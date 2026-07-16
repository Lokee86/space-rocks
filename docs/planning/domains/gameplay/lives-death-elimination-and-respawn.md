# Lives, Death, Elimination, And Respawn
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc is the authoritative P4 planning owner for lives, death, elimination, and respawn semantics.

It defines how a resolved mode tracks available ships, records death and attribution facts, decides whether and when a player may return, and hands lifecycle facts to spawning, match rules, results, and player experience. It defines policy boundaries without claiming that the implementation already exists.

## Ownership Boundary

This doc owns planning for:

```text
starting-lives meaning
per-player life accounting
shared team life-pool accounting
life consumption and life restoration policy boundaries
death event facts and attribution categories
respawn eligibility, trigger, and delay policy boundaries
respawn restoration defaults and mode overrides
between-life runtime-state reset policy boundaries
elimination and recovery semantics
player and team re-entry policy boundaries
manual-respawn stalling and AFK removal policy
simultaneous final-death policy boundaries
life/death/respawn result requirements
```

This doc does not own:

```text
mode objective or match-end policy
room membership, connection, removal, or reconnect execution
player-spawn profile placement or spawn protection behavior
runtime equipment and loadout ownership
combat damage formulas or collision detection
spectate presentation or client UI layout
match-result orchestration
packet or storage schemas
```

Modes and match rules select the life model and related policy as part of resolved rules. The lives/death/respawn system validates and applies those rules, exposes normalized lifecycle facts, and hands execution to the owning systems. Owning systems consume those facts rather than redefining life or elimination semantics.

## Settled Product Model

`starting_lives` means the total number of ships available to the participant, including the current active ship.

The first planning model supports both:

```text
per-player lives
shared team life pools
```

A life is consumed according to the resolved mode policy when the participant's active ship reaches a death outcome. The exact consumption point must be consistent for the selected mode and must not be inferred from a client-side respawn action.

Life transfer may be supported later, but it is not part of the first implementation slice.

Reaching zero lives does not universally require permanent elimination. A mode may define recovery, revival, re-entry, objective-based return, or another mechanism that makes a participant eligible again. Zero lives therefore closes the normal ship-availability path; it is not by itself a universal permanent-elimination rule.

Respawn trigger, respawn delay, restoration, loadout persistence, penalties, recovery, re-entry, and elimination are mode-defined unless this document explicitly identifies a default.

## Lives Ownership

The resolved mode owns the life model and policy. It selects whether lives are infinite, per-player, shared by team, recoverable, or otherwise constrained, and it defines how a death affects availability.

The lives owner maintains authoritative participant and, when selected, team-pool facts. The active ship is runtime state and is not the owner of durable player counters. A runtime ship may be destroyed and recreated without losing the participant's score, lives, elimination state, or historical death facts.

For a finite per-player pool, the participant's life count represents the remaining total ship opportunities under the resolved mode. For a shared team pool, the pool is authoritative for the team and individual active-ship state remains separate. The team system supplies normalized membership; it does not independently decrement or invent shared-life facts.

Infinite lives prevent death-based life exhaustion only. Forfeiture, objective failure, disconnect/removal, and other mode-defined elimination remain valid even when a participant or team has infinite lives.

The first slice does not include life transfer between participants or teams. Any future transfer must be an explicit mode policy with authoritative validation, attribution, result, and team-pool handoffs rather than an incidental mutation by UI or transport code.

## Death And Attribution

Death is an authoritative gameplay outcome, not merely the disappearance of a client entity. The death fact identifies the participant, the destroyed active ship, the relevant match and mode context, and the life/elimination transition selected by policy.

Death attribution supports:

```text
killer
assists
self-destruction
environmental death
unattributed death
```

A killer is the authoritative credited damaging participant or other eligible source under the mode's attribution rules. Assists are a separate set of credited contributors and must not be collapsed into the killer field. Self-destruction identifies a death caused by the participant's own action or owned effect. Environmental death identifies hazards or world causes. Unattributed death is valid when no eligible source can be established.

Attribution policy may define damage windows, ownership transfer, ties, source expiry, and whether a source is eligible for a penalty or reward. Those are mode/combat policy decisions, not reasons for results or clients to reconstruct attribution from visible effects.

Death penalties are mode-defined. A penalty may affect score, resources, objectives, cooldowns, loadout, or another mode-owned value, but the death system only provides the authoritative death transition and the normalized attribution facts needed by the selected policy.

## Respawn Trigger And Delay

Respawn trigger is mode-defined. A mode may trigger respawn automatically after death, require a player-controlled confirmation, wait for a team or objective event, or use another explicit trigger. A participant with no current respawn eligibility must not be respawned merely because a delay elapsed.

Respawn delay is mode-defined. When permitted by the mode, it may be exposed as a game-creation option. The option is validated against the selected mode's allowed range and behavior; clients must not apply an arbitrary local delay.

