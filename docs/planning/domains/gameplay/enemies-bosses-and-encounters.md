# Enemies, Bosses, And Encounters
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc plans the encounter and player-hazard content seam for asteroid pressure, enemy archetypes, enemy AI, enemy loadouts, boss phases, waves, encounter profiles, spawn profiles, and encounter pacing.

The goal is to make hostile and hazardous content authored and sequenced through explicit encounter structures while preserving the server-authoritative runtime model.

Asteroids are included in this plan. They may remain a separate runtime entity type, but they are encounter hazards and must be coordinated with enemy and boss spawning for balance, pacing, and threat pressure.

## Ownership Boundary

This doc owns:

* asteroid pressure planning
* asteroid spawn pressure integration
* enemy archetypes
* enemy AI and behavior profiles
* enemy and boss loadouts
* AI capability compatibility rules
* boss definitions
* boss phases
* wave definitions
* encounter profiles
* spawn profiles
* spawn director behavior
* threat pacing
* encounter events
* open-space and fixed-background encounter constraints
* mission and challenge interaction with encounters

This doc does not own:

* mode rule resolution
* player inventory
* player build eligibility
* durable progression reward formulas
* damage math
* weapon execution internals
* UI layout
* mission/content progression ordering

Mode rules may select `encounter_profile_id`, `spawn_profile_id`, and `arena_profile_id`, but the detailed enemy, boss, asteroid-pressure, wave, and spawn content behind those IDs belongs here.

Mission and challenge planning may also select encounter profiles, but mission progression structure belongs to the levels, missions, and content structure plan.

Progression and rewards may consume trusted encounter outcomes later, but raw wave, boss, and enemy events are not durable reward grants by default.

## Current Code Notes

The current server code already contains partial enemy hooks:

```text
EntityStore.Enemies exists.
target_kind = enemy exists.
damage.EntityTypeEnemy exists.
Radial effects already support enemy targets.
```

The current enemy map appears to reuse ship-shaped runtime data:

```text
Enemies map[string]*Ship
```

That should be treated as a current implementation shortcut or partial seam, not the intended long-term model.

The planned endpoint is:

```text
runtime.Ship
-> player-controlled ship/avatar

runtime.Enemy
-> hostile AI-controlled entity
```

Enemies and ships may share lower-level concepts such as position, velocity, rotation, health, shield, damage modifiers, weapon state, and collision body. Enemies should not inherit player-specific assumptions such as player input, player config, player lifecycle, player inventory, or player build ownership unless those concepts are deliberately abstracted into shared runtime components.

## Core Architecture

Encounter selection flow:

```text
Mode / Mission / Challenge selection
-> encounter_profile_id + spawn_profile_id + arena_profile_id
-> match start resolves EncounterRuntime
-> EncounterDirector steps during Game.Step
-> director coordinates asteroid pressure, enemy waves, and boss phases
-> director emits spawn intents and encounter events
-> Game applies spawn intents through authoritative runtime seams
-> enemies and bosses act through AI + loadout + weapons + damage seams
-> encounter observes wave, boss, and clear conditions
-> match rules/objectives decide whether the match ends
```

The encounter director decides what should happen. Game remains the owner of authoritative runtime mutation.

Encounter logic should not directly mutate broad runtime state outside controlled spawn, phase, and encounter-state seams.

## Encounter Hazard Content

Encounter hazard content includes:

```text
asteroids
enemies
bosses
future environmental hazards
```

Asteroids are not AI enemies, but they are player hazards. They affect encounter pressure, weapon economy, pickup economy, objective difficulty, and threat pacing.

Runtime entity types may remain distinct:

```text
runtime.Asteroid
runtime.Enemy
runtime.Projectile
runtime.Pickup
runtime.Ship
```

Conceptually:

```text
Asteroid
-> non-AI hazard entity

Enemy
-> AI hostile entity

Boss
-> boss-class AI hostile entity
```

Core rule:

```text
Asteroids are separate runtime entities, but they are encounter hazards.

Encounter planning owns asteroid spawn pressure alongside enemy and boss sequencing.
```

## Profile Resolution

Resolved match rules may carry encounter references:

```text
ResolvedMatchRules
- encounter_profile_id
- spawn_profile_id
- arena_profile_id
- difficulty_tier
```

Mode rules own selecting or validating those IDs. This doc owns what the IDs mean.

