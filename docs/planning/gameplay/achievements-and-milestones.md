# Achievements And Milestones

## Purpose

This doc plans the achievement and milestone seam for long-lived player goals, one-time recognition, threshold progress, live completion feedback, and reward handoff.

Achievements and milestones consume authoritative domain events.

Achievements and milestones do not own gameplay simulation, match rules, scoring, inventory ownership, currency mutation, reward application, physical persistence routing, or UI layout.

The achievement/milestone system answers:

```text
Did this player complete a named accomplishment?
Did this player cross a milestone threshold?
Should the player receive a live completion notification?
Should a completion produce a reward intent for the progression pipeline?
What durable achievement/milestone state must be stored?
```

## Ownership Boundary

This doc owns planning for:

```text
achievement definitions
milestone definitions
shared achievement/milestone catalog source of truth
definition categories
definition flags
domain-event consumption
achievement and milestone evaluation
completion detection
milestone threshold detection
durable achievement/milestone state
long-lived progress counters when needed
live completion notification outputs
bandwidth-conscious progress notification policy
Redis-backed short-term processing guards
reward-intent handoff into progression
devtools achievement/milestone testing behavior
```

This doc does not own:

```text
domain event emission by gameplay systems
gameplay simulation
scoring policy
match-end rules
match-result summary construction
XP formulas
currency formulas
inventory mutation
unlock mutation
GrantAward construction
player-data route selection
physical database table design
leaderboard eligibility
UI layout
HUD placement
achievement art/audio presentation
analytics/event history storage
```

Owning systems emit domain events.

Achievements and milestones consume those events and interpret them against definitions.

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

Achievements and milestones do not require a separate trusted-fact emitter.

The domain event system is the trusted live input seam.

Evaluator-local facts may be derived from domain events, but those facts are internal projections, not another emitted gameplay stream.

## Domain Event Input Model

Achievements and milestones consume authoritative domain events directly.

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

Not every event needs to exist immediately.

The architecture requirement is that the owning gameplay system emits the event, and achievements consume it later.

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

## Domain Event Identity Requirements

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

## Match Result Relationship

Match results are not the primary source of all achievement progress.

Match-end and match-result systems should emit final domain events such as:

```text
match_completed
match_won
score_finalized
objective_completed
mission_completed
challenge_completed
```

These final events are part of the same domain event model.

They are not a separate trusted-fact pipeline.

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

The shared catalog should generate or load equivalent client and server views.

Server responsibilities:

```text
evaluate authoritative progress
detect completion
update durable state
emit completion/progress outputs
emit reward intents
```

Client responsibilities:

```text
display achievement and milestone catalog data
display names, descriptions, categories, visibility, and thresholds
display live completion notifications
display tracked/pinned progress when provided by server
never authoritatively complete achievements
```

The client should receive compact IDs and progress values from the server and resolve display text locally from the shared catalog.

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
-> long-lived threshold progress track
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

## Definition Categories

Initial categories:

```text
combat
survival
collection
progression
mode
challenge
hidden
```

Later categories:

```text
social
event
seasonal
```

Categories are for organization, display, filtering, and future reward/event grouping.

They should not drive trust or persistence behavior directly.

## Definition Flags

Eligibility and behavior should use flags.

Recommended flags:

```text
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
```

Default behavior:

```text
single-player counts
multiplayer counts
guest counts through transient player-data route
local profile counts through local player-data route
authenticated account counts through online/account route
devtools counts unless or until trust policy blocks it
```

Flags should be positive behavior markers where possible.

Avoid scattering many one-off boolean fields through the definition model.

## Visibility

Visibility should be catalog-owned.

Recommended visibility behavior:

```text
visible
-> shown before completion

hidden
-> slot/category visible, details hidden until completion

secret
-> not shown until completion or discovery
```

Visibility does not determine whether the achievement can progress.

A secret achievement can still evaluate normally from domain events.

