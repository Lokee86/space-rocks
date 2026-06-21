# Match End And Results Flow

Parent index: [Player Experience](./!INDEX.md)

## Purpose

This document describes the current cross-system match end and results flow for Space Rocks.

It documents how an authoritative match ends, how final result facts become visible to the player, how result navigation intent leaves gameplay, and how durable or transient player stats are updated through the player-data boundary.

## Overview

Match end has two separate player-facing states:

```text
Local elimination
= the local player has reached zero lives.

Authoritative room match-over
= the game server room has entered GameOver.
```

Local elimination is presentation-only. It can update the HUD, set local game-over menu state, and request game-over audio, but it does not show match results.

Authoritative room match-over is the valid source for final match results. The game server detects that the match is complete from server-owned gameplay facts, moves the room from `InGame` to `GameOver`, stores a resolved match summary, broadcasts a room snapshot containing a presentation-safe `match_result`, and reports durable result data through the player-data runtime.

The client presents match results only after observing authoritative room match-over. The room snapshot result is cached by the client session flow, adapted into result rows by the client match-end flow, and rendered by the match-results UI.

The durable result commit path is separate from the visible result window. The result window is not a persistence surface. It displays presentation-safe rows and emits navigation intent.

## Participating systems

* Game server simulation - owns live gameplay facts such as player lifecycle, score, deaths, lives, and match-over rule evaluation.
* Game server rooms - owns room state, match lifecycle transitions, current match ID, resolved match summary storage, and once-only result reporting state.
* Game server networking - owns WebSocket packet routing, room snapshot broadcast, lifecycle tick integration, and return-to-lobby request handling.
* Game server match reporting - maps the resolved match summary into player-data match-result commands.
* Player-data runtime - validates mode and identity pairing, routes result writes by identity kind, applies transient or durable stat updates, and returns normalized stats.
* API server - owns authenticated-account match-result persistence and aggregate stat mutation for account-backed multiplayer results.
* Client session flow - caches room state and the latest valid room snapshot match result.
* Client match-end flow - distinguishes local elimination from authoritative room match-over and coordinates match-end presentation.
* Client match-results UI - mounts the visible result window, renders rows, and forwards result-window button intent.
* Client gameplay menu and HUD flows - own HUD visibility, match-over menu mode, and gameplay-session UI behavior.

## Authority boundaries

### Match completion authority

The game server is the authority for match completion.

The current match-over rule evaluates server-owned player lifecycle facts. A match is over when every player is classified as eliminated. A player is not eliminated while they still have an active ship or remaining lives that can lead to respawn.

The client does not decide that a match is over. Local player elimination is not enough to show final match results.

### Room lifecycle authority

Rooms own the authoritative room transition into `GameOver`.

When a started match begins, the room assigns a current match ID using this shape:

```text
<room_id>-match-<number>
```

When the game reports completion, the room transitions from `InGame` to `GameOver`. During that transition it builds and stores the resolved match summary if one has not already been stored.

The room keeps that resolved summary for both presentation snapshot projection and durable result reporting. It also tracks whether the summary has already been reported.

Returning to lobby is a room lifecycle action. It can happen only from `GameOver`. The game instance is stopped and cleared when the room returns to lobby, not when the result window appears.

### Result calculation authority

The game server owns the current result facts:

```text
match_id
mode
game_player_id
score
ship_deaths
won
```

Score and death counts come from server gameplay facts.

Winner resolution is currently mode-specific:

* Single-player clears winner flags.
* Multiplayer marks the unique highest-score player as the winner.
* A tied highest score produces no winner.

The client displays only the result projection it receives. It does not calculate scores, deaths, or winners.

### Presentation authority

The client owns only presentation and route intent.

The current visible result table displays:

```text
PLAYER | DEATHS | SCORE
```

The client result rows may carry `won`, but the current result row UI does not display win/loss state.

The presentation payload intentionally excludes durable identity internals such as account IDs and local profile IDs. Those are used by the durable reporting path, not by the result window.

### Durable player-data authority

Player-data owns stat routing and stat mutation below the game-server reporting boundary.

The game server reports one match-result command per player. Player-data validates that the submitted identity kind is allowed for the play mode, routes the write, applies aggregate stat behavior, and returns a result response.

Current routes:

