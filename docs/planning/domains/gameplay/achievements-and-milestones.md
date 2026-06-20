# Achievements And Milestones
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc plans the achievement and milestone seam for long-lived player goals, one-time recognition, threshold progress, live completion feedback, and reward handoff.

Achievements and milestones consume authoritative domain events directly.

There is no separate trusted-fact emitter for gameplay achievements. Evaluator-local facts are projections, not another emitted gameplay stream.

## Ownership Boundary

This doc owns:

* achievement definitions
* milestone definitions
* shared achievement/milestone catalog source of truth
* definition categories and flags
* domain-event consumption
* achievement and milestone evaluation
* completion detection
* milestone threshold detection
* durable achievement/milestone state
* long-lived progress counters when needed
* live completion notification outputs
* bandwidth-conscious progress notification policy
* Redis-backed short-term processing guards
* reward-intent handoff into progression
* devtools achievement/milestone testing behavior

This doc does not own:

* domain event emission by gameplay systems
* gameplay simulation
* scoring policy
* match-end rules
* match-result summary construction
* XP formulas
* currency formulas
* inventory mutation
* unlock mutation
* GrantAward construction
* player-data route selection
* physical database table design
* leaderboard eligibility
* UI layout
* HUD placement
* achievement art/audio presentation
* analytics/event history storage

Progression and rewards owns `GrantAward` construction.

Player-data owns identity routing and persistence application.

Inventory and hangar owns durable item ownership state.

Commerce owns shop, currency sinks, purchase receipts, refunds, and entitlement policy.

## Core Architecture

```text
Authoritative gameplay system
-> DomainEvent emitted by owning system
-> achievement/milestone evaluator consumes event
-> Redis short-term guard / dedupe / completion lock where needed
-> achievement or milestone state update
-> live notification output if completion or tracked progress should display
-> reward intent emitted if completion has rewards
-> Progression And Rewards builds GrantAward
-> player-data runtime routes and applies grants idempotently
```

## Domain Event Input Model

Examples of relevant gameplay domain events:

```text
damage_applied
damage_over_time_started
damage_over_time_tick
ship_death
pickup_collected
pickup_effect_applied
pickup_dropped
pickup_expired
bullet_blast
radial_effect_started
match_completed
match_won
score_finalized
objective_completed
mission_completed
challenge_completed
rare_drop_collected
boss_defeated
enemy_destroyed
asteroid_destroyed
```

Not every event needs to exist immediately. The owning gameplay system emits the event, and achievements consume it later.

Good ownership:

```text
damage system emits damage_applied
achievement evaluator consumes damage_applied
```

Bad ownership:

```text
damage system checks achievement definitions
damage system emits achievement_damage_progress
```

Good ownership:

```text
pickup system emits pickup_collected
achievement evaluator decides whether that pickup matters
```

Bad ownership:

```text
pickup system grants achievement completion directly
```

## Domain Event Identity And Match Result Relationship

Domain events consumed by achievements need stable identity for dedupe, replay protection, and completion locking.

Recommended event identity:

```text
event_id
```

or:

```text
match_id + sequence
```

Recommended event envelope fields:

```text
DomainEvent
- event_id or match_id + sequence
- type
- occurred_at
- match_id optional
- mode_id optional
- actor_player_id optional
- target_id optional
- target_type optional
- source_type optional
- source_id optional
- value fields
- metadata optional
```

The domain event system may carry more or fewer fields per event type, but achievements need enough identity and player attribution to process events safely.

Match results are not the primary source of all achievement progress, but they may emit final domain events such as:

```text
match_completed
match_won
score_finalized
objective_completed
mission_completed
challenge_completed
```

These final events are part of the same domain event model, not a separate trusted-fact pipeline.

Live gameplay achievements should be able to complete before match end when their source event occurs.

Examples:

```text
first pickup collected
first ship death
first boss defeated
first rare drop collected
shield broken
survive near death
```

These should not wait for match-result processing unless the achievement specifically depends on final match state.

## Shared Source Of Truth

Achievement and milestone definitions should use shared data immediately.

Planned source:

```text
shared/progression/achievements.toml
```

Possible later split:

```text
shared/progression/achievements.toml
shared/progression/milestones.toml
```

Initial preference is one shared progression catalog unless the file becomes too large or conceptually mixed.

