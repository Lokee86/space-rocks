# Season Format And Progression

Parent index: [Platform Planning](./!INDEX.md)

## Purpose

This document plans the season-format and campaign-progression domain for Space Rocks.

It defines how season/campaign periods work, how seasonal participation state is tracked, how campaign content becomes permanent, how season rewards interact with durable progression, and how online and single-player campaign paths stay aligned.

This is a platform-domain planning document. It is not a core XP document, not a leaderboard formula document, not a commerce/RMT policy document, and not an account-reset model.

## Overview

Space Rocks seasons are campaign-style live periods.

They are not account resets.

The planned model is:

```text
durable player progression
-> campaign season period
-> seasonal participation state
-> campaign objectives and rewards
-> durable grants through progression/rewards
-> archived season history
-> permanent content availability
```

A season gives structure, context, current activity, campaign goals, and period-scoped standings.

It does not wipe durable account state or make gameplay content disappear when the period ends.

## Current status

Active planning.

Current implementation already has useful foundations:

```text
server-authoritative match results
local profile progression state
authenticated account identity
guest transient state
progression/reward grant planning
leaderboard/ranking season-period planning
inventory and hangar ownership planning
commerce/RMT boundary planning
```

Current implementation does not yet have:

```text
season/campaign definitions
active season state
season participation state
campaign objective state
season reward-track state
season archive records
campaign replay/permanent conversion policy
single-player campaign parity system
season-to-leaderboard coordination
season-to-reward grant coordination
```

## Ownership boundary

This document owns planning for:

```text
season/campaign lifecycle
season calendar
active and archived season state
season period keys
campaign identity
campaign participation state
seasonal objective grouping
campaign reward-track state
season reward eligibility
season archive summaries
season-to-leaderboard coordination
season-to-progression grant triggers
online and single-player campaign parity policy
FOMO-light season policy
season content permanence policy
```

This document does not own:

```text
core XP formulas
durable account level
rank or insignia derivation
wallets
currency definitions
shop purchases
paid passes
RMT implementation
inventory ownership
owned ship/weapon/module persistence
leaderboard board definitions
rating formulas
match scoring
trusted match result construction
abuse invalidation
physical database schema
exact API route shapes
```

Progression And Rewards owns XP, level, durable grant construction, unlock grants, and reward idempotency.

Leaderboards And Rankings owns board definitions, seasonal board views, archived boards, ranking formulas, and ranking lifecycle.

Shop, Commerce, And Economy owns purchases, RMT boundaries, entitlement records, future paid models, and commerce review.

Inventory And Hangar owns owned ships, weapons, modules, inventory items, and ownership state.

Match Outcomes And Results owns trusted match facts that may feed season objectives or rewards.

## Core season model

A season is a campaign period.

A campaign season may define:

```text
campaign_id
season_id or period_key
display name
active window
campaign content set
campaign objectives
campaign reward track
eligible modes or missions
online participation state
single-player campaign state
archive behavior
```

Season lifecycle:

```text
planned
-> active
-> closing
-> archived
-> permanent content path
```

During the active period, the campaign may drive current objectives, current standings, live events, and period-specific reward goals.

After the active period, the campaign is archived. Earned rewards remain durable, and gameplay content introduced by the campaign remains available through permanent paths.

## Anti-reset policy

A new season must not reset durable player progress.

Seasons do not reset:

```text
XP
account level
rank or insignia
wallet balances
ships
weapons
modules
inventory
achievements
titles
permanent cosmetics
personal records
account identity
public profile identity
trust or integrity history
abuse or enforcement history
```

Season end may close or archive:

```text
current season standings
seasonal objective progress
campaign reward-track progress
season participation summaries
event-specific counters
current-season display state
```

Closing period state is not an account reset.

## Content permanence

Normal seasonal gameplay additions become permanent game content.

Examples:

```text
ships
weapons
modules
missions
encounters
bosses
mode variants
campaign challenge paths
campaign reward paths
```

The normal policy is:

```text
season introduces gameplay content
-> active season creates current participation path
-> season ends
-> content remains available permanently
```

Seasonal gameplay exclusivity is not normal policy.

Players who miss a campaign season should not be permanently locked out of gameplay power, campaign missions, or core content paths.

Possible permanent acquisition paths include:

```text
campaign replay
normal progression
mission chains
shop unlock paths
achievement paths
milestone paths
archived campaign paths
```

Holiday or special limited-time events may be exceptions, but they should be treated as event exceptions, not normal season policy.

## No early-access season rewards

Campaign seasons should not use early-access rewards as a normal incentive.

Avoid this model:

```text
play or pay now to use content before everyone else
```

Preferred model:

```text
this campaign introduces new permanent content and reward paths
```

