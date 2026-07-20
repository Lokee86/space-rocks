---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-708c-9e52-0f1b25c9f3f5
document_type: general
policy_exempt: false
summary: This document describes the game-server room match lifecycle boundary.
---
# Room Match Lifecycle

Parent index: [Game Server Rooms](./!INDEX.md)

## Purpose

This document describes the game-server room match lifecycle boundary.

It covers how a room starts a match, attaches a game instance, moves through room lifecycle states, detects authoritative match completion, stores resolved match results, reports those results once, and returns to the lobby.

## Overview

Room match lifecycle lives in `services/game-server/internal/rooms/`.

The room package owns room state transitions and match-owned room state. The game package owns simulation and match-over rule evaluation. Networking owns WebSocket session routing, active-player activation, room snapshot broadcasts, and lifecycle tick calls that observe room/game state.

The current room match state path is:

```text
Lobby
-> Starting
-> InGame
-> GameOver
-> Lobby
```

A normal multiplayer match starts from a lobby room when the owner requests start and all connected members are ready. A single-player match starts by creating a non-joinable room with one member, then immediately starting it.

A started room owns one `*game.Game` instance through `roomMatch`. `BeginNextMatch` increments the match number and ID, resets `roomMatch.activeSessionIDs` and the active count through `ResetActiveSessions`, resets `matchResultReported` and `matchResultReporting`, clears the prior summary, and does not rewrite member `PlayerID` values.

Current match IDs use this shape:

```text
<room_id>-match-<number>
```

When the game aggregate reports that the match is over, the room transitions from `InGame` to `GameOver`. During that transition, the room builds and stores a resolved match summary if one has not already been stored. That summary is then available to room snapshots and the match-result reporting integration.

The room does not stop the game instance when it enters `GameOver`. The game instance is stopped and cleared when the room returns to lobby or when room cleanup removes the room.

Related player and runtime boundaries:

* [Player Death And Despawn](../simulation/players/player-death-and-despawn.md)
* [Player Counters](../simulation/players/player-counters.md)
* [Lane Packet Projection](../simulation/runtime/lane-packet-projection.md)

Room match lifecycle owns phase transitions and match completion. Player death, lives mutation, and lane-native realtime projection live in their narrower simulation boundaries.

## Code root

`services/game-server/internal/rooms/`

## Responsibilities

* Own room-level match state through `roomMatch`.
* Store the active `*game.Game` instance for the room.
* Track active player count plus authoritative per-match `activeSessionIDs` ownership for room cleanup and network-session coordination.
* Track match number and current match ID.
* Reset active-session ownership, both reporting flags, and the prior resolved summary when a new match begins.
* Validate multiplayer start through room start rules.
* Allow single-player start from a lobby room with at least one member.
* Move room state from `Lobby` to `Starting` to `InGame` during start.
* Detect game completion by asking the game aggregate for its match decision.
* Move room state from `InGame` to `GameOver` when the game is complete.
* Build and preserve a resolved match summary at game over.
* Keep an existing resolved summary if one has already been stored.
* Expose resolved summary data for room snapshot projection and reporting.
* Report resolved match results once through a `MatchResultReporter` using an in-flight claim, with the external reporter call outside `Room.mu`.
* Leave the result unreported and release the matching claim when reporting fails so a later path may retry.
* Reset room members to not-ready when returning to lobby.
* Clear room state under `Room.mu` and call `Game.Stop` outside `Room.mu` when returning to lobby.
* Reject invalid lifecycle transitions with room domain errors.

## Does not own

* WebSocket transport.
* Per-connection session state.
* Active game player ID assignment on WebSocket sessions.
* Room snapshot encoding or broadcast.
* Gameplay simulation phase order.
* Match-over rule calculation inside the game aggregate.
* Player movement, weapons, collisions, score mutation, lives, death, or respawn.
* Client match-end UI.
* Client match-results presentation.
* Player-data sink routing.
* Rails/Postgres persistence.
* Local SQLite profile persistence.
* Guest stat storage.
* Room cleanup timer ownership.

