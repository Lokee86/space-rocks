# Space Rocks MCP

This folder contains the local MCP servers that connect agents to the Space Rocks repo and the local Godot/EngineForge bridge.

Use this package for agent access only. It is not a general app server.

## Folder purpose

- `tools/space-rocks-mcp` holds the MCP HTTP servers, shared transport helpers, repo tool groups, and EngineForge bridge adapters.
- The package provides a planning/inspection server and one write-capable server for implementation.
- The legacy `server.js` entrypoint is kept for compatibility and should not be expanded.

## Servers

- `server-info-next.js` runs the planning/inspection MCP server on port `8789`.
- It includes workspace read/search, read-only EngineForge diagnostics, optional Chrome/Plasmic tools when enabled, and Hermes CLI tools.
- It is no longer purely read-only because Hermes session prompts can cause code edits and consume model/API usage.
- `server-write.js` runs the write-capable MCP server on port `8788`.
- `server.js` is legacy/compatibility and should not be expanded.

## OAuth-protected MCP startup

Both current shared-server entrypoints load the local `.env` file through `dotenv` before the shared HTTP/OAuth modules are evaluated.
The shared HTTP server fails closed at startup if `AUTH0_ISSUER`, `AUTH0_AUDIENCE`, or `RESOURCE_SERVER_URL` are missing or invalid.

Required environment variables:

- `AUTH0_ISSUER` - Auth0 issuer URL, including the tenant domain, using HTTPS
- `AUTH0_AUDIENCE` - configured API audience value
- `RESOURCE_SERVER_URL` - canonical resource server URL, using HTTPS

The Auth0 Client ID and Client Secret are configured on the ChatGPT connector side and are not consumed by this server.

If the ngrok endpoint changes, update the ChatGPT connector URL as well. The configured audience and resource values must stay consistent with the OAuth setup.

Use `.env.example` as the local template; do not add secrets to tracked files.

## Workspace and project roots

- `WORKSPACE_ROOT` controls the active Info MCP read/search boundary and the default/allowed workspace boundary for Hermes `cwd`. When unset, it defaults to the Space Rocks repository root.
- `SPACE_ROCKS_REPO` remains the Space Rocks-specific repository root used by project-specific integrations and compatibility paths. Its package-location default is the actual Space Rocks repository root.
- EngineForge remains independently configured or discovered through `SPACE_ROCKS_GODOT_PROJECT`, `ENGINEFORGE_BRIDGE_FILE`, and `ENGINEFORGE_BRIDGE_URL`. Setting a broader `WORKSPACE_ROOT` does not widen EngineForge to the whole workspace.

## Shared modules

The main shared modules are:

- `shared/http_mcp_server.js` for the local HTTP transport and `/mcp` endpoint.
- `shared/responses.js` for MCP text/JSON response helpers.
- `shared/paths.js` for workspace and Space Rocks repository root resolution.
- `shared/text_files.js` for text-file detection, workspace walking, and workspace search helpers.
- `shared/restricted_commands.js` for the guarded workspace command registry, policy, and launcher.
- `shared/restricted_command_tools.js` for asynchronous restricted workspace command tools.
- `shared/repo_readonly_tools.js` for workspace read/search tools.
- `shared/repo_write_tools.js` for compatibility-named configured-workspace write tools.
- `shared/workspace_writes.js` for transactional text writes and exact replacements within `WORKSPACE_ROOT`.
- `shared/engineforge_bridge.js` for discovering and calling the local EngineForge bridge.
- `shared/engineforge_readonly_tools.js` for safe Godot bridge diagnostics.
- `shared/engineforge_write_tools.js` for Godot mutation tools.
- `shared/hermes_tools.js` for Hermes CLI MCP tools.
- `shared/job_manager.js` for the in-memory asynchronous process-job lifecycle, bounded output buffers, cancellation, timeouts, and retention.
- `shared/job_tools.js` for MCP status, output-read, cancellation, and list tools for process jobs.

## Chrome DevTools / Plasmic bridge

