import assert from "node:assert/strict";
import { test } from "node:test";

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

test("registerHermesTools exposes the new PTY tools", () => {
  const server = createMockServer();
  registerHermesTools(server, { manager: new HermesTerminalManager({ spawnPty: () => { throw new Error("not used"); } }) });
  for (const name of ["hermes_terminal_start", "hermes_terminal_send", "hermes_terminal_read", "hermes_terminal_resize", "hermes_terminal_close"]) assert.ok(server.tools.has(name), `${name} should be registered`);
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
  const { sessionId, session } = manager.start({ cwd: ".", cols: 90, rows: 20 });
  assert.match(sessionId, /^ht_/);
  assert.match(created[0].command, /hermes(\.exe)?$/i);
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
