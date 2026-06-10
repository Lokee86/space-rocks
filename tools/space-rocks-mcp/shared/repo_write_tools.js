import { promises as fs } from "node:fs";
import path from "node:path";

import { z } from "zod";

import { repoPath, repoRelative } from "./paths.js";
import { textResponse } from "./responses.js";
import { isProbablyTextFile } from "./text_files.js";
import { listAllowedCommands, runAllowedCommand } from "./allowed_commands.js";

async function fileExists(filePath) {
  try {
    await fs.access(filePath);
    return true;
  } catch {
    return false;
  }
}

function assertWritableTextFile(filePath) {
  if (!isProbablyTextFile(filePath)) {
    throw new Error("Refusing to write non-text or unsupported file type");
  }
}

export function registerRepoWriteTools(server) {
  server.registerTool(
    "ping",
    {
      title: "Ping",
      description: "Simple connection test for the Space Rocks write MCP server.",
      inputSchema: {
        message: z.string().optional(),
      },
    },
    async ({ message }) => {
      return textResponse(`Write MCP server is reachable. Message: ${message ?? "none"}`);
    }
  );

  server.registerTool(
    "write_repo_file",
    {
      title: "Write repo file",
      description: "Write a UTF-8 text file inside the Space Rocks repo.",
      inputSchema: {
        path: z.string(),
        text: z.string(),
        overwrite: z.boolean().optional(),
      },
    },
    async ({ path: requestedPath, text, overwrite = false }) => {
      const filePath = repoPath(requestedPath);
      assertWritableTextFile(filePath);

      if (!overwrite && await fileExists(filePath)) {
        throw new Error("File already exists. Set overwrite=true to replace it.");
      }

      await fs.mkdir(path.dirname(filePath), { recursive: true });
      await fs.writeFile(filePath, text, "utf8");

      return textResponse(`Wrote ${repoRelative(filePath)}`);
    }
  );

  server.registerTool(
    "replace_in_repo_file",
    {
      title: "Replace in repo file",
      description: "Replace exactly one text occurrence in a repo file.",
      inputSchema: {
        path: z.string(),
        expected: z.string(),
        replacement: z.string(),
      },
    },
    async ({ path: requestedPath, expected, replacement }) => {
      const filePath = repoPath(requestedPath);
      assertWritableTextFile(filePath);

      const text = await fs.readFile(filePath, "utf8");
      const count = text.split(expected).length - 1;

      if (count === 0) {
        throw new Error("Expected text was not found.");
      }

      if (count > 1) {
        throw new Error(`Expected text appears ${count} times. Refusing ambiguous replacement.`);
      }

      await fs.writeFile(filePath, text.replace(expected, replacement), "utf8");

      return textResponse(`Updated ${repoRelative(filePath)}`);
    }
  );

  server.registerTool(
    "list_allowed_commands",
    {
      title: "List allowed commands",
      description: "List allowlisted repo commands available to the write MCP server.",
      inputSchema: {},
    },
    async () => {
      return textResponse(JSON.stringify(listAllowedCommands(), null, 2));
    }
  );

  server.registerTool(
    "run_allowed_command",
    {
      title: "Run allowed command",
      description: "Run one allowlisted repo command by name.",
      inputSchema: {
        name: z.enum(listAllowedCommands()),
      },
    },
    async ({ name }) => {
      const result = await runAllowedCommand(name);
      return textResponse(JSON.stringify(result, null, 2));
    }
  );
}
