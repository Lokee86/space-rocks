# Modes And Match Rules

## Purpose

This doc plans the match-rule seam for how a selected mode becomes a resolved set of match rules.

## Ownership Boundary

This doc owns planning for `ModePreset`, `RoomModeConfig`, and `ResolvedMatchRules`, along with the policy pieces they compose.

It should cover objective policy, scoring policy, match-end policy, spawn profile, damage/team policy, progression eligibility, and result policy.

## Current Inputs

- `ModePreset`
- `RoomModeConfig`
- `ResolvedMatchRules`
- objective policy inputs
- scoring policy inputs
- match-end policy inputs
- spawn profile inputs
- damage/team policy inputs
- progression eligibility inputs
- result policy inputs

## Planned Outputs

- mode-to-rules planning boundaries
- rule resolution inputs and outputs
- a shared vocabulary for match-rule policy across gameplay and results planning

## Preset-Driven Room Modes

Modes are preset-driven room and match configurations. Rooms store selected config. Rules resolve config into authoritative match policy. Gameplay consumes resolved rules.

The baseline implementation must prove the seam with two modes.

### Step 1 - Preset-Driven Room Mode Foundation

Step 1 is an implementation plan, not pure foundation. The foundation must be proven through two real baseline modes.

- `ModePreset` = named preset or template for a room or match ruleset
- `RoomModeConfig` = concrete options selected when creating a room
- `ResolvedMatchRules` = server-validated rules consumed by gameplay

Mode is not the same thing as single-player or multiplayer. Single-player and multiplayer are session or hosting context, while mode governs the match rules.

### Preset-Driven Mode Model

- `ModePreset` is the named ruleset template selected by the player.
- `RoomModeConfig` is the selected preset plus allowed player-configurable options.
- `ResolvedMatchRules` is the authoritative server-resolved rules used by the match.

Players configure room options through presets, not arbitrary free-form toggles. Presets define policy-heavy groups.

Preset-owned policy groups:

- Scoring policy
- Match-end policy
- Objective policy
- Spawn profile
- Damage policy
- Team policy
- Progression eligibility
- Result policy
- Difficulty/scaling profile

Likely player-configurable option groups:

- Lives, within preset limits
- Target score, when the preset supports it
- Time limit, when the preset supports it
- Max players, when the preset supports it
- Difficulty tier, when the preset supports it
- Hazards and pickups toggles only when the preset exposes them

Implementation flow:

1. Client selects `ModePreset`
2. Client configures exposed options
3. Client sends requested `RoomModeConfig`
4. Server validates preset and options
5. Room stores validated `RoomModeConfig`
6. Room locks mode config when match starts
7. Rules resolve config into `ResolvedMatchRules`
8. Game simulation consumes `ResolvedMatchRules`
9. Match result includes mode-aware result data

### Two-Mode Baseline

One mode only proves naming, while two modes prove the ruleset seam can vary behavior.

Baseline mode 1: `survival_arcade`

- Describes the current play behavior made explicit.
- `scoring_policy`: current asteroid scoring
- `match_end_policy`: all players eliminated
- `objective_policy`: survive / score freely
- `spawn_policy`: current asteroid spawning
- `lives_policy`: configured lives
- `result_policy`: score + deaths
- Configurable option: lives 1-5

Baseline mode 2: `score_attack`

- Describes the proof mode that uses currently available systems.
- `scoring_policy`: current asteroid scoring
- `match_end_policy`: score target reached OR all players eliminated
- `objective_policy`: reach score target
- `spawn_policy`: current asteroid spawning
- `lives_policy`: configured lives
- `result_policy`: won/lost + score + deaths + target_score
- Configurable options: lives 1-5 and `target_score` from preset-approved values

Score Attack is preferred because it uses existing score, asteroid destruction, lives and death, match-over evaluation, match results, and room lifecycle.

Score Attack does not require enemies, waves, bosses, teams, PvP damage, new pickups, campaign state, progression grants, or new objective entities.

Mission support is preparatory and can be implemented before campaign, while campaign itself remains a late future wrapper over missions.

### Affected Systems

Shared contracts / SSoT:

- Mode preset IDs and option vocabularies become shared client/server language.
- Likely fields include `preset_id`, `lives`, `target_score`, mode summary, and mode identity in match results.

Client room creation / pregame:

- Presents presets.
- Presents allowed options.
- Sends requested `RoomModeConfig`.
- Replaces hardcoded Play Endless behavior with the selected mode config path.

Rooms:

- Store validated `RoomModeConfig`.
- Lock mode config when match starts.
- Expose selected mode summary in room snapshot if needed.
- Pass config into match start.
- Rooms do not define what the mode means.

Game rules / modes:

- Define preset registry.
- Validate config.
- Construct `ResolvedMatchRules`.
- Select match-end, scoring, objective, respawn and lives, damage, team, spawn, progression, and result policies.
- Likely starts near `services/game-server/internal/game/rules`, with exact package split as a gametime decision.

Game simulation / player lifecycle:

- Consumes resolved lives count.
- Consumes match-over and objective rules.
- Should not parse raw room config throughout simulation.

Scoring:

- Reuses current asteroid scoring for both baseline modes.
- Score Attack reads current score as objective progress.
- Scoring package remains policy-focused.

Spawning:

- Both baseline modes use current asteroid spawning.
- Spawn profile support is reserved for later mode presets.

Damage / targeting / collision:

- Current baseline keeps existing damage rules.
- PvP and team damage policy are future affected behavior.

Teams:

- Not implemented in Step 1.
- Must be treated as an affected future system.
- Mode policy must leave room for none, free-for-all, co-op, fixed teams, friendly fire, team spawn rules, team result summaries, and team scoring later.

Match Results:

- Result payload should include mode identity.
- Score Attack should carry `target_score` and success/failure.
- Visible UI can remain small at first.

Player-data / progression:

- Not implemented in Step 1.
- Future progression needs trusted mode-aware results.

Client lobby/session state:

- Room snapshots may expose selected preset, option summary, mode locked state, and display name.

Devtools:

- Future diagnostics should inspect `preset_id`, resolved rules summary, objective state, match-end condition, spawn profile, and scoring policy.

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

### Gametime Decisions

- Exact package split between `game/rules`, `game/rules/modes`, or `game/modes`.
- Exact shared data format for presets.
- Whether first client UI is a full selector or a minimal preset path.
- Exact room snapshot mode-summary shape.
- Exact team policy fields.
- Exact future mission option shape.

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Player Experience Systems](player-experience-systems.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)
- [Progression And Rewards](progression-and-rewards.md)
- [Enemies, Bosses, And Encounters](enemies-bosses-and-encounters.md)

## Open Planning Questions

- Which rule inputs are mode-owned versus shared with room configuration?
- Which parts of scoring and match-end policy are resolved before match start?
- Which progression checks belong in rule resolution versus reward resolution?
- Which gametime decisions should remain open until Step 1 implementation begins?
