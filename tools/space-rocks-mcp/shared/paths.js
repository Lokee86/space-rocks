import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const PACKAGE_REPO_ROOT = path.join(__dirname, "../../..");

export const SPACE_ROCKS_REPO_ROOT = path.resolve(
  process.env.SPACE_ROCKS_REPO ?? PACKAGE_REPO_ROOT
);
export const REPO_ROOT = SPACE_ROCKS_REPO_ROOT;
export const WORKSPACE_ROOT = path.resolve(
  process.env.WORKSPACE_ROOT ?? SPACE_ROCKS_REPO_ROOT
);

function assertWithinRoot(root, resolved, message) {
  if (resolved !== root && !resolved.startsWith(root + path.sep)) {
    throw new Error(message);
  }
}

export function repoPath(relativePath = ".") {
  const resolved = path.resolve(SPACE_ROCKS_REPO_ROOT, relativePath);
  assertWithinRoot(
    SPACE_ROCKS_REPO_ROOT,
    resolved,
    "Path escapes Space Rocks repo root"
  );

  return resolved;
}

export function repoRelative(absolutePath) {
  return path.relative(SPACE_ROCKS_REPO_ROOT, absolutePath).replaceAll("\\", "/");
}

export function workspacePath(relativePath = ".") {
  const resolved = path.resolve(WORKSPACE_ROOT, relativePath);
  assertWithinRoot(WORKSPACE_ROOT, resolved, "Path escapes workspace root");

  return resolved;
}

export function workspaceRelative(absolutePath) {
  const resolved = path.resolve(absolutePath);
  assertWithinRoot(WORKSPACE_ROOT, resolved, "Path escapes workspace root");

  return path.relative(WORKSPACE_ROOT, resolved).replaceAll("\\", "/");
}
