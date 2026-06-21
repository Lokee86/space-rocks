# Godot Editing
Parent index: [Agent](./!INDEX.md)

## Purpose

This doc guides safe agent edits to Godot scenes, scripts, imports, and tests.

## Overview

Canonical client runtime docs live under `docs/services/client/`, and canonical devtools docs live under `docs/devtools/`.

## Rules

- Use the read-only MCP/bridge tools for Godot diagnosis before guessing from scene files alone.
- Use the write MCP path only for intentional implementation edits.
- Do not guess bridge command names.
- Do not manually edit the installed EngineForge plugin.
- Be careful with Godot scene diffs.
- Godot may rewrite `uid`, `unique_id`, offsets, imports, and scene metadata.
- Do not revert unrelated user/editor changes casually.
- Avoid committing generated recordings and build artifacts.
- Do not hand-edit generated client packet/constants files.
- Client tests use GUT under `client/tests/`.
- Do not put client test helpers in `client/scripts/`.
- Devtools edits must route through real gameplay seams.

## Related docs

- [MCP Servers](./mcp-servers.md)
- [Repo Hygiene](./repo-hygiene.md)
- [Generated Files](./generated-files.md)
- [Testing](./testing.md)
- [Client Service](../services/client/!INDEX.md)
- [Client Devtools](../devtools/client/!INDEX.md)
- [Server Devtools](../devtools/server/!INDEX.md)

## Notes

This doc is for Godot editing safety, not canonical gameplay or devtools behavior.
