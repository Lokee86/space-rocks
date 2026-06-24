# Progression And Rewards
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc plans the durable progression and reward architecture for player advancement, XP, level-derived rank and insignia, earned reward grants, durable unlocks, rare persistent drops, and reward award construction.

The core purpose is to define how trusted gameplay, progression, account, or commerce events become durable `GrantAward` records that can be handed to player-data safely and idempotently.

## Ownership Boundary

This doc owns:

* XP award policy
* level advancement policy
* rank and insignia derivation policy
* earned-currency grant policy
* unlock reward policy
* rare persistent drop reward policy
* durable grant-source vocabulary
* `GrantAward` and `Grant` construction
* stable award/grant ID requirements
* progression-owned completion state
* reward idempotency expectations

This doc does not own:

* the currency system itself
* shop pricing, sinks, purchases, receipts, refunds, or commerce balance
* physical storage or database routing
* Local Profile versus Authenticated Account routing
* Rails/Postgres, SQLite, guest memory, or Redis implementation details
* runtime-only pickup or powerup behavior
* UI reward reveal layout
* trusted match-result facts
* achievement definitions
* inventory or hangar ownership model

Currency itself belongs to [Shop, Commerce, And Economy](shop-commerce-and-economy.md).

Player-data routing belongs to [Player Data And Persistence](../../services/player-data/!INDEX.md).

Inventory ownership belongs to [Inventory And Hangar](inventory-and-hangar.md).

Trusted match facts belong to [Match Outcomes And Results](match-outcomes-and-results.md).

Achievement and milestone reward definitions belong to this doc.

## Core Architecture

```text
Trusted Source Event
-> Eligibility / Policy Check
-> Reward Evaluation
-> GrantAward Construction
-> Player-Data Handler
-> Player-Data Runtime Route
-> Durable Store / API Application
```

Progression and rewards constructs `GrantAward` records.

Progression and rewards does not choose SQLite, Rails/Postgres, guest memory, or Redis directly.

The player-data handler/runtime receives the whole `GrantAward` and routes it by mode and identity to the correct player-data destination:

```text
Guest
-> transient player-data memory

Local Profile
-> local player-data route / SQLite-backed persistence

Authenticated Account
-> Rails API / Postgres-backed persistence
```

The selected player-data destination owns applying the grants.

## Durable Grant Sources

A durable grant source is a trusted reason to mutate player progression, earned currency balance, unlocks, titles, or other durable player-owned state.

Planned durable grant sources:

```text
match_completion
objective_completion
mission_completion
challenge_completion
achievement_completion
milestone_completion
rare_drop_collection
content_progression_unlock
shop_purchase
entitlement
refund_or_reversal
account_migration
admin_grant
devtools_test_grant
seasonal_or_event_reward
battle_pass_or_track_reward
```

Source meanings:

* `match_completion` - post-match XP, earned currency, unlock progress, or reward bundles from trusted match results.
* `objective_completion` - rewards from trusted objective success when objective rules produce persistent rewards. This will likely be limited to first-time objective completions unless a mode explicitly supports repeatable objective rewards.
* `mission_completion` - rewards from authored mission completion.
* `challenge_completion` - rewards from challenge completion.
* `achievement_completion` - one-time accomplishment rewards.
* `milestone_completion` - threshold-based track rewards.
* `rare_drop_collection` - durable collected rare drops only. Normal runtime pickups and runtime-only rare effects are excluded.
* `content_progression_unlock` - progression rules unlocking ships, weapons, modules, modes, missions, challenges, titles, cosmetics, or other content.
* `shop_purchase` - grants caused by a completed purchase flow. Purchase flow and pricing belong to commerce/economy.
* `entitlement` - durable ownership or access granted by an external account, platform, store, DLC, founder pack, promotional grant, or product-license system.
* `refund_or_reversal` - grant removal, correction, or ownership adjustment caused by refund/reversal handling.
* `account_migration` - grants or copied state caused by an explicit account migration path.
* `admin_grant` - operator-created correction, support grant, or account adjustment.
* `devtools_test_grant` - debug-only grant source for testing grant handling.
* `seasonal_or_event_reward` - limited-time event or season reward grants.
* `battle_pass_or_track_reward` - reward-track grants if that system exists later.

## Runtime-Only Effects Excluded From Progression Grants

Normal runtime pickups, powerups, encounter state, boss state, and wave state are not durable progression grant sources by default.

Excluded runtime effects include:

```text
encounter_clear
boss_defeat
wave_clear
normal runtime pickups
normal powerups
temporary weapon drops
softpoint run weapons
hardpoint temporary overwrites
same-weapon ammo increases
dedicated ammunition pickups
health pickups
shield pickups
temporary buffs/debuffs
runtime-only rare drops
```

These may affect the current match, but they do not enter the durable progression grant pipeline unless a later design explicitly promotes a specific event into a persistent reward.

