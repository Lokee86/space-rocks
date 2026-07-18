import { z } from "zod";

import { JOB_STATES, defaultProcessJobManager } from "./job_manager.js";
import { textResponse } from "./responses.js";

const jobIdSchema = z.string().regex(/^job_[A-Za-z0-9_-]+$/, "job_id must be a valid job_ ID");
const streamSchema = z.enum(["stdout", "stderr"]);
const stateSchema = z.enum(JOB_STATES);

function jsonResponse(value) {
  return textResponse(JSON.stringify(value, null, 2));
}

export function registerJobTools(server, { manager = defaultProcessJobManager } = {}) {
  server.registerTool(
    "job_status",
    {
      title: "Process Job Status",
      description: "Returns the current status and output metadata for a process job.",
      inputSchema: { job_id: jobIdSchema },
    },
    async ({ job_id }) => jsonResponse(manager.status(job_id))
  );

  server.registerTool(
    "job_read",
    {
      title: "Read Process Job Output",
      description: "Reads incremental stdout or stderr from a process job using a cursor.",
      inputSchema: {
        job_id: jobIdSchema,
        stream: streamSchema.default("stdout"),
        cursor: z.number().int().min(0).default(0),
        max_chars: z.number().int().min(1).max(50000).default(50000),
      },
    },
    async ({ job_id, stream, cursor, max_chars }) => jsonResponse(manager.read(job_id, {
      stream,
      cursor,
      maxChars: max_chars,
    }))
  );

  server.registerTool(
    "job_cancel",
    {
      title: "Cancel Process Job",
      description: "Cancels a queued or running process job.",
      inputSchema: { job_id: jobIdSchema },
    },
    async ({ job_id }) => jsonResponse(manager.cancel(job_id))
  );

  server.registerTool(
    "job_list",
    {
      title: "List Process Jobs",
      description: "Lists retained process jobs, optionally filtered by lifecycle state.",
      inputSchema: { state: stateSchema.optional() },
    },
    async ({ state }) => {
      const jobs = manager.list();
      return jsonResponse(state === undefined ? jobs : jobs.filter((job) => job.state === state));
    }
  );
}
