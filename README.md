# Space Rocks

Space Rocks is an Asteroids-inspired game project with a Godot client, a Go realtime game server, and a Ruby/Rails API server for backend account and platform concerns.

The project is in active development. Expect rough edges while systems, documentation, and tooling continue to move.

## Overview

Space Rocks uses a server-authoritative gameplay model. The Godot client presents the game and sends player input. The Go game server owns realtime gameplay simulation and authoritative match state. The Rails API server owns backend HTTP concerns such as account, auth, and platform services.

This README is the repository front door. For setup, local development workflow, tools, and handoff notes, start with [Developer onboarding](docs/developer.md).

## Repository Layout

```text
client/                 Godot client project
services/game-server/   Go realtime game server
services/api-server/    Ruby/Rails API server
shared/                 Shared source-of-truth data
tools/data_sync/        Shared data validation and generation tooling
bruno-api/              Bruno collection for local API smoke tests
docs/                   Project documentation
```

## Getting Started

Install the main local development tools:

```text
Godot 4.6.x
Go 1.26.x
Ruby / Rails
Python 3.10+
Git LFS
```

After cloning, install and pull Git LFS assets:

```bash
git lfs install
git lfs pull
```

Open the Godot project from:

```text
client/
```

Run the game server from:

```bash
cd services/game-server
go run ./cmd/game-server
```

For the full setup path, local commands, repo tools, and development cautions, use [Developer onboarding](docs/developer.md).

## Documentation

Primary documentation entry points:

* [Developer onboarding](docs/developer.md) - Setup, local workflow, development tools, verification, and handoff notes.
* [Documentation index](docs/!INDEX.md) - Browse the project documentation by area.
* [Documentation policy](docs/documentation-policy.md) - Rules for where documentation belongs.
* [Documentation procedure](docs/documentation-procedure.md) - Workflow for creating, moving, updating, graduating, and deleting docs.

Documentation areas:

* [Agent docs](docs/agent/!INDEX.md) - Agent workflow, testing expectations, MCP usage, and implementation guardrails.
* [Data docs](docs/data/!INDEX.md) - Source-of-truth files, generated outputs, schemas, and data-sync pipelines.
* [Devtools docs](docs/devtools/!INDEX.md) - Debug and development tooling.
* [Domain docs](docs/domains/!INDEX.md) - Cross-system player, platform, and technical flows.
* [Limits docs](docs/limits/!INDEX.md) - Temporary blockers, bugs, and transitional limitations.
* [Planning docs](docs/planning/!INDEX.md) - Future, unresolved, proposed, or not-yet-current work.
* [Protocol docs](docs/protocol/!INDEX.md) - HTTP, WebSocket, packet, and message-flow contracts.
* [Service docs](docs/services/!INDEX.md) - Runtime implementation docs for client, game-server, API server, player-data, and web.
* [Systems-design docs](docs/systems-design/!INDEX.md) - Conceptual mechanics, authority boundaries, invariants, and durable design rules.

## Development Entry Points

Use these docs instead of expanding this README with detailed workflow instructions:

* [Developer onboarding](docs/developer.md) for setup, tools, and handoff.
* [Testing](docs/agent/testing.md) for verification expectations.
* [MCP servers](docs/agent/mcp-servers.md) for Info MCP, Write MCP, and EngineForge/Godot bridge usage.
* [Source-of-truth map](docs/data/source-of-truth-map.md) for ownership questions.
* [Data sync and source-of-truth pipeline](docs/data/data-sync-and-ssot-pipeline.md) for generated outputs and shared data workflows.
* [API-server Bruno smoke tests](docs/devtools/api-server/bruno-smoke-tests.md) for Bruno usage.

## Assets And Git LFS

Source assets and binary game assets are part of the repo workflow. Git LFS is required for asset patterns such as images and audio.

Generated recordings, local editor state, caches, and build artifacts should not be committed.

## License

All rights reserved.