## Achievement Definition Shape

Recommended logical shape:

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

Field meanings:

```text
achievement_id
-> stable catalog ID

category
-> organization/display category

visibility
-> visible, hidden, or secret

flags[]
-> eligibility and behavior flags

event_inputs[]
-> domain event types relevant to this achievement

condition
-> rule evaluated against event/current state

reward_intents[]
-> declarative reward handoff, not direct grants

metadata
-> optional display, balancing, or event metadata
```

Example:

```text
AchievementDefinition
- achievement_id: achievement.first_match_completed
- category: progression
- visibility: visible
- flags:
    - counts_single_player
    - counts_multiplayer
    - counts_guest
    - counts_local_profile
    - counts_authenticated_account
    - counts_devtools
- event_inputs:
    - match_completed
- condition:
    type: once
- reward_intents:
    - source_type: achievement_completion
```

Example:

```text
AchievementDefinition
- achievement_id: achievement.first_pickup_collected
- category: collection
- visibility: visible
- flags:
    - counts_single_player
    - counts_multiplayer
    - counts_guest
    - counts_local_profile
    - counts_authenticated_account
    - counts_devtools
- event_inputs:
    - pickup_collected
- condition:
    type: once
```

## Milestone Definition Shape

Recommended logical shape:

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

Field meanings:

```text
milestone_id
-> stable catalog ID

category
-> organization/display category

visibility
-> visible, hidden, or secret

flags[]
-> eligibility and behavior flags

event_inputs[]
-> domain event types that can update the milestone

counter
-> named progress counter updated by relevant events

thresholds[]
-> ordered milestone thresholds

reward_intents[]
-> declarative reward handoff

metadata
-> optional display, balancing, or event metadata
```

Threshold shape:

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
- flags:
    - counts_single_player
    - counts_multiplayer
    - counts_guest
    - counts_local_profile
    - counts_authenticated_account
    - counts_devtools
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

Example:

```text
MilestoneDefinition
- milestone_id: milestone.pickups_collected
- category: collection
- visibility: visible
- flags:
    - counts_single_player
    - counts_multiplayer
    - counts_guest
    - counts_local_profile
    - counts_authenticated_account
    - counts_devtools
- event_inputs:
    - pickup_collected
- counter: pickups_collected_total
- thresholds:
    - threshold_id: 10
      value: 10
    - threshold_id: 100
      value: 100
    - threshold_id: 1000
      value: 1000
```

## Reward Intents

Achievement and milestone definitions may declare reward intents.

Reward intents are not grants.

Reward intents describe that a completion should feed the progression reward pipeline.

Recommended shape:

```text
RewardIntent
- source_type
- reward_ref optional
- metadata optional
```

Examples:

```text
RewardIntent
- source_type: achievement_completion
- reward_ref: reward.first_match_completed
```

```text
RewardIntent
- source_type: milestone_completion
- reward_ref: reward.matches_completed_10
```

Progression and rewards owns turning reward intent into `GrantAward`.

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

Evaluator non-responsibilities:

```text
emit gameplay domain events
mutate gameplay state
decide scoring
decide match outcome
construct GrantAward
apply grants
route player-data writes
choose database/backend
render UI
```

## Progress Storage Rules

Store completion state always.

Store progress only when it is:

```text
long-lived
player-visible
needed for future evaluation
not cheaply derivable
needed for cross-session continuation
```

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

Stats may expose readout summaries, but achievement progress should not depend on scraping UI/profile stats.

## Durable State Shape

Recommended logical state:

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

For ordered milestones, `highest_completed_threshold_id` may be enough.

For non-linear or special milestones, `completed_threshold_ids[]` should be available.

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

Durable storage owns:

```text
completed achievements
completed milestone thresholds
long-lived progress counters
GrantAward receipts
```

Example Redis key concepts:

```text
achievement_event_processed:{player_ref}:{event_id}
achievement_completion_lock:{player_ref}:{achievement_id}
milestone_completion_lock:{player_ref}:{milestone_id}:{threshold_id}
achievement_progress_throttle:{player_ref}:{milestone_id}
```

Exact key naming and TTL values are gametime implementation decisions.

## Completion Outputs

Achievement completion output:

```text
AchievementCompletion
- player_ref
- achievement_id
- source_event_ref
- completed_at
- reward_intents[]
```

Milestone completion output:

```text
MilestoneCompletion
- player_ref
- milestone_id
- threshold_id
- source_event_ref
- completed_at
- reward_intents[]
```

Completion outputs feed two paths:

```text
live notification path
reward handoff path
```

Live notification path sends compact completion information to the client.

Reward handoff path sends reward intent to progression and rewards.

## Live Notification Policy

Players should see achievement and milestone completions live when possible.

Default server-to-client behavior:

```text
achievement completion
-> send live completion notification

milestone threshold completion
-> send live completion notification

background milestone progress
-> do not send every increment by default
```

The client already has the shared catalog, so live packets should be compact.

Recommended completion packet shape:

```text
AchievementCompleted
- achievement_id
- completed_at
- reward_preview_refs optional
```

Recommended milestone packet shape:

```text
MilestoneCompleted
- milestone_id
- threshold_id
- completed_at
- reward_preview_refs optional
```

Reward reveal may be separate later.

## Bandwidth-Conscious Progress Updates

Do not broadcast every counter change.

Plan for tracked or pinned milestones.

Default behavior:

```text
untracked achievement
-> completion notification only

untracked milestone
-> threshold completion notification only

tracked/pinned milestone
-> throttled progress updates

menu/result refresh
-> can load fuller progress state from player-data
```

Progress update packet shape:

```text
MilestoneProgressUpdated
- milestone_id
- current_value
- completed_threshold_ids optional
```

Send progress updates only when:

```text
milestone is tracked/pinned
progress crosses a threshold
progress crosses a display checkpoint
server flushes a throttled update
match ends
client/menu requests refreshed state
```

The server should send IDs and values.

The client resolves display labels, descriptions, thresholds, and visibility from the shared catalog.

## Tracked And Pinned Milestones

The architecture should allow tracked/pinned milestones even if not implemented immediately.

Tracked state may eventually live in player-data or client preference state.

Tracked milestones can receive more frequent progress updates than background milestones.

Recommended limits:

```text
small fixed number of tracked milestones
server-side throttling
completion updates always allowed
progress deltas preferred over full state
```

Exact UI behavior belongs to client/UI planning.

## Reward Handoff

Achievements and milestones do not mutate rewards directly.

Reward flow:

```text
AchievementCompletion or MilestoneCompletion
-> reward_intents[]
-> Progression And Rewards
-> GrantAward construction
-> player-data runtime
-> routed durable application
```

Progression and rewards owns:

```text
GrantAward shape
Grant shape
award_id generation
grant_id generation
reward formula policy
XP grants
currency grants
unlock grants
inventory grants
ship-part grants
rare-drop grants
idempotent award construction
```

Achievements and milestones own only completion detection and reward-intent output.

## Guest Behavior

Guest achievement and milestone behavior should match non-guest behavior during gameplay.

Difference is storage durability.

Guest flow:

```text
Guest identity
-> domain events consumed normally
-> achievement/milestone state stored in transient player-data route
-> live notifications work normally
-> rewards apply to guest transient state where supported
```

If guest state is saved into a new Local Profile through an explicit supported profile creation flow, achievement and milestone state may be copied with other durable-shaped guest state where supported.

Achievements should not invent a separate guest persistence model.

## Local Profile And Authenticated Account Behavior

Local Profile and Authenticated Account should share the same logical achievement/milestone contracts.

Backing storage may differ:

```text
Local Profile
-> local player-data route / SQLite-backed persistence

Authenticated Account
-> Rails/Postgres-backed route

Guest
-> transient memory route
```

