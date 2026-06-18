# Space Rocks MCP

This folder contains the local MCP servers that connect agents to the Space Rocks repo and the local Godot/EngineForge bridge.

Use this package for agent access only. It is not a general app server.

## Folder purpose

- `tools/space-rocks-mcp` holds the MCP HTTP servers, shared transport helpers, repo tool groups, and EngineForge bridge adapters.
- The package provides one read-only server for planning and one write-capable server for implementation.
- The legacy `server.js` entrypoint is kept for compatibility and should not be expanded.

## Servers

- `server-info-next.js` runs the read-only info MCP server on port `8789`.
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

## Chrome DevTools / Plasmic bridge

The info/read MCP server can optionally expose Chrome DevTools and read-only Plasmic tools.

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

Note:

- `server-info-next.js` defaults to `8789`.
- `server-write.js` defaults to `8788`.
- PowerShell does not use `PORT=8788 command` syntax.

## Connector URLs

- Info MCP: `http://127.0.0.1:8789/mcp`
- Write MCP: `http://127.0.0.1:8788/mcp`

## Safety boundaries

- Never expose the write MCP server through ngrok. If remote access is needed, expose only the read-only info server.
- Keep the info server read-only. Do not add write tools to it.
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