Gameplay should consume resolved encounter runtime data, not raw room config.

A missing or unknown profile must either fail safely or use an explicit fallback. Silent partial encounter startup should be avoided.

## Encounter, Spawn, And Arena Profiles

Use three separate profile concepts:

```text
EncounterProfile
-> authored sequence and pressure plan

SpawnProfile
-> placement, cadence, caps, and pressure budget

ArenaProfile
-> bounds, background, spawn zones, and encounter anchors
```

### EncounterProfile

An encounter profile owns authored hostile-content structure.

Likely fields:

```text
EncounterProfile
- encounter_profile_id
- stages / waves
- asteroid pressure rules
- enemy spawn groups
- boss entries
- clear/completion rules
- threat pacing refs
```

Encounter profiles answer:

```text
What content appears?
What sequence does it follow?
What must be cleared?
When does a boss appear?
How much asteroid pressure belongs to this encounter?
```

### SpawnProfile

A spawn profile owns placement and pacing mechanics.

Likely fields:

```text
SpawnProfile
- asteroid spawn rules
- enemy spawn rules
- boss spawn rules
- spawn zones
- spawn cadence
- max active asteroids
- max active enemies
- max active total threat
- threat budget
- player-distance constraints
- spawn source weighting
```

Spawn profiles answer:

```text
Where can content appear?
How often can it appear?
How many entities can be active?
How much total threat is allowed?
How far from players should entities spawn?
```

### ArenaProfile

An arena profile gives open-space encounters enough structure without requiring full map geometry.

Likely fields:

```text
ArenaProfile
- arena_id
- world bounds
- background_id
- spawn_zones
- encounter_anchors
```

The fixed planet/background can remain visual at first. Named zones and anchors should still exist so later content can author spawns such as planet-side entry, rear arc reinforcements, or boss entry lanes without inventing a full level system.

## Asteroid Pressure And Spawning

Current asteroid spawning should move behind encounter/spawn profile control or be wrapped by an encounter-owned adapter.

The first implementation does not need to remove asteroid-specific runtime code. It does need to make asteroid pressure part of encounter pacing.

Asteroid pressure should support:

```text
active asteroid caps
spawn cadence
variant/drop-table compatibility
spawn zone selection
difficulty/threat scaling
match-over spawn stop
coordination with enemy and boss counts
```

Asteroid variants remain owned by the asteroid variant contract. Drop-table evaluation remains owned by the drop-table seam. Encounter planning controls when and how much asteroid pressure appears, not the internal asteroid variant metadata or drop evaluation logic.

## Enemy Runtime Model

Enemies should be real server-authoritative runtime entities.

Likely shape:

```text
Enemy
- id
- enemy_type
- enemy_class
- behavior_profile_id
- loadout_id
- current_phase_id optional
- x
- y
- rotation
- velocity
- health
- shield later
- collision_shape_id
- weapon_state
- behavior_state
- encounter_ref optional
- pending_despawn
- despawn_delay
```

Enemy class values:

```text
minion
elite
boss
```

Bosses are boss-class enemies with boss definition and phase state. They should not become a completely separate top-level runtime entity unless a later mechanic requires that split.

Targeting should continue to use:

```text
target_kind = enemy
```

Boss identity can be carried by enemy class, `boss_id`, or boss metadata.

## Enemy Archetypes

Enemy archetypes define content identity and connect runtime stats, behavior, loadout, collision, threat, scoring, drops, and presentation.

Likely shape:

```text
EnemyArchetype
- enemy_type
- enemy_class
- behavior_profile_id
- loadout_id
- health
- shield later
- collision_shape_id
- movement/presentation metadata
- threat_cost
- score_value
- drop_table optional
- presentation_id
```

The first implementation must include two real enemy archetypes. One enemy archetype only proves naming and risks becoming a special-case hostile asteroid.

### V0 Archetype: drone_ram

```text
enemy.drone_ram
- class: minion
- behavior: close_and_ram
- required AI capability: ram
- loadout: enemy_ram
- movement: seek nearest active player
- attack trigger: collision/overlap
```

Purpose:

```text
Proves movement, targeting, ramming weapon, collision, damage, destruction, and enemy events.
```

### V0 Archetype: gunner_standoff

```text
enemy.gunner_standoff
- class: minion
- behavior: standoff_gunner
- required AI capability: ranged_fire
- loadout: basic_cannon with enemy tuning
- movement: keep medium range from target
- attack trigger: target valid + attack policy allows + weapon fire succeeds
```

