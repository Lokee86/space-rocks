# Space Rocks MCP

This folder contains the local MCP servers that connect agents to the Space Rocks repo and the local Godot/EngineForge bridge.

Use this package for agent access only. It is not a general app server.

## Folder purpose

- `tools/space-rocks-mcp` holds the MCP HTTP servers, shared transport helpers, repo tool groups, and EngineForge bridge adapters.
- The package provides a planning/inspection server and one write-capable server for implementation.
- The legacy `server.js` entrypoint is kept for compatibility and should not be expanded.

## Servers

- `server-info-next.js` runs the planning/inspection MCP server on port `8789`.
- It includes repo read/search, read-only EngineForge diagnostics, optional Chrome/Plasmic tools when enabled, and Hermes CLI tools.
- It is no longer purely read-only because Hermes session prompts can cause code edits and consume model/API usage.
- `server-write.js` runs the write-capable MCP server on port `8788`.
- `server.js` is legacy/compatibility and should not be expanded.

## Shared modules

The main shared modules are:

- `shared/http_mcp_server.js` for the local HTTP transport and `/mcp` endpoint.
- `shared/responses.js` for MCP text/JSON response helpers.
- `shared/paths.js` for repo root resolution.
- `shared/text_files.js` for text-file detection, repo walking, and repo search helpers.
- `shared/allowed_commands.js` for the bounded shell command allowlist.
- `shared/repo_readonly_tools.js` for repo read/search tools.
- `shared/repo_write_tools.js` for bounded repo write tools.
- `shared/engineforge_bridge.js` for discovering and calling the local EngineForge bridge.
- `shared/engineforge_readonly_tools.js` for safe Godot bridge diagnostics.
- `shared/engineforge_write_tools.js` for Godot mutation tools.
- `shared/hermes_tools.js` for Hermes CLI MCP tools.

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

- Repo read tools: `ping`, `repo_root`, `list_repo_tree`, `read_repo_file`, `search_repo_text`
- Repo write tools: `ping`, `write_repo_file`, `replace_in_repo_file`, `list_allowed_commands`, `run_allowed_command`
- EngineForge read tools: bridge info, bridge status, route probing, command probing, project info, scene tree, node properties, editor logs
- EngineForge write tools: scene open/save/create, node create/delete/duplicate/reparent/property/transform, script create/edit/detach/delete/attach, resource create, material helpers, editor play/stop/pause, console clear, animation play/stop
- Hermes tools: `hermes_run`, `hermes_ping`, `hermes_help`, `hermes_session_status`, `hermes_sessions_list`, `hermes_session_send`. These tools provide access to the Hermes CLI; the session tools preserve continuous context by sending prompts to a named Hermes session.

## Hermes CLI Tools

The Hermes MCP tools provide bounded CLI access:

- `hermes_run` - Runs the Hermes CLI with arbitrary args and optional stdin
- `hermes_ping` - Confirms the Hermes CLI is available (runs `hermes --version`)
- `hermes_help` - Shows Hermes CLI help (runs `hermes --help`)
- `hermes_session_status` - Shows the current Hermes session status (runs `hermes status`)
- `hermes_sessions_list` - Lists all Hermes sessions (runs `hermes sessions list`)
- `hermes_session_send` - Sends a prompt to a Hermes session and returns the result

The `hermes_session_send` tool preserves continuous context by sending prompts to the same named Hermes session. The internal command shape is:

```
hermes chat -Q --continue <session_name> --query <prompt>
```

The default session name is `space-rocks-mcp`.

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
- Keep the write server local and bounded to repo writes, allowlisted commands, and EngineForge mutations.
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