The info/read MCP server can optionally expose Chrome DevTools and read-only Plasmic tools.
In this chat setup, `server-info-next.js` is the local MCP app used by ChatGPT for Plasmic work, and its Plasmic read/edit tools are exposed there behind `ENABLE_CHROME_DEVTOOLS=1`.
`server-write.js` is not the active ChatGPT Plasmic tool surface in this setup and should not be changed for Plasmic unless explicitly requested later.

Start it from `tools/space-rocks-mcp`:

```powershell
$env:ENABLE_CHROME_DEVTOOLS="1"
npm run start:info
```

## Tool groups

- Workspace read tools: `ping`, `workspace_root`, `repo_root`, `list_repo_tree`, `read_repo_file`, `search_repo_text`. The repo-named tools remain compatibility names and operate on workspace-relative paths after `WORKSPACE_ROOT` is configured; `repo_root` returns the configured workspace root as an alias of `workspace_root`.
- Repo write tools: `ping`, `write_repo_file`, `replace_in_repo_file`, `apply_repo_file_edits`
- Workspace command tools: `list_workspace_commands`, `command_job_start`
- EngineForge read tools: bridge info, bridge status, route probing, command probing, project info, scene tree, node properties, editor logs
- EngineForge write tools: scene open/save/create, node create/delete/duplicate/reparent/property/transform, script create/edit/detach/delete/attach, resource create, material helpers, editor play/stop/pause, console clear, animation play/stop
- Hermes tools: `hermes_run`, `hermes_ping`, `hermes_help`, `hermes_session_status`, `hermes_sessions_list`, `hermes_session_send`, `hermes_session_send_batch`, `hermes_job_start`, `hermes_terminal_start`, `hermes_terminal_send`, `hermes_terminal_read`, `hermes_terminal_resize`, `hermes_terminal_close`. These tools provide access to the Hermes CLI, asynchronous Hermes session jobs, and persistent Hermes PTY sessions.
- Process-job tools: `job_status`, `job_read`, `job_cancel`, `job_list`. These tools inspect and control retained asynchronous process jobs.

## Workspace write tools

The write server's file mutation tools operate inside the configured `WORKSPACE_ROOT`. The compatibility names `write_repo_file` and `replace_in_repo_file` are retained, but they no longer imply that the target must be inside only the Space Rocks repository.

- `write_repo_file` writes a UTF-8 text file and returns `{ "changed_files": [...] }`. Creating a missing file is allowed; replacing an existing file requires `overwrite: true` and the default remains `overwrite: false`.
- `replace_in_repo_file` replaces exactly one occurrence of `expected` with `replacement` and returns `{ "changed_files": [...] }`.
- `apply_repo_file_edits` applies a transactional batch and returns `{ "changed_files": [...] }`.

`apply_repo_file_edits` accepts an `edits` array with 1–100 entries. Each entry is one of:

```json
{ "type": "write", "path": "src/file.js", "text": "content", "overwrite": false }
```

```json
{ "type": "replace", "path": "src/file.js", "expected": "old", "replacement": "new" }
```

The full batch is preflighted before any target is mutated. Edits to the same file are applied in input order, each unique file is committed once, and each final file is staged in its target directory and installed through rename. If a later commit fails, the service makes a best-effort rollback to original contents and removes newly created targets.

Paths under `.worktrees/` and `.workingtrees/` are writable when they remain inside `WORKSPACE_ROOT`. `.git` path components, root escapes, unsupported file types, and paths containing existing symlink components are rejected.

## Write MCP command jobs

Write MCP exposes restricted workspace commands as asynchronous jobs:

- `list_workspace_commands` returns the fixed command IDs.
- `command_job_start` starts one command job and returns its job snapshot immediately.
- `job_status`, `job_read`, `job_cancel`, and `job_list` provide status, cursor-based stdout/stderr reads, cancellation, and retained-job listing.

The exact command IDs are:

`go`, `gofmt`, `python`, `pytest`, `ruby`, `bundle`, `rails`, `npm`, `node`, `godot`, `rg`, `grep`, `find`, `ls`, `cat`, `sed`, `head`, `tail`, `wc`, `diff`.

