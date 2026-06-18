# Telemetry
Parent index: [Server Devtools](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe server devtools telemetry documentation.

## Overview

TODO: summarize the server-side devtools telemetry surfaces and what they report.
Stub note: keep this focused on diagnostics and internal visibility.

## Debug-only scope

- TODO: define which telemetry is debug-only and which runtime areas it observes.
- Stub note: do not blur into production gameplay policy.

## Server authority

- TODO: describe which server-owned systems emit telemetry and how authority is preserved.
- Stub note: keep authority rules conceptual.

## Client presentation

- TODO: describe any client-visible telemetry views, overlays, or readouts tied to server devtools.
- Stub note: keep presentation details separate from server telemetry generation.

## Commands or controls

- TODO: describe any commands or controls that query or drive telemetry on the server.
- Stub note: this is intentionally incomplete.

## Telemetry

- TODO: describe status snapshots, shape catalogs, player counters, or other telemetry data that is exposed.
- Stub note: only note verified telemetry surfaces later.

## Build/runtime gates

- `services/game-server/internal/devtools/enabled_default.go`
- `services/game-server/internal/devtools/enabled_nodevtools.go`
- TODO: describe any other build or runtime gates when they are confirmed.

## Code map

- `services/game-server/internal/devtools/status.go`
- `services/game-server/internal/devtools/packets_generated.go`
- `services/game-server/internal/devtools/shape_catalog.go`
- `services/game-server/internal/devtools/player_counters.go`
- `services/game-server/internal/devtools/streamruntime/`
- TODO: add narrower code links when they are confirmed.

## Tests

- `services/game-server/internal/devtools/shape_catalog_test.go`
- `services/game-server/internal/devtools/player_counters_test.go`
- `services/game-server/internal/devtools/command_types_test.go`
- `services/game-server/internal/devtools/enabled_default_test.go`
- `services/game-server/internal/devtools/disabled_test.go`
- TODO: add any additional verified tests here.

## Related docs

- [Server Devtools](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add telemetry-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future server devtools telemetry documentation.
Do not treat it as canonical source material.
