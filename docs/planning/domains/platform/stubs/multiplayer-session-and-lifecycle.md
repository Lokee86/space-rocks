# Multiplayer Session And Lifecycle
Parent index: [Platform Planning](../!INDEX.md)

## Purpose

This doc plans the multiplayer session, connection, admission, room participation, match participation, disconnect, reconnect, and cleanup lifecycle.

It is grounded in the current implemented seams:

- Godot resolves a WebSocket target by session mode before connecting.
- The Go game-server exposes a single `/ws` route.
- `network_client.gd` owns client WebSocket transport.
- `services/game-server/internal/networking` owns WebSocket upgrade, sessions, read/write loops, transport logging, packet routing, and session activation/deactivation.
- `services/game-server/internal/rooms` owns room membership, room lifecycle, room cleanup, and each room's `*game.Game` lifecycle.
- `services/game-server/internal/game` owns live simulation, player lifecycle during a match, scoring, lives, death, respawn, and authoritative state packets.
- Account/auth routing follows the cross-mode routing and player-data rules from legacy design docs.

## Ownership Boundary

This doc owns planning for the multiplayer/session lifecycle across existing networking and rooms seams.

It owns planning for:

- WebSocket session lifecycle
- session identity handoff
- `authenticate_request` / `authenticate_result` flow
- admission decision handoff
- room create/join/leave routing
- room membership lifecycle
- ready/not-ready lifecycle
- start-game lifecycle
- room state transitions
- active game-player routing state
- disconnect behavior
- reconnect seam planning
- return-to-lobby behavior
- room cleanup triggers

It does not own:

- WebSocket packet schema
- packet encoding
- realtime protocol lanes
- match/mode rule policy
- gameplay simulation
- scoring/lives/death/respawn rules
- account/OAuth implementation
- player-data persistence
- matchmaking/discovery metadata
- UI layout

## Existing Runtime Shape

Current client-side ownership:

- `client/scripts/boot/` owns boot/session target selection.
- `SessionBootController` chooses the WebSocket target by session mode.
- `SessionNetworkTarget` maps single-player mode to `SINGLE_PLAYER_WS_URL` and multiplayer mode to `MULTIPLAYER_WS_URL`.
- `client/scripts/networking/network_client.gd` owns WebSocket connect, poll, raw send, raw receive, graceful close, and packet codec use.
- Scene and menu code should not pass raw WebSocket URLs.
- Both single-player and multiplayer currently use the same local `/ws` endpoint during development.

Current server-side ownership:

- `services/game-server/internal/networking` owns WebSocket upgrade, sessions, read/write loops, transport logging, adapter wiring, inbound routing, outbound writes, and session activation/deactivation.
- `services/game-server/internal/rooms` owns room creation, joining, leaving, readiness, lifecycle transitions, cleanup policy, and game instance ownership.
- `services/game-server/internal/game` owns live simulation and match/player lifecycle inside the active game.
- `services/game-server/internal/game/rules` owns match/mode policy decisions from plain facts.
- Rails/API remains outside realtime gameplay and is only consumed through explicit auth/player-data boundaries.

## Current Lifecycle Facts

The server route is:

```text
GET /ws
```

The WebSocket connection itself is session-only.

Connection does not imply room membership.

Room membership does not imply an active game player.

Active ships/game players are created only when `StartGameRequest` succeeds.

`/ws?room_id=...` no longer creates or joins rooms.

Networking routes room/membership packets to the room domain:

- `CreateRoomRequest`
- `JoinRoomRequest`
- `LeaveRoomRequest`
- `SetReadyRequest`
- `StartGameRequest`
- `ReturnToLobbyRequest`

Room lifecycle currently includes:

- `Lobby`
- `Starting`
- `InGame`
- `GameOver`

`Starting` is an admission-closed handoff state for pre-game coordination, future slow-client handling, final readiness/sync steps, and other server work before the room becomes `InGame`.

## Identity And Admission

Current identity states:

- Guest
- Local Profile
- Authenticated Account

Current play modes:

- Local Single-Player
- Online Multiplayer
- Multiplayer Simulation

Admission matrix:

| Mode | Guest | Local Profile | Authenticated Account |
| --- | --- | --- | --- |
| Local Single-Player | allowed | allowed | rejected |
| Online Multiplayer | rejected | rejected | allowed |
| Multiplayer Simulation | rejected by default | rejected by default | allowed |

Current auth/session flow:

```text
websocket upgrade
-> create websocket session
-> session starts as Guest identity
-> optional authenticate_request
-> Rails token verification through authclient
-> authenticate_result
-> SessionIdentity becomes Authenticated Account if token is valid
-> mode/room request
-> admission decision
-> room/game flow
```

If the game-server auth verifier is configured, multiplayer create/join requires an Authenticated Account identity.

If the auth verifier is not configured, local/no-auth multiplayer can still proceed because server-side admission remains authoritative.

Bearer tokens must not become gameplay identity.

The game-server must not read Rails auth tables directly.

## Identifier Boundaries

Current identifier roles:

- `PlayerID` is permanent and player-facing.
- `PlayerID` values are readable labels like `Player-1`, `Player-2`, `Player-3`.
- `PlayerID` must not be converted to UUID.
- `SessionID` is server-internal WebSocket/session identity.
- `SessionID` is a target for the internal UUID upgrade.
- `MemberID` is server-internal room-membership identity.
- `MemberID` is currently UUID v4.
- `MemberID` is reserved as the future disconnect/reconnect seam.
- `MemberID` should not be exposed in normal room snapshot packets.
- `currentGamePlayerID` is networking-owned active-game routing state only.
- `currentGamePlayerID` is not room membership identity and not player-facing identity.
- Account identity is separate from session identity and game player identity.

