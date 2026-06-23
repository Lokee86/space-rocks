# Matchmaking And Room Discovery

Parent index: [Platform Planning](./!INDEX.md)

## Purpose

This document plans the matchmaking and room-discovery domain for Space Rocks.

It defines how players find joinable rooms, enter matchmaking queues, receive room assignments, confirm assignments, and join through the normal room/session path.

This is a platform-domain planning document. It is not a room lifecycle document, not a game simulation document, not a mode-rules document, not a social graph document, and not an abuse-enforcement plan.

## Overview

Matchmaking and room discovery own the layer before room membership.

The planned flow is:

```text
player intent
-> room browser, queue, invite, or join-code entry path
-> searchable room candidate set
-> room fit decision
-> assignment target
-> confirmation
-> normal room join path
-> multiplayer session lifecycle
```

The API server owns matchmaking, room search, queue state, assignment, confirmation timeout, and fallback-room coordination.

The game server owns authoritative room instances, room lifecycle, room membership, final join validation, simulation, and game logic.

Matchmaking does not directly place a player into a room without the normal room/session path. It produces a join target. The client confirms the assignment, then joins through the game server’s normal room join flow.

## Current status

Active planning.

Current implementation already has:

```text
direct room creation
join by room code
room membership
room lifecycle states
ready state
start-game validation
room snapshots
authenticated-account multiplayer admission
game-server-owned room instances
```

Current implementation does not yet have:

```text
API-server-owned room browser
API-server-owned matchmaking queue
room discovery summaries
search/filter surface
assignment confirmation
assignment token or reservation seam
fallback room creation coordination
ratings-backed PvP matchmaking
Discord-backed social matchmaking facts
room title moderation integration
hosted room registry between API server and game server
```

## Ownership boundary

This document owns planning for:

```text
room browser and search
joinable room discovery
room search filters
RoomDiscoverySummary
matchmaking queue state
queue assignment policy
fallback room creation coordination
assignment target shape
assignment confirmation timeout
room fit decisions before join
join-code entry as a discovery surface
invite and party entry handoff into room assignment
hosted multiplayer discovery/registry seam
rating dependency for PvP matchmaking
```

This document does not own:

```text
room membership
room owner assignment
ready state
start-game validation
room cleanup
room lifecycle transitions
active game-player routing
WebSocket session lifecycle
game simulation
mode rules
scoring
objectives
match-end policy
leaderboard rating formulas
Discord relationship implementation
friend/block/mute source data
invite notification UX
room title moderation policy
abuse reports
admin enforcement actions
client/platform/build compatibility enforcement
exact HTTP endpoint shapes
exact packet schemas
deployment infrastructure
```

Multiplayer Session And Lifecycle owns room membership, ready state, room lifecycle, disconnect/reconnect, return-to-lobby, and cleanup.

Modes And Match Rules owns mode presets, room mode config validation, resolved match rules, scoring policy, objective policy, match-end policy, and result policy.

Social And Community Systems owns relationships, Discord integration, friends, blocks, mutes, parties, presence, invite creation, and invite notification.

Abuse And Enforcement Admin owns room title moderation, display-name moderation, reports, review, hiding, sanctions, appeals, and moderation tooling.

Leaderboards And Rankings owns Pilot Rating, Combat Rating, rating display surfaces, ranking boards, and rating lifecycle.

Deployment owns server processes, capacity, regions, health checks, scaling, routing infrastructure, and host lifecycle.

## Service ownership

The API server is the matchmaking authority.

It owns:

```text
room browser reads
search/filter execution
queue creation
queue cancellation
queue status
assignment decisions
assignment token creation
assignment confirmation timeout
fallback-room coordination
requester-safe discovery views
```

The game server is the room and game authority.

It owns:

```text
authoritative room instances
roomConfig validation accepted by room creation
room membership
room lifecycle
room capacity
room join validation
room ready state
game start
simulation
game result production
room cleanup
```

The planned service flow is:

```text
client requests browser/search/queue through API server
-> API server evaluates room discovery data
-> API server assigns an existing room or coordinates fallback room creation
-> API server returns a join target
-> client displays confirmation
-> user confirms before timeout
-> client joins through normal game-server room/session path
-> game server performs final join validation
```

API-server assignment does not replace game-server validation.

## Core discovery model

Matchmaking and discovery should use a requester-safe room discovery projection, not full room snapshots.

The model is:

```text
authoritative room instance
-> roomConfig
-> current join metadata
-> requester-safe RoomDiscoverySummary
-> browser/search/queue assignment
```

Room browser and matchmaking should not expose internal room state directly.

RoomDiscoverySummary should be mostly a searchable projection of `roomConfig` plus current join metadata.

