---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-79e9-b323-c2278366b7a5
document_type: general
policy_exempt: false
summary: This doc plans the match-rule seam for turning a selected mode into a resolved set of authoritative match rules.
---
# Modes And Match Rules
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc plans the match-rule seam for turning a selected mode into a resolved set of authoritative match rules.

It preserves the current gameplay direction while making the room-mode boundary explicit enough for `ModePreset`, `RoomModeConfig`, and `ResolvedMatchRules` to stay separate.

## Ownership Boundary

This doc owns planning for `ModePreset`, `RoomModeConfig`, `ResolvedMatchRules`, and the policy pieces they compose.

It covers:

```text
gameplay award policy
objective policy selection and composition
ranking policy
match-end policy
result policy
team-system selection and mode overrides
damage policy
lives/respawn policy
player spawn profile
Encounter Spawn Profile
join policy
progression eligibility
room-mode option validation
room-mode storage
match-start rule resolution
runtime match-fact evaluation
authoritative MatchDecision and match lock
```

This doc selects `player_spawn_profile_id` and one or more Encounter Spawn Profiles through `encounter_spawn_profile_id`, using only profile-declared validated options. It does not own profile internals or detailed enemy, wave, or level content behind encounter-related IDs. The dedicated [Encounter Spawn Profiles](encounter-spawn-profiles.md) plan owns non-player scheduling and spawn policy.

Detailed gameplay award and counter semantics belong to [Gameplay Awards And Counters](gameplay-awards-and-counters.md). This doc selects and composes the resolved award-policy references consumed by modes; it does not redefine award catalogue, attribution, assists, combos, streaks, distribution, mutation, visibility, or idempotency semantics. Detailed lives, death, elimination, and respawn semantics likewise belong to [Lives, Death, Elimination, And Respawn](lives-death-elimination-and-respawn.md).

Detailed player-spawn placement, safety, profile, and spawn-presentation semantics belong to [Player Spawn Profiles](player-spawn-profiles.md). This doc selects and composes the resolved player-spawn profile reference without duplicating that owner system's placement policy. Encounter Spawn Profile internals likewise remain with [Encounter Spawn Profiles](encounter-spawn-profiles.md).

The shared Objective Foundation owns objective definitions/schema, objective-local lifecycle and state-machine evaluation, discovery/visibility, timers, failure reasons, and normalized objective events/snapshots. Modes consume and select/compose objectives, then own match implications and results policy; they do not become an alternate owner of objective runtime behavior.

Detailed participation and joining semantics belong to [Participation And Joining](participation-and-joining.md). Modes select resolved participation, join, re-entry, and result-eligibility policy; they do not redefine the shared participation-state model or Multiplayer Lifecycle execution.

