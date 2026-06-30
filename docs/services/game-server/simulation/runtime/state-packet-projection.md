# Lane Packet Projection

Parent index: [Game Server Simulation Runtime](./!INDEX.md)

## Purpose

This document describes the active game-server lane packet projection path for realtime gameplay presentation.

The filename is historical. The active runtime path is lane-native projection, not combined `StatePacket` projection.

## Overview

The game server now projects gameplay presentation as lane packets:

```text
world_full / world_delta
overlay_full / overlay_delta
session_full / session_delta
event_batch
player_pause_state
```

The active flow is:

```text
networking write tick
-> realtime active packet planning
-> lane projection from authoritative game state
-> packet metrics / budgeting
-> packetcodec.Encode
-> WebSocket write
```

Projection is lane-specific rather than one combined gameplay snapshot.

## Lane ownership

Current gameplay presentation ownership is split as:

```text
world lane
= active entity presentation state such as ships, asteroids, bullets, pickups

overlay lane
= local-player presentation facts such as lives, score, loadout, cooldown-facing HUD facts

session lane
= durable match-local player session state and lifecycle-oriented read models

event batch
= transient presentation events sent separately from baseline/delta lanes
```

`player_pause_state` remains a separate same-session packet and is not part of the old combined gameplay packet removal.

## Responsibilities

The active server projection path owns:

* Projecting authoritative runtime state into lane-native packet families.
* Keeping world, overlay, session, and event ownership separate.
* Producing receiver-specific overlay/session/event output where needed.
* Preserving explicit event-batch drain semantics.
* Leaving packet encoding, transport timing, and write success handling to networking.

## Event semantics

Presentation event projection is non-draining until the active send path explicitly drains after successful active handling.

The important rule is:

```text
projection may inspect/copy pending presentation events
active send/write path is the drain point
```

Projection must not treat event access as an implicit flush.

## Does not own

The lane projection path does not own:

* WebSocket transport.
* Packet schema source-of-truth files.
* JSON encode/decode mechanics.
* Room lifecycle.
* Client rendering.
* Match rules or simulation mutation.

## Code map

Relevant active files include:

* `services/game-server/internal/protocol/realtime/`
* `services/game-server/internal/networking/websocket_write.go`
* `services/game-server/internal/networking/packetmetrics/`
* `services/game-server/internal/game/`
* `shared/packets/gameplay.toml`
* `shared/packets/outputs.toml`

## Related docs

* [Gameplay packets](../../../../protocol/gameplay-packets.md)
* [Game Server Simulation Runtime](./!INDEX.md)
* [Presentation Event Queue](./presentation-event-queue.md)

## Notes

Historical combined `StatePacket` projection is no longer the active gameplay send path.
