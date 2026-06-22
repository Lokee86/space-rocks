# Leaderboards And Rankings

Parent index: [Platform Planning](./!INDEX.md)

## Purpose

This document plans the leaderboard and ranking surface for Space Rocks.

It defines which ranking surfaces exist, which durable facts they consume, how boards are scoped and filtered, where ranking information is visible, and how board lifecycle, privacy, and maintenance behave.

This is a platform-domain planning document. It is not a current implementation document, not a scoring policy document, not a match rules document, and not an abuse enforcement plan.

## Overview

Leaderboards and rankings compare durable eligible player performance.

They do not rank grind volume.

The planned ranking model is:

```text
eligible durable result or durable calculated stat
-> board definition
-> scope and filter selection
-> ranking calculation
-> visible board/profile/post-match/website surface
```

Online rankings are built from durable Authenticated Account results and stats.

Single-player rankings are built from durable Local Profile results and stats.

Guest play is excluded from rankings because unsaved guest play does not produce durable ranking data.

The same board definitions should be usable for online and local ranking surfaces where the required facts exist. Online rankings support multiplayer, social, region, season, and website surfaces. Local rankings use local durable data and omit multiplayer/social/public web filters.

## Current status

Active planning.

There is no current public leaderboard persistence or ranking system.

Current implementation already has useful foundations:

```text
server-authoritative match results
player-data identity and mode routing
authenticated-account result persistence
local profile durable result persistence
guest transient result handling
result_id idempotency
account_id as internal authenticated-account identity
```

Current implementation does not yet have:

```text
leaderboard board definitions
public ranking entries
public profile ranking identifiers
ranking privacy controls
seasonal board lifecycle
Pilot Rating
Combat Rating
website ranking surface
skill aggregate ranking calculations
board catalog source of truth
```

## Ownership boundary

This document owns planning for:

```text
leaderboard product surface
ranking board families
ranking filters and scopes
single-player/local ranking behavior
online ranking behavior
Pilot Rating and Combat Rating surface ownership
ranking visibility surfaces
public listing privacy behavior
board definition model
board lifecycle states
ranking entry visibility states
website ranking surface requirements
ranking source-of-truth expectations
```

This document does not own:

```text
mode rule definitions
scoring formulas
match-end rules
match result construction
achievement definitions
progression rank/title derivation
XP rank or insignia derivation
display-name moderation policy
account authentication
social relationship implementation
abuse enforcement
manual admin tooling
physical database schema
exact API request/response shape
website visual design
```

Modes and match rules own what each mode produces.

Match outcomes own authoritative match result finalization.

Game integrity owns result classification, review status, and lane-specific eligibility.

Account and identity owns account identity, public profile identity, display names, and display-name moderation.

Social systems own friends, blocks, mutes, teams/clans if added later, and relationship data.

Leaderboards consume those systems and decide how eligible facts become visible comparison surfaces.

## Core ranking model

A ranking board is a named, scoped comparison surface.

Every board should define:

```text
what it ranks
which facts it consumes
which mode or ruleset it belongs to
how entries are sorted
which tie-breakers apply
which filters are supported
which surfaces may display it
which lifecycle state it is in
```

Use `Primary Sort Value` for the main value a board ranks by.

Examples:

```text
High Score
-> primary sort value: score
-> direction: descending

Fastest Clear
-> primary sort value: completion_time
-> direction: ascending

Fewest Deaths Clear
-> primary sort value: deaths
-> direction: ascending

Pilot Rating
-> primary sort value: pilot_rating
-> direction: descending
```

## Identity and data sources

### Guest

Guest play is excluded from rankings.

Guest may appear in immediate match presentation, but guest results do not create durable ranking entries.

```text
Guest
-> no durable ranking entries
-> no online rankings
-> no local profile rankings
-> no website rankings
```

### Local Profile

Local Profile rankings are local and single-player only.

They are populated from durable local profile stats and results.

They use the same board definitions where local durable data has the required facts.

They do not support multiplayer, social, region, season, public profile, or website filters.

```text
Local Profile
-> durable local ranking source
-> single-player rankings
-> local board browser
-> local post-match ranking impact
-> no public website surface
-> no online sync
```

Local single-player rankings do not have seasons.

### Authenticated Account

Authenticated Account rankings are online ranking sources.

They are populated from durable account-backed stats, results, ratings, and board entries.

```text
Authenticated Account
-> durable online ranking source
-> public board listings by default
-> privacy opt-out supported
-> website ranking surface
-> seasonal/event ranking support
```

