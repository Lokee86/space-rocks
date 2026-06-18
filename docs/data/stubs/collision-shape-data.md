# Collision Shape Data
Parent index: [Data](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe exported collision shape data documentation.

## Overview

TODO: summarize how collision shape data is authored, exported, and consumed.
Stub note: keep this focused on collision data ownership.

## Source files

- `shared/collisions/collision_shapes.json`
- TODO: add any other confirmed collision data source files.

## Configuration

- TODO: describe any collision export configuration if it exists.
- Stub note: configuration details are not yet confirmed.

## Generated outputs

- `services/game-server/internal/game/physics/collision_shapes.go`
- `services/game-server/internal/game/physics/collision_shapes_test.go`
- `services/game-server/tests/physics/collision_shapes_test.go`
- TODO: add any other generated or exported collision outputs when they are confirmed.

## Consumers

- Game-server physics and collision code.
- Game-server physics and collision tests.
- TODO: add any other confirmed consumers.

## Pipeline usage

- TODO: describe the collision export workflow if it is confirmed.
- Stub note: keep this focused on the data export path.

## Validation commands

- `go test ./services/game-server/tests/physics/...`
- `go test ./services/game-server/internal/game/physics/...`
- TODO: add any other verified collision validation commands.

## Failure modes

- Out-of-sync exported collision shapes.
- Missing or stale collision data after shape edits.
- Invalid collision shape source data.
- TODO: add any other verified failure modes.

## Code or source map

- `shared/collisions/`
- `services/game-server/internal/game/physics/`
- `services/game-server/tests/physics/`
- TODO: add narrower collision data source-map entries when they are confirmed.

## Related docs

- [Data](../!README.md)
- TODO: add collision-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future collision shape data documentation.
Do not treat it as canonical source material.
