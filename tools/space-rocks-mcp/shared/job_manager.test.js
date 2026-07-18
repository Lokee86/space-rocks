import assert from "node:assert/strict";
import { EventEmitter } from "node:events";
import { test } from "node:test";

import { ProcessJobManager, terminateProcessTree } from "./job_manager.js";

class FakeProcess extends EventEmitter {
  constructor() {
    super();
    this.stdout = new EventEmitter();
    this.stderr = new EventEmitter();
    this.stdin = new EventEmitter();
    this.stdin.endCalls = 0;
    this.stdin.end = () => { this.stdin.endCalls += 1; };
    this.pid = FakeProcess.nextPid++;
    this.exitCode = null;
    this.killed = false;
    this.killCalls = [];
  }

  close(code = 0, signal = null) {
    this.exitCode = code;
    this.emit("close", code, signal);
  }

  kill(signal = "SIGTERM") {
    this.killCalls.push(signal);
    this.killed = true;
    this.close(null, signal);
  }
}

FakeProcess.nextPid = 1000;

function tick() {
  return new Promise((resolve) => setImmediate(resolve));
}

function wait(milliseconds) {
  return new Promise((resolve) => setTimeout(resolve, milliseconds));
}

function createManager({ processes = [], ...options } = {}) {
  const spawned = [];
  const manager = new ProcessJobManager({
    retentionMs: 60_000,
    spawnProcess: (command, args, spawnOptions) => {
      const child = processes.shift() ?? new FakeProcess();
      spawned.push({ child, command, args, spawnOptions });
      return child;
    },
    ...options,
  });
  return { manager, spawned };
}

test("queues jobs at the configured concurrency limit", async (t) => {
  const { manager, spawned } = createManager({ concurrency: 2 });
  t.after(() => manager.closeAll());

  const first = manager.start({ command: "one" });
  const second = manager.start({ command: "two" });
  const third = manager.start({ command: "three" });

  assert.match(first.jobId, /^job_[A-Za-z0-9_-]+$/);
  assert.equal(manager.status(first.jobId).state, "running");
  assert.equal(manager.status(second.jobId).state, "running");
  assert.equal(manager.status(third.jobId).state, "queued");
  assert.equal(spawned.length, 2);
  assert.equal(spawned[0].spawnOptions.shell, false);
  assert.equal(spawned[0].child.stdin.endCalls, 1);

  spawned[0].child.close(0);
  await tick();
  assert.equal(manager.status(third.jobId).state, "running");
});

test("captures stdout and completes a real process", async (t) => {
  const manager = new ProcessJobManager({ retentionMs: 60_000, defaultTimeoutMs: 5_000 });
  t.after(() => manager.closeAll());

  const job = manager.start({
    command: process.execPath,
    args: ["-e", "process.stdout.write('real-output')"],
  });
  let status;
  for (let attempt = 0; attempt < 100; attempt += 1) {
    status = manager.status(job.jobId);
    if (["succeeded", "failed", "cancelled", "timed_out"].includes(status.state)) break;
    await wait(10);
  }

  assert.equal(status.state, "succeeded");
  assert.equal(manager.read(job.jobId).data, "real-output");
});

test("falls back to child kill when Windows taskkill fails", async () => {
  for (const mode of ["error", "nonzero", "throw"]) {
    const child = new FakeProcess();
    const taskkill = new EventEmitter();
    const spawnProcess = mode === "throw"
      ? () => { throw new Error("taskkill unavailable"); }
      : () => {
        setImmediate(() => mode === "error" ? taskkill.emit("error", new Error("taskkill failed")) : taskkill.emit("close", 7));
        return taskkill;
      };

    terminateProcessTree(child, { platform: "win32", spawnProcess });
    await tick();
    assert.deepEqual(child.killCalls, ["SIGTERM"], mode);
  }
});

test("retains bounded output and reports cursor rollover offsets", (t) => {
  const { manager, spawned } = createManager({ maxOutputChars: 5 });
  t.after(() => manager.closeAll());

  const job = manager.start({ command: "output" });
  const child = spawned[0].child;
  child.stdout.emit("data", "abcde");
  child.stderr.emit("data", "err");

  const firstRead = manager.read(job.jobId, { cursor: 0, maxChars: 3 });
  assert.deepEqual(firstRead, {
    jobId: job.jobId,
    stream: "stdout",
    data: "abc",
    cursor: 0,
    nextCursor: 3,
    startOffset: 0,
    endOffset: 5,
    truncated: false,
  });

  child.stdout.emit("data", "fgh");
  const rolledRead = manager.read(job.jobId, { cursor: firstRead.nextCursor });
  assert.equal(rolledRead.data, "defgh");
  assert.equal(rolledRead.startOffset, 3);
  assert.equal(rolledRead.endOffset, 8);
  assert.equal(rolledRead.truncated, false);
  assert.equal(manager.read(job.jobId, { stream: "stdout", cursor: 0 }).truncated, true);
  assert.equal(manager.read(job.jobId, { stream: "stderr", cursor: 0 }).data, "err");
});

