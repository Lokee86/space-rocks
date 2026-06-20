# Match Result Reporting

Parent index: [Game Server Integrations](./!README.md)

## Purpose

This document describes the game-server integration boundary that reports authoritative match results into the player-data runtime.

It keeps the game-server reporting path separate from match simulation, client match-results presentation, player-data sink routing, Rails/Postgres persistence, local SQLite persistence, and guest transient stats.

## Overview

The game server owns the authoritative match facts used for match-result reporting. When a room reaches game over, the room lifecycle builds and stores a resolved match summary. The networking lifecycle then reports that resolved summary once through the configured match-result reporter.

The current implemented flow is:

```text
room game-over lifecycle
-> resolved match summary
-> rooms.ReportResolvedMatchResultOnce
-> matchreporting.RuntimeReporter
-> player_data_record_match_result command per player
-> in-process services/player-data runtime sink
```

The game server does not write player-data stores directly. It maps authoritative match facts into player-data commands and sends those commands across the player-data packet boundary. The player-data runtime validates the mode/identity pairing and routes the command to the correct backing behavior.

Related upstream boundaries:

* [Room Match Lifecycle](../rooms/room-match-lifecycle.md)
* [Player Counters](../simulation/players/stubs/player-counters.md)
* [Player Death And Despawn](../simulation/players/stubs/player-death-and-despawn.md)

Match-result reporting owns the reporting integration to player-data, not simulation counter mutation, death flow, room phase transition, or player-data persistence internals.

The current backing behaviors are outside this game-server integration boundary:

* authenticated account -> player-data Rails adapter -> API-server internal match-results endpoint
* local profile -> player-data local store
* guest -> player-data guest behavior

The game server treats accepted duplicate responses from player-data as successful reports. This makes the reporting path safe to retry while the room still holds the same resolved summary.

## Code root

`services/game-server/`

## Responsibilities

* Build the authoritative match summary when the room enters `game_over`.
* Preserve the resolved match summary after game over.
* Generate one match-result record command per player in the resolved summary.
* Generate a stable per-player `result_id` from `match_id` and `game_player_id`.
* Carry authoritative score, ship-death, and winner facts into player-data commands.
* Carry the room member identity context into player-data commands.
* Prefer authenticated-account identity over local-profile identity when both are present.
* Use guest identity when no durable player identity is present.
* Send encoded player-data commands into the configured player-data sink.
* Mark the room result as reported only after the reporter succeeds.
* Leave the room unmarked when reporting fails so a later lifecycle call may retry.
* Log match-result reporting start, skip, failure, and success events.

## Does not own

* Live simulation state.
* Match-over rule calculation inside the game aggregate.
* Client match-results UI.
* Room snapshot presentation shape.
* Player-data store routing.
* Guest/local/account stats mutation.
* Rails internal match-result endpoint behavior.
* Rails/Postgres schema or transactions.
* Embedded SQLite schema or migrations.
* Leaderboards.
* Match-history read APIs.
* General profile CRUD.
* Auth token verification.

## Domain roles

The game-server match-result reporting integration participates in the player-data commit path.

Current roles:

* Authoritative match fact producer.
* Match summary resolver.
* Player-data command mapper.
* In-process player-data runtime caller.
* Retry-safe reporting gate for resolved room summaries.

The game server remains the gameplay authority. Player-data remains the identity and store-routing authority. API-server remains the authenticated-account persistence owner.

## Protocols and APIs

The game-server reporting surface is not a public HTTP endpoint and not a client WebSocket packet.

It is an internal service boundary between the game-server room lifecycle and the in-process player-data runtime. The game server consumes the `rooms.MatchResultReporter` interface, and the concrete runtime reporter sends generated player-data packets through a `PlayerDataSink`. The data crossing the boundary is a trusted resolved match result: match ID, result ID, play mode, identity, score, ship deaths, and win flag. The boundary does not own player-data validation, storage selection, aggregate stat mutation, or presentation-safe room snapshot output.

### Room reporter interface

`services/game-server/internal/rooms/match_result_reporter.go` defines the reporting interface:

