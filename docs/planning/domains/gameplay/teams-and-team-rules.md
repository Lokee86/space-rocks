---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-79bd-9c74-9246b3c0336b
document_type: general
policy_exempt: false
summary: This doc is the authoritative P4 planning owner for team structure and team-rule semantics.
---
# Teams And Team Rules
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc is the authoritative P4 planning owner for team structure and team-rule semantics.

It defines how room team configuration becomes authoritative membership, how team relationships affect participation and damage, and what team facts are handed to spawning, match rules, and results.

## Current Implementation Status

The initial server-side owner system is implemented: configuration validation, canonical rosters, deterministic assignment for all four structures, room-owned membership, simulation relationship queries, snapshot projection, and retained team facts for removed participants.

Current implementation authority: [Team Configuration And Membership](../../../services/game-server/rooms/team-configuration-and-membership.md).

This planning document remains authoritative for deferred product choices and future expansion beyond that implemented foundation.

## Ownership Boundary

This doc owns planning for:

```text
team structures
room team configuration
team assignment and reassignment
team balancing
team participation
team relationship facts
team colour policy
team-aware damage relationships
team-aware spawning handoff
team score/objective aggregation defaults
team elimination semantics boundary
team forfeiture
team result requirements
```

This doc does not own:

```text
mode objectives or match-end policy
whether a mode enables PvP
exact damage formulas
player-spawn profile behavior
room membership and reconnect execution
match-result orchestration
client UI layout
packet or storage schemas
```

Modes and match rules select the team structure and related policy references as part of resolved rules. The team system validates team configuration, resolves authoritative membership, and exposes normalized team facts. Owning systems consume those facts rather than redefining team semantics.

## Settled Product Model

There are four initial team structures:

```text
FFA
Co-op
Custom Teams
Auto-balanced Teams
```

Architecture may later support mode-controlled mid-match rebalancing or team switching. Neither behavior is implemented in the first slice.

Room team configuration is fixed at room creation. Changing team structure, assignment mode, team count, or related room team configuration requires room reconfiguration or a new room; it is not mutable lobby configuration.

Single-player team representation is deliberately implementation-defined and simple because it has no product-level significance.

## FFA

FFA is the current multiplayer baseline.

FFA has no shared-team grouping between participants. When PvP is enabled, FFA participants are opposing sides.

FFA structure does not itself enable PvP. Team structure and whether PvP is enabled remain separate policy concerns.

## Co-op

Co-op uses one implicit shared team.

Players do not configure or select multiple team slots in Co-op. The shared team provides authoritative team membership for aggregation, relationships, elimination, spawning handoff, and results where the selected mode uses those concepts.

## Custom Teams

Custom Teams always exposes eight team slots.

Creators do not configure team count. Empty teams and uneven teams are allowed.

Custom assignment mode is selected at room creation and is one of:

```text
Player-selected
Owner-assigned
```

### Player-selected

Each player controls only their own team assignment. The room owner cannot move other players.

A player may change their own assignment only before readying and before match start.

### Owner-assigned

Only the room owner assigns or reassigns players.

A valid reassignment before match start clears the affected player's ready state.

If the owner leaves, existing team assignments remain. The new owner receives owner-assignment authority under the normal room owner-transfer lifecycle.

## Auto-balanced Teams

At room creation, the creator chooses:

```text
team count from 2 through 8
room player capacity
```

Assignment is immediate. Teams rebalance when players join or leave until match start.

Balancing uses team size only. It does not use skill, score, latency, party, history, or another hidden weighting.

Assignment uses deterministic round-robin among the smallest teams. The implementation must define a stable team ordering and a stable rebalance ordering so identical room facts produce identical assignments.

Assignments are visible before match start.

Leaving and rejoining does not preserve the prior assignment in the first slice. The returning player is assigned from the current room facts.

## Readiness And Assignment Locking

Players cannot switch teams after readying.

Any valid pre-start reassignment clears the affected player's ready state. This applies whether reassignment is initiated by that player, the owner under Owner-assigned rules, or Auto-balanced join/leave rebalancing.

Match start locks assignments for the first slice. Mid-match team switching and rebalancing are not implemented.

## Participation And In-Game Joining

Authoritative team membership is distinct from room membership and active match participation.

Participation And Joining owns normalized participation-state and eligibility semantics. Multiplayer Lifecycle owns room membership, admission/connection execution, and reconnect machinery. The team system consumes normalized connected and active-participation facts; removed or disconnected players stop contributing to live team evaluation while their historical player and team facts remain available for results.