| Result identity       | Mode          | Player-data route              |
| --------------------- | ------------- | ------------------------------ |
| Guest                 | Single-player | Transient guest memory         |
| Local Profile         | Single-player | Local profile store            |
| Authenticated Account | Multiplayer   | Rails/API-backed account store |

The game server does not write SQLite, Rails, or Postgres player-data tables directly.

### Account persistence authority

The API server owns authenticated-account result persistence.

For account-backed multiplayer results, player-data routes through the Rails adapter to the API server internal match-results endpoint. Rails stores accepted match-result rows and updates aggregate account stats in a transaction.

The API server rejects unknown users and invalid input. Duplicate `result_id` values are treated as duplicate results rather than reapplied stat mutations.

## Flow summary

### 1. Match starts

A room starts a match from lobby state.

For single-player, the server creates a non-joinable room and starts the game immediately for the local session.

For multiplayer, the room starts only after the owner starts and room start rules pass.

At match start, the room:

```text
increments match number
sets current match ID
clears previous resolved summary
clears match-result-reported state
moves through Starting into InGame
```

### 2. Local elimination may occur before room match-over

When the local player reaches zero lives, the server emits death/lives facts through gameplay event flow.

The client handles this as local elimination:

```text
self-death event with lives == 0
-> client death flow delegates to MatchEndFlow
-> HUD receives final lives and game-over presentation
-> gameplay menu enters game-over presentation
-> game-over audio may be requested
```

This does not show match results.

The room may still not be in `GameOver`, especially in multiplayer where other players may still be active or pending respawn.

### 3. Server detects authoritative room match-over

The game server lifecycle tick asks the room whether the game has reached match-over.

The room asks the game aggregate for the match decision. If the game is complete, the room transitions to `GameOver`.

During the transition:

```text
game match decision is complete
-> room MarkGameOver runs
-> room builds resolved match summary
-> room stores resolved summary once
-> room state becomes GameOver
```

The stored summary is not rebuilt on repeated game-over handling.

### 4. Room snapshot exposes presentation-safe results

After room game-over is detected, networking broadcasts a room snapshot.

The snapshot contains:

```text
room_state: game_over
match_result:
  match_id
  mode
  players:
    game_player_id
    score
    ship_deaths
    won
```

The snapshot result is presentation-safe. It excludes account IDs and local profile IDs.

The client session flow caches a match result only when the snapshot contains a dictionary with a non-empty `match_id`. If a later snapshot does not contain a valid result, the cached result is cleared so stale results do not appear in later sessions.

### 5. Client presents match results

The client match-end flow observes room state through the room-state provider.

When the room state is `GameOver`, the client handles authoritative room match-over once:

```text
room state provider returns GameOver
-> MatchEndFlow guards against repeated handling
-> HUD hides and locks for match-over
-> gameplay menu enables match-over overlay behavior
-> game-over audio may be requested
-> cached match_result is read
-> result player entries become presentation rows
-> MatchResultsFlow mounts the result window
```

Repeated `GameOver` snapshots must not remount duplicate result windows.

If there is no valid result provider, no match result, or no player array, the result window can still open with empty rows.

### 6. Result-window intent leaves gameplay

The result window emits intent only.

Current button behavior:

| Button       | Single-player intent | Multiplayer intent |
| ------------ | -------------------- | ------------------ |
| Lobby/Replay | Replay               | Return to lobby    |
| Menu         | Return to pregame    | Return to pregame  |
| Quit         | Quit to main menu    | Quit to main menu  |

The signal path is:

```text
MatchResultWindow
-> MatchResultsFlow
-> MatchEndFlow
-> GameplayComposition
-> GameplaySessionController
```

Session-level owners execute the consequences:

* Replay closes the current connection, resets gameplay, clears session context, and emits replay intent.
* Return to lobby sends a return-to-lobby request and resets gameplay presentation.
* Return to pregame closes the connection, resets gameplay, clears session context, and emits pregame route intent.
* Quit closes the connection, resets gameplay, clears session context and boot flow, and shows the main menu.

### 7. Game server reports durable results

After room game-over is detected, the game server also reports the resolved summary through the match-result reporting gate.

The reporting path is:

```text
room game-over lifecycle
-> resolved match summary
-> rooms.ReportResolvedMatchResultOnce
-> matchreporting.RuntimeReporter
-> player_data_record_match_result command per player
-> services/player-data runtime sink
```