`account_id` remains the internal authenticated-account identity.

A separate public UUID should be used for public profile URLs and public ranking lookup.

Recommended identity split:

```text
account_id
-> internal authenticated-account identity

public_profile_id
-> public profile and website lookup identity

display_name
-> moderated presentation identity
```

Display names are not durable identity and do not need to be globally unique.

## Board families

### Single-result performance boards

Single-result boards rank one stored eligible result.

Examples:

```text
Survival Arcade - High Score
Survival Arcade - Longest Survival
Score Attack - Fastest Clear
Score Attack - Fewest Deaths Clear
Mission - Fastest Clear
Mission - Highest Completion Score
Boss Encounter - Fastest Kill
Boss Encounter - Fewest Deaths
Perfect Clear boards
```

These boards should point back to the result that earned the entry.

### Skill aggregate boards

Skill aggregate boards rank calculated durable performance stats.

They are allowed only when they measure skill, consistency, or earned performance.

Examples:

```text
clear rate
win rate
average score per eligible match
score per minute
average deaths per clear
perfect clear rate
boss clear rate
mission grade average
objective completion rate
```

Skill aggregate boards require minimum sample thresholds before a player can appear on the public board.

Exact thresholds are gametime balancing decisions and should be decided when the relevant aggregate exists.

The system should be able to show a calculated stat on a private profile before making it eligible for public ranking.

### Excluded grind boards

Leaderboards should not rank raw grind volume.

Excluded public board types include:

```text
total games played
total score
total asteroids destroyed
total kills
total wins
total currency earned
```

These values may still exist as private stats, profile facts, diagnostics, or progression inputs. They should not become public leaderboard boards.

### Rating boards

Ratings are separate from scoreboards and aggregate stat boards.

Planned rating families:

```text
Pilot Rating
-> PvE performance rating

Combat Rating
-> PvP competitive rating
```

Pilot Rating is not assumed to be Elo. It should be based on eligible PvE performance once modes, missions, challenges, and difficulty tiers are mature enough.

Combat Rating is the PvP rating surface and may use a more conventional competitive rating model later.

Pilot Rating and Combat Rating should not be collapsed into one generic player rank.

### Profile calculated stats

Some calculated stats should be profile-visible without necessarily becoming boards.

Examples:

```text
average score
score per minute
clear rate
perfect clear rate
average deaths per clear
best mode
best mission grade
Pilot Rating
Combat Rating
```

Only selected, sample-thresholded calculated stats should become public ranking boards.

## Filters and scopes

Boards should support filters and scopes where the required data exists.

### Common online filters

```text
mode
ruleset
season or period
region
friends
room
party
last match
around my rank
solo or co-op
human / bot / TAS lane
ship or loadout category where reasonable
```

Not every board supports every filter.

Each board definition should declare its supported filters.

### Single-player local filters

Local single-player rankings use the same board options where local durable data exists, but omit multiplayer/public filters.

Supported local filters may include:

```text
mode
ruleset
local profile
last match
solo
ship or loadout category where supported
```

Unsupported local filters include:

```text
global online
region
friends
public website ranking
Combat Rating unless local PvP exists later
```

### Room, party, and last-match scopes

Room, party, and last-match scopes are ranking filters or view presets.

They reuse the same board definitions and sorting rules, but restrict the candidate set.

```text
Room
-> players currently in the room

Party
-> party members

Last Match
-> participants in the last completed match
```

These scopes do not create separate durable global board records.

They can mimic durable boards for local comparison without becoming persistent public board categories.

### Region

Region should be result-derived, not freely player-selected.

Preferred source order:

```text
match server region
matchmaking region
deployment region
fallback region only if needed
```

Store and display normalized regions only.

Examples:

```text
na
eu
apac
sa
oce
```

The ranking surface should not expose raw IP-derived data.

### Friends and social filters

Friends filters depend on social systems.

The leaderboard system should treat friends as a scope/filter over existing board definitions.

Social systems own the relationship graph, blocks, mutes, and future teams/clans.

## Seasons, periods, and lifetime records

Seasonal boards are the primary public competitive surface.

Planned period surfaces:

```text
current season
archived seasons
current event
archived events
```

Lifetime public boards are optional and may be excluded because they become difficult for new players to contest.

Lifetime personal records may still appear on profiles.

Single-player local rankings do not have seasons. They are persistent local records based on available local durable data.

Archived online boards must respect later privacy opt-outs.

## Privacy and public listing

Online ranking entries are public by default.

Players may opt out of public leaderboard listings.

