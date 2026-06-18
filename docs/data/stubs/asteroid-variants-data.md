# Asteroid Variants Data
Parent index: [Data](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe asteroid variant data source documentation.

## Overview

TODO: summarize how asteroid variant data is authored and consumed by gameplay.
Stub note: keep this focused on asteroid data ownership.

## Source files

- `shared/asteroids/variants.toml`
- TODO: add any other confirmed asteroid variant source files.

## Configuration

- TODO: describe any variant generation configuration if it exists.
- Stub note: configuration details are not yet confirmed.

## Generated outputs

- `services/game-server/internal/game/asteroids/variants.go`
- `services/game-server/internal/game/asteroids/variants_test.go`
- `services/game-server/internal/game/simulation_asteroids.go`
- TODO: add any other generated or exported asteroid outputs when they are confirmed.

## Consumers

- Game-server asteroid simulation.
- Game-server asteroid tests.
- TODO: add any other confirmed consumers.

## Pipeline usage

- TODO: describe the asteroid variant export workflow if it is confirmed.
- Stub note: keep this focused on the data export path.

## Validation commands

- `go test ./services/game-server/internal/game/asteroids/...`
- `go test ./services/game-server/internal/game/...`
- TODO: add any other verified asteroid validation commands.

## Failure modes

- Out-of-sync asteroid variant data.
- Missing or stale variant outputs after source edits.
- Invalid asteroid variant source data.
- TODO: add any other verified failure modes.

## Code or source map

- `shared/asteroids/`
- `services/game-server/internal/game/asteroids/`
- `services/game-server/internal/game/simulation_asteroids.go`
- TODO: add narrower asteroid data source-map entries when they are confirmed.

## Related docs

- [Data](../!README.md)
- TODO: add asteroid-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future asteroid variants data documentation.
Do not treat it as canonical source material.
