# Ships

Parent index: [Entities](./!README.md)

## Purpose

This document defines the systems-design model for ship entities in Space Rocks.

It documents what a ship is conceptually, what authority owns ship behavior, how live ship state differs from durable player state, and which invariants must hold across server simulation, realtime packets, client presentation, combat, respawn, and future build/loadout work.

## Overview

Ships are server-authoritative live world entities.

The current implemented ship entity is the player-controlled runtime avatar stored by the game server in `game.entities.Players`. It represents a player’s active in-world ship during a match.

A ship can currently:

```text
move
rotate
fire equipped weapons
carry health and shields
carry temporary damage options and modifiers
collide with asteroids and pickups
hold a packet-facing target copy
enter pending despawn after fatal damage
project render-facing state to clients
```

A ship is not the same thing as a player, room member, websocket session, local pilot profile, account, inventory item, or durable loadout.

The core split is:

```text
playerSession
= durable per-match player state

runtime.Ship
= live active avatar/world entity

StatePacket.players
= packet-facing active ship state

client player rendering
= presentation of server-owned active ship state
```

A player can exist in the match without an active ship. Pending-respawn and eliminated players remain represented by player session and lifecycle state, but may be absent from active ship state.

## Conceptual model

The current ship model is built around session-to-avatar creation.

```text
room/networking activation
-> game creates playerSession
-> playerSession creates runtime.Ship
-> simulation mutates runtime.Ship
-> StatePacket.players projects runtime.Ship
-> client renders ship presentation
```

The ship owns active entity state only.

Current active ship state includes:

```text
ship id
ship type id
resolved ship stats
position
rotation
velocity
current input
client viewport config copy
equipped primary and secondary weapons
weapon cooldown and ammo state
target kind
target id
health
shields
damage modifiers
damage options
temporary invulnerability timer
pending despawn flag
despawn delay
```

Durable player state stays outside the ship:

```text
score
lives
ship death count
respawn cooldown
spawn position
pause/freeze suspension
durable target selection
player armory defaults
match lifecycle classification
room membership
profile/account persistence
```

Those values belong to player sessions, room/networking state, player-data systems, or future inventory/loadout systems.

## Ship identity and type

The current runtime ship id is the game player id.

```text
runtime.Ship.ID == player id
game.entities.Players[playerID] == active ship entity
```

The current default ship type is:

```text
v_wing
```

Ship type is carried as `ShipTypeID` in runtime state and projected as `ship_type` in packet state.

The current server model preserves ship type through the player session:

```text
playerSession.ShipTypeID
-> playerSession.NewShip
-> runtime.Ship.ShipTypeID
-> ShipState.ship_type
```

Respawn recreates a new active ship from the existing session, so ship type and resolved stats come from the session rather than from client presentation.

Unknown or unsupported ship type resolution currently falls back to default ship modifiers and default effective stats. This keeps the runtime safe while full selectable ship variants remain planned work.

## Ship stats

`ShipStats` are the resolved effective runtime values copied from session state into the active ship.

Current resolved ship stats include:

```text
rotation speed
thrust force
max speed
damping
max health
bullet cooldown
bullet damage
bullet speed
bullet lifetime
bullet spawn offset
collision shape id
```

Conceptually, ship stats own chassis behavior and survivability. Weapon firing, projectile tuning, damage intent, ammunition, and impact effects belong to the weapon profile model.

Some current ship stat fields still reflect older ship-side weapon tuning. The durable design direction is that weapon profiles own weapon behavior, while ship or build data may later modify weapon behavior through explicit build/stat seams.

## Lifecycle

Ships are created when a player becomes an active game participant.

Initial creation flow:

```text
Game.AddPlayer
-> plan safe initial spawn
-> create playerSession
-> session.NewShip(spawn position)
-> store runtime.Ship in game.entities.Players
-> attach camera view
```

Respawn creation flow:

```text
respawn request
-> session exists
-> session can respawn
-> no active ship currently exists
-> safe respawn position selected
-> session.NewShip(respawn position)
-> store runtime.Ship in game.entities.Players
-> reattach camera view
```

Fatal damage does not immediately delete the ship.

Fatal player damage currently:

```text
stores or updates camera view at death position
marks the ship pending despawn
clears ship velocity and input
increments session ship deaths
decrements session lives when life options allow it
sets respawn cooldown when lives remain
records a ship death event
```

A pending-despawn ship may still be present in `game.entities.Players` for presentation timing, but it is no longer a normal active participant.

Pending-despawn ships must not:

```text
receive normal input
move normally
fire weapons
take further collision damage
collect pickups
be treated as active lifecycle participants
be normal target candidates
```

The ship is removed from the active player entity map only after its despawn delay completes.

