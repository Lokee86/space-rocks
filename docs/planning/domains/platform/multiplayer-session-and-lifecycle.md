# Multiplayer Session And Lifecycle

Parent index: [Platform Planning](./!INDEX.md)

## Purpose

This document plans the multiplayer session and lifecycle domain for Space Rocks.

It defines how online players connect, authenticate, create or join rooms, wait in lobbies, start matches, load into synchronized play, join mid-session, disconnect, reconnect, return after match end, and leave or get removed from rooms.

This is a platform-domain planning document. It is not a matchmaking document, not a game simulation document, not a mode-rules document, not a realtime protocol document, and not an abuse-enforcement plan.

## Overview

Multiplayer lifecycle owns the layer from WebSocket session through room participation and match participation.

The planned lifecycle is:

```text
websocket connection
-> authenticated online session
-> room join or room create
-> room membership
-> lobby / queued waiting state
-> countdown
-> Starting synchronized handoff
-> InGame match participation
-> disconnect / reconnect / mid-session join handling
-> GameOver result and next-action state
-> individual return, leave, timeout, kick, ban, or cleanup
```

The game server owns authoritative room instances, room membership, room lifecycle, active game ownership, final join validation, reconnect handling, and room cleanup.

The API server owns matchmaking, room browser, queue, assignment, confirmation, and fallback-room coordination before the player enters the normal game-server room/session path.

All final room joins go through the game server’s normal room/session lifecycle.

## Current status

Active planning.

Current implementation already has:

```text
WebSocket session lifecycle
authenticated-account multiplayer admission
room create / join / leave
ready state
owner-gated start
Lobby -> Starting -> InGame -> GameOver
room snapshots
match result summary in room snapshots
return-to-lobby path
room cleanup after empty rooms
```

Current implementation does not yet fully have:

```text
active in-game reconnect
member-disconnected room state
paused active player reconnect handling
mid-session join
queued join during Starting
spectator capacity
spectator packet lane
member-local return-to-lobby
kick / ban split
room-lifetime bans
join order tracking
Starting as a real synchronized loading handoff
multiplayer no-action timeout outside queued/waiting states
```

## Ownership boundary

This document owns planning for:

```text
WebSocket session lifecycle
session-to-room handoff
online room admission execution
room membership lifecycle
ready state lifecycle
countdown lifecycle
Starting synchronized handoff
join loading lifecycle
mid-session join structure
disconnect and active reconnect
member-local return-to-lobby
room owner transfer
kick and room-lifetime ban behavior
room cleanup triggers
room lifecycle state transitions
room capacity and reservation lifecycle
spectator admission lifecycle
multiplayer no-action timeout
lifecycle diagnostics and event facts
```

This document does not own:

```text
matchmaking queue assignment
room browser search and filters
room discovery summaries
OAuth/account implementation
token model details
packet encoding
realtime packet lane design
spectator packet payloads
mode rule definitions
scoring policy
objective policy
match-end result policy
gameplay simulation
pause-seam mechanics
player-data persistence
leaderboard eligibility formulas
abuse sanctions and appeals
UI layout
exact API endpoint shapes
exact packet schemas
deployment infrastructure
```

Matchmaking and room discovery owns browser, queue, assignment, confirmation, fallback-room coordination, and room discovery metadata.

Modes and match rules owns start policy, join policy, mid-session join policy, spectator allowance, result participation policy, and mode-specific lifecycle behavior.

Realtime protocol architecture owns packet lanes, packet payloads, snapshots, deltas, and encoding.

Game simulation owns active ships, pause mechanics, deaths, respawn, scoring, pickups, collision, and authoritative match state.

Match outcomes owns final result locking and result handoff.

Game integrity owns classification of disconnect/reconnect churn, automation lanes, ranked eligibility, and suspicious lifecycle behavior.

## Service ownership

The game server is the room lifecycle authority.

It owns:

```text
authoritative room instances
room membership
room lifecycle state
room capacity
queued room join reservations
final join validation
ready state
countdown
Starting handoff
active reconnect
mid-session join execution
kick and ban execution
room cleanup
game instance ownership
```

