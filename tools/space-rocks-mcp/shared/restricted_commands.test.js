import assert from "node:assert/strict";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";
import { test } from "node:test";

import { ProcessJobManager } from "./job_manager.js";
import { COMMAND_REGISTRY, RestrictedCommandService } from "./restricted_commands.js";

function makeManager() {
  const starts = [];
  return {
    starts,
    start(options) {
      starts.push(options);
      return { jobId: `job_${starts.length}`, state: "queued", command: options.command, args: options.publicArgs, cwd: options.cwd, metadata: options.metadata };
    },
  };
}

async function makeRoot(t) {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "restricted-commands-"));
  t.after(() => fs.rm(root, { recursive: true, force: true }));
  return root;
}

test("lists exactly the fixed restricted command registry", () => {
  const expected = ["go", "gofmt", "python", "pytest", "ruby", "bundle", "rails", "npm", "node", "godot", "rg", "grep", "find", "ls", "cat", "sed", "head", "tail", "wc", "diff"];
  assert.deepEqual(Object.keys(COMMAND_REGISTRY), expected);
  assert.deepEqual(new RestrictedCommandService({ processJobManager: makeManager() }).listCommands(), expected);
  for (const blocked of ["git", "cmd", "powershell", "pwsh", "bash", "sh", "wsl", "npx"]) assert.equal(Object.hasOwn(COMMAND_REGISTRY, blocked), false);
});

test("uses default root cwd, permits worktrees, and rejects lexical escapes", async (t) => {
  const root = await makeRoot(t);
  await fs.mkdir(path.join(root, ".worktrees", "demo"), { recursive: true });
  const manager = makeManager();
  const service = new RestrictedCommandService({ root, processJobManager: manager });
  const realRoot = await fs.realpath(root);

  await service.start({ command: "ls" });
  await service.start({ command: "ls", cwd: ".worktrees/demo" });
  assert.equal(manager.starts[0].cwd, realRoot);
  assert.equal(manager.starts[1].cwd, await fs.realpath(path.join(root, ".worktrees", "demo")));
  await assert.rejects(service.start({ command: "ls", cwd: "../outside" }), /escapes workspace root/);
});

test("rejects a cwd symlink that resolves outside the workspace", async (t) => {
  const root = await makeRoot(t);
  const outside = await fs.mkdtemp(path.join(os.tmpdir(), "restricted-outside-"));
  t.after(() => fs.rm(outside, { recursive: true, force: true }));
  const link = path.join(root, "linked");
  try {
    await fs.symlink(outside, link, "junction");
  } catch (error) {
    t.skip(`symlink setup unavailable: ${error.message}`);
    return;
  }

  const service = new RestrictedCommandService({ root, processJobManager: makeManager() });
  await assert.rejects(service.start({ command: "ls", cwd: "linked" }), /resolves outside workspace root/);
});

test("rejects restricted command policies and allows narrow exceptions", async (t) => {
  const root = await makeRoot(t);
  const service = new RestrictedCommandService({ root, processJobManager: makeManager() });
  const cases = [
    ["node", ["-e", "code"]], ["node", ["-eCODE"]], ["node", ["--eval=code"]], ["node", ["-p", "code"]], ["node", ["-pCODE"]], ["node", ["--print", "code"]],
    ["python", ["-c", "code"]], ["python", ["-cCODE"]], ["python", ["-m", "other"]], ["python", ["-mother"]], ["ruby", ["-e", "code"]], ["ruby", ["-eCODE"]],
    ["npm", ["exec", "tool"]], ["npm", ["x", "tool"]], ["rails", ["runner", "code"]], ["rails", ["console"]],
    ["go", ["env", "-w", "KEY=value"]], ["go", ["-C", "repo", "env", "-w", "KEY=value"]], ["go", ["clean", "-modcache"]], ["go", ["-C", "repo", "clean", "-modcache"]], ["bundle", ["exec", "go", "build"]],
    ["bundle", ["config", "--global", "x"]], ["bundle", ["config", "--system", "x"]],
  ];
  for (const [command, args] of cases) await assert.rejects(service.start({ command, args }), undefined, `${command} ${args.join(" ")}`);
  await service.start({ command: "node", args: ["--profile"] });
  await service.start({ command: "python", args: ["--version"] });
  await service.start({ command: "ruby", args: ["-w"] });
  await service.start({ command: "python", args: ["-m", "pytest"] });
  for (const nested of ["rails", "rake", "ruby", "rspec"]) await service.start({ command: "bundle", args: ["exec", nested] });
});

