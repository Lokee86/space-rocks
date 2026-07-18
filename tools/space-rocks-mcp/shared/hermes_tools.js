import crypto from "node:crypto";
import { EventEmitter } from "node:events";
import path from "node:path";
import { spawn, execFileSync } from "node:child_process";
import { z } from "zod";
import pty from "node-pty";

import { WORKSPACE_ROOT } from "./paths.js";
import { textResponse } from "./responses.js";

const HERMES_ROOT = path.resolve("C:/Users/archa/AppData/Local/hermes");
const HERMES_ROOT_DESCRIPTION = `Allowed cwd values: WORKSPACE_ROOT (${WORKSPACE_ROOT}), descendants of WORKSPACE_ROOT, ${HERMES_ROOT}, or descendants of ${HERMES_ROOT}. Defaults to WORKSPACE_ROOT.`;

const MAX_OUTPUT_CHARS = 50000;
const DEFAULT_COLS = 120;
const DEFAULT_ROWS = 30;
const DEFAULT_IDLE_MS = 400;
const DEFAULT_TIMEOUT_MS = 30000;
const DEFAULT_HERMES_JOB_TIMEOUT_MS = 300000;
const INITIAL_STARTUP_TIMEOUT_MS = 5000;
const INITIAL_STARTUP_IDLE_MS = 75;
const SESSION_IDLE_TTL_MS = 15 * 60 * 1000;
const SESSION_IDLE_SWEEP_MS = 60 * 1000;
const SESSION_ID_BYTES = 18;
const HERMES_COMMAND = process.env.HERMES_COMMAND || (() => {
  try { return execFileSync("where", ["hermes"], { encoding: "utf8" }).split(/\r?\n/).find(Boolean) || "hermes"; }
  catch { return "hermes"; }
})();