The API server is the matchmaking and discovery authority.

It owns:

```text
room browser reads
search/filter execution
matchmaking queue
queue assignment
assignment confirmation
fallback-room coordination
requester-safe discovery views
```

The planned service flow is:

```text
client authenticates online
-> client requests browser / queue / invite / code path
-> API server may assign or reserve a join target
-> client joins through game-server room/session path
-> game server performs final validation
-> game server creates or updates room membership
```

API-server assignment does not replace game-server validation.

## Identity and identifiers

Multiplayer lifecycle uses separate identifiers for separate jobs.

```text
SessionID
-> WebSocket connection identity
-> server-internal
-> new connection gets a new session identity

AccountID
-> authenticated account identity
-> production online multiplayer identity

MemberID
-> internal room-membership identity
-> reconnect seam
-> not exposed in normal room snapshots

PlayerID
-> readable player-facing label
-> active simulation routing label
-> allocated by the existing room/player flow

currentGamePlayerID
-> networking-owned active-game routing field
-> not durable account or room identity
```

`PlayerID` is the room-visible player label and the active simulation routing label.

A separate `RoomSlotID` is not required while room slot identity and `PlayerID` remain equivalent.

`MemberID` remains the internal reconnect spine. Normal clients should not receive `MemberID` in room snapshots.

Only active players count as match participants.

Queued joiners, pending joiners, loading joiners, disconnected-but-not-active observers, result viewers, and spectators do not count as match participants unless the selected mode explicitly says otherwise.

## Connection and authentication lifecycle

A WebSocket connection is session-only.

Connection does not imply room membership.

Authentication does not imply room membership.

The normal online flow is:

```text
websocket upgrade
-> session created
-> session starts unauthenticated or guest-shaped
-> authenticate_request
-> API auth verification
-> authenticate_result
-> authenticated account session
-> room create / join / reconnect / queue handoff
```

Production online multiplayer requires authenticated account identity.

The game server must verify account identity through the API auth boundary. It must not read Rails auth tables directly.

Bearer tokens and OAuth provider identities must not become gameplay identity.

## Room membership lifecycle

Room membership begins when the game server accepts a create, join, late-join, GameOver join, or reconnect claim.

Room membership is separate from active match participation.

Room membership may exist while a player is:

```text
in lobby
ready
unready
in countdown
loading through Starting
active in game
disconnected and reconnectable
queued for mid-session join
spectating
viewing GameOver results
returned to post-match lobby state
leaving
kicked
banned
expired
```

`MemberID` should remain stable while the member remains in the room.

`PlayerID` should remain stable while the member remains in the room.

A member leaves the room when they intentionally leave, are kicked, are banned, expire from a timeout, or the room is cleaned up.

## Join order

V2 should track join order.

Recommended shape:

```text
join_sequence
-> room-local monotonic integer
-> assigned when room membership is created
-> never reused inside the room
```

Join order is useful for:

```text
owner transfer
auditing
debugging
stable room history
diagnostics
snapshot ordering where needed
```

Default owner transfer should use the connected, non-banned remaining member with the earliest join order.

If implementation temporarily uses lowest `PlayerID`, that is an implementation shortcut, not the intended ownership model.

## Room states

The core room states are:

```text
Lobby
Starting
InGame
GameOver
Closed
```

### Lobby

Lobby is a queued/waiting room state.

Lobby members may wait, ready, unready, select allowed pre-match options, or leave.

Lobby presence does not trigger multiplayer no-action timeout by default.

A player sitting unready in Lobby is a room-management issue, not an automatic timeout issue.

Lobby members may still be removed by:

```text
manual leave
owner kick
room-lifetime ban
disconnect / reconnect expiry where applicable
room cleanup if the room closes
mode-specific lifecycle policy
```

### Starting

`Starting` is a real synchronized handoff state.

It owns:

```text
final locked start segment
match creation
participant loading
loading confirmation
automatic loading confirmation timeout
final synchronization
activation into active play
```

The final visible second of the countdown happens during `Starting`.