Purpose:

```text
Proves ranged AI, enemy use of the shared weapon system, context-specific weapon tuning, and archetype variation.
```

## AI And Behavior Profiles

AI behavior owns decision-making. It does not own damage values, projectile construction, weapon cooldowns, or runtime mutation.

AI owns:

```text
target selection
movement decision
attack intent selection
behavior specificity / AI level
boss pattern behavior
```

Likely shape:

```text
EnemyBehaviorProfile
- target_selection_policy
- movement_policy
- attack_selection_policy
- ai_level / specificity
```

AI may inspect loadout, weapon, cooldown, range, and AI capability data when making decisions. Weapon data may inform behavior, but weapon data does not author behavior by itself.

Correct relationship:

```text
AI chooses what to try.
Loadout defines what the enemy can do.
Weapons/capabilities define how attacks execute.
Damage defines how harm resolves.
Game applies runtime mutation.
```

Movement is selected by the behavior profile.

Allowed:

```text
AI reads weapon range/cooldown/capability
-> decides whether to advance, hold, retreat, fire, or switch attacks
```

Not desired:

```text
weapon profile says desired_range = medium
-> movement automatically becomes standoff
```

Better:

```text
behavior_profile: standoff_gunner
movement_policy: maintain_medium_range
attack_policy: fire_when_useful
loadout: basic_cannon with enemy tuning
```

### AI Levels

Different AI profiles may choose targets and attacks differently.

Planning categories:

```text
low_ai
- nearest valid target
- simple movement
- first usable attack
- basic cooldown/range checks

medium_ai
- target by range, threat, health, or simple priority
- choose weapon based on situation
- reposition more intentionally

boss_ai
- set attack patterns
- phase-driven behavior
- possible randomization inside defined patterns
```

Bosses will likely use authored attack patterns instead of generic enemy weapon selection.

## Enemy And Boss Loadouts

Enemies and bosses use loadouts, but these are not player inventory loadouts.

Likely shape:

```text
EnemyLoadout
- weapon/capability entries
- context-specific weapon tuning fields
- ammo policy
- optional phase overrides
```

Enemy behavior and enemy loadout must validate together.

Example:

```text
behavior_profile: close_and_ram
requires capability: ram

loadout:
- enemy_ram capability
```

A ramming behavior with no ramming weapon is invalid.

Boss phases may reload loadout when specified. Phase reload resets cooldown and ammo by default.

Boss weapons will likely use infinite ammo policies. Preserve cooldown or ammo across phases only if a later phase explicitly requires that behavior.

## AI Capability Tags

AI capability tags describe what an AI behavior can execute with a loadout entry.

They do not need to mirror weapon tags.

They do not need to expose damage/effect contracts like `area`, `radial`, or `over_time` unless AI behavior actually needs to select or validate against those distinctions.

Initial AI capability tags should stay narrow:

```text
ram
ranged_fire
```

Likely later AI capability tags:

```text
burst_fire
pattern_fire
self_cast
summon_or_spawn
```

Avoid damage/effect tags as AI capability tags unless they become behavior-relevant:

```text
area_damage
over_time
radial_effect
```

Rule:

```text
Behavior profiles require AI capability tags.
Loadout entries provide AI capability tags.
Weapon/effect/damage metadata may exist under the weapon system, but only becomes an AI capability tag when behavior needs to select or validate against it.
```

Examples:

```text
drone_ram
- behavior: close_and_ram
- requires capability: ram
- loadout provides: ram
```

```text
gunner_standoff
- behavior: standoff_gunner
- requires capability: ranged_fire
- loadout provides: ranged_fire
```

## Weapon And Capability Integration

Enemies and bosses use the same weapon system and weapon catalog as players.

Weapon IDs identify conceptual weapons. Do not create separate enemy or boss weapon IDs unless the weapon is conceptually different.

Preferred tuning model:

```text
weapon_id: basic_cannon
player tuning fields
enemy tuning fields
boss tuning fields
```

This avoids unnecessary duplication such as:

```text
basic_cannon
enemy_basic_cannon
boss_basic_cannon
```

unless those are genuinely different weapons.

Projectile ownership should use `source_id`. Source kind can be derived from runtime lookup/context when necessary. Do not add `source_kind` or `owning_player_id` unless lookup ambiguity becomes a real problem.

