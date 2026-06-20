# Local Pilot Profile Flow

Parent index: [Player Experience](./!INDEX.md)

## Purpose

This document describes the current cross-system player-experience flow for local pilot selection, Guest fallback, local profile management, profile readout, and single-player profile handoff.

It is a domain document. It explains how the player-facing flow moves across client presentation, game-server hosting, room membership, match reporting, and player-data persistence. Implementation details belong in the related service, protocol, and data docs.

## Overview

The local pilot profile flow lets a player choose the identity context used for local single-player.

Current identity choices for this flow are:

```text
Guest
= local single-player fallback identity
= no durable local_profile_id
= transient player-data route

Local Profile
= durable local-only identity
= keyed by local_profile_id
= backed by player-data local storage when available
```

The flow starts in the single-player pregame menu. The client applies the saved local default, displays the active callsign, lets the player open the local pilot selector, and lets the player load Guest or a saved local profile. Local profiles can also be created, renamed, or deleted from the selector flow.

The client owns player-facing selection state and presentation. It does not persist local profiles directly. Local profile persistence is owned by the player-data service, currently hosted through game-server HTTP routes on the local data-handler surface.

When single-player gameplay starts, the client sends only the selected `local_profile_id` when the active identity is Local Profile. Guest starts without a `local_profile_id`. The game server attaches the local profile ID to the single-player room member when supplied, keeps gameplay player IDs separate from profile IDs, and later reports authoritative match results through player-data. Player-data routes those match results to Guest transient memory or Local Profile storage based on identity.

## Participating systems

* Godot client - owns the pregame local pilot selector, active single-player profile context, callsign display, profile readout presentation, and single-player boot handoff.
* Client HTTP API layer - calls the game-server-hosted player-data HTTP routes for profile reads and local profile management.
* Game-server HTTP process - hosts player-data HTTP routes on the local data-handler surface.
* Game-server WebSocket/session flow - receives the single-player start request and passes the selected `local_profile_id` into room membership when present.
* Rooms - preserve member-level account/local-profile references separately from gameplay player IDs.
* Gameplay - owns live match state, score, deaths, and match facts, but not profile persistence.
* Match reporting - converts authoritative match summaries into player-data record commands.
* services/player-data - owns local profile CRUD/default behavior, profile stat reads, identity-based store routing, Guest transient stats, and local-profile stat writes.
* Embedded SQLite local store - backs Local Profile data in the standard no-tag development build.
* Rails API server - owns Authenticated Account identity and online account persistence, but does not own Local Profile storage.

## Authority boundaries

### Client authority

The client may:

* select Guest or Local Profile for local single-player
* remember the active single-player profile context during the client session
* display the active callsign in the pregame menu
* open local pilot selector and subpanel UI
* validate callsign input before submitting create or rename requests
* call local profile HTTP endpoints through the client API layer
* include `local_profile_id` in the single-player start request when Local Profile is active
* display profile stats returned by player-data

The client must not:

* write SQLite directly
* choose Guest, Local Profile, or Authenticated Account backing stores directly
* mutate profile stats from presentation code
* treat display name as durable identity
* use `local_profile_id` as a gameplay player ID
* route profile readout directly to Rails stats endpoints

### Player-data authority

Player-data owns:

* local profile listing
* local profile creation
* local profile display-name updates
* local profile deletion
* default Guest/local-profile selection
* server-generated `local_profile_id` values
* local profile stat storage and stat reads
* Guest transient stats
* Guest-to-local-profile stat seeding during profile creation
* identity-based store routing
* local profile unavailability behavior

The local profile ID is the durable local identity key. Display name is mutable callsign/presentation data.

### Game-server authority

The game server owns:

* hosting the player-data HTTP handlers in the current in-process deployment shape
* receiving the single-player WebSocket start request
* creating the single-player room
* attaching `local_profile_id` to the room member when the client supplies one
* preserving account/local-profile identity separately from session ID, room member ID, and gameplay player ID
* reporting authoritative match results to player-data

The game server does not own local profile persistence, local SQLite schema, or profile stat mutation.

### Room and gameplay authority