Architecturally, `Starting` is the lock boundary.

### InGame

`InGame` is active match state.

It owns:

```text
active play
active reconnect
mid-session join
spectator admission where allowed
mode/match runtime participation rules
```

`InGame` is the only normal lifecycle state for mid-session join.

### GameOver

`GameOver` is the post-match result and next-action state.

GameOver can be joined for result viewing and return-to-lobby / next-game flow.

Return-to-lobby is member-local in V2. One player returning does not force every other player back to lobby.

GameOver players who do nothing may be removed by no-action timeout.

### Closed

Closed rooms do not accept normal joins, reconnects, or queued join activation.

Closed rooms exist only for cleanup, diagnostics, or future admin/debug behavior.

## Ready, countdown, and start lifecycle

The normal start flow is:

```text
Lobby
-> mode/match start policy passes
-> owner or allowed start actor requests start
-> visible countdown begins
-> unlocked countdown may be cancelled
-> Starting begins for final locked segment
-> match is created and synchronized
-> active players are activated
-> InGame
```

Any player disconnecting or unreadying during the unlocked countdown cancels the countdown.

The countdown represents the countdown to match start.

The final visible second of the countdown happens during `Starting`, not before it.

Start policy is decided by match/mode rules.

Lifecycle owns the countdown and state transition machinery. Mode/match rules decide whether starting is allowed.

## Starting and loading

Starting has two timing concepts:

```text
countdown to Starting
-> pre-lock countdown from Lobby

Starting timer
-> synchronized loading and final timer to active match start
```

Loading confirmation timeout is automatic confirmation.

If loading confirmation times out, the default behavior is to treat the participant as confirmed unless activation fails.

Casual default behavior for failed loading is to remove the failed participant and continue.

Mode/match rules may define stricter behavior, including cancelling back to Lobby, removing failed participants, or marking failed participants disconnected.

Starting should expose enough lifecycle state for the client to show a loading screen or dialog.

## Join and loading lifecycle

Joining should expose enough lifecycle state for the client to show a joining/loading dialog.

Useful conceptual states:

```text
join_requested
join_validating
join_queued
join_loading
join_waiting_for_activation
join_failed
joined_lobby
joined_ingame
joined_gameover
```

The exact packet names belong to realtime protocol planning.

A joining player must load before becoming visible or active in the match.

If the player cancels or disconnects before activation, the pending join or reservation should be released unless the member has already entered a reconnectable state.

## Queued joins and capacity reservations

Queued join requests reserve capacity regardless of source.

This includes:

```text
matchmaking assignment joins
direct code joins that are queued
Starting joins queued until InGame
spectator joins where spectator capacity exists
```

Queued joins during `Starting` reserve capacity and wait for `InGame` admission resolution.

A queued join should expire or release its reservation if the requester disconnects, cancels, fails admission, or exceeds the join queue timeout.

## Mid-session join

V2 must structurally support mid-session join.

Rules:

```text
InGame is the only normal lifecycle state for mid-session join.
Join requests during Starting are queued for when the game enters InGame.
Joiners must load before becoming visible or active.
Mid-session join consumes room capacity the same way lobby join does.
Only active players count as match participants.
```

Joiners enter according to the selected mode.

If a mode supports spectator entry, the joiner may enter as a spectator.

Otherwise, the joiner enters according to match/mode rules.

By default, a late joiner enters in the normal match-start state unless match/mode rules override that.

Mode/match rules decide:

```text
whether mid-session join is allowed
whether the joiner enters active play or spectator state
spawn/default state
late-join cutoff rules
result eligibility
ranked or competitive restrictions
```

## Spectators

Spectators are structurally supported by V2.

Spectators have separate capacity from active players.

Spectators should use a separate packet lane from active players.

Spectators likely need less game-state than active players, but the exact payload belongs to realtime protocol implementation.

Lifecycle owns:

```text
spectator admission state
spectator capacity counting
spectator join and leave lifecycle
```

Realtime protocol owns:

```text
spectator packet lane
spectator state payload
spectator delivery cadence
```

