## Targeting

Parent index: [Combat](./!README.md)

## Purpose

This document defines the gameplay targeting model for Space Rocks combat and interaction systems.

It explains what a target means, which systems may request or consume target state, where authority lives, and which invariants must be preserved when targeting is used by input, combat, devtools, telemetry, pickups, projectiles, or future lock-on systems.

## Overview

Targeting is the match-local concept of a player selecting an entity reference in authoritative gameplay state.

The canonical target identity is:

```text
target_kind
target_id
```

`target_kind` identifies the target family. `target_id` identifies the entity or player inside that family.

Current canonical gameplay target kinds are:

```text
player
enemy
pickup
asteroid
bullet
```

Targeting is server-authoritative. The client may build local visual candidates and send target intent, but the game server decides whether the target is accepted. Accepted target state is confirmed only when it appears in authoritative gameplay state readback.

A selected target is not automatically combat behavior. Targeting records intent or focus. It does not, by itself, damage the target, collect the pickup, fire a weapon, command an entity, validate a devtools command, or make an entity eligible for every targeting consumer.

## Conceptual model

Targeting has three separate layers:

```text
client candidate
= what the local client thinks the player clicked or can click

canonical selected target
= what the server has accepted and stored for the player

consumer-specific target use
= what another system does, or refuses to do, with the selected target
```

The client candidate layer is presentation-side and transient. It may be based on rendered positions, synchronized world state, visual pick radii, local mouse position, and UI input rules.

The canonical selected target layer is server-owned match state. It stores a generic target reference on the player session and copies it to the active ship when one exists.

The consumer layer belongs to the system consuming target state. Weapon fire, radial effect filtering, devtools commands, telemetry, spectate behavior, and pickup collection each have their own rules. They must not treat selected target state as automatic permission to act.

## Target identity

Generic gameplay targeting must use:

```text
target_kind + target_id
```

For a player target, the canonical shape is:

```text
target_kind = "player"
target_id = <player id>
```

Player-only code may use a direct player ID when it already owns a player-specific context. It should not create a parallel generic targeting model.

`target_player_id` is not canonical gameplay targeting. It is a legacy player-only compatibility surface for debug/devtools command paths. New normal gameplay systems must not introduce or expand `target_player_id`.

## Authority rules

The game server owns:

* canonical target acceptance
* target existence validation
* point-based target selection validation
* selected target storage
* selected target clearing
* selected target readback through gameplay state
* target lifecycle classification as active, inactive, or missing

The client owns:

* local input translation
* mouse action priority
* visual candidate construction
* client-side candidate picking
* outbound target request packets
* presentation and readback display

The protocol and data pipeline own:

* packet field names
* packet type strings
* generated client/server packet helpers
* schema drift checks

Devtools own:

* command-specific target interpretation
* whether a debug command can resolve from the current gameplay target
* player-only command restrictions

Combat systems own:

* whether selected target state affects weapon fire
* whether an effect can include a target kind
* whether a target can receive damage
* whether an interaction has any gameplay consequence

## Selection model

Client selection is a request.

The normal selection flow is:

```text
local input
-> client visual candidate pick
-> select_target_at_position_request
-> server target existence check
-> server collision-body point check
-> session-owned target state update
-> active ship target copy when present
-> StatePacket.players[*].target_kind / target_id
-> client readback
```

The point-based request carries:

```text
x
y
target_kind
target_id
```

The server does not blindly accept the clicked candidate. It verifies that the requesting player session exists, the requested target reference is non-empty, the target exists in authoritative state, a matching server-side target candidate exists, and the submitted point is inside that candidate’s authoritative collision body.

If validation fails, the existing selected target remains unchanged.

Clearing target state is explicit. A clear request stores the empty target for the requesting player when the player session exists.

## Session and active ship model

Targeting is stored on the player session as match-local state.

The active ship carries a packet-facing copy of that session target:

```text
playerSession.Targeting
= authoritative match-local selected target

runtime.Ship.TargetKind / TargetID
= active avatar copy

StatePacket.players[*].target_kind / target_id
= client readback for active ships
```

This split matters because a player can have a session without an active ship. A dead or pending-respawn player can still hold, update, or clear selected target state. When the player respawns, the newly created ship receives the session-owned target copy.

`StatePacket.players` only contains active avatar state. Session-owned target state can exist while absent from that active-ship readback.

## Target status and lifecycle

Target status is separate from target identity.

The current target status model is:

```text
active
inactive
missing
```

A player target is active when the player world state exists and is targetable. A pending-respawn player is inactive, not missing. A removed player is missing.

Asteroid, bullet, and enemy targets are active when present and not pending despawn. They are inactive when present but pending despawn. They are missing when absent or nil.

Pickup targets are currently active when present and missing when absent or nil.

Missing-target cleanup clears selected targets that refer to removed entities. It must not clear targets merely because the target is temporarily inactive.

## Candidate and priority model

The client can build visual candidates for:

```text
player
enemy
pickup
asteroid
bullet
```

Current client picking chooses among visible valid candidates within their pick radius. Higher `pick_rank` wins first. When rank ties, target kind priority is used.