Early access creates avoidable fairness, ownership, availability, and communication problems.

## Seasonal progression state

Seasonal progression is separate from durable account progression.

Possible season state:

```text
season_points
season_track_level
season_objective_progress
season_challenge_progress
season_completion_state
season_participation_summary
claimed_reward_refs
archived_campaign_record
```

This state belongs to the season/campaign period.

It should not replace durable progression state.

When seasonal rewards are earned, they should be granted through the normal durable reward path.

## XP policy

Seasons should avoid XP rewards.

XP belongs to durable account progression and should not become the main seasonal incentive.

Normal gameplay inside campaign content may still award ordinary XP through standard progression rules.

Season-specific reward design should prefer:

```text
ships
weapons
modules
hardwires where appropriate
Orebits
titles
badges
cosmetics
rare drops
campaign unlocks
shop or progression access unlocks
achievement rewards
milestone rewards
```

## Power reward policy

Campaign seasons may grant gameplay-affecting rewards.

Allowed reward categories include:

```text
ships
weapons
modules
hardwires where appropriate
gameplay unlocks
campaign-specific unlocks
```

Power rewards must not become permanently unavailable because a player missed the active season.

Rule:

```text
campaign seasons may grant gameplay-affecting rewards,
but those rewards must remain obtainable through permanent non-seasonal paths
after the campaign period ends
```

This allows season rewards to matter without creating permanent missed-season disadvantage.

## FOMO-light policy

Time-limited events inherently create some FOMO. The design should not amplify that in a predatory way.

Acceptable pressure:

```text
current campaign is active
current standings close later
current event participation is time-boxed
campaign archive records preserve who participated
```

Avoid:

```text
daily-login punishment loops
artificial claim pressure
predatory grind pacing
paid skip pressure
paid catch-up pressure
paid seasonal power pressure
temporary power vaults
permanent missed-season disadvantage
```

Campaigns should feel current and worth playing, not coercive.

## Monetization boundary

Season monetization is not decided in this document.

Season/campaign systems are live-service work and may create real operating costs. Monetization options should remain open, subject to explicit commerce, product, platform, and legal review.

This document does not approve, define, or require any paid season model.

V0 should not include:

```text
paid season pass
paid battle pass
paid skips
paid loot boxes
premium seasonal entitlement
paid seasonal power exclusivity
```

Future monetization, if reviewed and approved, must not violate:

```text
anti-reset policy
FOMO-light policy
content permanence policy
no seasonal gameplay exclusivity
permanent acquisition paths for power rewards
single-player campaign parity
```

Compatible future options may include:

```text
cosmetic supporter packs
campaign supporter bundles
profile cosmetics
ship skins
visual, audio, or UI cosmetics
DLC-style campaign packs
supporter subscriptions
founder or supporter entitlements
permanent content purchases
non-exclusive convenience bundles
```

Higher-risk options require explicit review before planning or implementation:

```text
paid season tracks
paid skips
premium currency
paid consumables
paid catch-up
limited-time paid bundles
paid power bundles
```

Commerce and RMT planning owns any final decision.

## Single-player campaign parity

Single-player should receive the same campaign content or equivalent campaign content.

Online campaign path:

```text
multiplayer, co-op, or ranking-aware campaign version
authenticated account participation state
online seasonal boards where applicable
```

Single-player campaign path:

```text
local/offline equivalent campaign version
local profile campaign state
same or equivalent gameplay rewards
no public seasonal ranking requirement
```

Single-player does not need identical live standings, but it should not become second-class for permanent campaign gameplay content.

Local single-player rankings do not become seasonal public boards.

## Guest, local, and authenticated behavior

Guest season participation is transient.

```text
Guest
-> transient participation only
-> no durable online season record
-> no public season standings
```

Local Profile campaign participation is local and single-player only.

```text
Local Profile
-> local campaign state
-> local reward ownership
-> no online season import
-> no public seasonal rankings
```

Authenticated Account campaign participation is the online season path.

```text
Authenticated Account
-> durable online campaign state
-> eligible online season rewards
-> seasonal leaderboard participation where applicable
-> archived public/private season history according to privacy policy
```

Local and online campaign state should remain separate.

## Leaderboard relationship

Seasonal leaderboards are period-scoped ranking surfaces.

Season format owns the active period and campaign context.

Leaderboards And Rankings owns the board definitions, ranking calculations, entry lifecycle, privacy behavior, and archived boards.

Typical flow:

```text
season period exists
-> leaderboard board supports season scope
-> eligible results populate current season board
-> season closes
-> board is archived
-> durable progression remains unchanged
```

Seasonal boards do not imply account resets.

## Campaign archive behavior

When a season ends, the system should archive useful history without deleting durable facts.

