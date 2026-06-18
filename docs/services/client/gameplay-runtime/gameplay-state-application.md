# Gameplay State Application

Parent index: [Gameplay Runtime](!README.md)

## Purpose

This document describes how the client applies authoritative gameplay state packets to client presentation systems.

It covers packet normalization, gameplay-state fanout, world-state forwarding, first-state handling, and the boundary between gameplay runtime orchestration and world rendering.

## Overview

Gameplay state application begins after the client networking layer has received and classified a gameplay state packet.

The client does not treat raw packets as a general-purpose read model. Instead, gameplay packets enter `GameplayStateFlow`, are normalized by `GameplayStatePacketReader`, and then pass through `GameplayStateApplyFlow` in a fixed fanout order.

The current path is:

```text
ClientConnectionService.gameplay_state_received
-> SessionNetworkController._on_gameplay_state_received
-> GameplaySessionController.handle_gameplay_state
-> GameplayStateFlow.handle_gameplay_state_packet
-> GameplayStatePacketReader.read
-> GameplayComposition.apply_gameplay_state
-> GameplayShellFlow.apply_gameplay_state
-> GameplayFlowComposer.apply_gameplay_state
-> GameplayStateApplyFlow.apply_state
```

`GameplayStateApplyFlow` is the central client-side application seam. It applies normalized state to runtime consumers, HUD summary state, world sync, alive/respawn restoration, and server-event presentation.

World entity rendering is not owned here. `GameplayWorldStateApplyFlow` adapts the normalized gameplay state into the `WorldSync.apply_state(...)` call.

## Code root

* `client/`

## Responsibilities

* Receive gameplay state packets from the gameplay runtime path.
* Normalize packet fields into a stable client-side state dictionary.
* Keep packet-field access isolated in packet reader code.
* Apply normalized gameplay state in a predictable order.
* Forward world-state fields to world sync.
* Apply HUD summary state from authoritative gameplay state.
* Mark gameplay input as ready only after gameplay state has been received.
* Apply alive/respawn restoration after world state has been applied.
* Route server presentation events after state application.
* Return a first-state result so the gameplay shell can emit gameplay-start lifecycle events.
* Keep runtime state application separate from packet schema ownership and world entity rendering.

## Does not own

* Raw WebSocket transport.
* Packet decoding.
* Packet schema source-of-truth files.
* Server-authoritative simulation.
* Gameplay outcome decisions.
* Collision, score, lives, respawn validity, or match-over authority.
* Room membership or lobby state decisions.
* Entity node creation, cleanup, interpolation, or visual wrap behavior.
* Detailed HUD widget behavior.
* Detailed input or targeting behavior.
* Persistent player data.

## Domain roles

### Packet-to-state normalization

`GameplayStatePacketReader` is the packet-facing boundary. It reads generated packet keys and converts the packet into a normalized dictionary consumed by gameplay runtime flows.

This keeps raw packet-field access out of broad presentation code.

### Gameplay state fanout

`GameplayStateApplyFlow` owns the application order for normalized state. It does not perform authoritative gameplay decisions. It coordinates presentation consumers that need the latest server state.

### World-state adapter

`GameplayWorldStateApplyFlow` adapts the normalized gameplay state into the `WorldSync` API. It is intentionally narrow: gameplay runtime decides when world state is applied, while world sync decides how entities are rendered.

### First-state lifecycle signal

The state application path tracks whether gameplay state has already been received. The first successful gameplay-state application produces a result that allows the shell to emit gameplay-start behavior once.

## Protocols and APIs

### Packet input

Gameplay state packets enter through `GameplayStateFlow.handle_gameplay_state_packet(packet)`.

That method uses:

```gdscript
GameplayStatePacketReader.read(packet)
```

and forwards the normalized state to gameplay composition.

### Normalized state shape

`GameplayStatePacketReader.read(packet)` currently normalizes these fields:

```text
self_id
server_players
player_sessions
player_lifecycle
server_bullets
server_asteroids
server_pickups
total_asteroids
server_events
server_sent_msec
has_lives
lives
```

The reader uses generated packet constants from:

```text
client/scripts/generated/networking/packets/packets.gd
```

Player lifecycle state is normalized through:

```text
client/scripts/gameplay/lifecycle/player_lifecycle.gd
```

### Application order

`GameplayStateApplyFlow.apply_state(...)` applies normalized gameplay state in this order:

```text
1. Apply gameplay state to devtools context.
2. Mark gameplay input as having received gameplay state.
3. Apply gameplay-state summary to HUD flow.
4. Apply world state through GameplayWorldStateApplyFlow.
5. Apply alive/respawn restoration.
6. Apply server presentation events.
7. Mark gameplay state as received.
8. Return first-state result when applicable.
```