In-game joiners use Auto-balanced assignment or mode-controlled assignment. The selected mode decides whether in-game joining is allowed; the team system resolves the required assignment before the joiner becomes an active participant, while Lifecycle executes admission and activation.

## Team Colours

Team modes force team colours.

Custom player display colours are not available in team modes. Lobby and gameplay presentation consume the resolved team colour for each assigned player.

Exact colour identifiers and palette mapping are implementation-level data decisions.

## Damage Relationships

There is no player same-team friendly fire. This invariant applies to player-on-player damage; it does not globally prohibit enemy-on-enemy friendly fire or other source relationships. Enemy friendly fire and other non-player relationships are resolved by [Damage And Healing Rules](damage-and-healing-rules.md) under source and mode policy.

When PvP is enabled, player damage is inter-team only. Same-team player damage is always rejected.

Team structure and PvP enablement remain separate policy concerns:

```text
team system
-> exposes same-team / opposing-team relationship

mode and damage policy
-> decides whether PvP is enabled
-> permits player damage only for an opposing relationship
```

FFA participants count as opposing sides when PvP is enabled. Co-op participants share the same team. Custom and Auto-balanced participants use their resolved team membership.

## Spawning Handoff

Player-spawn profiles own team-aware spawn placement.

The team system supplies resolved membership and same-team/opposing-team relationships. A player-spawn profile may group teammates and separate opposing teams by default, but exact placement, distance, ordering, safety, and fallback behavior belong to [Player Spawn Profiles](player-spawn-profiles.md).

Team rules do not embed spawn coordinates or placement algorithms.

## Aggregation, Elimination, And Forfeiture

Team score and objective values aggregate by sum unless the selected mode defines another aggregation.

Modes may override the aggregation operation where their objective or ranking semantics require it. The override must be explicit in resolved rules; consumers must not infer a different operation from mode identity.

Team elimination is mode-defined. The team system exposes normalized membership and active-participation facts, while the selected mode defines what elimination means and whether it contributes to match end.

A team with no remaining connected players forfeits.

The first slice can apply that forfeiture immediately from normalized connected-player facts. After Multiplayer Lifecycle V2 exists, a grace period may replace immediate forfeiture so reconnectable players can retain team viability until the grace period expires. Participation And Joining supplies normalized participation/eligibility semantics; Lifecycle owns disconnect/reconnect timing and connected-state facts; team rules own the resulting team-forfeiture requirement.

## Result Requirements

Every match output includes standard individual player results.

Team modes additionally include standard team results. Match output therefore includes standard team results plus individual player results rather than replacing player rows with team-only output.

The team system provides the result requirements and normalized aggregation scope. Modes provide mode-specific outcome meaning and any explicit aggregation override. The authoritative `MatchDecision` locks winning and outcome facts. `EndOfMatchFlow` and `MatchSummary` preserve and emit the resolved team and player results without reconstructing membership.

Historical facts for players who participated before disconnect remain eligible for individual results and for the team aggregates defined by the locked match facts.

## System Handoffs

```text
room creation
-> selected team structure and allowed configuration
-> team-system validation
-> fixed room team configuration
-> pre-start authoritative assignments
-> match-start assignment lock
-> normalized team membership and relationship facts
-> damage / spawning / objective / elimination consumers
-> locked MatchDecision
-> team and individual result output
```

### Modes And Match Rules

Modes select the team structure, allowed team configuration, PvP policy, aggregation overrides, team-elimination meaning, and in-game join assignment policy. They do not redefine assignment, balancing, colour, or baseline team-result semantics; damage policy, including friendly-fire permissions, is resolved by [Damage And Healing Rules](damage-and-healing-rules.md) from mode and source rules.

### Multiplayer Session And Lifecycle

Lifecycle owns room membership, owner transfer, ready-state storage, join/leave execution, connected-state facts, and future reconnect grace timing. Team rules consume those facts to preserve or update assignments, clear readiness when required, assign in-game joiners, and determine forfeiture.

### Player Experience

Player-facing room creation presents the team options allowed by the selected preset. Lobby presentation shows team slots, assignments, team colours, and assignment controls permitted by the resolved structure and assignment mode. Result presentation receives both team and individual result rows for team modes.

### Damage And Spawning

Damage consumes authoritative team relationships plus separate PvP enablement. Player spawning consumes authoritative team relationships through the selected player-spawn profile.

### Match Outcomes And Results

