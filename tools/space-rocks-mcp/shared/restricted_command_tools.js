import { z } from "zod";

import { defaultRestrictedCommandService, RESTRICTED_COMMAND_REGISTRY } from "./restricted_commands.js";
import { textResponse } from "./responses.js";

const commandSchema = z.enum(Object.keys(RESTRICTED_COMMAND_REGISTRY));

export function registerRestrictedCommandTools(server, { commandService = defaultRestrictedCommandService } = {}) {
  server.registerTool(
    "list_workspace_commands",
    {
      title: "List workspace commands",
      description: "List the fixed restricted commands available as asynchronous workspace jobs.",
      inputSchema: {},
    },
    async () => textResponse(JSON.stringify({ commands: commandService.listCommands() }, null, 2))
  );

  server.registerTool(
    "command_job_start",
    {
      title: "Start workspace command job",
      description: "Start one restricted workspace command asynchronously and return its job snapshot.",
      inputSchema: {
        command: commandSchema,
        args: z.array(z.string()).default([]),
        cwd: z.string().optional(),
        timeout_ms: z.number().int().min(1000).max(3600000).default(300000),
        env: z.record(z.string(), z.string()).optional(),
      },
    },
    async ({ command, args = [], cwd, timeout_ms = 300000, env }) => {
      const snapshot = await commandService.start({ command, args, cwd, timeoutMs: timeout_ms, env });
      return textResponse(JSON.stringify(snapshot, null, 2));
    }
  );
}