The shared catalog should generate or load equivalent client and server views. Server authoritatively evaluates progress, detects completion, updates durable state, and emits progress/completion outputs and reward intents. Client consumes the shared catalog for display and tracking, and never authoritatively completes achievements.

Likely data-sync direction:

```text
shared/progression/achievements.toml
-> generated Go progression catalog
-> generated GDScript achievement catalog
```

Exact generator implementation is a gametime implementation decision, but definitions should be shared from the start.

## Achievement Versus Milestone

Use two concepts:

```text
Achievement
-> one-time accomplishment

Milestone
-> threshold-based long-lived progress track
```

Do not use “tiered achievement” language.

Threshold-based progress belongs to milestones.

Examples:

```text
achievement.first_match_completed
achievement.first_pickup_collected
achievement.first_boss_defeated

milestone.matches_completed
milestone.total_score
milestone.pickups_collected
milestone.ship_deaths
```

## Definition Model

Categories are for organization, display, and filtering; they do not drive trust.

Eligibility and behavior use flags, and the model should avoid lots of one-off fields.

Visibility is catalog-owned.

Achievements are one-time accomplishments.

Milestones are threshold-based long-lived progress tracks.

Do not use “tiered achievement” language.

Examples omit default flags unless the definition differs from defaults.

### Shared Rules

```text
Initial categories:
combat
survival
collection
progression
mode
challenge
hidden

Later categories:
social
event
seasonal

Recommended flags:
counts_single_player
counts_multiplayer
counts_guest
counts_local_profile
counts_authenticated_account
counts_devtools
hidden
secret
event_limited
ranked_only_later
online_only_later
local_only_later

Default behavior:
single-player counts
multiplayer counts
guest counts through transient player-data route
local profile counts through local player-data route
authenticated account counts through online/account route
devtools counts unless or until trust policy blocks it

Visibility:
visible -> shown before completion
hidden -> slot/category visible, details hidden until completion
secret -> not shown until completion or discovery
```

### AchievementDefinition

```text
AchievementDefinition
- achievement_id
- category
- visibility
- flags[]
- event_inputs[]
- condition
- reward_intents[]
- metadata optional
```

Inline notes:

```text
achievement_id -> stable catalog ID
category -> organization/display category
visibility -> visible, hidden, or secret
flags[] -> eligibility and behavior flags
event_inputs[] -> relevant domain event types
condition -> rule evaluated against event/current state
reward_intents[] -> declarative reward handoff, not direct grants
metadata -> optional display, balancing, or event metadata
```

Example:

```text
AchievementDefinition
- achievement_id: achievement.first_match_completed
- category: progression
- visibility: visible
- event_inputs:
    - match_completed
- condition:
    type: once
- reward_intents:
    - source_type: achievement_completion
```

### MilestoneDefinition

```text
MilestoneDefinition
- milestone_id
- category
- visibility
- flags[]
- event_inputs[]
- counter
- thresholds[]
- reward_intents[]
- metadata optional
```

Inline notes:

```text
milestone_id -> stable catalog ID
category -> organization/display category
visibility -> visible, hidden, or secret
flags[] -> eligibility and behavior flags
event_inputs[] -> domain event types that can update the milestone
counter -> named progress counter updated by relevant events
thresholds[] -> ordered milestone thresholds
reward_intents[] -> declarative reward handoff
metadata -> optional display, balancing, or event metadata
```

### MilestoneThreshold

```text
MilestoneThreshold
- threshold_id
- value
- reward_intents[]
- metadata optional
```

Example:

```text
MilestoneDefinition
- milestone_id: milestone.matches_completed
- category: progression
- visibility: visible
- event_inputs:
    - match_completed
- counter: matches_completed_total
- thresholds:
    - threshold_id: 1
      value: 1
    - threshold_id: 10
      value: 10
    - threshold_id: 50
      value: 50
```

### RewardIntent

Reward intents are not grants.

Progression and rewards owns turning reward intent into `GrantAward`.

```text
RewardIntent
- source_type
- reward_ref optional
- metadata optional
```

Example:

```text
RewardIntent
- source_type: achievement_completion
- reward_ref: reward.first_match_completed
```

Achievements and milestones must not directly apply XP, currency, unlocks, inventory items, titles, or ship parts.