Mode/match rules own:

```text
whether spectators are allowed
whether spectators can join mid-session
what spectators can see
whether any spectator state affects results
```

Spectators are not match participants unless the selected mode explicitly treats them as participants.

## Disconnect and active reconnect

Active in-game reconnect is required in V2.

Disconnect does not mean leave.

The planned reconnect flow is:

```text
active player disconnects
-> room member is marked disconnected
-> active game player is paused through the existing pause seam
-> reconnect claim is accepted for the same member
-> new session attaches to the existing member
-> active ship control is restored
```

Reconnect should be built around `MemberID`, not exposed through normal room snapshots.

The exact paused-player mechanics belong to the pause/gameplay seam. This document owns the lifecycle relationship, not pause behavior.

Reconnect timeout is a multiplayer no-action timeout because a disconnected active player is holding room and match resources.

Mode/match rules may define the final result of reconnect expiry.

## GameOver and return-to-lobby

Return-to-lobby is an individual decision.

Any player can trigger return-to-lobby for themselves.

Only the player who clicks the button returns to lobby or next-game waiting state.

Other players remain in GameOver until they take action, leave, or time out and are automatically removed.

GameOver join is allowed for result viewing and return-to-lobby / next-game flow where capacity and mode policy allow it.

A GameOver joiner is not a match participant for the completed match.

V2 should avoid treating return-to-lobby as a whole-room reset triggered by one player.

The room may remain in `GameOver` while individual members enter a post-match lobby-ready state.

A later whole-room transition can occur when remaining members have returned, left, timed out, or when lifecycle policy says it is safe.

## No-action timeout

V2 should add multiplayer no-action timeout, but not for lobby or queue waiting.

Excluded from no-action timeout:

```text
matchmaking queue
room lobby
queued join request
queued Starting join reservation
normal ready / waiting lobby state
```

Included in no-action timeout:

```text
Starting / loading state where activation resources are held
pending mid-session join activation
disconnected active-player reconnect state
GameOver result-viewing / next-action state
optional spectator idle state if spectator capacity becomes scarce
```

Recommended timeout categories:

```text
starting_loading_timeout
mid_session_join_activation_timeout
in_game_disconnected_reconnect_timeout
game_over_no_action_timeout
optional spectator_idle_timeout
```

Loading confirmation timeout is separate from no-action timeout.

Loading confirmation timeout is automatic confirmation. No-action timeout is removal, expiry, or lifecycle cleanup.

## Owner behavior

Room ownership transfers immediately when the owner drops or disconnects.

Reserved ownership can be scaffolded for a future policy, but it is not V2 behavior.

Default owner transfer should use join order.

Room owners should have tools to remove players, especially for start-blocking reasons.

Owner authority is room lifecycle authority, not abuse enforcement authority.

## Kick and ban

Player removal is split into kick and ban.

```text
kick
-> simple room removal
-> player can rejoin

ban
-> room-lifetime permanent removal
-> player cannot rejoin that room
```

Kick is a room-management action.

Ban is a room-lifetime room-management restriction.

Production online room bans should target `AccountID`.

The room may also keep `MemberID`, `PlayerID`, display name, and join order context for audit and diagnostics, but the durable room-lifetime block should be account-based.

Likely default scope:

```text
Lobby
-> owner can kick or ban

Unlocked countdown
-> owner can kick or ban
-> removal cancels countdown

Starting
-> failed participants are removed by lifecycle or mode policy
-> owner kick/ban is not default behavior

InGame
-> owner kick/ban of active players is not default behavior

GameOver
-> owner kick/ban may be allowed for post-match room management
```

Mode-specific or admin-specific removal behavior belongs to the relevant mode, integrity, or admin systems.

## Capacity model

V2 uses separate capacity concepts.

```text
player_capacity
spectator_capacity
queued_player_reservations
queued_spectator_reservations
```

Lobby members, active players, pending active-player joins, queued active-player joins, and GameOver members waiting for next game count against player capacity.

Spectators count against spectator capacity.

Queued requests reserve the relevant capacity while queued.

