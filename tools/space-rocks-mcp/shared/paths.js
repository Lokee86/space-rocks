import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export const REPO_ROOT = path.resolve(
  process.env.SPACE_ROCKS_REPO ?? path.join(__dirname, "../../..")
);

export function repoPath(relativePath = ".") {
  const resolved = path.resolve(REPO_ROOT, relativePath);

  if (resolved !== REPO_ROOT && !resolved.startsWith(REPO_ROOT + path.sep)) {
    throw new Error("Path escapes Space Rocks repo root");
  }

  return resolved;
}

export function repoRelative(absolutePath) {
  return path.relative(REPO_ROOT, absolutePath).replaceAll("\\", "/");
}
