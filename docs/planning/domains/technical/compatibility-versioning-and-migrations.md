# Compatibility Versioning And Migrations

Parent index: [Technical Planning](./!INDEX.md)

## Purpose

This doc plans cross-cutting compatibility, versioning, and migration policy for Space Rocks.

It defines how old clients, old local profiles, old data schemas, generated contract drift, and service-version mismatches should be handled as the project moves toward packaged builds and hosted multiplayer.

The goal is not long-term backwards compatibility for every old build. The goal is to keep users moving forward to the latest version without corrupting accumulated progress.

## Overview

This doc keeps compatibility rules, versioning policy, and migration gates aligned so release-shaped builds can move forward without breaking progress or contract drift safety.

## Current status

Active planning.

## Ownership Boundary

This doc owns planning for compatibility rules, build admission, versioning policy, migration gates, deprecation timing, removal rules, and cross-surface drift policy.

Protocol details stay in protocol planning. Schema details stay in data and service docs. Release-shape details stay in build and environment planning.

## Does Not Belong

* Exact realtime packet layouts.
* Exact HTTP endpoint schemas.
* Exact database schema definitions.
* Exact migration implementation code.
* Account, auth, or abuse policy.
* Build packaging details.
* Current implementation authority.

## Compatibility Surfaces

Compatibility planning applies to these surfaces:

* Client build version.
* Game-server version.
* API-server version.
* Player-data version.
* Realtime protocol version.
* HTTP/OpenAPI contract version.
* Local profile storage version.
* Hosted player-data schema version.
* Generated contract outputs.
* Saved match results, run records, loadouts, unlocks, and progression data.

These surfaces do not need to version together unless the changed contract requires it.

## Versioning Policy

Each service or contract surface may version independently.

A service version, protocol version, API contract version, player-data schema version, and generated-file version should not be treated as the same thing by default.

Version only the surface that actually changed, as long as compatibility remains intact.

Examples:

* The API server may change internally without changing the HTTP contract version.
* The game-server may change internally without changing the realtime protocol version.
* The client may change UI behavior without changing any hosted service contract.
* A player-data schema change may require a migration without changing the realtime protocol.

## Version Support Policy

| Stage                  | Policy                                                               |
| ---------------------- | -------------------------------------------------------------------- |
| Local Development      | Loose compatibility; coordinated local breakage is acceptable.       |
| Local Packaged Beta    | Best-effort forward migration; no forever-support promise.           |
| Dev-Hosted Multiplayer | Forced upgrades are acceptable.                                      |
| Hosted Staging         | Should follow production compatibility expectations where practical. |
| Hosted Production      | Current version only at first.                                       |

Hosted production should block incompatible clients before login or gameplay access.

Incompatible clients should fail early with an update-required path rather than reaching room, matchmaking, or gameplay flows.

Production may support current-plus-previous compatibility later, but that is not an initial requirement.

## Local Packaged Compatibility

Local packaged single-player should bundle matching client and server pieces.

Compatibility for local packaged single-player is package-controlled. The packaged client should not need to negotiate with arbitrary server versions for normal local play.

If the bundled server, local profile storage, or generated contracts change, the package should include the matching runtime and migration path.

## Realtime Protocol Compatibility

Realtime protocol compatibility should use a connection-level compatibility check first.

The project should not try to support many mixed packet versions early.

Unknown packet types should fail safely. They should not be ignored silently if ignoring them could hide a broken client/server mismatch.

Protocol version should change when the realtime contract changes. Server version and protocol version should not move together unless the server change actually changes the realtime contract.

## HTTP And OpenAPI Compatibility

Before public production, HTTP/API surfaces may change with coordinated client and service updates.

Once hosted production exists, breaking API changes require one of:

* compatibility support,
* a migration path,
* a forced client upgrade,
* or removal from release scope until the break is resolved.

OpenAPI and generated clients should be treated as source-of-truth contract surfaces where applicable.

## Local Profile Migration

Local profile migration should protect existing progress before upgrading.

Local profile upgrades should either create an automatic backup before migration or offer the user a backup option before migration. Automatic backup is the preferred default when practical.

Local profile migration should be forward-only unless a rollback path is explicitly planned.

If migration fails, the game should preserve the old data and fail visibly rather than silently corrupting or overwriting progress.

The long-term goal is to provide repair, export, or equivalent recovery options where practical so users can move to the latest version with accumulated progress intact.

## Hosted Data Migration

Hosted production data migrations require stricter gates than local profile migrations.

Hosted migrations should include:

* backup planning,
* rollback planning where practical,
* migration validation,
* release coordination,
* failure visibility,
* and clear ownership for recovery.

Hosted migration details belong in the relevant data and service docs. This doc owns the cross-cutting policy that hosted data migrations must be treated as release-sensitive operations.

## Generated Contract Drift

Generated contract drift blocks release-shaped builds.

Release-shaped builds include packaged beta/testing builds, dev-hosted multiplayer builds, staging candidates, and production candidates.

Local development remains loose. It does not need to block or warn on every drift case unless a local tool chooses to do so.

Drift examples include:

* OpenAPI output out of sync with source contracts.
* Generated Godot contract files stale.
* Generated Go contract files stale.
* Shared data definitions stale.
* Data-sync output not matching source-of-truth files.

If generated files do not match the relevant source-of-truth, the build should not be considered releasable.

## Deprecation And Removal

Pre-production and beta builds do not receive long-term compatibility support unless explicitly marked.

Hosted production removals should be more deliberate. Removing an old route, packet field, data field, migration path, or generated contract shape should require confirming that no supported build still depends on it.

Old migration support may be removed later, but only after the supported-version policy says it is safe.

## Release Gates

Compatibility gates should apply to release-shaped builds.

A release-shaped build should not pass if:

* client/server compatibility is known to be broken,
* generated contracts are stale,
* required migrations are missing,
* local profile migration risks data loss without backup,
* hosted data migration lacks backup or rollback planning,
* or incompatible production clients can reach normal login/gameplay paths.

Exact checks may vary by build shape.

## Implementation sequence

1. Keep each compatibility surface versioned independently unless a shared contract truly changes.
2. Apply package-controlled compatibility for local packaged single-player builds.
3. Use connection-level checks and safe failure for realtime protocol mismatches.
4. Require backup, rollback, and failure visibility for local profile and hosted data migrations.
5. Block release-shaped builds when generated contracts drift from source-of-truth definitions.
6. Tighten hosted production admission around incompatible clients and unsafe deprecations.

## Related docs

* [Planning](../../!INDEX.md)
* [Development Roadmap](../../development-roadmap.md)
* [Protocol Planning](../../protocol/!INDEX.md)
* [Data Planning](../../data/!INDEX.md)
* [Services Planning](../../services/!INDEX.md)
* [Build Release And Environment Matrix](build-release-and-environment-matrix.md)
* [Realtime Protocol Architecture](../../../protocol/realtime-protocol-architecture.md)

## Open decisions

* What exact backup mechanism should local profile migration use?
* What repair or export UX should failed local profile migrations provide?
* Which generated drift checks are required for each release-shaped build?
* Which hosted migrations require full rollback versus backup-and-forward repair?
* When, if ever, should production support current-plus-previous compatibility?

## Notes

Keep versioning and migration policy cross-cutting; service-specific implementation details belong in the owning service or data docs.
