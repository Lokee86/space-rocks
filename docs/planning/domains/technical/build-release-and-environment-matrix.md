---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7c73-9680-8ddabeccdfba
document_type: general
policy_exempt: false
summary: This doc plans the build, release, environment, packaging, and deployment-shape seam for Space Rocks.
---
# Build Release And Environment Matrix

Parent index: [Technical Planning](./!INDEX.md)

## Purpose

This doc plans the build, release, environment, packaging, and deployment-shape seam for Space Rocks.

It defines which build and environment shapes are expected to exist, what each shape is allowed to use, and what must be true before a build shape is considered usable for testing or production.

## Overview

This matrix keeps local, dev-hosted, staging, and production build shapes separated so packaging, storage, devtools, and release gates stay aligned.

## Current status

The local packaged single-player alpha build shape is implemented for Windows and macOS. Native release-gate jobs export the Godot client, bundle the local-only server and platform credential helper, complete a deterministic server-authoritative single-player match, verify the resolved result and local stats across a complete restart, and upload the exact tested artifacts. The Windows artifact has passed this gate locally; the macOS implementation is exercised by the native macOS job. Manual promotion smoke is limited to visible presentation and ordinary player-driven interaction.

## Ownership Boundary

This doc owns planning for build shapes, release gates, local/dev/hosted environments, packaging boundaries, embedded SQLite versus non-embedded builds, devtools availability by build type, and deployment assumptions.

It should stay on release shape, environment matrix, and packaging/deployment constraints rather than account policy, gameplay rules, realtime protocol design, or detailed hosting-provider setup.

## Does Not Belong

* Detailed protocol format design.
* Detailed player-data schema design.
* Matchmaking or room-discovery policy.
* Account, identity, or abuse-enforcement policy.
* Hosting-provider-specific infrastructure instructions.
* Current implementation authority.

## Build And Environment Shapes

| Shape                             | Role                                                                                              |
| --------------------------------- | ------------------------------------------------------------------------------------------------- |
| Local Development                 | Fast local iteration with editor/dev runners, local services, debug tools, and local data stores. |
| Local Packaged Single-Player Alpha | First release-shaped testing target for local single-player packaging.                            |
| Dev-Hosted Multiplayer            | Multiplayer testing target using hosted or semi-hosted services before full staging readiness.    |
| Hosted Staging                    | Production-like hosted validation environment.                                                    |
| Hosted Production                 | Public online production environment.                                                             |

## Local Development

Local development exists for fast iteration.

It may use:

* Godot editor client.
* Local Go game-server runner.
* Local Rails API server.
* Local player-data service.
* Local Postgres where needed.
* Embedded SQLite where useful.
* Debug flags, devtools, overlays, and local-only shortcuts.

Local development is not a releasable build shape and does not need production constraints.

## Local Packaged Single-Player Alpha

Local packaged single-player is the first release-shaped testing target.

These builds are alpha/testing builds, not production builds. They may keep devtools open or enabled for diagnostics, bug reporting, and testing.

Local packaged single-player should not require hosted auth, hosted multiplayer, matchmaking, or durable online services.

It should include:

* Packaged Godot client.
* Separately packaged local server variant.
* Local profile support.
* Embedded SQLite support.
* Local-only server binding.
* Single-player match start/end flow.
* Local results and stats persistence.

Game logic should remain server-owned. There are no plans to duplicate authoritative game simulation inside the client.

## Local Single-Player Server Packaging

Packaged single-player should bundle a separate server variant with separate build flags or configuration.

Expected differences may include:

* API/auth disabled, stubbed, or locally restricted.
* Embedded SQLite enabled.
* Multiplayer admission disabled.
* Public room discovery disabled.
* Local-only bind required.
* Packaged process lifecycle controlled by the client.
* Server simulation retained as the authoritative runtime.

The bundled server should be locally locked. The likely default is binding only to `127.0.0.1` or localhost, rejecting non-loopback access, and disabling general multiplayer routes.

