---
author: brian
created: "2026-07-27"
document_id: 82eaa581-34a5-4ac3-93bf-fca9ecbc819f
document_type: architecture
policy_exempt: false
summary: Recipient-specific realtime interest filtering, coarse player locator delivery, and spectate-view anchoring.
---
# Network Interest

Parent index: [Game Server Networking](./!INDEX.md)

## Purpose

This document describes the implemented game-server network-interest boundary.

It covers recipient-specific world projection, wrap-aware camera regions, entry/exit hysteresis, lifecycle transitions at the presentation boundary, coarse player locator delivery, and the server-side spectate view-target handoff.

## Overview

Network interest reduces active gameplay traffic by filtering the published presentation snapshot separately for each receiving session before realtime full, delta, lifecycle, and hot-lane candidates are built.

The authoritative game world is not filtered. Simulation entities continue to exist and advance globally. Interest answers only which entity records are relevant to one recipient's presentation.

```text
authoritative GameplayPresentationSnapshot
-> resolve recipient camera/view target
-> apply recipient-specific interest filter
-> build recipient baselines, deltas, lifecycle changes, and hot updates
-> schedule and encode lane candidates
-> write to the recipient's WebRTC channels
```

The current implementation filters ships, asteroids, bullets, and pickups. The receiver's own ship and selected spectate target are always relevant. Far players that leave detailed ship interest remain represented by low-cadence `player_locator` records so the client can continue drawing off-screen indicators.

## Code root

```text
services/game-server/internal/protocol/realtime/
services/game-server/internal/game/visibility/
services/game-server/internal/networking/
```

## Responsibilities

Network interest owns:

- Resolving the presentation camera for one receiving realtime session.
- Filtering ship, asteroid, bullet, and pickup presentation records per recipient.
- Reusing the shared toroidal camera-region predicate used by spawning visibility checks.
- Applying different entry and exit margins to prevent boundary churn.
- Keeping the receiver ship and selected spectate target relevant regardless of distance.
- Letting normal recipient baseline/delta/lifecycle projection express interest entry, exit, and re-entry.
- Building the coarse `player_locator` packet family for all active player identities.
- Sending player locators through the existing `ships` logical lane and `sr.ships` physical DataChannel.
- Accepting `set_view_target_request` and `clear_view_target_request` over the WebSocket control path.
- Storing the selected view target on the receiving WebSocket session.
- Applying the selected view target when building recipient-specific realtime output.

## Does not own

Network interest does not own:

- Authoritative entity existence, movement, collision, damage, death, respawn, scoring, or match rules.
- Client camera rendering or off-screen indicator layout.
- Player hue assignment or team-colour policy.
- WebRTC peer/DataChannel mechanics.
- Realtime JSON representation, compact aliases, tuple encoding, or numeric quantization.
- Candidate priority policy after interest has selected the relevant presentation records.
- The mutable simulation collision broad-phase index under `internal/game/spatial`.
- Room membership or spectate-target eligibility rules.
- Durable player/session identity.

## Domain roles

### Recipient presentation projection

`applyNetworkInterest(...)` runs against a copied `GameplayPresentationSnapshot` for one realtime session. It filters the snapshot maps before the ordinary realtime projection code compares the recipient's current and previous state.

The implementation currently scans the published presentation maps directly. It does not query or reuse the mutable simulation collision index.

### Camera and view-target resolution

The interest camera is resolved in this order:

1. A selected `viewTargetID` with a current player locator.
2. The recipient camera view published in the presentation snapshot.
3. The recipient's own player locator when no camera view was published.

If none of those provide a position, interest filtering is skipped for that build and the unfiltered presentation snapshot continues through projection.

The camera dimensions come from `runtime.CameraView` and `runtime.CameraConfig`. Position tests use `internal/game/visibility.Contains`, which applies toroidal shortest-distance math through the shared world-space helpers.

### Interest margins and hysteresis

The generated server constants are:

```text
entry margin:      320 world units
exit margin:       480 world units
projectile margin: 640 world units
```

An entity that was not relevant in the previous recipient world projection uses the entry margin. An entity that was already relevant uses the wider exit margin.