Privacy controls should be granular enough to support different visibility surfaces.

Possible visibility flags:

```text
show_on_public_boards
show_on_recent_views
show_on_friends_boards
show_on_profile_rankings
show_on_all_public_ranking_surfaces
```

Exact flag names are gametime implementation decisions.

Privacy controls affect display and listing. They do not delete underlying durable results, private stats, integrity records, or account history.

Archived boards must respect current privacy settings.

## Human, bot, and TAS ranking lanes

Ranking lanes should consume game-integrity eligibility.

Planned lanes:

```text
human
bot
tas
```

Human boards are the default public competitive surface.

Bot and TAS boards may exist as separate board categories if intentionally exposed.

Bot and TAS results must not appear in human boards.

Debug, load-test, and integrity-test results should not appear in public ranking boards.

## Team rankings

Teams are ad-hoc.

Persistent teams and clans are not required for this ranking plan.

Team boards may exist, but a team entry lists the whole ad-hoc team that earned the result.

Planning shape:

```text
TeamLeaderboardEntry
- board_id
- team_members
- primary_sort_value
- tie_break_values
- achieved_at
- match or result refs
```

Future persistent teams or clans belong to social and community planning.

## Visibility surfaces

### In-game rankings browser

The in-game rankings browser is the primary game-client ranking surface.

It should support:

```text
board catalog
filter selection
top entries
my standing
around-my-rank entries
room / party / last-match views
entry details
local and online ranking contexts where available
```

### Post-match results

Post-match results should show relevant ranking impact.

Examples:

```text
new personal best
new board entry
current rank
rank change
next nearby entry
not ranked reason when safe
```

Post-match should not need to show a full board. It should show the relevant outcome and route the player to the board browser where appropriate.

### Player profile

Profiles may show:

```text
top leaderboard placements
personal records
current Pilot Rating
current Combat Rating
season standing
selected calculated stats
ranking badges or percentile indicators
```

Profile ranking panels must respect privacy settings.

### Lobby and player card

Lobby surfaces should stay compact.

Possible display fields:

```text
Pilot Rating
Combat Rating
best visible leaderboard badge
season tier or standing
progression rank/title where relevant
```

The lobby should not become a full leaderboard browser.

### Website

The website ranking surface is required.

Website surfaces should include:

```text
public leaderboard browser
public player profiles
player ranking pages
season standings
event standings
archived boards
shareable board URLs
shareable public profile URLs
```

The website and in-game client should consume the same ranking catalog or source of truth.

## Board definitions and source of truth

Ranking boards should be maintained from one source of truth.

The in-game client and website must not maintain separate hardcoded leaderboard catalogs.

A board definition should specify:

```text
board_id
display_name
board_family
source_type
mode_scope
ruleset_scope
primary_sort_value
sort_direction
tie_breakers
minimum_sample_requirement
required_result_facts
required_durable_stats
supported_filters
supported_scopes
supported_surfaces
lifecycle state
version or ruleset compatibility
```

Example:

```text
board_id: survival_arcade.high_score
display_name: Survival Arcade High Score
board_family: single_result_performance
source_type: match_result
mode_scope: survival_arcade
primary_sort_value: score
sort_direction: descending
tie_breakers:
- deaths ascending
- achieved_at ascending
```

The exact source-of-truth implementation is deferred.

Possible approaches include:

```text
shared catalog file
database-seeded catalog
API-served catalog
generated client and website constants
hybrid generated catalog plus runtime lifecycle state
```

The invariant is that board definitions must not drift between API, client, and website.

## Ranking entry state

A visible board entry should preserve why it is visible or hidden.

Useful entry states:

```text
active
hidden_by_privacy
hidden_by_admin
invalidated_by_lifecycle
invalidated_by_integrity
archived
stale
recalculated
```

Avoid using generic `removed` as the primary stored state. The useful state is why an entry is no longer visible or ranked.

Underlying durable results should not be deleted merely because a board entry is hidden, invalidated, recalculated, or removed from public listing.

## Lifecycle, maintenance, and ownership

Ruleset and versioning tie directly into leaderboard lifecycle.

Lifecycle and maintenance issues include:

```text
season reset
event closure
ruleset/version archival
global scoring bug correction
board definition correction
broad rebuild from source of truth
privacy opt-out propagation
```

These belong to leaderboard/ranking lifecycle and maintenance planning.

Abuse and enforcement issues include:

```text
hacking
exploitation
integrity invalidation
account enforcement
abuse-related result removal
repeat offender handling
```

