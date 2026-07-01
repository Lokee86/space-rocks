# Lane Packet Projection

Parent index: [Game Server Simulation Runtime](./!INDEX.md)

## Purpose

This document describes the active game-server lane packet projection path for realtime gameplay presentation.


## Overview

The game server projects authoritative gameplay state into lane packet candidates and results.

The active flow is:

```text
authoritative game state
-> realtime projection / planning
-> lane candidate selection and packet shaping
-> packetmetrics summaries
-> networking write integration
-> WebSocket write
```

Projection is lane-specific rather than one combined gameplay snapshot.

## Code root

```text
services/game-server/internal/protocol/realtime/
services/game-server/internal/networking/websocket_write.go
```

The realtime package owns candidate construction, scheduling, metadata, wire packet assembly, numeric wire quantization, and delta comparison. The websocket write loop owns successful delivery and post-write state changes.

## Responsibilities

The active server projection path owns:

* Projecting authoritative runtime state into lane-native packet families.
* Keeping world, overlay, session, and event ownership separate.
* Producing receiver-specific overlay/session/event output where needed.
* Preserving explicit event-batch drain semantics.
* Leaving packet encoding, transport timing, and write success handling to networking.

## Does not own

The lane projection path does not own:

* WebSocket transport.
* Packet schema source-of-truth files.
* JSON encode/decode mechanics.
* Room lifecycle.
* Client rendering.
* Match rules or simulation mutation.
* Realtime scheduling and planning policy in `services/game-server/internal/protocol/realtime/`.
* WebSocket write integration, write success/failure handling, and post-write state changes in networking.

## Protocols and APIs

Canonical gameplay-family overview: [Gameplay packets](../../../../protocol/gameplay-packets.md)
Canonical detailed lane protocol: [Realtime WebSocket Protocol](../../../../protocol/realtime-websocket-protocol.md)

This doc only covers the projection-side service boundary. It does not define wire lifecycle, transport behavior, baseline rules, sequencing, or resync/control semantics.

## Data ownership

The lane projection path owns the transient projection results used to build lane packets from authoritative game state.

It does not own packet schemas, generated constants, or wire-format definitions. Those belong to packet schemas and protocol docs.

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

`player_pause_state` remains a separate same-session packet and is handled independently from lane packet projection.

## Delta projection behavior

The realtime projection path builds lane records from authoritative game state before delta comparison.

Field-delta comparison is current behavior for these update groups:

```text
world lane
= ship, bullet, asteroid, and pickup updates

overlay lane
= receiver updates

session lane
= player session and player lifecycle updates
```

Creates remain full records. Deletes remain identity lists. Update groups carry partial maps with the identity key plus changed fields only.

Client lane state merges partial update maps into existing records and preserves omitted fields. Omitted fields mean unchanged, not cleared.

Numeric wire quantization is implemented in the realtime projection and wire-record path before delta comparison. The active server implementation uses `services/game-server/internal/protocol/realtime/quantize/` and `services/game-server/internal/protocol/realtime/quantize_world.go` as the quantization boundary for outbound lane projection. It should not truncate authoritative simulation state for packet-size savings.

The ownership boundary remains:

```text
simulation
= authoritative gameplay state

realtime projection
= lane packet shaping and delta comparison

packetcodec
= JSON encode/decode mechanics

networking
= WebSocket write integration and write success/failure handling
```

## Event semantics

Presentation event projection is non-draining until the active send path explicitly drains after successful active handling.

The important rule is:

```text
projection may inspect or copy pending presentation events
active send/write path is the drain point after successful active handling
```

Projection, shadow, and inspection paths must not treat event access as an implicit flush.

## Code map

Relevant active files include:

* `services/game-server/internal/protocol/realtime/` - lane candidates, metadata, planning, scheduling, wire packets, metrics bridge, and shadow/parity helpers.
* `services/game-server/internal/networking/websocket_write.go` - active write integration and post-write state changes.
* `services/game-server/internal/networking/packetmetrics/` - sent lane metric summaries and packet metrics helpers.
* `services/game-server/internal/networking/` - websocket session and outbound delivery boundaries.
* `shared/packets/gameplay.toml` - shared gameplay schema and realtime packet type values.
* `shared/packets/outputs.toml` - generated output routing for packet constants and builders.

## Tests

Relevant server tests include:

* `services/game-server/internal/protocol/realtime/*_test.go`
* `services/game-server/internal/networking/websocket_write_test.go`
* `services/game-server/internal/networking/room_snapshot_test.go`
* `services/game-server/internal/networking/room_error_test.go`
* `services/game-server/internal/networking/session_auth_test.go`
* `services/game-server/internal/networking/player_pause_state_test.go`
* `services/game-server/internal/networking/outbound/debug_status_presentation_test.go`
* `services/game-server/internal/networking/outbound/debug_shape_catalog_presentation_test.go`

## Related docs

* [Gameplay packets](../../../../protocol/gameplay-packets.md)
* [Realtime WebSocket Protocol](../../../../protocol/realtime-websocket-protocol.md)
* [Game Server Simulation Runtime](./!INDEX.md)
* [Presentation Event Queue](./presentation-event-queue.md)
* [Packet Schemas](../../../../data/packet-schemas.md)

## Notes