A manual respawn flow must expose an authoritative pending-respawn state. The participant may stall manual respawn only for the configured allowance. When the AFK timer expires, the participant is removed according to mode policy. Exact timing and removal behavior are configurable. Removal ends active evaluation while preserving historical participant facts for results.

A delay does not itself decide elimination, match end, spawn placement, or spawn protection. It only gates the selected respawn trigger after eligibility has been established.

## Respawn Restoration

The default respawn restoration is:

```text
full health
full shields
cooldown reset for cooldowns shorter than a tunable threshold initially around 10 seconds
removal of temporary effects
```

Effects may explicitly persist through death. Modes may override restoration behavior, including health, shields, cooldown treatment, temporary effects, or other runtime state. The resolved policy must identify overrides explicitly; consumers must not infer them from mode identity or from the existence of a respawn.

The cooldown threshold is a tunable implementation-level value, initially expected to be around 10 seconds. The threshold applies only to the default restoration rule and does not prevent a mode from defining a different rule.

Restoration is authoritative and occurs as part of respawn execution. Presentation may animate the transition, but it must not grant health, shields, cooldown resets, or effect removal independently.

## Loadout And Runtime State

Loadout persistence and between-life loadout changes are mode-defined. A mode may preserve the selected loadout, require a re-selection, permit a limited change window, or apply another explicit policy.

Runtime equipment remains owned by the runtime equipment/loadout system. The lives/death/respawn owner requests the resolved between-life policy and reports the life transition; it does not become an alternate inventory, equipment, or loadout authority.

The active ship's transient runtime state is recreated or restored through the respawn path. Durable player counters such as lives and score remain on the player-session or mode-owned state, not on the destroyed runtime ship.

## Spawn Protection Handoff

Spawn protection is owned by the player spawn profile. This is a hard seam with high gameplay and balance risk that may require tuning, replacement, or removal.

The lives/death/respawn owner supplies the respawn context and the selected profile reference. The player-spawn profile owns protection duration, protected interactions, movement or attack restrictions, cancellation conditions, and placement-related safety behavior. Lives and respawn rules must not embed spawn-protection timers or bypass the profile.

A mode may select a profile or disable protection where the profile contract permits it. Any profile override must remain visible in resolved rules so gameplay and balance tests can exercise it explicitly.

## Elimination, Recovery, And Re-entry

Elimination is mode-defined. A participant may become eliminated when lives are exhausted, when a team pool is exhausted, when an objective fails, on forfeiture, on disconnect/removal, or through another mode-defined condition. These conditions must not be treated as universally equivalent: a mode may provide recovery or re-entry after zero lives or another elimination transition.

Eliminated-player participation is also mode-defined and may include:

```text
spectating
leaving the match
non-combat participation
later re-entry
```

An eliminated participant is removed from active combat evaluation unless the resolved mode explicitly allows a non-combat or recovery state. The lifecycle state must distinguish active, pending respawn, eliminated, removed, and any mode-specific recoverable state needed by the rules.

Team elimination is mode-defined. A team may remain viable while an individual is eliminated, or the mode may end the team's combat participation when its resolved team condition is met. Team membership and connected/active facts come from the team and multiplayer lifecycle owners; the selected mode determines the elimination consequence.

Recovery and re-entry must be authoritative, mode-approved transitions. They must not be simulated as a new unrelated participant, and they must preserve the historical facts required by results.

## Teams And Shared Lives

For modes using shared team life pools, the team system supplies authoritative membership and normalized team participation facts. The lives owner applies the resolved pool policy and exposes the remaining shared availability without duplicating team assignment or team elimination authority.

A shared pool may consume one life per team ship death, per participant death, or according to another explicit mode rule. The selected rule must be resolved before match start and must be applied consistently to simultaneous or near-simultaneous deaths where the mode cares about ordering.

A team can be eliminated even when individual participants have remaining personal lives if the mode defines the shared pool or another team condition as decisive. Conversely, a team need not be eliminated merely because one member reaches zero personal lives.

Life transfer is not implemented in the first slice. Team rules still own team membership and team relationship facts, while modes own whether team elimination contributes to match end.

## Participation, Disconnect, And Reconnect

Mid-match join life state is mode-defined. A permitted joiner may receive the mode's starting life state, a constrained entry state, a shared team pool allocation, or another explicit state. Lifecycle owns admission and activation; lives rules resolve the life state before the joiner becomes an active participant.

Reconnect restores the player's previous match state. Reconnect must not silently reset lives, active/eliminated state, attribution history, loadout policy state, or pending-respawn state unless the resolved mode explicitly defines that behavior.

Disconnected or removed players leave active evaluation. Their historical participant facts remain available to results, including prior lives/death facts, attribution, team membership at the relevant time, and the reason for removal where the result policy includes it.

