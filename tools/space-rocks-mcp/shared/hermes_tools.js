import { spawn } from "node:child_process";
import { z } from "zod";
import { REPO_ROOT, repoPath } from "./paths.js";
import { textResponse } from "./responses.js";

const DEFAULT_TIMEOUT_MS = 600000;
const MAX_OUTPUT_CHARS = 50000;

function truncateOutput(output) {
  if (output.length > MAX_OUTPUT_CHARS) {
    return output.slice(0, MAX_OUTPUT_CHARS) + `\n\n[Output truncated at ${MAX_OUTPUT_CHARS} characters]`;
  }
  return output;
}

async function runHermes({ args = [], stdin, cwd, timeout_ms = DEFAULT_TIMEOUT_MS }) {
  const resolvedCwd = cwd ? repoPath(cwd) : REPO_ROOT;

  return new Promise((resolve) => {
    const child = spawn("hermes", args, {
      cwd: resolvedCwd,
      shell: false,
    });

    let stdout = "";
    let stderr = "";
    let timedOut = false;

    const timeout = setTimeout(() => {
      timedOut = true;
      child.kill();
    }, timeout_ms);

    child.stdout.on("data", (data) => {
      stdout += data.toString();
    });

    child.stderr.on("data", (data) => {
      stderr += data.toString();
    });

    child.stdin.on("error", () => {
      // Ignore stdin errors (e.g., if process exits early)
    });

    if (stdin !== undefined) {
      child.stdin.write(stdin);
    }

    child.stdin.end();

    child.on("close", (code) => {
      clearTimeout(timeout);

      resolve({
        command: "hermes",
        args,
        cwd: cwd || ".",
        exit_code: code,
        timed_out: timedOut,
        stdout: truncateOutput(stdout),
        stderr: truncateOutput(stderr),
      });
    });

    child.on("error", (err) => {
      clearTimeout(timeout);
      resolve({
        command: "hermes",
        args,
        cwd: cwd || ".",
        exit_code: 1,
        timed_out: timedOut,
        stdout: truncateOutput(stdout),
        stderr: truncateOutput(stderr + err.message),
      });
    });
  });
}

const sessionNameSchema = z.string().regex(
  /^[a-zA-Z0-9_\-\. ]+$/,
  "session_name must contain only letters, numbers, dash, underscore, dot, and space"
);

const cwdSchema = z.string().optional();

const timeoutSchema = z.number().int().min(1000).max(600000).default(600000);

const promptSchema = z.string().min(1, "prompt must be non-empty");

export function registerHermesTools(server) {
  server.registerTool(
    "hermes_run",
    {
      title: "Hermes Run",
      description: "Runs the Hermes CLI with arbitrary args and optional stdin.",
      inputSchema: {
        args: z.array(z.string()).default([]),
        stdin: z.string().optional(),
        cwd: cwdSchema,
        timeout_ms: timeoutSchema,
      },
    },
    async ({ args, stdin, cwd, timeout_ms }) => {
      const result = await runHermes({ args, stdin, cwd, timeout_ms });
      return textResponse(JSON.stringify(result, null, 2));
    }
  );

  server.registerTool(
    "hermes_ping",
    {
      title: "Hermes Ping",
      description: "Confirms the Hermes CLI is available.",
      inputSchema: {},
    },
    async () => {
      const result = await runHermes({ args: ["--version"] });
      return textResponse(JSON.stringify(result, null, 2));
    }
  );

  server.registerTool(
    "hermes_help",
    {
      title: "Hermes Help",
      description: "Shows Hermes CLI help.",
      inputSchema: {},
    },
    async () => {
      const result = await runHermes({ args: ["--help"] });
      return textResponse(JSON.stringify(result, null, 2));
    }
  );

  server.registerTool(
    "hermes_session_status",
    {
      title: "Hermes Session Status",
      description: "Shows the current Hermes session status.",
      inputSchema: {},
    },
    async () => {
      const result = await runHermes({ args: ["status"] });
      return textResponse(JSON.stringify(result, null, 2));
    }
  );

  server.registerTool(
    "hermes_sessions_list",
    {
      title: "Hermes Sessions List",
      description: "Lists all Hermes sessions.",
      inputSchema: {},
    },
    async () => {
      const result = await runHermes({ args: ["sessions", "list"] });
      return textResponse(JSON.stringify(result, null, 2));
    }
  );

  server.registerTool(
    "hermes_session_send",
    {
      title: "Hermes Session Send",
      description: "Sends a prompt to a Hermes session and returns the result.",
      inputSchema: {
        prompt: promptSchema,
        session_name: sessionNameSchema.default("space-rocks-mcp"),
        cwd: cwdSchema,
        timeout_ms: timeoutSchema,
      },
    },
    async ({ prompt, session_name, cwd, timeout_ms }) => {
      const result = await runHermes({
        args: ["chat", "-Q", "--continue", session_name, "--query", prompt],
        cwd,
        timeout_ms,
      });
      return textResponse(JSON.stringify(result, null, 2));
    }
  );
}