## Space Rocks MCP Write Server and Godot Bridge

Use this skill when implementing Space Rocks changes through the local write MCP server, especially changes that touch the Godot editor, scenes, nodes, scripts, resources, or bounded repo writes.

## Core rule

Use the write MCP server only for implementation work.

Use the read-only info MCP server first when diagnosing, planning, inspecting repo state, inspecting Godot scene state, reading editor logs, or checking bridge status.

Do not expose the write MCP server through ngrok or any remote tunnel.

## Server split

Space Rocks has two local MCP servers under:

```text
tools/space-rocks-mcp/
```

Info MCP:

```text
server-info-next.js
port 8789
http://127.0.0.1:8789/mcp
```

Use it for:

* repo reads
* repo search
* Godot bridge status
* current scene inspection
* scene tree inspection
* node property inspection
* editor log reads
* planning and diagnosis

Write MCP:

```text
server-write.js
port 8788
http://127.0.0.1:8788/mcp
```

Use it for:

* bounded repo file writes
* exact repo text replacements
* allowlisted test commands
* Godot scene mutations
* Godot node mutations
* Godot script/resource mutations
* editor play/stop/pause
* console clearing
* animation play/stop

## Startup commands

Run from:

```text
tools/space-rocks-mcp/
```

WSL/Linux:

```bash
PORT=8789 node server-info-next.js
PORT=8788 node server-write.js
```

PowerShell:

```powershell
$env:PORT=8789; node server-info-next.js
$env:PORT=8788; node server-write.js
```

Package scripts may also be available:

```bash
npm run start:info-next
npm run start:write
```

If tools do not appear after changing MCP config, restart the Codex/session using the MCP server. Tool lists can be session-cached.

## Bridge dependency

The MCP server does not contain EngineForge. It wraps the local Godot bridge.

Bridge discovery normally reads:

```text
client/.godot/engineforge/bridge.json
```

The bridge comes from the installed Godot plugin:

```text
client/addons/engineforge_bridge/engineforge_bridge.gd
```

Do not edit the installed EngineForge plugin manually.

If bridge discovery fails:

* confirm Godot is running
* confirm the EngineForge plugin is installed and active
* confirm `client/.godot/engineforge/bridge.json` exists
* confirm the bridge reports `/status`
* confirm real command support from `/capabilities`

## Command shape

The bridge command shape is:

```json
{
  "category": "scene",
  "action": "getTree",
  "params": {}
}
```

Think in:

```text
category / action / params
```

Do not guess dotted names such as:

```text
scene.tree
scene.current
node.properties
```

Use the MCP wrapper tools or confirmed `/capabilities` names.

## Write MCP repo tools

Use these for repo-side changes:

```text
write_repo_file
replace_in_repo_file
list_allowed_commands
run_allowed_command
```

Prefer `replace_in_repo_file` for small targeted edits where the expected text is unique.

Use `write_repo_file` for new files or full-file replacement only when the whole file content is known.

Do not use repo write tools for binary files, imported assets, generated files, or Godot editor-owned metadata unless explicitly required.

## Allowlisted commands

The write server only runs named allowlisted commands.

Known command names:

```text
go_server_tests
godot_unit_tests
tools_boundary_tests
data_sync_tests
```

Use `list_allowed_commands` if unsure.

Run the smallest relevant command after implementation. Do not run unrelated test suites just to look busy.

## Godot write tools

Use these when changing live Godot editor state.

Scene tools:

```text
godot_scene_create
godot_scene_open
godot_scene_save
```

Node tools:

```text
godot_node_create
godot_node_delete
godot_node_duplicate
godot_node_reparent
godot_node_set_property
godot_node_set_transform
```

Script tools:

```text
godot_script_create
godot_script_edit
godot_script_attach
godot_script_detach
godot_script_delete
```

Resource tools:

```text
godot_resource_create
godot_resource_create_material
godot_resource_set_material_property
```

Editor tools:

```text
godot_editor_play
godot_editor_stop
godot_editor_pause
godot_console_clear
```

Animation tools:

```text
godot_animation_play
godot_animation_stop
```

## Choosing repo edits vs Godot bridge edits

Use repo text edits when:

* changing ordinary `.gd` code
* changing docs
* changing small config files
* applying exact, reviewable text replacements
* the Godot editor does not need to own the mutation

Use Godot bridge edits when:

* creating/opening/saving scenes
* creating/deleting/reparenting nodes
* setting node properties through the editor model
* changing transforms
* creating resources/materials
* attaching/detaching scripts
* validating behavior in the running editor

For `.tscn` files, prefer bridge edits when the change is structural. Godot scene diffs can rewrite IDs, offsets, ownership, imports, and metadata.

## Safe Godot mutation workflow

Before mutating a scene:

1. Confirm the target scene.
2. Inspect the current scene tree or scene file.
3. Decide whether the bridge or a text edit is safer.
4. Open the target scene if needed.
5. Apply one small mutation.
6. Save the scene if the mutation should persist.
7. Inspect the resulting scene/tree or diff.
8. Run only the smallest relevant verification.

Do not batch large unrelated scene edits into one bridge session.

## Script edit workflow

For script edits through the bridge:

* Prefer exact `oldText` / `newText` replacement.
* Use full `contents` replacement only for small files or newly-created scripts.
* Keep GDScript files small.
* Create a new script or helper before growing a large file.
* Preserve existing signal names, node paths, exported properties, and public call sites unless the task explicitly changes them.

## Scene/node safety

Be careful with Godot scene diffs. Godot may rewrite:

* `uid`
* `unique_id`
* offsets
* anchors
* imports
* scene metadata

Do not revert unrelated Godot/editor changes unless explicitly instructed.

Do not casually touch:

```text
client/.godot/
*.import
```

Avoid committing generated recordings and temporary artifacts:

```text
*.avi
tmp/
*/tmp/
client/.godot/
```

## Testing and verification

Use the smallest relevant verification.

Server gameplay changes:

```text
run_allowed_command name=go_server_tests
```

Godot client unit changes:

```text
run_allowed_command name=godot_unit_tests
```

MCP/tool boundary changes:

```text
run_allowed_command name=tools_boundary_tests
```

Data sync changes:

```text
run_allowed_command name=data_sync_tests
```

For UI/scene changes, automated tests may not prove the visual result. Also report what should be smoke-tested manually in Godot.

## Troubleshooting

If the write server tools are missing:

* confirm Codex is pointed at `http://127.0.0.1:8788/mcp`
* confirm `server-write.js` is running
* restart the Codex/session if the tool list is stale

If bridge commands fail:

* confirm Godot is running
* confirm EngineForge bridge status is OK
* confirm the target scene is open when using scene/node tools
* confirm node paths are exact
* inspect available bridge capabilities instead of guessing command names

If Godot state appears stale:

* inspect the active scene again
* save/reopen the target scene if needed
* avoid continuing mutation work against an unknown editor state

If the write MCP server is reachable from outside the local machine:

* stop it
* remove the tunnel/exposure
* restart it locally only

## Reporting

After using the write server, report:

* changed files
* bridge mutations performed
* verification run
* verification result
* manual smoke-test notes, if any

Keep reports short and concrete.