Bullets use the projectile margin on entry and `800` world units on exit (`640 + (480 - 320)`). This provides additional lead space for fast-moving projectiles while retaining the same hysteresis difference.

Previous relevance is derived from the recipient's stored world baseline projection. There is no second authoritative entity-membership store.

### Entity-family policy

Current policy is:

```text
ships
= camera region with entry/exit hysteresis
= receiver ship always relevant
= selected view target always relevant

asteroids
= camera region with entry/exit hysteresis

bullets
= larger projectile region with entry/exit hysteresis

pickups
= camera region with entry/exit hysteresis
```

Session, overlay, match, team, objective, score, and player lifecycle facts are not removed by this world-presentation filter merely because a ship is distant.

### Presentation lifecycle behavior

Interest exit is not authoritative destruction.

The normal recipient projection interprets filtered membership changes as presentation lifecycle changes:

```text
outside -> inside
= lifecycle create/full record for that recipient

inside -> inside
= ordinary delta/hot updates

inside -> outside
= lifecycle delete/removal for that recipient

outside -> inside again
= a new lifecycle create/full record for that recipient
```

The entity continues existing in the authoritative simulation throughout an interest-only exit.

### Coarse player locator

`player_locator` is a separate packet family containing:

```text
id
x
y
velocity_x
velocity_y
active
```

The server projects locators from the unfiltered player-locator snapshot, so distant players remain represented after their detailed ship record leaves interest.

Locator behavior:

- Logical lane: `ships`.
- Physical channel: existing unordered/unreliable `sr.ships` DataChannel.
- Cadence: approximately 5 Hz through `PlayerLocatorCadenceTicks`.
- Immediate eligibility when locator membership or `active` state changes.
- Independent packet sequence and projection state from `ship_delta`.
- No new DataChannel and no reliable retransmission.

The packet is coarse presentation data. Durable existence and lifecycle truth remain in session/player lifecycle state, while detailed nearby ship state remains in world and ship lifecycle/hot packets.

### Spectate view target

The client sends:

```text
set_view_target_request(view_target_player_id)
clear_view_target_request
```

These packets use the existing WebSocket control connection. The server stores the selected player ID on the WebSocket session and supplies it to recipient-specific realtime projection.

A selected view target becomes the interest-camera anchor when its locator exists. The target ship is also forced into detailed ship interest regardless of distance. Clearing the view target restores normal receiver-camera interest.

## Protocols and APIs

### Server-internal projection API

The primary projection boundary is:

```go
applyNetworkInterest(
    snapshot game.GameplayPresentationSnapshot,
    state RealtimeSessionState,
    viewTargetID string,
) game.GameplayPresentationSnapshot
```

It returns a recipient-filtered presentation snapshot and does not mutate authoritative game state.

### Player locator packet surface

The server builds `PlayerLocatorPacket` as a delta-kind candidate with `LaneShips` metadata. The packet crosses the boundary to the client as compact or readable JSON according to the normal realtime wire codec.

The client consumes the records as a replace-all coarse locator snapshot for that packet sequence. Missing locator records therefore mean the player is no longer present in the current coarse locator projection; they are not interpreted as detailed ship destruction.

### View-target request surface

`set_view_target_request` and `clear_view_target_request` are client-to-server intent/control packets. They do not change gameplay target state and do not grant gameplay authority to the client.

## Data ownership

The authoritative presentation snapshot owns the immutable per-build entity and locator inputs.

The realtime session owns recipient-local projection state:

- world baseline and delta projections
- player-locator projection and sequence
- hot-lane cadence state
- selected WebSocket-session view-target ID

The visibility package owns only wrap-aware region predicates. It does not retain entities or recipient interest sets.

## Code map

### Primary implementation

```text
services/game-server/internal/protocol/realtime/network_interest.go
services/game-server/internal/protocol/realtime/player_locator.go
services/game-server/internal/protocol/realtime/active.go
services/game-server/internal/protocol/realtime/planner.go
services/game-server/internal/protocol/realtime/baseline.go
services/game-server/internal/protocol/realtime/payload.go
```

### Presentation snapshot and shared visibility