`local_profile_import` is intentionally excluded as a normal grant source.

## GrantAward And Grant Model

A `GrantAward` is one durable award envelope.

A `Grant` is one durable mutation requested inside a `GrantAward`.

Different grant kinds can have different application behavior, but every grant uses the same outer shape so the award can be handed to player-data consistently.

### GrantAward

```text
GrantAward
- award_id
- source_type
- source_id
- player_ref
- eligibility_result
- grants[]
- metadata
```

Field meanings:

* `award_id` - compact deterministic ID for the whole award envelope.
* `source_type` - durable grant source, such as `match_completion` or `achievement_completion`.
* `source_id` - source-specific stable reference, such as match ID, achievement ID, objective ID, receipt ID, or entitlement ID.
* `player_ref` - player identity reference used by player-data routing. Progression does not resolve this into a database route.
* `eligibility_result` - result of progression/trust/policy checks.
* `grants[]` - individual durable grants inside this award.
* `metadata` - optional source context such as `mode_id`, `period_key`, `event_key`, or balancing version.

### Grant

```text
Grant
- grant_id
- grant_kind
- target_ref optional
- amount optional
- metadata optional
```

Planned grant kinds:

```text
xp
earned_currency
unlock
inventory_item
ship_part
rare_drop
entitlement
reversal
```

Possible later grant kinds:

```text
custom_title
cosmetic
reward_track_progress
```

Examples:

```text
Grant
- grant_id: compact deterministic ID
- grant_kind: xp
- amount: 150
```

```text
Grant
- grant_id: compact deterministic ID
- grant_kind: earned_currency
- target_ref: currency.primary
- amount: 25
```

```text
Grant
- grant_id: compact deterministic ID
- grant_kind: unlock
- target_ref: ship.scout
```

```text
Grant
- grant_id: compact deterministic ID
- grant_kind: inventory_item
- target_ref: module.railgun_overcharger
- amount: 1
- metadata:
    rarity: rare
```

## Award And Grant ID Rules

`award_id` and `grant_id` must be stable, deterministic, and compact.

They should not be long concatenated strings of UUIDs.

Planning rule:

```text
award_id = compact deterministic key from:
- source_type
- source_id
- player_ref
```

```text
grant_id = compact deterministic key from:
- award_id
- grant_kind
- target_ref if present
- line discriminator if needed
```

The readable fields remain on the award/grant for audit and debugging. The compact ID exists for dedupe, storage, transaction safety, and retry handling.

Implementation may use a deterministic 128-bit hash or equivalent compact key. Exact encoding is a gametime implementation decision.

## Player-Data Handoff

Progression and rewards emits `GrantAward` records to player-data.

Progression does not directly apply grants or choose a backing store.

Player-data owns route selection and grant application.

## Idempotency Model

Progression must provide stable `award_id` and `grant_id` values.

Player-data persistence owns the actual idempotency storage, receipt storage, transaction storage, and durable write implementation.

Ledger behavior splits by grant class:

```text
Non-currency durable ownership grants
-> long-term unique durable state

Currency grants
-> short-term idempotency receipt
-> atomic balance / transaction application
```

### Non-Currency Durable Ownership Grants

Non-currency durable ownership grants create long-term durable state.

Examples:

```text
unlock
ship
weapon
module
inventory_item
ship_part
custom_title
non-stackable rare item
entitlement ownership
```

Dedupe behavior:

```text
apply once
record durable ownership/unlock/item state
duplicate grant resolves to already-owned / already-applied
```

### Currency Grants

Currency grants mutate a balance and therefore need transaction-style idempotency.

Examples:

```text
earned_currency from match completion
earned_currency from objective completion
earned_currency from reward event
```

Dedupe behavior:

```text
require grant_id receipt
apply balance mutation atomically
duplicate grant_id must not apply balance delta again
```

Short-term idempotency receipts are a player-data implementation detail.

### Stackable Inventory Grants

Stackable inventory grants may need the currency-style receipt path if duplicate quantities are valid.

Unique inventory/unlock grants can usually dedupe against durable ownership state.

## Progression State

Initial durable progression state:

```text
PlayerProgression
- total_xp
- level
- completed_objectives
- unlocked_content
- unlocked_titles
- achievement_state
```

Rules:

```text
total_xp = source of truth
level = stored/cacheable, but derivable from total_xp
rank = derived from level
insignia = derived from level
custom titles = unlockable content
```

XP curve, max level, post-cap XP behavior, and prestige/reset behavior are gametime balancing decisions.

XP will likely continue past max level in some form. Prestige/reset remains possible.

Rank and insignia are tied directly to XP level.

Custom titles are separate unlockable content and may be granted through achievements, milestones, events, rare drops, or other reward sources.

## Objective Completion State

First-time objective completion and periodic objective completion use different state paths.