Local locking is meant to prevent the packaged single-player server from becoming a LAN or Internet multiplayer server. It is not meant to secure the process against the local machine owner.

## Dev-Hosted Multiplayer

Dev-hosted multiplayer is a multiplayer testing environment that may use hosted or semi-hosted services before full staging exists.

It may use:

* Exported or development client builds.
* Hosted or semi-hosted game-server.
* Hosted or semi-hosted API server.
* Non-production databases.
* Real or test auth configuration.
* Explicit debug, admin, or test affordances.

Current controlled alpha multiplayer intentionally leaves runtime server devtools available to alpha testers and there is no plan to gate them during alpha. This is a non-production testing exception, not a public-player capability or production security model.

Dev-hosted multiplayer is not production and should not replace hosted staging. It should not be treated as durable player-data authority, leaderboard authority, production release readiness, or public launch infrastructure.

## Hosted Staging

Hosted staging is the production-like validation environment.

It should use:

* Hosted API server.
* Hosted game-server.
* Hosted player-data path.
* Hosted database.
* Real auth integration where practical.
* Production-like environment variables.
* Observability and diagnostic logging.
* Compatibility and admission checks.

Hosted staging exists to prove that the hosted online stack works before production.

## Hosted Production

Hosted production is the public online environment.

It should require:

* Production client/export.
* Hosted API server.
* Hosted game-server.
* Durable player-data path.
* Real auth.
* Strict compatibility admission.
* No embedded SQLite for multiplayer.
* No production client devtools capability.
* Strongly gated server admin/devtool paths.
* Inbound devtools commands must be disabled or server-authorized and integration-tested.
* Rollback and recovery expectations.

Production should block incompatible clients before login or gameplay access.

## Storage Policy

| Shape                             | SQLite Policy |
| --------------------------------- | ------------- |
| Local Development                 | Allowed.      |
| Local Packaged Single-Player Alpha | Allowed.      |
| Dev-Hosted Multiplayer            | Not allowed.  |
| Hosted Staging                    | Not allowed.  |
| Hosted Production                 | Not allowed.  |

SQLite must not be used in multiplayer builds.

Local packaged single-player may use embedded SQLite for local profiles and local persistence. Multiplayer builds should use the normal hosted or configured durable data path instead.

## Devtools Policy

| Shape                             | Devtools Policy                                    |
| --------------------------------- | -------------------------------------------------- |
| Local Development                 | Enabled/open.                                      |
| Local Packaged Single-Player Alpha | Likely enabled/open for testing.                   |
| Dev-Hosted Multiplayer            | Controlled alpha may be open to testers; post-alpha must be gated. |
| Hosted Staging                    | Strongly gated.                                    |
| Hosted Production Client          | Disabled completely.                               |
| Hosted Production Server          | Strongly gated; some tools may become admin tools. |

Production client builds must disable devtools capability entirely, including the ability to send devtools packets. This must be stronger than hiding UI toggles.

Production server builds may retain strongly gated dev/admin tooling because some devtools may later double as admin tools. These paths must require explicit authorization and must not be reachable through normal player capability.

### Alpha Devtools Exception And Post-Alpha Release Gate

Current controlled alpha multiplayer intentionally leaves server runtime devtools available to alpha testers. There is no requirement to add player authorization or disable inbound devtools commands for alpha testing.

This exception applies only to controlled, non-production alpha environments. It must not be carried into public post-alpha multiplayer.

Before any public post-alpha multiplayer release, inbound server devtools commands must either be disabled entirely or require explicit server-side developer/admin authorization that ordinary players cannot obtain.

Hiding client UI, removing hotkeys, or relying on a package-level build helper is not sufficient unless the inbound networking route actually enforces the gate. The current `nodevtools` helper must not be treated as satisfying this release gate until inbound command routing consults it and integration tests prove devtools commands are rejected.

This is a release blocker, not an alpha blocker.

## Bot And TAS Policy

