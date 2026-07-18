import assert from "node:assert/strict";
import path from "node:path";
import { test } from "node:test";

import { WORKSPACE_ROOT } from "./paths.js";
import { HermesTerminalManager, registerHermesTools } from "./hermes_tools.js";

function createMockServer() {
  const tools = new Map();
  return { tools, registerTool(name, config, handler) { tools.set(name, { config, handler }); } };
}

function makeFakePty({ onWrite, onKill } = {}) {
  let onData;
  let onExit;
  return {
    writes: [], resizeCalls: [],
    onData(cb) { onData = cb; },
    onExit(cb) { onExit = cb; },
    write(data) { this.writes.push(data); onWrite?.(data, (text) => onData?.(text), (code = 0, signal = null) => onExit?.({ exitCode: code, signal })); },
    resize(cols, rows) { this.resizeCalls.push([cols, rows]); },
    kill() { onKill?.(); onExit?.({ exitCode: 0, signal: null }); },
    emit(text) { onData?.(text); },
  };
}

test("registerHermesTools exposes PTY tools without the async job by default", () => {
  const server = createMockServer();
  registerHermesTools(server);
  for (const name of ["hermes_terminal_start", "hermes_terminal_send", "hermes_terminal_read", "hermes_terminal_resize", "hermes_terminal_close"]) assert.ok(server.tools.has(name), `${name} should be registered`);
  assert.equal(server.tools.has("hermes_job_start"), false);
  const sendSchema = server.tools.get("hermes_terminal_send").config.inputSchema;
  assert.equal(sendSchema.session_id.format, "regex");
  assert.equal(sendSchema.append_enter.def.defaultValue, true);
});

test("HermesTerminalManager starts, queues sends, reads, resizes, and closes sessions", async () => {
  const created = [];
  const manager = new HermesTerminalManager({
    spawnPty: (command, args, options) => {
      const pty = makeFakePty({ onWrite: (data, emit) => setImmediate(() => emit(`echo:${data}`)) });
      Object.assign(pty, { command, args, options });
      created.push(pty);
      return pty;
    },
  });
  const { sessionId, session } = manager.start({ cols: 90, rows: 20 });
  assert.match(sessionId, /^ht_/);
  assert.match(created[0].command, /hermes(\.exe)?$/i);
  assert.equal(created[0].options.cwd, WORKSPACE_ROOT);
  await session.settleStartup();
  assert.equal(session.read({ maxChars: 100 }), "");
  const first = await manager.send(sessionId, { input: "hello", appendEnter: true, idleMs: 5, timeoutMs: 200 });
  assert.match(first, /echo:hello\n/);
  assert.equal(manager.read(sessionId, { maxChars: 100 }), "");
  manager.resize(sessionId, 120, 40);
  assert.deepEqual(created[0].resizeCalls.at(-1), [120, 40]);
  const closed = manager.close(sessionId);
  assert.equal(closed.closed, true);
  assert.throws(() => manager.send(sessionId, { input: "x" }), /Unknown Hermes PTY session/);
});

test("HermesTerminalManager preserves the newest output on rollover and adjusts unread state", () => {
  const manager = new HermesTerminalManager({ spawnPty: () => makeFakePty() });
  const { session } = manager.start({ cwd: "." });
  session._appendOutput("a".repeat(50010));
  assert.equal(session.output.length, 50000);
  session.readOffset = session.bufferStartOffset + 10;
  session._appendOutput("b".repeat(25));
  assert.equal(session.output.length, 50000);
  assert.match(session.output.slice(-25), /^b+$/);
  assert.equal(session.readOffset >= session.bufferStartOffset, true);
});

test("Hermes tools accept the fixed Hermes root cwd", async () => {
  const created = [];
  const manager = new HermesTerminalManager({
    spawnPty: (command, args, options) => {
      const pty = makeFakePty();
      Object.assign(pty, { command, args, options });
      created.push(pty);
      return pty;
    },
  });
  const runHermesCalls = [];
  const server = createMockServer();
  registerHermesTools(server, {
    manager,
    runHermesImpl: async (options) => {
      runHermesCalls.push(options);
      return { ok: true, cwd: options.cwd ?? null };
    },
  });

  const hermesRoot = path.resolve("C:/Users/archa/AppData/Local/hermes");
  await server.tools.get("hermes_run").handler({ args: ["--version"], cwd: hermesRoot });
  await server.tools.get("hermes_session_send").handler({ prompt: "ping", cwd: hermesRoot });
  await server.tools.get("hermes_session_send_batch").handler({ prompts: ["one", "two"], cwd: hermesRoot });
  assert.deepEqual(runHermesCalls.map((call) => call.cwd), [hermesRoot, hermesRoot, hermesRoot, hermesRoot]);

  const toolDescription = server.tools.get("hermes_terminal_start").config.inputSchema.cwd.description;
  assert.match(toolDescription, /WORKSPACE_ROOT/);
  assert.match(toolDescription, /C:\\Users\\archa\\AppData\\Local\\hermes/);

  await server.tools.get("hermes_terminal_start").handler({});
  assert.equal(created[0].options.cwd, WORKSPACE_ROOT);
});