## Entry surfaces

### Direct room creation

Direct room creation remains a primitive room entry path.

Discovery owns:

```text
room visibility/searchability choice
room_title or room_name
initial discovery metadata
roomConfig projection for browser/search
fallback-room creation expectations
```

Room lifecycle owns the actual created room, membership, owner, ready state, and start-game flow.

### Join by code

Join by code is a direct entry path, not matchmaking.

Join by code may bypass public listing, but it must not bypass:

```text
authentication
admission
room capacity
room lifecycle state
room visibility rules that apply to code join
account restrictions
game-server final join validation
```

Join-code requests should return safe failure reasons. They should not expose private moderation, enforcement, or social relationship details.

### Room browser

The room browser lists requester-visible joinable rooms by default.

Filters narrow the joinable-room list.

The browser should not list non-joinable rooms by default. Non-joinable room visibility belongs to later spectate, history, admin, diagnostics, or website surfaces if those are planned separately.

The browser candidate rule is:

```text
visible to this requester
+
joinable right now
```

### Matchmaking queue

The matchmaking queue exists to get players into playable rooms quickly.

Queue behavior should prefer short matchmaking times.

The default queue strategy is:

```text
look for an acceptable immediate room
-> if none exists, create or coordinate a fallback room as soon as practical
-> assign the player or party to that room
-> continue filling that room through browser/queue where appropriate
```

Casual, PvE, co-op, and unranked queues should use fast-fill behavior.

Ranked PvP matchmaking is different. It depends on ratings and may wait longer for fairer assignments.

### Invite and party entry

Room invites are social-owned.

The likely product shape is a local invite system that consumes or coordinates with broader Discord-backed social systems.

Social owns:

```text
invite creation
invite notification
party membership
presence
friends
blocks
mutes
Discord relationship integration
relationship privacy
```

Matchmaking owns:

```text
how an invite resolves to a room entry target
whether a target room can accept the invited player or party
whether a party fits the target room
how invite/party entry hands off to assignment and join
```

Matchmaking should consume normalized social facts. It should not directly depend on raw Discord API behavior.

## Room browser and filters

Most discovery “fields” are search filters.

Room browser filters may include:

```text
mode preset
rules summary
ranked or casual
rating band, where applicable
current player count
max players
party size support
region or server label
ping or latency bucket
room_title or room_name search
visibility available to requester
friends or invite eligibility later
```

Filters are not the same thing as room-owned state. They are ways to search requester-safe room summaries.

The browser should default to all joinable rooms visible to the requester, then narrow that set by filters.

## RoomDiscoverySummary

`RoomDiscoverySummary` is a requester-safe, searchable projection of roomConfig plus current join metadata.

Likely roomConfig-derived values:

```text
room_title or room_name
visibility/searchability
mode preset
rules summary
max players
ranked or casual
rating band, where applicable
party allowance
region or server label, if applicable
```

Likely runtime join metadata:

```text
current player count
joinable state
room state category
assignment or reservation availability
ping or latency bucket, if known
created or updated time bucket
```

RoomDiscoverySummary must not expose:

```text
account IDs
member IDs
session IDs
owner account ID
internal deployment node
hidden moderation state
private invite tokens
raw infrastructure data
debug-only room state
```

`display_name` should not be used for room presentation because display name is already an account/player identity concept.

Use `room_title` or `room_name`.

The exact final name remains open.

## Assignment and confirmation

Matchmaking assignments require user confirmation.

The planned flow is:

```text
client enters queue or selects a joinable room
-> API server returns a join target
-> client displays confirmation
-> user confirms within timeout
-> client joins through normal game-server room/session path
-> game server validates final join
```

Assignment should include a token or reservation-style seam so confirmation does not race room capacity.

Useful assignment target data:

```text
room_code or join_handle
assignment_token
expires_at
room_title or room_name
mode summary
player count
max players
ranked or casual
region or latency summary
reserved party size, if applicable
```

The API server owns assignment and confirmation timeout.

The game server owns final room capacity and join validation.

Open implementation decision:

```text
whether reservation state lives in the API server, game server, or both
```

The invariant is that assignment confirmation should not create obvious over-assignment or stale-join behavior.

## Queue state

Queue state should be explicit.

Useful queue states:

```text
not_queued
queued
searching
assigned
confirming
joining
cancelled
expired
failed
```

Queue status should expose safe player-facing information:

```text
status
selected mode or preset
ranked or casual
region preference, if applicable
elapsed time
confirmation expiry, if assigned
cancel allowed
```

Exact queue status shape belongs to API product surface and contract planning.

## Fallback room creation

Fallback room creation is required.

