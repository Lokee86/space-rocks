# Packet Schema Pipeline
Parent index: [Data](./!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe packet schema source files and generated packet outputs.

## Overview

TODO: summarize how packet schema TOML becomes generated packet outputs for client and server code.
Stub note: keep this focused on packet schema flow.

## Source files

- `shared/packets/outputs.toml`
- `shared/packets/gameplay.toml`
- `shared/packets/debug.toml`
- `shared/packets/lobby.toml`
- `shared/packets/player_data.toml`
- TODO: add any other confirmed packet schema source files.

## Configuration

- `tools/data_sync/config.toml`
- `tools/data_sync/!README.md`
- TODO: add any other packet pipeline configuration roots when they are confirmed.

## Generated outputs

- `client/scripts/generated/networking/packets/packets.gd`
- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/runtime/packets_generated.go`
- `services/game-server/internal/devtools/packets_generated.go`
- TODO: add any other generated packet outputs when they are confirmed.

## Consumers

- Client networking and packet readers.
- Game-server networking and packet dispatch.
- TODO: add any other confirmed consumers.

## Pipeline usage

- `tools/data_sync -push -packets -go -gds`
- `tools/data_sync -check -packets -go -gds`
- `tools/data_sync -validate -packets`
- TODO: add any other verified packet pipeline commands.

## Validation commands

- `tools/data_sync -validate -packets`
- `tools/data_sync -check -packets -go -gds`
- `tools/data_sync -diff -packets -go -gds`
- TODO: add any other verified validation commands.

## Failure modes

- Missing or conflicting packet sections across source files.
- Stale generated packet outputs after schema edits.
- Invalid packet schema or unsupported packet routing configuration.
- TODO: add any other verified failure modes.

## Code or source map

- `shared/packets/`
- `tools/data_sync/data_sync/packets_sync.py`
- `tools/data_sync/data_sync/packet_toml.py`
- `tools/data_sync/data_sync/packet_rendering.py`
- `tools/data_sync/data_sync/generators/go_packets.py`
- `tools/data_sync/data_sync/generators/gds_packets.py`
- `tools/data_sync/data_sync/generators/ts_packets.py`
- TODO: add narrower packet pipeline source-map entries when they are confirmed.

## Related docs

- [Data](../!README.md)
- [Data Sync](../../../tools/data_sync/!README.md)
- TODO: add packet-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future packet schema pipeline documentation.
Do not treat it as canonical source material.
