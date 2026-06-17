# Levels, Missions, And Content Structure

## Purpose

This doc plans the match-content composition layer for Space Rocks.

It defines how the pre-lobby/create flow assembles a room content configuration from shared catalog refs before a room is created.

The goal is to keep authored content modular by composing refs into `RoomContentConfig`, while leaving behavior, validation, mutation, scoring, objectives, encounters, rewards, and results with their owning systems.

## Ownership Boundary

This doc owns:

* `RoomContentConfig` planning
* mission catalog refs
* challenge catalog refs
* `MatchLevel` catalog refs
* room content preset planning
* arena policy planning
* composition of content refs during room creation
* shared content catalog organization

This doc does not own:

* mode behavior
* mode option semantics
* objective execution
* scoring policy
* match result facts
* reward grants
* build/loadout restrictions
* enemy, boss, wave, asteroid, or spawn behavior
* campaign mechanics
* progression unlock rules

Modes, objectives, scoring policy, match rules, and build restrictions belong to [Modes And Match Rules](modes-and-match-rules.md).

Enemy, boss, asteroid pressure, encounter, spawn, and repeating event internals belong to [Enemies, Bosses, And Encounters](enemies-bosses-and-encounters.md).

Mission, challenge, and objective completion facts belong to [Match Outcomes And Results](match-outcomes-and-results.md).

Durable reward grants, unlock consequences, and content progression state belong to [Progression And Rewards](progression-and-rewards.md).

Build eligibility enforcement belongs to [Player Build And Loadouts](player-build-and-loadouts.md).

## Core Architecture

Room content is configured before room creation.

The flow is:

```text
UI / pre-lobby create phase
-> optional preset fills UI controls
-> UI builds RoomContentConfig
-> create room with RoomContentConfig
-> server validates refs/options through owning systems
-> room stores validated RoomContentConfig
-> players join / select builds / ready
-> match start resolves runtime rules and content
-> gameplay consumes resolved systems
```

`RoomContentConfig` is not a predefined playable catalog object.

It is a configured object built by the UI and passed to room creation.

The room stores validated refs and selected options. Match start resolves those refs through the owning systems.

Gameplay should consume resolved mode rules, objective rules, encounter runtime, spawn runtime, match-level state, and other resolved runtime objects. Gameplay should not read raw room-content config directly.

## RoomContentConfig

`RoomContentConfig` is the main object planned by this doc.

It composes selected refs and options for a room.

Likely shape:

```text
RoomContentConfig
- mode_ref
- selected_mode_options
- objective_refs[]
- mission_ref optional
- challenge_refs[]
- match_level_track_ref optional
- encounter_profile_ref
- spawn_profile_ref
- arena_policy_ref optional
- content_modifier_refs[]
```

Rules:

```text
RoomContentConfig composes refs only.
RoomContentConfig does not define behavior.
Server validates refs/options before storing them on the room.
Match start resolves refs through owning systems.
Gameplay consumes resolved runtime objects only.
```

`RoomContentConfig` may reference modes and objectives, but it does not define what those modes or objectives mean.

Examples:

```text
mode_ref
-> Modes And Match Rules

objective_refs
-> Modes And Match Rules / objective policy

mission_ref
-> Levels, Missions, And Content Structure

challenge_refs
-> Levels, Missions, And Content Structure

match_level_track_ref
-> Levels, Missions, And Content Structure

encounter_profile_ref / spawn_profile_ref
-> Enemies, Bosses, And Encounters

arena_policy_ref
-> unresolved arena policy planning
```

## Room Content Presets

Room content presets are optional.

They exist only to pre-fill UI controls used to build `RoomContentConfig`.

They do not replace `RoomContentConfig`.

They are not authoritative predefined playable entries.

Likely shape:

```text
RoomContentPreset
- preset_id
- default_mode_ref
- default_mode_options
- default_objective_refs[]
- default_mission_ref optional
- default_challenge_refs[]
- default_match_level_track_ref optional
- default_encounter_profile_ref
- default_spawn_profile_ref
- default_arena_policy_ref optional
```

Example:

```text
room_content_preset.arcade_default
-> preselects survival_arcade
-> preselects default mode options
-> preselects default encounter/spawn refs
-> preselects default MatchLevel track if active
```

The UI may apply a preset, modify controls, and then submit the resulting `RoomContentConfig`.

## Arcade, Freeplay, And Endless

Freeplay, endless, and arcade play use the same `RoomContentConfig` path as other match content.

They are not special permanent match-start flows.

The default play path should become a room content configuration assembled from default UI state or an optional room content preset.

Example:

```text
arcade default preset
-> mode_ref: survival_arcade
-> selected default mode options
-> default encounter_profile_ref
-> default spawn_profile_ref
-> default match_level_track_ref if active
```

Arcade/freeplay/endless does not require a mission victory objective.

Mode rules continue to own normal victory, loss, scoring, lives, and match-end behavior.

Challenges may still be active in arcade/freeplay/endless if selected or automatically activated by later policy.

## Missions

A mission is a run whose objectives are victory conditions.

Missions are likely the future home for story or campaign-related content, but non-prep mechanical campaign work is deferred indefinitely.

For now, missions remain non-story and reward-oriented.

Rules:

```text
mission_ref is optional on RoomContentConfig.
Mission objectives become victory conditions.
Missions are differentiated from normal mode victory by mission-owned objective victory content.
Missions may later become unrepeatable once completed.
Mission completion facts belong to Match Outcomes And Results.
Mission reward and unlock consequences belong to Progression And Rewards.
```

Likely shape:

```text
MissionDefinition
- mission_id
- objective_refs[]
- victory_objective_refs[]
- availability_policy_ref optional
- completion_policy_ref optional
- metadata optional
```

A mission does not grant rewards directly.

A mission does not define objective behavior directly.

A mission selects or references objectives that the match-rule system evaluates.

## Challenges

A challenge is an objective overlay.

Challenges may exist inside any `RoomContentConfig`.

A challenge may or may not be selectable. It may be automatic, event-active, mission-attached, mode-attached, hidden, daily, weekly, optional, or manually selected later.

A challenge may or may not be repeatable. It may have cooldowns, timers, period gates, event gates, or one-time completion rules before it can be completed again.

Challenges do not inherently define match victory.

Rules:

```text
Challenges may or may not be selectable.
Challenges may or may not be repeatable.
Challenges may have cooldowns/timers before repeat.
Challenges may be active in any RoomContentConfig.
Challenges do not inherently define match victory.
Challenge completion facts belong to Match Outcomes And Results.
Challenge reward and unlock consequences belong to Progression And Rewards.
```

Working shape:

```text
ChallengeDefinition
- challenge_id
- objective_refs[]
- availability_policy_ref optional
- activation_policy
- repeat_policy
- cooldown_policy_ref optional
- selectable
- auto_activate_policy_ref optional
- completion_policy_ref
- metadata optional
```

Possible activation policies:

```text
selected
automatic
event_active
mission_attached
mode_attached
```

Possible repeat policies:

```text
repeatable
once
cooldown_gated
period_gated
event_gated
```

These field names are planning placeholders and may change before challenge implementation.

## MatchLevels

Use `MatchLevel` for match progression brackets.

A `MatchLevel` is not a map level and not player progression level.

A `MatchLevel` is a score bracket or match progression layer that can activate rule modifier refs and event trigger refs.

Rules:

```text
MatchLevels are triggered during the match.
Once triggered, they persist for the rest of the match.
MatchLevels may activate rule modifier refs.
MatchLevels may activate idempotent event refs.
MatchLevels may activate repeating event refs.
Repeating events own their own internal timing/conditions in the target system.
Scoring policy belongs to Modes And Match Rules.
```

