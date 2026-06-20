# Data Sync And SSOT Pipeline
Parent index: [Data](./!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe the shared source-of-truth pipeline for generated data and schema sync.

## Overview

TODO: summarize how constants, packets, drop tables, player-data schema, OpenAPI contracts, and generator outputs fit into the sync pipeline.
Stub note: keep this focused on data ownership and pipeline behavior, not feature policy.

## Source files

- `shared/constants/server_constants.toml`
- `shared/constants/server_entities.toml`
- `shared/constants/client/presentation.toml`
- `shared/constants/client/shell.toml`
- `shared/constants/client/lobby.toml`
- `shared/packets/outputs.toml`
- `shared/packets/gameplay.toml`
- `shared/packets/debug.toml`
- `shared/packets/lobby.toml`
- `shared/drop_tables/basicasteroids.toml`
- `shared/player_data/`
- TODO: add any other confirmed source files.

## Configuration

- `tools/data_sync/config.toml`
- `tools/data_sync/!README.md`
- TODO: add any other pipeline configuration roots when they are confirmed.

## Generated outputs

- `client/scripts/generated/constants/constants.gd`
- `services/game-server/internal/constants/constants.go`
- `client/scripts/generated/networking/packets/packets.gd`
- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/runtime/packets_generated.go`
- `services/game-server/internal/devtools/packets_generated.go`
- `services/game-server/internal/game/drops/drop_tables.go`
- TODO: add any other generated outputs when they are confirmed.

## Consumers

- Client runtime and generated client scripts.
- Game-server runtime and generated Go code.
- Player-data service runtime.
- API server runtime.
- TODO: add any other confirmed consumer groups.

## Pipeline usage

- TODO: describe how source files are discovered, validated, and synchronized.
- Stub note: keep this focused on ownership mapping and pipeline handoff.

## Validation commands

- `tools/data_sync -validate`
- `tools/data_sync -check -constants -go -gds`
- `tools/data_sync -check -packets -go -gds`
- `tools/data_sync -check -drop-tables -go`
- TODO: add any other verified validation commands.

## Failure modes

- Duplicate source-of-truth ownership across multiple roots.
- Stale generated outputs after source edits.
- Missing or mismatched data-sync managed blocks.
- TODO: add any other verified failure modes.

## Code or source map

- `tools/data_sync/!README.md`
- `tools/data_sync/data_sync/`
- `docs/data/!README.md`
- `docs/data/stubs/`
- TODO: add narrower source-map entries when they are confirmed.

## Related docs

- [Data](../!README.md)
- [Data Sync](../../../tools/data_sync/!README.md)
- [Source of Truth Map](source-of-truth-map.md)
- TODO: add other related docs when they exist.

## Notes

Stub note: this document is a placeholder for future project-wide data and schema sync documentation.
Do not treat it as canonical source material.