Rooms own room membership and member identity attachments.

Gameplay owns live player state and match facts. It does not own durable player identity or player-data persistence.

Gameplay player identity remains separate from:

```text
local_profile_id
account_id
session_id
room member identity
display name
callsign
```

### API-server authority

Rails/API owns Authenticated Account identity and online account-backed persistence.

The Local Profile flow is local-only. It is not an API-server account cache, not an online account, and not a synced Authenticated Account.

## Flow summary

### Enter single-player pregame

```text
Main Menu
-> Single Player
-> Pregame Menu in single-player mode
-> apply saved Guest/local-profile default
-> update callsign indicator
```

When the single-player pregame menu opens, the client asks the local profile API for the saved default selection.

If the saved default is Guest, missing, invalid, unavailable, or cannot be loaded, the client applies Guest as the active single-player profile context.

If the saved default is Local Profile and includes a valid `local_profile_id` and display name, the client selects that local profile context and updates the visible callsign.

### Open local pilot selector

```text
Pregame Menu
-> Select Pilot
-> local pilot selector mounted in the transmission panel
-> local profiles listed from player-data
-> Guest row appended as fallback
```

The selector lists durable local profiles returned by player-data and always includes Guest as a selectable fallback row.

Guest is not a local profile record. It has no `local_profile_id`, cannot be renamed, and cannot be deleted.

### Load Guest

```text
select Guest
-> persist default identity_kind = guest
-> set active single-player context to Guest
-> update callsign to Guest
```

Loading Guest stores Guest as the local default and clears the active local profile ID from the single-player context.

### Load Local Profile

```text
select local profile
-> persist default identity_kind = local_profile + local_profile_id
-> set active single-player context to selected local profile
-> update callsign to selected display name
```

Loading a local profile stores the selected `local_profile_id` as the local default and uses the profile display name as the player-facing callsign.

### Create Local Profile

```text
Create
-> callsign entry subpanel
-> validate callsign
-> POST local profile create request
-> player-data generates local_profile_id
-> optional Guest stat seed
-> refresh selector
```

Create opens a subpanel callsign entry flow.

The accepted callsign shape is:

```text
^[A-Za-z0-9_-]+$
```

Blank callsigns are rejected before submission. Callsigns outside the accepted pattern are rejected before submission and are also rejected by the player-data HTTP handler.

When the active single-player identity is Guest, the create request sends `seed_from_guest_stats = true`. Player-data can copy current transient Guest stats into the newly created local profile.

When the active single-player identity is already Local Profile, the create request sends `seed_from_guest_stats = false`, and the new profile starts with zero stats.

Creating a profile does not make the client write local storage directly. Player-data creates the profile, generates the `local_profile_id`, and returns the created profile summary.

### Edit Local Profile

```text
Edit selected local profile
-> callsign entry subpanel
-> validate new callsign
-> update display_name only
-> refresh selector
-> update active callsign if edited profile is active
```

Edit is available only for Local Profile rows.

Renaming changes display name only. It does not change `local_profile_id`, stats, or match-result history.

If the edited profile is the active single-player context, the client updates the active callsign after the rename succeeds.

### Delete Local Profile

```text
Delete selected local profile
-> confirmation subpanel
-> DELETE local profile request
-> player-data deletes profile-owned local data
-> default resets to Guest if needed
-> client applies Guest if deleted profile was active
-> refresh selector
```

Delete is available only for Local Profile rows.

The delete action is sent only after confirmation. Player-data owns deletion of the local profile record and associated local profile data. If the deleted profile was the stored default, the local store resets the default to Guest.

If the client deleted the currently active local profile, the client applies Guest as the active single-player context.

### Profile readout

```text
Profile button
-> active profile context resolved
-> POST /api/player-data/profile
-> player-data loads stats by identity
-> client displays callsign and normalized stats
```

Profile readout uses the active pregame context.

For Guest and Local Profile, the client sends the selected play mode and identity context without an authenticated-account bearer token. For Local Profile, the request includes `local_profile_id`.

Player-data returns normalized stats. The client shapes those stats for display and uses the selected local callsign context for local-profile presentation.