Lifecycle owns connection, removal, reconnect execution, and any reconnect grace timing. The lives/death/respawn owner consumes normalized lifecycle facts and applies the mode-defined consequence. It does not own transport membership or invent connection state.

## Match-End Timing Handoff

The selected mode owns whether elimination contributes to match end and when the match is considered resolved.

Post-elimination match-end timing is configurable to support:

```text
immediate ending
delayed resolution
campaign presentation
death sequences
future mode-specific behavior
```

The lives/death/respawn owner emits normalized final-death and elimination facts. Match rules and outcome orchestration decide whether to end immediately, wait for a configured resolution window, allow a death sequence, or apply another mode policy.

Simultaneous final-death handling is only required for modes where it matters. A mode that cares must define the tie, ordering, survival, and result consequences explicitly. Modes that do not care may use their ordinary elimination transition without adding special global behavior.

## Result Requirements

Which life, death, respawn, elimination, recovery, and re-entry facts appear in results is strictly mode-defined.

The lives/death/respawn owner provides normalized historical facts and the result requirements selected by the mode. `MatchDecision`, `EndOfMatchFlow`, and `MatchSummary` preserve and emit those facts without reconstructing life counts or attribution from runtime entities.

Result policy may include, for example:

```text
starting and remaining lives
death count and death reasons
killer and assist attribution
self-destruction, environmental, and unattributed death facts
respawn count and delay-related facts
elimination or recovery outcome
team-pool contribution or team elimination facts
forfeiture, disconnect, or removal reason
```

None of these fields is universally required. Historical facts for disconnected or removed participants remain eligible when the selected mode's result policy includes them.

## System Handoffs

```text
resolved mode rules
-> life model, death policy, respawn policy, and elimination policy
-> authoritative player/team life state
-> death event and attribution facts
-> life transition and respawn eligibility
-> mode-defined trigger and delay
-> runtime restoration and loadout handoff
-> player-spawn profile placement and protection
-> active, eliminated, recovery, or re-entry state
-> match outcome evaluation
-> locked MatchDecision and mode-defined results
```

### Modes And Match Rules

Modes select life scope, starting lives, infinite-life behavior, death penalties, respawn trigger and delay, restoration overrides, loadout policy, elimination/recovery rules, team elimination, simultaneous-final-death behavior, mid-match join state, AFK removal behavior, match-end timing, and result fields. They do not redefine runtime equipment, spawn placement, connection execution, or result orchestration.

### Multiplayer Session And Lifecycle

Lifecycle owns room membership, admission, activation, connection state, removal, reconnect execution, and future reconnect grace timing. Lives/death/respawn consumes those facts, preserves previous match state on reconnect, and removes disconnected or AFK-removed participants from active evaluation without erasing historical result facts.

### Combat And Runtime Ship

Combat owns damage and collision outcomes and supplies authoritative death causes or source candidates. Runtime ship owns the live avatar state. Lives/death/respawn converts an authoritative death into the mode-defined life and lifecycle transition without moving durable counters onto the ship.

### Player Loadout And Equipment

The loadout/equipment system owns equipment and runtime equipment state. Lives/death/respawn supplies the mode-defined between-life persistence or change policy and requests the restoration/reset operation; it does not own inventory or loadout data.

### Player Spawn Profiles

Player-spawn profiles own placement, spawn safety, and spawn protection. Respawn supplies the eligible participant and profile context. The profile returns the spawn outcome and protection state without lives rules embedding placement mechanics.

### Match Outcomes And Results

Outcome rules consume normalized active, eliminated, recovery, team-pool, and final-death facts. Match-result orchestration locks the decision and emits only the life/death/respawn facts required by the selected mode.

### Player Experience

Client presentation displays authoritative life, death, respawn, elimination, spectate, and re-entry states. It may collect manual-respawn input and present delay or protection feedback, but it cannot advance lives, grant respawn, or decide elimination locally.

## Implementation Direction

The first implementation slice should proceed from resolved mode policy into authoritative lifecycle facts:

```text
1. Define normalized life models for finite per-player, shared-team, and infinite lives.
2. Define starting_lives as total ships including the current active ship.
3. Record authoritative death causes and killer/assist attribution categories.
4. Apply mode-defined life consumption, penalties, and respawn eligibility.
5. Apply mode-defined trigger, configurable delay, and manual-respawn AFK removal.
6. Route default restoration and explicit mode overrides through runtime state owners.
7. Hand placement and spawn protection to the selected player-spawn profile.
8. Expose active, pending, eliminated, recovery, re-entry, and removed facts.
9. Feed locked outcome and mode-defined result requirements without reconstructing facts.
10. Preserve seams for life transfer and richer recovery without implementing them.
```