### Ramming And Non-Projectile Weapons

Ramming and other non-projectile attacks must be added to the weapon seam.

Ramming is not enemy-owned contact damage.

Correct model:

```text
ramming behavior
-> requires ramming weapon/capability in loadout
-> collision adapter triggers the ramming weapon
-> weapon/capability provides damage spec and timing rules
-> damage seam resolves the damage
```

V0 decision:

```text
Implement ramming as a collision-triggered weapon capability inside the weapons seam.

Any weapon-system refactor required to support this belongs to the implementation slice.
```

## Faction And Damage Eligibility

Faction or allegiance policy gates legal targeting, collision consequences, and damage eligibility.

Exact package placement is a gametime decision because the game package needs refactoring.

Initial required rules:

```text
players damage enemies
enemies damage players
enemies do not damage enemies by default
enemy projectiles do not damage enemies by default
player projectiles do not damage players by default
```

AI target selection may reference faction policy, but damage and collision legality should not be purely AI-owned.

Faction policy should be close enough to collision/damage adapters that illegal harm is blocked even if AI selection or projectile source data is malformed.

## Boss Definitions And Phases

Bosses are boss-class enemy entities with boss definitions and phase state.

Likely shape:

```text
BossDefinition
- boss_id
- enemy_type
- phases[]
```

```text
BossPhase
- phase_id
- trigger
- behavior_profile_id optional
- loadout_id optional
- tuning overrides optional
- attack_pattern_id optional
- spawn_group_refs optional
```

Initial phase triggers:

```text
on_spawn
health_percent_below
timer_elapsed
```

Boss phases may reload AI and loadout when specified.

Phase transition flow:

```text
phase trigger fires
-> boss phase changes
-> reload behavior if specified
-> reload loadout if specified
-> reset cooldown/ammo by default
-> emit boss_phase_started
```

Boss phases should support authored attack patterns. Randomization may be allowed inside a phase pattern later, but the pattern itself should remain phase-owned.

## Fake Boss Test Seam

A fake or test boss may be added to prove the boss implementation seams.

The fake boss is not player-facing content, not balance content, and not progression-eligible.

It should prove the entire boss seam, not only a phase flag.

Required proof:

```text
boss spawned
boss identity exists
boss health threshold triggers phase transition
phase transition reloads AI if specified
phase transition reloads loadout if specified
cooldown/ammo resets on reload
boss_phase_started event emits
boss can fire/use phase loadout
boss can take damage
boss can be defeated
boss_defeated event emits
encounter director can observe boss defeat
packet/client presentation plumbing works
```

Fake boss phase 2 should reload both AI and loadout, even if the actual behavior is simple.

## Encounter Director

The encounter director decides what should spawn and when.

```text
EncounterRuntime
+ EncounterDirector.Step(...)
-> SpawnIntent[]
-> EncounterEvent[]
```

Spawn intents may include:

```text
asteroid_spawn_intent
enemy_spawn_intent
boss_spawn_intent
```

The director should respect:

```text
encounter profile
spawn profile
arena profile
active asteroid count
active enemy count
active total threat
wave state
boss state
match-over state
```

The director should not spawn during match over.

Game applies returned spawn intents through existing authoritative runtime seams.

## Domain Events

Planned encounter-related domain events:

```text
encounter_started
enemy_spawned
enemy_destroyed
wave_started
wave_cleared
boss_spawned
boss_phase_started
boss_defeated
encounter_completed
```

Achievements and milestones can consume these later.

Progression should not grant durable rewards directly from raw wave, enemy, or boss events unless match, objective, mission, or challenge policy promotes them into durable reward sources.

## Base Implementation Plan

V0 should prove asteroid pressure, two enemy archetypes, enemy weapon use, ramming, and the fake boss seam.

Implementation direction:

```text
1. Keep asteroid runtime entity separate.
2. Move or wrap timed asteroid spawning behind encounter/spawn profile control.
3. Add encounter profile fields for asteroid pressure.
4. Add spawn profile caps for asteroids, enemies, and total threat.
5. Add or formalize runtime.Enemy.
6. Add enemy packet/client sync.
7. Add EnemyArchetype catalog with drone_ram and gunner_standoff.
8. Add EnemyBehaviorProfile support.
9. Add EnemyLoadout support.
10. Add AI capability tags for behavior/loadout validation.
11. Add non-projectile ramming weapon support to the weapon seam.
12. Add enemy use of the existing projectile weapon fire path.
13. Add enemy projectile source_id attribution.
14. Add faction/allegiance legality checks where collision/damage adapters need them.
15. Add enemy movement policies for seek target and standoff.
16. Add nearest-active-player enemy target policy.
17. Add projectile/enemy and enemy/player collision handling.
18. Add enemy damage/destruction through the damage seam.
19. Emit enemy_spawned and enemy_destroyed.
20. Add EncounterDirector profile that can coordinate asteroid, enemy, and boss spawning.
21. Add fake/test boss with two phases.
22. Make fake boss phase 2 reload AI and loadout.
23. Emit boss_spawned, boss_phase_started, and boss_defeated.
24. Exclude fake boss from progression/content balance.
25. Ensure match-over stops all encounter-driven spawning.
```

Early slices should favor proving seams over building a large content catalog.

## Testing Direction

Important future tests:

```text
archetype/loadout validation rejects missing required AI capability
drone_ram requires ram capability
gunner_standoff requires ranged_fire capability
enemy uses weapon_id with enemy tuning
enemy projectile source_id is set correctly
enemy projectile can damage player
enemy projectile does not damage enemy by default
player projectile can damage enemy
ramming weapon triggers from collision
ramming damage routes through the damage seam
director coordinates asteroid pressure
director spawns both V0 archetypes
director respects max active asteroids
director respects max active enemies
director respects max active total threat
director does not spawn during match over
wave starts and clears
fake boss spawns
fake boss phase transition triggers at health threshold
phase transition reloads AI/loadout
phase reload resets cooldown/ammo
boss_phase_started emits once
boss_defeated emits once
encounter director observes boss defeat
unknown encounter_profile_id fails safely or falls back explicitly
```

## Related Docs

* [Planning](../../!INDEX.md)
* [Player Experience Systems](player-experience-systems.md)
* [Modes And Match Rules](modes-and-match-rules.md)
* [Levels, Missions, And Content Structure](levels-missions-and-content-structure.md)
* [Player Build And Loadouts](player-build-and-loadouts.md)
* [Progression And Rewards](progression-and-rewards.md)
* [Progression And Rewards](progression-and-rewards.md)
* [Damage](../../../systems-design/combat/damage.md)
* [Weapons](../../../systems-design/combat/weapons.md)
* [Asteroid Variant Contract](../../../protocol/asteroid-variant-contract.md)
* [Drop Table System](../../../data/drop-tables.md)

## Open Gametime Decisions

* Exact package split for encounter runtime, enemy runtime, AI behavior, and spawn director.
* Exact shared data format for enemy archetypes, behavior profiles, loadouts, boss definitions, encounter profiles, spawn profiles, and arena profiles.
* Exact enemy packet shape and client presentation mapping.
* Exact collision shape catalog structure for enemy and boss bodies.
* Exact faction/allegiance policy package placement.
* Exact fallback behavior for missing encounter, spawn, arena, enemy, or boss profile IDs.
* Exact boss phase state packet shape.
* Exact fake boss presentation.
* Exact asteroid pressure formula and threat-cost values.
* Exact AI capability tag vocabulary beyond the V0 tags.

## Core Invariants

```text
Asteroids are separate runtime entities, but they are encounter hazards.

Encounter planning owns asteroid spawn pressure alongside enemy and boss sequencing.

Enemies and bosses are server-authoritative runtime entities.

runtime.Enemy is the planned endpoint for AI-controlled hostile entities.

Enemies use AI behavior profiles, enemy loadouts, the shared weapon system, faction legality, collision, and damage seams.

Enemy behavior chooses target, movement, and attack intent.

AI may inspect loadout, weapon, cooldown, range, and capability data when making decisions.

Loadout defines what AI capabilities are available.

AI capability tags define behavior/loadout compatibility.

Weapon IDs identify conceptual weapons.

Context-specific tuning fields allow player, enemy, and boss tuning without duplicating weapon IDs.

Ramming is a collision-triggered weapon capability in the weapon seam.

Faction policy gates legal targeting, collision consequences, and damage eligibility.

Damage resolves harm.

Game applies runtime mutation.

Boss phases may reload AI and loadout.

Boss phase reload resets cooldown/ammo by default.

The fake boss exists only to prove the boss seam, not to define real boss content.

Match-over stops all encounter-driven spawning.
```
