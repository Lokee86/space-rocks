import crypto from "node:crypto";
import { spawn } from "node:child_process";

export const JOB_STATES = Object.freeze(["queued", "running", "succeeded", "failed", "cancelled", "timed_out"]);
const TERMINAL_STATES = new Set(["succeeded", "failed", "cancelled", "timed_out"]);
const DEFAULT_CONCURRENCY = 4;
const DEFAULT_MAX_OUTPUT_CHARS = 50_000;
const DEFAULT_TIMEOUT_MS = 5 * 60 * 1000;
const DEFAULT_RETENTION_MS = 15 * 60 * 1000;
const DEFAULT_TERMINATION_GRACE_MS = 5_000;

function newJobId() { return `job_${crypto.randomBytes(18).toString("base64url")}`; }
function isoTime(timestamp) { return timestamp === null ? null : new Date(timestamp).toISOString(); }
function positiveInteger(value, name) {
  if (!Number.isInteger(value) || value < 1) throw new Error(`${name} must be a positive integer`);
  return value;
}
function timeoutValue(value, name) { return value === null ? null : positiveInteger(value, name); }
function cloneMetadata(metadata) {
  if (metadata === null || metadata === undefined) return null;
  return structuredClone(metadata);
}

class RollingOutputBuffer {
  constructor(maxChars) { this.maxChars = maxChars; this.text = ""; this.totalOffset = 0; }
  append(chunk) {
    const text = Buffer.isBuffer(chunk) ? chunk.toString() : String(chunk);
    this.totalOffset += text.length;
    this.text = `${this.text}${text}`.slice(-this.maxChars);
  }
  get startOffset() { return this.totalOffset - this.text.length; }
  read(cursor = 0, maxChars = this.maxChars) {
    const requestedCursor = Math.max(0, cursor);
    const startOffset = this.startOffset;
    const effectiveCursor = Math.max(requestedCursor, startOffset);
    const endOffset = Math.min(effectiveCursor + maxChars, this.totalOffset);
    return {
      data: this.text.slice(effectiveCursor - startOffset, endOffset - startOffset),
      cursor: requestedCursor,
      nextCursor: endOffset,
      startOffset,
      endOffset: this.totalOffset,
      truncated: requestedCursor < startOffset,
    };
  }
  snapshot() {
    return { startOffset: this.startOffset, totalOffset: this.totalOffset, availableChars: this.text.length };
  }
}

export function terminateProcessTree(child, { platform = process.platform, spawnProcess = spawn } = {}) {
  if (!child || (child.exitCode !== null && child.exitCode !== undefined) || child.killed) return;
  if (platform === "win32" && child.pid) {
    let fallbackUsed = false;
    const fallback = () => {
      if (fallbackUsed) return;
      fallbackUsed = true;
      try { child.kill?.("SIGTERM"); } catch {}
    };
    try {
      const taskkill = spawnProcess("taskkill", ["/pid", String(child.pid), "/T", "/F"], { shell: false, windowsHide: true });
      taskkill.on("error", fallback);
      taskkill.on("close", (code) => { if (code !== 0) fallback(); });
    } catch {
      fallback();
    }
    return;
  }
  if (child.pid) {
    try { process.kill(-child.pid, "SIGTERM"); return; } catch { /* Fall back when no process group exists. */ }
  }
  try { child.kill?.("SIGTERM"); } catch {}
}

export class ProcessJobManager {
  constructor({
    concurrency = DEFAULT_CONCURRENCY,
    maxOutputChars = DEFAULT_MAX_OUTPUT_CHARS,
    defaultTimeoutMs = DEFAULT_TIMEOUT_MS,
    retentionMs = DEFAULT_RETENTION_MS,
    sweepIntervalMs = Math.min(retentionMs, 60_000),
    terminationGraceMs = DEFAULT_TERMINATION_GRACE_MS,
    spawnProcess = spawn,
    terminateProcessTree: terminate = terminateProcessTree,
    now = Date.now,
  } = {}) {
    this.concurrency = positiveInteger(concurrency, "concurrency");
    this.maxOutputChars = positiveInteger(maxOutputChars, "maxOutputChars");
    this.defaultTimeoutMs = timeoutValue(defaultTimeoutMs, "defaultTimeoutMs");
    this.retentionMs = positiveInteger(retentionMs, "retentionMs");
    this.sweepIntervalMs = positiveInteger(sweepIntervalMs, "sweepIntervalMs");
    this.terminationGraceMs = positiveInteger(terminationGraceMs, "terminationGraceMs");
    this.spawnProcess = spawnProcess;
    this.terminateProcessTree = terminate;
    this.now = now;
    this.jobs = new Map();
    this.queue = [];
    this.activeCount = 0;
    this.closed = false;
    this.pumping = false;
    this.sweepTimer = setInterval(() => this.cleanupExpired(), this.sweepIntervalMs);
    this.sweepTimer.unref?.();
  }

