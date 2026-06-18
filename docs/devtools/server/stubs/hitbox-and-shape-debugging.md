# Hitbox And Shape Debugging
Parent index: [Server Devtools](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe server hitbox and shape debug catalog documentation.

## Overview

TODO: summarize the debug tools that expose hitbox and shape information.
Stub note: keep this focused on debug tooling and diagnostics.

## Debug-only scope

- TODO: define which shape and hitbox tools are debug-only and what they affect.
- Stub note: do not blur into production gameplay behavior.

## Server authority

- TODO: describe which server-owned systems expose shape and hitbox debug data.
- Stub note: keep authority rules conceptual.

## Client presentation

- TODO: describe any client-visible overlays or readouts tied to hitbox and shape debugging.
- Stub note: keep presentation details separate from server debug data generation.

## Commands or controls

- TODO: describe the commands or controls that query or update shape debug state.
- Stub note: this is intentionally incomplete.

## Telemetry

- TODO: describe any shape catalogs, IDs, or debug logs emitted by hitbox and shape tools.
- Stub note: only note verified telemetry surfaces later.

## Build/runtime gates

- `services/game-server/internal/devtools/enabled_default.go`
- `services/game-server/internal/devtools/enabled_nodevtools.go`
- TODO: describe any other build or runtime gates when they are confirmed.

## Code map

- `services/game-server/internal/devtools/shape_catalog.go`
- `services/game-server/internal/devtools/shape_ids.go`
- `services/game-server/internal/devtools/shape_catalog_test.go`
- `services/game-server/internal/devtools/shape_ids_test.go`
- `services/game-server/internal/devtools/player_labels/`
- `services/game-server/internal/devtools/hitboxes/devtools_server_hitbox_overlay.go`
- `services/game-server/internal/devtools/hitboxes/debug_shape_catalog_packet_reader.go`
- `services/game-server/internal/devtools/hitboxes/debug_shape_id_resolver.go`
- TODO: add narrower code links when they are confirmed.

## Tests

- `services/game-server/internal/devtools/shape_catalog_test.go`
- `services/game-server/internal/devtools/shape_ids_test.go`
- `services/game-server/internal/devtools/enabled_default_test.go`
- `services/game-server/internal/devtools/disabled_test.go`
- TODO: add any additional verified tests here.

## Related docs

- [Server Devtools](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add hitbox/shape-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future server hitbox and shape debugging documentation.
Do not treat it as canonical source material.
