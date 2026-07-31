---
author: brian
created: "2026-07-31"
document_id: 3f46bce4-d10a-489c-a78f-f81927be312e
document_type: general
policy_exempt: false
summary: Records the decision that the Go game server owns authoritative gameplay state and outcomes while clients own intent collection and presentation.
---
# ADR-0001: Server-Authoritative Gameplay

Parent index: [Architecture Decisions](!INDEX.md)

## Purpose

This record explains why the Go game server owns authoritative gameplay state and outcomes while clients own intent collection and presentation.

## Overview

One authoritative simulation owner prevents clients, transports, and presentation layers from creating conflicting gameplay truth. The server accepts or rejects intent and projects results; the client renders those results defensively.

## Status

Accepted.

## Context

Space Rocks supports single-player and multiplayer through the same room, session, simulation, and realtime protocol architecture. Clients collect input and render state over a network where packets can be delayed, reordered, duplicated, dropped, or deliberately forged. Allowing each client to decide movement, collisions, damage, score, lives, respawn validity, targeting, pause, or match outcomes would create conflicting sources of truth and make multiplayer integrity impossible to preserve.

## Decision

The Go game server owns authoritative live gameplay state and outcomes. Client packets express requests, intent, configuration, or observations. The server validates those packets against session, room, active-player, match, and gameplay rules before mutating simulation state.

The Godot client owns local input collection, transport operation, defensive packet acceptance, read-model application, rendering, HUD, audio, effects, menus, and navigation. It does not infer acceptance from a sent packet or recalculate authoritative outcomes.

## Consequences

### Ownership and dependencies

- Gameplay policy and mutation remain under game-server simulation owners.
- Networking routes identity and intent but does not become gameplay authority.
- Client presentation depends on server-projected state and transient presentation events.
- Debug tooling must route through server-owned devtools and gameplay seams rather than mutating client state as authority.

### State, lifecycle, and operations

- WebSocket connection, authentication, room membership, active gameplay participation, and player-facing identity remain distinct states.
- Disconnect, reconnection, resync, and packet-loss behavior must preserve server authority.
- Client-side prediction, if introduced, remains reversible presentation and cannot become authoritative mutation.

### Compatibility and migration

- New gameplay systems must establish their server owner before client presentation is added.
- Moving an outcome to client authority requires a superseding ADR, anti-cheat and reconciliation design, protocol migration, and new behavioral contracts.

## Alternatives Considered

- **Client-authoritative gameplay:** rejected because conflicting clients and untrusted packets cannot produce one reliable multiplayer truth.
- **Shared client/server authority:** rejected because ambiguous mutation ownership produces desync and bypass paths.
- **Server authority only for competitive multiplayer:** rejected because separate single-player mechanics would duplicate simulation ownership and drift.

## Verification

- Game-server simulation and room lifecycle tests.
- Client tests proving inbound authoritative state drives presentation.
- Protocol and runtime scenarios covering invalid identity, stale match packets, resync, death, respawn, score, and match outcomes.
- Narrow Pitlord rules where a concrete dependency bypass is statically detectable.

## Risks and Debt

- Poor recovery or interest management can make correct server state appear frozen or missing on clients; observability and resync behavior remain critical.
- Client prediction may add reconciliation complexity later.

## Related docs

- [Realtime client-server flow](../domains/technical/realtime-client-server-flow.md)
- [Source-of-truth flow](../domains/technical/source-of-truth-flow.md)
- [Game-server simulation](../services/game-server/simulation/!INDEX.md)
- [Gameplay state application](../services/client/gameplay-runtime/gameplay-state-application.md)

## Notes

A client-rendered fact is not authoritative merely because it is visible. The server-owned state and accepted transition remain the source of truth.