  start({ command, args = [], publicArgs, cwd, env, timeoutMs, metadata } = {}) {
    if (this.closed) throw new Error("ProcessJobManager is closed");
    if (typeof command !== "string" || command.length === 0) throw new Error("command must be a non-empty string");
    if (!Array.isArray(args) || args.some((arg) => typeof arg !== "string")) throw new Error("args must be an array of strings");
    const exposedArgs = publicArgs === undefined ? args : publicArgs;
    if (!Array.isArray(exposedArgs) || exposedArgs.some((arg) => typeof arg !== "string")) throw new Error("publicArgs must be an array of strings");
    const job = {
      jobId: newJobId(), command, args: [...args], publicArgs: [...exposedArgs], metadata: cloneMetadata(metadata), cwd, env,
      timeoutMs: timeoutMs === undefined ? this.defaultTimeoutMs : timeoutValue(timeoutMs, "timeoutMs"),
      state: "queued", createdAtMs: this.now(), startedAtMs: null, completedAtMs: null,
      exitCode: null, exitSignal: null, error: null, child: null, timer: null, terminationTimer: null, finished: false,
      stdout: new RollingOutputBuffer(this.maxOutputChars), stderr: new RollingOutputBuffer(this.maxOutputChars),
    };
    this.jobs.set(job.jobId, job);
    this.queue.push(job);
    this._pump();
    return this._snapshot(job);
  }

  list() { this.cleanupExpired(); return [...this.jobs.values()].map((job) => this._snapshot(job)); }
  status(jobId) { return this._snapshot(this._getJob(jobId)); }

  read(jobId, { stream = "stdout", cursor = 0, maxChars = this.maxOutputChars } = {}) {
    const job = this._getJob(jobId);
    if (stream !== "stdout" && stream !== "stderr") throw new Error("stream must be stdout or stderr");
    if (!Number.isInteger(cursor) || cursor < 0) throw new Error("cursor must be a non-negative integer");
    positiveInteger(maxChars, "maxChars");
    return { jobId, stream, ...job[stream].read(cursor, maxChars) };
  }

  cancel(jobId) {
    const job = this._getJob(jobId);
    if (TERMINAL_STATES.has(job.state)) return this._snapshot(job);
    job.state = "cancelled";
    if (job.child) this._terminate(job, "cancel");
    else {
      job.finished = true;
      job.completedAtMs = this.now();
      this._pump();
    }
    return this._snapshot(job);
  }

  cleanupExpired(now = this.now()) {
    const expiredJobIds = [];
    for (const [jobId, job] of this.jobs) {
      if (job.finished && job.completedAtMs !== null && now - job.completedAtMs >= this.retentionMs) {
        this.jobs.delete(jobId);
        expiredJobIds.push(jobId);
      }
    }
    return expiredJobIds;
  }

  closeAll() {
    if (this.closed) return;
    this.closed = true;
    clearInterval(this.sweepTimer);
    this.sweepTimer = null;
    for (const job of this.jobs.values()) {
      if (!job.finished) {
        if (!TERMINAL_STATES.has(job.state)) job.state = "cancelled";
        if (job.child) this._terminate(job, "close");
        job.completedAtMs = this.now();
      }
      clearTimeout(job.timer);
      job.timer = null;
      clearTimeout(job.terminationTimer);
      job.terminationTimer = null;
      job.finished = true;
    }
    this.activeCount = 0;
    this.queue = [];
    this.jobs.clear();
  }