Production may support explicitly flagged bot/TAS-friendly accounts, runs, or environments.

Bot/TAS production lanes must be deliberate, auditable, and separable from normal trust, rankings, rewards, and eligibility where required.

Account flags can mark accounts allowed to use bot/TAS behavior. Run flags can mark specific matches or results as bot/TAS-involved. Both may be needed.

## Release Gates

Release gates define the minimum pass/fail bar for a build shape.

They are not all the same. Local development only needs basic sanity checks. Local packaged single-player needs packaging and local-runtime verification. Hosted multiplayer targets need service, auth, compatibility, and operational checks.

The local packaged single-player alpha gate verifies:

Automated on the native target operating system:

* Client exports successfully from a clean staged project copy.
* The matching local-server and credential-helper executables are present in the package.
* The bundled server binds only to loopback; normal packages default to `127.0.0.1:8080` and the gate uses an isolated loopback port.
* The client starts and owns the bundled server process.
* The server health endpoint becomes ready.
* DPAPI or Keychain credential save/load/delete works through the exported client.
* Local profile creation and default selection work.
* A real single-player room starts through the normal WebSocket and realtime transport seams.
* The gate assigns a deterministic score, exhausts the player's lives, and reaches the server-owned `game_over` state.
* The resolved match result identifies the active match and expected player score.
* Local game, score, high-score, and ship-death statistics persist across a complete client/server restart.
* Quitting the exported client cleans up the bundled server process.
* The package completes without hosted auth, API, matchmaking, or multiplayer services.
* The exact tested artifact receives a versioned SHA-256 manifest and is uploaded by CI.

Manual before alpha promotion while visual presentation remains intentionally outside the headless gate:

* The package launches normally with visible graphics and working controls.
* A player-driven single-player match can be played without the deterministic smoke controls.
* The visible results screen presents the completed match and persisted statistics correctly.

The implementation entry point is `tools/release/package_local_alpha.py`; the promotion checklist is `tools/release/local_alpha_manual_smoke.md`.

## Packaging Expectations

Local packaged single-player should be packaged as a client plus locally locked server runtime.

Hosted multiplayer builds should not depend on embedded SQLite or local-only server behavior.

Production packages should separate development-only and production-capable behavior through build flags, configuration, or equivalent hard gates.

## Deployment Assumptions

Deployment-provider choice is not decided here.

This doc should keep enough structure to support future decisions about hosted API, hosted game-server, hosted player-data, staging, production, rollback, and release promotion without committing to a specific infrastructure provider too early.

## Implementation sequence

1. Finalize the local packaged single-player alpha gate and keep it aligned with local packaging constraints.
2. Lock the local packaged single-player server to local-only behavior with the existing embedded storage policy.
3. Define dev-hosted multiplayer constraints so hosted services can be used without becoming production assumptions.
4. Keep hosted staging production-like for auth, compatibility, and observability checks.
5. Tighten hosted production around compatibility admission, devtool gating, the post-alpha inbound devtools release gate, and rollback/recovery expectations.

## Related docs

* [Planning](../../!INDEX.md)
* [Development Roadmap](../../development-roadmap.md)
* [Network Observability And Packet Budget](network-observability-and-packet-budget.md)
* [Website And Web Presence](../web/website-and-web-presence.md)
* [Matchmaking And Room Discovery](../platform/matchmaking-and-room-discovery.md)
* [Account And Identity Systems](../platform/account-and-identity-systems.md)
* [Game Integrity Policy](../platform/security-and-admin/game-integrity-policy.md)
* [Shop, Commerce, And Economy](../gameplay/shop-commerce-and-economy.md)

## Open decisions

* Which visible presentation checks, if any, are stable enough to automate without brittle UI testing?
* Which server devtools become admin tools?
* How are bot/TAS flags represented in results, telemetry, leaderboards, and reward eligibility?

## Notes

Keep this doc focused on build and environment boundaries rather than provider-specific deployment instructions.
