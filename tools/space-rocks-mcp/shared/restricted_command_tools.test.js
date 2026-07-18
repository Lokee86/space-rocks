import assert from "node:assert/strict";
import { test } from "node:test";

import { registerRestrictedCommandTools } from "./restricted_command_tools.js";

function createMockServer() {
  const tools = new Map();
  return {
    tools,
    registerTool(name, config, handler) {
      tools.set(name, { config, handler });
    },
  };
}

function makeCommandService() {
  const starts = [];
  return {
    starts,
    listCommands() { return ["node", "ls"]; },
    async start(options) {
      starts.push(options);
      return { jobId: "job_command", state: "queued", args: options.publicArgs ?? options.args, metadata: options.metadata };
    },
  };
}

function parseResponse(response) {
  return JSON.parse(response.content[0].text);
}

test("registers restricted command tools with fixed command and timeout schemas", () => {
  const server = createMockServer();
  registerRestrictedCommandTools(server, { commandService: makeCommandService() });

  assert.deepEqual([...server.tools.keys()], ["list_workspace_commands", "command_job_start"]);
  const schema = server.tools.get("command_job_start").config.inputSchema;
  assert.ok(schema.command.options.includes("node"));
  assert.equal(schema.args.def.defaultValue.length, 0);
  assert.equal(schema.timeout_ms.def.defaultValue, 300000);
  assert.equal(schema.timeout_ms.safeParse(999).success, false);
  assert.equal(schema.timeout_ms.safeParse(1000).success, true);
  assert.equal(schema.timeout_ms.safeParse(3600000).success, true);
  assert.equal(schema.timeout_ms.safeParse(3600001).success, false);
  assert.equal(schema.command.safeParse("git").success, false);
});

test("returns command list JSON and forwards async job arguments", async () => {
  const commandService = makeCommandService();
  const server = createMockServer();
  registerRestrictedCommandTools(server, { commandService });

  const listResponse = await server.tools.get("list_workspace_commands").handler({});
  assert.deepEqual(parseResponse(listResponse), { commands: ["node", "ls"] });
  assert.match(listResponse.content[0].text, /\n/);

  const jobResponse = await server.tools.get("command_job_start").handler({
    command: "node",
    args: ["--version"],
    cwd: ".worktrees/demo",
    timeout_ms: 1234,
    env: { TEST_VALUE: "yes" },
  });
  assert.deepEqual(commandService.starts[0], {
    command: "node",
    args: ["--version"],
    cwd: ".worktrees/demo",
    timeoutMs: 1234,
    env: { TEST_VALUE: "yes" },
  });
  assert.deepEqual(parseResponse(jobResponse), {
    jobId: "job_command",
    state: "queued",
    args: ["--version"],
  });
});

test("uses schema defaults when starting a command job", async () => {
  const commandService = makeCommandService();
  const server = createMockServer();
  registerRestrictedCommandTools(server, { commandService });

  await server.tools.get("command_job_start").handler({ command: "ls" });
  assert.deepEqual(commandService.starts[0], {
    command: "ls",
    args: [],
    cwd: undefined,
    timeoutMs: 300000,
    env: undefined,
  });
});

test("propagates restricted command policy errors", async () => {
  const server = createMockServer();
  registerRestrictedCommandTools(server, {
    commandService: {
      listCommands: () => ["node"],
      async start() {
        throw new Error("node eval and print modes are not allowed");
      },
    },
  });

  await assert.rejects(
    server.tools.get("command_job_start").handler({ command: "node", args: ["-eCODE"] }),
    /node eval and print modes are not allowed/
  );
});