  _getJob(jobId) {
    this.cleanupExpired();
    const job = this.jobs.get(jobId);
    if (!job) throw new Error(`Unknown process job: ${jobId}`);
    return job;
  }

  _pump() {
    if (this.closed || this.pumping) return;
    this.pumping = true;
    try {
      while (this.activeCount < this.concurrency && this.queue.length > 0) {
        const job = this.queue.shift();
        if (job.state === "queued") this._startJob(job);
      }
    } finally { this.pumping = false; }
  }

  _startJob(job) {
    const options = {
      shell: false, windowsHide: true,
      ...(job.cwd === undefined ? {} : { cwd: job.cwd }),
      ...(job.env === undefined ? {} : { env: job.env }),
    };
    if (process.platform !== "win32") options.detached = true;
    let child;
    try { child = this.spawnProcess(job.command, job.args, options); }
    catch (error) { this._failBeforeStart(job, error); return; }

    job.child = child;
    job.state = "running";
    job.startedAtMs = this.now();
    this.activeCount += 1;
    if (child.stdin) {
      child.stdin.on?.("error", () => {});
      try { child.stdin.end?.(); } catch {}
    }
    child.stdout?.on("data", (chunk) => job.stdout.append(chunk));
    child.stderr?.on("data", (chunk) => job.stderr.append(chunk));
    child.on("error", (error) => { job.error = error instanceof Error ? error.message : String(error); });
    child.on("close", (code, signal) => this._finish(job, code, signal));
    if (job.timeoutMs !== null) {
      job.timer = setTimeout(() => {
        if (!job.finished) { job.state = "timed_out"; this._terminate(job, "timeout"); }
      }, job.timeoutMs);
      job.timer.unref?.();
    }
  }

  _failBeforeStart(job, error) {
    job.state = "failed";
    job.error = error instanceof Error ? error.message : String(error);
    job.finished = true;
    job.completedAtMs = this.now();
  }

  _terminate(job, reason) {
    if (job.terminationTimer === null) {
      job.terminationTimer = setTimeout(() => this._forceFinish(job), this.terminationGraceMs);
      job.terminationTimer.unref?.();
    }
    try { this.terminateProcessTree(job.child, { reason }); }
    catch (error) { job.error = `${job.error ? `${job.error}; ` : ""}${error.message ?? error}`; }
  }

  _forceFinish(job) {
    if (job.finished) return;
    job.finished = true;
    clearTimeout(job.timer);
    job.timer = null;
    job.terminationTimer = null;
    job.exitCode = job.child?.exitCode ?? null;
    job.error = `${job.error ? `${job.error}; ` : ""}Process did not exit within the ${this.terminationGraceMs}ms termination grace period`;
    job.completedAtMs = this.now();
    this.activeCount = Math.max(0, this.activeCount - 1);
    this._pump();
  }

  _finish(job, code, signal) {
    if (job.finished) return;
    job.finished = true;
    clearTimeout(job.timer);
    job.timer = null;
    clearTimeout(job.terminationTimer);
    job.terminationTimer = null;
    job.exitCode = code ?? null;
    job.exitSignal = signal ?? null;
    if (!TERMINAL_STATES.has(job.state)) job.state = job.error || (code !== null && code !== 0) ? "failed" : "succeeded";
    job.completedAtMs = this.now();
    this.activeCount = Math.max(0, this.activeCount - 1);
    this._pump();
  }

  _snapshot(job) {
    return {
      jobId: job.jobId, command: job.command, args: [...job.publicArgs], metadata: cloneMetadata(job.metadata), cwd: job.cwd ?? null,
      state: job.state, timeoutMs: job.timeoutMs,
      createdAt: isoTime(job.createdAtMs), startedAt: isoTime(job.startedAtMs), completedAt: isoTime(job.completedAtMs),
      exitCode: job.exitCode, exitSignal: job.exitSignal, error: job.error,
      stdout: job.stdout.snapshot(), stderr: job.stderr.snapshot(),
    };
  }
}

export const defaultProcessJobManager = new ProcessJobManager();
process.once("exit", () => defaultProcessJobManager.closeAll());
