---
author: brian
created: "2026-07-30"
document_id: f6918c7a-dd86-4ea2-80c9-492e9af63a59
document_type: development
policy_exempt: false
summary: This document maps Space Rocks production ownership boundaries to canonical current documentation.
---
# Documentation Coverage

Parent index: [Development Documentation](./!INDEX.md)

## Purpose

This document maps Space Rocks production ownership boundaries to canonical current documentation.

## Overview

Coverage is organized by runtime, protocol, source-of-truth pipeline, and durable stateful flow. A link is valid only when the owning document explains responsibility, authority, lifecycle, state, failure behavior, non-ownership boundaries, and relevant tests or contracts.

## Runtime coverage

| Production boundary | Code root | Canonical current owner |
| --- | --- | --- |
| Godot client application and session shell | `client/` | [Client](../services/client/!INDEX.md), [App shell and session](../services/client/app-shell-and-session/!INDEX.md) |
| Authoritative Go game server | `services/game-server/` | [Game server](../services/game-server/!INDEX.md), [Game aggregate](../services/game-server/simulation/runtime/game-aggregate.md) |
| Rails API and account-backed persistence | `services/api-server/` | [API server](../services/api-server/!INDEX.md) |
| Local player-data service | `services/player-data/` | [Player data](../services/player-data/!INDEX.md) |
| Diagnostic aggregator | `services/diagnostic-aggregator/` | [Diagnostic aggregator](../services/diagnostic-aggregator/!INDEX.md) |
| Web and devlog surface | `web/` | [Web service](../services/web/!INDEX.md), [Web domain](../domains/web/!INDEX.md) |
| Shared packet, constant, and collision sources | `shared/`, `tools/data_sync/` | [Data](../data/!INDEX.md), [Source-of-truth map](../data/source-of-truth-map.md) |

## Protocol and state coverage

| Durable flow or contract | Canonical current owner |
| --- | --- |
| WebSocket control and session lifecycle | [Realtime WebSocket protocol](../protocol/realtime-websocket-protocol.md), [WebSocket session lifecycle](../services/game-server/networking/websocket-session-lifecycle.md) |
| WebRTC gameplay transport and compact wire state | [WebRTC gameplay transport](../protocol/realtime-webrtc-gameplay-transport.md), [Realtime compact wire mapping](../services/game-server/networking/realtime-compact-wire-mapping.md) |
| Inbound and outbound packet authority | [Gameplay packets](../protocol/gameplay-packets.md), [Game-server inbound routing](../services/game-server/networking/inbound-packet-routing.md), [Client inbound routing](../services/client/networking-flow/inbound-packet-routing.md) |
| Room membership, identity, teams, and match lifecycle | [Rooms](../services/game-server/rooms/!INDEX.md), [Room match lifecycle](../services/game-server/rooms/room-match-lifecycle.md) |
| Authoritative simulation phase order and runtime entity state | [Simulation runtime](../services/game-server/simulation/runtime/!INDEX.md), [Simulation loop](../services/game-server/simulation/runtime/simulation-loop-and-phase-order.md) |
| Player lifecycle, lives, death, respawn, and participation | [Player lifecycle](../services/game-server/simulation/players/player-lifecycle-overview.md), [Player session state](../services/game-server/simulation/players/player-session-state.md) |
| Combat, damage, projectiles, and radial effects | [Combat runtime](../services/game-server/simulation/combat/!INDEX.md), [Combat systems design](../systems-design/combat/!INDEX.md) |
| Match rules, objectives, outcomes, and scoring | [Match simulation](../services/game-server/simulation/match/!INDEX.md), [Scoring](../services/game-server/simulation/scoring/!INDEX.md) |
| Account identity and profile-backed data | [Platform domain](../domains/platform/!INDEX.md), [Auth and OAuth](../services/api-server/auth-and-oauth.md) |
| Observability event contract and emitted diagnostics | [Observability](../observability/!INDEX.md), [Canonical event emission](../observability/canonical-event-emission.md) |
| Generated packet, constant, collision, and observability outputs | [Data pipeline](../data/data-sync-and-ssot-pipeline.md), [Generated protocol](../protocol/generated/!INDEX.md), [Generated observability](../observability/generated/!INDEX.md) |
| Build, release, environment, and deployment behavior | [Developer workflow](../developer.md), [Build and release planning](../planning/domains/technical/build-release-and-environment-matrix.md), [Cloudflare deployment](../services/web/cloudflare-pages-deployment.md) |

## Coverage update rules

Update this map when a production runtime, package group, packet family, persistent model, mutation flow, concurrency boundary, recovery seam, or authority owner is added, removed, renamed, or materially reassigned.

Broad links do not replace focused documentation. When one subsystem contains several independent stateful flows, each flow requires a focused owner elsewhere in the documentation tree.

## Related docs

- [Behavioral contract matrix](behavioral-contract-matrix.md)
- [Documentation policy](../documentation-policy.md)
- [Source-of-truth map](../data/source-of-truth-map.md)
- [Services](../services/!INDEX.md)

## Notes

Planning documents may describe future replacements but do not satisfy current coverage for implemented behavior.