## Domain roles

Room match lifecycle participates in the player experience session flow.

The room is the server-side authority for whether a room is in lobby, starting, in game, or game over. The game aggregate is the server-side authority for whether the simulated match is complete. Networking is the live transport/session adapter that calls the room lifecycle, activates connected players, broadcasts snapshots, and reports resolved results.

Important lifecycle boundaries:

* A WebSocket connection does not imply room membership.
* Room membership does not imply active gameplay participation.
* Active game players are created only after a successful room start.
* Room state can enter `GameOver` only from `InGame`.
* Return to lobby can happen only from `GameOver`.
* Match results are resolved from authoritative game facts, not from client UI state.
* Match-result reporting is retry-safe because the room stores a resolved summary, a reported flag, and a `matchResultReporting` claim; failure releases the matching claim for retry and stale completion cannot affect a later match.

## Protocols and APIs

The room match lifecycle is not a public protocol surface by itself. It is an internal service boundary consumed by networking handlers and lifecycle tick code.

Clients trigger this behavior indirectly through generated realtime packets handled by networking:

```text
start_game_request
start_single_player_request
return_to_lobby_request
leave_room_request
```

Networking decodes those packets, validates session context, calls room manager or room aggregate methods, and broadcasts `room_snapshot` or sends `room_error` packets. Room lifecycle owns the authority behind successful state transitions. Networking owns packet decode, live session fields, activation/deactivation, enqueueing, and broadcast.

### Internal room lifecycle calls

`RoomManager.StartRoomGame(roomID, sessionID)`:

* normalizes the room ID
* finds the room
* calls `Room.StartGameForSession(sessionID, game.New)`
* returns a room domain error when lookup, membership, or start validation fails

`Room.StartGameForSession(sessionID, newGame)`:

* under `Room.mu`, validates the exact session/member and start rules, transitions `Lobby` to `Starting`, and captures the reserved game
* releases `Room.mu` for the game factory and `Game.Start` (`startGameCall`)
* reacquires `Room.mu` and commits only if the state and game reservation still match; installs the created game, moves `Starting` to `InGame`, and begins the next match
* a superseded start stops only a game it does not own; concurrent starts have exactly one successful reservation/commit

`Room.StartGameForMember(playerID, newGame)` remains the compatibility/internal player-ID entry point.

`RoomManager.CreateStartedSinglePlayerRoom(sessionID)`:

* creates a non-joinable single-player room
* adds the session as the room member
* delegates to `Room.StartSinglePlayerGame`

`Room.StartSinglePlayerGame(newGame)`:

* requires a lobby room with at least one member
* moves `Lobby` to `Starting`
* creates a game instance when needed
* starts the game simulation loop
* moves `Starting` to `InGame`
* begins the next match and assigns a match ID

`Room.MarkGameOverIfComplete()`:

* returns false unless the room is `InGame`
* asks the game aggregate whether the match is over
* captures state, game, match, joinability, and member values under `Room.mu`; calls `Game.MatchDecision` and `Game.PlayerMatchFacts` outside `Room.mu`; reacquires the lock and commits only if the captured state, game, and match still match and the room remains `InGame`

`Room.MarkGameOver()`:

* accepts only `InGame`
* builds and stores a resolved match summary if none exists
* changes room state to `GameOver`

`MarkGameOver` retains supported nil-game behavior: an `InGame` room with a nil game can transition to `GameOver` without a summary.

`RoomManager.ReturnRoomToLobby(roomID, sessionID)`:

* finds the room
* calls `Room.ResetToLobbyForSession(sessionID)`

`Room.ResetToLobbyForSession(sessionID)`:

* under one `Room.mu` acquisition, resolves the session, validates `GameOver`, clears readiness, clears the game pointer, resets active-session ownership, and changes room state to `Lobby`
* calls `Game.Stop` outside `Room.mu`

