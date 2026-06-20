# Active Player Avatar State

Parent index: [Game Server Simulation Players](./!README.md)

## Purpose

This document describes active player avatar state in the game-server simulation.

It explains what `runtime.Ship` and `game.entities.Players` own, how active player ships are created and removed, how they are advanced during the simulation tick, and how their render-facing state is projected into `StatePacket.players`.

## Overview

Active player avatar state is the live ship entity state for a player during a match.

The current runtime boundary is:

```text
player session
-> creates active ship avatar
-> stores ship in game.entities.Players
-> simulation mutates ship state
-> StatePacket.players projects active ship/render state
```

`runtime.Ship` is not the durable player session. It is the active avatar/world entity for a player who currently has a ship in the simulation.

The active avatar owns current ship facts such as:

```text
position
rotation
velocity
input
client viewport config copy
ship type
ship stats
health
shields
damage modifiers
damage options
temporary invulnerability
equipped ship weapons
weapon cooldown and ammo state
target kind and target id copy
pending despawn state
despawn delay
```

Durable player facts such as score, lives, respawn cooldown, spawn position, pause/freeze suspension, and long-lived target selection are session-owned. Those facts live in `playerSession` and are projected through `StatePacket.player_sessions` or `StatePacket.player_lifecycle`, not through `StatePacket.players`.

`StatePacket.players` is therefore active avatar/render state only. Pending-respawn and eliminated players may be absent from `StatePacket.players` while still present in `StatePacket.player_sessions` and `StatePacket.player_lifecycle`.

## Code root

```text
services/game-server/internal/game/
```

Primary supporting package:

```text
services/game-server/internal/game/runtime/
```

## Responsibilities

Active player avatar state owns the game-server side of:

* Creating a live ship avatar from a player session.
* Storing active player ships in `game.entities.Players`.
* Mutating active ship position, rotation, velocity, input, health, shields, targeting copy, weapon state, invulnerability, and despawn timers.
* Keeping active ship movement server-authoritative.
* Applying per-tick ship motion through the motion seam.
* Keeping active ship position wrapped inside the toroidal world bounds.
* Updating active avatar weapon cooldown and ammo state.
* Firing active avatar weapons when input and simulation gates allow it.
* Marking a live ship pending despawn after fatal player damage.
* Delaying active ship removal during the collision despawn window.
* Removing ships from `game.entities.Players` once pending despawn completes.
* Projecting active avatar state into `StatePacket.players`.

## Does not own

Active player avatar state does not own:

* Room membership or lobby identity.
* WebSocket session identity.
* Durable player score.
* Durable player lives.
* Respawn cooldown ownership.
* Match lifecycle classification.
* Room match lifecycle.
* Pause/freeze ownership.
* Camera fallback state after death.
* Client rendering or interpolation.
* Packet wire encoding.
* Weapon profile policy internals.
* Projectile spawn intent construction.
* Damage math.
* Collision primitive math.
* Account/profile persistence.

Those systems may read or mutate active avatars through game-owned seams, but they own their own boundaries.

## Domain roles

Active player avatar state participates in the player-facing gameplay domain by representing the player ship that can currently:

```text
move
rotate
shoot
take damage
collide
be targeted
collect pickups
die
produce render-facing state
```

It also participates in the technical simulation domain by keeping the live entity map separate from durable player session state.

A player can exist in the match without an active avatar. That happens when the player has a session but no live ship, such as during pending respawn or elimination.

## Runtime state model

The live avatar type is:

```go
type Ship struct {
    ID                       string
    ShipTypeID               string
    Stats                    ShipStats
    X                        float64
    Y                        float64
    Rotation                 float64
    Velocity                 physics.Vector2
    Input                    InputState
    Config                   ClientConfig
    ShipWeapons              weapons.ShipWeapons
    WeaponState              weapons.State
    TargetKind               string
    TargetID                 string
    Health                   int
    Shields                  int
    DamageModifiers          []damage.DamageModifier
    DamageOptions            DamageOptions
    InvulnerabilityRemaining float64
    PendingDespawn           bool
    DespawnDelay             float64
}
```

The active avatar map is:

```go
game.entities.Players map[string]*runtime.Ship
```

The key is the game player ID. The ship `ID` uses the same value.

The map means “this player currently has a ship entity stored in the simulation.” It does not always mean the player is eligible for every behavior. Pending-despawn ships remain in the map during their despawn delay but are blocked from movement, input, targeting candidates, collision damage, pickup collection, and firing.

## Avatar creation

Initial avatar creation happens through `Game.AddPlayer`.

The current flow is:

```text
allocate player id
-> plan safe initial player spawn
-> create playerSession
-> create runtime.Ship from session.NewShip
-> store session in game.playerSessions
-> store ship in game.entities.Players
-> seed camera view from ship position
-> initialize per-player pending presentation event lane
```

`playerSession.NewShip` copies the session-owned configuration needed by the live ship:

```text
player id
ship type
resolved ship stats
viewport config
current target selection
damage options
equipped primary and secondary weapons
```

It initializes ship health from the resolved ship stats.

Respawn avatar creation also uses `session.NewShip`, but only after respawn gates pass:

```text
session exists
session.CanRespawn()
no active ship currently exists for player id
safe respawn position selected
new ship stored in game.entities.Players
camera view reset to respawned ship position
```

Initial spawn and respawn placement are planning concerns owned by player spawn/respawn logic. Active avatar state receives the selected spawn position and stores the live ship.

## Simulation behavior

Active avatars are advanced during `Game.Step`.

For active matches, the player portion of the simulation currently runs in this order:

```text
step player sessions
-> step player weapons
-> step players
-> remove ready players
```

`stepPlayerWeapons` advances primary and secondary weapon slot state for every live ship in `game.entities.Players`.

`stepPlayers` then advances each active avatar through:

```text
motion.AdvanceShipWithMovePolicy
-> camera view position update
-> pending-despawn skip
-> primary fire check
-> secondary fire check
```

Movement uses the motion package. The active avatar supplies runtime state; the motion package performs per-entity movement math and world wrapping.

When movement is allowed, ship stepping:

```text
ticks temporary invulnerability
reads input axes
updates rotation
applies thrust to velocity
applies damping
clamps velocity to max speed
integrates position
wraps position inside world bounds
```

When movement is blocked by suspension or pending despawn, active-avatar movement does not advance normally. The relevant gate clears input rather than allowing stale controls to keep affecting the ship.

## Input and active avatars

Input packets do not create active avatars.

`Game.HandlePacket` handles respawn and client config before active avatar lookup. For normal input, it then requires:

```text
player id has active ship in game.entities.Players
player can receive input
packet type is input
```

If the active ship is missing, normal input is ignored. This is why pending-respawn and eliminated players can still have session state but cannot move or shoot.

Client config packets can update session and camera config even before the active ship lookup returns. If an active ship exists, the ship config is also updated.

Pause requests require an active ship and are ignored for missing or pending-despawn avatars.

## Pending despawn and removal

Fatal player damage does not immediately delete the active avatar.

The current fatal flow is:

```text
store/update camera view at death position
-> mark ship pending despawn
-> clear ship velocity and input
-> increment session ship deaths
-> decrement session lives when life options allow it
-> set respawn cooldown when lives remain
-> record ship_death event
```

`runtime.Ship.MarkPendingDespawn` marks the avatar as pending despawn, sets the despawn delay, clears velocity, and clears input.

While pending despawn:

```text
the ship may still exist in game.entities.Players
the despawn delay ticks down
movement input is ignored
weapon firing is skipped
collision damage is skipped
target candidate generation skips it
match lifecycle treats it as not having an active ship
```

`removeReadyPlayers` deletes the ship from `game.entities.Players` after `ship.ReadyForRemoval()` returns true.

This delayed removal allows death/despawn presentation to observe the ship briefly while gameplay eligibility has already ended.

## State packet projection

`Game.statePacket` projects active avatars by iterating `game.entities.Players`:

```go
players := make(map[string]runtime.ShipState, len(game.entities.Players))
for id, player := range game.entities.Players {
    players[id] = player.State()
}
```

`runtime.Ship.State()` produces `runtime.ShipState`.

Current projected active avatar fields are:

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

`StatePacket.players` does not include:

```text
score
lives
respawn cooldown
spawn position
ship death count
pause/freeze suspension state
match lifecycle status
camera fallback position
room membership state
```

Those belong to other packet fields or service docs.

The same state packet also projects:

```text
StatePacket.player_sessions
StatePacket.player_lifecycle
```

Use those fields for durable session read-model state and lifecycle classification.

Do not infer player lifecycle from `StatePacket.players` alone. A player can be missing from `StatePacket.players` and still be pending respawn, or can be present during a despawn delay while no longer counted as an active participating ship.

## Lifecycle classification interaction

Match lifecycle classification reads both session and active-avatar facts.

The game builds a match snapshot from `game.playerSessions` and `game.entities.Players`:

```text
session exists
active ship exists
active ship is not pending despawn
session has remaining lives
```

The rules package classifies the player as:

```text
active
pending_respawn
eliminated
```

