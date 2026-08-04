---
author: brian
created: "2026-08-04"
document_id: 37e2d24f-2a4d-431f-bbcc-16456ed2ec4d
document_type: general
policy_exempt: false
summary: Routes Space Rocks change areas to their canonical documentation, primary implementation boundaries, and verification owners.
---
# Space Rocks Maintainer Map

Parent index: [Documentation](./!INDEX.md)

## Purpose

This document routes maintainers to the canonical documentation and primary implementation boundary for common Space Rocks changes.

## Overview

Use this map when the owner of a change is unclear. It routes by maintainer intent rather than repeating the repository tree. Detailed behavior, contracts, state, and failure handling remain in the linked canonical documents.

## Change-area routing

| Change area | Canonical documentation | Primary implementation boundary | Verification |
| --- | --- | --- | --- |
| Client startup, shutdown, viewport, and network target selection | [Client app shell and session](services/client/app-shell-and-session/!INDEX.md) | `client/scripts/boot/`, `client/scripts/config/`, `client/scripts/shell/` | Client GUT tests and packaged-client gates |
| Client authentication and bearer credential handling | [Client auth session flow](services/client/auth-session-flow.md) | `client/scripts/auth/`, `client/scripts/api/` | Client auth tests and API contract tests |
| Client menus, local pilot, lobby, and match-result presentation | [Client service index](services/client/!INDEX.md) | `client/scripts/ui/`, `client/scripts/main_menu/`, `client/scripts/lobby/` | Client GUT tests |
| Client gameplay state application and presentation | [Client gameplay runtime](services/client/gameplay-runtime/!INDEX.md) | `client/scripts/gameplay/`, `client/scripts/world/` | Client gameplay and world-sync tests |
| Client developer tooling | [Client devtools](devtools/client/!INDEX.md) | `client/scripts/devtools/`, `client/scenes/devtools/` | Client devtools tests and runtime scenarios |
| Game-server process startup, shutdown, routes, and runtime configuration | [Game-server process](services/game-server/process/!INDEX.md) | `services/game-server/cmd/game-server/` | Command-package tests and deployed runtime scenarios |
| Realtime sessions, WebSocket/WebRTC transport, packet routing, and backpressure | [Game-server networking](services/game-server/networking/!INDEX.md) | `services/game-server/internal/networking/`, `services/game-server/internal/protocol/` | Networking and protocol tests |
| Rooms, membership, lobby rules, and match lifecycle | [Game-server rooms](services/game-server/rooms/!INDEX.md) | `services/game-server/internal/rooms/` | Room and networking integration tests |
| Authoritative gameplay simulation | [Game-server simulation](services/game-server/simulation/!INDEX.md) | `services/game-server/internal/game/` | Game package and integration tests |
| Server developer tooling and game-control adapters | [Server devtools](devtools/server/!INDEX.md) | `services/game-server/internal/devtools/`, game control seams | Devtools tests and runtime scenarios |
| Authenticated accounts, OAuth, API health, and match-result persistence | [API server](services/api-server/!INDEX.md) | `services/api-server/` | Rails controller, service, model, and contract tests |
| Guest, local-profile, and account-backed player-data routing | [Player data](services/player-data/!INDEX.md) | `services/player-data/` | Player-data Go tests and API-server integration tests |
| Diagnostic report intake, validation, storage, and retrieval | [Diagnostic aggregator](services/diagnostic-aggregator/!INDEX.md) | `services/diagnostic-aggregator/` | Diagnostic-aggregator Go tests and Bruno smoke tests |
| Public HTTP, lobby, gameplay, and realtime wire contracts | [Protocol documentation](protocol/!INDEX.md) | `shared/contracts/`, `shared/packets/`, protocol adapters | Contract, generator, server, and client protocol tests |
| Constants, packet schemas, collision data, drop tables, and generated outputs | [Data documentation](data/!INDEX.md) | `shared/`, `tools/data_sync/` | Data-sync validation and generator tests |
| Deployment, production topology, image publishing, update, and rollback | [Operations](operations/!INDEX.md) | `deploy/production/`, release workflows, service Dockerfiles | Compose validation, deployment verification, release gates |
| Website and devlog | [Web service](services/web/!INDEX.md) | `web-astro/` | Web build and deployment checks |
| Cross-system product flows | [Domain documentation](domains/!INDEX.md) | Multiple participating systems | Participating service and protocol tests |
| Gameplay mechanics and durable authority rules | [Systems design](systems-design/!INDEX.md) | Server authority plus client presentation owners | Behavioral-contract matrix and focused tests |
| Future work and unresolved decisions | [Planning](planning/!INDEX.md) | No current implementation owner until shipped | Acceptance criteria in the owning plan |
| Current defects and accepted ceilings | [Limits](limits/!INDEX.md) | Affected current owners | Reproduction and regression tests where applicable |
| Documentation governance and coverage | [Documentation policy](documentation-policy.md) | `docs/`, `docs-standard.json`, `.standards/` | Shared documentation checker and Pitlord documentation policy |

## Boundaries

- The Go game server owns authoritative gameplay outcomes; the Godot client owns input collection and presentation.
- Rails owns authenticated accounts and account-backed persistence; the game server consumes those boundaries through explicit APIs.
- Player-data owns profile/store routing and does not import game-server implementation packages.
- Diagnostic-aggregator owns bounded diagnostic reports and remains independent whether run standalone or through its hosted adapter.
- Shared TOML, JSON, and contract sources own generated constants and packet/reference outputs; generated files are not hand-edited.
- Deployment and release mechanics are operational owners, not substitutes for service architecture documents.
- Planning and research never serve as the sole owner of implemented behavior.

## Related docs

- [Documentation coverage](development/documentation-coverage.md)
- [Behavioral-contract matrix](development/behavioral-contract-matrix.md)
- [Documentation policy](documentation-policy.md)
- [Developer guide](developer.md)

## Notes

Component indexes provide the next navigation layer. Focused code maps remain inside the current documents that own each behavior.
