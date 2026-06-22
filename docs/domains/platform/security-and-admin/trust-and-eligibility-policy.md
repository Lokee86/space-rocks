# Trust And Eligibility Policy

Parent index: [Security And Admin](./!INDEX.md)

## Purpose

This doc records the current trust and eligibility policy for Space Rocks.

It explains which identity states, match facts, player-data routes, and result writes are eligible to become trusted platform facts under the current implementation.

This is a current platform domain doc, not a future anti-cheat planning doc.

## Overview

Space Rocks currently separates trust by identity, play mode, and authority boundary.

The current identity states are:

```text
Guest
Local Profile
Authenticated Account
```

Current trust policy:

```text
Guest
-> local transient only
-> not online-trusted

Local Profile
-> durable local-only profile
-> not online-trusted

Authenticated Account
-> Rails/API-backed online identity
-> only identity that may create online-trusted facts
```

The current implementation has no general anti-cheat verdict system, no review queue, no public leaderboard eligibility system, and no durable trust-verdict object.

Existing trust is enforced through narrower boundaries:

```text
identity kind
play mode
authenticated account requirement
server-authoritative match facts
player-data mode/identity validation
authenticated-account Rails persistence
result_id idempotency
local/offline import exclusion policy
```

## Trust vocabulary

Current docs and implementation imply these trust categories:

| Category                    | Meaning                                                                                          |
| --------------------------- | ------------------------------------------------------------------------------------------------ |
| Guest facts                 | Transient local facts held in runtime memory.                                                    |
| Local Profile facts         | Local-only durable facts stored through the local profile route.                                 |
| Authenticated Account facts | Online account-backed facts routed by `account_id`.                                              |
| Trusted match facts         | Game-server-produced match result facts.                                                         |
| Online-trusted facts        | Authenticated Account facts produced through valid online-authoritative flow.                    |
| Presentation facts          | Client-facing display facts that do not own persistence or authority.                            |
| Imported local facts        | Local/offline facts copied or transferred from local state. These are not online-trusted.        |
| Devtools/debug facts        | Development or debug-sourced facts. These are not currently a trusted online progression source. |

`online-trusted` does not mean the fact is true because the client presented it. It means the fact belongs to an Authenticated Account and was produced through the server-owned authoritative path that current systems accept.

## Current authority chain

The current trusted match-result chain is:

```text
gameplay authority
-> room match lifecycle
-> game-server match summary
-> match-result reporter
-> player-data runtime sink
-> player-data mode/identity validation
-> selected player-data route
-> backing store persistence
```

Current account-backed persistence chain:

```text
Authenticated Account
-> WebSocket auth verifies bearer token through Rails/API
-> multiplayer admission attaches account_id
-> gameplay produces authoritative result facts
-> player-data routes authenticated_account result by account_id
-> Rails/API persists match result and aggregate stats
```

The client is not part of the authority chain for score, deaths, wins, match-over state, or persistent account stat mutation.

## Identity trust policy

### Guest

Guest is a temporary local identity.

Guest may:

```text
play local single-player
route stats to transient guest memory
seed a Local Profile through supported local-profile creation flow
present local profile readout data
```

Guest must not:

```text
enter production online multiplayer
create online-trusted facts
write account-backed stats
write leaderboard results
become an Authenticated Account
migrate into online account state
```

Guest stats are useful for local transient play and local profile seeding only.

### Local Profile

Local Profile is a durable local-only identity.

Local Profile may:

```text
play local single-player
route stats to the local profile route
persist local match results and local aggregate stats when local storage is configured
participate in local profile UI and local callsign flow
```

Local Profile must not:

```text
enter production online multiplayer
create online-trusted facts
write account-backed stats
write leaderboard results
become an Authenticated Account
sync into an Authenticated Account
migrate into online account state
```

Local Profile is not a Rails/API cache.

Authenticated Account is not a synced Local Profile.

### Authenticated Account

Authenticated Account is the current online identity.

Authenticated Account may:

```text
authenticate through Rails/API
prove identity to the game server through WebSocket authentication
enter production online multiplayer
route account-backed profile reads by account_id
route account-backed match-result writes by account_id
create online-trusted facts through valid server-authoritative flow
```

Authenticated Account is the only identity that may create online-trusted facts.

The current canonical cross-system account identity is:

