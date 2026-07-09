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
|| Info MCP | 8789 | server-info-next.js | ChatGPT / planning | Repo reads/search, read-only Godot diagnostics, optional Chrome/Plasmic tools, default Hermes session tools |
|| Write MCP | 8788 | server-write.js | Codex / implementation | Bounded repo writes, allowlisted commands, Godot bridge mutations |

## Hermes CLI Tools

Info MCP exposes Hermes CLI tools for continuous context across prompt sequences:

- `hermes_run` - Runs the Hermes CLI with arbitrary args and optional stdin
- `hermes_ping` - Confirms Hermes CLI is available (runs `hermes --version`)
- `hermes_help` - Shows Hermes CLI help (runs `hermes --help`)
- `hermes_session_status` - Shows the current Hermes session status (runs `hermes status`)
- `hermes_sessions_list` - Lists all Hermes sessions (runs `hermes sessions list`)
- `hermes_session_send` - Sends a prompt into a named Hermes session

### hermes_session_send

`hermes_session_send` sends prompts into a named Hermes session. Reusing the same `session_name` preserves continuous Hermes context across prompt sequences.

**Internal command shape:**
```
hermes --continue <session_name> -z <prompt>
```

**Default session name:** `space-rocks-mcp`

**Input parameters:**
- `prompt` (required) - The prompt string to send to the Hermes session
- `session_name` (optional, default: `space-rocks-mcp`) - The named session to continue
- `cwd` (optional) - Repo-relative working directory
- `timeout_ms` (optional, default: 600000) - Timeout in milliseconds (1000-600000 range)

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