The queue should create fallback rooms as soon as practical when no acceptable joinable room exists.

Planned fallback flow:

```text
client enters queue
-> API server searches joinable room summaries
-> no acceptable immediate room exists
-> API server coordinates fallback room creation
-> game server creates authoritative room instance from validated roomConfig
-> API server receives join target
-> client confirms assignment
-> client joins game server normally
```

The API server decides fallback creation is needed.

The game server creates and owns the room instance.

The game server must validate the roomConfig it accepts.

## Ratings and PvP matchmaking

PvP matchmaking depends on ratings.

Ranked PvP matchmaking should not be implemented before rating support exists.

Rating-backed PvP matchmaking may need:

```text
Combat Rating
provisional rating
placement matches
party rating aggregation
rating-band widening over wait time
new-player protection
ranked eligibility checks
```

Casual room discovery and direct room entry do not require ratings first.

Planning split:

```text
casual / PvE / unranked
-> fast-fill strategy
-> broad filters
-> quick fallback room creation
-> shortest practical matchmaking time

ranked PvP
-> rating-constrained strategy
-> narrower eligibility
-> possible longer wait
-> no implementation before ratings exist
```

Pilot Rating and Combat Rating ownership belongs to Leaderboards And Rankings.

Matchmaking consumes rating facts once they exist.

## Ready behavior

All multiplayer modes require all current players ready before match start.

This is a universal room lifecycle rule.

Matchmaking does not own ready state and must not bypass ready requirements.

The flow remains:

```text
assignment / browser / invite / code join
-> normal room join
-> room membership
-> all players ready
-> start-game validation
-> match start
```

## Client/platform/build compatibility

Client, platform, and build compatibility should be screened before the player reaches online matchmaking.

Standard policy:

```text
out-of-date or incompatible client
-> cannot come online
-> cannot authenticate into online play
-> cannot reach matchmaking
```

This document should not treat build compatibility as a matchmaking filter except as an upstream assumption.

## Social and Discord handoffs

Social owns Discord-backed relationship facts.

Future relationship-aware matchmaking should consume facts such as:

```text
blocked relationship exists
muted relationship exists
friend relationship exists
party membership exists
invite eligibility exists
```

Blocks should always block communication where technically possible.

Blocks affect matchmaking co-placement only where practical.

Mutes are communication-focused and do not imply matchmaking exclusion by themselves.

Matchmaking consumes normalized social facts and returns safe failure reasons.

## Moderation and enforcement handoffs

Room title moderation belongs to abuse and enforcement planning.

`room_title` or `room_name` is a public text surface and should use the same moderation system as account display names.

Abuse And Enforcement Admin owns:

```text
banned terms
classifier checks
LLM or review queue
report handling
room hiding
sanctions
appeals
admin moderation tools
```

Matchmaking owns only how the moderated room title appears in discovery.

If a room is hidden, delisted, restricted, or enforcement-blocked, matchmaking consumes that state and removes or limits the room from browser/search/queue assignment.

## Hosted multiplayer and scaling prep

Most hosted multiplayer preparation is not needed for a single-server launch, but the domain should include the seam for scaling and portfolio strength.

Matchmaking may eventually consume:

```text
available regions
server health
server capacity
room registry
room advertisement state
room creation availability
latency bucket
deployment routing target
```

Deployment owns:

```text
server processes
regions
capacity reporting
health checks
scaling
host lifecycle
routing infrastructure
```

Matchmaking consumes deployment facts for discovery and assignment.

## API-server and game-server room registry

Because the API server owns matchmaking and the game server owns room instances, a room registry seam is required.

Possible implementation shapes:

```text
game server pushes room registry updates to API server
API server polls game-server room summaries
shared backing store or registry
single-server shortcut for early launch
```

The exact mechanism is open.

The invariant is:

```text
API server has enough requester-safe room facts to search, filter, queue, and assign.
Game server remains authoritative for actual room state and final joins.
```

## Failure categories

Failure categories should be explicit and safe.

Useful categories:

```text
room_not_found
room_full
room_not_joinable
room_private
invite_required
auth_required
account_restricted
mode_unavailable
region_unavailable
queue_expired
assignment_expired
matchmaking_unavailable
```

Avoid exposing precise private reasons such as hidden moderation state, enforcement details, or exact social relationship blockers.

Client/platform/build mismatch should normally be handled before online matchmaking and should not be treated as a normal matchmaking failure.

## Implementation sequence