```text
ReportMatchResult(summary playerdata.MatchResultSummary) error
```

`rooms.ReportResolvedMatchResultOnce` calls this interface after a room has a resolved summary.

### Runtime reporter sink

`services/game-server/internal/matchreporting/runtime_reporter.go` defines the sink-facing boundary:

```text
HandlePlayerDataCommand(payload []byte) ([]byte, error)
```

`RuntimeReporter` encodes each generated command with the player-data codec, sends it to the sink, and decodes the player-data result response.

### Player-data command shape

`BuildRecordMatchResultCommands` maps each player summary into a generated player-data packet:

```text
type: player_data_record_match_result
result_id: <match_id>:<game_player_id>
match_id: <match_id>
identity:
  identity_kind: authenticated_account | local_profile | guest
  account_id: <account UUID, if authenticated>
  local_profile_id: <local profile ID, if local profile>
context:
  play_mode: single_player | multiplayer
score: <authoritative score>
ship_deaths: <authoritative ship deaths>
won: <resolved winner flag>
```

The response packet is expected to be `player_data_record_match_result_result`.

A response with `accepted: true` is success. A duplicate response is still success when it is accepted. A response with `accepted: false`, an invalid response payload, or a sink error causes reporting to fail.

## Data ownership

### Game-server-owned data

The game server owns the match facts it reports:

* `match_id`
* match mode
* per-player `game_player_id`
* score
* ship deaths
* winner flag
* room-member account ID, when present
* room-member local profile ID, when present

The room lifecycle stores the resolved summary once the room transitions to game over. A later attempt to mark game over does not rebuild the summary.

### Player-data-owned data

The player-data runtime owns:

* mode/identity validation
* identity-based store routing
* guest behavior
* local-profile storage behavior
* authenticated-account Rails adapter behavior
* aggregate stats updates
* duplicate handling below the packet boundary
* returned normalized stats

The game server does not choose the backing store directly.

### API-server-owned data

For authenticated accounts, the API server owns Rails/Postgres persistence. It stores authenticated-account aggregate stats and match-result rows behind the internal match-results endpoint.

The game server does not call that endpoint directly. The player-data Rails adapter owns that HTTP call.

## Reporting lifecycle

### Match start reset

When a room starts a new match, `BeginNextMatch` increments the room match number, sets the current match ID, clears the previous resolved summary, and clears the reported flag.

Current match IDs use this shape:

```text
<room_id>-match-<number>
```

### Match-over transition

The room game-over lifecycle checks whether the authoritative game is complete. When the room transitions from `in_game` to `game_over`, it builds the resolved summary if one does not already exist.

The summary builder currently:

* reads player match facts from the game aggregate
* derives `single_player` mode when the room is not joinable
* derives `multiplayer` mode otherwise
* copies account/local-profile identity from room membership by game player ID
* builds winner flags through `playerdata.BuildMatchResultSummary`

Multiplayer winner resolution currently marks a unique highest-score player as the winner. Tied highest scores award no wins. Single-player summaries clear winner flags.

### Normal reporting path

The gameplay lifecycle tick advances the room game-over lifecycle. When that transition succeeds, it calls `rooms.ReportResolvedMatchResultOnce`.

`ReportResolvedMatchResultOnce`:

* returns false for a nil room
* uses a noop reporter when no reporter is provided
* skips when the room has already been marked reported
* skips when the room has no resolved summary
* calls the configured reporter
* leaves the room unmarked if the reporter fails
* marks the room reported only after the reporter succeeds

### Room-exit reporting path

The websocket session also attempts to report a resolved match result before requested room leave or disconnected room cleanup.

This protects against losing an already-resolved summary during room exit. It still uses `ReportResolvedMatchResultOnce`, so it follows the same reported flag and retry rules.

## Failure and idempotency model

A successful report marks the room result as reported.

A failed report does not mark the room result as reported. The resolved summary remains on the room, so later lifecycle paths may retry.

The reporter treats accepted duplicate player-data responses as success. Duplicate handling is owned below the game-server reporting boundary, using the generated `result_id`.

