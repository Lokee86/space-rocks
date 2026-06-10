# Space Rocks MCP Servers

This is the permanent agent-facing reference for the Space Rocks MCP split.

Use this as the first stop when you need to decide which MCP server to connect to, which bridge commands are safe, and how to start the local services.

## Server Split

| Server | Port | Entry file | Consumer | Role |
|---|---:|---|---|---|
| Info MCP | 8789 | server-info-next.js | ChatGPT / planning | Read/search repo plus read-only Godot bridge diagnostics |
| Write MCP | 8788 | server-write.js | Codex / implementation | Bounded repo writes, allowlisted commands, Godot bridge mutations |

## What Each Server Is For

- Info MCP is the planning and read-only server.
- Write MCP is the implementation server.
- ChatGPT and other planning agents should use Info MCP.
- Codex and implementation work should use Write MCP.
- Info MCP must never import write tools.
- Write MCP should stay local only.
- Do not expose Write MCP through ngrok.
- If you need to publish one server for remote access, only expose Info MCP.

## EngineForge / Godot Bridge Dependency

Both MCP servers depend on the local EngineForge/Godot bridge plugin that runs inside the Godot project.

- The bridge command set comes from `/capabilities` and the installed plugin.
- Do not assume guessed command names exist.
- Do not use stale names like `scene.current` or `scene.tree`.
- For bridge diagnostics, prefer the read-only MCP tools first.

## Bridge Command Format

EngineForge bridge commands use this shape:

```json
{
  "category": "scene",
  "action": "getTree",
  "params": {}
}
```

Think of the command as `category/action/params`.

## Confirmed Read-Only Bridge Commands

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

## Confirmed Write Bridge Commands

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

## Startup Commands

Run these from `tools/space-rocks-mcp/`.

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
$env:PORT=8789; node server-info-next.js
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

## Codex VS Code Connection Notes

- Point Codex at the Write MCP server for implementation work.
- Use the local HTTP MCP endpoint on port `8788`.
- If Codex does not see the tools after a config change, restart the Codex session.
- Session reload matters: Codex needed a new session before it could see the MCP tools.
- If the server is running but Codex still shows an old tool list, assume the session cache is stale before debugging the server itself.

## ChatGPT / This Assistant Connection Notes

- Use the Info MCP server for planning, inspection, and read-only diagnostics.
- Use the local HTTP MCP endpoint on port `8789`.
- Keep ChatGPT and other planning agents off the write server.
- If you are only reading repo state or checking the Godot bridge, Info MCP is the correct server.

## Ngrok Rule

- Info MCP only may be exposed through ngrok.
- Never expose Write MCP through ngrok.
- Write MCP is meant to stay local and bounded to implementation work.

## Practical Usage Guide

Use Info MCP when you need:

- repo search
- repo reads
- bridge status checks
- scene tree inspection
- editor log reads

Use Write MCP when you need:

- bounded repo writes
- allowlisted command execution
- scene edits
- script edits
- resource creation
- play/stop actions in the editor bridge

## Troubleshooting Checklist

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

