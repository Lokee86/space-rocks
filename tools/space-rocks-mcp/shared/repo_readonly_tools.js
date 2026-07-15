import { promises as fs } from "node:fs";

import { z } from "zod";

import { WORKSPACE_ROOT, workspacePath } from "./paths.js";
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
    "workspace_root",
    {
      title: "Show workspace root",
      description: "Returns the configured workspace root.",
      inputSchema: {},
    },
    async () => {
      return textResponse(WORKSPACE_ROOT);
    }
  );

  server.registerTool(
    "repo_root",
    {
      title: "Show workspace root (compatibility alias)",
      description: "Compatibility alias for workspace_root; returns the configured workspace root.",
      inputSchema: {},
    },
    async () => {
      return textResponse(WORKSPACE_ROOT);
    }
  );

  server.registerTool(
    "list_repo_tree",
    {
      title: "List workspace tree",
      description: "List files and directories under a workspace-relative path.",
      inputSchema: {
        path: z.string().optional(),
        max_files: z.number().int().min(1).max(2000).optional(),
      },
    },
    async ({ path: requestedPath = ".", max_files = 500 }) => {
      const root = workspacePath(requestedPath);
      const entries = await walkDirectory(root, max_files);
      return textResponse(entries.join("\n"));
    }
  );

  server.registerTool(
    "read_repo_file",
    {
      title: "Read workspace file",
      description: "Read a text file from the workspace by workspace-relative path.",
      inputSchema: {
        path: z.string(),
        max_chars: z.number().int().min(1).max(50000).optional(),
      },
    },
    async ({ path: requestedPath, max_chars = 20000 }) => {
      const filePath = workspacePath(requestedPath);

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
      title: "Search workspace text",
      description: "Search text files in the workspace for a string.",
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
      const root = workspacePath(requestedPath);
      const result = await searchText(root, query, max_files, max_matches);
      return textResponse(JSON.stringify(result, null, 2));
    }
  );
}
