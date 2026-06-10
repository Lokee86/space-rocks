import { promises as fs } from "node:fs";
import path from "node:path";

import { repoRelative } from "./paths.js";

const IGNORED_DIRS = new Set([
  ".git",
  "node_modules",
  ".godot",
  ".import",
  "dist",
  "build",
  "target",
  "tmp",
]);

const TEXT_EXTENSIONS = new Set([
  ".go",
  ".gd",
  ".ts",
  ".js",
  ".mjs",
  ".py",
  ".pyi",

  // Ruby / Rails
  ".rb",
  ".erb",
  ".rake",
  ".ru",
  ".gemspec",
  ".builder",

  // Data / config
  ".json",
  ".toml",
  ".yaml",
  ".yml",
  ".xml",
  ".csv",

  // Docs / text
  ".md",
  ".txt",

  // Godot
  ".tscn",
  ".tres",
  ".cfg",

  // Extensionless or special text files
  ".gitignore",
  "Gemfile",
  "Gemfile.lock",
  "Rakefile",
  "config.ru",
  ".ruby-version",

  // Go module files
  ".mod",
  ".sum",
]);

export function isProbablyTextFile(filePath) {
  const base = path.basename(filePath);
  const ext = path.extname(filePath);
  return TEXT_EXTENSIONS.has(ext) || TEXT_EXTENSIONS.has(base);
}

export async function walkDirectory(root, maxFiles) {
  const results = [];

  async function walk(current) {
    if (results.length >= maxFiles) return;

    const entries = await fs.readdir(current, { withFileTypes: true });

    for (const entry of entries) {
      if (results.length >= maxFiles) return;

      if (entry.name.startsWith(".") && entry.name !== ".gitignore" && entry.name !== ".github") {
        continue;
      }

      const fullPath = path.join(current, entry.name);
      const relPath = repoRelative(fullPath);

      if (entry.isDirectory()) {
        if (IGNORED_DIRS.has(entry.name)) continue;

        results.push(`${relPath}/`);
        await walk(fullPath);
        continue;
      }

      if (entry.isFile()) {
        results.push(relPath);
      }
    }
  }

  await walk(root);
  return results;
}

export async function searchText(root, query, maxFiles, maxMatches) {
  const matches = [];
  let scanned = 0;
  const lowerQuery = query.toLowerCase();

  async function walk(current) {
    if (scanned >= maxFiles || matches.length >= maxMatches) return;

    const entries = await fs.readdir(current, { withFileTypes: true });

    for (const entry of entries) {
      if (scanned >= maxFiles || matches.length >= maxMatches) return;

      const fullPath = path.join(current, entry.name);

      if (entry.isDirectory()) {
        if (IGNORED_DIRS.has(entry.name)) continue;

        await walk(fullPath);
        continue;
      }

      if (!entry.isFile() || !isProbablyTextFile(fullPath)) continue;

      scanned++;

      let text;
      try {
        text = await fs.readFile(fullPath, "utf8");
      } catch {
        continue;
      }

      const lines = text.split(/\r?\n/);

      for (let i = 0; i < lines.length; i++) {
        if (lines[i].toLowerCase().includes(lowerQuery)) {
          matches.push({
            file: repoRelative(fullPath),
            line: i + 1,
            text: lines[i].slice(0, 300),
          });

          if (matches.length >= maxMatches) break;
        }
      }
    }
  }

  await walk(root);

  return {
    scanned_files: scanned,
    matches,
  };
}