Capacity reservations must be released on cancellation, disconnect before activation, failed admission, timeout, kick, ban, or room cleanup.

## Match participants and results

Only active players count as match participants.

Not match participants by default:

```text
queued joiners
pending joiners
loading joiners
spectators
GameOver result viewers
members waiting for next game
```

All match participants affect win/loss/results unless match/mode rules say otherwise.

Ranked and competitive modes may define stricter result participation and late-join rules.

Match result finalization belongs to match outcomes.

Lifecycle must preserve these result invariants:

```text
final results lock once
disconnect does not rebuild results
reconnect does not rebuild results
GameOver joins do not alter completed match results
member-local return-to-lobby does not alter completed match results
cleanup must not discard unreported resolved results
```

## Mode and match rule relationship

Lifecycle provides state transition machinery.

Mode/match rules decide whether transitions are allowed and how mode-specific cases behave.

Lifecycle should provide facts such as:

```text
room_state
member_count
connected_member_count
player_capacity
spectator_capacity
queued_reservations
ready_state
countdown_state
starting_state
disconnect_state
match_elapsed_time
active_player_count
spectator_count
result_locked
```

Mode/match rules decide:

```text
start allowed
join allowed
mid-session join allowed
spectator allowed
late-join initial state
failed loading behavior
disconnect expiry behavior
result participation behavior
ranked or competitive restrictions
return / next-game policy where mode-specific
```

Lifecycle owns execution. Mode/match rules own policy.

## Matchmaking handoff

Matchmaking and room discovery own the pre-room path.

The lifecycle entry point is the normal game-server join path:

```text
client receives or selects join target
-> client joins through game-server session path
-> game server validates account/session/room/capacity/reservation
-> room membership or queued reservation is created
-> lifecycle state is sent to client
```

Matchmaking assignments may reserve capacity, but game-server final validation remains authoritative.

## Diagnostics and lifecycle events

V2 should expose useful lifecycle facts for debugging, audits, integrity review, and future support tooling.

Useful event concepts:

```text
session_connected
session_authenticated
room_join_requested
room_join_queued
room_join_accepted
room_join_rejected
member_joined
member_left
member_kicked
member_banned
member_disconnected
member_reconnect_started
member_reconnected
member_reconnect_expired
owner_transferred
countdown_started
countdown_cancelled
starting_entered
loading_confirmation_received
loading_confirmation_auto_confirmed
loading_failed_participant_removed
match_started
mid_session_join_requested
mid_session_join_accepted
mid_session_join_rejected
spectator_joined
spectator_left
active_player_paused_for_disconnect
active_player_restored_after_reconnect
return_to_lobby_requested
member_returned_to_lobby
game_over_no_action_timeout
room_cleanup_scheduled
room_closed
```

Not every event needs durable event infrastructure in the first implementation. Some can begin as room or network logs.

## Implementation sequence

1. Update related links to point at the canonical platform lifecycle plan.
2. Clarify identifier roles around `SessionID`, `AccountID`, `MemberID`, `PlayerID`, and active match participants.
3. Add join order tracking to room membership.
4. Add disconnected member state without treating disconnect as leave.
5. Route active disconnect through the pause seam.
6. Add reconnect claim handling that restores active ship control.
7. Make `Starting` a real synchronized handoff state.
8. Add countdown-to-Starting and Starting-to-InGame timing.
9. Add loading confirmation and automatic confirmation timeout.
10. Add casual failed-loading removal-and-continue behavior.
11. Add queued join reservations, including queued joins during `Starting`.
12. Add mid-session join structure for `InGame`.
13. Add spectator capacity and spectator lifecycle state.
14. Add member-local return-to-lobby behavior.
15. Add GameOver join/result-viewing behavior.
16. Add multiplayer no-action timeout for non-queue, non-lobby lifecycle states.
17. Split owner removal into kick and room-lifetime ban.
18. Add immediate owner transfer on disconnect/drop using join order.
19. Add lifecycle diagnostics/log events.
20. Preserve result finalization and reporting during disconnect, reconnect, return, and cleanup.

