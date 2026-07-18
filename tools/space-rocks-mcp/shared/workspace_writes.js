import { promises as fs } from "node:fs";
import path from "node:path";

import { WORKSPACE_ROOT } from "./paths.js";
import { isProbablyTextFile } from "./text_files.js";

const MAX_BATCH_EDITS = 100;
const REPLACEABLE_RENAME_ERRORS = new Set(["EEXIST", "ENOTEMPTY", "EPERM"]);

function errorMessage(error) {
  return error instanceof Error ? error.message : String(error);
}

function isMissing(error) {
  return error?.code === "ENOENT";
}

export class WorkspaceWriteService {
  constructor({ root = WORKSPACE_ROOT, fileSystem = fs } = {}) {
    this.root = path.resolve(root);
    this.fileSystem = fileSystem;
  }

  apply(edits) {
    return this.applyBatch(edits);
  }

  async applyBatch(edits) {
    const plans = await this._preflight(edits);
    const committed = [];

    try {
      for (const plan of plans) {
        await this._assertNoSymlinkComponents(plan.target);
        committed.push(plan);
        await this._commitContent(plan.target, plan.finalText, plan.mode);
      }
    } catch (error) {
      const rollbackFailures = await this._rollback(committed);
      const rollbackMessage = rollbackFailures.length > 0
        ? ` Rollback failures: ${rollbackFailures.join("; ")}`
        : "";
      throw new Error(`Workspace write commit failed: ${errorMessage(error)}${rollbackMessage}`);
    }

    return { changedFiles: plans.map((plan) => plan.relativePath) };
  }

  async _preflight(edits) {
    if (!Array.isArray(edits) || edits.length === 0) {
      throw new Error("Workspace write batch must contain at least one edit");
    }
    if (edits.length > MAX_BATCH_EDITS) {
      throw new Error(`Workspace write batch cannot contain more than ${MAX_BATCH_EDITS} edits`);
    }

    const plansByTarget = new Map();
    const plans = [];
    for (let index = 0; index < edits.length; index += 1) {
      const edit = edits[index];
      try {
        const type = edit?.type ?? edit?.op;
        if (type !== "write" && type !== "replace") {
          throw new Error("edit type must be write or replace");
        }

        const target = await this._resolveTarget(edit.path);
        const key = process.platform === "win32" ? target.toLowerCase() : target;
        let plan = plansByTarget.get(key);
        if (!plan) {
          const original = await this._readOriginal(target);
          plan = {
            target,
            relativePath: this._relativePath(target),
            exists: original.exists,
            originalText: original.text,
            mode: original.mode,
            finalText: original.text,
            producedInBatch: false,
          };
          plansByTarget.set(key, plan);
          plans.push(plan);
        }

        if (type === "write") {
          if (typeof edit.text !== "string") throw new Error("write edit text must be a string");
          if (plan.exists && !plan.producedInBatch && edit.overwrite !== true) {
            throw new Error("File already exists. Set overwrite=true to replace it");
          }
          plan.finalText = edit.text;
          plan.producedInBatch = true;
          continue;
        }

        if (typeof edit.expected !== "string" || edit.expected.length === 0) {
          throw new Error("replace edit expected must be a non-empty string");
        }
        if (typeof edit.replacement !== "string") throw new Error("replace edit replacement must be a string");
        if (plan.finalText === null) throw new Error("replace edit target does not exist");

        const first = plan.finalText.indexOf(edit.expected);
        if (first < 0) throw new Error("Expected text was not found");
        if (plan.finalText.indexOf(edit.expected, first + edit.expected.length) >= 0) {
          throw new Error("Expected text must occur exactly once");
        }
        plan.finalText = `${plan.finalText.slice(0, first)}${edit.replacement}${plan.finalText.slice(first + edit.expected.length)}`;
        plan.producedInBatch = true;
      } catch (error) {
        throw new Error(`Invalid workspace edit ${index}: ${errorMessage(error)}`);
      }
    }

    return plans;
  }

  async _resolveTarget(requestedPath) {
    if (typeof requestedPath !== "string" || requestedPath.length === 0) {
      throw new Error("edit path must be a non-empty string");
    }

    const target = path.resolve(this.root, requestedPath);
    const relative = path.relative(this.root, target);
    if (!relative || relative === ".." || relative.startsWith(`..${path.sep}`) || path.isAbsolute(relative)) {
      throw new Error("Path escapes workspace root");
    }

    const components = relative.split(path.sep);
    if (components.some((component) => component.toLowerCase() === ".git")) {
      throw new Error("Path cannot contain a .git component");
    }
    if (!isProbablyTextFile(target)) {
      throw new Error("Refusing to write non-text or unsupported file type");
    }

    await this._assertNoSymlinkComponents(target);
    return target;
  }

  async _assertNoSymlinkComponents(target) {
    const relative = path.relative(this.root, target);
    let current = this.root;
    for (const component of relative.split(path.sep)) {
      if (!component || component === ".") continue;
      current = path.join(current, component);
      try {
        const stat = await this.fileSystem.lstat(current);
        if (stat.isSymbolicLink()) throw new Error("Path contains an existing symlink component");
      } catch (error) {
        if (isMissing(error)) return;
        throw error;
      }
    }
  }

  async _readOriginal(target) {
    try {
      const stat = await this.fileSystem.lstat(target);
      if (!stat.isFile()) throw new Error("Target must be a regular file");
      return {
        exists: true,
        text: await this.fileSystem.readFile(target, "utf8"),
        mode: stat.mode & 0o7777,
      };
    } catch (error) {
      if (isMissing(error)) return { exists: false, text: null, mode: null };
      throw error;
    }
  }

  _relativePath(target) {
    return path.relative(this.root, target).replaceAll("\\", "/");
  }

  async _commitContent(target, text, mode) {
    const directory = path.dirname(target);
    await this.fileSystem.mkdir(directory, { recursive: true });
    const temporaryDirectory = await this.fileSystem.mkdtemp(path.join(directory, ".mcp-write-"));
    const staged = path.join(temporaryDirectory, "content");
    try {
      await this.fileSystem.writeFile(staged, text, "utf8");
      if (mode !== null) await this.fileSystem.chmod(staged, mode);
      await this._renameReplacing(staged, target);
    } finally {
      await this.fileSystem.rm(temporaryDirectory, { recursive: true, force: true }).catch(() => {});
    }
  }

  async _renameReplacing(staged, target) {
    try {
      await this.fileSystem.rename(staged, target);
    } catch (error) {
      if (!REPLACEABLE_RENAME_ERRORS.has(error?.code)) throw error;
      await this.fileSystem.rm(target, { force: true });
      await this.fileSystem.rename(staged, target);
    }
  }

  async _rollback(committed) {
    const failures = [];
    for (const plan of committed.toReversed()) {
      try {
        await this._assertNoSymlinkComponents(plan.target);
        if (plan.exists) await this._commitContent(plan.target, plan.originalText, plan.mode);
        else await this.fileSystem.rm(plan.target, { force: true });
      } catch (error) {
        failures.push(`${plan.relativePath}: ${errorMessage(error)}`);
      }
    }
    return failures;
  }
}

export const defaultWorkspaceWriteService = new WorkspaceWriteService();