Archive records may preserve:

```text
campaign_id
season_id or period_key
active window
completion state
earned reward refs
participation summary
final leaderboard refs
campaign objectives completed
campaign-specific titles or badges earned
```

Archive behavior should support:

```text
player profile history
campaign replay or review
website history where applicable
leaderboard archive links
admin/support inspection
```

Archived visibility must respect privacy policy and later visibility changes where applicable.

## Events and limited-time exceptions

Normal campaign seasons should make gameplay additions permanent.

Limited-time events are allowed as exceptions, especially for holiday or special-event content.

Event exceptions should be explicit and narrow.

Examples:

```text
holiday mode variant
temporary event encounter
limited-time cosmetic challenge
special event leaderboard
```

Event exceptions should not become a loophole for permanent missed gameplay power.

## Implementation sequence

1. Define the `SeasonCampaign` identity model.
2. Define season lifecycle states and period keys.
3. Define active and archived season state.
4. Define seasonal participation state for authenticated accounts.
5. Define local single-player campaign state.
6. Define seasonal objective grouping.
7. Define campaign reward-track state.
8. Route season rewards through Progression And Rewards `GrantAward` handling.
9. Define permanent acquisition paths for campaign power rewards.
10. Coordinate season period keys with Leaderboards And Rankings.
11. Add archive records for completed campaign seasons.
12. Add campaign replay or permanent-content conversion rules.
13. Add single-player equivalent campaign content rules.
14. Reserve monetization handoff points to Commerce without implementing paid season behavior.
15. Add admin/support inspection for season participation and reward grants.

First useful slice:

```text
one campaign_id
+ active period key
+ one campaign objective group
+ one durable reward grant
+ local and authenticated participation records
+ archive-on-close behavior
```

## Open decisions

Implementation-shape decisions remain:

```text
exact season length
exact campaign_id and season_id naming
exact season lifecycle state names
exact season participation schema
exact local campaign state schema
exact reward-track shape
whether archived campaigns are fully replayable, converted, or both
exact campaign archive UI
exact single-player parity rules per campaign type
exact holiday/special-event exception policy
exact permanent acquisition path per power reward
exact campaign website surface
exact monetization review process
```

These are not open policy questions:

```text
whether seasons reset accounts
whether seasons are campaign-style periods
whether normal campaign gameplay content becomes permanent
whether normal seasonal gameplay exclusivity exists
whether early access is a normal season reward
whether seasons should focus on XP rewards
whether power rewards are allowed
whether power rewards require permanent acquisition paths
whether single-player needs same or equivalent campaign content
```

## Core invariants

```text
Seasons do not reset accounts.

Seasons are campaign-style live periods.

Seasonal progression is separate from durable account progression.

Seasonal gameplay additions become permanent content.

Seasonal gameplay exclusivity is not normal policy.

Holiday or special events may be limited-time exceptions.

Campaign seasons can grant power rewards.

Power rewards must have permanent acquisition paths.

Seasons should avoid XP rewards.

Normal campaign gameplay may still grant ordinary XP through standard progression.

Early-access season rewards are not normal policy.

FOMO should remain light and non-predatory.

Single-player receives the same or equivalent campaign content.

Local campaign state and online campaign state remain separate.

Guest season participation is transient.

Seasonal leaderboards are period-scoped and archived.

Seasonal leaderboards do not imply account resets.

Season monetization remains open but undecided.

Any season monetization belongs to commerce/RMT review.

Monetization must not create resets, paid pressure, permanent missed power, or predatory FOMO.

Season rewards use the normal durable grant path.

Campaign content permanence is part of the season contract.
```

## Related docs

* [Platform Planning](./!INDEX.md)
* [Leaderboards And Rankings](leaderboards-and-rankings.md)
* [Progression And Rewards](../gameplay/progression-and-rewards.md)
* [Inventory And Hangar](../gameplay/inventory-and-hangar.md)
* [Shop, Commerce, And Economy](../gameplay/shop-commerce-and-economy.md)
* [Match Outcomes And Results](../gameplay/match-outcomes-and-results.md)
* [Modes And Match Rules](../gameplay/modes-and-match-rules.md)
* [Levels, Missions, And Content Structure](../gameplay/levels-missions-and-content-structure.md)
* [Game Integrity Policy](security-and-admin/game-integrity-policy.md)

## Notes

This document intentionally treats seasons as campaign structure, not account churn.

The goal is to support freshness, live context, campaign releases, current standings, and durable rewards without turning seasons into resets, missed-power traps, or monetization funnels.

The planning direction is stable: campaigns can matter, rewards can include power, content becomes permanent, single-player gets parity, and monetization remains reviewed future work rather than baked into the season model.
