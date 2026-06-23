# Build Release And Environment Matrix

Parent index: [Technical Planning](./!INDEX.md)

## Purpose

This doc plans the build, release, environment, packaging, and deployment-shape seam for Space Rocks.

It defines which build and environment shapes are expected to exist, what each shape is allowed to use, and what must be true before a build shape is considered usable for testing or production.

## Overview

This matrix keeps local, dev-hosted, staging, and production build shapes separated so packaging, storage, devtools, and release gates stay aligned.

## Current status

Active planning.

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
| Local Packaged Single-Player Beta | First release-shaped testing target for local single-player packaging.                            |
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

## Local Packaged Single-Player Beta

Local packaged single-player is the first release-shaped testing target.

These builds are beta/testing builds, not production builds. They may keep devtools open or enabled for diagnostics, bug reporting, and testing.

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
* Rollback and recovery expectations.

Production should block incompatible clients before login or gameplay access.

## Storage Policy

| Shape                             | SQLite Policy |
| --------------------------------- | ------------- |
| Local Development                 | Allowed.      |
| Local Packaged Single-Player Beta | Allowed.      |
| Dev-Hosted Multiplayer            | Not allowed.  |
| Hosted Staging                    | Not allowed.  |
| Hosted Production                 | Not allowed.  |

SQLite must not be used in multiplayer builds.

Local packaged single-player may use embedded SQLite for local profiles and local persistence. Multiplayer builds should use the normal hosted or configured durable data path instead.

## Devtools Policy

| Shape                             | Devtools Policy                                    |
| --------------------------------- | -------------------------------------------------- |
| Local Development                 | Enabled/open.                                      |
| Local Packaged Single-Player Beta | Likely enabled/open for testing.                   |
| Dev-Hosted Multiplayer            | Explicitly gated; testing/admin use allowed.       |
| Hosted Staging                    | Strongly gated.                                    |
| Hosted Production Client          | Disabled completely.                               |
| Hosted Production Server          | Strongly gated; some tools may become admin tools. |

Production client builds must disable devtools capability entirely, including the ability to send devtools packets. This must be stronger than hiding UI toggles.

Production server builds may retain strongly gated dev/admin tooling because some devtools may later double as admin tools. These paths must require explicit authorization and must not be reachable through normal player capability.

## Bot And TAS Policy

Production may support explicitly flagged bot/TAS-friendly accounts, runs, or environments.

Bot/TAS production lanes must be deliberate, auditable, and separable from normal trust, rankings, rewards, and eligibility where required.

Account flags can mark accounts allowed to use bot/TAS behavior. Run flags can mark specific matches or results as bot/TAS-involved. Both may be needed.

## Release Gates

Release gates define the minimum pass/fail bar for a build shape.

They are not all the same. Local development only needs basic sanity checks. Local packaged single-player needs packaging and local-runtime verification. Hosted multiplayer targets need service, auth, compatibility, and operational checks.

The first local packaged single-player beta gate should likely verify:

* Client exports successfully.
* Bundled local server starts.
* Client connects to bundled local server.
* Local profile create/load works.
* Single-player match starts.
* Single-player match ends.
* Results screen appears.
* Local stats persist.
* Quitting cleans up the local server process.
* Hosted auth and multiplayer services are not required.

## Packaging Expectations

Local packaged single-player should be packaged as a client plus locally locked server runtime.

Hosted multiplayer builds should not depend on embedded SQLite or local-only server behavior.

Production packages should separate development-only and production-capable behavior through build flags, configuration, or equivalent hard gates.

## Deployment Assumptions

Deployment-provider choice is not decided here.

This doc should keep enough structure to support future decisions about hosted API, hosted game-server, hosted player-data, staging, production, rollback, and release promotion without committing to a specific infrastructure provider too early.

## Implementation sequence

1. Finalize the local packaged single-player beta gate and keep it aligned with local packaging constraints.
2. Lock the local packaged single-player server to local-only behavior with the existing embedded storage policy.
3. Define dev-hosted multiplayer constraints so hosted services can be used without becoming production assumptions.
4. Keep hosted staging production-like for auth, compatibility, and observability checks.
5. Tighten hosted production around compatibility admission, devtool gating, and rollback/recovery expectations.

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

* What exact mechanism should lock the packaged single-player server to local-only access?
* What exact build flags distinguish the packaged single-player server variant from multiplayer server variants?
* What is the final minimum gate for the first local packaged single-player beta?
* Which server devtools become admin tools?
* How are bot/TAS flags represented in results, telemetry, leaderboards, and reward eligibility?

## Notes

Keep this doc focused on build and environment boundaries rather than provider-specific deployment instructions.
