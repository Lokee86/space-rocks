# Player Input Routing

Parent index: [Game Server Simulation Players](./!INDEX.md)

## Purpose

This document describes the game-server simulation boundary for player input routing after networking has already resolved the active game player.

It covers how input, pause, respawn, and client config gameplay packets reach `Game.HandlePacket`, how the game validates the active player state before applying input, and how stored input is consumed by movement and weapon simulation.

## Overview

Player input routing is server-authoritative.

The client may send movement, firing, pause, respawn, and viewport/config requests, but those packets are only requests or input reports. The game server decides whether the packet is associated with an active game player, whether the player can currently receive input, and which player-owned simulation state should mutate.

The current routing path is:

```text
WebSocket message
-> inbound packet routing
-> current room + current game player lookup
-> Game.HandlePacket(playerID, packet)
-> player/session/camera mutation
-> Game.Step(...)
-> movement and weapon simulation consume stored input
-> state packet projection
```

Networking owns the active room/player handoff. The game simulation owns what happens after `Game.HandlePacket` receives a `playerID` and decoded `ClientPacket`.

`Game.HandlePacket` currently handles these player-side gameplay packet types:

```text
input
respawn
client_config
pause_request
```

Targeting packets are also gameplay packets, but they route through the game targeting API instead of the player input-routing path.

## Code root

```text
services/game-server/internal/game/
services/game-server/internal/networking/inbound/
shared/packets/gameplay.toml
```

## Responsibilities

The player input-routing boundary owns:

* Applying decoded `input` packets to the active runtime player ship.
* Rejecting input when the player is missing, pending despawn, or suspended.
* Routing respawn requests into the player respawn path.
* Routing pause requests into the player pause/suspension path.
* Applying valid client viewport config to player session, camera view, and active runtime ship state.
* Preserving the split between durable player session state and active ship/avatar state.
* Ensuring stored movement/fire input is consumed only by authoritative simulation phases.
* Keeping client input capture separate from server-side input acceptance and simulation effects.

## Does not own

This boundary does not own:

* WebSocket lifecycle.
* Packet decode mechanics.
* Packet schema source-of-truth files.
* Room membership.
* Active game player assignment.
* Client-side key bindings or input polling.
* Client-side menu, pause overlay, respawn overlay, or HUD behavior.
* Target selection semantics.
* Weapon fire policy.
* Projectile construction.
* Respawn eligibility rules beyond routing the request to the respawn path.
* Pause/suspension semantics beyond calling the pause path and honoring its input gates.
* State packet write loops or outbound transport.
* Devtools command routing.

Those responsibilities belong to networking, protocol, data, rooms, client, combat, targeting, respawn, pause/suspension, devtools, or runtime docs as appropriate.

## Domain roles

### Active game player handoff

The WebSocket connection itself does not imply an active game player.

Before `Game.HandlePacket` is called, inbound networking requires the session to have:

```text
current room
current game player ID
```

For direct gameplay packets:

```text
input
respawn
client_config
```

`HandleGameplayPacket` treats the packet type as consumed even when no current room or game player exists. In that case, no gameplay mutation is applied.

When both values exist, the packet is forwarded to:

```text
room.GameInstance().HandlePacket(currentGamePlayerID, packet)
```

This means player input routing starts only after networking has already resolved the per-session game player identity.

### Packet dispatch inside the game

`Game.HandlePacket` locks the game instance before inspecting the packet.

Respawn requests are handled first:

```text
respawn
-> game.respawnPlayer(playerID)
-> return
```

Client config packets are handled in two phases. If the config has positive visible-world dimensions, the game updates session and camera view config before active-player lookup. If the player entity is active, the same valid config is also copied onto the runtime ship.

After config pre-handling, `Game.HandlePacket` requires an active runtime player ship:

```text
game.entities.Players[playerID]
```

If no active ship exists, the packet returns without further mutation. This prevents input and pause from recreating or mutating inactive players.

The active-player switch currently handles:

```text
input
-> if playerCanReceiveInput(...)
-> player.SetInput(packet.Input)

pause_request
-> game.togglePlayerPaused(playerID)

client_config
-> if dimensions are valid
-> player.SetConfig(packet.Config)
```

### Input acceptance gate

Input packets do not immediately move or fire the ship. They replace the stored `runtime.InputState` on the active runtime ship.

The input acceptance gate is:

```text
playerCanReceiveInput(playerID, player)
```

The gate returns false when:

```text
player is pending despawn
player session is missing
player session suspension state is active
```

Suspension includes both normal pause and dev freeze:

```text
SuspensionState.Paused
SuspensionState.DevFrozen
```

If the gate rejects input, the current input packet is ignored and does not overwrite the ship's stored input.

### Stored input state

Accepted input is stored on the active runtime ship:

```text
runtime.Ship.Input
```

