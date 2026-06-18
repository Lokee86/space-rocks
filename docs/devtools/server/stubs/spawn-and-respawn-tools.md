# Spawn And Respawn Tools
Parent index: [Server Devtools](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe server devtools spawn, respawn, clear, and placement command documentation.

## Overview

TODO: summarize the debug tools that spawn, respawn, clear, or place entities.
Stub note: keep this focused on debug tooling and diagnostics.

## Debug-only scope

- TODO: define which spawn and respawn tools are debug-only and what they affect.
- Stub note: do not blur into production gameplay behavior.

## Server authority

- TODO: describe which server-owned systems execute spawn and respawn commands.
- Stub note: keep authority rules conceptual.

## Client presentation

- TODO: describe any client-visible effects or UI readouts tied to spawn and respawn tools.
- Stub note: keep presentation details separate from server command behavior.

## Commands or controls

- TODO: describe spawn, respawn, clear, placement, and related debug controls.
- Stub note: this is intentionally incomplete.

## Telemetry

- TODO: describe any logs, status messages, or counters emitted by spawn and respawn commands.
- Stub note: only note verified telemetry surfaces later.

## Build/runtime gates

- `services/game-server/internal/devtools/enabled_default.go`
- `services/game-server/internal/devtools/enabled_nodevtools.go`
- TODO: describe any other build or runtime gates when they are confirmed.

## Code map

- `services/game-server/internal/devtools/spawn_entity.go`
- `services/game-server/internal/devtools/spawn_player.go`
- `services/game-server/internal/devtools/spawn_asteroid.go`
- `services/game-server/internal/devtools/spawn_bullet.go`
- `services/game-server/internal/devtools/spawn_pickup.go`
- `services/game-server/internal/devtools/respawn_handler.go`
- `services/game-server/internal/devtools/respawn_player.go`
- `services/game-server/internal/devtools/placement_requests.go`
- `services/game-server/internal/devtools/clear_entities.go`
- TODO: add narrower code links when they are confirmed.

## Tests

- `services/game-server/internal/devtools/clear_entities_test.go`
- `services/game-server/internal/devtools/enabled_default_test.go`
- `services/game-server/internal/devtools/disabled_test.go`
- TODO: add any additional verified tests here.

## Related docs

- [Server Devtools](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add spawn/respawn-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future server devtools spawn and respawn tools documentation.
Do not treat it as canonical source material.