Profile readout does not mutate stats.

### Start single-player gameplay

```text
Play Endless
-> client reads active single-player context
-> if Local Profile, send local_profile_id in start_single_player_request
-> if Guest, send no local_profile_id
-> game server creates single-player room
-> room member stores local_profile_id when supplied
-> gameplay starts
```

The single-player boot handoff carries only the local profile ID needed by the game server to preserve the local-profile identity reference.

The game server does not need the local profile display name to start gameplay.

### Match result persistence handoff

```text
gameplay match facts
-> room match summary
-> local_profile_id copied from room member when present
-> match reporting builds player-data command
-> player-data routes write by identity kind
```

When the match ends, gameplay and room code provide authoritative score, death, win, and match facts.

If the room member has a `local_profile_id`, match reporting sends a Local Profile player-data record command. If no local profile ID is present, the result routes as Guest.

Player-data owns the resulting stat update.

## Inputs and outputs

Current inputs:

```text
selected pregame mode
saved default local profile selection
user selector choice
callsign text
identity_kind
local_profile_id
seed_from_guest_stats
single-player start request
authoritative match facts
```

Current outputs:

```text
selector rows
active single-player profile context
visible pregame callsign
default_profile response
created local_profile_id
profile readout stats
single-player room member local_profile_id
player-data match-result write identity
updated Guest or Local Profile stats
```

## Data crossing the flow

Local pilot selector data:

```text
local_profile_id
display_name
identity_kind
```

Local profile create data:

```text
display_name
seed_from_guest_stats
```

Local profile default data:

```text
identity_kind
local_profile_id
display_name
```

Profile readout data:

```text
play_mode
identity_kind
local_profile_id
callsign
activity_status
total_score
high_score
games_played
wins
ship_deaths
```

Single-player boot data:

```text
local_profile_id
```

Match-result persistence data:

```text
match_id
result_id
identity_kind
local_profile_id
score
ship_deaths
won
play_mode
```

## Out of scope

This domain flow does not define:

* local profile physical SQLite schema
* HTTP request/response contract source-of-truth
* generated packet schema
* client scene implementation details
* client code maps
* game-server code maps
* player-data code maps
* Authenticated Account login
* Discord OAuth
* online multiplayer admission
* local-to-online profile migration
* Local Profile to Authenticated Account sync
* progression, unlocks, inventory, achievements, or rewards
* leaderboard eligibility
* anti-cheat trust policy
* future player-data service extraction

## Related docs

* [Player Experience](./!INDEX.md)
* [Platform Account And Identity Current State](../platform/account-and-identity-current-state.md)
* [Client](../../services/client/!INDEX.md)
* [Client Local Pilot Flow](../../services/client/pregame-menu-flow/local-pilot-flow.md)
* [Client Profile Flow](../../services/client/pregame-menu-flow/profile-flow.md)
* [Client HTTP API Flow](../../services/client/client-http-api-flow.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Player Data](../../services/player-data/!INDEX.md)
* [Local Profiles HTTP API](../../services/player-data/local-profiles-http-api.md)
* [Profile Stats Flow](../../services/player-data/profile-stats-flow.md)
* [Runtime And Store Routing](../../services/player-data/runtime-and-store-routing.md)
* [Game Server Player Data HTTP Hosting](../../services/game-server/integrations/player-data-http-hosting.md)
* [Game Server Match Result Reporting](../../services/game-server/integrations/match-result-reporting.md)
* [HTTP Contract Enforcement](../../protocol/http-contract-enforcement.md)
* [Player Data Schema](../../data/player-data-schema.md)

## Notes

Local Profile is durable local identity, not Guest with saves and not an Authenticated Account cache.

Guest remains selectable even when no local profile storage is available.

Display name and callsign are presentation identity. Store routing uses `identity_kind`, `local_profile_id`, and `account_id`.

The profile readout path and local profile management path are related but separate. Local profile management changes profile records and defaults. Profile readout loads normalized stats for the active context.

The current implementation hosts player-data HTTP routes in the game-server process, but player-data owns the local profile behavior behind those routes.