## Open decisions

Implementation-shape decisions remain:

```text
exact timeout durations
exact reconnect expiry result per mode
exact kick/ban allowed states
exact ban identity fallback for dev/no-auth multiplayer
exact GameOver member/result-viewer capacity treatment
exact queued join expiry duration
exact spectator capacity limits
exact spectator packet payload
exact lifecycle failure codes
exact room snapshot fields for Starting/loading/queued/reconnect states
exact implementation path for member-local return-to-lobby over current whole-room reset
```

These are not open policy questions:

```text
whether V2 is the completed domain plan
whether active reconnect belongs in V2
whether disconnect is distinct from leave
whether active players pause through the pause seam on disconnect
whether mid-session join is structurally supported
whether Starting is a real synchronized handoff state
whether queued joins reserve capacity
whether lobby is excluded from no-action timeout
whether return-to-lobby is member-local
whether kick and ban are separate
whether bans are room-lifetime permanent
whether join order should be tracked
whether spectators have separate capacity
whether only active players count as match participants
whether start and join policy are mode/match-rule decisions
```

## Core invariants

```text
Connection does not imply room membership.

Authentication does not imply room membership.

Room membership does not imply active match participation.

SessionID is connection-scoped.

AccountID is authenticated account identity.

MemberID is internal room-membership identity and reconnect seam.

MemberID is not exposed in normal room snapshots.

PlayerID is the readable room/player label and active simulation routing label.

Only active players count as match participants by default.

The game server owns authoritative room lifecycle and final join validation.

The API server owns matchmaking, discovery, assignment, and confirmation before room/session handoff.

All final joins go through the normal game-server room/session path.

Queued join requests reserve capacity regardless of source.

Lobby is a queued/waiting state and does not trigger no-action timeout by default.

Starting is the locked synchronized handoff state.

The final visible countdown segment happens during Starting.

Loading confirmation timeout is automatic confirmation.

Casual failed-loading default removes the failed participant and continues.

InGame is the normal mid-session join state.

Join requests during Starting queue for InGame.

Joiners must load before becoming visible or active.

Spectators have separate capacity and a separate packet lane.

Disconnect is not leave.

Active in-game reconnect is required.

Disconnected active players are paused through the pause seam.

Reconnect restores active ship control.

Room ownership transfers immediately when the owner drops or disconnects.

Kick and ban are separate.

Kick allows rejoin.

Ban is room-lifetime permanent.

Return-to-lobby is member-local.

One player returning to lobby does not force all GameOver members back to lobby.

GameOver can be joined for result viewing and next-game flow.

Lifecycle must not mutate locked match results during disconnect, reconnect, return-to-lobby, or cleanup.

Mode/match rules decide start, join, mid-session join, spectator, result participation, and mode-specific lifecycle policy.

Lifecycle owns execution machinery, not mode policy.
```

## Related docs

* [Platform Planning](./!INDEX.md)
* [Account And Identity Systems](account-and-identity-systems.md)
* [Matchmaking And Room Discovery](matchmaking-and-room-discovery.md)
* [Social And Community Systems](social-and-community-systems.md)
* [Leaderboards And Rankings](leaderboards-and-rankings.md)
* [Game Integrity Policy](security-and-admin/game-integrity-policy.md)
* [Modes And Match Rules](../gameplay/modes-and-match-rules.md)
* [Match Outcomes And Results](../gameplay/match-outcomes-and-results.md)
* [Player Experience Systems](../gameplay/player-experience-systems.md)
* [Realtime Protocol Architecture](../../protocol/realtime-protocol-architecture.md)

## Notes

This document replaces the earlier stub-level multiplayer lifecycle plan.

The most important V2 behavior is active in-game reconnect. Lobby and GameOver reconnect may exist, but they are not the central reconnect goal.

The most important lifecycle split is that disconnect, leave, kick, ban, timeout, and cleanup are different lifecycle outcomes.

The most important cross-domain boundary is that lifecycle owns state transitions and execution, while mode/match rules decide whether a transition is allowed for the selected mode.