## Evaluation Runtime

The achievement/milestone evaluator consumes domain events and definition data.

Evaluator responsibilities:

```text
load relevant definitions
filter by definition flags and current identity/mode context
project domain event into evaluator-local inputs
read current achievement/milestone state
update progress when needed
detect achievement completion
detect milestone threshold completion
guard duplicate processing where needed
emit live notification outputs
emit reward-intent outputs
```

It does not mutate gameplay, construct `GrantAward`, route player-data writes, or render UI.

## Progress Storage Rules

Store completion state always.

Store progress only when it is long-lived, player-visible, needed for future evaluation, not cheaply derivable, or needed for cross-session continuation.

Do not store every domain event.

Examples:

```text
achievement.first_match_completed
-> store completed state

milestone.matches_completed
-> store long-lived counter and completed thresholds

milestone.total_score
-> store long-lived counter if achievement/milestone progress owns it

achievement.fire_weapon_three_times_in_one_match
-> match-local transient progress only, if it exists later

achievement.win_without_dying
-> may evaluate from match-local state and complete at match end
```

Stats are not the general achievement counter store.

Player-data/progression owns long-lived achievement/milestone counters where needed.

## Durable State Shape

AchievementState and MilestoneState remain the planned logical shapes.

```text
AchievementState
- player_ref
- achievement_id
- status
- completed_at optional
- last_source_event_ref optional
- last_award_ids[]
- metadata optional
```

Status values:

```text
locked
in_progress
completed
```

Recommended milestone state:

```text
MilestoneState
- player_ref
- milestone_id
- current_value
- completed_threshold_ids[]
- highest_completed_threshold_id optional
- last_progress_at optional
- last_source_event_ref optional
- last_award_ids[]
- metadata optional
```

For ordered milestones, `highest_completed_threshold_id` may be enough. For non-linear or special milestones, `completed_threshold_ids[]` should be available.

Exact physical table layout belongs to player-data persistence.

## Redis And Idempotency

Redis is the logical short-term guard for high-frequency event processing.

Redis may own:

```text
event processing dedupe
short-TTL processed event keys
completion locks
award dispatch guards
live progress throttle windows
temporary match/session counters if needed
```

Redis should not become the long-term achievement database.
Durable storage owns completed achievements, completed milestone thresholds, long-lived progress counters, and `GrantAward` receipts.

Example Redis key concepts: `achievement_event_processed:{player_ref}:{event_id}`, `achievement_completion_lock:{player_ref}:{achievement_id}`, `milestone_completion_lock:{player_ref}:{milestone_id}:{threshold_id}`, `achievement_progress_throttle:{player_ref}:{milestone_id}`.

Exact key naming and TTL values are gametime implementation decisions.

## Live Notifications And Bandwidth

Players should see achievement and milestone completions live when possible.

Completion outputs:

```text
AchievementCompletion
- player_ref
- achievement_id
- source_event_ref
- completed_at
- reward_intents[]

MilestoneCompletion
- player_ref
- milestone_id
- threshold_id
- source_event_ref
- completed_at
- reward_intents[]
```

Default behavior:

```text
achievement completion -> send live completion notification
milestone threshold completion -> send live completion notification
untracked background progress -> do not send every increment
tracked/pinned milestone -> throttled progress updates
menu/result refresh -> can load fuller progress state from player-data
```

Progress update packet:

```text
MilestoneProgressUpdated
- milestone_id
- current_value
- completed_threshold_ids optional
```

The server should send compact ID/value packets, and the client resolves names, descriptions, thresholds, and visibility from the shared catalog.

Tracked or pinned milestones may eventually live in player-data or client preference state, and a small fixed number should be throttled with progress deltas preferred over full state.

Reward reveal may be separate later.

## Reward Intent Handoff

Achievements and milestones emit reward intents only, and those intents are not grants.

Progression And Rewards turns reward intents into `GrantAward`, and player-data applies the resulting grants idempotently.

Achievements and milestones do not directly mutate XP, currency, unlocks, inventory, titles, or ship parts.

## Identity, Trust, And Devtools

Guest achievement and milestone behavior matches normal gameplay behavior, with guest state routed through transient player-data.

Local Profile and Authenticated Account share the same logical achievement/milestone contract, while player-data owns route selection and persistence application.