## Movement and world position

Ship movement is server-authoritative.

The client sends input intent. The server stores that input on the active ship only when input gates allow it. During simulation, ship motion reads active ship state and resolved ship stats to update rotation, velocity, and position.

Current movement behavior:

```text
read left/right input as rotation axis
read forward/back input as thrust axis
apply rotation speed
apply thrust force
apply damping
clamp velocity to max speed
integrate position
wrap position inside world bounds
```

The world is toroidal. Server movement wraps ship positions inside the authoritative world bounds. Client rendering converts server positions into continuous visual positions around the active ViewAnchor.

The client may interpolate and render the ship, but it does not own authoritative position, rotation, velocity, collision state, or movement outcomes.

## Combat participation

A ship is a combat participant only while it is an active, eligible avatar.

Ships can currently participate in combat through:

```text
primary weapon fire
secondary weapon fire
projectile ownership
ship/asteroid collision
damage resolution
health and shield mutation
fatal damage and despawn
pickup collection
target state projection
```

Weapon fire is server-owned. The ship carries equipped runtime weapons and mutable weapon state, but weapon profiles own firing metadata, projectile intent, damage intent, ammo policy, and impact-effect metadata.

Damage is resolved outside the ship entity. Damage resolution decides what happened from a damage request; game-owned combat code applies that result to ship health and shields.

Fatal damage crosses from combat into player lifecycle by marking the active ship pending despawn and mutating durable session counters through player-owned seams.

## Targeting

Ships carry packet-facing target fields:

```text
target_kind
target_id
```

These fields are a live avatar copy.

The durable target selection belongs to player session targeting state. When a target changes, game-owned targeting code updates the session and applies the current target to the active ship if one exists. When a new ship is created from the session, the session target is copied onto the ship.

This keeps targeting stable across death and respawn without making the active ship the durable owner of target selection.

A ship’s presence in active state does not automatically make every target interaction valid. Targetability still depends on server-owned lifecycle, pending-despawn, suspension, and candidate-selection gates.

## Collision model

The server owns collision behavior.

Current ship collision uses imported collision shape data from the shared collision-shape export. The active ship resolves a collision body from its `Stats.CollisionShapeID`.

The current default collision shape id is:

```text
v_wing
```

Current lookup behavior safely falls back to the default ship shape for missing or unsupported ship shape ids.

The client may render collision debug overlays through devtools, but client presentation does not decide collision behavior. Collision shape export data and server collision loading are the authority for gameplay collision.

## Authority rules

Ship authority follows these rules:

```text
The server owns ship creation.

The server owns active ship storage.

The server owns movement simulation.

The server owns collision participation.

The server owns damage application to ship health and shields.

The server owns pending despawn and active ship removal.

The server owns weapon fire outcomes.

The server owns pickup collection outcomes.

The server owns target selection authority.

The client owns input collection and presentation only.

The player session owns durable per-match player state.

Room and networking systems own membership and connection routing, not ship behavior.

Player-data and future inventory systems own durable profile/account/equipment state, not active ship mutation.
```

Client-provided values must not be treated as authority for ship position, collision shape, ship stats, health, shields, weapon cooldown, ammo, damage, or lifecycle.

## Invariants

Ship systems must preserve these invariants:

```text
A runtime ship is live avatar state, not durable player state.

A player session may exist without an active ship.

StatePacket.players is active ship state only.

Player lifecycle must not be inferred only from StatePacket.players.

Durable score, lives, respawn cooldown, and ship death count remain session-owned.

Respawn recreates a ship from session state.

Fatal damage marks the active ship pending despawn before removal.

Pending-despawn ships are not normal gameplay participants.

Ship movement remains server-authoritative.

Ship collision behavior remains server-authoritative.

Ship type and collision shape selection must not become client-authoritative.

Client rendering consumes ship state but does not own ship truth.

Weapon profiles own weapon behavior; ships carry equipped runtime weapons and mutable weapon state.

Future loadout, inventory, and owned-ship systems must feed ship setup through explicit server-owned resolution rather than direct client mutation.
```

## Participating systems

Ships participate in these systems:

```text
Game server simulation
= authoritative active ship state, movement, combat participation, lifecycle, respawn, state projection

Realtime protocol
= input intent from client and active ship state from server

Client world sync
= presentation, interpolation, ViewAnchor-relative rendering, visual read models

Combat
= weapons, projectiles, collisions, damage, radial effects, pickups, death events

Player lifecycle
= session state, lives, respawn cooldown, lifecycle classification, pending respawn, elimination

Collision shape data
= exported shape data consumed by server collision lookup

Planning and limits
= future ship variants, loadouts, owned ships, hardpoints, modules, and full build resolution
```

