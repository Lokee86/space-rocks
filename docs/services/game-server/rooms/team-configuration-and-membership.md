# Team Configuration And Membership

Parent index: [Game Server Rooms](./!INDEX.md)

## Purpose

This document describes the current game-server ownership boundary for room team configuration, deterministic assignment, match membership, and player relationship queries.

## Overview

Rooms own team configuration and assignment before a match starts. The simulation receives the locked result and preserves it on player sessions and historical participant records.

```text
room creation team config
-> teams.ValidateConfig
-> canonical participant roster
-> teams.ResolveAssignments
-> room membership projection
-> Game.SetTeamStructure + player session TeamID
-> relationship and result consumers
```

Supported structures are:

```text
ffa
co_op
custom
auto_balanced
```

Canonical team IDs are `team_1` through `team_8`. FFA uses no gameplay team, while co-op resolves participants onto the shared co-op team model. Custom rooms accept player-selected or owner-assigned memberships. Auto-balanced rooms accept a fixed team count from two through eight.

## Code root

```text
services/game-server/internal/game/teams/
services/game-server/internal/rooms/
```

Simulation readback lives in:

```text
services/game-server/internal/game/team_membership.go
```

## Responsibilities

This boundary owns:

- validating room team structure and assignment mode
- validating auto-balanced team counts
- rejecting blank or duplicate participant IDs
- producing a sorted canonical roster before assignment
- deterministic FFA, co-op, custom, and auto-balanced assignment
- validating canonical team IDs
- locking room-owned team membership into the match
- preserving team membership after participant removal
- exposing self, same-team, opposing, and unaffiliated relationships
- projecting team membership through room snapshots and activation

## Does not own

This boundary does not own:

- mode victory or ranking rules
- shared-life pool mutation
- friendly-fire and damage eligibility
- team score calculation
- client team-selection UI
- account or party identity
- post-match result persistence

Those systems consume normalized team membership rather than inventing their own assignment state.

## Domain roles

Rooms are the assignment authority. `Game` is the runtime consumer and historical record keeper.

The room may rebalance or accept an owner assignment while the room is still configurable. Once the match begins, the simulation reads the locked structure and `TeamID`; it does not rebalance players or reinterpret team membership.

Removed participants keep their last authoritative team in `participantRecords`, allowing outcomes, awards, deaths, and results to use the correct historical relationship after the active player session is gone.

## Protocols and APIs

Primary internal APIs include:

```go
func ValidateConfig(config Config) error
func ResolveAssignments(config Config, participantIDs []string, requested Assignments) (Assignments, error)
func CanonicalRoster(participantIDs []string) ([]string, error)
func RelationshipBetween(...) (Relationship, error)

func (game *Game) SetTeamStructure(structure teams.Structure)
func (game *Game) PlayerTeam(playerID string) teams.ID
func (game *Game) PlayerRelationship(leftPlayerID, rightPlayerID string) teams.Relationship
```

Room creation and room-team methods apply these contracts before player activation. Room and networking snapshots expose the resolved structure and memberships to clients; clients do not submit authoritative runtime relationships.

## Data ownership

`teams.Config` owns the selected structure, optional custom assignment mode, and auto-balanced team count.

`teams.Assignments` maps canonical player IDs to canonical team IDs. The room owns the locked assignment map during lobby and start flow. Active `playerSession.TeamID` and historical participant records carry the resolved result inside the simulation.

Relationship queries are derived from the locked structure, player IDs, and resolved team IDs. They are not independently persisted.

## Code map

```text
services/game-server/internal/game/teams/types.go
services/game-server/internal/game/teams/config.go
services/game-server/internal/game/teams/roster.go
services/game-server/internal/game/teams/resolve.go
services/game-server/internal/game/teams/relationships.go
services/game-server/internal/game/team_membership.go
services/game-server/internal/rooms/room_creation.go
services/game-server/internal/rooms/room_team.go
services/game-server/internal/rooms/room_join.go
services/game-server/internal/rooms/room_projections.go
services/game-server/internal/networking/player_activation.go
services/game-server/internal/networking/room_snapshot.go
```

## Tests

Coverage includes configuration validation, canonical rosters, every assignment structure, room creation, room membership changes, activation handoff, relationship queries, snapshot projection, and preservation after participant removal.

```text
services/game-server/internal/game/teams/*_test.go
services/game-server/internal/game/team_membership_test.go
services/game-server/internal/rooms/room_team_test.go
services/game-server/internal/rooms/room_creation_test.go
services/game-server/internal/networking/team_membership_activation_test.go
```

## Related docs

- [Game Server Rooms](./!INDEX.md)
- [Room Membership And Identity](room-membership-and-identity.md)
- [Lobby And Start Rules](lobby-and-start-rules.md)
- [Modes And Match Rules](../simulation/match/modes-and-match-rules.md)
- [Match Outcomes And Results](../simulation/match/match-outcomes-and-results.md)
- [Teams And Team Rules Planning](../../../planning/domains/gameplay/teams-and-team-rules.md)

## Notes

Team assignment and balancing remain room-owned. The simulation records and consumes the result so combat, lives, awards, objectives, and results share one relationship vocabulary.
