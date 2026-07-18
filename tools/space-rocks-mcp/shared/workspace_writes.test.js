import assert from "node:assert/strict";
import { promises as fs } from "node:fs";
import os from "node:os";
import path from "node:path";
import { test } from "node:test";

import { WorkspaceWriteService } from "./workspace_writes.js";

async function makeRoot(t) {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "workspace-write-"));
  t.after(() => fs.rm(root, { recursive: true, force: true }));
  return root;
}

async function write(root, relativePath, text) {
  const target = path.join(root, relativePath);
  await fs.mkdir(path.dirname(target), { recursive: true });
  await fs.writeFile(target, text, "utf8");
  return target;
}

test("creates, overwrites, and reports workspace-relative changed paths", async (t) => {
  const root = await makeRoot(t);
  const service = new WorkspaceWriteService({ root });

  const created = await service.applyBatch([{ type: "write", path: ".worktrees/demo.js", text: "one" }]);
  assert.deepEqual(created, { changedFiles: [".worktrees/demo.js"] });
  assert.equal(await fs.readFile(path.join(root, ".worktrees/demo.js"), "utf8"), "one");

  await assert.rejects(
    service.apply([{ type: "write", path: ".worktrees/demo.js", text: "two" }]),
    /overwrite=true/
  );
  const overwritten = await service.apply([{ type: "write", path: ".worktrees/demo.js", text: "two", overwrite: true }]);
  assert.deepEqual(overwritten.changedFiles, [".worktrees/demo.js"]);
  assert.equal(await fs.readFile(path.join(root, ".worktrees/demo.js"), "utf8"), "two");
});

test("replaces exactly one occurrence", async (t) => {
  const root = await makeRoot(t);
  await write(root, "notes.md", "alpha\nbeta\n");
  const service = new WorkspaceWriteService({ root });

  await service.applyBatch([{ type: "replace", path: "notes.md", expected: "beta", replacement: "gamma" }]);
  assert.equal(await fs.readFile(path.join(root, "notes.md"), "utf8"), "alpha\ngamma\n");

  await assert.rejects(
    service.applyBatch([{ type: "replace", path: "notes.md", expected: "alpha", replacement: "x" }, { type: "replace", path: "notes.md", expected: "alpha", replacement: "y" }]),
    /Expected text was not found|Expected text must occur exactly once/
  );
});

test("applies same-file edits in input order and commits once", async (t) => {
  const root = await makeRoot(t);
  await write(root, "script.js", "one two");
  const service = new WorkspaceWriteService({ root });

  const result = await service.applyBatch([
    { type: "replace", path: "script.js", expected: "one", replacement: "two" },
    { type: "replace", path: "script.js", expected: "two two", replacement: "three" },
    { type: "write", path: "script.js", text: "final" },
    { type: "replace", path: "script.js", expected: "final", replacement: "done" },
  ]);

  assert.deepEqual(result.changedFiles, ["script.js"]);
  assert.equal(await fs.readFile(path.join(root, "script.js"), "utf8"), "done");
});

test("preflights the full batch before changing any target", async (t) => {
  const root = await makeRoot(t);
  await write(root, "first.txt", "original");
  await write(root, "second.txt", "untouched");
  const service = new WorkspaceWriteService({ root });

  await assert.rejects(
    service.applyBatch([
      { type: "write", path: "first.txt", text: "would change", overwrite: true },
      { type: "replace", path: "second.txt", expected: "missing", replacement: "changed" },
    ]),
    /Expected text was not found/
  );
  assert.equal(await fs.readFile(path.join(root, "first.txt"), "utf8"), "original");
  assert.equal(await fs.readFile(path.join(root, "second.txt"), "utf8"), "untouched");
});

test("rejects empty, oversized, escaping, git, and unsupported batches", async (t) => {
  const root = await makeRoot(t);
  const service = new WorkspaceWriteService({ root });

  await assert.rejects(service.applyBatch([]), /at least one edit/);
  await assert.rejects(service.applyBatch(Array.from({ length: 101 }, () => ({ type: "write", path: "x.txt", text: "x" }))), /more than 100/);
  await assert.rejects(service.applyBatch([{ type: "write", path: "../outside.txt", text: "x" }]), /escapes workspace root/);
  await assert.rejects(service.applyBatch([{ type: "write", path: ".git/config", text: "x" }]), /.git component/);
  await assert.rejects(service.applyBatch([{ type: "write", path: "image.png", text: "x" }]), /non-text or unsupported/);
});

test("rejects an existing symlink component", async (t) => {
  const root = await makeRoot(t);
  const realDirectory = path.join(root, "real");
  const linkedDirectory = path.join(root, "linked");
  await fs.mkdir(realDirectory);
  try {
    await fs.symlink(realDirectory, linkedDirectory, "junction");
  } catch (error) {
    t.skip(`symlink setup unavailable: ${error.message}`);
    return;
  }

  const service = new WorkspaceWriteService({ root });
  await assert.rejects(
    service.applyBatch([{ type: "write", path: "linked/file.js", text: "x" }]),
    /symlink component/
  );
});

test("rolls back committed targets after a later commit failure", async (t) => {
  const root = await makeRoot(t);
  await write(root, "first.txt", "first-original");
  await write(root, "second.txt", "second-original");
  const firstTarget = path.join(root, "first.txt");
  const secondTarget = path.join(root, "second.txt");
  const failingFileSystem = {
    ...fs,
    rename: async (source, target) => {
      if (target === secondTarget) throw new Error("injected rename failure");
      return fs.rename(source, target);
    },
  };
  const service = new WorkspaceWriteService({ root, fileSystem: failingFileSystem });

  await assert.rejects(
    service.applyBatch([
      { type: "write", path: "first.txt", text: "first-new", overwrite: true },
      { type: "write", path: "second.txt", text: "second-new", overwrite: true },
    ]),
    /Workspace write commit failed: injected rename failure/
  );
  assert.equal(await fs.readFile(firstTarget, "utf8"), "first-original");
  assert.equal(await fs.readFile(secondTarget, "utf8"), "second-original");
});