Achievements should not choose the storage backend.

Player-data runtime owns route selection.

## Devtools Behavior

Devtools should support achievement and milestone testing.

Supported devtools directions:

```text
inject domain events in dev mode
trigger progress in dev mode
trigger completion in dev mode
exercise reward handoff in dev mode
test live notification packets
test GrantAward handoff
```

Devtools-triggered achievement progress is allowed for testing unless or until trust policy blocks it.

Do not add extra code early just to prevent online/account progress from devtools events if that prevention adds complexity and blocks testing.

Later anti-cheat/trust policy may add stricter rules for ranked, online, or public progression.

## Eligibility And Trust Policy

Achievement definitions should use flags to express intended eligibility.

Anti-cheat and trust policy may later override or restrict progression for:

```text
ranked modes
public multiplayer
devtools events
admin/testing sessions
offline/local-only sessions
modified rooms
event-limited content
```

Initial default:

```text
single-player counts
multiplayer counts
guest counts
local profile counts
authenticated account counts
devtools counts unless blocked later
```

Mode-specific or account-specific restrictions should be definition flags and trust-policy decisions, not hardcoded into gameplay emitters.

## Client Responsibilities

Client owns presentation.

Client should:

```text
load/generated shared achievement catalog
display achievement and milestone names
display descriptions
display visibility states
display categories
display thresholds
display completion notifications
display tracked/pinned milestone progress
request/load full achievement state where needed
```

Client should not:

```text
authoritatively complete achievements
authoritatively increment milestone counters
award currency
award XP
award unlocks
award inventory
decide online trust
```

The client can show local prediction later only if explicitly designed, but server authoritative completion remains the source of truth.

## Server Responsibilities

Server owns authoritative evaluation.

Server should:

```text
consume authoritative domain events
evaluate definitions against current state
apply Redis processing guards where needed
update achievement/milestone state through player-data route
emit compact live notifications
emit reward intents
handoff reward intents to progression and rewards
```

Server should not:

```text
hardcode display strings outside shared catalog
mutate inventory directly from achievement code
mutate currency directly from achievement code
depend on client-side counters
persist every domain event as achievement history
```

## Player-Data Contract Direction

Future logical player-data schema should include achievement and milestone state.

Possible future shared logical schema source:

```text
shared/player_data/achievements.toml
```

Expected logical concepts:

```text
AchievementState
MilestoneState
TrackedMilestonePreference later
CompletedAchievement
CompletedMilestoneThreshold
AchievementProgressCounter
```

Likely future player-data operations:

```text
LoadAchievements(identity)
ApplyAchievementProgress(identity, progress_update)
ApplyAchievementCompletion(identity, completion)
ApplyMilestoneProgress(identity, progress_update)
ApplyMilestoneCompletion(identity, completion)
SaveTrackedMilestones(identity, tracked_refs) later
```