```text
First-time objective completions
-> durable unique database entry/table

Periodic objective completions
-> timestamped progression state
```

First-time objective completion should be stored as durable unique completion state, keyed by player identity and objective reference.

```text
FirstObjectiveCompletion
- player_ref
- objective_ref
- first_completed_at
- award_id
```

Periodic objective completion should use timestamped progression state so daily, weekly, or other time-window rewards can be evaluated without treating every completion as a permanent first-time unlock.

```text
PeriodicObjectiveCompletionState
- objective_ref
- period_key
- last_completed_at
- completion_count
- last_award_id
```

Example period keys:

```text
daily:2026-06-16
weekly:2026-W25
event:<event_id>:<period>
```

## Unlock And Target References

Unlock grants should use content references rather than hardcoded categories.

Expected `target_ref` values include:

```text
ship.scout
weapon.railgun
module.railgun_overcharger
consumable.repair_charge
title.first_win
mode.score_attack
mission.mars_01
challenge.no_deaths_01
```

Expected unlock targets include:

```text
ships
weapons
modules
modes
missions
challenges
rank-linked insignia
custom titles
cosmetics/colors
future content refs
```

The system should not assume this list is exhaustive.

## Inventory Rewards, Ship Parts, And Rare Drops

Persistent rare drops resolve to inventory items.

Not all rare drops are persistent. Runtime-only rare drops do not enter the durable progression grant pipeline.

## Earned Currency Boundary

Progression and rewards owns earned-currency grant policy.

Commerce and economy owns the currency system itself.

Progression may produce:

```text
Grant
- grant_kind: earned_currency
- target_ref: currency.primary
- amount: <earned amount>
```

Progression does not own:

```text
currency naming
price tables
shop catalog
currency sinks
purchase receipts
refund policy
premium currency policy
commerce balance
```

## Guest Progression Behavior

Guest progression follows existing player-data behavior and uses transient memory until saved into a durable profile through an existing supported flow.

## Eligibility Inputs

Progression reward evaluation consumes eligibility and trust results. It does not own all trust policy.

Likely eligibility inputs:

```text
identity kind
mode
source_type
source_id
trusted result facts
debug/devtools status
completion status
first-time completion status
period key when relevant
trust policy result
```

Game integrity polciy owns which sources are trusted for online progression, which debug/devtools grants are excluded, and which local facts are not online-trusted.

## Reward Formula Policy

Reward formulas are gametime balancing decisions.

Match, objective, mission, challenge, daily-first, weekly-first, event-style, and achievement rewards may use different formulas.

The architecture should support formula variation by:

```text
source_type
source_id
mode_id
period_key
event_key
balancing version
```

This document does not lock exact XP, currency, item, unlock, or reward amounts.

## Current Inputs

* match outcome and result data
* progression eligibility inputs
* trusted source events
* XP inputs
* level and rank derivation inputs
* earned-currency grant inputs
* unlock inputs
* reward grant inputs
* rare persistent drop inputs
* `GrantAward` construction inputs
* idempotent award/grant ID inputs
* player-data handoff inputs

## Planned Outputs

* progression award planning boundaries
* durable grant-source vocabulary
* runtime-only exclusion rules
* `GrantAward` shape
* `Grant` shape
* stable `award_id` / `grant_id` requirements
* progression state planning
* objective completion state planning
* earned-currency grant boundary
* rare-drop inventory boundary
* player-data handoff expectations
* idempotency expectations

## Related Docs

* [Match Outcomes And Results](match-outcomes-and-results.md)
* [Shop, Commerce, And Economy](shop-commerce-and-economy.md)
* [Player Data And Persistence](../../services/player-data/!INDEX.md)
* [Trust And Eligibility Policy](../../../domains/platform/security-and-admin/trust-and-eligibility-policy.md)
* [Game Integrity Policy](../platform/security-and-admin/game-integrity-policy.md)
* [Inventory And Hangar](inventory-and-hangar.md)
* [Player Build And Loadouts](player-build-and-loadouts.md)

## Open Gametime Decisions

* exact XP curve
* exact max-level behavior
* post-cap XP behavior
* prestige/reset policy
* exact reward formulas
* exact earned-currency amounts
* exact daily/weekly period policies
* exact compact ID encoding
* exact physical schema for first-time objective completions
* exact physical schema for periodic completion state
* exact UI reward reveal shape

## Core Invariants

* Progression builds `GrantAward` records.
* Player-data owns identity/storage routing.
* Commerce/economy owns currency system policy.
* Runtime pickups and powerups are not progression grants by default.
* Persistent rare drops resolve to inventory items.
* Currency grants require receipt-backed idempotency and atomic balance application.
* Rank and insignia derive from XP level.
* Custom titles are unlockable content.
* First-time objective completion state belongs to progression.
* Guest progression uses existing transient-memory behavior.