`Room.ResetToLobby(playerID)` remains the compatibility/internal player-ID entry point.

Networking calls `deactivateRoomPlayers` after a successful return-to-lobby request. That clears connected sessions' `SessionContext.GamePlayerID` values and calls `ResetActivePlayerCount`; the room reset already clears active-session ownership, so networking cleanup is idempotent rather than the sole owner of the reset.

### Game-over lifecycle tick

Networking runs `tickSessionGameplayLifecycle` for each WebSocket session. On each server tick, it only proceeds when the session has an active game player and the room can send gameplay presentation state.

When `rooms.TickRoomGameOverLifecycle` advances the room to game over, it:

* logs the game-over transition
* broadcasts a room snapshot
* returns true to the networking tick

The networking tick then calls `rooms.ReportResolvedMatchResultOnce`.

### Match-result reporting gate

`ReportResolvedMatchResultOnce` is the room-owned reporting gate. It does not own the concrete player-data sink. It owns the once-only room-level decision around a stored resolved summary.

The function:

* returns false for a nil room
* substitutes a noop reporter when no reporter is provided
* under `Room.mu`, rejects an already-reported or in-flight result, requires a summary, captures the room match ID and summary match ID, deep-copies `Players`, and sets `matchResultReporting=true`
* calls the reporter outside `Room.mu`
* reacquires `Room.mu` and clears or commits only when the room match ID, in-flight flag, and current summary match ID still match the claim
* on failure leaves the result unreported and allows retry by clearing the matching claim
* on success marks the result reported only for the matching claim
* prevents an old completion from clearing a newer claim, and prevents reporter mutation from aliasing the stored summary

The same reporting gate is also called before requested room leave and disconnected room cleanup so an already-resolved match result is not lost when the member exits.

## Data ownership

### Room-owned data

The room match lifecycle owns:

* room state
* active game instance reference
* active player count
* `activeSessionIDs`
* match number
* current match ID
* resolved match summary
* match-result-reported flag
* `matchResultReporting` in-flight flag

`roomMatch.BeginNextMatch` increments the match number and ID, calls `ResetActiveSessions` to clear `activeSessionIDs` and active count without rewriting member `PlayerID` values, resets `matchResultReported` and `matchResultReporting`, and clears the prior summary.

### Game-owned data

The game aggregate owns the authoritative gameplay facts used by room lifecycle:

* match-over decision
* per-player match facts
* score
* ship deaths
* active ship presence
* remaining lives
* pending respawn versus eliminated classification

`Game.MatchDecision()` evaluates whether the match is over. `Game.PlayerMatchFacts()` exposes the score and ship-death facts used to build the room's resolved summary.

### Player-data-owned data

The player-data contracts own the resolved summary shape and winner resolution helpers used by the room summary builder:

* `MatchMode`
* `PlayerMatchSummary`
* `MatchResultSummary`
* `BuildMatchResultSummary`
* `ResolveWinners`

The room summary builder selects `single_player` mode when the room is not joinable and `multiplayer` mode otherwise.

Current winner behavior:

* single-player summaries clear all winner flags
* multiplayer summaries mark the unique highest-score player as the winner
* tied highest scores produce no winner

### Networking-owned data

Networking owns live session data around the room lifecycle:

* WebSocket session ID
* synchronized `SessionContext{Room, RoomID, GamePlayerID}`
* room session registry
* outbound room snapshot and error packets

Networking activates connected room members into game players after a successful start. The room lifecycle does not assign `SessionContext.GamePlayerID` on WebSocket sessions.

### Presentation-safe room snapshot data

Room snapshots include a presentation-safe match result summary derived from the room's resolved summary.

The snapshot result includes:

```text
match_id
mode
players[].game_player_id
players[].score
players[].ship_deaths
players[].won
```

It intentionally excludes account IDs and local profile IDs.

## Lifecycle states