```text
services/game-server/internal/game/presentation_snapshot.go
services/game-server/internal/game/visibility.go
services/game-server/internal/game/visibility/region.go
services/game-server/internal/game/space/
```

### Session and inbound control handling

```text
services/game-server/internal/networking/websocket_session.go
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/inbound/gameplay.go
services/game-server/internal/networking/inbound_adapter.go
```

### Schema and generated data

```text
shared/constants/server_constants.toml
shared/constants/client/presentation.toml
shared/packets/gameplay.toml
shared/packets/outputs.toml
shared/packets/realtime_wire.toml
services/game-server/internal/constants/constants.go
services/game-server/internal/game/packets.go
services/game-server/internal/protocol/realtime/packets_generated.go
services/game-server/internal/protocol/realtimewire/generated.go
client/scripts/generated/networking/packets/packets.gd
client/scripts/generated/networking/realtime_wire_generated.gd
```

### Client consumers

```text
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/networking/inbound/client_inbound_coordinator.gd
client/scripts/networking/realtime/realtime_packet_pipeline.gd
client/scripts/protocol/realtime/player_locator_applier.gd
client/scripts/protocol/realtime/player_locator_state.gd
client/scripts/world/world_sync.gd
client/scripts/gameplay/presentation/os_indicator_controller.gd
client/scripts/gameplay/spectate/gameplay_spectate_flow.gd
client/legacy/player_render/player_sync.gd
```

### Important non-ownership boundary

```text
services/game-server/internal/game/spatial/
```

This package remains the mutable authoritative simulation broad phase. Current network interest does not read or repurpose it.

## Tests

Server coverage includes:

```text
services/game-server/internal/protocol/realtime/network_interest_test.go
services/game-server/internal/networking/gameplay_packets_test.go
services/game-server/internal/networking/inbound/context_test.go
services/game-server/internal/networking/inbound/resync_test.go
```

Client coverage includes:

```text
client/tests/unit/protocol/realtime/test_player_locator.gd
client/tests/unit/networking/inbound/test_client_inbound_coordinator.gd
client/tests/unit/networking/test_server_packet_dispatcher.gd
client/tests/unit/world/test_world_sync.gd
client/tests/unit/world/player_render/test_player_render_api.gd
client/tests/unit/gameplay/spectate/test_gameplay_spectate_flow.gd
```

Verification should prove:

- Wrap-aware inclusion works across world seams.
- Entry and exit margins provide hysteresis.
- Receiver and spectate-target ships remain relevant.
- Interest entry/exit produces recipient presentation lifecycle transitions without authoritative destruction.
- Locator packets remain independent from detailed ship updates.
- Locator packets route through the client dispatcher and realtime pipeline.
- Locator-only positions continue driving indicators.
- Remote hue state survives detailed render-interest exit.
- Spectate selection updates both server interest anchoring and client camera focus.

## Related docs

- [Game Server Networking](./!INDEX.md)
- [Realtime WebRTC Gameplay Transport](../../../protocol/realtime-webrtc-gameplay-transport.md)
- [Gameplay Packets](../../../protocol/gameplay-packets.md)
- [Realtime WebSocket Protocol](../../../protocol/realtime-websocket-protocol.md)
- [Realtime Compact Wire Mapping](realtime-compact-wire-mapping.md)
- [Client Inbound Packet Routing](../../client/networking-flow/inbound-packet-routing.md)
- [Client World Sync Coordinator](../../client/world-sync/world-sync-coordinator.md)
- [Client Gameplay Presentation Flow](../../client/presentation-flow/gameplay-presentation-flow.md)
- [Client Spectate Session And Camera Flow](../../client/spectate-flow/spectate-session-and-camera-flow.md)
- [Toroidal Spatial Query Index](../simulation/world/spatial-query-index.md)
- [Toroidal Wrap](../../../systems-design/world/toroidal-wrap.md)

## Notes

The name `player_locator` describes a coarse presentation packet, not a separate physical transport lane. It shares `sr.ships` with detailed ship movement because both are supersedable positional data.

Interest membership is currently inferred from the prior recipient world baseline projection. Future optimization may change how candidates are found, but must preserve the same authority and presentation-lifecycle semantics.