The locked decision and final facts carry team membership, team outcomes, and individual outcomes. Result orchestration preserves those facts and emits standard team plus individual results.

## Implementation Direction

The first implementation slice should proceed from room configuration into authoritative runtime facts:

```text
1. Define the four team-structure identifiers and configuration shapes.
2. Validate fixed room-creation team configuration.
3. Resolve Custom assignment authority and readiness clearing.
4. Resolve deterministic Auto-balanced assignment and pre-start rebalancing.
5. Lock membership at match start.
6. Expose normalized team relationship and participation facts.
7. Route damage, player-spawn profiles, aggregation, elimination, and results through those facts.
8. Preserve seams for later mode-controlled mid-match switching/rebalancing without implementing them.
```

Implementation should keep team policy in a focused gameplay owner. Room, lifecycle, UI, damage, spawning, and result code should route facts or execute their own policy rather than becoming alternate team authorities.

## Testing Direction

Important future checks:

```text
FFA is the multiplayer baseline and treats participants as opposing when PvP is enabled
Co-op resolves one implicit shared team
Custom Teams always exposes eight slots
Custom Teams permits empty and uneven teams
Player-selected assignment lets each player move only themselves
Owner-assigned assignment lets only the owner move players
valid pre-start reassignment clears the affected player's ready state
players cannot switch after readying
owner transfer preserves assignments and transfers owner-assignment authority
Auto-balanced validates 2-8 teams and room capacity
Auto-balanced assignment is immediate, deterministic, and size-only
Auto-balanced join/leave rebalancing runs until match start
Auto-balanced assignments are visible before match start
leave/rejoin receives a fresh assignment
room team configuration cannot mutate in the lobby
match start locks team assignments
team modes force team colours and reject custom player display colours
same-team player damage is always rejected
PvP enablement varies independently from team structure
player-spawn profiles receive team relationship facts without team rules owning placement
team values sum unless resolved mode policy explicitly overrides aggregation
team elimination follows mode policy
no connected players causes team forfeiture
team modes emit standard team results and individual player results
in-game joiners receive Auto-balanced or mode-controlled assignment before activation
single-player does not require product-significant team representation
```

## Related Docs

- [Gameplay Planning](./!INDEX.md)
- [Modes And Match Rules](modes-and-match-rules.md)
- [Participation And Joining](participation-and-joining.md)
- [Damage And Healing Rules](damage-and-healing-rules.md)
- [Gameplay Awards And Counters](gameplay-awards-and-counters.md)
- [Player Experience Systems](player-experience-systems.md)
- [Player Spawn Profiles](player-spawn-profiles.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)
- [Multiplayer Session And Lifecycle](../platform/multiplayer-session-and-lifecycle.md)

## Remaining Implementation-Level Decisions

- Exact type and field names for team structure, configuration, assignment authority, membership, and relationship facts.
- Stable team identifiers, display names, ordering, and colour palette mapping.
- Deterministic player ordering used when Auto-balanced join/leave changes require multiple assignments to be recomputed.
- Exact validation and error vocabulary for room creation and assignment actions.
- Exact owner-transfer and ready-state event/packet handoff shapes.
- Exact representation of FFA opposing-side relationships and the simple single-player representation.
- Exact aggregation-policy contract used when a mode overrides sum.
- Exact normalized connected-player facts used for immediate forfeiture.
- Exact future reconnect grace-period contract after Multiplayer Lifecycle V2.
- Exact team and individual result field names and presentation projection.
- Exact package boundaries and persistence/packet shapes chosen at implementation time.

There are no remaining product-level team decisions blocking P4 system planning.

## Core Invariants

```text
Teams And Team Rules is the authoritative team-semantics planning owner.

FFA, Co-op, Custom Teams, and Auto-balanced Teams are the initial structures.

Room team configuration is fixed at room creation.

Custom Teams always has eight slots; Auto-balanced Teams configures 2-8 teams.

Players cannot switch teams after readying, and valid pre-start reassignment clears ready state.

Match start locks team assignment for the first slice.

Team modes force team colours.

Player same-team friendly fire is never allowed.

Team relationship and PvP enablement remain separate policy concerns.

Player-spawn profiles own team-aware placement.

Team values sum unless the mode explicitly defines another aggregation.

Team elimination is mode-defined.

A team with no connected players forfeits; future Lifecycle V2 may provide a reconnect grace period.

Team modes produce standard team results plus individual player results.

Mid-match team switching and rebalancing are not implemented in the first slice.
```