Modes own match-end policy. A mode may define multiple simultaneous end conditions, and each mode explicitly defines precedence between those conditions. The authoritative game/match layer evaluates the resolved policy and produces one `MatchDecision` through the shared match-lock path. A locked match may end as `completed`, `failed`, `cancelled`, `invalid`, or `administratively terminated`. Once `MatchDecision` locks, the match can never resume; pause is separate from match end.

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
-> MatchRuntimeState / MatchFacts
-> authoritative MatchDecision
-> match lock
-> EndOfMatchFlow
-> MatchSummary
```

Rooms store the validated config, then lock that configuration when the match starts. This pre-start configuration lock is distinct from the one-time `MatchDecision` lock produced later by the authoritative game/match layer.

Gameplay consumes only `ResolvedMatchRules`, not raw room config.

Every normal mode uses the same one-time authoritative match-lock path. By default, locking stops gameplay mutation while result processing and presentation/lifecycle transitions continue. A mode may explicitly bypass or override this default post-end freeze/continuation behavior where its rules require it.

## Preset-Driven Room Modes

Modes are preset-driven room and match configurations.

Rooms store selected config.

Rules resolve config into authoritative match policy.

Gameplay consumes resolved rules.

The baseline implementation must prove the seam with two real modes, not one naming example.

### P4 Preset-Driven Room Mode Foundation

The P4 seam must be proven through two real modes. Players configure room options through presets rather than arbitrary free-form toggles, and presets compose explicitly separate policies:

```text
gameplay award policy
objective policy selection and composition
ranking policy
match-end policy
result policy
team-system selection and mode overrides
damage policy
lives/respawn policy
player spawn profile
Encounter Spawn Profile
join policy
progression eligibility
```

Gameplay award, objective progress, and final ranking are separate concepts. A gameplay score may contribute to an objective without becoming the result-ranking metric.

Likely player-configurable option groups include preset-approved starting lives, target score, time limit, maximum players, team setup, difficulty, hazards, and pickups. Starting lives count total ships, including the initial ship. Both finite and infinite lives are valid resolved policies, and starting lives may be exposed as a preset-approved room option.

Match-end policy is explicit per mode rather than limited to one universal condition. A mode may resolve multiple conditions that become true in the same evaluation, with a mode-declared precedence ordering determining the authoritative result.

Mode is not the same thing as single-player or multiplayer. Single-player and multiplayer are `session_context`; mode governs gameplay rules through `mode_id`. In-game joining remains future functionality controlled by resolved join policy and executed by lifecycle.

### Arcade Survival Baseline

Arcade Survival is the baseline configuration that other modes modify. It adds no special objective or ranking rule.

```text
mode_id: arcade_survival
baseline rules with no mode-specific objective
FFA by default
no player damage
configurable finite or infinite lives
encounter_spawn_profile_id: playercentric_asteroids_v1
ends when no active participants remain
records final match facts
no required competitive ranking
```

The baseline also resolves a player spawn profile separately from the Encounter Spawn Profile.

```text
player_spawn_profile_id: basic_safe_spawn_v1
```

### Score Attack Overrides

Score Attack uses the Arcade Survival baseline and applies explicit objective, ranking, match-end, and result overrides.

```text
mode_id: score_attack

objective:
reach target score

ranking:
completion time, lower is better

match end:
immediate when the first valid participant reaches target
or no active participants remain before anyone succeeds

