# Modes And Match Rules
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc plans the match-rule seam for turning a selected mode into a resolved set of authoritative match rules.

It preserves the current gameplay direction while making the room-mode boundary explicit enough for `ModePreset`, `RoomModeConfig`, and `ResolvedMatchRules` to stay separate.

## Ownership Boundary

This doc owns planning for `ModePreset`, `RoomModeConfig`, `ResolvedMatchRules`, and the policy pieces they compose.

It covers:

```text
objective policy
scoring policy
match-end policy
progression eligibility
result policy
room-mode option validation
room-mode storage
match-start rule resolution
spawn profile ID selection
encounter profile ID selection
```

This doc may select spawn profile IDs or encounter profile IDs, but it does not own detailed enemy, wave, or level content behind those IDs.

## Core Architecture

`ModePreset` is the named preset or template for a room or match ruleset.

`RoomModeConfig` is the concrete options selected when creating a room.

`ResolvedMatchRules` is the server-validated rules object consumed by gameplay.

The flow is:

```text
ModePreset
-> allowed room options
-> requested RoomModeConfig
-> server validation
-> stored RoomModeConfig
-> match-start resolution
-> ResolvedMatchRules
-> game simulation
-> match result
```

Rooms store the validated config, then lock that config when the match starts.

Gameplay consumes only `ResolvedMatchRules`, not raw room config.

## Preset-Driven Room Modes

Modes are preset-driven room and match configurations.

Rooms store selected config.

Rules resolve config into authoritative match policy.

Gameplay consumes resolved rules.

The baseline implementation must prove the seam with two real modes, not one naming example.

### Step 1 - Preset-Driven Room Mode Foundation

Step 1 is an implementation plan, not pure foundation. The seam must be proven through two baseline modes.

Players configure room options through presets, not arbitrary free-form toggles. Presets define policy-heavy groups.

Preset-owned policy groups:

```text
scoring policy
match-end policy
objective policy
spawn policy
damage policy
team policy
progression eligibility
result policy
difficulty / scaling profile
```

Likely player-configurable option groups:

```text
lives, within preset limits
target_score, when the preset supports it
time limit, when the preset supports it
max players, when the preset supports it
difficulty tier, when the preset supports it
hazards and pickups toggles only when the preset exposes them
```

Mode is not the same thing as single-player or multiplayer. Single-player and multiplayer are session or hosting context, while mode governs the match rules.

One mode only proves naming, while two modes prove the ruleset seam can vary behavior.

Baseline mode 1: `survival_arcade`

```text
scoring_policy: current asteroid scoring
match_end_policy: all players eliminated
objective_policy: survive / score freely
spawn_policy: current asteroid spawning
lives_policy: configured lives
result_policy: score + deaths
configurable option: lives 1-5
```

Baseline mode 2: `score_attack`

```text
scoring_policy: current asteroid scoring
match_end_policy: score target reached OR all players eliminated
objective_policy: reach score target
spawn_policy: current asteroid spawning
lives_policy: configured lives
result_policy: won/lost + score + deaths + target_score
configurable options: lives 1-5 and target_score from preset-approved values
```

`score_attack` is preferred because it uses existing score, asteroid destruction, lives and death, match-over evaluation, match results, and room lifecycle.

`score_attack` does not require enemies, waves, bosses, new pickups, campaign state, progression grants, or new objective entities.

Mission support is preparatory and can be implemented before campaign, while campaign itself remains a late future wrapper over missions.

### Affected Systems

Shared contracts / SSoT:

```text
Mode preset IDs and option vocabularies become shared client/server language.
Likely fields include `preset_id`, `lives`, `target_score`, mode summary, and mode identity in match results.
```

Client room creation / pregame:

```text
Presents presets.
Presents allowed options.
Sends requested `RoomModeConfig`.
Replaces hardcoded Play Endless behavior with the selected mode config path.
```

Rooms:

```text
Store validated `RoomModeConfig`.
Lock mode config when match starts.
Expose selected mode summary in room snapshot if needed.
Pass config into match start.
Rooms do not define what the mode means.
```

