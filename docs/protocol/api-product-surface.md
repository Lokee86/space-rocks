<!-- documentation-policy-exempt: current API surface map; intentionally omits full protocol doc shape because adding policy sections would blur its purpose. -->
# API Product Surface

Parent index: [Protocol](./!INDEX.md)

## Purpose

This document maps current API product surfaces for Space Rocks.

It is a surface inventory, not the owner of exact HTTP request/response shape. Exact HTTP shape belongs to [HTTP API Contracts](./http-api-contracts.md) and `shared/contracts/http/openapi.yaml`. Detailed behavior belongs in the linked service and protocol docs for each surface.

## Overview

This doc is the centralized map for current API product surfaces.

It stays focused on surface ownership and audience, not HTTP schema, request validation, or endpoint shape.

## Current Surface Map

| Surface | Status | Consumer | Host | Behavior owner | Detail docs |
| --- | --- | --- | --- | --- | --- |
| Rails auth/session | Current | Godot auth flow and browser session handling | `services/api-server` | `services/api-server` | [Auth And OAuth](../services/api-server/auth-and-oauth.md), [Client HTTP API Flow](../services/client/client-http-api-flow.md), [HTTP API Contracts](./http-api-contracts.md) |
| Discord OAuth login session | Current | Godot browser-assisted Discord login | `services/api-server` | `services/api-server` | [Auth And OAuth](../services/api-server/auth-and-oauth.md), [Client HTTP API Flow](../services/client/client-http-api-flow.md), [HTTP API Contracts](./http-api-contracts.md) |
| Current authenticated user | Current | Client auth/session and account readout | `services/api-server` | `services/api-server` | [Auth And OAuth](../services/api-server/auth-and-oauth.md), [Client HTTP API Flow](../services/client/client-http-api-flow.md), [HTTP API Contracts](./http-api-contracts.md) |
| Public authenticated player stats | Current | Client account stats readout | `services/api-server` | `services/api-server` | [Player Stats And Match Results](../services/api-server/player-stats-and-match-results.md), [HTTP API Contracts](./http-api-contracts.md) |
| Player profile readout | Current | Client pregame profile flow | `services/game-server` hosting `POST /api/player-data/profile` | `services/player-data` | [Player Data HTTP API](./player-data-http-api.md), [Profile Stats Flow](../services/player-data/profile-stats-flow.md), [Client HTTP API Flow](../services/client/client-http-api-flow.md), [Route Composition](../services/game-server/process/route-composition.md), [HTTP API Contracts](./http-api-contracts.md) |
| Local profile management | Current | Client local pilot flow | `services/game-server` hosting local-profile routes | `services/player-data` | [Local Profiles HTTP API](../services/player-data/local-profiles-http-api.md), [Client HTTP API Flow](../services/client/client-http-api-flow.md), [Profile Stats Flow](../services/player-data/profile-stats-flow.md), [Route Composition](../services/game-server/process/route-composition.md), [HTTP API Contracts](./http-api-contracts.md) |
| Internal token verification | Current | Game-server auth verifier and player-data Rails adapter | `services/api-server` | `services/api-server` | [Internal API Surface](../services/api-server/internal-api-surface.md), [Auth And OAuth](../services/api-server/auth-and-oauth.md), [Route Composition](../services/game-server/process/route-composition.md) |
| Internal authenticated-account stats read | Current | Player-data Rails adapter | `services/api-server` | `services/api-server` | [Internal API Surface](../services/api-server/internal-api-surface.md), [Player Stats And Match Results](../services/api-server/player-stats-and-match-results.md), [Profile Stats Flow](../services/player-data/profile-stats-flow.md) |
| Internal match-result persistence | Current | Player-data Rails adapter and game-server match-result reporting | `services/api-server` | `services/api-server` | [Player Stats And Match Results](../services/api-server/player-stats-and-match-results.md), [Match Result Sinks](../services/player-data/match-result-sinks.md), [Internal API Surface](../services/api-server/internal-api-surface.md) |
| Rails health | Current | Deployment checks and service monitors | `services/api-server` | `services/api-server` | [Runtime And Health](../services/api-server/runtime-and-health.md) |
| Game-server health | Current | Deployment checks and service monitors | `services/game-server` | `services/game-server` process | [Route Composition](../services/game-server/process/route-composition.md) |

## Does Not Belong

- Exact JSON schemas.
- OpenAPI enforcement details.
- Service implementation internals.
- Code maps.
- Test commands.
- Realtime WebSocket packet details.

## Related Docs

- [Protocol](./!INDEX.md)
- [HTTP API Contracts](./http-api-contracts.md)
- [Planning API Product Surface](../planning/protocol/api-product-surface.md)
- [Player Data HTTP API](./player-data-http-api.md)
- [Auth And OAuth](../services/api-server/auth-and-oauth.md)
- [Internal API Surface](../services/api-server/internal-api-surface.md)
- [Player Stats And Match Results](../services/api-server/player-stats-and-match-results.md)
- [Runtime And Health](../services/api-server/runtime-and-health.md)
- [Client HTTP API Flow](../services/client/client-http-api-flow.md)
- [Local Profiles HTTP API](../services/player-data/local-profiles-http-api.md)
- [Profile Stats Flow](../services/player-data/profile-stats-flow.md)
- [Match Result Sinks](../services/player-data/match-result-sinks.md)
- [Route Composition](../services/game-server/process/route-composition.md)

## Notes

This doc is a map, not a replacement for OpenAPI, service docs, or implementation code.

The current rows should stay aligned with the live product surface, but the exact route shapes belong to the contract docs.

Future API surface planning belongs in [Planning API Product Surface](../planning/protocol/api-product-surface.md).
