# API Product Surface

Parent index: [Protocol Planning](./!INDEX.md)

## Purpose

This document owns the planned API product surface map for Space Rocks.

It is planning-owned and non-current. The current surface map lives in [API Product Surface](../../protocol/api-product-surface.md), and exact HTTP shape continues to live in [HTTP API Contracts](../../protocol/http-api-contracts.md).

## Overview

This document tracks future API product surface ownership and the planned split between likely consumers, hosts, and behavior owners.

It does not define exact request or response shape, and it does not replace the current protocol doc.

## Current Status

The API product surface split is established:

- current surface inventory lives in current protocol docs
- planned surface ownership lives here
- exact HTTP shape remains owned by OpenAPI and HTTP contract docs

This planning doc is intentionally non-current.

## Decisions Made

- The current protocol doc owns current surface mapping only.
- Planned API product surfaces should live in planning docs until they become current.
- Exact HTTP shape stays with OpenAPI and the HTTP contract docs.
- Future product surfaces should stay separate from implementation detail and service code maps.

## Open Decisions

- Which future planned surfaces become current first.
- Which planned surfaces require dedicated service ownership docs before implementation.
- Which planned surfaces should split into smaller planning docs once they gain detailed behavior.
- Which planned surfaces remain deferred until the product direction is confirmed.

## Expected Ownership

| Surface | Status | Consumer | Host | Behavior owner | Detail docs |
| --- | --- | --- | --- | --- | --- |
| Room browser / matchmaking discovery | Planned | Client lobby and room discovery UI | TBD | TBD | TBD |
| Shop / wallet / purchase surface | Planned | Client commerce and ownership UI | TBD | TBD | TBD |
| Inventory / hangar / loadout surface | Planned | Client collection and setup UI | TBD | TBD | TBD |
| Leaderboards / rankings surface | Planned | Client ranking and comparison UI | TBD | TBD | TBD |
| Social / community surface | Planned | Client social and community UI | TBD | TBD | TBD |
| Website / account portal surface | Planned | Web account and support portal | TBD | TBD | TBD |
| Admin / support / moderation surface | Planned | Internal admin and support tooling | TBD | TBD | TBD |

## Implementation Sequence

1. Keep the current protocol surface map authoritative for implemented HTTP and API behavior.
2. Grow a dedicated current service or protocol doc only when a planned surface becomes real.
3. Move a planned row into current protocol docs when the surface is implemented.
4. Add or update service docs for the owning runtime once implementation begins.
5. Keep OpenAPI and service docs aligned when planned surfaces become current.

## Related Docs

- [Protocol Planning](./!INDEX.md)
- [API Product Surface](../../protocol/api-product-surface.md)
- [HTTP API Contracts](../../protocol/http-api-contracts.md)
- [Player Data HTTP API](../../protocol/player-data-http-api.md)
- [Auth And OAuth](../../services/api-server/auth-and-oauth.md)
- [Internal API Surface](../../services/api-server/internal-api-surface.md)
- [Player Stats And Match Results](../../services/api-server/player-stats-and-match-results.md)
- [Runtime And Health](../../services/api-server/runtime-and-health.md)
- [Client HTTP API Flow](../../services/client/client-http-api-flow.md)
- [Local Profiles HTTP API](../../services/player-data/local-profiles-http-api.md)
- [Profile Stats Flow](../../services/player-data/profile-stats-flow.md)
- [Match Result Sinks](../../services/player-data/match-result-sinks.md)
- [Route Composition](../../services/game-server/process/route-composition.md)

## Notes

This planning doc stays focused on future surface ownership and avoids exact endpoint shape.

When a planned surface becomes current, move its authoritative surface mapping into the current protocol doc and keep its detailed behavior in the owning service docs.
