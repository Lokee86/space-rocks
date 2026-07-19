# Workspace MCP
Parent index: [Agent](./!INDEX.md)

## Purpose

The MCP server is maintained as its own sibling repository rather than inside Space Rocks.

## Location

- Local repository: `D:\\!bin\\workspace-mcp`
- GitHub repository: `Lokee86/workspace-mcp`
- Space Rocks remains the default configured project through the untracked MCP `.env`; the MCP source itself is not owned by this repository.

## Local usage

- `WORKSPACE_ROOT` should normally be `D:\\!bin` so the MCP can operate across local repositories.
- `PROJECT_ROOT` selects the active project for project-relative integrations.
- `GODOT_PROJECT_ROOT` selects the Godot project used by EngineForge.
- Info MCP uses port `8789`.
- Write MCP uses port `8788`.
- The consolidated development server uses port `8889`.

For server architecture, tool contracts, startup instructions, and tests, use the sibling repository's `README.md` and `docs/mcp-servers.md`.

## Workflow boundary

- MCP read/write tools are the primary implementation path.
- Hermes is reserved for delegated sub-agent work and parallel independent workstreams.
- Space Rocks documentation should describe how the project consumes the MCP, not duplicate the MCP implementation documentation.