Room lifecycle states are defined in `services/game-server/internal/rooms/constants.go`.

Current states:

```text
Lobby
Starting
InGame
GameOver
Closed
```

Current transition rules:

```text
Lobby -> Starting
Starting -> InGame
InGame -> GameOver
GameOver -> Lobby
```

`Closed` exists as a room state constant, but normal match lifecycle does not transition through it.

Invalid transition behavior:

* starting from `Starting` or `InGame` returns `room_in_game`
* starting from other non-lobby states returns `invalid_room_state`
* marking game over from any non-`InGame` state returns `invalid_room_state`
* returning to lobby from any non-`GameOver` state returns `invalid_room_state`

## Code map

Primary room lifecycle files:

* `services/game-server/internal/rooms/room.go` - Room aggregate root and owned lifecycle components.
* `services/game-server/internal/rooms/constants.go` - Room states, limits, and room error codes.
* `services/game-server/internal/rooms/room_match.go` - Game instance, active player count, match ID, resolved summary, and reported-state storage.
* `services/game-server/internal/rooms/room_match_access.go` - Locked room accessors for game instance, active-player count, current match ID, resolved summary, and reporting state.
* `services/game-server/internal/rooms/room_lifecycle.go` - Room state transitions, game start, game-over transition, single-player start, and return-to-lobby behavior.
* `services/game-server/internal/rooms/lifecycle.go` - Room manager lifecycle entry points for start, single-player start, and return to lobby.
* `services/game-server/internal/rooms/lifecycle_tick.go` - Game-over lifecycle tick and once-only match-result reporting claim gate.
* `services/game-server/internal/rooms/room_membership_mutations.go` - Per-match active-session ownership and activation.
* `services/game-server/internal/rooms/room_projections.go` - Coherent `GameplayContext` and room snapshots.
* `services/game-server/internal/rooms/room_lifecycle_concurrency_test.go` - Lifecycle and reporting concurrency coverage.
* `services/game-server/internal/rooms/room_membership_mutations_test.go` - Activation and deactivation mutation coverage.
* `services/game-server/internal/rooms/room_match_summary.go` - Resolved match summary builder.
* `services/game-server/internal/rooms/match_result_reporter.go` - Match-result reporting interface and noop reporter.
* `services/game-server/internal/rooms/room_lobby.go` - Start preconditions and ready-state mutation.
* `services/game-server/internal/rooms/room_rule_adapter.go` - Pure room-rule decision to room-domain-error adapter.
* `services/game-server/internal/rooms/roomrules/start.go` - Pure start-rule policy.

Related networking files:

* `services/game-server/internal/networking/room_handlers.go` - Adapts start, single-player start, return-to-lobby, and leave packets into room lifecycle calls.
* `services/game-server/internal/networking/player_activation.go` - Activates and deactivates active game players around room lifecycle transitions.
* `services/game-server/internal/networking/websocket_gameplay_tick.go` - Calls room game-over lifecycle tick and match-result reporting.
* `services/game-server/internal/networking/websocket.go` - Reports resolved results before requested leave or disconnect cleanup.
* `services/game-server/internal/networking/room_snapshot.go` - Projects resolved match summaries into presentation-safe room snapshots.

Related game files:

* `services/game-server/internal/game/game.go` - Game construction, start, and stop behavior.
* `services/game-server/internal/game/match.go` - Match decision and player match facts exposed to room lifecycle.
* `services/game-server/internal/game/rules/match.go` - Pure match-over and player lifecycle classification policy.
* `services/game-server/internal/game/simulation.go` - Simulation step behavior, including reduced stepping after match over.

Related player-data contract files:

* `services/game-server/internal/playerdata/types.go` - Match mode and match summary structures.
* `services/game-server/internal/playerdata/summary.go` - Match summary builder.
* `services/game-server/internal/playerdata/resolve.go` - Winner resolution policy.

Important non-ownership boundaries:

