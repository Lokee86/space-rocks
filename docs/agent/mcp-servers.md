# Space Rocks MCP Servers
Parent index: [Agent](./!INDEX.md)

## Purpose

This is the permanent agent-facing reference for the Space Rocks MCP split.

## Overview

Use this as the first stop when you need to decide which MCP server to connect to, which bridge commands are safe, and how to start the local services.

## Rules

- Info MCP is the planning/inspection server.
- Info MCP includes Hermes session tools for agent orchestration.
- Info MCP must never import repo write tools or EngineForge write tools.
- Info MCP intentionally registers Hermes CLI tools.
- Hermes CLI tools are not a general terminal bridge.
- Hermes tools can run Hermes CLI args, but not arbitrary shell commands.
- Write MCP is the implementation server.
- ChatGPT and other planning agents should use Info MCP.
- Codex and implementation work should use Write MCP.
- Write MCP should stay local only.
- Do not expose Write MCP through ngrok.
- Do not expose Info MCP through ngrok while Hermes tools are enabled by default.
- Treat Info MCP as trusted-local only because Hermes session prompts can mutate files and consume usage.
- The bridge command set comes from `/capabilities` and the installed plugin.
- Do not assume guessed command names exist.
- Do not use stale names like `scene.current` or `scene.tree`.
- For bridge diagnostics, prefer the read-only MCP tools first.

## Server Split

|| Server | Port | Entry file | Consumer | Role |
||---|---:|---|---|---|
|| Info MCP | 8789 | server-info-next.js | ChatGPT / planning | Workspace reads/search, read-only Godot diagnostics, optional Chrome/Plasmic tools, default Hermes session tools |
|| Write MCP | 8788 | server-write.js | Codex / implementation | Bounded repo writes, allowlisted commands, Godot bridge mutations |

## Shared process-job modules

- `tools/space-rocks-mcp/shared/job_manager.js` owns the in-memory asynchronous process-job lifecycle, process spawning, cancellation, timeouts, bounded output buffers, concurrency, and retention.
- `tools/space-rocks-mcp/shared/job_tools.js` exposes the read/control MCP surface: `job_status`, `job_read`, `job_cancel`, and `job_list`.
- The Info MCP server uses one shared `defaultProcessJobManager` for `hermes_job_start` and the four process-job tools.

## Workspace and project roots

- `WORKSPACE_ROOT` controls the active Info MCP read/search boundary and the default/allowed workspace boundary for Hermes `cwd`. It defaults to the Space Rocks repository root when unset.
- `SPACE_ROCKS_REPO` remains the Space Rocks-specific repository root for project-specific integrations and compatibility paths. Its package-location default remains the actual Space Rocks repository root.
- The canonical `workspace_root` tool returns the configured workspace root. Existing repo-named read tools (`repo_root`, `list_repo_tree`, `read_repo_file`, and `search_repo_text`) remain compatibility names but operate workspace-relative after `WORKSPACE_ROOT` is configured; `repo_root` returns the same configured workspace root.
- EngineForge remains independently configured or discovered through `SPACE_ROCKS_GODOT_PROJECT`, `ENGINEFORGE_BRIDGE_FILE`, and `ENGINEFORGE_BRIDGE_URL`. A broader workspace does not widen EngineForge access beyond its configured Godot project and local bridge.

## Write MCP workspace file tools

The Write MCP file tools operate inside the configured `WORKSPACE_ROOT`:

- `write_repo_file` and `replace_in_repo_file` retain their compatibility names but target the configured workspace.
- `apply_repo_file_edits` accepts an `edits` array of 1–100 discriminated entries:
  - `{ "type": "write", "path": "...", "text": "...", "overwrite": false }`
  - `{ "type": "replace", "path": "...", "expected": "...", "replacement": "..." }`
- `write_repo_file` permits creating missing files, but an existing target requires `overwrite: true`; its default remains `false`.
- `replace_in_repo_file` and replace edits require exactly one occurrence of `expected`.
- All three mutation tools return JSON containing `changed_files`.

The full batch is preflighted before mutation. Same-file edits are applied in input order, each unique file is committed once, and each final file is staged in its target directory and installed through rename. If a later commit fails, the service makes a best-effort rollback to original contents and removes newly created targets.

Descendants under `.worktrees/` and `.workingtrees/` are writable when they remain inside `WORKSPACE_ROOT`. `.git` components, paths escaping the root, unsupported file types, and paths containing existing symlink components are rejected.

## Hermes CLI Tools

Info MCP exposes Hermes CLI tools for continuous context across prompt sequences:

- `hermes_run` - Runs the Hermes CLI with arbitrary args and optional stdin
- `hermes_ping` - Confirms Hermes CLI is available (runs `hermes --version`)
- `hermes_help` - Shows Hermes CLI help (runs `hermes --help`)
- `hermes_session_status` - Shows the current Hermes session status (runs `hermes status`)
- `hermes_sessions_list` - Lists all Hermes sessions (runs `hermes sessions list`)
- `hermes_session_send` - Sends a prompt into a named Hermes session
- `hermes_session_send_batch` - Sends multiple prompts into a named Hermes session through the connector
- `hermes_job_start` - Starts an asynchronous Hermes session prompt job and returns immediately with a job snapshot

### Asynchronous Hermes session jobs

Use `hermes_job_start` when a Hermes prompt should run without holding the MCP request open. The lifecycle is:

1. Start the job with `prompt`, optional `session_name`, optional allowed `cwd`, and optional `timeout_ms`.
2. Receive the immediate response containing the opaque `job_...` ID and initial job snapshot.
3. Poll `job_status` until the state is terminal.
4. Read incremental `stdout` or `stderr` with `job_read`, passing each response's `nextCursor` as the next `cursor`. Cursor reads preserve total offsets, rolling-buffer start offsets, and `truncated` when older output has rolled out.
5. Use `job_cancel` when cancellation is required; use `job_list` to inspect retained jobs, optionally filtered by state.

The lifecycle states are `queued`, `running`, `succeeded`, `failed`, `cancelled`, and `timed_out`.

The shared manager allows four running jobs by default. Each job keeps a separate rolling 50,000-character buffer for stdout and stderr. Completed jobs are retained for 15 minutes before status and output expire. Hermes jobs default to a 5-minute timeout; `timeout_ms` is configurable from 1,000 ms through a maximum of one hour.

Only one queued or running asynchronous Hermes job is allowed for each `session_name`. Unrelated session names and unrelated process jobs may overlap, subject to the shared concurrency limit. A new job for a session is allowed after the prior job reaches a terminal state or expires.

### hermes_session_send

`hermes_session_send` sends a single prompt into a named Hermes session. Reusing the same `session_name` preserves continuous Hermes context across prompt sequences.

**Internal command shape:**
```
hermes chat -Q --continue <session_name> --query <prompt>
```

**Default session name:** `space-rocks-mcp`

**Input parameters:**
- `prompt` (required) - The prompt string to send to the Hermes session
- `session_name` (optional, default: `space-rocks-mcp`) - The named session to continue
- `cwd` (optional) - Allowed working directory; defaults to `WORKSPACE_ROOT`

**Important notes:**
- This is session continuation through Hermes, not one-shot workflow guidance.
- It is not a general terminal bridge.
- It does not expose arbitrary shell commands.
- It cannot execute bash, PowerShell, git, Python, npm, or arbitrary shell commands.
- Hermes session prompts can mutate files and consume model/API usage.

### hermes_session_send_batch

`hermes_session_send_batch` queues multiple prompts to the same named Hermes session through the connector. `hermes_session_send` remains the single-prompt tool.

**Internal command shape:**
```
hermes chat -Q --continue <session_name> --query <prompt>
```

**Default session name:** `space-rocks-mcp`

**Input parameters:**
- `prompts` (required) - Array of non-empty prompt strings; minimum one prompt
- `session_name` (optional, default: `space-rocks-mcp`) - Same naming restrictions as `hermes_session_send`
- `cwd` (optional) - Allowed working directory; defaults to `WORKSPACE_ROOT`

**Behavior:**
- Starts all sends before awaiting results so prompts are queued without waiting for each prior response.
- Preserves input order in the returned `results` array.
- Each result includes `index`, `prompt_preview`, and the existing Hermes result object.
- `prompt_preview` is short and truncated to avoid dumping large prompt bodies.

**Important notes:**
- This is session continuation through Hermes, not one-shot workflow guidance.
- It is not a general terminal bridge.
- It does not expose arbitrary shell commands.
- It cannot execute bash, PowerShell, git, Python, npm, or arbitrary shell commands.
- Hermes session prompts can mutate files and consume model/API usage.

## Bridge Usage

EngineForge bridge commands use this shape:

```json
{
  "category": "scene",
  "action": "getTree",
  "params": {}
}
```

Think of the command as `category/action/params`.

Use these from the Info MCP server when you need diagnostics or a safe read path:

- `scene.getActive`
- `scene.getTree`
- `project.getInfo`
- `project.scan`
- `editor.getState`
- `console.getLogs`
- `node.getProperties`

Practical use:

- `scene.getActive` for the current scene selection/state.
- `scene.getTree` for the active scene tree.
- `project.getInfo` for project metadata.
- `project.scan` when you want the bridge to rescan the project.
- `editor.getState` for editor mode/state checks.
- `console.getLogs` for editor log inspection.
- `node.getProperties` for inspecting a node by path.

Use these from the Write MCP server when you are intentionally changing Godot state:

- `scene.open`
- `scene.save`
- `node.create`
- `node.delete`
- `node.setProperty`
- `node.setTransform`
- `script.create`
- `script.edit`
- `resource.create`
- `editor.play`
- `editor.stop`

These are the practical write-side commands to reach for first.

## Startup notes

Both MCP servers depend on the local EngineForge/Godot bridge plugin that runs inside the Godot project.

Run these from `tools/space-rocks-mcp/`.

## OAuth-protected startup

The current shared HTTP MCP entrypoints load the local `.env` file through `dotenv` before the shared HTTP/OAuth modules are evaluated.
The OAuth seam fails closed at startup unless these HTTPS-safe environment values are present and valid:

- `AUTH0_ISSUER`
- `AUTH0_AUDIENCE`
- `RESOURCE_SERVER_URL`

`AUTH0_ISSUER` is normalized to include its trailing slash so it matches Auth0 JWT `iss` claims.
`AUTH0_AUDIENCE` and `RESOURCE_SERVER_URL` are preserved exactly as configured.

The Auth0 Client ID and Client Secret live on the ChatGPT connector side; they are not consumed by this server.
If the ngrok endpoint changes, update the ChatGPT connector URL too. Keep the configured audience and resource values aligned with the OAuth setup.

The protected MCP routes are `POST /mcp`, `GET /mcp`, and `DELETE /mcp`.
They require a bearer token and validate it with Auth0 JWKS via `jose.jwtVerify`, restricted to RS256, the configured issuer, the configured audience, and standard expiry/not-before checks.
The design intentionally does not add scopes.

Public routes stay public:

- `GET /` health response
- `GET /.well-known/oauth-protected-resource` metadata response
- `OPTIONS /mcp` CORS preflight

Protected responses advertise `WWW-Authenticate: Bearer resource_metadata="..."` without a scope parameter.

### WSL / Linux

Info MCP:

```bash
PORT=8789 node server-info-next.js
```

Write MCP:

```bash
PORT=8788 node server-write.js
```

If you prefer the package scripts:

```bash
npm run start:info-next
```

```bash
npm run start:write
```

### PowerShell

Info MCP:

```powershell
$env:WORKSPACE_ROOT="D:\!bin"; $env:PORT=8789; node server-info-next.js
```

Write MCP:

```powershell
$env:PORT=8788; node server-write.js
```

If you prefer npm scripts in PowerShell:

```powershell
npm run start:info-next
```

```powershell
npm run start:write
```

## Codex connection notes

- Point Codex at the Write MCP server for implementation work.
- Use the local HTTP MCP endpoint on port `8788`.
- If Codex does not see the tools after a config change, restart the Codex session.
- Session reload matters: Codex needed a new session before it could see the MCP tools.
- If the server is running but Codex still shows an old tool list, assume the session cache is stale before debugging the server itself.

## ChatGPT connection notes

- ChatGPT uses Info MCP for planning, inspection, and agent orchestration.
- Use the local HTTP MCP endpoint on port `8789`.
- Info MCP is trusted-local because Hermes session prompts can mutate files and consume usage.
- Keep ChatGPT and other planning agents off the write server.
- If you are only reading repo state or checking the Godot bridge, Info MCP is the correct server.

## Ngrok rule

- Do not expose Info MCP through ngrok while Hermes tools are enabled by default.
- Never expose Write MCP through ngrok.
- Write MCP is meant to stay local and bounded to implementation work.

## Practical usage guide

Use Info MCP when you need:

- workspace search
- workspace reads
- bridge status checks
- scene tree inspection
- editor log reads

Use Write MCP when you need:

- bounded repo writes
- transactional workspace file writes and exact replacements
- allowlisted command execution
- scene edits
- script edits
- resource creation
- play/stop actions in the editor bridge

## Troubleshooting

- Confirm the right server is running on the right port.
- Confirm the consumer is using the correct server: ChatGPT/planning on Info MCP, Codex/implementation on Write MCP.
- Confirm the EngineForge bridge plugin is installed and active in the Godot project.
- Confirm the bridge exposes `/capabilities`.
- Confirm the command uses the real `category/action/params` shape.
- Do not guess alternate command names.
- Do not use `scene.current` or `scene.tree`.
- If the tools do not appear in Codex after a config change, restart the session.
- If Write MCP is reachable from outside the machine, stop and remove that exposure.
- If Godot bridge reads fail, check the local bridge state before changing MCP wiring.

## Related docs

- [Session Primer](./session-primer.md)
- [Repo Hygiene](./repo-hygiene.md)
- [Godot Editing](./godot-editing.md)
- [Prompting And Reporting](./prompting-and-reporting.md)

## Notes

This doc owns agent-facing MCP usage, not general Godot/client implementation facts.