The room marks the result as reported only after the reporter succeeds.

If reporting fails, the room keeps the resolved summary unreported so a later lifecycle path can retry. Reporting is also attempted before requested room leave and disconnected room cleanup so an already-resolved result is not lost during room exit.

### 8. Player-data commits stats

Each player result command includes:

```text
result_id: <match_id>:<game_player_id>
match_id
identity
context.play_mode
score
ship_deaths
won
```

Player-data validates `context.play_mode` and `identity.identity_kind` before calling a backing store.

Accepted writes update aggregate stats according to the selected route:

* Guest results update transient process-local stats.
* Local Profile results update local profile stats and local match-result rows when the embedded SQLite store is active.
* Authenticated Account results route through the Rails/API-backed account store when configured.

`result_id` is the idempotency key. Accepted duplicates are surfaced as duplicate results and are treated as successful by the game-server reporter.

## Result projections

The flow uses two different result projections.

### Presentation result projection

The presentation result projection is for the client result window.

It contains:

```text
match_id
mode
players[].game_player_id
players[].score
players[].ship_deaths
players[].won
```

It does not contain:

```text
account_id
local_profile_id
Rails user id
OAuth provider id
token
database row id
```

The client currently renders only player ID, ship deaths, and score.

### Durable result projection

The durable result projection is for player-data reporting.

It contains the trusted facts needed to update stats:

```text
result_id
match_id
play_mode
identity kind
account_id when authenticated
local_profile_id when local profile
score
ship_deaths
won
```

This projection is not sent to the result window.

## Inputs and outputs

Current inputs:

* server player lifecycle state
* remaining lives
* active ship presence
* player score
* player ship deaths
* room joinability
* room member account identity
* room member local profile identity
* room lifecycle state
* room snapshot `match_result`
* result-window button presses
* return-to-lobby packet request
* player-data result response

Current outputs:

* room state `GameOver`
* stored resolved match summary
* presentation-safe `room_snapshot.match_result`
* visible result rows
* match-over HUD hide/lock state
* match-over gameplay menu overlay state
* result-window route intent
* `player_data_record_match_result` commands
* player-data accepted/duplicate/rejected responses
* updated guest, local-profile, or account stats
* room return-to-lobby transition when requested

## Out of scope

This document does not define:

* future mode-specific match rules
* future score attack or objective result formats
* future mission, challenge, achievement, or progression summary rows
* result-window layout beyond current domain behavior
* packet schema source-of-truth details
* database schema details
* HTTP endpoint contract details
* leaderboard eligibility
* [Trust And Eligibility Policy](../platform/trust-and-eligibility-policy.md)
* reconnect behavior
* future `EndOfMatchFlow`, `MatchSummary`, or `MatchSummaryDispatcher` implementation details

Those belong in service, protocol, data, systems-design, limits, or planning documentation.

## Related docs

* [Player Experience](./!INDEX.md)
* [Platform](../platform/!INDEX.md)
* [Client](../../services/client/!INDEX.md)
* [Client Match End Flow](../../services/client/match-end-flow/!INDEX.md)
* [Gameplay Menu Flow](../../services/client/gameplay-menu-flow/!INDEX.md)
* [Gameplay Runtime](../../services/client/gameplay-runtime/!INDEX.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Game Server Rooms](../../services/game-server/rooms/!INDEX.md)
* [Game Server Integrations](../../services/game-server/integrations/!INDEX.md)
* [Game Server Simulation](../../services/game-server/simulation/!INDEX.md)
* [Player Data](../../services/player-data/!INDEX.md)
* [Trust And Eligibility Policy](../platform/trust-and-eligibility-policy.md)
* [API Server](../../services/api-server/!INDEX.md)
* [Data](../../data/!INDEX.md)
* [Protocol](../../protocol/!INDEX.md)
* [Match Outcomes And Results Planning](../../planning/domains/gameplay/match-outcomes-and-results.md)

## Notes

Local elimination and authoritative room match-over must remain separate. Showing final match results from local elimination would be incorrect.

The client result window is a presentation surface, not a persistence surface.

The current visible result table intentionally omits kills, rewards, progression, achievements, and durable identity.

The current implemented flow has a resolved match summary and player-data reporting path. The broader planned `EndOfMatchFlow`, `MatchSummary`, and `MatchSummaryDispatcher` architecture remains future planning rather than current implementation.