* `services/game-server/internal/networking` owns WebSocket session routing, not room lifecycle authority.
* `services/game-server/internal/game` owns simulation and match-over evaluation, not room state transitions.
* `services/game-server/internal/matchreporting` owns concrete match-result sink mapping, not room lifecycle state.
* `services/player-data` owns stats mutation and store routing, not live room state.
* `services/api-server` owns authenticated-account persistence, not live match lifecycle.

## Tests

Primary room lifecycle tests:

* `services/game-server/internal/rooms/room_lifecycle_test.go`

  * Verifies multiplayer start transitions, owner/ready rejection, single-player start, match ID advancement, reported-state reset, return-to-lobby behavior, and game-over summary storage.
* `services/game-server/internal/rooms/lifecycle_tick_test.go`

  * Verifies game-over lifecycle ticking, room snapshot broadcast trigger, once-only reporting, retry behavior after reporting failure, and nil/missing-summary cases.
* `services/game-server/internal/rooms/room_match_summary_test.go`

  * Verifies single-player guest/local-profile summaries, multiplayer account summaries, winner selection, tied-winner clearing, identity preservation after player ID rekey, and no summary rebuild after game over.
* `services/game-server/internal/rooms/room_lobby_test.go`

  * Verifies lobby ready/start behavior and that Player-ID reuse cannot authorize a missing session.
* `services/game-server/internal/rooms/roomrules/start_test.go`

  * Verifies pure start-rule policy for ownership, readiness, disconnected members, and invalid states.
* `services/game-server/internal/rooms/room_lifecycle_concurrency_test.go`

  * Verifies external factory/start/stop/decision/facts calls without `Room.mu`, concurrent start with one winner, concurrent game-over with one transition, existing-summary preservation, nil-game behavior, reporting claims, failure retry, stale-claim isolation, and deep-copy mutation isolation.
* `services/game-server/internal/rooms/room_membership_mutations_test.go`

  * Verifies idempotent activation, exact active counts, consecutive-match ownership reset, and session deactivation.
* `services/game-server/internal/rooms/room_lifecycle_test.go`

  * Verifies Player-ID reuse cannot authorize a missing session for ready/start/reset.

Related networking tests:

* `services/game-server/internal/networking/player_activation_test.go`

  * Verifies rollback and consecutive-match activation of the same identity while preserving account identity.
* `services/game-server/internal/networking/room_snapshot_test.go`

  * Verifies match-result projection into room snapshots.
* `services/game-server/internal/networking/websocket_test.go`

  * Verifies resolved match results are reported before requested leave or disconnect.
* `services/game-server/internal/networking/gameplay_packets_test.go`

  * Verifies single-player start packet behavior and local-profile ID attachment.

Related game tests:

* `services/game-server/internal/game/simulation_match_over_test.go`

  * Verifies post-match-over simulation behavior does not continue normal asteroid spawning and remains cleanup-safe.

Suggested verification command:

```text
go test -buildvcs=false ./services/game-server/internal/rooms ./services/game-server/internal/networking ./services/game-server/internal/game
```

## Related docs

* [Game Server Rooms](./!INDEX.md)
* [Game Server](../!INDEX.md)
* [Game Server Networking](../networking/!INDEX.md)
* [Room Network Adapter](../networking/room-network-adapter.md)
* [Match Result Reporting](../integrations/match-result-reporting.md)
* [Game Server Simulation](../simulation/!INDEX.md)
* [Player Data](../../player-data/!INDEX.md)
* [Client Match End Flow](../../client/match-end-flow/!INDEX.md)
* [Protocol](../../../protocol/!INDEX.md)
* [Data](../../../data/!INDEX.md)

## Notes

Legacy documentation supplied two still-current ownership rules: the client does not own authoritative match lifecycle, and WebSocket connection, room membership, and active gameplay participation are separate states.

Room game-over state and local player elimination are different concepts. The room reaches `GameOver` only when the server room lifecycle observes that the authoritative game match decision is complete.