The packet shape is generated from shared packet data and currently includes:

```text
forward
back
right
left
primary_fire
secondary_fire
```

Movement and shooting are not resolved in `HandlePacket`. They are resolved during simulation stepping.

### Movement consumption

During `Game.Step`, player simulation calls:

```text
stepPlayers
-> motion.AdvanceShipWithMovePolicy(player, delta, bounds, game.playerCanMove(...))
```

`playerCanMove` uses the same suspension and pending-despawn style gate as input acceptance. If movement is not allowed, the motion path clears the ship input and returns without applying movement.

When movement is allowed, the motion package consumes:

```text
Input.Left
Input.Right
Input.Back
Input.Forward
```

and applies rotation, thrust, damping, velocity limiting, and wrapped world movement.

### Weapon-fire consumption

After movement, `stepPlayers` checks stored fire input:

```text
Input.PrimaryFire
Input.SecondaryFire
```

Primary fire routes to:

```text
firePlayerPrimaryWeapon
```

Secondary fire routes to:

```text
firePlayerSecondaryWeapon
```

Both paths are gated by:

```text
worldSimulationOptions.BulletsCanMove()
playerCanShoot(playerID, player)
```

`playerCanShoot` rejects pending-despawn players, missing sessions, suspended players, and players whose primary weapon cooldown is not zero. The weapon-specific fire policy then applies slot-specific cooldown, ammo, and equipped-weapon checks.

The input-routing boundary does not create projectiles directly. It stores input; weapon simulation later consumes that state.

### Pause request routing

A `pause_request` packet routes through `Game.HandlePacket` and calls:

```text
togglePlayerPaused(playerID)
```

The pause path requires both a player session and active runtime ship. It ignores pending-despawn players.

When pausing, the server:

```text
sets session suspension paused state
clears player input
clears player velocity
```

When resuming, the server:

```text
clears player input
sets post-resume invulnerability
```

Networking separately enqueues the pause-state response after routing the pause packet. The simulation player input-routing path owns the mutation; networking owns the outbound enqueue.

### Respawn request routing

A `respawn` packet routes to:

```text
game.respawnPlayer(playerID)
```

before active player lookup. This allows an inactive player with a remaining session and valid cooldown to request respawn.

Respawn itself owns eligibility, safe placement, ship recreation, and camera reattachment. Input routing only dispatches the request to that path.

### Client config routing

A `client_config` packet carries viewport dimensions:

```text
visible_world_width
visible_world_height
```

The game accepts the config only when both values are positive.

Valid config is copied to:

```text
playerSession.Config
cameraView.Config
runtime.Ship.Config
```

Session and camera config may update even before the active player switch, as long as the session or camera view exists. Runtime ship config updates only when the active player entity exists.

This config supports server-side camera/visibility behavior. It does not give the client authority over world bounds, movement, collision, targeting, or gameplay results.

## Protocols and APIs

### Inbound gameplay packet surface

The relevant inbound packet surface is the generated realtime gameplay packet family.

The caller is the client over the WebSocket connection. The networking layer decodes packets and forwards eligible gameplay packets to the current room's game instance. The game server owns authority behind accepted gameplay consequences.

Data crossing this surface includes:

```text
ClientPacket.Type
ClientPacket.Input
ClientPacket.Config
```

Input data is treated as player intent, not authoritative outcome. Config data is treated as a viewport-size report used by server visibility/camera state, not a world-rule change.

### Current player-routed packet table

```text
input
-> Game.HandlePacket
-> active player lookup
-> playerCanReceiveInput
-> runtime.Ship.SetInput

pause_request
-> Game.HandlePacket
-> active player lookup
-> togglePlayerPaused

respawn
-> Game.HandlePacket
-> respawnPlayer
-> returns before active player lookup

client_config
-> Game.HandlePacket
-> positive dimension check
-> session config
-> camera view config
-> active player config when present
```

Targeting packets are excluded from this document even though they are gameplay packets:

```text
set_target_player_request
select_target_at_position_request
clear_target_request
```

Those route through game targeting APIs rather than through player input storage.

## Data ownership

Input routing owns no durable data.

Transient and runtime data touched by this boundary includes:

```text
game.ClientPacket
runtime.InputState
runtime.ClientConfig
runtime.Ship.Input
runtime.Ship.Config
runtime.CameraView.Config
playerSession.Config
playerSession.Suspension
```

Packet source-of-truth lives in:

```text
shared/packets/gameplay.toml
```

Generated game-server packet files include:

```text
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
```

Generated files should not be edited manually.

## Code map

Primary implementation files:

* `services/game-server/internal/game/input.go` - `Game.HandlePacket` dispatch for input, respawn, pause, and client config packets.
* `services/game-server/internal/game/pause.go` - input, movement, shooting, collision-damage, score, pause, and suspension gates.
* `services/game-server/internal/game/simulation_players.go` - per-tick player stepping, camera position update, stored fire input checks, and weapon fire invocation.
* `services/game-server/internal/game/motion/motion.go` - movement consumption of stored input and movement-policy clearing.
* `services/game-server/internal/game/runtime/ship.go` - runtime ship input/config setters, input clearing, pending-despawn state, and ship state projection.
* `services/game-server/internal/game/runtime/state.go` - runtime ship, camera, and entity-store state shapes.
* `services/game-server/internal/game/runtime/suspension.go` - pause/dev-freeze suspension state.

Related implementation files:

* `services/game-server/internal/networking/inbound/gameplay.go` - inbound gameplay packet recognition and handoff to `Game.HandlePacket`.
* `services/game-server/internal/networking/client_packet_router.go` - inbound router wiring after packet decode.
* `services/game-server/internal/networking/player_activation.go` - current game player assignment for active gameplay sessions.
* `services/game-server/internal/networking/player_pause_state.go` - outbound pause-state enqueue after pause request routing.
* `services/game-server/internal/game/session.go` - respawn request handling and player session state used by input routing.
* `services/game-server/internal/game/players.go` - player/session creation and camera view initialization.
* `services/game-server/internal/game/player_weapons.go` - game adapter from stored fire input to weapon fire results.
* `services/game-server/internal/game/simulation.go` - authoritative simulation phase order.

Source-of-truth and generated files:

* `shared/packets/gameplay.toml` - gameplay packet, input, and client config source definitions.
* `services/game-server/internal/game/packets.go` - generated `ClientPacket` and packet type constants.
* `services/game-server/internal/game/runtime/packets_generated.go` - generated `InputState`, `ClientConfig`, and runtime packet state structs.

Important non-ownership boundaries:

* `services/game-server/internal/networking/` owns WebSocket transport and session context.
* `services/game-server/internal/rooms/` owns room membership and room game-instance access.
* `services/game-server/internal/game/targeting.go` owns target request semantics.
* `services/game-server/internal/game/weapons/` owns weapon fire policy.
* `client/scripts/gameplay/input/` owns client-side input polling and routing.
* `client/scripts/networking/outbound/` owns client-side packet send helpers.
* `docs/protocol/` owns protocol-level packet behavior.
* `docs/data/` owns packet source-of-truth and generation pipeline documentation.

## Tests

Relevant focused tests include:

* `services/game-server/tests/game/pause_test.go`
* `services/game-server/tests/game/respawn_test.go`
* `services/game-server/internal/networking/gameplay_packets_test.go`
* `services/game-server/tests/protocol/packetcodec_test.go`
* `services/game-server/internal/game/player_weapons_test.go`
* `services/game-server/tests/game/devtools_test.go`

Current verified behavior includes:

* Fresh players accept input and move.
* Fresh players can shoot from stored input.
* Paused players do not move or shoot.
* Pause requests toggle pause state.
* Pausing clears input and velocity.
* Resuming clears input and grants temporary invulnerability.
* New input is ignored while paused.
* Pause toggles are ignored for dead/inactive players.
* Respawn packets route through respawn request handling.
* Client config packets reach the game instance and update session/camera config.
* Input packet decode preserves movement and fire booleans.

Broader verification should include the game-server Go test suite when changing packet routing, player gates, pause behavior, respawn behavior, movement, weapon fire, or generated packet fields.

## Related docs

* [Game Server Simulation Players](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Game Server Networking](../../networking/!INDEX.md)
* [Inbound Packet Routing](../../networking/inbound-packet-routing.md)
* [Gameplay Network Adapter](../../networking/gameplay-network-adapter.md)
* [Realtime Protocol](../../../../protocol/!INDEX.md)
* [Gameplay Packets](../../../../protocol/stubs/gameplay-packets.md)
* [Packet Schema Pipeline](../../../../data/stubs/packet-schema-pipeline.md)
* [Player Pause And Suspension](player-pause-and-suspension.md)
* [Player Respawn](player-respawn.md)
* [Active Player Avatar State](active-player-avatar-state.md)
* [Player Session State](player-session-state.md)
* [Player Camera View State](player-camera-view-state.md)
* [Weapons And Projectile Fire](../combat/weapons-and-projectile-fire.md)

## Notes

Legacy architecture notes correctly identified that the client collects input while the game server owns authoritative gameplay outcomes. This document narrows that fact to the current game-server player input-routing implementation.

`Game.HandlePacket` is intentionally not the full inbound packet router. It is the game-owned packet mutation surface after networking has already decoded the packet and resolved the current game player.

The current secondary-fire pre-gate uses `playerCanShoot`, which checks primary cooldown before the slot-specific weapon fire policy runs. That is current behavior in this boundary, not a general weapon-system design rule.