This ordering matters.

World state is applied before alive/respawn restoration so restored presentation state can operate against the latest rendered entity state.

Server presentation events are applied after world state so effects can resolve against current visual coordinates.

### World-state forwarding

`GameplayWorldStateApplyFlow` forwards these normalized fields into world sync:

```gdscript
world_sync.apply_state(
    state["self_id"],
    state.get("server_players", {}),
    state.get("server_bullets", {}),
    state.get("server_asteroids", {}),
    state.get("server_pickups", {})
)
```

World sync owns entity-family synchronization after this point.

### First-state result

`GameplayStateApplyFlow` returns a `GameplayStateApplyResult`.

The result indicates whether the current application was the first received gameplay state. `GameplayShellFlow` uses that result to emit gameplay-start behavior after the first state is applied.

## Data ownership

Gameplay state application owns transient application state only.

It may temporarily hold or pass normalized dictionaries during a state-application cycle, but it does not persist authoritative gameplay state.

The state application path does not own durable player data, room data, match results, or server state.

The main stateful boundary is whether gameplay state has already been received. That flag exists to support first-state lifecycle behavior and runtime input readiness.

## Code map

### Packet normalization

* `client/scripts/gameplay/state/gameplay_state_flow.gd`
* `client/scripts/gameplay/state/gameplay_state_packet_reader.gd`
* `client/scripts/gameplay/lifecycle/player_lifecycle.gd`
* `client/scripts/generated/networking/packets/packets.gd`

### State application

* `client/scripts/gameplay/state/gameplay_state_apply_flow.gd`
* `client/scripts/gameplay/state/gameplay_state_apply_result.gd`
* `client/scripts/gameplay/runtime/gameplay_world_state_apply_flow.gd`

### Runtime callers

* `client/scripts/session/session_network_controller.gd`
* `client/scripts/session/gameplay_session_controller.gd`
* `client/scripts/gameplay/gameplay_composition.gd`
* `client/scripts/shell/gameplay_shell_flow.gd`
* `client/scripts/gameplay/runtime/gameplay_flow_composer.gd`

### State consumers

* `client/scripts/world/world_sync.gd`
* `client/scripts/gameplay/respawn/gameplay_alive_restore_flow.gd`
* `client/scripts/gameplay/events/gameplay_event_lifecycle_flow.gd`
* `client/scripts/gameplay/events/gameplay_event_flow.gd`
* `client/scripts/devtools/gameplay_devtools_context.gd`
* `client/scripts/gameplay/input/gameplay_input_context.gd`
* `client/scripts/shell/gameplay_hud_flow.gd`

### Source-of-truth boundaries

* `shared/packets/gameplay.toml`
* `shared/packets/outputs.toml`
* `services/game-server/internal/game/`

## Tests

Relevant tests include:

* `client/tests/unit/test_gameplay_state_packet_reader.gd`
* `client/tests/unit/test_gameplay_state_apply_flow.gd`
* `client/tests/unit/gameplay/test_gameplay_flow_composer.gd`
* `client/tests/unit/test_gameplay_session_controller.gd`
* `client/tests/unit/test_session_network_controller.gd`
* `client/tests/unit/gameplay/test_gameplay_alive_restore_flow.gd`
* `client/tests/unit/gameplay/test_gameplay_event_lifecycle_flow.gd`
* `client/tests/unit/gameplay/test_gameplay_event_controller.gd`

The expected verification path is the client GUT test suite.

## Related docs

* [Gameplay Runtime](!README.md)
* [World Sync](../world-sync/!README.md)
* [Runtime composition](runtime-composition.md)
* [Gameplay session lifecycle](gameplay-session-lifecycle.md)
* [Runtime processing](runtime-processing.md)
* [Gameplay packets](../../../protocol/stubs/gameplay-packets.md) - Stub: gameplay realtime packet documentation.
* [Input and targeting](../input-and-targeting.md) - Client input and targeting documentation.
* [HUD and gameplay UI](../hud-and-gameplay-ui.md) - Client HUD and gameplay UI documentation.

## Notes

This document describes the active client implementation, not future packet architecture.

Packet schema authority belongs to protocol and shared packet source files. Server gameplay authority belongs to the game server. This document only covers the client application path after gameplay packets have reached the gameplay runtime.

`GameplayWorldStateApplyFlow` lives under gameplay runtime because it is part of state fanout. Detailed entity rendering, interpolation, ViewAnchor behavior, and continuous visual coordinates belong in world-sync documentation.