The game server logs reporting failures but does not mutate persistent stats itself.

## Code map

Primary game-server files:

* `services/game-server/cmd/game-server/main.go`
* `services/game-server/cmd/game-server/player_data_http.go`
* `services/game-server/internal/matchreporting/mapper.go`
* `services/game-server/internal/matchreporting/runtime_reporter.go`
* `services/game-server/internal/rooms/match_result_reporter.go`
* `services/game-server/internal/rooms/lifecycle_tick.go`
* `services/game-server/internal/rooms/room_lifecycle.go`
* `services/game-server/internal/rooms/room_match.go`
* `services/game-server/internal/rooms/room_match_access.go`
* `services/game-server/internal/rooms/room_match_summary.go`
* `services/game-server/internal/networking/websocket.go`
* `services/game-server/internal/networking/websocket_gameplay_tick.go`

Game-server match-summary contract files:

* `services/game-server/internal/playerdata/types.go`
* `services/game-server/internal/playerdata/summary.go`
* `services/game-server/internal/playerdata/resolve.go`

Related generated/source files:

* `shared/player_data/match_result.toml`
* `shared/packets/player_data.toml`
* `services/player-data/protocol/packets.go`

Related player-data implementation files:

* `services/player-data/playerdata/dispatcher.go`
* `services/player-data/playerdata/mode_policy.go`
* `services/player-data/playerdata/store_router.go`
* `services/player-data/playerdata/rails_store.go`
* `services/player-data/playerdata/guest_memory_store.go`
* `services/player-data/playerdata/embeddedsqlite/sqlite_store.go`

Important non-ownership boundaries:

* `services/player-data/playerdata/store_router.go` owns identity-based store selection.
* `services/player-data/playerdata/dispatcher.go` owns player-data packet dispatch and validation responses.
* `services/api-server/app/controllers/internal/player_data/match_results_controller.rb` owns authenticated-account match-result persistence intake.
* `services/api-server/app/services/player_stats/apply_match_result.rb` owns Rails aggregate stat mutation.
* `services/game-server/internal/networking/room_snapshot.go` owns presentation-safe room snapshot projection and intentionally excludes account/local-profile IDs from client match-result presentation.

## Tests

Primary game-server tests:

* `services/game-server/internal/matchreporting/mapper_test.go`
* `services/game-server/internal/matchreporting/runtime_reporter_test.go`
* `services/game-server/internal/rooms/lifecycle_tick_test.go`
* `services/game-server/internal/rooms/room_lifecycle_test.go`
* `services/game-server/internal/rooms/room_match_summary_test.go`
* `services/game-server/internal/networking/websocket_test.go`
* `services/game-server/internal/networking/room_snapshot_test.go`
* `services/game-server/internal/playerdata/summary_test.go`
* `services/game-server/internal/playerdata/resolve_test.go`
* `services/game-server/internal/playerdata/types_test.go`

Related player-data tests:

* `services/player-data/playerdata/dispatcher_test.go`
* `services/player-data/playerdata/store_router_test.go`
* `services/player-data/playerdata/rails_store_test.go`
* `services/player-data/playerdata/guest_memory_store_test.go`
* `services/player-data/playerdata/embeddedsqlite/sqlite_store_test.go`

Related API-server tests:

* `services/api-server/test/controllers/internal/player_data/match_results_controller_test.rb`
* `services/api-server/test/services/player_stats/apply_match_result_test.rb`
* `services/api-server/test/models/player_match_result_test.rb`
* `services/api-server/test/models/player_stat_test.rb`

## Related docs

* [Game Server Integrations](./!README.md)
* [Game Server](../../!README.md)
* [Game Server Rooms](../../rooms/!README.md)
* [Game Server Networking](../../networking/!README.md)
* [Player Data](../../../player-data/!README.md)
* [Player Stats And Match Results](../../../api-server/player-stats-and-match-results.md)

## Notes

This document intentionally stays at the game-server integration boundary.

The client match-results window uses presentation-safe room snapshot data. That data path is related, but it is not the same as durable match-result reporting.