Commands launch with `shell: false`; command arguments are passed as arguments and are not shell-expanded. `cwd` defaults to `WORKSPACE_ROOT`, must exist as a directory, and must resolve inside the real workspace root. `.worktrees/` and `.workingtrees/` descendants are valid when they remain inside that root. Output buffering, timeouts, retention, and cancellation come from the shared `ProcessJobManager`.

The blocked command IDs are `git`, `cmd`, `powershell`, `pwsh`, `bash`, `sh`, `wsl`, and `npx`. Argument and environment policy also blocks Node eval/print, Python `-c` and non-`pytest` modules, Ruby `-e`, npm exec/x, Rails runner/console, Go `env -w` and `clean -modcache`, unsafe Bundler modes, NULs, oversized arguments, and protected environment overrides such as `PATH`, `NODE_OPTIONS`, `PYTHONPATH`, `RUBYOPT`, `GEM_HOME`, `GEM_PATH`, `COMSPEC`, `SHELL`, and `GIT_`/`SSH_` variables.

This is a guarded process runner, not an OS filesystem sandbox. A workspace-owned script can still perform whatever its language runtime permits.

## Hermes CLI Tools

The Hermes MCP tools provide bounded CLI access:

- `hermes_run` - Runs the Hermes CLI with arbitrary args and optional stdin
- `hermes_ping` - Confirms the Hermes CLI is available (runs `hermes --version`)
- `hermes_help` - Shows Hermes CLI help (runs `hermes --help`)
- `hermes_session_status` - Shows the current Hermes session status (runs `hermes status`)
- `hermes_sessions_list` - Lists all Hermes sessions (runs `hermes sessions list`)
- `hermes_session_send` - Sends a prompt to a Hermes session and returns the result
- `hermes_session_send_batch` - Sends multiple prompts to a Hermes session and returns the results
- `hermes_terminal_start` - Starts a persistent interactive Hermes PTY session and returns session metadata plus initial readable output
- `hermes_terminal_send` - Sends input to a PTY session, optionally appending Enter, and returns incremental output after prompt activity settles
- `hermes_terminal_read` - Reads unread incremental output from a PTY session
- `hermes_terminal_resize` - Resizes a PTY session
- `hermes_terminal_close` - Closes a PTY session and removes it from the manager
- `hermes_job_start` - Starts an asynchronous Hermes session prompt job and returns its job snapshot immediately

## Asynchronous Hermes jobs

Asynchronous Hermes jobs use the shared in-memory process-job manager. The workflow is:

1. Call `hermes_job_start` with a prompt, optional `session_name`, optional allowed `cwd`, and optional `timeout_ms`.
2. The tool returns immediately with the job snapshot and opaque `job_...` ID.
3. Poll `job_status` with that ID until the job reaches a terminal state.
4. Call `job_read` for `stdout` or `stderr`, passing the returned cursor as the next `cursor` to read incremental output. The response preserves `nextCursor`, total offsets, rollover start offsets, and `truncated`.
5. Call `job_cancel` when the job should be stopped before completion.

The process-job states are `queued`, `running`, `succeeded`, `failed`, `cancelled`, and `timed_out`.

The manager defaults are:

- maximum four concurrently running jobs;
- a 50,000-character rolling buffer for each job's stdout and stderr stream;
- completed-job retention for 15 minutes, after which status and output are no longer available;
- a 5-minute default timeout for Hermes jobs, configurable per job up to a maximum of one hour.

Only one queued or running asynchronous job is allowed for a given Hermes `session_name`. Jobs for different session names, and unrelated process jobs, may overlap subject to the shared concurrency limit. A new job may use the session after the previous job reaches a terminal state or expires.

The `hermes_session_send` tool preserves continuous context by sending prompts to the same named Hermes session. The internal command shape is:

```
hermes chat -Q --continue <session_name> --query <prompt>
```

The default session name is `space-rocks-mcp`.