result:
winner / success / failure
completion time
final score
deaths
target score
```

Score is gameplay award state and objective progress for Score Attack. Completion time is its final ranking metric. The authoritative decision locks on the same evaluation that first observes a valid participant reaching the target; later gameplay mutations cannot alter the winner or ranking inputs.

`score_attack` uses existing score, asteroid destruction, lives and death, match-over evaluation, match results, and room lifecycle. It does not require enemies, waves, bosses, new pickups, campaign state, progression grants, or new objective entities.

### Team-System Selection

Teams are required in the first P4 foundation. Modes select the team structure and permitted room-creation configuration, then consume authoritative membership and relationship facts from [Teams And Team Rules](teams-and-team-rules.md).

The initial selectable structures are FFA, Co-op, Custom Teams, and Auto-balanced Teams. The team-system plan owns assignment, balancing, participation, colour, team relationship facts, aggregation defaults, forfeiture, and team-result requirements. Damage policy, including friendly-fire permissions, belongs to [Damage And Healing Rules](damage-and-healing-rules.md).

Modes retain ownership of mode-specific objective aggregation overrides, team-elimination meaning, whether PvP is enabled, and in-game join policy. Team structure and damage policy remain separate.

### Participation, Spawning, And Match Lock

Player spawning and encounter spawning are independent rule-selectable seams:

```text
player_spawn_profile_id: basic_safe_spawn_v1
encounter_spawn_profile_id: playercentric_asteroids_v1
```

Runtime evaluation uses normalized active participation facts:

```text
disconnected players stop participating in live rule evaluation
their accumulated match facts remain available for final results
match rules evaluate normalized active participation facts
MatchDecision locks once inside the authoritative game/match layer
rooms consume the locked decision rather than reconstructing it
```

Removed or disconnected players therefore cannot block team elimination or match completion, while their historical participation remains available to `MatchSummary`.

Mission support is preparatory and can be implemented before campaign, while campaign itself remains a late future wrapper over missions.

### Affected Systems

Shared contracts / SSoT:

```text
Mode preset IDs and option vocabularies become shared client/server language.
Likely fields include `preset_id`, `mode_id`, `session_context`, starting-lives policy, `target_score`, team configuration, both spawn-profile IDs, mode summary, and result facts.
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
Rooms consume a locked `MatchDecision`; they do not reconstruct match end or winners.
```

Game rules / modes:

```text
Define preset registry.
Validate config.
Construct `ResolvedMatchRules`.
Compose the Arcade Survival baseline plus mode/config overrides.
Select gameplay award and objective policy, ranking, match-end, result, team-system configuration, damage, lives/respawn, player-spawn, one or more Encounter Spawn Profiles, join, and progression-eligibility policies using only profile-declared validated options. Match-end policy may define multiple simultaneous end conditions and must define their precedence. Objective runtime evaluates the selected objective definitions and emits authoritative objective facts.
Evaluate normalized `MatchFacts`, resolve the mode-declared precedence, and lock one authoritative `MatchDecision` through the shared one-time match-lock path. A locked decision carries a terminal match status of `completed`, `failed`, `cancelled`, `invalid`, or `administratively terminated`.
Likely starts near `services/game-server/internal/game/rules`, with exact package split as a gametime decision.
```

Game simulation / player lifecycle:

```text
Consumes resolved finite or infinite lives policy; finite starting_lives counts total ships.
Exposes normalized active participation and historical participant facts.
Removes disconnected players from live rule and team evaluation without discarding accumulated facts.
Consumes separate player and Encounter Spawn Profiles.
Should not parse raw room config throughout simulation.
After `MatchDecision` locks, gameplay mutation stops by default while result processing and presentation/lifecycle transitions continue; a mode may explicitly bypass or override that post-end behavior.
```

Scoring:

```text
Reuses current asteroid gameplay awards for both baseline modes.
Score Attack reads score as objective progress but ranks completion by elapsed time.
Gameplay award, objective, and ranking policy remain separate.
```

Spawning:

```text
Player spawn policy resolves through `player_spawn_profile_id`.
Encounter spawn policy resolves separately through `encounter_spawn_profile_id`.
Both baseline modes use `playercentric_asteroids_v1` for encounter spawning.
```

Damage / targeting / collision:

```text
Arcade Survival defaults to no player damage.
Team relationship and PvP enablement resolve separately.
Detailed damage and healing semantics belong to [Damage And Healing Rules](damage-and-healing-rules.md).
Damage consumes authoritative same-team/opposing-team relationships.
Player same-team damage is prohibited.
Enemy friendly fire and other relationship permissions are source/mode policy.
PvP player damage is inter-team only.
```

Teams:

```text
Required in the P4 foundation.
Selects FFA, Co-op, Custom Teams, or Auto-balanced Teams through the team-system owner.
Resolved rules carry the selected team configuration plus mode-owned aggregation overrides, elimination meaning, PvP enablement, and in-game join policy.
Authoritative assignment, balancing, relationships, forfeiture, colours, and result requirements come from Teams And Team Rules.
```

Match Results:

```text
Result facts distinguish gameplay `mode_id` from `session_context`.
Arcade Survival records final facts without requiring competitive ranking.
Score Attack carries winner/success/failure, completion time, final score, deaths, and target score.
Team and FFA result scopes derive from resolved policy and locked MatchDecision.
Disconnected participants remain available for final summary aggregation.
Visible UI can remain small at first.
```

Player-data / progression:

```text
Not implemented in the initial P4 rules slice.
Future progression needs trusted mode-aware results.
```

Client lobby/session state:

```text
Room snapshots may expose selected preset, option summary, mode locked state, and display name.
```

Devtools:

```text
Future diagnostics should inspect `preset_id`, `mode_id`, resolved rules summary, active MatchFacts, objective/ranking state, locked MatchDecision, both spawn profiles, and separate award/team/damage policies.
```

## Implementation Planning

The first implementation stage should proceed from low-level owner contracts toward integration:

```text
1. Define the low-level policy contracts and runtime fact boundary.
2. Integrate the team-system configuration and authoritative membership contracts.
3. Define lives and both spawn-profile seams.
4. Define authoritative MatchDecision, terminal match statuses, simultaneous-condition precedence, post-end behavior, and the one-time match-lock path.
5. Construct ResolvedMatchRules from baseline plus mode/config overrides.
6. Express Arcade Survival through the baseline.
7. Express Score Attack through explicit overrides.
8. Wire room creation and presentation.
```

This order prevents room storage and UI representation from becoming the authority for policies whose owner systems have not yet been defined.

### P4 Foundation Completion Criteria

- `arcade_survival` exists as the explicit baseline preset with no mode-specific objective or required competitive ranking.
- `score_attack` exists as a second explicit preset built from baseline overrides.
- Gameplay award, objective progress, and ranking metric are separate policy contracts.
- Score Attack reaches a target score, ends immediately on the first valid success, and ranks by completion time.
- FFA, Co-op, Custom Teams, and Auto-balanced Teams are selectable through the team-system owner.
- Game-creation team configuration resolves through the authoritative team-system plan.
- Team-system membership and damage policy remain separate.
- Finite and infinite lives are valid; finite `starting_lives` counts total ships.
- Player and encounter spawning resolve through separate profile IDs.
- `playercentric_asteroids_v1` names the existing encounter-spawning behavior.
- Match rules evaluate normalized active participant facts while retaining historical participant facts.
- Every normal mode produces one authoritative locked `MatchDecision` in the game/match layer.
- Each normal mode may define multiple simultaneous end conditions and explicitly defines their precedence.
- A locked `MatchDecision` ends the match as `completed`, `failed`, `cancelled`, `invalid`, or `administratively terminated`.
- A locked match never resumes; pause remains separate from match end.
- The default post-end behavior stops gameplay mutation while result processing and presentation/lifecycle transitions continue.
- A mode can explicitly bypass or override the default post-end freeze/continuation behavior.
- Rooms consume the locked decision rather than reconstructing it.
- `CreateRoomRequest` or an equivalent room creation path can carry selected mode and team config.
- Server validates requested preset and options, and the room stores validated `RoomModeConfig`.
- Match start resolves `ResolvedMatchRules`; gameplay does not read raw room config.
- Result facts keep `session_context` separate from gameplay `mode_id`.
- Existing single-player and multiplayer create/start flows still work through Arcade Survival.

## Testing Direction

The main checks for this seam are:

```text
room creation rejects invalid preset or option combinations
room storage preserves validated RoomModeConfig
match start resolves ResolvedMatchRules once and uses those rules in gameplay
arcade_survival preserves current play through baseline rules without a special objective
score_attack ends in the same evaluation that first observes a valid target-score success
score_attack ranks successful participants by completion time, lower first
score remains distinct from ranking inputs
all four initial team structures resolve through the team-system owner
mode-selected team configuration and overrides validate where allowed
team and damage policies vary independently
finite starting_lives count total ships
infinite lives do not consume a finite lives counter
player_spawn_profile_id and encounter_spawn_profile_id resolve independently
the baseline player spawn profile is basic_safe_spawn_v1
arcade_survival and score_attack use playercentric_asteroids_v1 encounter spawning
disconnected players leave active evaluation without losing accumulated facts
removed players do not block elimination or match completion
multiple end conditions observed in one evaluation resolve according to the mode's explicit precedence
all normal modes use the same one-time authoritative match-lock path
MatchDecision locks once in the game/match layer and cannot be resumed
pause does not produce a match-end decision
locked matches may produce completed, failed, cancelled, invalid, or administratively terminated outcomes
default post-end behavior stops gameplay mutation while result processing and presentation/lifecycle transitions continue
mode-specific post-end bypass or override behavior is explicit and testable
rooms consume the locked MatchDecision without recomputing winners
target_score affects only score_attack
match results carry mode_id separately from session_context
room snapshots expose the mode summary only if the client needs it
```

## Related Docs

- [Planning](../../!INDEX.md)
- [Player Experience Systems](player-experience-systems.md)
- [Gameplay Awards And Counters](gameplay-awards-and-counters.md)
- [Objectives And Objective Runtime](objectives-and-objective-runtime.md)
- [Teams And Team Rules](teams-and-team-rules.md)
- [Participation And Joining](participation-and-joining.md)
- [Damage And Healing Rules](damage-and-healing-rules.md)
- [Lives, Death, Elimination, And Respawn](lives-death-elimination-and-respawn.md)
- [Player Spawn Profiles](player-spawn-profiles.md)
- [Encounter Spawn Profiles](encounter-spawn-profiles.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)
- [Progression And Rewards](progression-and-rewards.md)

## Open Gametime Decisions

- Exact package split between `game/rules`, `game/rules/modes`, or `game/modes`.
- Exact shared data format for presets.
- Whether first client UI is a full selector or a minimal preset path.
- Exact room snapshot mode-summary shape.
- Exact resolved award-policy reference and contract shape between modes and Gameplay Awards And Counters.
- Exact objective policy selection and composition contracts.
- Exact ranking policy contracts and tie handling.
- Exact implementation contract for expressing mode-declared match-end conditions, precedence, and post-end behavior.
- Exact result-policy shape and participant/team result aggregation details; the locked terminal outcome vocabulary is settled.
- Exact finite/infinite lives and respawn policy contracts.
- Exact implementation contract for the selected Encounter Spawn Profiles is defined in [Encounter Spawn Profiles](encounter-spawn-profiles.md); remaining decisions there are implementation-level.
- Exact damage-policy contract and PvP enablement values.
- Exact join-policy and re-entry eligibility contracts.
- Exact future mission option shape.

## Core Invariants

```text
ModePreset names the preset.
RoomModeConfig carries the validated room options.
ResolvedMatchRules is the only rule object consumed by gameplay.
Rooms store validated config, not raw free-form mode flags.
Match-start rule resolution is the seam where config becomes authoritative rules.
Arcade Survival is the baseline configuration and adds no special objective or ranking rule.
score_attack uses baseline gameplay awards and encounter spawning, but adds explicit objective, completion-time ranking, match-end, and result overrides.
Gameplay award, objective progress, and ranking metric remain separate.
Score Attack ends immediately when the first valid participant reaches target score.
Teams are part of the first P4 foundation; modes select the authoritative team system rather than redefining its semantics.
Team-system membership and damage policy remain separate.
Finite and infinite lives apply to both baseline modes.
Finite starting_lives counts total ships, including the initial ship.
target_score applies only to score_attack.
Player and encounter spawning are separate profile seams.
The baseline player-spawn profile is basic_safe_spawn_v1, with detailed placement owned by Player Spawn Profiles.
The existing encounter-spawning profile is playercentric_asteroids_v1.
Disconnected players do not participate in live rule evaluation, but their historical facts remain available for results.
Modes may define multiple simultaneous end conditions, and each mode explicitly defines their precedence.
Every normal mode uses the same one-time authoritative match-lock path.
MatchDecision locks once in the authoritative game/match layer and may end the match as completed, failed, cancelled, invalid, or administratively terminated.
A locked match never resumes; pause is separate from match end.
By default, a locked match stops gameplay mutation while result processing and presentation/lifecycle transitions continue; a mode may explicitly bypass or override this behavior.
Rooms consume the locked decision and never reconstruct match end or winners.
session_context remains separate from gameplay mode_id.
No gameplay system should read raw room config directly.
```