This keeps active avatar presence separate from player participation state.

The important boundary is:

```text
game.entities.Players
= live avatar storage

playerSession
= durable player match state

rules.MatchDecision
= lifecycle classification from plain facts

StatePacket.players
= active avatar/render projection

StatePacket.player_sessions
= session read model

StatePacket.player_lifecycle
= lifecycle read model
```

## Targeting interaction

Active avatars carry a copy of the player’s current target fields:

```text
TargetKind
TargetID
```

The durable target selection is stored in `playerSession.Targeting`.

When a target changes, game-owned targeting code updates the session targeting state and applies that targeting to the active ship if one exists.

When a ship is created from a session, `session.Targeting.ApplyToShip(ship)` copies the session target onto the new avatar.

This means respawned avatars inherit session-owned target selection. Active avatar state only carries the current packet-facing copy.

## Damage and collision interaction

Active avatars are damageable and collidable only while gameplay gates allow it.

Ship/asteroid collision flow iterates `game.entities.Players`, then skips a ship when:

```text
ship is pending despawn
session is missing
session is suspended
ship is temporarily invulnerable
ship damage options reject damage
```

Fatal damage mutates active avatar state by marking the ship pending despawn. Durable counter changes then go through session/counter ownership.

The damage package does not own active avatar storage. It calculates damage results. Game-owned combat code applies those results to the active ship by mutating health and shields.

## Protocols and APIs

Active player avatar state has no direct HTTP API.

The runtime surfaces that affect active avatars are game-server service methods and realtime packet consequences.

The main service methods are:

```text
Game.AddPlayer
Game.RemovePlayer
Game.HandlePacket
Game.Step
Game.StatePacket
Game.MatchDecision
```

`Game.AddPlayer` is called by networking activation after room start so connected room members receive game-player avatars.

`Game.HandlePacket` consumes decoded client gameplay packets and may update active avatar input, client config, pause state, respawn state, or target state through game-owned routes.

`Game.Step` advances authoritative active avatar state.

`Game.StatePacket` exposes active avatar state to clients through `StatePacket.players`.

The active avatar protocol surface is for presentation and gameplay synchronization. The client consumes active ship state to render ships, interpolate movement, show health/shield/weapon state, and align local player presentation. The client does not own authoritative active avatar mutation.

## Data ownership

Active avatar state is in-memory game runtime state.

It reads:

```text
player sessions
ship stats
ship armory
damage options
targeting state
input packets
client config packets
world simulation options
collision shapes
constants
```

It mutates:

```text
game.entities.Players
runtime.Ship position
runtime.Ship rotation
runtime.Ship velocity
runtime.Ship input
runtime.Ship config
runtime.Ship health
runtime.Ship shields
runtime.Ship weapon state
runtime.Ship target copy
runtime.Ship invulnerability timer
runtime.Ship pending despawn fields
camera view position during active movement
```

It does not persist account/profile data.

Packet shape source data for `ShipState`, `InputState`, `ClientConfig`, `PlayerSessionState`, and `StatePacket` lives under:

```text
shared/packets/gameplay.toml
```

Generated server packet/runtime output includes:

```text
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
```

## Invariants

Active player avatar state must preserve these rules:

* `runtime.Ship` is live avatar state, not durable session state.
* `game.entities.Players` stores live ship entities only.
* Durable score and lives remain session-owned.
* Respawn cooldown remains session-owned.
* Match lifecycle must not be inferred from active avatar map presence alone.
* Pending-despawn ships must not receive input, move normally, fire, collide, or become target candidates.
* Pending-despawn ships are removed only after their despawn delay completes.
* State packet active avatar projection must not duplicate durable session ownership.
* Client rendering observes `StatePacket.players`; it does not own active avatar authority.
* Active avatar motion stays server-authoritative.
* Active avatar creation must use session-owned ship config, stats, armory, damage options, and target selection.
* Active avatar deletion must clear or update dependent game-owned state through the owning seams.

## Code map

Primary implementation files:

```text
services/game-server/internal/game/game.go
services/game-server/internal/game/players.go
services/game-server/internal/game/session.go
services/game-server/internal/game/simulation.go
services/game-server/internal/game/simulation_players.go
services/game-server/internal/game/simulation_weapons.go
services/game-server/internal/game/input.go
services/game-server/internal/game/state_packet.go
services/game-server/internal/game/match.go
services/game-server/internal/game/player_targeting.go
services/game-server/internal/game/targeting.go
services/game-server/internal/game/combat.go
services/game-server/internal/game/combat_damage_application.go
services/game-server/internal/game/pause.go
```

Runtime entity files:

```text
services/game-server/internal/game/runtime/state.go
services/game-server/internal/game/runtime/ship.go
services/game-server/internal/game/runtime/ship_stats.go
services/game-server/internal/game/runtime/damage_options.go
services/game-server/internal/game/runtime/life_options.go
services/game-server/internal/game/runtime/suspension.go
services/game-server/internal/game/runtime/packets_generated.go
```

Motion and spatial support:

```text
services/game-server/internal/game/motion/motion.go
services/game-server/internal/game/space/space.go
services/game-server/internal/game/physics/collision_shapes.go
```

Generated/source packet files:

```text
shared/packets/gameplay.toml
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
```

Related activation and room boundaries:

```text
services/game-server/internal/networking/player_activation.go
services/game-server/internal/rooms/room_lifecycle.go
services/game-server/internal/rooms/room_match.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/rooms/
services/game-server/internal/networking/
services/game-server/internal/game/rules/
services/game-server/internal/game/weapons/
services/game-server/internal/game/damage/
client/
```

`rooms` owns room state and match lifecycle.

`networking` owns websocket sessions and activation routing.

`rules` owns plain match lifecycle classification.

`weapons` owns weapon firing policy and projectile spawn intent calculation.

`damage` owns pure damage result calculation.

`client` owns presentation.

## Tests and verification

Relevant game integration tests:

```text
services/game-server/tests/game/state_packet_lifecycle_test.go
services/game-server/tests/game/match_decision_test.go
services/game-server/tests/game/game_over_test.go
services/game-server/tests/game/movement_test.go
services/game-server/tests/game/respawn_test.go
services/game-server/tests/game/pause_test.go
services/game-server/tests/game/collision_test.go
services/game-server/tests/game/player_counters_test.go
services/game-server/tests/game/ship_type_test.go
services/game-server/tests/game/ship_stats_test.go
```

Relevant package tests:

```text
services/game-server/internal/game/runtime/entity_health_test.go
services/game-server/internal/game/runtime/ship_stats.go
services/game-server/internal/game/motion/
services/game-server/internal/game/rules/
```

Current coverage includes:

* player avatar creation through `Game.AddPlayer`
* active player movement and toroidal wrapping
* state packet player projection
* player lifecycle projection for active, pending-respawn, and eliminated players
* missing active ships not appearing in `StatePacket.players`
* pending respawn after death and delayed removal
* respawned avatars preserving session lives
* respawn safety against asteroids and other players
* pause clearing input and velocity
* paused players not moving or shooting
* resumed players receiving temporary invulnerability
* pending/dead players ignoring pause toggles
* fatal collision marking player death and respawn state
* match-over evaluation from active ship presence and remaining lives

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Focused verification for active avatar state:

```bash
cd services/game-server
go test -buildvcs=false ./tests/game -run 'StatePacket|MatchDecision|GameOver|Movement|Respawn|Pause|Collision|PlayerCounters|Ship'
```

Focused verification for runtime ship state:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game/runtime
```

## Related docs

* [Game Server Simulation Players](./!README.md)
* [Game Server Simulation](../!README.md)
* [Game Server](../../!README.md)
* [Game Server Networking](../../networking/!README.md)
* [Gameplay Network Adapter](../../networking/gameplay-network-adapter.md)
* [Room Match Lifecycle](../../rooms/room-match-lifecycle.md)
* [Room Membership And Identity](../../rooms/room-membership-and-identity.md)
* [Game Server Simulation Runtime](../runtime/!README.md)
* [Game Server Simulation Combat](../combat/!README.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Weapons And Projectile Fire](../combat/weapons-and-projectile-fire.md)
* [Player Session State](player-session-state.md)
* [Player Counters](player-counters.md)
* [Player Death And Despawn](player-death-and-despawn.md)
* [Player Respawn](player-respawn.md)
* [Player Input Routing](player-input-routing.md)
* [Player Pause And Suspension](player-pause-and-suspension.md)
* [Player Camera View State](player-camera-view-state.md)
* [State Packet Projection](../runtime/stubs/state-packet-projection.md)
* [Gameplay Packets](../../../../protocol/stubs/gameplay-packets.md)
* [Data](../../../../data/!README.md)

## Notes

The legacy architecture doc’s most relevant migrated rule is that `runtime.Ship` is active ship/world state only. It is not the owner of durable score, lives, respawn, or match lifecycle state.

`StatePacket.players` can temporarily include pending-despawn ships during the despawn delay. Consumers that need gameplay eligibility should use lifecycle/session/read-model fields and server-owned targeting/collision gates rather than treating active avatar packet presence as the complete participation model.