The `hermes_session_send_batch` tool queues multiple prompts to the same named session through the connector. It starts all sends before awaiting results, preserves input order in `results`, and returns each item with `index`, `prompt_preview`, and the existing Hermes result object.

The persistent PTY tools use one Hermes process per session and are intended for interactive workflows:

- `cwd` defaults to `WORKSPACE_ROOT` and must resolve to that root, one of its descendants, the separate Hermes installation root, or one of that root's descendants
- `session_id` values are generated by the MCP server and are unguessable
- `hermes_terminal_send` uses `\r` for Enter by default
- unread output is buffered and returned in a cleaned, MCP-readable form
- each PTY session keeps only the newest 50,000 characters in the MCP output buffer; this is a rolling transport/recovery buffer, not the Hermes conversation transcript
- Hermes persists its own session history separately, so older conversation context remains available through Hermes even if old terminal output rolls out of the MCP buffer
- abandoned PTY sessions expire after idle timeout and are cleaned up without keeping Node alive

`cwd` is optional; this shape starts Hermes in its local installation directory.

```js
hermes_terminal_start({
  cwd: "C:\\Users\\archa\\AppData\\Local\\hermes"
})
```

The input shape is:

- `prompts` - required array of non-empty prompt strings, minimum one prompt
- `session_name` - same default and naming restrictions as `hermes_session_send`
- `cwd` - optional allowed working directory; defaults to `WORKSPACE_ROOT`

**Important**: Info MCP does not expose:
- General shell
- Arbitrary command strings
- Arbitrary shell commands through Hermes
- Repo write tools
- EngineForge write tools

## Start commands

WSL/Linux:

```bash
cd tools/space-rocks-mcp
PORT=8789 node server-info-next.js
PORT=8788 node server-write.js
```

PowerShell:

```powershell
cd D:\!bin\space-rocks\tools\space-rocks-mcp
$env:WORKSPACE_ROOT="D:\!bin"
node server-info-next.js
node server-write.js
```

## Plasmic write bridge

Plasmic mutation tools are exposed only from `server-write.js`.
Set `ENABLE_CHROME_DEVTOOLS=1` to enable the Chrome DevTools / Plasmic bridge.

Initial write tools:

- `plasmic_insert_html`
- `plasmic_change_element`
- `plasmic_delete_element`

Broader tools such as `createComponent`, token mutation, animation mutation, and page meta are intentionally deferred.
Test edits on a small known element before making larger changes.

Note:

- `server-info-next.js` defaults to `8789`.
- `server-write.js` defaults to `8788`.
- PowerShell does not use `PORT=8788 command` syntax.

## Connector URLs

- Info MCP: `http://127.0.0.1:8789/mcp`
- Write MCP: `http://127.0.0.1:8788/mcp`

## Safety boundaries

- Info MCP must not import repo write tools.
- Info MCP must not import EngineForge write tools.
- Info MCP must not expose a general shell.
- Info MCP intentionally exposes Hermes session tools for continuous context, which can cause code edits and consume model/API usage.
- Do not expose Info MCP through ngrok while Hermes tools are default-enabled.
- Treat Info MCP as trusted-local only.
- Keep the write server local and bounded to workspace writes, restricted command jobs, and EngineForge mutations.
- Do not edit `package.json` or the installed EngineForge plugin from this README workflow.
- The MCP server does not contain EngineForge itself. It wraps the local Godot bridge by reading `client/.godot/engineforge/bridge.json`.
- The Godot bridge is provided by the installed plugin at `client/addons/engineforge_bridge/engineforge_bridge.gd`.
- Do not edit the installed EngineForge plugin manually.

## Common troubleshooting

- If the wrong tools appear, confirm you started the correct server on the correct port.
- If the MCP tool list looks stale in Codex, restart the Codex session.
- If the bridge cannot connect, confirm Godot is running and the EngineForge plugin is installed.
- If bridge discovery fails, check `client/.godot/engineforge/bridge.json`.
- If a bridge command fails, verify the real command name from `/capabilities` instead of guessing.
- If the write server is reachable outside the machine, stop and remove that exposure.
- If you only need inspection or planning, use the info server instead of the write server.