## Room Participation Flow

Current intended room participation flow:

```text
WebSocket session exists
-> create or join room
-> room membership created
-> owner/member state assigned by rooms
-> ready state updated through SetReadyRequest
-> StartGameRequest validates start preconditions
-> room transitions Lobby -> Starting -> InGame
-> room owns game instance lifecycle
-> game creates active players/ships
-> networking stores currentGamePlayerID for active-game routing
-> game-over policy resolves through game/rules
-> room transitions InGame -> GameOver
-> ReturnToLobbyRequest may reset GameOver -> Lobby
-> leave/disconnect may detach session/member
-> empty rooms schedule cleanup
```

Rooms own membership and lifecycle decisions.

Networking owns packet routing and session field mutation.

Gameplay owns live simulation after match start.

## Disconnect And Reconnect Planning

Current implemented facts:

- WebSocket sessions are connection-scoped.
- Room membership is owned by `rooms`.
- Session activation/deactivation stays in `networking` when it mutates per-connection session fields.
- Empty rooms schedule cleanup after members or active players leave.
- `MemberID` is reserved as the future reconnect seam.
- `MemberID` is not exposed in normal room snapshots.

Planned reconnect work should build around `MemberID`, not `PlayerID`, `SessionID`, `account_id`, or `currentGamePlayerID`.

Reconnect planning should decide:

- whether a disconnected member remains in room membership for a grace window
- whether their active game player remains alive, inactive, frozen, AI-controlled, or eliminated
- whether reconnect can reclaim the same member identity
- whether reconnect is allowed during Lobby, Starting, InGame, and GameOver
- how reconnect interacts with owner selection
- how reconnect interacts with ready state
- how reconnect interacts with match result eligibility
- how reconnect interacts with leaderboards/ranked/trust policy later

Do not implement reconnect by exposing `MemberID` in normal room snapshots.

## Cleanup Planning

Cleanup ownership belongs to rooms.

Current cleanup direction:

- empty rooms schedule cleanup after members/active players leave
- room cleanup uses room-owned cleanup state/timers
- networking may detach session references
- rooms decide when room state should be removed
- gameplay/game instances are stopped or cleared through room lifecycle paths

Cleanup planning should preserve:

- networking does not own room deletion policy
- game does not own room deletion policy
- rooms do not own WebSocket transport
- game instances do not outlive their owning room lifecycle

## Boundary Rules

Networking owns transport and session mutation:

- WebSocket upgrade
- read/write loops
- packet decode/route
- outbound responses
- websocket session fields
- session activation/deactivation
- `currentGamePlayerID` routing state

Rooms own room lifecycle:

- create
- join
- leave
- ready
- owner selection
- start-game transition
- return-to-lobby transition
- game-over room transition
- cleanup policy
- game instance ownership

Game owns active match lifecycle:

- player session counters
- active ships
- lives
- death
- respawn
- score
- match-over facts
- state packets
- player lifecycle status in `StatePacket.player_lifecycle`

Account and identity systems own:

- Guest / Local Profile / Authenticated Account rules
- OAuth/account token policy
- authenticated account identity
- local profile identity
- account/profile routing expectations

Cross-mode routing owns:

- whether a session identity may enter a requested play mode
- data-route implications of Guest, Local Profile, or Authenticated Account

Realtime protocol architecture owns:

- packet lanes
- snapshots
- deltas
- event delivery
- encoding direction

Matchmaking and room discovery owns:

- finding or assigning rooms before join
- public/private discovery metadata
- queue/discovery behavior later

## Related Docs

- [Planning](../../../!INDEX.md)
- [Account And Identity Systems](../account-and-identity-systems.md)
- [Matchmaking And Room Discovery](matchmaking-and-room-discovery.md)
- [Anti-Cheat Policy](game-integrity-policy.md)
- [Realtime Protocol Architecture](../../../protocol/realtime-protocol-architecture.md)
- [Deployment And Packaging](../../technical/stubs/deployment-and-packaging.md)
- [Modes And Match Rules](../../gameplay/modes-and-match-rules.md)
- [Systems Design](../../../systems-design/!INDEX.md)
- [Player Data and Persistence](player-data-and-persistence.md)
- [Account And Identity Current State](../../../../domains/platform/account-and-identity-current-state.md)

## Open Planning Questions

- Should this doc move out of `platform/stubs/` once the lifecycle plan becomes active?
- Which lifecycle state should appear in room snapshots?
- Which lifecycle state should remain server-internal only?
- What is the first reconnect-supported state: Lobby, GameOver, or active InGame?
- Should disconnected InGame members retain active ships, become inactive, or be eliminated?
- How long should reconnect grace windows last?
- How should owner selection behave when the owner disconnects?
- How should `Starting` handle slow clients or failed final readiness?
- Which disconnect/reconnect cases affect match result eligibility?
- Which lifecycle events should become domain events for achievements, results, or diagnostics?
- Which lifecycle diagnostics belong in network logs versus room logs?
- Which lifecycle behavior changes when hosted online multiplayer uses deployed infrastructure?
