---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7e16-9711-cfc452b6e940
document_type: general
policy_exempt: false
summary: This document defines the game-server process boundary for temporarily co-hosting the logically independent diagnostic-aggregator service. It records the permitted composition-root hosting exception and the forbidden runtime and domain...
---
# Diagnostic-Aggregator Hosting

Parent index: [Game Server Integrations](./!INDEX.md)

## Purpose

This document defines the game-server process boundary for temporarily co-hosting the logically independent diagnostic-aggregator service. It records the permitted composition-root hosting exception and the forbidden runtime and domain dependencies.

## Overview

`diagnostic-aggregator` is a separate service with its own diagnostic application and report-storage responsibilities. It is temporarily co-hosted by the game-server process for deployment and process-composition reasons. Co-hosting does not make diagnostic-aggregator part of the game-server runtime, gameplay domain, networking domain, rooms, match reporting, or player-data domain.

The sole import exception is:

```text
services/game-server/cmd/game-server/ -> services/diagnostic-aggregator/hosted
```

Only the game-server composition-root hosting adapter under `services/game-server/cmd/game-server/` may import `services/diagnostic-aggregator/hosted`.

No package under `services/game-server/internal/`, no gameplay, networking, rooms, or match-reporting package, and no player-data package may import any diagnostic-aggregator package.

No game-server or player-data runtime/domain code may call diagnostic handlers, report services, stores, or other diagnostic-aggregator internals in-process.

The permitted service flow is:

```text
producer/client
-> HTTP transport/API
-> diagnostic-aggregator handler/application service
-> report store
```

The forbidden flow is:

```text
game-server/player-data runtime
-> diagnostic application service/store/direct handler call
```

The composition root may construct, register, and close the hosted service. That lifecycle authority does not grant report-processing authority or permission to make direct diagnostic calls. The composition root only assembles the process and its transport wiring.

Future detachment changes process composition and service addressing only. Callers continue through the transport/API contract; detachment must not require changing game-server or player-data runtime/domain callers to use diagnostic-aggregator internals.

## Responsibilities

The game-server composition root is responsible for:

- Constructing the hosted diagnostic-aggregator service through its hosting adapter.
- Registering the hosted HTTP transport/API with the process composition.
- Closing the hosted service during process shutdown.
- Supplying process-level configuration and lifecycle wiring needed for co-hosting.

The diagnostic-aggregator service is responsible for:

- Receiving diagnostic requests through its HTTP transport/API.
- Running diagnostic handler and application-service behavior.
- Validating and aggregating diagnostic reports.
- Persisting reports through its report store.

Producers and clients are responsible for sending diagnostic information through the HTTP transport/API contract.

## Code root

`services/game-server/`

## Does not own

Game-server integration hosting does not own:

- Diagnostic report-processing policy.
- Diagnostic handlers or application services.
- Diagnostic report stores or persistence internals.
- Gameplay simulation, networking, rooms, or match-reporting behavior.
- Player-data runtime/domain behavior or player-data storage.
- A direct in-process diagnostic call path.
- A static architecture guard; this document records the boundary and intended dependency rule, but does not claim that an automated guard already exists.

## Domain roles

The game-server composition root is a process host and lifecycle coordinator. It may construct, register, and close diagnostic-aggregator, but it does not become the diagnostic report-processing authority. Diagnostic-aggregator remains authoritative for diagnostic handlers, application services, validation, report construction, and report storage. Game-server gameplay, networking, rooms, match reporting, and player-data remain separate domain owners.

## Data ownership

Diagnostic-aggregator owns diagnostic report processing, finalized report data, report identifiers, retention, and report-store data. Producers own the diagnostic material and correlation context they submit. The game server owns gameplay and process data; player-data owns player-data aggregates and persistence. Co-hosting does not transfer these ownership responsibilities.

## Protocols and APIs

The integration boundary is the diagnostic-aggregator HTTP transport/API. Producers and clients send requests to that API, and the diagnostic-aggregator handler/application service processes them before writing to the report store.

Game-server and player-data code must remain callers through that transport/API contract when they participate in a permitted producer flow. They must not import or call diagnostic handlers, application services, stores, or internals directly in-process.

Co-hosting may use an in-process HTTP server or equivalent process-local transport registration, but process-local addressing does not change the API boundary or authorize direct package calls.

## Code map

Hosting exception:

- `services/game-server/cmd/game-server/` - composition-root hosting adapter; the only game-server location permitted to import `services/diagnostic-aggregator/hosted`.
- `services/diagnostic-aggregator/hosted` - hosted-service adapter used for temporary process co-hosting.

Forbidden dependency surfaces:

- `services/game-server/internal/` - no diagnostic-aggregator imports.
- Game-server gameplay, networking, rooms, and match-reporting packages - no diagnostic-aggregator imports or direct calls.
- Player-data packages - no diagnostic-aggregator imports or direct calls.

The exact handler, application-service, and report-store implementation paths belong to the diagnostic-aggregator service and are not game-server integration dependencies.

## Tests

Tests for this boundary should verify the HTTP transport/API behavior and service lifecycle wiring without granting runtime or domain code a direct diagnostic dependency. Co-hosting tests may verify construction, registration, request routing, and close behavior at the composition root.

This document does not claim that a static architecture guard already exists. Any future dependency check should enforce the import and direct-call rules stated here without treating co-hosting as domain ownership.

## Related docs

- [Game Server Integrations](./!INDEX.md)
- [Game Server](../!INDEX.md)
- [Services index](../../!INDEX.md)
- Diagnostic-aggregator service documentation, when available

## Notes

Co-hosting is a deployment arrangement, not an ownership merger. The game-server composition root may host the service, but diagnostic-aggregator remains logically independent and retains authority over diagnostic report processing and storage.
