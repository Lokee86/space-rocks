# Systems Plan Index

This is the broad index for major future system plans under `docs/planning`.

Use this page to find the owning planning doc for a system area.

- Phase sequencing belongs in [platform-and-progression-roadmap.md](platform-and-progression-roadmap.md).
- Unscheduled backlog notes belong in [domain-backlog.md](domain-backlog.md).
- Detailed system ownership belongs in the system-specific planning docs linked below.

### Player-facing Game Systems

- [Player Experience Systems](player-experience-systems.md) - umbrella map for the player-facing flow.
- [Modes And Match Rules](modes-and-match-rules.md) - mode resolution, objective policy, scoring policy, and match-end rules.
- [Match Outcomes And Results](match-outcomes-and-results.md) - trusted match outcome and post-match result handoff.
- [Progression And Rewards](progression-and-rewards.md) - XP, rank, currency, unlocks, and reward grants.
- [Achievements And Milestones](achievements-and-milestones.md) - definitions, trusted facts, and milestone grants.
- [Inventory And Hangar](inventory-and-hangar.md) - owned ships, weapons, modules, and build eligibility inputs.
- [Enemies, Bosses, And Encounters](enemies-bosses-and-encounters.md) - enemies, bosses, waves, and encounter structure.
- [Player Build And Loadouts](player-build-and-loadouts.md) - ship variants, weapons, modules, eligibility, and loadout selection.

### Platform / Backend / Account Systems

- [Account And Identity Systems](account-and-identity-systems.md) - guest, local profile, and authenticated account planning.
- [Player Data And Persistence](player-data-and-persistence.md) - shared contracts, local SQLite, Rails/Postgres, and service extraction.
- [Leaderboards And Rankings](leaderboards-and-rankings.md) - leaderboard eligibility, rankings, and persistence.
- [Matchmaking And Room Discovery](matchmaking-and-room-discovery.md) - matchmaking, discovery metadata, and room creation.
- [API Product Surface](api-product-surface.md) - API-owned account and profile surfaces.
- [Anti-Cheat And Trust Policy](anti-cheat-and-trust-policy.md) - trust boundaries for progression and rankings.
- [Platform And Progression Roadmap](platform-and-progression-roadmap.md) - phase order, dependency order, and decision gates for platform and progression work.

### Technical Foundation / Operations Systems

- [Network Observability And Packet Budget](network-observability-and-packet-budget.md) - gameplay packet budget and packet-size visibility.
- [Realtime Protocol Architecture](realtime-protocol-architecture.md) - realtime protocol lanes, snapshots, and codec direction.
- [Logging And Diagnostics](logging-and-diagnostics.md) - future logging and diagnostics planning.
- [Devtools And Telemetry](devtools-and-telemetry.md) - internal telemetry and debug-only presentation.
- [Testing And Smoke Strategy](testing-and-smoke-strategy.md) - smoke coverage and verification planning.
- [Deployment And Packaging](deployment-and-packaging.md) - local packaging and hosted deployment shape.
- [Data Sync And Ssot Pipeline](data-sync-and-ssot-pipeline.md) - shared generator and schema sync pipeline.
- [Domain Backlog](domain-backlog.md) - cross-cutting backlog detail for combat expansion, player-data pipeline work, infrastructure, deployment, and verification follow-up.

### Notes

- `platform-and-progression-roadmap.md` coordinates sequencing and gates, not detailed ownership.
- `player-build-and-loadouts.md` is the system-specific home for build and loadout planning.
- `domain-backlog.md` is the parking lot for unscheduled items and cross-cutting backlog notes that do not yet have a dedicated owner doc.