Likely shape:

```text
MatchLevelTrack
- match_level_track_id
- levels[]
```

```text
MatchLevel
- match_level_id
- threshold
- rule_modifier_refs[]
- idempotent_event_refs[]
- repeating_event_refs[]
```

Event categories:

```text
idempotent_event_refs
-> fire once when the MatchLevel is reached

repeating_event_refs
-> become active once the MatchLevel is reached
-> repeat timing/conditions are owned internally by the target system
```

Example:

```text
MatchLevel reached at score 5000
-> idempotent_event_ref: boss_intro_warning
-> repeating_event_ref: meteor_swarm_pressure
```

The MatchLevel system activates the repeating event ref. It does not own meteor swarm timing, spawn cadence, boss behavior, asteroid pressure, or encounter mutation.

Those belong to the target system, likely encounter/spawn planning.

### Multiplayer Score Source

The exact multiplayer score source for MatchLevels is a future decision.

Scoring policy belongs to [Modes And Match Rules](modes-and-match-rules.md).

Possible score sources include:

```text
aggregate match score
individual player score
team score
mode-specific score source
```

MatchLevel should consume the score state selected by mode-owned scoring policy.

The content system should not hardcode aggregate, individual, or team scoring behavior.

## Arena Planning

Arena planning is unresolved and needs a dedicated design pass.

Space Rocks currently uses open toroidal multiplayer space. That makes artificial arena boundaries, orbital zones, spawn anchors, and authored spatial constraints non-trivial.

The plan should not assume that conceptual arena ideas such as orbital lanes, planet-side anchors, or fixed boundaries are implementable until they are proven against the actual toroidal multiplayer model.

Open arena problems:

```text
How do artificial boundaries work in toroidal space?
How does the boundary behave when players split apart?
How do multiplayer players share or separate arena constraints?
How do spawn zones or anchors make sense when space wraps?
How do bosses or events occupy meaningful location?
How are boundaries presented clearly to clients?
How are boundary rules enforced authoritatively by the server?
```

Current decision:

```text
Arena implementation is deferred.
Arena planning remains explicit but unresolved.
Do not commit to complex arena boundaries, orbital lanes, or authored zones yet.
```

`RoomContentConfig` may keep an optional placeholder:

```text
arena_policy_ref optional
```

Early implementation may use a default toroidal arena policy while advanced arena behavior remains deferred.

## Catalog Ownership

Catalogs should be split now and merged later only if the split proves unhelpful.

Recommended ownership paths:

```text
shared/modes/modes.toml
shared/modes/objectives.toml

shared/content/missions.toml
shared/content/challenges.toml
shared/content/match_levels.toml
shared/content/room_content_presets.toml
shared/content/arena_policies.toml
```

Modes and objectives live with the modes/match-rules system because they own rule behavior.

Content catalogs reference mode and objective refs, but do not define their behavior.

The shared-data pipeline should be generic enough that adding new catalogs does not require redesigning the pipeline.

Pipeline goal:

```text
Adding a new shared catalog should not require repeated data-sync pipeline surgery.
```

## Completion And Reward Boundary

This doc does not own rewards.

Mission, challenge, and objective completion facts belong to [Match Outcomes And Results](match-outcomes-and-results.md).

Progression and rewards consumes trusted completion facts and builds durable `GrantAward` records.

Rules:

```text
Missions do not directly grant rewards.
Challenges do not directly grant rewards.
MatchLevels do not directly grant rewards.
Arena policies do not directly grant rewards.
Progression And Rewards owns durable reward and unlock consequences.
```

## Implementation Planning

Initial implementation should prove the room-content configuration seam without implementing campaign systems, advanced arena boundaries, or full challenge mechanics.

Recommended sequence:

```text
1. Add shared catalog loading support for split content catalogs.
2. Add missions catalog shape.
3. Add challenges catalog shape.
4. Add MatchLevel catalog shape.
5. Add room content preset catalog shape.
6. Add arena policy placeholder catalog shape.
7. Add RoomContentConfig request/storage shape.
8. Make the UI/create-room path build RoomContentConfig.
9. Make optional presets pre-fill UI controls only.
10. Validate RoomContentConfig refs server-side.
11. Store validated RoomContentConfig on the room.
12. Resolve RoomContentConfig at match start through owning systems.
13. Preserve current arcade/freeplay behavior through the default room content config.
14. Add MatchLevel runtime state for persistent triggered levels.
15. Add support for idempotent event refs and repeating event refs.
16. Keep repeating event internals owned by the target system.
17. Leave arena behavior at default toroidal policy until a real arena model is designed.
```

Early implementation should favor proving the composition seam over building a large mission or challenge catalog.

## Testing Direction

Important future tests:

```text
RoomContentConfig validates known refs.
RoomContentConfig rejects unknown refs.
Room stores validated RoomContentConfig.
Room content presets only pre-fill UI defaults.
Preset application does not bypass RoomContentConfig validation.
Match start resolves RoomContentConfig once.
Gameplay consumes resolved runtime objects, not raw room content config.
Default arcade/freeplay config preserves current play behavior.
Mission refs can activate victory objective refs.
Challenge refs can be present without defining match victory.
MatchLevel triggers once at threshold.
Triggered MatchLevels persist for the rest of the match.
Idempotent MatchLevel event refs fire once.
Repeating MatchLevel event refs activate once.
Repeating event internals are owned by the target system.
Scoring source policy remains mode-owned.
Arena policy defaults safely to current toroidal behavior.
```

## Related Docs

* [Systems Plan Index](../systems-plan-index.md)
* [Player Experience Systems](player-experience-systems.md)
* [Modes And Match Rules](modes-and-match-rules.md)
* [Enemies, Bosses, And Encounters](enemies-bosses-and-encounters.md)
* [Match Outcomes And Results](match-outcomes-and-results.md)
* [Progression And Rewards](progression-and-rewards.md)
* [Player Build And Loadouts](player-build-and-loadouts.md)

## Open Gametime Decisions

* Exact `RoomContentConfig` field names.
* Exact mission catalog TOML schema.
* Exact challenge catalog TOML schema.
* Exact MatchLevel catalog TOML schema.
* Exact room content preset catalog TOML schema.
* Exact arena policy placeholder schema.
* Exact package placement.
* Exact data-sync generator implementation.
* Exact challenge activation and repeat field names.
* Exact MatchLevel event ref field names.
* Exact multiplayer score source behavior for MatchLevels.
* Exact toroidal multiplayer-compatible arena model.

## Deferred Work

Explicitly deferred:

```text
campaign/story mechanics
content unlock/progression graph
advanced arena implementation
actual challenge implementation details
exact multiplayer MatchLevel score behavior
full mission catalog content
full challenge catalog content
```

## Core Invariants

```text
RoomContentConfig is configured during the pre-lobby/create phase.

RoomContentConfig composes refs; it does not define behavior.

Optional presets only pre-fill UI controls for RoomContentConfig.

Arcade/freeplay/endless is the default room content configuration path.

Missions are runs whose objectives are victory conditions.

Challenges are objective overlays that may be selectable, automatic, repeatable, gated, or attached to other content.

MatchLevels are match-persistent once triggered.

MatchLevels may activate idempotent event refs and repeating event refs.

Repeating event internals are owned by the target system.

Scoring policy belongs to Modes And Match Rules.

Objectives are match rules.

Mission and challenge completion facts belong to Match Outcomes And Results.

Rewards and unlock consequences belong to Progression And Rewards.

Arena implementation is unresolved and deferred for a dedicated toroidal multiplayer design pass.

Catalogs should be split now and supported by a generic enough shared-data pipeline.
```