Client consumes shared catalog data and never authoritatively completes achievements or milestones. Server authoritatively evaluates completion, updates durable state, and emits compact notifications and reward intents.

Devtools should support injecting events, triggering progress and completion, exercising reward handoff, and testing notification packets. Devtools-triggered progress is allowed unless or until trust policy blocks it, and early complexity should not be added just to block online/account devtools progress if that makes testing harder.

Achievement definitions should use flags to express intended eligibility. Trust policy may later restrict ranked, public multiplayer, online, devtools, offline/local-only, modified-room, or event-limited progression.

## Initial V0 Set

Recommended V0 achievements:

```text
achievement.first_match_completed
achievement.first_pickup_collected
achievement.first_ship_death
```

Recommended V0 milestones:

```text
milestone.matches_completed
milestone.total_score
milestone.pickups_collected
milestone.ship_deaths
```

Avoid asteroid/enemy/boss milestones until those domain events have clean actor attribution and stable event identity.

Avoid weapon-specialist milestones until weapon usage events and resolved weapon identity are cleanly emitted.

## Implementation Planning

Recommended implementation sequence:

```text
1. Define the shared achievement/milestone catalog.
2. Add data-sync support or a direct shared loader for that catalog.
3. Add stable domain event identity requirements.
4. Add a pure achievement/milestone evaluator.
5. Add transient in-memory state for early evaluator tests.
6. Add V0 definitions for first-match, first-pickup, first-death, matches-completed, total-score, pickups-collected, and ship-deaths.
7. Consume existing domain events where available.
8. Add final match domain events where needed.
9. Add compact live completion notifications.
10. Add tracked/pinned milestone throttling.
11. Add reward-intent output.
12. Connect reward intent to Progression And Rewards GrantAward construction.
13. Add player-data logical achievement/milestone contracts.
14. Add guest/local/account route parity.
15. Add Redis processing guards where duplicate processing becomes possible.
16. Add devtools event/progress/completion hooks.
```

Early slices should favor seams over complete catalog size.

The first implementation should prove:

```text
domain event consumed
achievement completed
client notified live
state prevents duplicate completion
reward intent can be emitted
```

## Testing Direction

Important future tests:

```text
shared catalog
-> load syncs for client and server

domain event completion
-> achievements complete from events
-> milestones increment and complete thresholds from events

duplicate protection
-> events and completions do not double-apply

notifications and throttling
-> completions notify once
-> background progress does not spam live packets
-> tracked milestone progress is throttled
-> untracked milestones only send threshold completion

visibility
-> hidden and secret achievement visibility works

routing and trust
-> guest/local/account routing works
-> devtools can trigger progress/completion
-> definition flags filter eligibility
```

## Related Docs

* [Match Outcomes And Results](match-outcomes-and-results.md)
* [Progression And Rewards](progression-and-rewards.md)
* [Player Data And Persistence](../platform/player-data-and-persistence.md)
* [Anti-Cheat And Trust Policy](../platform/anti-cheat-and-trust-policy.md)

## Open Gametime Decisions

* exact shared catalog file layout
* exact flag vocabulary for future event-specific rules
* exact Redis key naming and TTL values
* exact live notification packet shape for progress throttling
* exact physical schema for achievement and milestone state
* exact tracked/pinned milestone preference storage

## Core Invariants

* Achievements and milestones consume authoritative domain events directly.
* There is no separate trusted-fact emitter for gameplay achievements.
* Evaluator-local facts are projections, not another emitted stream.
* Definitions are shared source of truth from the start.
* Client needs the same catalog data for display and tracking.
* Server authoritatively evaluates completion.
* Achievements are one-time accomplishments.
* Milestones are threshold-based long-lived progress tracks.
* Eligibility and behavior use flags.
* Completion state is durable.
* Long-lived progress is durable when needed.
* High-frequency domain events are not stored long-term by achievements.
* Redis guards short-term event processing, completion locks, award dispatch, and throttling.
* Live completion notifications should happen.
* Untracked background progress should not spam packets.
* Tracked or pinned milestones can receive throttled progress updates.
* Achievements and milestones emit reward intents only.
* Progression and Rewards builds `GrantAward`.
* Devtools can exercise progress and completion unless trust policy later blocks it.
