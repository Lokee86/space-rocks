# Source Of Truth Map
Parent index: [Data](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe project-wide data and source-of-truth ownership.

## Overview

TODO: summarize the major source-of-truth domains, their owning files, and their generated outputs.
Stub note: keep this focused on data ownership rather than feature policy.

## Source files

- `docs/design/source-of-truth-map.md`
- `shared/constants/`
- `shared/packets/`
- `shared/drop_tables/`
- `services/player-data/data/`
- `services/api-server/db/`
- TODO: add any other confirmed project source-of-truth roots.

## Configuration

- `tools/data_sync/config.toml`
- TODO: add any other data pipeline configuration roots when they are confirmed.

## Generated outputs

- `client/scripts/generated/constants/constants.gd`
- `services/game-server/internal/constants/constants.go`
- `client/scripts/generated/networking/packets/packets.gd`
- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/runtime/packets_generated.go`
- `services/game-server/internal/devtools/packets_generated.go`
- TODO: add other generated outputs when they are confirmed.

## Consumers

- Client runtime and generated client scripts.
- Game-server runtime and generated Go code.
- Player-data service runtime.
- API server runtime.
- TODO: add any other confirmed consumer groups.

## Pipeline usage

- TODO: describe how source-of-truth roots are discovered, validated, and synchronized.
- Stub note: keep this focused on ownership mapping and pipeline handoff.

## Validation commands

- `tools/data_sync -validate`
- `tools/data_sync -check -constants -go -gds`
- `tools/data_sync -check -packets -go -gds`
- TODO: add any other verified validation commands.

## Failure modes

- Duplicate source-of-truth ownership across multiple roots.
- Stale generated outputs after source edits.
- Missing or mismatched data-sync managed blocks.
- TODO: add any other verified failure modes.

## Code or source map

- `tools/data_sync/!README.md`
- `tools/data_sync/data_sync/`
- `docs/design/source-of-truth-map.md`
- `docs/data/!README.md`
- `docs/data/stubs/`
- TODO: add narrower source-map entries when they are confirmed.

## Related docs

- [Data](../!README.md)
- [Documentation procedure](../../documentation-procedure.md)
- [Data Sync](../../../tools/data_sync/!README.md)
- TODO: add source-of-truth-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future project-wide source-of-truth documentation.
Do not treat it as canonical source material.