test("validates command arguments and environment overrides", async (t) => {
  const root = await makeRoot(t);
  const manager = makeManager();
  const service = new RestrictedCommandService({ root, processJobManager: manager, processEnvironment: { BASE_VALUE: "base" } });

  await service.start({ command: "ls", args: ["--name"], env: { OK_VALUE: "yes", OTHER_2: "two" } });
  assert.equal(manager.starts.at(-1).env.OK_VALUE, "yes");
  assert.equal(manager.starts.at(-1).env.BASE_VALUE, "base");
  for (const name of ["PATH", "PATHEXT", "NODE_OPTIONS", "PYTHONPATH", "RUBYOPT", "BUNDLE_GEMFILE", "GEM_HOME", "GEM_PATH", "COMSPEC", "SHELL", "GIT_CONFIG", "SSH_AUTH_SOCK", "lowercase"]) {
    await assert.rejects(service.start({ command: "ls", env: { [name]: "blocked" } }), /not allowed/);
  }
  await assert.rejects(service.start({ command: "ls", args: ["x\0y"] }), /NUL/);
  await assert.rejects(service.start({ command: "ls", args: Array.from({ length: 201 }, () => "x") }), /at most 200/);
  await assert.rejects(service.start({ command: "ls", args: ["x".repeat(4097)] }), /at most 4096/);
  await assert.rejects(service.start({ command: "ls", env: Object.fromEntries(Array.from({ length: 51 }, (_, index) => [`KEY_${index}`, "x"])) }), /more than 50/);
  await assert.rejects(service.start({ command: "ls", env: { VALUE: "x".repeat(4097) } }), /4096/);
});

test("resolves executables, attaches metadata, and redacts sensitive public args", async (t) => {
  const root = await makeRoot(t);
  const manager = makeManager();
  const resolved = [];
  const service = new RestrictedCommandService({
    root,
    processJobManager: manager,
    resolveExecutable: (command) => { resolved.push(command); return `resolved-${command}`; },
  });

  const snapshot = await service.start({
    command: "node",
    args: ["--token", "secret", "--api-key=abc", "--safe", "value", "--auth"],
    timeoutMs: 1234,
  });
  const start = manager.starts[0];
  assert.equal(snapshot.jobId, "job_1");
  assert.equal(start.command, "resolved-node");
  assert.deepEqual(start.publicArgs, ["--token", "[redacted]", "--api-key=[redacted]", "--safe", "value", "--auth"]);
  assert.deepEqual(start.metadata, { kind: "workspace_command", command: "node" });
  assert.equal(start.timeoutMs, 1234);
  assert.deepEqual(resolved, ["node"]);
});

test("resolves platform-specific executables", async (t) => {
  const root = await makeRoot(t);
  const manager = makeManager();
  const service = new RestrictedCommandService({ root, platform: "win32", processJobManager: manager, processEnvironment: { GODOT_EXECUTABLE: "custom-godot.exe" } });
  for (const command of ["node", "npm", "bundle", "rails", "godot"]) await service.start({ command });
  assert.equal(manager.starts[0].command, process.execPath);
  assert.equal(manager.starts[1].command, "npm.cmd");
  assert.equal(manager.starts[2].command, "bundle.bat");
  assert.equal(manager.starts[3].command, "rails.bat");
  assert.equal(manager.starts[4].command, "custom-godot.exe");

  const defaultGodotManager = makeManager();
  const defaultGodotService = new RestrictedCommandService({ root, platform: "win32", processJobManager: defaultGodotManager, processEnvironment: {} });
  await defaultGodotService.start({ command: "godot" });
  assert.equal(defaultGodotManager.starts[0].command, "C:\\Godot.exe");
});

test("runs node version through the real process job manager", async (t) => {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "restricted-command-real-"));
  const manager = new ProcessJobManager({
    defaultTimeoutMs: 2000,
    retentionMs: 1000,
    sweepIntervalMs: 100,
    terminationGraceMs: 100,
  });
  t.after(() => {
    manager.closeAll();
    return fs.rm(root, { recursive: true, force: true });
  });

  const service = new RestrictedCommandService({ root, processJobManager: manager });
  const initial = await service.start({ command: "node", args: ["--version"] });
  const terminalStates = new Set(["succeeded", "failed", "cancelled", "timed_out"]);
  let status = initial;
  for (let attempt = 0; attempt < 100 && !terminalStates.has(status.state); attempt += 1) {
    await new Promise((resolve) => setTimeout(resolve, 10));
    status = manager.status(initial.jobId);
  }

  assert.equal(status.state, "succeeded");
  assert.deepEqual(status.metadata, { kind: "workspace_command", command: "node" });
  assert.match(manager.read(initial.jobId, { stream: "stdout" }).data, /^v/);
});
