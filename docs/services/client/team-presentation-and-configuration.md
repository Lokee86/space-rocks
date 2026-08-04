---
author: brian
created: "2026-08-04"
document_id: 019f7d55-fb2c-7f7b-9b24-2dbb3a7c1e52
document_type: general
policy_exempt: false
summary: Current client team configuration controls and team presentation mapping.
---
# Client Team Presentation And Configuration

Parent index: [Client](./!INDEX.md)

## Purpose

This document defines the current client-owned team configuration controls and presentation mapping.

## Overview

The client exposes room team configuration choices, sends team-related requests, and formats server-resolved team facts for presentation. The server remains authoritative for configuration validation, assignment, roster membership, lock state, and participant relationships.

## Status And Boundary

This document covers the implemented client-side team slice. The server remains authoritative for team configuration, assignment, roster, lock state, and relationships. The client formats configuration choices, sends requests, and presents resolved team facts; it does not validate or assign teams authoritatively.

## Configuration Controls

`multiplayer_room_setup_readout.gd` exposes the room's team structure, Custom assignment mode, Auto-balanced team count, and room capacity. The assignment selector is visible only for `custom`; the team-count selector is visible only for `auto_balanced`. Custom offers `owner_assigned` and `player_selected`. Auto-balanced offers counts from 2 through 8. The assembled values are sent by `client_connection_service.gd` in the existing create-room request.

Manual assignment requests are sent through `send_set_team_assignment_request(target_player_id, team_id)`. Server errors and permission rules remain server-owned.

## Presentation Mapping

`TeamPresentation` defines the canonical display IDs `team_1` through `team_8`, names them `TEAM 1` through `TEAM 8`, and returns `NO TEAM` for an unknown or empty ID. It derives a team color from a shared base color and the generated player hue constants. `ids_for_count` exposes the first N IDs, clamped to the eight supported IDs. It formats `ffa`, `co_op`, `custom`, and `auto_balanced`, plus the two custom assignment modes; unknown values use the FFA or automatic fallback labels.

Match-end presentation preserves a non-empty `team_id` on player result rows. The client therefore displays resolved server facts rather than reconstructing membership locally.

## Code Map

- `client/scripts/teams/team_presentation.gd` — canonical client display and color mapping.
- `client/scripts/ui/transmission_displays/multiplayer_room_setup_readout.gd` — room setup controls and visibility rules.
- `client/scripts/networking/client_connection_service.gd` — create-room and assignment request routing.
- `client/scripts/gameplay/match_end/match_end_flow.gd` — preserves `team_id` in player result rows.
- `client/scripts/generated/networking/packets/packets.gd` — generated wire fields and packet constructors; do not hand-edit.

## Not Client-Owned

The client does not own team validation, canonical roster ordering, assignment balancing, readiness clearing, assignment locking, lifecycle activation, relationship calculation, or failure semantics. Future team HUD, team roster presentation, team objectives, and mode-specific team mechanics remain planning work unless separately implemented.

## Related Docs

- [Client](./!INDEX.md)
- [Game-Server Teams And Team Membership](../game-server/simulation/teams-and-team-membership.md)
- [Teams And Team Rules](../../planning/domains/gameplay/teams-and-team-rules.md)

## Notes

Future team HUD, roster, objective, and mode-specific presentation must extend this client boundary without recreating server team authority.
