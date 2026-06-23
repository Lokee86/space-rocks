# Domain Backlog
Parent index: [Planning](./!INDEX.md)

## Role

This file tracks unscheduled, cross-domain, or not-yet-routed planning items.

It is a triage and routing document, not a feature backlog. Detailed system plans belong in the owner docs for the relevant domain, service, protocol, data, or devtools area.

Current limitations belong in [docs/limits/current-system-limits.md](../limits/current-system-limits.md) and [docs/limits/player-build-limits.md](../limits/player-build-limits.md).

Detailed player build and loadout planning belongs in [docs/planning/player-build-and-loadouts.md](domains/gameplay/player-build-and-loadouts.md).

Roadmap phase sequencing belongs in [docs/planning/development-roadmap.md](development-roadmap.md).

## How To Use This File

Use this file to capture an item only until it has a clear owner, a stable decision, or a better home in another doc.

When an item is routed, move it to the owning document and remove it from this file.

If an item is already fully specified in an owner doc, do not duplicate it here.

## Triage States

`Needs owner`: the item has no clear owning doc or system yet.

`Active decision`: the item needs an unresolved cross-domain decision before routing can continue.

`Blocked`: the item cannot move forward because another dependency or constraint must land first.

`Parked`: the item is acknowledged but intentionally deferred for now.

`Routed`: the item has a clear owner doc and should be moved there.

`Cut`: the item is no longer being pursued and should be removed from active planning.

## Retention Rules

Keep an item in this file only if it blocks another system, protects correctness/security/trust, defines ownership or service boundaries, represents an unresolved decision, or lacks a better owner doc.

Once an item has a clearer owner, move it out of this file instead of letting it linger as backlog.

Do not use this file to accumulate feature ideas, implementation notes, or detailed subsystem plans.

## Active Cross-Domain Decisions

| Decision | Why It Matters | Blocks | Owner |
| --- | --- | --- | --- |
| Player-data contract enforcement | Prevents schema drift across local profile, account, game-server, and API surfaces. | Durable progression, profile migration, loadout persistence. | [data-sync-and-ssot-pipeline.md](../data/data-sync-and-ssot-pipeline.md) and [Player Data And Persistence](../services/player-data/!INDEX.md) |
| Player-data service boundary | Decides whether player-data remains in-process or becomes an extracted service. | Local profile persistence, match-result writes, loadout reads. | [Player Data And Persistence](../services/player-data/!INDEX.md) and [API Product Surface](protocol/api-product-surface.md) |
| Online admission and auth routing | Multiplayer cannot be trusted until identity and admission rules are explicit. | Hosted multiplayer, account rewards, rankings. | [account-and-identity-systems.md](domains/platform/account-and-identity-systems.md) and [game-integrity-policy.md](domains/platform/security-and-admin/game-integrity-policy.md) |
| Durable progression grants | Match results and durable rewards need idempotent grant writes. | Currency, unlocks, account progression, profile progression. | [progression-and-rewards.md](domains/gameplay/progression-and-rewards.md) and [match-outcomes-and-results.md](domains/gameplay/match-outcomes-and-results.md) |

## Parked But Accepted

| Item | State | Reopen When | Owner |
| --- | --- | --- | --- |
| Prediction/reconciliation layer | Parked | packet budget/protocol work proves client prediction is needed. | [realtime-protocol-architecture.md](protocol/realtime-protocol-architecture.md) |
| Local play packaging | Parked | local distribution becomes a release target. | [Build Release And Environment Matrix](domains/technical/build-release-and-environment-matrix.md) |
| Hosted game-server deployment | Parked | online multiplayer moves from local/dev to hosted. | [Build Release And Environment Matrix](domains/technical/build-release-and-environment-matrix.md) |
| Matchmaking or room discovery metadata | Parked | public rooms or non-direct-join flows are planned. | [Matchmaking And Room Discovery](domains/platform/matchmaking-and-room-discovery.md) |

## Routed Gameplay Areas

| Area | Owner Doc | Notes |
| --- | --- | --- |
| Weapons and loadouts | [player-build-and-loadouts.md](domains/gameplay/player-build-and-loadouts.md) | Route weapon selection, loadout, and equip ownership here. |
| Enemy, boss, encounter, asteroid variant, and drop behavior | [enemies-bosses-and-encounters.md](domains/gameplay/enemies-bosses-and-encounters.md) | Keep simulation, encounter, and enemy-behavior planning in the owner doc. |
| Damage/effect presentation and player-facing feedback | [player-experience-systems.md](domains/gameplay/player-experience-systems.md) | Route feedback, presentation, and player-visible outcome work here. |
| Progression rewards and reward-bearing drops | [progression-and-rewards.md](domains/gameplay/progression-and-rewards.md) | Use this doc for reward flow and durable progression routing. |
| Inventory and hangar ownership/acquisition | [inventory-and-hangar.md](domains/gameplay/inventory-and-hangar.md) | Keep acquisition and ownership rules in the inventory owner doc. |

## Routed Platform And Progression Areas

| Area | Owner Doc | Notes |
| --- | --- | --- |
| Account product surface | [API Product Surface](protocol/api-product-surface.md) | Route exposed account-facing surface work here instead of backlog bullets. |
| Leaderboards | [Leaderboards And Rankings](domains/platform/leaderboards-and-rankings.md) | Keep ranking and board ownership in the dedicated stub doc. |
| Currency and economy | [progression-and-rewards.md](domains/gameplay/progression-and-rewards.md) | Use this doc for durable reward flow and economy routing. |
| Rewards | [progression-and-rewards.md](domains/gameplay/progression-and-rewards.md) | Route reward-bearing progression here. |
| Inventory | [inventory-and-hangar.md](domains/gameplay/inventory-and-hangar.md) | Keep ownership and acquisition in the inventory owner doc. |
| Account identity | [account-and-identity-systems.md](domains/platform/account-and-identity-systems.md) | Route identity, linking, and admission-adjacent ownership here. |

## Routed Technical Areas

| Area | Owner Doc | Notes |
| --- | --- | --- |
| Realtime protocol | [realtime-protocol-architecture.md](protocol/realtime-protocol-architecture.md) | Keep protocol ownership here instead of in backlog items. |
| Network observability and packet budget | [network-observability-and-packet-budget.md](domains/technical/network-observability-and-packet-budget.md) | Route packet sizing, measurement, and observability work here. |
| Testing and smoke strategy | [Verification And Quality Gates](domains/technical/verification-and-quality-gates.md) | Keep smoke-test and verification planning in the owner doc. |
| Build Release And Environment Matrix | [Build Release And Environment Matrix](domains/technical/build-release-and-environment-matrix.md) | Route packaging and deployment details here. |
