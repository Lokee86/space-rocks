---
author: brian
created: "2026-08-04"
document_id: 019f7d55-fb2c-7f7b-9b24-2dbb3a7c1e51
document_type: general
policy_exempt: false
summary: Current game-server team configuration, assignment, roster, relationship, and lifecycle ownership.
---
# Game-Server Teams And Team Membership

Parent index: [Game Server Simulation](./!INDEX.md)

## Status

This is the canonical current-service document for the implemented team-system slice on main. It documents server behavior only; future mode mechanics remain in [Teams And Team Rules](../../../planning/domains/gameplay/teams-and-team-rules.md).

## Ownership

The focused `game/teams` package owns team identifiers, structure/configuration validation, canonical roster ordering, assignment resolution, and participant relationships. The room owns room-scoped configuration, mutable lobby assignments, readiness effects, and the assignment lock. The game owns the active and historical player-to-team fact and reads the locked structure for relationship queries. Networking routes room requests and applies the locked room snapshot when activating players; it does not resolve team policy.

## Configuration

`teams.Config` supports these structures:

- `ffa`: participants resolve to `NoTeam`; non-self participants are opposing for relationship queries.
- `co_op`: participants resolve to `team_1`.
- `custom`: uses `player_selected` or `owner_assigned`; the available team IDs are `team_1` through `team_8`.
- `auto_balanced`: uses `AutoTeamCount` from 2 through 8 and has no manual assignment mode.

FFA is the default room configuration. FFA and Co-op reject assignment mode and team-count fields. Custom rejects a team count and requires a valid assignment mode. Auto-balanced rejects a manual assignment mode and validates its count. Invalid structures, team IDs, blank participant IDs, and duplicate participant IDs are rejected with errors.

Room creation validates and stores this configuration with room capacity. The room configuration is the source for subsequent assignment resolution.

## Assignment And Roster

`CanonicalRoster` copies and sorts participant IDs lexically, rejecting blank or duplicate IDs. Baseline FFA and Co-op assignment uses that roster. Custom assignment requires exactly one valid, non-`NoTeam` assignment for every participant and rejects unknown or missing participants.

Auto-balanced assignment uses the canonical roster, the first configured team IDs, and team size only. For each participant it chooses the currently smallest team, using stable team order to break ties. This makes equal room facts produce deterministic assignments.

The room rebalances Auto-balanced assignments from the current membership roster while in the lobby. Membership changes can therefore change assignments; the resolver does not preserve a departed member's prior assignment. A changed assignment clears a non-bot member's ready state. Custom assignment changes are available only in the lobby, only for Custom rooms, and only to valid team IDs:

- Player-selected members may move only themselves and only while not ready.
- Owner-assigned members may be moved only by the room owner. Moving a non-bot target clears that target's ready state.

The room exposes defensive assignment snapshots. Bots participate in the same locked assignment snapshot; owner-assigned bot changes do not clear bot readiness.

## Lifecycle Integration And Failure Behavior

Starting a room resolves assignments against the current membership and locks them. `TeamStartSnapshot` is unavailable before the lock and returns a defensive copy after the lock. A failed start rolls back the lock and leaves the room in the lobby. Returning to the lobby unlocks assignments. Assignment mutation after the room leaves the lobby is rejected.

Networking accepts team fields in room creation and routes manual assignment requests to the room. At activation, networking reads the locked snapshot, sets the game structure, and calls `AddPlayerWithTeam` or `AddBotWithTeam` for assigned members. Missing assignments are not activated. If session binding or member activation fails, the newly added player is rolled back or removed. Deactivation clears session-to-game-player bindings and resets the active-player count; it does not redefine historical team facts.

## Runtime State And Invariants

The game stores the team on the durable participant/session or participant record, not only on the live ship. `PlayerTeam` can therefore read a team after a player entity is removed. Respawn preserves membership; rollback removes the new membership and participant state; a fresh game does not inherit a prior game's facts.

`PlayerRelationship` returns `self`, `same_team`, `opposing`, or `unaffiliated`. Unknown players, invalid structure reads, and invalid relationship inputs fall back to `unaffiliated` at the game boundary. FFA participants are opposing, Co-op participants are same-team, and Custom/Auto-balanced participants are same-team only when their non-empty team IDs match. `NoTeam` in Custom/Auto-balanced produces an unaffiliated relationship.

The current implementation does not implement mid-match reassignment or rebalancing, team objectives, team elimination, forfeiture, team scoring aggregation, friendly-fire policy, or team-aware spawn placement. Those consumers may use the current membership facts when separately implemented, but the team package does not claim those policies.

## Client Responsibility

The client has a focused presentation/configuration helper at `client/scripts/teams/team_presentation.gd`. It maps the eight canonical IDs to display names, hue shifts, colors, and visible IDs for a requested count; it formats the four structure names and the two custom assignment modes. The room-setup readout exposes the structure, custom assignment mode, auto-balanced count, and capacity fields and sends them through the existing networking client. Client presentation consumes server-provided `team_id` values, including match-end player rows; it is not authoritative for assignment or relationships.

## Code Map

- `services/game-server/internal/game/teams/types.go` — IDs, structures, assignment modes, relationships, and config types.
- `services/game-server/internal/game/teams/config.go` — configuration validation.
- `services/game-server/internal/game/teams/roster.go` — canonical roster validation/order.
- `services/game-server/internal/game/teams/baseline_assignments.go` — FFA and Co-op resolution.
- `services/game-server/internal/game/teams/custom_assignments.go` — Custom validation.
- `services/game-server/internal/game/teams/auto_assignments.go` — deterministic Auto-balanced resolution.
- `services/game-server/internal/game/teams/resolve.go` — structure dispatch.
- `services/game-server/internal/game/teams/relationships.go` — relationship resolution.
- `services/game-server/internal/rooms/room_team.go` — room configuration, lobby assignment, lock, and snapshot.
- `services/game-server/internal/game/team_membership.go` — game-owned structure/team/relationship reads.
- `services/game-server/internal/networking/room_handlers.go` — inbound room team requests.
- `services/game-server/internal/networking/player_activation.go` — locked snapshot handoff to game players.
- `client/scripts/teams/team_presentation.gd` — client display mapping.
- `client/scripts/ui/transmission_displays/multiplayer_room_setup_readout.gd` — client team configuration controls.

## Tests

Current coverage includes team type/config validation, canonical rosters, baseline/custom/auto-balanced assignment, relationship resolution, room assignment permissions and readiness, roster rebalancing, lock/snapshot defensive copying, failed-start rollback, lobby unlock, game membership/respawn/history behavior, and networking activation of member and bot assignments. Relevant files include `internal/game/teams/*_test.go`, `internal/rooms/room_team_test.go`, `internal/rooms/room_creation_test.go`, `internal/game/team_membership_test.go`, and `internal/networking/team_membership_activation_test.go`.

## Related Docs

- [Game Server Simulation](./!INDEX.md)
- [Teams And Team Rules](../../../planning/domains/gameplay/teams-and-team-rules.md)
- [Client Team Presentation And Configuration](../../client/team-presentation-and-configuration.md)
