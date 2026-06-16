# Domain Backlog

## Backlog Boundary

This file tracks planning and backlog items only.

Implemented behavior belongs in docs/design, docs/client, docs/server, docs/api, or docs/devtools.

Current limitations belong in [docs/limits/current-system-limits.md](../limits/current-system-limits.md) and [docs/limits/player-build-limits.md](../limits/player-build-limits.md).

Detailed player build and loadout planning belongs in [docs/planning/player-build-and-loadouts.md](player-build-and-loadouts.md).

Roadmap phase sequencing belongs in [docs/planning/platform-and-progression-roadmap.md](platform-and-progression-roadmap.md).

## Combat Systems

### Weapons

- Add additional weapon profiles.
- Add client equip and presentation flows.
- Add focused tests for new fire and profile rules.

### Damage

- Add client render events for damage presentation.
- Add area falloff rules.
- Extend DoT into broader status-effect handling.
- Add richer presentation and telemetry for damage outcomes.

### Radial Effects

- Add shockwave or knockback payloads.
- Expand hazardous fields.
- Add status-effect payloads.
- Add falloff rules.
- Add richer presentation events.
- Add additional radial weapons.

### Drop Tables

- Add multi-drop tables with more than one table entry.
- Add additional drop table definitions for other source types.
- Add a minimum drop count policy if needed.
- Add explicit per-source routing.
- Add client-facing presentation polish for drop events.

### Asteroid Variants

- Add per-variant stats behavior through `stats_profile`.
- Add per-variant drop behavior through `drop_table`.
- Add rare variant weighting through lower nonzero weights.

## Player Data And Progression

### Player-Data Pipeline

- Add `-player-data` to the data-sync domain set.
- Add likely generated player-data outputs.
- Add Rails migration skeleton generation.
- Add embedded DB migration skeleton generation.
- Add player-data contract tests.
- Add schema-drift enforcement for player-data contracts.

### Service Boundaries

- Add `services/player-data-server` extraction if the in-process runtime is split.
- Define the player-data service contract from the shared logical schema.
- Add SQLite-backed persistence and migrations for the extracted local service.
- Make the game-server consume player-data service APIs for loadout and profile reads plus match-result writes.
- Make the client consume player-data service APIs for local profile UI.
- Add admission package and routing matrix tests.
- Add room mode and session identity fields.
- Add behavior-preserving admission wiring where needed.
- Handle the Local Profile rename in `services/player-data-server`.
- Add store contract tests.
- Add local profile schema migration/versioning.

### Auth And Account Routing

- Rails token verification hardening.
- Go auth client hardening.
- Websocket auth handshake hardening.
- Enforce online multiplayer admission.
- Expand OAuth support.
- Add JWT only if selected later.
- Harden game-server auth integration.
- Client token storage.

### Progression Grants

- Add live progression grant transport.
- Add an internal HTTP grant path from the game-server to the owning player-data service.
- Add authenticated-account grant transport to `services/api-server`.
- Add local-profile grant transport to extracted `services/player-data-server` if that service exists.
- Make grant writes idempotent with `grant_id` or `event_id`.
- Decouple durable rewards from end-of-match summary handling.

### Account Product Surface

- Account linking or local-to-online migration.
- Online leaderboards.
- Anti-cheat/trust policy.
- Currency.
- Ship parts.
- Rare drops.
- Unlock tokens.
- Account-affecting rewards.

## Client Presentation

### Weapon And Equipment UI

- Add weapon UI.
- Add equip presentation.
- Add player-build and loadout UI once the build model exists.

### Damage And Effect Presentation

- Add client render events for damage presentation.
- Add radial-effect presentation.
- Add richer gameplay effect presentation where tied to implemented server events.

### Devtools Pickup Rendering

- Devtools pickup selector should share the same presentation/catalog source as client pickup rendering when implemented.

## Infrastructure And Deployment

### Local Packaging

- Add local play packaging that may launch or bundle the Go game server with the Godot client.

### Hosted Multiplayer

- Add hosted online game-server deployment using the room/websocket structure.
- Add matchmaking or room discovery metadata if selected later.

### Networking And Prediction

- Add prediction/reconciliation as an explicit separate client layer if added.
- Keep prediction separate from authoritative game rules.

### Smoke And Verification

- Add full gameplay/network smoke testing hardening.
- Review world-dimension balance for gameplay.
