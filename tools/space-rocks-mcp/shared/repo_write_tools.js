import { z } from "zod";

import { textResponse } from "./responses.js";
import { listAllowedCommands, runAllowedCommand } from "./allowed_commands.js";
import { defaultWorkspaceWriteService } from "./workspace_writes.js";

function changedFilesResponse(result) {
  return textResponse(JSON.stringify({ changed_files: result.changedFiles }));
}

export function registerRepoWriteTools(server, { workspaceWriteService = defaultWorkspaceWriteService } = {}) {
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
      title: "Write workspace file",
      description: "Write a UTF-8 text file inside the configured workspace.",
      inputSchema: {
        path: z.string(),
        text: z.string(),
        overwrite: z.boolean().optional(),
      },
    },
    async ({ path: requestedPath, text, overwrite = false }) => {
      const result = await workspaceWriteService.applyBatch([{ type: "write", path: requestedPath, text, overwrite }]);
      return changedFilesResponse(result);
    }
  );

  server.registerTool(
    "replace_in_repo_file",
    {
      title: "Replace in workspace file",
      description: "Replace exactly one text occurrence in a configured workspace file.",
      inputSchema: {
        path: z.string(),
        expected: z.string(),
        replacement: z.string(),
      },
    },
    async ({ path: requestedPath, expected, replacement }) => {
      const result = await workspaceWriteService.applyBatch([{ type: "replace", path: requestedPath, expected, replacement }]);
      return changedFilesResponse(result);
    }
  );

  server.registerTool(
    "apply_repo_file_edits",
    {
      title: "Apply workspace file edits",
      description: "Apply a transactional batch of text writes and exact replacements inside the configured workspace.",
      inputSchema: {
        edits: z.array(z.discriminatedUnion("type", [
          z.object({
            type: z.literal("write"),
            path: z.string(),
            text: z.string(),
            overwrite: z.boolean().optional(),
          }),
          z.object({
            type: z.literal("replace"),
            path: z.string(),
            expected: z.string(),
            replacement: z.string(),
          }),
        ])).min(1).max(100),
      },
    },
    async ({ edits }) => {
      const result = await workspaceWriteService.applyBatch(edits);
      return changedFilesResponse(result);
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