test("marks zero exits succeeded and non-zero exits failed", (t) => {
  const first = new FakeProcess();
  const second = new FakeProcess();
  const { manager, spawned } = createManager({
    concurrency: 1,
    processes: [first, second],
  });
  t.after(() => manager.closeAll());

  const success = manager.start({ command: "success" });
  first.close(0);
  const failure = manager.start({ command: "failure" });
  assert.equal(manager.status(success.jobId).state, "succeeded");
  assert.equal(manager.status(failure.jobId).state, "running");
  second.close(7, "SIGTERM");
  assert.equal(manager.status(failure.jobId).state, "failed");
  assert.equal(manager.status(failure.jobId).exitCode, 7);
  assert.equal(spawned.length, 2);
});

test("spawns private args but exposes only public args and metadata", (t) => {
  const { manager, spawned } = createManager();
  t.after(() => manager.closeAll());

  const job = manager.start({
    command: "private-args",
    args: ["--token", "super-secret"],
    publicArgs: ["--token", "[redacted]"],
    metadata: { requestId: "request-1" },
  });

  assert.deepEqual(spawned[0].args, ["--token", "super-secret"]);
  assert.deepEqual(job.args, ["--token", "[redacted]"]);
  assert.deepEqual(job.metadata, { requestId: "request-1" });
  assert.equal(JSON.stringify(manager.status(job.jobId)).includes("super-secret"), false);
  assert.deepEqual(manager.list()[0].args, ["--token", "[redacted]"]);

  const defaulted = manager.start({ command: "defaults", args: ["--visible"] });
  assert.deepEqual(defaulted.args, ["--visible"]);
});

test("times out a running process through the injectable terminator", async (t) => {
  const child = new FakeProcess();
  const terminations = [];
  const { manager } = createManager({
    processes: [child],
    terminateProcessTree: (process, details) => {
      terminations.push({ process, details });
      process.close(null, "SIGTERM");
    },
  });
  t.after(() => manager.closeAll());

  const job = manager.start({ command: "slow", timeoutMs: 25 });
  await new Promise((resolve) => setTimeout(resolve, 40));

  assert.equal(manager.status(job.jobId).state, "timed_out");
  assert.equal(terminations.length, 1);
  assert.equal(terminations[0].details.reason, "timeout");
});

test("cancels queued and running jobs", async (t) => {
  const first = new FakeProcess();
  const second = new FakeProcess();
  const terminations = [];
  const { manager, spawned } = createManager({
    concurrency: 1,
    processes: [first, second],
    terminateProcessTree: (process, details) => {
      terminations.push(details.reason);
      process.close(null, "SIGTERM");
    },
  });
  t.after(() => manager.closeAll());

  const running = manager.start({ command: "running" });
  const queued = manager.start({ command: "queued" });
  assert.equal(manager.cancel(queued.jobId).state, "cancelled");
  assert.equal(spawned.length, 1);

  assert.equal(manager.cancel(running.jobId).state, "cancelled");
  await tick();
  assert.deepEqual(terminations, ["cancel"]);
  assert.equal(manager.status(running.jobId).state, "cancelled");
  assert.equal(spawned.length, 1);
  assert.equal(manager.list().find((job) => job.jobId === queued.jobId).state, "cancelled");

  const next = manager.start({ command: "next" });
  assert.equal(next.state, "running");
  assert.equal(spawned.length, 2);
  second.close(0);
});

test("force-finalizes non-closing cancellation and timeout jobs", async (t) => {
  const first = new FakeProcess();
  const second = new FakeProcess();
  const { manager, spawned } = createManager({
    concurrency: 1,
    processes: [first, second],
    terminationGraceMs: 20,
    terminateProcessTree: () => {},
  });
  t.after(() => manager.closeAll());

  const cancelled = manager.start({ command: "cancelled" });
  const queued = manager.start({ command: "queued" });
  manager.cancel(cancelled.jobId);
  await wait(35);

  const cancelledStatus = manager.status(cancelled.jobId);
  assert.equal(cancelledStatus.state, "cancelled");
  assert.match(cancelledStatus.error, /did not exit within the 20ms termination grace period/);
  assert.ok(cancelledStatus.completedAt);
  assert.equal(manager.status(queued.jobId).state, "running");
  assert.equal(spawned.length, 2);

  const timeoutChild = new FakeProcess();
  const { manager: timeoutManager } = createManager({
    processes: [timeoutChild],
    terminationGraceMs: 20,
    terminateProcessTree: () => {},
  });
  t.after(() => timeoutManager.closeAll());
  const timedOut = timeoutManager.start({ command: "timed-out", timeoutMs: 10 });
  await wait(40);

  const timedOutStatus = timeoutManager.status(timedOut.jobId);
  assert.equal(timedOutStatus.state, "timed_out");
  assert.match(timedOutStatus.error, /did not exit within the 20ms termination grace period/);
  assert.ok(timedOutStatus.completedAt);
});

test("expires completed jobs after retention", (t) => {
  let now = 1_000;
  const child = new FakeProcess();
  const { manager } = createManager({
    processes: [child],
    retentionMs: 100,
    sweepIntervalMs: 10_000,
    now: () => now,
  });
  t.after(() => manager.closeAll());

  const job = manager.start({ command: "retained" });
  child.close(0);
  assert.equal(manager.list().length, 1);

  now += 100;
  assert.deepEqual(manager.cleanupExpired(), [job.jobId]);
  assert.deepEqual(manager.list(), []);
  assert.throws(() => manager.status(job.jobId), /Unknown process job/);
});
