# Player-Data Schema Source Of Truth

Parent index: [Design Legacy](./!INDEX.md)

This doc defines the shared logical schema contract for account-shaped player data used by Local Profile and Authenticated Account paths.

## Purpose

Local Profile and Authenticated Account must share the same logical account-shaped player-data concepts.

## Problem

Hand-writing Rails schema, embedded DB schema, and Go playerdata structs separately risks schema drift.

## Boundary Summary

- `shared/player_data/*.toml` owns logical player-data contracts.
- It does not own HTTP request/response shapes.
- HTTP request/response shapes live in `shared/contracts/http/openapi.yaml`.
- Rails migrations own the Rails/Postgres physical schema.
- Embedded SQLite physical storage exists in the standard no-tag development build.
- `-tags noembeddedsqlite` deployment/restricted builds omit embedded SQLite and do not include `modernc.org/sqlite`.
- Embedded DB migrations own local physical storage when embedded local storage is enabled.
- Stores must satisfy the logical player-data contract even if physical tables differ.
- See [Project source-of-truth map](source-of-truth-map.md) for the broader ownership map.
- If HTTP API payloads are involved, see [HTTP contracts](../api/http-contracts.md).

## Core Rule

- `shared/player_data/stats.toml` and `shared/player_data/match_result.toml` are the source of truth for logical account-shaped player-data schema.
- `shared/packets/player_data.toml` defines the player-data packet protocol.
- `services/player-data/protocol/packets.go` is generated from the packet SSoT.
- Rails/Postgres physical stats persistence exists for authenticated accounts.
- Embedded SQLite physical stats persistence exists for local profiles in the standard no-tag development build.
- Both stores implement the same logical stats contract.
- Go `MatchResultSummary` structs and builders now exist in the player-data runtime and mirror the shared logical schema.
- Gameplay-facing code depends on playerdata contracts, not Rails tables or embedded DB tables.
- The logical player-data concepts include:
  - Profile
  - Loadout
  - Progression
  - Unlocks
  - Stats
  - MatchResultSummary
- Live simulation state is excluded.

## V1 Stats Contract

The initial logical `Stats` contract is summary-only and intended for match-resolution commits.

V1 stat fields:

- `total_score`
- `high_score`
- `ship_deaths`
- `games_played`
- `wins`
For V1 multiplayer, the winner is the authenticated player with the highest match score.
`wins` is account/multiplayer-only; Local Profile uses the shared core stats fields and intentionally excludes `wins`.

`MatchResultSummary` supports:

- `match_id`
- `mode`
- `player summaries`
- `account_id`
- `local_profile_id`
- `score`
- `ship_deaths`
- `won`

Guest summaries use no durable identity.
Wins remain account/multiplayer-only, and Local Profile excludes wins.

## Read/Write Symmetry

The logical `Stats` contract is used for both writes and reads through the player-data runtime.

- Writes use `RecordMatchResult` through the player-data runtime.
- Reads use `LoadStats` through the player-data runtime.
- The runtime selects the backing store after mode and identity validation.
- Backing store selection is not a client concern.
- The same logical stats payload is normalized for profile display regardless of whether it came from guest memory, SQLite, or Rails/Postgres.
- Guest memory is transient and separate from local profile persistence.
- `ship_deaths` comes from authoritative server match facts, not client-side presentation counting.
- The client must not count or mutate profile stats from game-over presentation.

## Logical Schema Versus Physical Database Schema

This SSoT is for logical player-data contracts, not raw database DDL.

Logical schema examples:

- `PlayerProfile` has `display_name` and profile metadata.
- `PlayerLoadout` has selected ship, primary weapon, secondary weapon, and additional equipment fields.
- `PlayerProgression` has unlocks, milestones, stats, or progress markers.
- `MatchResultSummary` has account/profile relevant match summary fields.

Physical schema examples:

- Rails/Postgres tables, indexes, constraints, migrations.
- Embedded DB tables, indexes, constraints, migrations in the standard no-tag development build.

Physical schemas may differ because Rails/Postgres and the embedded DB may have different storage needs.

Physical schemas must still satisfy the shared logical contract.

Go `MatchResultSummary` structs and builders exist in the player-data runtime, and the game-server reports resolved `MatchResultSummary` through `services/player-data` for both write and read flows.
`services/player-data` routes `RecordMatchResult` and `LoadStats` by identity kind: Authenticated Account uses Rails/Postgres through `RailsStore`, Local Profile uses embedded SQLite in the standard no-tag development build, and Guest uses guest/no-durable behavior.

## Scope

- Logical player-data contracts for the current `Stats` and `MatchResultSummary` shapes.
- Runtime adapters that read and write those logical contracts through `services/player-data`.

## Non-Goals

- Physical table design details.
- Gameplay state ownership outside player-data contracts.
- Transport or packet routing beyond the player-data packet protocol.

## Source Layout

Current logical schema sources:

- `shared/player_data/stats.toml`
- `shared/player_data/match_result.toml`

These files define logical schema contracts and are not physical database schemas.

## Rails/Postgres Boundary

Rails owns online authenticated account persistence.

Rails migrations own the physical Postgres schema.

Rails physical schema should satisfy the shared logical player-data schema.

`account_id` is the authenticated account UUID identity in player-data contracts. Rails `user_id` stays an internal foreign key to `users.id`.
`local_profile_id` is the local profile identity.
`playerdata` core owns the logical contract and receives local-store construction through dependency injection.
`playerdata` core does not import the embedded SQLite package.

Rails should not use raw SQL as the cross-service SSoT.

## Embedded DB Boundary

Embedded DB owns Local Profile persistence.

Embedded DB physical schema may differ from Rails/Postgres.
Embedded DB physical schema is present in the standard no-tag development build.

Embedded DB physical schema must satisfy the same logical player-data contract.

Embedded DB must not store Discord access tokens, Rails bearer tokens as local profile identity, or online account secrets.

## Store Parity Rules

Local Profile and Authenticated Account should expose equivalent conceptual player-data operations.

Store parity rules:

- Profile shape must remain conceptually aligned.
- Loadout shape must remain conceptually aligned.
- Progression/unlocks shape must remain conceptually aligned.
- Match result summary shape must remain conceptually aligned where applicable.

Parity does not require identical physical tables.

Both stores are implementations of the same player-data contract, not sources of independent domain truth.

`NoDurableStore` for Guest may return defaults or reject durable writes, but should not pretend to persist account-shaped data.
Guest uses singleton in-memory unsaved stats.
Authenticated Account uses Rails/Postgres physical stats persistence.
Local Profile uses embedded SQLite physical stats persistence in the standard no-tag development build.
Profile reads are implemented through the data-handler and player-data runtime.

## Related Limits

- [Current System Limits](../limits/current-system-limits.md)
- [Planning Notes](../planning/domain-backlog.md)