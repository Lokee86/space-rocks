---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7981-b569-b7f2a57540fe
document_type: general
policy_exempt: false
summary: This document describes the current cross-system match end and results flow for Space Rocks.
---
# Match End And Results Flow

Parent index: [Player Experience](./!INDEX.md)

## Purpose

This document describes the current cross-system path from authoritative mode evaluation to room match-over, client results presentation, and durable stat reporting.

## Overview

```text
simulation facts
-> mode-owned MatchDecision
-> locked MatchSummary
-> room GameOver transition
-> presentation-safe room snapshot
-> client results flow

locked MatchSummary
-> game-server result reporter
-> player-data identity route
-> guest, local-profile, or account stat update
```

Local player elimination and authoritative match-over remain separate. A local player can be eliminated while the match continues. Final results appear only after the room reaches `GameOver` from a server-owned terminal decision.

## Participating systems

- Game-server simulation: owns current player, score, objective, team, life, and participation facts.
- Modes and match rules: decide terminal status, end reason, winners, placement data, and mode-specific targets.
- Match-results runtime: combines locked decisions with retained participant and objective facts.
- Rooms: store the once-only summary and own the `GameOver` transition.
- Networking: projects a presentation-safe result in room snapshots.
- Client match-end flow: distinguishes local elimination from room match-over and renders result rows.
- Game-server reporting integration: translates the locked summary into player-data commands.
- Player-data: validates identity/mode routing and updates aggregate stats.
- API server: persists authenticated-account results.

## Authority boundaries

### Match completion

The selected mode is the authority for whether current facts are terminal. Arcade Survival ends when no active participants remain. Score Attack ends on the first recorded target-score success, or fails when no active participant remains before success.

The client never infers final match completion from local lives, HUD state, or a local death event.

### Result construction

The match-results runtime is the final summary owner. It consumes retained participant facts, including players already removed from active sessions, and preserves team, score, deaths, disposition, completion time, target, objective, mission, and challenge data selected by the mode.

The end-of-match flow locks its first successful summary and returns clones thereafter.

### Room lifecycle

Rooms own the transition into `GameOver`, store the summary for the current match, and expose it to both presentation and reporting. Returning to lobby is a later room action; opening the result window does not destroy the game or mutate persistence.

### Presentation

The client receives only a presentation-safe projection. Durable account and local-profile identity stay on the reporting path.

Repeated `GameOver` snapshots do not remount duplicate result windows. Local elimination can update local presentation but cannot expose final results.

### Persistence

The game server reports trusted result facts once per player. Player-data routes guest results to memory, local-profile results to embedded SQLite, and authenticated-account results to the API server. Result IDs provide idempotence.

## Flow summary

1. Room start locks mode, teams, and match ID.
2. Simulation updates lives, counters, objectives, participation, and retained participant history.
3. Mode evaluation produces a terminal decision when its ordered end conditions are met.
4. Match-results resolves participant and team outcomes and locks a summary.
5. Room enters `GameOver` and broadcasts a presentation-safe result projection.
6. Client match-end flow presents results once and emits replay, lobby, pregame, or quit intent.
7. Game-server reporting maps the same locked summary into player-data commands.
8. Player-data applies idempotent aggregate stat updates through the identity-appropriate store.

## Inputs and outputs

Inputs:

- resolved match rules
- authoritative participant, team, score, death, objective, and completion facts
- room match identity and trace identity
- retained facts for departed or forfeited participants
- client result-window intent

Outputs:

- locked terminal status, reason, winners, outcomes, and placements
- immutable `MatchSummary`
- room `GameOver`
- presentation-safe snapshot result
- client result rows and route intent
- idempotent player-data result commands
- updated transient or durable aggregate stats

## Out of scope

This document does not define:

- result-window visual layout
- detailed database schemas
- achievement or progression award policy
- leaderboards
- reconnect policy
- future campaign debrief presentation
- packet source definitions

## Related docs

- [Player Experience](./!INDEX.md)
- [Modes And Match Rules](../../services/game-server/simulation/match/modes-and-match-rules.md)
- [Match Outcomes And Results](../../services/game-server/simulation/match/match-outcomes-and-results.md)
- [Objective Runtime](../../services/game-server/simulation/match/objective-runtime.md)
- [Game Server Rooms](../../services/game-server/rooms/!INDEX.md)
- [Match Result Reporting](../../services/game-server/integrations/match-result-reporting.md)
- [Player Data Match Result Sinks](../../services/player-data/match-result-sinks.md)
- [API Player Stats And Match Results](../../services/api-server/player-stats-and-match-results.md)

## Notes

One locked summary feeds both presentation and persistence, but each receives only the projection it needs. Local elimination remains a nonterminal player-experience state unless the selected mode also resolves the match.