function truncateOutput(output) { return output.length > MAX_OUTPUT_CHARS ? output.slice(0, MAX_OUTPUT_CHARS) + `\n\n[Output truncated at ${MAX_OUTPUT_CHARS} characters]` : output; }
function stripControlNoise(text) { return text.replace(/\u001b\[[0-9;?]*[ -/]*[@-~]/g, "").replace(/[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f]/g, (ch) => (ch === "\n" || ch === "\r" || ch === "\t" ? ch : "")).replace(/\r(?!\n)/g, "\n").replace(/[ \t]+\n/g, "\n"); }
function normalizeOutput(text) { return stripControlNoise(text); }
function newSessionId() { return `ht_${crypto.randomBytes(SESSION_ID_BYTES).toString("base64url")}`; }
function isWithinRoot(candidate, root) { return candidate === root || candidate.startsWith(root + path.sep); }
function validateHermesCwd(cwd) {
  if (cwd === undefined) return WORKSPACE_ROOT;
  const resolved = path.resolve(cwd);
  if (isWithinRoot(resolved, WORKSPACE_ROOT) || isWithinRoot(resolved, HERMES_ROOT)) return resolved;
  throw new Error(`cwd must be WORKSPACE_ROOT, a descendant of WORKSPACE_ROOT, ${HERMES_ROOT}, or a descendant of ${HERMES_ROOT}`);
}
function safeWorkspaceCwd(cwd) { return validateHermesCwd(cwd); }
function boundedTailAppend(current, addition) { const next = current + addition; if (next.length <= MAX_OUTPUT_CHARS) return { text: next, dropped: 0 }; const dropped = next.length - MAX_OUTPUT_CHARS; return { text: next.slice(dropped), dropped }; }

class HermesTerminalSession extends EventEmitter {
  constructor({ cwd, cols, rows, spawnPty = pty.spawn }) { super(); this.cwd = cwd; this.cols = cols; this.rows = rows; this.spawnPty = spawnPty; this.createdAt = new Date().toISOString(); this.lastActivityAt = Date.now(); this.output = ""; this.bufferStartOffset = 0; this.totalOutputChars = 0; this.readOffset = 0; this.closed = false; this.exitCode = null; this.exitSignal = null; this.pending = Promise.resolve(); this._start(); }
  _start() { this.pty = this.spawnPty(HERMES_COMMAND, ["chat", "--cli"], { name: "xterm-color", cols: this.cols, rows: this.rows, cwd: this.cwd, env: process.env, handleFlowControl: true }); this.pty.onData((data) => this._appendOutput(data)); this.pty.onExit(({ exitCode, signal }) => { this.exitCode = exitCode; this.exitSignal = signal; this.closed = true; this.emit("exit", { exitCode, signal }); }); }
  _appendOutput(chunk) { const normalized = normalizeOutput(chunk); if (!normalized) return; this.totalOutputChars += normalized.length; const { text, dropped } = boundedTailAppend(this.output, normalized); this.output = text; this.bufferStartOffset = this.totalOutputChars - this.output.length; if (dropped > 0) this.readOffset = Math.max(this.bufferStartOffset, this.readOffset - dropped); this.lastActivityAt = Date.now(); this.emit("output"); }
  enqueue(task) { const next = this.pending.then(task, task); this.pending = next.then(() => undefined, () => undefined); return next; }
  resize(cols, rows) { if (this.closed) throw new Error("Session is closed"); this.pty.resize(cols, rows); this.cols = cols; this.rows = rows; this.lastActivityAt = Date.now(); }
  consumeFromOffset(startOffset, maxChars = MAX_OUTPUT_CHARS) { const sliceStart = Math.max(startOffset, this.bufferStartOffset) - this.bufferStartOffset; if (sliceStart >= this.output.length) return ""; return this.output.slice(sliceStart, sliceStart + maxChars); }
  consumeUnread({ maxChars = MAX_OUTPUT_CHARS } = {}) { const unread = this.consumeFromOffset(this.readOffset, maxChars); this.readOffset = Math.max(this.readOffset, this.bufferStartOffset + unread.length); return unread; }
  waitForOutputAfter(startTotalOffset, { timeoutMs, idleMs, requireNewOutput = true }) { const startedAt = Date.now(); return new Promise((resolve) => { let idleTimer; let timeoutTimer; let seenNewOutput = !requireNewOutput; const cleanup = () => { clearTimeout(idleTimer); clearTimeout(timeoutTimer); this.off("output", onOutput); this.off("exit", onExit); }; const finish = () => { cleanup(); resolve(); }; const onOutput = () => { if (this.totalOutputChars > startTotalOffset) seenNewOutput = true; if (!seenNewOutput) return; clearTimeout(idleTimer); idleTimer = setTimeout(finish, idleMs); idleTimer.unref?.(); }; const onExit = () => finish(); this.on("output", onOutput); this.on("exit", onExit); timeoutTimer = setTimeout(finish, Math.max(timeoutMs - (Date.now() - startedAt), 0)); timeoutTimer.unref?.(); if (seenNewOutput) { idleTimer = setTimeout(finish, idleMs); idleTimer.unref?.(); } }); }
  async settleStartup() { await this.waitForOutputAfter(this.totalOutputChars, { idleMs: INITIAL_STARTUP_IDLE_MS, timeoutMs: INITIAL_STARTUP_TIMEOUT_MS, requireNewOutput: true }); }
  async send(input, { appendEnter = true, idleMs = DEFAULT_IDLE_MS, timeoutMs = DEFAULT_TIMEOUT_MS } = {}) { return this.enqueue(async () => { if (this.closed) throw new Error("Session is closed"); const startReadOffset = this.readOffset; const startTotalOffset = this.totalOutputChars; this.pty.write(appendEnter ? `${input}\r` : input); await this.waitForOutputAfter(startTotalOffset, { idleMs, timeoutMs, requireNewOutput: false }); const produced = this.consumeFromOffset(startReadOffset, MAX_OUTPUT_CHARS); this.readOffset = Math.max(this.readOffset, Math.max(startReadOffset, this.bufferStartOffset) + produced.length); return produced; }); }
  read({ maxChars = MAX_OUTPUT_CHARS } = {}) { return this.consumeUnread({ maxChars }); }
  close() { if (this.closed) return { closed: true, exitCode: this.exitCode, exitSignal: this.exitSignal }; this.closed = true; try { this.pty.kill(); } catch {} return { closed: true, exitCode: this.exitCode, exitSignal: this.exitSignal }; }
  expired(now = Date.now()) { return !this.closed && now - this.lastActivityAt > SESSION_IDLE_TTL_MS; }
}

export class HermesTerminalManager { constructor({ spawnPty = pty.spawn } = {}) { this.spawnPty = spawnPty; this.sessions = new Map(); this._sweepTimer = setInterval(() => this.expireIdleSessions(), SESSION_IDLE_SWEEP_MS); this._sweepTimer.unref?.(); process.on("exit", () => this.closeAll()); } start({ cwd, cols = DEFAULT_COLS, rows = DEFAULT_ROWS } = {}) { const session = new HermesTerminalSession({ cwd: safeWorkspaceCwd(cwd), cols, rows, spawnPty: this.spawnPty }); const sessionId = newSessionId(); this.sessions.set(sessionId, session); return { sessionId, session }; } get(sessionId) { const session = this.sessions.get(sessionId); if (!session) throw new Error(`Unknown Hermes PTY session: ${sessionId}`); return session; } send(sessionId, options) { const session = this.get(sessionId); if (session.closed) throw new Error(`Hermes PTY session is closed: ${sessionId}`); return session.send(options.input, options); } read(sessionId, options) { const session = this.get(sessionId); if (session.closed) throw new Error(`Hermes PTY session is closed: ${sessionId}`); return session.read(options); } resize(sessionId, cols, rows) { const session = this.get(sessionId); if (session.closed) throw new Error(`Hermes PTY session is closed: ${sessionId}`); return session.resize(cols, rows); } close(sessionId) { const session = this.sessions.get(sessionId); if (!session) throw new Error(`Unknown Hermes PTY session: ${sessionId}`); const result = session.close(); this.sessions.delete(sessionId); return result; } expireIdleSessions(now = Date.now()) { for (const [sessionId, session] of this.sessions.entries()) if (session.closed || session.expired(now)) this.sessions.delete(sessionId); } closeAll() { for (const sessionId of [...this.sessions.keys()]) this.close(sessionId); } }

const sessionNameSchema = z.string().regex(/^[a-zA-Z0-9_\-\. ]+$/, "session_name must contain only letters, numbers, dash, underscore, dot, and space");
const sessionIdSchema = z.string().regex(/^ht_[A-Za-z0-9_-]+$/, "session_id must be a Hermes PTY session id");
const cwdSchema = z.string().optional().describe(HERMES_ROOT_DESCRIPTION);
const inputSchema = z.string().min(1, "input must be non-empty");
export const defaultHermesTerminalManager = new HermesTerminalManager();
function renderSessionResult(sessionId, session, extra = {}) { return { session_id: sessionId, cwd: session.cwd, cols: session.cols, rows: session.rows, created_at: session.createdAt, closed: session.closed, exit_code: session.exitCode, exit_signal: session.exitSignal, unread_chars: Math.max(0, session.totalOutputChars - session.readOffset), ...extra }; }
async function runHermes({ args = [], stdin, cwd }) { const resolvedCwd = validateHermesCwd(cwd); return new Promise((resolve) => { const child = spawn("hermes", args, { cwd: resolvedCwd, shell: false }); let stdout = ""; let stderr = ""; child.stdout.on("data", (data) => { stdout += data.toString(); }); child.stderr.on("data", (data) => { stderr += data.toString(); }); child.stdin.on("error", () => {}); if (stdin !== undefined) child.stdin.write(stdin); child.stdin.end(); child.on("close", (code) => resolve({ command: "hermes", args, cwd: cwd || ".", exit_code: code, timed_out: false, stdout: truncateOutput(stdout), stderr: truncateOutput(stderr) })); child.on("error", (err) => resolve({ command: "hermes", args, cwd: cwd || ".", exit_code: 1, timed_out: false, stdout: truncateOutput(stdout), stderr: truncateOutput(stderr + err.message) })); }); }
function previewPrompt(prompt) { return prompt.length > 80 ? `${prompt.slice(0, 80)}…` : prompt; }
const ACTIVE_HERMES_JOB_STATES = new Set(["queued", "running"]);
function hasActiveHermesJob(manager, sessionName) { return manager.list().some((job) => ACTIVE_HERMES_JOB_STATES.has(job.state) && job.metadata?.kind === "hermes_session_job" && job.metadata.session_name === sessionName); }
function registerHermesJobTool(server, manager) {
  server.registerTool("hermes_job_start", {
    title: "Hermes Job Start",
    description: "Starts an asynchronous Hermes session prompt job.",
    inputSchema: { prompt: inputSchema, session_name: sessionNameSchema.default("space-rocks-mcp"), cwd: cwdSchema, timeout_ms: z.number().int().min(1000).max(3600000).default(DEFAULT_HERMES_JOB_TIMEOUT_MS) },
  }, async ({ prompt, session_name, cwd, timeout_ms }) => {
    if (hasActiveHermesJob(manager, session_name)) throw new Error(`A Hermes job is already queued or running for session: ${session_name}`);
    const resolvedCwd = validateHermesCwd(cwd);
    const args = ["chat", "-Q", "--continue", session_name, "--query", prompt];
    return textResponse(JSON.stringify(manager.start({
      command: HERMES_COMMAND,
      args,
      publicArgs: ["chat", "-Q", "--continue", session_name, "--query", "[redacted]"],
      cwd: resolvedCwd,
      timeoutMs: timeout_ms,
      metadata: { kind: "hermes_session_job", session_name, prompt_chars: prompt.length },
    }), null, 2));
  });
}
function registerTerminalTools(server, manager) { server.registerTool("hermes_terminal_start", { title: "Hermes Terminal Start", description: "Starts a persistent interactive Hermes PTY session.", inputSchema: { cwd: cwdSchema, cols: z.number().int().min(40).max(400).default(DEFAULT_COLS), rows: z.number().int().min(10).max(200).default(DEFAULT_ROWS) } }, async ({ cwd, cols, rows }) => { const { sessionId, session } = manager.start({ cwd, cols, rows }); await session.settleStartup(); return textResponse(JSON.stringify(renderSessionResult(sessionId, session, { output: session.consumeUnread({ maxChars: MAX_OUTPUT_CHARS }) }), null, 2)); }); server.registerTool("hermes_terminal_send", { title: "Hermes Terminal Send", description: "Writes input to a persistent Hermes PTY session and waits for prompt activity to settle.", inputSchema: { session_id: sessionIdSchema, input: inputSchema, append_enter: z.boolean().default(true), idle_ms: z.number().int().min(20).max(120000).default(DEFAULT_IDLE_MS), timeout_ms: z.number().int().min(1000).max(300000).default(DEFAULT_TIMEOUT_MS) } }, async ({ session_id, input, append_enter, idle_ms, timeout_ms }) => { const session = manager.get(session_id); const output = await manager.send(session_id, { input, appendEnter: append_enter, idleMs: idle_ms, timeoutMs: timeout_ms }); return textResponse(JSON.stringify(renderSessionResult(session_id, session, { output }), null, 2)); }); server.registerTool("hermes_terminal_read", { title: "Hermes Terminal Read", description: "Reads unread incremental output from a Hermes PTY session.", inputSchema: { session_id: sessionIdSchema, max_chars: z.number().int().min(1).max(MAX_OUTPUT_CHARS).default(MAX_OUTPUT_CHARS) } }, async ({ session_id, max_chars }) => { const session = manager.get(session_id); return textResponse(JSON.stringify(renderSessionResult(session_id, session, { output: session.read({ maxChars: max_chars }) }), null, 2)); }); server.registerTool("hermes_terminal_resize", { title: "Hermes Terminal Resize", description: "Resizes a persistent Hermes PTY session.", inputSchema: { session_id: sessionIdSchema, cols: z.number().int().min(20).max(400).default(DEFAULT_COLS), rows: z.number().int().min(5).max(200).default(DEFAULT_ROWS) } }, async ({ session_id, cols, rows }) => { const session = manager.get(session_id); manager.resize(session_id, cols, rows); return textResponse(JSON.stringify(renderSessionResult(session_id, session, { resized: true }), null, 2)); }); server.registerTool("hermes_terminal_close", { title: "Hermes Terminal Close", description: "Closes a persistent Hermes PTY session.", inputSchema: { session_id: sessionIdSchema } }, async ({ session_id }) => textResponse(JSON.stringify({ session_id, ...manager.close(session_id) }, null, 2))); }
export function registerHermesTools(server, { manager = defaultHermesTerminalManager, runHermesImpl = runHermes, processJobManager } = {}) { server.registerTool("hermes_run", { title: "Hermes Run", description: "Runs the Hermes CLI with arbitrary args and optional stdin.", inputSchema: { args: z.array(z.string()).default([]), stdin: z.string().optional(), cwd: cwdSchema } }, async ({ args, stdin, cwd }) => textResponse(JSON.stringify(await runHermesImpl({ args, stdin, cwd }), null, 2))); server.registerTool("hermes_ping", { title: "Hermes Ping", description: "Confirms the Hermes CLI is available.", inputSchema: {} }, async () => textResponse(JSON.stringify(await runHermesImpl({ args: ["--version"] }), null, 2))); server.registerTool("hermes_help", { title: "Hermes Help", description: "Shows Hermes CLI help.", inputSchema: {} }, async () => textResponse(JSON.stringify(await runHermesImpl({ args: ["--help"] }), null, 2))); server.registerTool("hermes_session_status", { title: "Hermes Session Status", description: "Shows the current Hermes session status.", inputSchema: {} }, async () => textResponse(JSON.stringify(await runHermesImpl({ args: ["status"] }), null, 2))); server.registerTool("hermes_sessions_list", { title: "Hermes Sessions List", description: "Lists all Hermes sessions.", inputSchema: {} }, async () => textResponse(JSON.stringify(await runHermesImpl({ args: ["sessions", "list"] }), null, 2))); server.registerTool("hermes_session_send", { title: "Hermes Session Send", description: "Sends a prompt to a Hermes session and returns the result.", inputSchema: { prompt: inputSchema, session_name: sessionNameSchema.default("space-rocks-mcp"), cwd: cwdSchema } }, async ({ prompt, session_name, cwd }) => textResponse(JSON.stringify(await runHermesImpl({ args: ["chat", "-Q", "--continue", session_name, "--query", prompt], cwd }), null, 2))); server.registerTool("hermes_session_send_batch", { title: "Hermes Session Send Batch", description: "Sends multiple prompts to a Hermes session and returns the results.", inputSchema: { prompts: z.array(inputSchema).min(1), session_name: sessionNameSchema.default("space-rocks-mcp"), cwd: cwdSchema } }, async ({ prompts, session_name, cwd }) => textResponse(JSON.stringify({ session_name, count: prompts.length, results: await Promise.all(prompts.map((prompt, index) => runHermesImpl({ args: ["chat", "-Q", "--continue", session_name, "--query", prompt], cwd }).then((result) => ({ index, prompt_preview: previewPrompt(prompt), result })))) }, null, 2))); registerTerminalTools(server, manager); if (processJobManager !== undefined) registerHermesJobTool(server, processJobManager); }