test("Hermes tools reject unrelated external cwd values", async () => {
  const manager = new HermesTerminalManager({ spawnPty: () => makeFakePty() });
  const server = createMockServer();
  registerHermesTools(server, { manager, runHermesImpl: async () => ({ ok: true }) });
  const externalCwd = path.join(path.dirname(WORKSPACE_ROOT), "hermes-external");
  await assert.rejects(() => server.tools.get("hermes_terminal_start").handler({ cwd: externalCwd }), /cwd must be WORKSPACE_ROOT/);
});

test("HermesTerminalManager send returns the retained delta even when rollover happens during send", async () => {
  let emit;
  const manager = new HermesTerminalManager({ spawnPty: () => makeFakePty({ onWrite: () => {} }) });
  const { sessionId, session } = manager.start({ cwd: "." });
  emit = (text) => session._appendOutput(text);
  const sendPromise = manager.send(sessionId, { input: "go", idleMs: 5, timeoutMs: 200 });
  emit("x".repeat(49990));
  emit("y".repeat(100));
  const output = await sendPromise;
  assert.match(output, /y+$/);
  assert.equal(manager.read(sessionId, { maxChars: 100 }), "");
});

test("HermesTerminalManager collects initial output during startup", async () => {
  let onData;
  const manager = new HermesTerminalManager({ spawnPty: () => ({ onData(cb) { onData = cb; }, onExit() {}, write() {}, resize() {}, kill() {} }) });
  const { sessionId, session } = manager.start({ cwd: "." });
  setTimeout(() => onData?.("banner> "), 5);
  await session.settleStartup();
  assert.equal(session.read({ maxChars: 100 }), "banner> ");
  manager.close(sessionId);
});

function makeFakeProcessJobManager(jobs = []) {
  const starts = [];
  return {
    jobs,
    starts,
    list() { return jobs; },
    start(options) {
      starts.push(options);
      return { jobId: "job_test", state: "queued", args: options.publicArgs, metadata: options.metadata, cwd: options.cwd };
    },
  };
}

test("hermes_job_start registers, starts immediately, and redacts the prompt", async () => {
  const processJobManager = makeFakeProcessJobManager();
  const server = createMockServer();
  registerHermesTools(server, { processJobManager });
  assert.equal(server.tools.has("hermes_job_start"), true);

  const response = await server.tools.get("hermes_job_start").handler({
    prompt: "private prompt",
    session_name: "session-one",
    cwd: WORKSPACE_ROOT,
    timeout_ms: 123,
  });
  const job = JSON.parse(response.content[0].text);
  const start = processJobManager.starts[0];

  assert.equal(job.state, "queued");
  assert.match(start.command, /hermes(\.exe)?$/i);
  assert.deepEqual(start.args, ["chat", "-Q", "--continue", "session-one", "--query", "private prompt"]);
  assert.deepEqual(start.publicArgs, ["chat", "-Q", "--continue", "session-one", "--query", "[redacted]"]);
  assert.equal(start.publicArgs.includes("private prompt"), false);
  assert.equal(start.cwd, WORKSPACE_ROOT);
  assert.equal(start.timeoutMs, 123);
  assert.deepEqual(start.metadata, { kind: "hermes_session_job", session_name: "session-one", prompt_chars: 14 });
});

test("hermes_job_start validates Hermes cwd", async () => {
  const processJobManager = makeFakeProcessJobManager();
  const server = createMockServer();
  registerHermesTools(server, { processJobManager });

  await assert.rejects(
    () => server.tools.get("hermes_job_start").handler({ prompt: "hello", session_name: "session-one", cwd: path.resolve("C:/outside-space-rocks") }),
    /cwd must be WORKSPACE_ROOT/
  );
  assert.equal(processJobManager.starts.length, 0);
});

test("hermes_job_start rejects an active session job and allows terminal jobs", async () => {
  const jobs = [{ state: "running", metadata: { kind: "hermes_session_job", session_name: "session-one" } }];
  const processJobManager = makeFakeProcessJobManager(jobs);
  const server = createMockServer();
  registerHermesTools(server, { processJobManager });
  const handler = server.tools.get("hermes_job_start").handler;

  await assert.rejects(
    () => handler({ prompt: "second", session_name: "session-one", cwd: WORKSPACE_ROOT }),
    /already queued or running/
  );
  assert.equal(processJobManager.starts.length, 0);

  jobs[0].state = "succeeded";
  await handler({ prompt: "after completion", session_name: "session-one", cwd: WORKSPACE_ROOT });
  assert.equal(processJobManager.starts.length, 1);
});