## Service implementation

The current authoritative implementation lives in the game-server simulation.

The main service implementation docs are:

```text
Active Player Avatar State
Player Lifecycle Overview
Player Session State
Player Respawn
Player Death And Despawn
Player Input Routing
Weapons And Projectile Fire
Collision To Damage Flow
State Packet Projection
```

This systems-design document owns the conceptual model and invariants. Service docs own detailed code paths, tests, generated files, runtime flow, and implementation maps.

## Protocol and data relationships

Ships are projected through realtime gameplay state packets.

The active ship state projection is:

```text
StatePacket.players
```

Current ship-facing packet fields include:

```text
id
ship_type
x
y
rotation
health
shields
thrusting
target_kind
target_id
primary_weapon_id
primary_ammo_policy
primary_cooldown_remaining
primary_ammo_remaining
secondary_weapon_id
secondary_ammo_policy
secondary_cooldown_remaining
secondary_ammo_remaining
```

Durable per-match player state is projected separately through:

```text
StatePacket.player_sessions
StatePacket.player_lifecycle
```

Packet shape source data lives under the gameplay packet source-of-truth files and is generated into both server and client packet helpers.

Collision shape data is exported from Godot scene collision nodes into shared JSON and loaded by the server. Collision data is not owned by the client at runtime, even though the source shape comes from Godot scenes.

## Client presentation

The client renders ships from server-owned state.

Current client presentation responsibilities include:

```text
applying normalized state to WorldSync
creating or removing player presentation nodes
interpolating player positions
rendering local and remote player hues
aligning world presentation around ViewAnchor
exposing player visual positions to targeting and presentation read models
showing HUD and weapon state derived from packets
```

The current client receives `ship_type`, but player rendering does not yet select a different ship scene from it. Future ship scene mapping belongs to the player-build and ship-variant implementation work.

## Planning boundary

Future player-build work is expected to expand ship setup into a fuller model:

```text
ShipVariant
LoadoutSelection
ResolvedPlayerBuild
OwnedShip
weapon points
module slots
hardwired modules
weight class
starting equipment state
shield policy
```

Those are planning concepts until implemented.

The current ship systems-design boundary remains compatible with that direction:

```text
ship entity
= live avatar runtime state

ship variant
= future chassis definition and setup input

owned ship
= future durable inventory/hangar state

loadout selection
= future pre-match player choice

resolved player build
= future authoritative match-start setup object
```

Future enemy and boss work may share lower-level concepts such as position, velocity, rotation, health, shields, damage modifiers, weapon state, and collision bodies. It must not inherit player-specific assumptions such as player input, player session lifecycle, player inventory, or player build ownership unless those concepts are explicitly modeled for enemies.

## Related docs

* [Entities](./!README.md)
* [Weapons](../combat/weapons.md)
* [Damage](../combat/damage.md)
* [Targeting](../combat/targeting.md)
* [Pickups](../combat/pickups.md)
* [Active Player Avatar State](../../services/game-server/simulation/players/active-player-avatar-state.md)
* [Player Lifecycle Overview](../../services/game-server/simulation/players/player-lifecycle-overview.md)
* [Player Session State](../../services/game-server/simulation/players/player-session-state.md)
* [Player Respawn](../../services/game-server/simulation/players/player-respawn.md)
* [Player Death And Despawn](../../services/game-server/simulation/players/player-death-and-despawn.md)
* [Weapons And Projectile Fire](../../services/game-server/simulation/combat/weapons-and-projectile-fire.md)
* [Collision To Damage Flow](../../services/game-server/simulation/combat/collision-to-damage-flow.md)
* [State Packet Projection](../../services/game-server/simulation/runtime/state-packet-projection.md)
* [World Sync Coordinator](../../services/client/world-sync/world-sync-coordinator.md)
* [View Anchor And Visual Coordinates](../../services/client/world-sync/view-anchor-and-visual-coordinates.md)
* [Gameplay Packets](../../protocol/gameplay-packets.md)
* [Collision Shape Data](../../data/collision-shape-data.md)
* [Player Build And Loadouts](../../planning/domains/gameplay/player-build-and-loadouts.md)
* [Enemies Bosses And Encounters](../../planning/domains/gameplay/enemies-bosses-and-encounters.md)
* [Player Build Limits](../../limits/player-build-limits.md)

## Notes

The server owns ship type resolution, resolved stats, and collision behavior. The client consumes `ship_type`; it does not decide collision behavior.

The current default ship type and collision shape id are both `v_wing`.

The current runtime type name `Ship` is also used by a not-yet-active enemy map shape in the entity store. This document describes the implemented player ship/avatar model, not a completed enemy entity model.
