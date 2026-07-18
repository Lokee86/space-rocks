import assert from "node:assert/strict";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";
import { test } from "node:test";

import { WorkspaceWriteService } from "./workspace_writes.js";
import { registerRepoWriteTools } from "./repo_write_tools.js";

function createMockServer() {
  const tools = new Map();
  return {
    tools,
    registerTool(name, config, handler) {
      tools.set(name, { config, handler });
    },
  };
}

function parseResponse(response) {
  return JSON.parse(response.content[0].text);
}

async function makeRoot(t) {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "repo-write-tools-"));
  t.after(() => fs.rm(root, { recursive: true, force: true }));
  return root;
}

test("registers mutation tools and forwards single writes, replacements, and batches", async () => {
  const calls = [];
  const workspaceWriteService = {
    async applyBatch(edits) {
      calls.push(edits);
      return { changedFiles: edits.map((edit) => edit.path).filter((path, index, all) => all.indexOf(path) === index) };
    },
  };
  const server = createMockServer();
  registerRepoWriteTools(server, { workspaceWriteService });

  for (const name of ["write_repo_file", "replace_in_repo_file", "apply_repo_file_edits", "list_allowed_commands", "run_allowed_command"]) {
    assert.ok(server.tools.has(name), `${name} should be registered`);
  }
  assert.match(server.tools.get("write_repo_file").config.description, /configured workspace/);
  assert.match(server.tools.get("replace_in_repo_file").config.description, /configured workspace/);

  const writeResult = await server.tools.get("write_repo_file").handler({ path: "src/file.js", text: "one" });
  assert.deepEqual(calls[0], [{ type: "write", path: "src/file.js", text: "one", overwrite: false }]);
  assert.deepEqual(parseResponse(writeResult), { changed_files: ["src/file.js"] });

  const replaceResult = await server.tools.get("replace_in_repo_file").handler({ path: "src/file.js", expected: "one", replacement: "two" });
  assert.deepEqual(calls[1], [{ type: "replace", path: "src/file.js", expected: "one", replacement: "two" }]);
  assert.deepEqual(parseResponse(replaceResult), { changed_files: ["src/file.js"] });

  const edits = [
    { type: "write", path: "a.js", text: "one", overwrite: true },
    { type: "replace", path: "a.js", expected: "one", replacement: "two" },
    { type: "write", path: "b.js", text: "three" },
  ];
  const batchResult = await server.tools.get("apply_repo_file_edits").handler({ edits });
  assert.deepEqual(calls[2], edits);
  assert.deepEqual(parseResponse(batchResult), { changed_files: ["a.js", "b.js"] });
});

test("preserves default overwrite rejection through the write tool", async (t) => {
  const root = await makeRoot(t);
  await fs.writeFile(path.join(root, "existing.txt"), "old", "utf8");
  const server = createMockServer();
  registerRepoWriteTools(server, { workspaceWriteService: new WorkspaceWriteService({ root }) });

  await assert.rejects(
    server.tools.get("write_repo_file").handler({ path: "existing.txt", text: "new" }),
    /overwrite=true/
  );
  const response = await server.tools.get("write_repo_file").handler({ path: "existing.txt", text: "new", overwrite: true });
  assert.deepEqual(parseResponse(response), { changed_files: ["existing.txt"] });
  assert.equal(await fs.readFile(path.join(root, "existing.txt"), "utf8"), "new");
});

test("propagates WorkspaceWriteService errors", async () => {
  const server = createMockServer();
  registerRepoWriteTools(server, {
    workspaceWriteService: {
      async applyBatch() {
        throw new Error("workspace service failed");
      },
    },
  });

  await assert.rejects(
    server.tools.get("apply_repo_file_edits").handler({ edits: [{ type: "write", path: "file.txt", text: "x" }] }),
    /workspace service failed/
  );
});