The current kind priority order is:

```text
player
enemy
pickup
asteroid
bullet
unknown
```

Server-side position validation does not use client visual priority as authority. The server validates the specific `target_kind` and `target_id` submitted by the client.

## Combat and interaction boundary

Targeting is not damage.

A selected target does not imply:

```text
weapon can fire
projectile will home
target is damageable
target should receive area damage
pickup is collectable
devtools command can act on it
spectate camera should follow it
```

Each consuming system must make its own eligibility decision.

Examples:

* Weapons own fire rules, cooldowns, ammo, projectile creation, and any future lock-on requirements.
* Damage owns damage resolution after a valid damage request exists.
* Radial effects own their own target filters and spatial coverage checks.
* Pickups own collection and effect application.
* Devtools own whether a command can resolve from the gameplay target.
* Spectate owns camera target selection separately from gameplay target state.

This boundary prevents targeting from becoming a hidden global permission system.

## Target kinds and adjacent filters

Canonical gameplay targeting currently uses `bullet` for projectile targets because gameplay state exposes active projectiles through the bullet state map.

Radial effects use a separate target-filter vocabulary that includes `projectile`. That filter belongs to radial effect evaluation, not canonical player-selected target identity.

The two concepts are related but not identical:

```text
canonical selected target
= player intent/readback target identity

radial target filter
= effect-specific inclusion rule for hit evaluation
```

Consumers should translate deliberately when crossing that boundary. They should not assume every target-kind enum in one package is interchangeable with another package’s targeting or filtering vocabulary.

## Participating systems

Game server simulation participates as the authority for selected target state, target validation, target status, and state-packet projection.

Client input and targeting participate by translating mouse/input intent into target requests and reading authoritative confirmation from gameplay state.

Realtime gameplay packets participate by carrying target requests and target readback fields between client and server.

Packet schema data participates by defining the packet fields and generated packet helpers used by both services.

Devtools participate as target-state consumers and inspectors, but devtools command targeting remains command-specific.

Combat systems participate only when they explicitly consume selected target state or target-kind filters for a mechanic.

## Service implementations

Targeting implementation is primarily covered by service documentation rather than this systems-design document.

Current implementation touchpoints include:

```text
services/game-server/internal/game/targeting.go
services/game-server/internal/game/player_targeting.go
services/game-server/internal/game/targeting/targeting.go
services/game-server/internal/game/player/target_status.go
services/game-server/internal/game/session.go
services/game-server/internal/game/runtime/state.go
services/game-server/internal/game/runtime/ship.go
services/game-server/internal/networking/inbound/gameplay.go
client/scripts/gameplay/input/
client/scripts/gameplay/targeting/
client/scripts/generated/networking/packets/packets.gd
shared/packets/gameplay.toml
```

The detailed game-server implementation map belongs in the game-server simulation targeting docs. The detailed client implementation map belongs in the client input and targeting doc. Packet source and generated output ownership belongs in protocol and data docs.

## Invariants

* Canonical gameplay target identity is `target_kind` plus `target_id`.
* New normal gameplay paths must not introduce `target_player_id`.
* `target_player_id` remains a player-only devtools/debug compatibility surface.
* Client candidate selection is not authority.
* Server-selected target state must be confirmed through authoritative gameplay state readback.
* Point-based selection must validate both the requested target identity and server-side collision-body containment.
* Invalid target selection must not overwrite the previous accepted target.
* Empty target clearing is explicit.
* Player session targeting is the match-local ownership point.
* Active ship target fields are projection copies of session targeting.
* Respawned ships inherit session-owned target state.
* Target status must distinguish inactive existing targets from missing removed targets.
* Missing-target cleanup must clear missing targets without clearing inactive existing targets.
* Targeting state alone must not imply damage, collection, lock-on, command eligibility, scoring, or presentation effects.
* Consumer systems must apply their own target-kind and eligibility rules.
* Devtools player-only commands may resolve from gameplay target state only when the canonical target kind is `player`.
* Spectate target selection is separate from gameplay target selection.

## Related docs

* [Combat](./!README.md)
* [Game Server Simulation Targeting](../../services/game-server/simulation/targeting/!README.md)
* [Canonical Target State](../../services/game-server/simulation/targeting/canonical-target-state.md)
* [Target Selection And Status](../../services/game-server/simulation/targeting/target-selection-and-status.md)
* [Input And Targeting](../../services/client/input-and-targeting.md)
* [Gameplay Packets](../../protocol/gameplay-packets.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Devtools](../../devtools/!README.md)
* [Radial Effects](radial-effects.md)
* [Pickups](pickups.md)
* [Weapons](weapons.md)
* [Damage](damage.md)

## Notes

The quarantine boundary remains: generic gameplay targeting is `target_kind` plus `target_id`; `target_player_id` must not leak back into normal gameplay.

Current selected target readback is attached to active ship state. Session-owned targeting can outlive the active ship, but it is not separately projected through `player_sessions`.

`enemy` is already a canonical target kind in the targeting model, even though enemy gameplay is still broader planning work outside this document.
