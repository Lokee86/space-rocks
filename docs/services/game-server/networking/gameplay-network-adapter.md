# Gameplay Network Adapter

Parent index: [Game Server Networking](./!README.md)

## Purpose

This document covers the game-server inbound gameplay network adapter.

The adapter is the server networking seam that translates already-decoded client gameplay packets into room/game calls. It does not own gameplay rules, simulation behavior, packet decoding, WebSocket transport, or outbound state broadcasting.

## Overview

Gameplay packets reach this adapter after the inbound client packet router has already handled devtools shortcuts, decoded the packet, and given auth, telemetry, and lobby handlers the first chance to consume it.

The gameplay adapter receives:

- the current session
- the current decoded `game.ClientPacket`
- the session's current room
- the session's current game player ID

It then decides whether the packet is a gameplay packet and forwards it to the active room's `GameInstance()`.

There are two handling paths:

```text
input / respawn / client_config
-> require no special target/pause switch
-> forward to Game.HandlePacket(playerID, packet)

target / pause requests
-> require current room and game player ID
-> adapt request-specific fields into game API calls
```

For `input`, `respawn`, and `client_config`, a missing room or game player ID is treated as consumed once the packet type is recognized. This prevents those gameplay packets from falling through to other routing paths.

For target and pause request packets, a missing room or game player ID returns `false`, because the adapter cannot prove it owns the packet without an active gameplay context.

## Code root

`services/game-server/internal/networking/inbound/`

## Runtime surface

The adapter handles these inbound gameplay packet types:

```text
input
respawn
client_config
set_target_player_request
select_target_at_position_request
clear_target_request
pause_request
```

Packet field ownership stays split:

* packet shape and packet type constants come from generated game packet code
* routing ownership lives in networking/inbound
* gameplay mutation lives in the game instance
* target validation and target state mutation live in game targeting code
* pause state broadcast scheduling is requested through the session

## Responsibilities

The gameplay network adapter owns:

* recognizing gameplay packet types after earlier inbound routing stages have passed
* requiring an active room and current game player ID before mutating gameplay state
* forwarding normal gameplay packets to the authoritative game instance
* adapting target packets into game targeting calls
* forwarding pause requests to the game instance
* asking the session to enqueue the local player pause state after pause changes
* keeping networking concerns separate from simulation and targeting rules

## Non-responsibilities

The gameplay network adapter does not own:

* WebSocket connection lifecycle
* raw packet decoding
* packet source-of-truth definitions
* auth handling
* telemetry handling
* lobby room lifecycle handling
* room ownership or room membership rules
* simulation ticking
* player input application rules
* respawn mechanics
* client viewport/config semantics beyond forwarding valid packets
* target validation internals
* pause state implementation
* outbound state packet construction

## Packet handling flow

```text
RouteClientPacket
-> DecodePacket
-> HandleAuth
-> HandleTelemetry
-> HandleLobby
-> HandleGameplay
```

The gameplay handler receives only packets that were not consumed by earlier handlers.

For direct gameplay packets:

```text
input
respawn
client_config
```

the adapter:

1. checks the current room
2. checks the current game player ID
3. forwards the packet to `room.GameInstance().HandlePacket(playerID, packet)`
4. returns `true`

For target request packets:

```text
set_target_player_request
select_target_at_position_request
clear_target_request
```

the adapter:

1. checks the current room
2. checks the current game player ID
3. maps request fields into the game targeting API
4. returns `true` after handing the packet to the game instance

For `pause_request`, the adapter:

1. checks the current room
2. checks the current game player ID
3. forwards the packet to `Game.HandlePacket`
4. calls `session.EnqueuePlayerPauseState()`
5. returns `true`

## Authority and trust boundary

The gameplay adapter trusts the networking session only for session context:

* current room
* current game player ID
* ability to enqueue player pause state

It does not trust the client to own gameplay authority.

Client packets are requests or input reports. Authoritative mutation remains inside the game instance. The adapter may pass packet fields through, but validation and state changes belong to game systems.

Targeting is a clear example:

* the adapter reads `TargetID`, `TargetKind`, `X`, and `Y` from the packet
* the game targeting implementation decides whether the target exists and whether the selected position is valid
* the adapter does not duplicate those targeting rules

## Gameplay mutation handoff

`Game.HandlePacket` owns direct gameplay packet effects:

* `respawn` requests call the game respawn path
* `client_config` updates session/camera/player config when dimensions are valid
* `input` updates player input only when the player can receive input
* `pause_request` toggles player pause state

The adapter should remain thin. New gameplay packet types should be added here only when networking needs to choose the correct authoritative game API. The actual gameplay rule should live under the relevant game package.

## Code map

Primary implementation files:

* `services/game-server/internal/networking/inbound/router.go`

  * Orders inbound packet handling.
  * Calls the gameplay handler after devtools, auth, telemetry, and lobby handlers.

* `services/game-server/internal/networking/inbound/gameplay.go`

  * Owns gameplay packet recognition and adaptation.
  * Bridges session context to the current room game instance.

Related implementation files:

* `services/game-server/internal/networking/inbound/lobby.go`

  * Neighbor inbound adapter for lobby packets.
  * Useful comparison for keeping gameplay and lobby routing separate.

* `services/game-server/internal/game/packets.go`

  * Generated packet type constants and packet field shapes.
  * Includes the client packet fields consumed by this adapter.

* `services/game-server/internal/game/input.go`

  * Implements `Game.HandlePacket`.
  * Owns respawn, client config, input, and pause mutation behavior.

* `services/game-server/internal/game/targeting.go`

  * Owns target selection, target clearing, target existence checks, and click-position validation.

Important non-ownership boundaries:

* `networking/inbound` owns routing and adaptation only.
* `game` owns gameplay authority.
* `rooms` owns room state and room-to-game-instance access.
* generated packet code owns packet names and packet field shapes.
* outbound networking owns state/event delivery back to clients.

Player docs own the actual input rules, pause mechanics, and respawn implementation. This adapter only routes packets into the correct game-side APIs.

Related player docs:

* [Player Input Routing](../simulation/players/stubs/player-input-routing.md)
* [Player Pause And Suspension](../simulation/players/stubs/player-pause-and-suspension.md)
* [Player Respawn](../simulation/players/stubs/player-respawn.md)

## Tests and verification

Expected verification should cover:

```text
go test -buildvcs=false ./services/game-server/internal/networking/inbound/...
go test -buildvcs=false ./services/game-server/internal/game/...
```

Useful behavioral checks:

* input packets reach `Game.HandlePacket`
* respawn packets reach `Game.HandlePacket`
* client config packets reach `Game.HandlePacket`
* target-player requests call `SetPlayerTarget`
* position target requests call `SelectTargetAtPosition`
* clear-target requests call `ClearTarget`
* pause requests call `Game.HandlePacket` and enqueue player pause state
* packets without an active room/player do not mutate gameplay state

## Related docs

* [Game Server Networking](./!README.md)
* [Game Server](../!README.md)

## Notes

This document focuses on the gameplay adapter's current routing and handoff responsibilities.