```text
account_id
```

Rails `user_id` remains an API-server internal database identity.

Bearer tokens prove an account session but are not gameplay identity.

Display names and callsigns are presentation identity, not routing identity.

## Play-mode eligibility

Current play-mode and identity policy:

| Play mode                | Guest    | Local Profile | Authenticated Account                   |
| ------------------------ | -------- | ------------- | --------------------------------------- |
| `single_player`          | allowed  | allowed       | rejected by player-data mode validation |
| `multiplayer`            | rejected | rejected      | allowed                                 |
| `multiplayer_simulation` | rejected | rejected      | allowed                                 |

Production multiplayer create and join require Authenticated Account identity.

If auth verification is unavailable, multiplayer create/join returns:

```text
auth_unavailable
```

If the session is unauthenticated, multiplayer create/join returns:

```text
auth_required
```

Guest and Local Profile are not admitted to online multiplayer.

## Match-result trust policy

The game server owns match-result authority.

Gameplay and room lifecycle own:

```text
match-over state
score
ship deaths
winner flag
match summary construction
result reporting trigger
```

Player-data accepts these as trusted upstream facts from the game server. Player-data does not recompute:

```text
score
ship deaths
winner state
match-over eligibility
room result
```

The API server also does not decide match outcome. It persists authenticated-account match-result facts submitted through the internal player-data route after upstream resolution.

The client must not be trusted for:

```text
score
deaths
wins
match completion
match outcome
match result persistence
account-backed stat mutation
```

The client may present result data returned through server-owned projections.

## Player-data trust policy

Player-data owns route selection and mode/identity validation.

Current player-data routes:

| Identity kind           | Route               | Backing behavior                             |
| ----------------------- | ------------------- | -------------------------------------------- |
| `guest`                 | guest route         | process-local transient memory               |
| `local_profile`         | local profile route | embedded SQLite when configured              |
| `authenticated_account` | account route       | Rails/API-backed persistence when configured |

Player-data validates identity and play-mode combinations before mutating backing stores.

Invalid mode/identity combinations are rejected before store mutation.

Player-data must not treat client-selected storage as authoritative. The client selects identity context, not SQLite, Rails/Postgres, or guest memory.

## Account-backed persistence eligibility

Authenticated-account profile reads require valid account identity.

Authenticated-account match-result writes require:

```text
identity_kind = authenticated_account
account_id
result_id
match_id
score
ship_deaths
won
valid internal service boundary
```

Rails/Postgres persistence is reached only through the authenticated-account player-data route.

The API server persists:

```text
PlayerStat
PlayerMatchResult
```

`result_id` is the current idempotency key for match-result writes.

Duplicate `result_id` submissions are accepted as duplicates and must not apply aggregate stats again.

## Local/offline import policy

Local/offline facts do not become online-trusted by being copied, imported, linked, displayed next to account UI, or associated with an Authenticated Account.

Current policy forbids importing or transferring these from Guest or Local Profile state into online-trusted account state:

```text
currency
inventory
unlocks
progression
achievements
leaderboard scores
ranked results
trusted match history
commerce entitlements
competitive challenge completions
anti-cheat-sensitive facts
```

The only possible local/online transfer category is:

```text
preferences/settings
```

Even preferences/settings import from offline to online requires screening and trust verification before it can exist.

The preferred direction for preferences/settings transfer is online to offline export.

## Migration and merge exclusions

Current policy permanently excludes:

```text
Guest to Authenticated Account migration
Local Profile to Authenticated Account migration
local-to-multiplayer migration
account merge
automatic Local Profile / Authenticated Account sync
```

Provider linking, account recovery, account deletion, and provider conflict handling must not imply account merge or data transfer.

## Development and devtools trust policy

Current devtools and debug behavior is not an online-trusted progression source.

Devtools may exercise gameplay, telemetry, debug packets, local testing, grant-handling paths, and presentation flows, but devtools/debug-sourced facts must not silently become online-trusted account facts.

Development-only auth bypass, if added later, must be build-flagged, environment-gated, unavailable on live deployed servers, and unable to create online-trusted facts.

The current production policy is:

```text
live deployed servers must not allow auth bypass
```

## Progression and reward eligibility dependency

Progression and rewards consumes eligibility and trust results. It does not own the trust policy.

