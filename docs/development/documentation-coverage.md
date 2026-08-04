---
author: brian
created: "2026-08-04"
document_id: b17ca694-dc3f-49f2-9e3e-a114bd9de747
document_type: general
policy_exempt: false
summary: Maps Space Rocks production boundaries, executables, stateful flows, persistence models, contracts, and recovery seams to canonical current documentation.
---
# Documentation Coverage

Parent index: [Development](./!INDEX.md)

## Purpose

This document maps production implementation boundaries to canonical current documentation.

## Overview

Coverage is organized by responsibility rather than file count. Small packages may share a focused subsystem owner, but independent executables, stateful flows, persistent models, public contracts, and recovery seams require explicit current owners.

A coverage entry means the linked document must explain the boundary in prose. This table is navigation and audit evidence; it is not a substitute for the owning document.

## Repository and runtime coverage

| Production boundary | Canonical current owner | Coverage notes |
| --- | --- | --- |
| Godot application entry, shutdown, viewport, and session boot | [Client app shell and session](../services/client/app-shell-and-session/!INDEX.md) | Covers application composition, network target selection, hosted connectivity, shutdown, and room-session state. |
| Client authentication and credential persistence | [Client auth session flow](../services/client/auth-session-flow.md) | Covers login, OAuth handoff, bearer storage, logout, and account identity. |
| Client HTTP API helpers and request results | [Client HTTP API flow](../services/client/client-http-api-flow.md) | Covers public API wrappers and failure result ownership. |
| Client pregame, menu, lobby, gameplay menu, spectate, and match-end workflows | [Client service index](../services/client/!INDEX.md) | Routed to focused sub-indexes and workflow owners. |
| Client gameplay runtime, world sync, and presentation bridges | [Client gameplay runtime](../services/client/gameplay-runtime/!INDEX.md) | Covers applied state, runtime composition, event presentation, and entity synchronization. |
| Client logging and local rolling archives | [Client logging](../services/client/client-logging.md) | Covers event envelopes, local JSONL output, maintenance, and failure isolation. |
| Client developer tools and runtime scenarios | [Client devtools](../devtools/client/!INDEX.md) | Covers debug-only controls, overlays, telemetry, and scenario presentation. |
| Game-server executable and process lifecycle | [Game-server process](../services/game-server/process/!INDEX.md) | Covers startup, routes, shutdown, runtime configuration, and hosted integrations. |
| WebSocket/WebRTC session lifecycle and routing | [Game-server networking](../services/game-server/networking/!INDEX.md) | Covers admission, inbound/outbound flow, backpressure, transport, and resync. |
| Realtime projection, scheduling, quantization, prioritization, and baselines | [Realtime protocol architecture](../planning/protocol/realtime-protocol-architecture.md) and [current protocol docs](../protocol/!INDEX.md) | Current exact wire behavior is owned by protocol and service docs; unresolved later phases remain in planning. |
| Room manager, membership, lobby, match lifecycle, and cleanup | [Game-server rooms](../services/game-server/rooms/!INDEX.md) | Covers room state transitions, ownership, snapshots, cleanup, and reporting handoff. |
| Authoritative gameplay aggregate and simulation loop | [Game aggregate](../services/game-server/simulation/runtime/game-aggregate.md) and [simulation loop](../services/game-server/simulation/runtime/simulation-loop-and-phase-order.md) | Covers state ownership, phase order, deterministic RNG, entity store, and event queue. |
| Players, lifecycle, counters, respawn, pause, camera, and input | [Player simulation](../services/game-server/simulation/players/!INDEX.md) | Covers durable session state and active avatar lifecycle. |
| Combat, weapons, collision damage, and radial effects | [Combat simulation](../services/game-server/simulation/combat/!INDEX.md) | Covers request/resolution ownership, projectile fire, and radial coverage. |
| Pickups, drops, effects, and lifecycle | [Pickup simulation](../services/game-server/simulation/pickups/!INDEX.md) | Covers collection, drops, effects, and entity lifetime. |
| World physics, spawning, visibility, spatial index, and toroidal motion | [World simulation](../services/game-server/simulation/world/!INDEX.md) | Covers world-state mechanics and spatial/recovery boundaries. |
| Scoring and gameplay awards | [Scoring policy and awards](../services/game-server/simulation/scoring/scoring-policy-and-awards.md) | Covers award requests, counters, and score policy. |
| Canonical targeting | [Targeting](../services/game-server/simulation/targeting/!INDEX.md) | Covers target identity, selection, status, and consumers. |
| Server developer tooling | [Server devtools](../devtools/server/!INDEX.md) | Covers debug authority, command routing, telemetry, spawn/control adapters, and gates. |
| Rails API executable, health, accounts, OAuth, internal APIs, and match results | [API server](../services/api-server/!INDEX.md) | Covers runtime, auth, observability, and account-backed result persistence. |
| Player-data codec, store routing, local SQLite, guest memory, Rails store, and HTTP handlers | [Player data](../services/player-data/!INDEX.md) | Covers persistence modes, dispatch, APIs, observability, and result sinks. |
| Diagnostic-aggregator executable and hosted adapter | [Diagnostic aggregator](../services/diagnostic-aggregator/!INDEX.md) | Must cover both standalone and in-process composition modes without merging ownership. |
| Diagnostic report API, validation, redaction, construction, storage, recovery, rotation, and retention | [Diagnostic aggregator](../services/diagnostic-aggregator/!INDEX.md) | Focused service owners cover report flow and the JSONL store lifecycle. |
| Production Compose topology and deployment verification | [Operations](../operations/!INDEX.md) | Covers deploy bundle, environment, ports, volumes, health, update, rollback, and recovery. |
| Website, devlog content, Plasmic integration, and Cloudflare Pages | [Web service](../services/web/!INDEX.md) | Covers static site architecture, content, and deployment. |
| Shared constants and generated outputs | [Constants](../data/constants.md) | Covers source TOML, generation, consumers, and validation. |
| Packet schemas and generated packet code | [Packet schemas](../data/packet-schemas.md) | Covers source files, generated outputs, routing, and checks. |
| Realtime compact-wire physical contract | [Realtime compact wire mapping](../services/game-server/networking/realtime-compact-wire-mapping.md) and [generated reference](../protocol/generated/realtime-wire-reference.md) | Schema owns exact encoding; runtime applies generated descriptors. |
| Collision-shape source and generated consumers | [Collision-shape data](../data/collision-shape-data.md) | Covers source, import/export, consumers, and debugging. |
| Drop-table source and generated consumers | [Drop tables](../data/drop-tables.md) | Covers source, generation, validation, and runtime consumption. |
| HTTP OpenAPI contract and enforcement | [HTTP API contracts](../protocol/http-api-contracts.md) and [contract enforcement](../protocol/http-contract-enforcement.md) | Covers exact public/internal request-response shapes and tests. |
| Observability event contract and generated bindings | [Observability contract](../data/observability-contract.md) | Covers event source, generated bindings, consumers, and canonical emission. |
| Data-sync pipeline | [Data-sync and source-of-truth pipeline](../data/data-sync-and-ssot-pipeline.md) | Covers commands, configuration, generation, validation, and failures. |
| Repository checks and documentation enforcement | [Developer guide](../developer.md), [documentation policy](../documentation-policy.md), and this coverage map | Covers CI, shared checker, Pitlord, data validation, and test gates. |

## Coverage review rules

Update this map when a production package, executable, public command family, independent stateful flow, mutation boundary, persistence model, concurrency boundary, machine-readable contract, or recovery seam is added, removed, renamed, or reassigned.

A broad owner must be split when it can no longer explain independent state transitions and failure boundaries clearly. Planning documentation may link to this map but does not satisfy current implementation coverage.

## Related docs

- [Maintainer map](../maintainer-map.md)
- [Behavioral-contract matrix](behavioral-contract-matrix.md)
- [Documentation audit](documentation-audit.md)
- [Documentation policy](../documentation-policy.md)

## Notes

The P4 feature branches are not treated as current `main` coverage until their implementation merges. Their planning documents remain separate and must graduate affected facts during integration.
