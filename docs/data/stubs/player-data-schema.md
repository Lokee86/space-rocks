# Player Data Schema
Parent index: [Data](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe player-data schema and source documentation.

## Overview

TODO: summarize the player-data schema sources and the current logical-schema ownership.
Stub note: keep this focused on source-of-truth schema behavior.

## Source files

- `shared/player_data/match_result.toml`
- `shared/player_data/stats.toml`
- TODO: add any other confirmed player-data schema source files.

## Configuration

- `tools/data_sync/config.toml`
- `tools/data_sync/!README.md`
- TODO: describe any player-data schema pipeline configuration when it is confirmed.

## Generated outputs

- `services/player-data/protocol/packets.go`
- `services/player-data/data/player-data.sqlite3`
- `services/game-server/services/player-data/data/player-data.sqlite3`
- TODO: add any other generated or persisted outputs when they are confirmed.

## Consumers

- Player-data service runtime.
- API server HTTP surfaces.
- Game-server player-data integrations.
- TODO: add any other confirmed consumers.

## Pipeline usage

- TODO: describe how the player-data schema is synced or validated if the pipeline is confirmed.
- Stub note: keep this focused on schema source usage.

## Validation commands

- `tools/data_sync -validate`
- `tools/data_sync -check -packets -go -gds`
- TODO: add any other verified player-data schema validation commands.

## Failure modes

- Missing or inconsistent player-data schema sources.
- Stale generated or persisted outputs after schema edits.
- Invalid schema shape or unsupported field changes.
- TODO: add any other verified failure modes.

## Code or source map

- `shared/player_data/`
- `services/player-data/playerdata/`
- `services/player-data/httpapi/`
- `services/api-server/app/controllers/internal/player_data/`
- `services/api-server/app/controllers/api/internal/player_data/`
- TODO: add narrower player-data schema source-map entries when they are confirmed.

## Related docs

- [Data](../!README.md)
- [Data Sync](../../../tools/data_sync/!README.md)
- TODO: add player-data-schema-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future player-data schema documentation.
Do not treat it as canonical source material.