Exact operation names and physical schema are gametime implementation decisions.

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
1. Promote achievements-and-milestones.md out of stubs.
2. Define shared achievement/milestone catalog file.
3. Add minimal data-sync support or direct shared loader for the catalog.
4. Add stable event identity requirements to the domain event seam.
5. Add pure achievement/milestone evaluator.
6. Add transient in-memory state for initial evaluator tests.
7. Add V0 definitions for first-match, first-pickup, first-death, matches-completed, total-score, pickups-collected, and ship-deaths.
8. Consume existing domain events where available.
9. Add final match domain events where needed.
10. Add compact live completion notification packets.
11. Add milestone progress throttling for tracked/pinned progress.
12. Add reward-intent output.
13. Connect reward intent to Progression And Rewards GrantAward construction.
14. Add player-data logical achievement/milestone contracts.
15. Add guest/local/account route parity.
16. Add Redis processing guards where duplicate processing becomes possible.
17. Add devtools event/progress/completion testing hooks.
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
achievement definitions load from shared catalog
client and server catalogs stay in sync
achievement completes from domain event
milestone counter increments from domain event
milestone threshold completes once
duplicate event does not duplicate completion
duplicate completion does not duplicate GrantAward
completion notification is emitted once
background progress does not spam live packets
tracked milestone progress is throttled
untracked milestone sends threshold completion only
hidden achievement visibility works
secret achievement visibility works
guest progress uses transient route
local profile progress uses local route
authenticated account progress uses account route
devtools can trigger progress/completion
definition flags filter eligibility
match_completed domain event can complete achievements
score_finalized domain event can update score milestones
progress state persists only where needed
completion state always persists
```

## Current Inputs

```text
shared achievement/milestone definitions
domain events
domain event identity
player identity
mode identity
definition flags
visibility rules
current achievement state
current milestone state
long-lived progress counters
Redis processing guards
reward intents
devtools event/progress inputs
```

## Planned Outputs

```text
shared achievement/milestone source-of-truth plan
achievement definition shape
milestone definition shape
category vocabulary
flag vocabulary
domain-event consumption boundary
event identity requirements
evaluation runtime boundary
durable state shape
Redis idempotency role
live completion notification policy
bandwidth-conscious progress update policy
tracked/pinned milestone planning
reward-intent handoff shape
guest/local/account behavior
devtools behavior
implementation sequence
testing direction
```

## Related Docs

* [Systems Plan Index](../systems-plan-index.md)
* [Player Experience Systems](player-experience-systems.md)
* [Progression And Rewards](progression-and-rewards.md)
* [Match Outcomes And Results](match-outcomes-and-results.md)
* [Inventory And Hangar](inventory-and-hangar.md)
* [Player Build And Loadouts](player-build-and-loadouts.md)
* [Modes And Match Rules](modes-and-match-rules.md)
* [Player Data And Persistence](../platform/player-data-and-persistence.md)
* [Anti-Cheat And Trust Policy](../platform/anti-cheat-and-trust-policy.md)
* [Source Of Truth Map](../../design/source-of-truth-map.md)
* [Player-Data Schema Source Of Truth](../../design/player-data-schema-ssot.md)

## Open Gametime Decisions

```text
exact TOML schema
exact generated Go catalog shape
exact generated GDScript catalog shape
exact event ID format
exact Redis key naming
exact Redis TTL values
exact progress throttling interval
exact tracked/pinned milestone limit
exact V0 reward intents
exact achievement notification packet fields
exact milestone notification packet fields
exact reward reveal timing
exact player-data logical schema file layout
exact physical persistence schema
exact hidden/secret UI treatment
exact anti-cheat restrictions for public/ranked online progression
exact devtools safeguards if stricter account protection is needed
```

## Core Invariants

```text
Achievements consume domain events.
Milestones consume domain events.
There is no separate trusted-fact emitter for gameplay achievements.
Evaluator-local facts are projections, not another emitted stream.
Gameplay systems do not know about achievement definitions.
Domain events do not decide achievement progress.
Achievements are one-time accomplishments.
Milestones are threshold-based long-lived progress tracks.
Definitions are shared data from the start.
Client displays catalog data and notifications.
Server authoritatively evaluates completion.
Completion state is durable.
Long-lived milestone progress is durable when needed.
High-frequency domain events are not stored long term by achievements.
Redis guards short-term processing, locking, dedupe, and throttling.
Redis is not the achievement database.
Completion notifications should be live.
Background progress should not spam packets.
Tracked/pinned milestones may receive throttled progress updates.
Achievements and milestones emit reward intents only.
Progression builds GrantAward records.
Player-data applies grants idempotently.
Inventory, currency, unlocks, and XP are never directly mutated by achievement code.
Guest behavior matches normal gameplay behavior with transient storage.
Local Profile and Authenticated Account share the same logical achievement/milestone contract.
Devtools can exercise achievement progress and completion unless trust policy later blocks it.
```