1. Define the API-server-owned matchmaking/search/queue boundary.
2. Define the game-server-owned room registry summary boundary.
3. Define `RoomDiscoverySummary` as a requester-safe projection of roomConfig plus current join metadata.
4. Rename room presentation identity to `room_title` or `room_name`, not `display_name`.
5. Define browser behavior as requester-visible joinable rooms by default.
6. Define initial browser filters over joinable room summaries.
7. Define queue state and queue status.
8. Define assignment target shape.
9. Define assignment confirmation with timeout.
10. Define assignment token or reservation semantics.
11. Define fallback room creation coordination from API server to game server.
12. Ensure fallback rooms are created quickly when no acceptable room exists.
13. Keep all final room joins on the normal game-server room/session path.
14. Add room title moderation handoff to abuse/enforcement/admin.
15. Reserve social/Discord relationship seams for invites, parties, blocks, mutes, and friends-only discovery.
16. Define PvP matchmaking as rating-dependent.
17. Add ranked PvP matching only after ratings and ranked eligibility exist.
18. Add hosted multiplayer room registry, region, and capacity seams as scaling prep.
19. Keep client/platform/build compatibility outside matchmaking by screening before online access.

## Open decisions

```text
exact room presentation field name: room_title or room_name
exact RoomDiscoverySummary shape
exact room registry mechanism between API server and game server
exact fallback room creation command/API boundary
exact assignment token shape
exact reservation ownership
exact confirmation timeout duration
whether browser joins reserve slots or only queue assignments
exact initial browser filters
exact rating model for PvP matchmaking
exact party rating aggregation rule
exact queue widening policy for ranked PvP
exact Discord/local-social integration boundary for invites, blocks, mutes, and friends-only discovery
exact visibility flag names
```

These are not open policy questions:

```text
whether the API server owns matchmaking
whether the game server owns authoritative room instances
whether room browser lists joinable rooms by default
whether queue can create fallback rooms
whether matchmaking should minimize wait time
whether PvP matchmaking needs ratings first
whether all modes require all players ready
whether assignment requires confirmation
whether final join goes through normal room/session path
whether room title moderation belongs with display-name moderation
whether client/platform/build compatibility is screened before matchmaking
```

## Core invariants

```text
Matchmaking owns pre-room search, queue, assignment, and discovery.

Matchmaking does not own room lifecycle.

The API server owns matchmaking, room browser, queue, assignment, and fallback-room coordination.

The game server owns authoritative room instances, room lifecycle, room membership, final join validation, and simulation.

Room browser lists requester-visible joinable rooms by default.

Filters narrow the joinable room set.

Most discovery fields are search filters.

RoomDiscoverySummary is a requester-safe projection of roomConfig plus current join metadata.

RoomDiscoverySummary is not a full room snapshot.

Use room_title or room_name, not display_name.

Room title moderation belongs to abuse/enforcement/admin alongside display-name moderation.

Room invites are social-owned.

Discord/social integration owns relationships, blocks, mutes, friends, parties, and invite facts.

Matchmaking consumes normalized social facts when they exist.

All modes require all current players ready before match start.

Matchmaking never bypasses ready state, admission, room capacity, or final game-server join validation.

Client/platform/build compatibility is screened before online matchmaking.

Queue creates fallback rooms quickly when no acceptable room exists.

PvP matchmaking requires ratings before implementation.

Ranked PvP matchmaking is separate from casual room discovery and fast-fill matchmaking.

Assignment requires user confirmation with timeout.

Assignment should use token or reservation semantics to avoid stale or overfilled joins.

Hosted multiplayer and scaling seams are included, but deployment owns infrastructure.
```

## Related docs

* [Platform Planning](./!INDEX.md)
* [Account And Identity Systems](account-and-identity-systems.md)
* [Multiplayer Session And Lifecycle](multiplayer-session-and-lifecycle.md)
* [Social And Community Systems](social-and-community-systems.md)
* [Leaderboards And Rankings](leaderboards-and-rankings.md)
* [Game Integrity Policy](security-and-admin/game-integrity-policy.md)
* [Abuse And Enforcement Admin](security-and-admin/abuse-and-enforcement-admin.md)
* [Modes And Match Rules](../gameplay/modes-and-match-rules.md)
* [API Product Surface](../../protocol/api-product-surface.md)
* [Realtime Protocol Architecture](../../protocol/realtime-protocol-architecture.md)
* [Build Release And Environment Matrix](../technical/build-release-and-environment-matrix.md)

## Notes

This document plans the full matchmaking and room-discovery domain system.

The first useful implementation can be smaller, but the ownership boundaries should still follow this plan.

The most important architectural decision is that matchmaking belongs to the API server while authoritative room instances remain game-server-owned.

The most important product decision is that matchmaking should get players into playable rooms quickly. Casual queues should create fallback rooms quickly rather than waiting for ideal matches. Ranked PvP is the exception and should wait for rating support before implementation.
