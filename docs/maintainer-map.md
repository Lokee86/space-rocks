---
author: brian
created: "2026-07-30"
document_id: 3ca84732-93ba-4598-864b-2a69544bc8d8
document_type: architecture
policy_exempt: false
summary: Routes common Space Rocks changes to canonical documentation and primary implementation boundaries.
---
# Space Rocks Maintainer Map

Parent index: [Documentation](!INDEX.md)

## Purpose

This document routes maintainers to the canonical documentation and primary implementation boundary for common Space Rocks changes.

## Overview

Use this map when the owning service, domain, protocol, data source, or verification surface is unclear. It is not a repository inventory and does not replace focused service documents or their code maps.

## Change-area routing

| Change area | Canonical documentation | Primary implementation boundary | Verification |
| --- | --- | --- | --- |
| Client boot, session, menus, input, HUD, and presentation | [Client services](services/client/!INDEX.md) | `client/` | Godot client tests and runtime smoke checks |
| Client networking, packet routing, world sync, and hosted connectivity | [Client networking flow](services/client/networking-flow/!INDEX.md), [Client world sync](services/client/world-sync/!INDEX.md), [Protocols](protocol/!INDEX.md) | `client/` | Client networking tests and multiplayer smoke scenarios |
| Game-server startup, routes, shutdown, rooms, and membership | [Game-server process](services/game-server/process/!INDEX.md), [Rooms](services/game-server/rooms/!INDEX.md) | `server/` | Go server tests and room lifecycle scenarios |
| Authoritative simulation, players, combat, encounters, scoring, and world | [Game-server simulation](services/game-server/simulation/!INDEX.md), [Systems design](systems-design/!INDEX.md) | `server/internal/game/` | Go simulation tests and runtime scenarios |
| API authentication, profiles, inventory, stats, and persistence | [API server](services/api-server/!INDEX.md), [Platform domains](domains/platform/!INDEX.md) | `api-server/` | Rails tests and HTTP smoke flows |
| Local player-data service and storage routing | [Player data](services/player-data/!INDEX.md), [Player-data protocol](protocol/player-data-http-api.md) | `services/player-data/`, `server/` integrations | Go service tests and HTTP contract checks |
| Realtime packets, HTTP contracts, WebRTC, and WebSocket behavior | [Protocol documentation](protocol/!INDEX.md), [Packet schemas](data/packet-schemas.md) | `shared/`, `server/`, `client/`, `api-server/` | Protocol generators, contract tests, and deployed connectivity checks |
| Source-of-truth data, constants, generated outputs, and data sync | [Data documentation](data/!INDEX.md), [Source-of-truth map](data/source-of-truth-map.md) | `shared/`, `tools/data_sync/` | `data-sync` and generated-artifact checks |
| Logging, diagnostics, telemetry, and diagnostic aggregation | [Observability](observability/!INDEX.md), [Devtools](devtools/!INDEX.md), [Diagnostic aggregator](services/diagnostic-aggregator/!INDEX.md) | `services/log-aggregator/`, service logging adapters, `client/` devtools | Observability contracts and runtime measurement scenarios |
| Build, packaging, deployment, environment, and release | [Developer workflow](developer.md), [Technical planning](planning/domains/technical/!INDEX.md) | `.github/workflows/`, deployment files, installer and release scripts | CI, packaged-client, compose, and server smoke checks |
| Product rules, authority, and future gameplay or platform work | [Domains](domains/!INDEX.md), [Systems design](systems-design/!INDEX.md), [Planning](planning/!INDEX.md) | Owning client, server, API, shared, or web boundary | Behavioral-contract matrix and owning subsystem tests |

## Boundaries

- The Go game server owns authoritative multiplayer simulation; the Godot client owns input and presentation.
- Rails owns account-backed web/API persistence; local player-data paths remain separately documented.
- `shared/` and the data-sync pipeline own cross-runtime source data and generated contracts, not runtime behavior.
- Planning documents describe future work only; current implementation routes point to `docs/services/`, `docs/protocol/`, `docs/data/`, `docs/domains/`, or `docs/systems-design/`.

## Related docs

- [Documentation coverage](development/documentation-coverage.md)
- [Source-of-truth flow](domains/technical/source-of-truth-flow.md)
- [Behavioral contract matrix](development/behavioral-contract-matrix.md)

## Notes

Update this map when a service, domain, protocol owner, source-of-truth boundary, or major verification surface changes.
