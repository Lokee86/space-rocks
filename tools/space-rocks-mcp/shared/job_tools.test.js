import assert from "node:assert/strict";
import { test } from "node:test";

import { registerJobTools } from "./job_tools.js";

function createMockServer() {
  const tools = new Map();
  return {
    tools,
    registerTool(name, config, handler) {
      tools.set(name, { config, handler });
    },
  };
}

function readJsonResponse(response) {
  assert.equal(response.content[0].type, "text");
  return JSON.parse(response.content[0].text);
}

test("registerJobTools exposes the process job lifecycle tools", () => {
  const server = createMockServer();
  registerJobTools(server, { manager: { status() {}, read() {}, cancel() {}, list() { return []; } } });

  assert.deepEqual([...server.tools.keys()], ["job_status", "job_read", "job_cancel", "job_list"]);
  assert.equal(server.tools.get("job_status").config.inputSchema.job_id.safeParse("job_valid").success, true);
  assert.equal(server.tools.get("job_status").config.inputSchema.job_id.safeParse("not-a-job").success, false);
  assert.equal(server.tools.get("job_read").config.inputSchema.stream.safeParse("stderr").success, true);
  assert.equal(server.tools.get("job_list").config.inputSchema.state.safeParse("running").success, true);
  assert.equal(server.tools.get("job_list").config.inputSchema.state.safeParse("unknown").success, false);
});

test("job tools delegate lifecycle calls and preserve read cursor fields", async () => {
  const calls = [];
  const jobs = [
    { jobId: "job_1", state: "running" },
    { jobId: "job_2", state: "succeeded" },
  ];
  const manager = {
    status(jobId) {
      calls.push(["status", jobId]);
      return { jobId, state: "running" };
    },
    read(jobId, options) {
      calls.push(["read", jobId, options]);
      return {
        jobId,
        stream: options.stream,
        data: "output",
        cursor: options.cursor,
        nextCursor: 12,
        startOffset: 4,
        endOffset: 12,
        truncated: true,
      };
    },
    cancel(jobId) {
      calls.push(["cancel", jobId]);
      return { jobId, state: "cancelled" };
    },
    list() {
      calls.push(["list"]);
      return jobs;
    },
  };
  const server = createMockServer();
  registerJobTools(server, { manager });

  assert.deepEqual(readJsonResponse(await server.tools.get("job_status").handler({ job_id: "job_1" })), {
    jobId: "job_1",
    state: "running",
  });
  assert.deepEqual(readJsonResponse(await server.tools.get("job_read").handler({
    job_id: "job_1",
    stream: "stderr",
    cursor: 8,
    max_chars: 50,
  })), {
    jobId: "job_1",
    stream: "stderr",
    data: "output",
    cursor: 8,
    nextCursor: 12,
    startOffset: 4,
    endOffset: 12,
    truncated: true,
  });
  assert.deepEqual(readJsonResponse(await server.tools.get("job_cancel").handler({ job_id: "job_1" })), {
    jobId: "job_1",
    state: "cancelled",
  });
  assert.deepEqual(readJsonResponse(await server.tools.get("job_list").handler({ state: "running" })), [jobs[0]]);
  assert.deepEqual(readJsonResponse(await server.tools.get("job_list").handler({})), jobs);
  assert.deepEqual(calls, [
    ["status", "job_1"],
    ["read", "job_1", { stream: "stderr", cursor: 8, maxChars: 50 }],
    ["cancel", "job_1"],
    ["list"],
    ["list"],
  ]);
});