Game rules / modes:

```text
Define preset registry.
Validate config.
Construct `ResolvedMatchRules`.
Select match-end, scoring, objective, respawn and lives, damage, team, spawn, progression, and result policies.
Likely starts near `services/game-server/internal/game/rules`, with exact package split as a gametime decision.
```

Game simulation / player lifecycle:

```text
Consumes resolved lives count.
Consumes match-over and objective rules.
Should not parse raw room config throughout simulation.
```

Scoring:

```text
Reuses current asteroid scoring for both baseline modes.
Score Attack reads current score as objective progress.
Scoring package remains policy-focused.
```

Spawning:

```text
Both baseline modes use current asteroid spawning.
Spawn profile support is reserved for later mode presets.
```

Damage / targeting / collision:

```text
Current baseline keeps existing damage rules.
PvP and team damage policy are future affected behavior.
```

Teams:

```text
Not implemented in Step 1.
Must be treated as an affected future system.
Mode policy must leave room for none, free-for-all, co-op, fixed teams, friendly fire, team spawn rules, team result summaries, and team scoring later.
```

Match Results:

```text
Result payload should include mode identity.
Score Attack should carry `target_score` and success/failure.
Visible UI can remain small at first.
```

Player-data / progression:

```text
Not implemented in Step 1.
Future progression needs trusted mode-aware results.
```

Client lobby/session state:

```text
Room snapshots may expose selected preset, option summary, mode locked state, and display name.
```

Devtools:

```text
Future diagnostics should inspect `preset_id`, resolved rules summary, objective state, match-end condition, spawn profile, and scoring policy.
```

## Implementation Planning

Step 1 is the baseline proof for the mode seam.

The plan is to keep the current play flow intact under `survival_arcade`, then add `score_attack` as the second explicit proof mode.

### Step 1 Completion Criteria

- `survival_arcade` exists as an explicit preset.
- `score_attack` exists as a second explicit preset.
- `CreateRoomRequest` or an equivalent room creation path can carry selected mode config.
- Server validates requested preset and options.
- Room stores validated `RoomModeConfig`.
- Room snapshot exposes selected mode summary if needed by client or lobby.
- Match start resolves `ResolvedMatchRules`.
- Configured lives affect both modes.
- `target_score` affects only Score Attack.
- Survival Arcade ends on elimination.
- Score Attack ends on target score or elimination.
- Match result includes mode identity.
- Score Attack result includes `target_score` and success or failure.
- Existing current play flow still works through Survival Arcade.
- Current multiplayer create and start flow is not broken by mode config.

## Testing Direction

The main checks for this seam are:

```text
room creation rejects invalid preset or option combinations
room storage preserves validated RoomModeConfig
match start resolves ResolvedMatchRules once and uses those rules in gameplay
survival_arcade preserves the current play behavior
score_attack ends on target score or elimination
configured lives affect both baseline modes
target_score affects only score_attack
match results carry mode identity and score_attack success or failure
room snapshots expose the mode summary only if the client needs it
```

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Player Experience Systems](player-experience-systems.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)
- [Progression And Rewards](progression-and-rewards.md)

## Open Gametime Decisions

- Exact package split between `game/rules`, `game/rules/modes`, or `game/modes`.
- Exact shared data format for presets.
- Whether first client UI is a full selector or a minimal preset path.
- Exact room snapshot mode-summary shape.
- Exact team policy fields.
- Exact future mission option shape.
- Exact spawn profile ID vocabulary.
- Exact encounter profile ID vocabulary.

## Core Invariants

```text
ModePreset names the preset.
RoomModeConfig carries the validated room options.
ResolvedMatchRules is the only rule object consumed by gameplay.
Rooms store validated config, not raw free-form mode flags.
Match-start rule resolution is the seam where config becomes authoritative rules.
survival_arcade remains the explicit current-play baseline.
score_attack uses the same baseline scoring and spawning behavior, but adds a target-score end condition.
lives apply to both baseline modes.
target_score applies only to score_attack.
No gameplay system should read raw room config directly.
```
