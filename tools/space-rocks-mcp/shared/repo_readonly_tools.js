import { promises as fs } from "node:fs";

import { z } from "zod";

import { REPO_ROOT, repoPath } from "./paths.js";
import { textResponse } from "./responses.js";
import { isProbablyTextFile, walkDirectory, searchText } from "./text_files.js";

export function registerRepoReadonlyTools(server) {
  server.registerTool(
    "ping",
    {
      title: "Ping",
      description: "Simple connection test for the Space Rocks MCP server.",
      inputSchema: {
        message: z.string().optional(),
      },
    },
    async ({ message }) => {
      return textResponse(`MCP server is reachable. Message: ${message ?? "none"}`);
    }
  );

  server.registerTool(
    "repo_root",
    {
      title: "Show repo root",
      description: "Returns the configured Space Rocks repo root.",
      inputSchema: {},
    },
    async () => {
      return textResponse(REPO_ROOT);
    }
  );

  server.registerTool(
    "list_repo_tree",
    {
      title: "List repo tree",
      description: "List files and directories under a repo-relative path.",
      inputSchema: {
        path: z.string().optional(),
        max_files: z.number().int().min(1).max(2000).optional(),
      },
    },
    async ({ path: requestedPath = ".", max_files = 500 }) => {
      const root = repoPath(requestedPath);
      const entries = await walkDirectory(root, max_files);
      return textResponse(entries.join("\n"));
    }
  );

  server.registerTool(
    "read_repo_file",
    {
      title: "Read repo file",
      description: "Read a text file from the Space Rocks repo by repo-relative path.",
      inputSchema: {
        path: z.string(),
        max_chars: z.number().int().min(1).max(50000).optional(),
      },
    },
    async ({ path: requestedPath, max_chars = 20000 }) => {
      const filePath = repoPath(requestedPath);

      if (!isProbablyTextFile(filePath)) {
        throw new Error("Refusing to read non-text or unsupported file type");
      }

      const text = await fs.readFile(filePath, "utf8");

      if (text.length > max_chars) {
        return textResponse(`${text.slice(0, max_chars)}\n\n[TRUNCATED at ${max_chars} chars]`);
      }

      return textResponse(text);
    }
  );

  server.registerTool(
    "search_repo_text",
    {
      title: "Search repo text",
      description: "Search text files in the Space Rocks repo for a string.",
      inputSchema: {
        query: z.string(),
        path: z.string().optional(),
        max_files: z.number().int().min(1).max(1000).optional(),
        max_matches: z.number().int().min(1).max(200).optional(),
      },
    },
    async ({
      query,
      path: requestedPath = ".",
      max_files = 300,
      max_matches = 50,
    }) => {
      const root = repoPath(requestedPath);
      const result = await searchText(root, query, max_files, max_matches);
      return textResponse(JSON.stringify(result, null, 2));
    }
  );
}