Current trusted source event policy is not fully implemented as a broad progression system yet, but the existing boundary is:

```text
server-authoritative match facts
-> eligibility / policy check
-> reward or stats mutation path
-> player-data routing
-> durable store when eligible
```

Progression, rewards, achievements, inventory, commerce, and leaderboards must not independently decide that local/offline facts are online-trusted.

They must consume trust and eligibility decisions from the platform policy boundary once those systems exist.

## Leaderboard eligibility dependency

There is no current public leaderboard persistence or ranking system.

Existing policy already constrains future leaderboard eligibility:

```text
leaderboard facts must require Authenticated Account identity
leaderboard facts must require trusted server-authoritative result facts
leaderboard facts must not accept Guest or Local Profile results as online-trusted
leaderboard facts must not accept local/offline imports
leaderboard facts must not trust client-submitted score or result claims
```

Future leaderboard planning must consume this policy rather than redefining identity trust.

## Trust-sensitive facts

Trust-sensitive facts include any data that can affect online account value, competitive standing, durable progression, or public ranking.

Current trust-sensitive categories include:

```text
account-backed stats
match results
wins
score
ship deaths
progression
currency
inventory
unlocks
achievements
leaderboard submissions
ranked results
trusted match history
commerce entitlements
competitive challenge completions
```

Only Authenticated Account flow can produce online-trusted versions of these facts, and only where current systems actually implement the write path.

## Current limitations

The current implementation does not yet have:

```text
general trust verdict records
anti-cheat verdicts
anti-farming policy enforcement
debug_tainted result markers
review_required state
admin enforcement workflow
leaderboard eligibility checks
ranked-mode eligibility checks
custom-room eligibility checks
modded-room eligibility checks
public match-history eligibility checks
durable progression grant eligibility enforcement
```

Current trust protection is mostly structural:

```text
server-authoritative gameplay
identity separation
play-mode validation
authenticated multiplayer admission
player-data route selection
account-backed internal persistence
result_id idempotency
local/offline import exclusion
```

## Active issues

`start_single_player_request` does not currently reject an already-authenticated WebSocket session directly at the game-server boundary.

The intended model remains Guest or Local Profile for local single-player, and player-data mode validation rejects `single_player + authenticated_account`.

This is documented as a current system limit.

## Out of scope

This doc does not define:

```text
future anti-cheat implementation
anti-farming implementation
abuse review queues
admin enforcement tools
ban or suspension policy
appeals
leaderboard ranking formulas
ranked matchmaking policy
custom-room policy
modded-room policy
future progression formula policy
future reward formula policy
exact database schemas
exact HTTP or packet contracts
client UI layout
```

Future anti-cheat planning belongs in a separate platform planning doc.

Future abuse and enforcement planning belongs in a separate platform planning doc.

Protocol shapes belong in protocol docs. Physical data shape belongs in data and service docs.

## Related docs

* [Platform](domains/platform/!INDEX.md)
* [Account And Identity Current State](account-and-identity-current-state.md)
* [Account Backed Profile Flow](account-backed-profile-flow.md)
* [Player Data Routing Flow](player-data-routing-flow.md)
* [Gameplay Session Flow](gameplay-session-flow.md)
* [Match End And Results Flow](match-end-and-results-flow.md)
* [Local Pilot Profile Flow](local-pilot-profile-flow.md)
* [Realtime Client Server Flow](realtime-client-server-flow.md)
* [Player-data Match Result Sinks](match-result-sinks.md)
* [Player-data Runtime And Store Routing](runtime-and-store-routing.md)
* [API-server Player Stats And Match Results](player-stats-and-match-results.md)
* [Game-server Match Result Reporting](match-result-reporting.md)
* [Current System Limits](current-system-limits.md)
* [Account And Identity Systems planning](account-and-identity-systems.md)

## Notes

This doc captures current policy implied by the implemented identity, player-data, match-result, and account-backed persistence flows.

It intentionally does not keep the old combined “game integrity polciy” ownership model. Trust and eligibility policy decides whether a fact is eligible to become trusted. Anti-cheat decides whether gameplay or result generation appears manipulated. Abuse and enforcement decide what happens after suspicious, rejected, or disputed behavior.

Current implementation has the first layer structurally, but not the full future anti-cheat or enforcement layers.