These belong to game integrity, abuse, and enforcement planning. Leaderboards consume the resulting eligibility or invalidation decision and update visible board state.

Admin and support issues include:

```text
stale entries
manual correction
mistaken visibility state
profile or display correction
admin-triggered refresh
public listing correction not tied to abuse
```

These belong to admin/support tooling. Leaderboards consume the correction and update visible board state.

## API and persistence expectations

Exact endpoint shapes belong to protocol and API contract docs when implemented.

Planning expectations:

```text
clients can read board catalog
clients can read board entries
clients can read my standing
clients can read around-my-rank windows
clients can read room / party / last-match ranking views
website can read public boards and public profiles
internal systems can update or rebuild board state from eligible durable facts
```

Public clients must not submit trusted leaderboard scores.

Leaderboard writes and recalculations should derive from trusted durable results, durable calculated stats, rating state, or internal maintenance processes.

## Implementation sequence

1. Define the board definition model and source-of-truth expectation.
2. Define public profile UUID behavior for ranking URLs and public lookup.
3. Add ranking privacy settings for public, recent, friends, profile, and broad public listing behavior.
4. Define the initial board catalog from mature mode/ruleset data.
5. Add local ranking reads from durable Local Profile data where required facts exist.
6. Add online ranking reads from durable Authenticated Account data.
7. Add post-match ranking impact display.
8. Add in-game ranking browser.
9. Add website leaderboard and public profile surfaces.
10. Add seasonal and archived board lifecycle behavior.
11. Add skill aggregate ranking support with minimum sample requirements.
12. Add Pilot Rating once PvE performance inputs are mature enough.
13. Add Combat Rating once PvP/ranked inputs are mature enough.
14. Add team board entries for ad-hoc teams where modes support team results.
15. Add lifecycle maintenance, admin correction, and integrity invalidation consumption.

## Open decisions

Implementation-shape decisions remain:

```text
exact privacy flag names
exact public profile UUID field name
exact board catalog source-of-truth implementation
which calculated stats become profile-only versus public boards
exact minimum sample thresholds for each aggregate board
exact Pilot Rating formula
exact Combat Rating formula
exact initial board catalog once modes and rulesets mature
exact website route shape
exact board entry persistence shape
exact archived-board privacy propagation implementation
```

These are not open policy questions:

```text
whether guest results are ranked
whether grind boards exist
whether local rankings sync online
whether public listings are opt-out
whether archived boards respect later privacy opt-outs
whether Pilot Rating and Combat Rating are separate
whether persistent teams are required
whether website ranking surface exists
```

## Core invariants

```text
Leaderboards and rankings compare durable eligible performance.

Guest play is excluded from rankings.

Local Profile rankings are local-only and use durable local data.

Authenticated Account rankings use durable account-backed data.

Public online listings are enabled by default and can be opted out.

Archived boards must respect later privacy opt-outs.

Rankings do not rank raw grind volume.

Skill aggregate boards require sample thresholds.

Pilot Rating is the PvE rating surface.

Combat Rating is the PvP rating surface.

Single-player uses the same board definitions where local data supports them.

Single-player local rankings do not have seasons.

Online boards may use seasons, events, regions, friends, room, party, and last-match scopes.

Room, party, and last-match scopes are filters or view presets, not separate durable global boards.

Team boards use ad-hoc teams and list the whole team.

Persistent teams and clans are not required.

Board definitions are maintained from one source of truth.

The in-game client and website must not maintain separate leaderboard catalogs.

Public clients never submit trusted leaderboard scores.

Board lifecycle and maintenance are separate from abuse enforcement and admin/support correction.

Underlying durable results are not deleted merely because a board entry is hidden, invalidated, archived, recalculated, or removed from public listing.
```

## Related docs

* [Platform Planning](./!INDEX.md)
* [Account And Identity Systems](account-and-identity-systems.md)
* [Game Integrity Policy](game-integrity-policy.md)
* [Trust And Eligibility Policy](trust-and-eligibility-policy.md)
* [API Product Surface](../../protocol/api-product-surface.md)
* [Website And Web Presence](../web/website-and-web-presence.md)

## Notes

This document intentionally plans the completed non-stub leaderboard and ranking product surface, not a minimal first slice.

The exact board catalog should wait until modes, rulesets, scoring, match summaries, and player experience systems are mature enough to produce meaningful comparisons.

The planning direction is stable: durable performance rankings, no grind boards, local and online sources separated, website required, privacy respected, and board definitions maintained from a single source of truth.
