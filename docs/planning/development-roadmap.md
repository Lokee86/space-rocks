# Platform And Progression Roadmap
Parent index: [Planning](./!INDEX.md)

## Opening Context

The menu/profile/local-pilot/match-results/stats-refresh vertical slice is complete and green. That is an important milestone, but it also marks a shift in how the project needs to be planned.

We have reached the point where isolated feature slices are no longer enough on their own. Future work will overlap more across auth, observability, packet strategy, progression, enemies, bullet hell, ship variants, unlocks, and account surfaces.

Because those systems now intersect more tightly, seams and ownership boundaries need to be planned before more feature growth lands. The goal is to keep the codebase scalable and avoid coupling the next wave of systems into the wrong layers.

## Roadmap Role

This roadmap coordinates phase order, dependency order, and decision gates for the platform and progression track.

Detailed system ownership belongs in the system-specific planning docs linked from `docs/planning/!INDEX.md`.

## Roadmap Purpose

This document coordinates the larger phases that follow the completed menu-flow vertical slice. It frames the work that needs to happen next, highlights the dependencies between the major areas, and sets up the order in which the platform and progression systems should be planned.

## Major Planning Pillars

- Network budget and runtime observability
- Realtime protocol architecture and state delivery strategy
- Auth and account surface completion
- Player experience structure and system boundaries
- Progression, rewards, unlocks, and player-data contracts
- Gameplay expansion

## Current Completed Baseline

- Discord auth works through the Godot browser handoff.
- Rails owns bearer token issuance and `/api/auth/me`.
- The Go game-server verifies tokens through Rails for websocket auth.
- Local Profile and Authenticated Account stats route through the player-data runtime.
- Match results and stats refresh are complete.
- The World Telemetry Overlay exists behind devtools.
- Gameplay packets are known to exceed 4KB at times before enemies or bullet hell exist.

## Phase Order

- Phase A - Network Budget + Runtime Observability Foundation
- Phase B - Realtime Protocol Architecture
- Phase C - Player Experience Foundation
- Phase D - Progression Foundation
- Phase E - Gameplay Expansion

Phase A determines the priority and order of Phase B work, not whether launch-grade realtime protocol work is expected.

## Phase A

Phase A remains the packet-budget and observability gate for later realtime protocol work. The detailed packet budget, diagnostics, devtools visibility, completion criteria, and decision gate now live in [network-observability-and-packet-budget.md](domains/technical/network-observability-and-packet-budget.md).

## Phase B

Phase B remains the realtime protocol seam for authoritative multiplayer state delivery. The detailed lanes, snapshot model, quantization path, and protobuf target now live in [realtime-protocol-architecture.md](protocol/realtime-protocol-architecture.md).

## Phase C - Player Experience Foundation

Phase C plans the player-facing game structure before progression persistence. Phase C starts with preset-driven room mode configuration.

### Purpose

Phase C defines the player-facing game structure: what kind of room or match the player creates, what rules govern that match, what options are configurable, what systems are affected by those rules, and what later progression must consume.

### Step 1 - Preset-Driven Room Mode Foundation

The detailed mode and match-rules plan now lives in [modes-and-match-rules.md](domains/gameplay/modes-and-match-rules.md).

That doc owns `ModePreset`, `RoomModeConfig`, `ResolvedMatchRules`, preset-owned policy groups, the `survival_arcade` and `score_attack` baseline modes, affected systems, Step 1 completion criteria, and the open gametime decisions.

### Step 2 - Player Build And Loadout Foundation

Step 2 planning now lives in [player-build-and-loadouts.md](domains/gameplay/player-build-and-loadouts.md).

That doc owns the detailed build flow, `ShipVariant`, `weight_class`, weapon points, weapon classification, module slots, `BuildEligibility`, `EligibleBuildOptions`, `LoadoutSelection`, `ResolvedPlayerBuild`, `RuntimeEquipmentState`, shield support, `OwnedShip` and hardwired module boundaries, and pickup interaction.

Inventory and hangar acquisition details belong in [inventory-and-hangar.md](domains/gameplay/inventory-and-hangar.md) when you need the ownership and acquisition layer.

## Phase D - Progression Foundation

Phase D carries the trusted post-match and progression systems that depend on the player experience flow. The detailed planning now lives in:

- [Match Outcomes And Results](domains/gameplay/match-outcomes-and-results.md)
- [Progression And Rewards](domains/gameplay/progression-and-rewards.md)
- [Inventory And Hangar](domains/gameplay/inventory-and-hangar.md)
- [Player Data And Persistence](domains/platform/stubs/player-data-and-persistence.md)

## Phase E - Gameplay Expansion

Phase E carries the gameplay growth track that depends on the earlier system seams. The detailed planning now lives in:

- [Enemies, Bosses, And Encounters](domains/gameplay/enemies-bosses-and-encounters.md)
- [Modes And Match Rules](domains/gameplay/modes-and-match-rules.md)
- [Network Observability And Packet Budget](domains/technical/network-observability-and-packet-budget.md)
- [Realtime Protocol Architecture](protocol/realtime-protocol-architecture.md)