Implementation should keep life/death/respawn policy in a focused gameplay owner. Modes, lifecycle, combat, runtime ship, equipment, spawning, outcome, and presentation code should route facts or execute their own policy rather than becoming alternate authorities.

## Testing Direction

Important future checks:

```text
starting_lives includes the current active ship
finite per-player lives decrement according to resolved death policy
shared team pools are authoritative and do not duplicate team membership ownership
infinite lives prevent death-based exhaustion but not forfeiture or other mode-defined elimination
life transfer is not available in the first slice
killer, assists, self-destruction, environmental, and unattributed deaths are distinguishable
death penalties follow mode policy
respawn trigger and delay follow mode policy
permitted respawn delay options are validated by the selected mode
manual-respawn stalling invokes the configurable AFK removal behavior
full health and full shields are the default restoration
cooldowns shorter than the initial approximately 10-second threshold reset by default
temporary effects are removed by default unless explicitly persistent
modes can override restoration behavior
loadout persistence and between-life changes remain mode-defined
runtime equipment remains owned by the loadout/equipment system
spawn protection is supplied by the player-spawn profile and remains replaceable/removable
zero lives does not universally imply permanent elimination
recovery and re-entry are explicit authoritative mode transitions
eliminated participation follows mode policy
team elimination follows mode policy
simultaneous final deaths receive special handling only where the mode requires it
mid-match joiners receive mode-defined life state
reconnect restores previous match state
disconnected or removed players leave active evaluation but remain eligible for historical results
post-elimination timing supports immediate and delayed mode-defined resolution
result facts are emitted strictly according to mode requirements
```

## Related Docs

- [Gameplay Planning](./!INDEX.md)
- [Modes And Match Rules](modes-and-match-rules.md)
- [Gameplay Awards And Counters](gameplay-awards-and-counters.md)
- [Teams And Team Rules](teams-and-team-rules.md)
- [Player Spawn Profiles](player-spawn-profiles.md)
- [Player Build And Loadouts](player-build-and-loadouts.md)
- [Player Experience Systems](player-experience-systems.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)
- [Multiplayer Session And Lifecycle](../platform/multiplayer-session-and-lifecycle.md)

## Remaining Implementation-Level Decisions

- Exact life-state, death-event, attribution, respawn-request, and elimination type/field names.
- Exact ownership and storage location for per-player life state and shared team pools.
- Exact point at which a death consumes a life for each supported life model.
- Exact killer eligibility, assist window, source expiry, and attribution tie rules.
- Exact death-penalty contract and mode override representation.
- Exact respawn-trigger and manual-confirmation event flow.
- Exact allowed respawn-delay range, game-creation option shape, and validation errors.
- Exact cooldown threshold value and classification of cooldowns shorter than the threshold.
- Exact temporary-effect persistence markers and restoration override shape.
- Exact loadout persistence and between-life change-window handoff.
- Exact player-spawn profile contract for protection and its tuning/removal controls.
- Exact recoverable and re-entry lifecycle states.
- Exact team-pool consumption and simultaneous-death ordering rules.
- Exact AFK timer defaults, reset events, and removal behavior.
- Exact reconnect preservation and grace-period contract.
- Exact post-elimination timing and simultaneous-final-death decision contracts.
- Exact result field names, historical retention, and presentation projection.
- Exact package boundaries and packet/storage shapes chosen at implementation time.

There are no remaining product-level lives, death, elimination, or respawn decisions blocking P4 system planning.

## Core Invariants

```text
Lives, Death, Elimination, And Respawn is the authoritative planning owner for these semantics.

starting_lives counts total ships, including the current active ship.

Per-player lives and shared team life pools are supported.

Life transfer is not in the first implementation slice.

Zero lives does not universally require permanent elimination.

Respawn trigger and delay are mode-defined; permitted delay may be a game-creation option.

Default respawn restoration is full health, full shields, short-cooldown reset, and temporary-effect removal.

Effects may explicitly persist through death, and modes may override restoration.

Loadout persistence and between-life loadout changes are mode-defined.

Runtime equipment and durable player counters remain with their owning systems, not the active ship.

Player-spawn profiles own spawn protection as a hard, high-risk seam.

Death attribution supports killer, assists, self-destruction, environmental, and unattributed deaths.

Eliminated participation, team elimination, recovery, and re-entry are mode-defined.

Post-elimination match-end timing is configurable and owned by match rules/outcome orchestration.

Simultaneous final-death handling is required only for modes where it matters.

Mid-match join life state is mode-defined, and reconnect restores previous match state.

Infinite lives prevent death-based exhaustion only; other mode-defined elimination remains valid.

Manual-respawn stalling results in configurable AFK removal.

Life/death/respawn facts appear in results strictly according to mode policy.

Disconnected or removed participants leave active evaluation while historical facts remain available to results.
```
