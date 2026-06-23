<!-- documentation-policy-exempt: planned API surface map; intentionally uses the same lightweight map shape as the current API surface map because full planning sections would blur its purpose. -->
# API Product Surface

Parent index: [Protocol Planning](./!INDEX.md)

## Purpose

This document maps planned API product surfaces for Space Rocks.
It is planning-owned and non-current.
Current implemented surfaces live in ../../protocol/api-product-surface.md.
Exact HTTP shape lives in ../../protocol/http-api-contracts.md and shared/contracts/http/openapi.yaml.

## Overview

This doc tracks likely future product API surface areas and their rough ownership.
It does not define endpoints, methods, schemas, request bodies, response bodies, status codes, or implementation flow.

## Current Status

Current API product surfaces are mapped in current protocol docs.
This doc only tracks planned surfaces that are not yet current.
Planned rows should move into the current map only when implemented.

## Planned Surface Map

| Surface | Status | Consumer | Likely host | Behavior owner | Planning docs |
| --- | --- | --- | --- | --- | --- |
| Room browser / matchmaking discovery | Planned | Client lobby and discovery UI | TBD | TBD | [Matchmaking and Room Discovery](../domains/platform/matchmaking-and-room-discovery.md) |
| Matchmaking queue / status | Planned | Client queue and status UI | TBD | TBD | [Matchmaking and Room Discovery](../domains/platform/matchmaking-and-room-discovery.md) |
| Room invite / join-code support | Planned | Client invite and room entry UI | TBD | TBD | [Matchmaking and Room Discovery](../domains/platform/matchmaking-and-room-discovery.md) |
| Leaderboards / rankings | Planned | Client ranking and comparison UI | TBD | TBD | [Leaderboards and Rankings](../domains/platform/leaderboards-and-rankings.md) |
| Match history / recent matches | Planned | Client match history UI | TBD | TBD | TBD |
| Public profile / profile visibility | Planned | Client profile and visibility UI | TBD | TBD | [Social and Community Systems](../domains/platform/social-and-community-systems.md) |
| Friends / blocks / recent players | Planned | Client social management UI | TBD | TBD | [Social and Community Systems](../domains/platform/social-and-community-systems.md) |
| Party invites / party presence | Planned | Client party and presence UI | TBD | TBD | [Social and Community Systems](../domains/platform/social-and-community-systems.md) |
| Reports / moderation intake | Planned | Client report and support flows | TBD | TBD | [Social and Community Systems](../domains/platform/social-and-community-systems.md) |
| Inventory / hangar | Planned | Client collection and setup UI | TBD | TBD | [Inventory and Hangar](../domains/gameplay/inventory-and-hangar.md) |
| Loadout save / load | Planned | Client loadout management UI | TBD | TBD | [Inventory and Hangar](../domains/gameplay/inventory-and-hangar.md) |
| Progression / unlocks | Planned | Client progression UI | TBD | TBD | [Progression and Rewards](../domains/gameplay/progression-and-rewards.md) |
| Wallet / shop / purchase / receipts | Planned | Client commerce and ownership UI | TBD | TBD | [Shop Commerce and Economy](../domains/gameplay/shop-commerce-and-economy.md) |
| Website account portal | Planned | Web account and support portal | TBD | TBD | [Website and Web Presence](../domains/web/website-and-web-presence.md) |
| Admin / support tools | Planned | Internal admin and support tooling | TBD | TBD | TBD |

## Does Not Belong

- exact endpoint paths
- HTTP methods
- request or response schemas
- status-code catalogs
- OpenAPI enforcement details
- service implementation internals
- realtime WebSocket packet details

## Related Docs

- [Protocol Planning](./!INDEX.md)
- [API Product Surface](../../protocol/api-product-surface.md)
- [HTTP API Contracts](../../protocol/http-api-contracts.md)
- [Player Data HTTP API](../../protocol/player-data-http-api.md)
- [Matchmaking and Room Discovery](../domains/platform/matchmaking-and-room-discovery.md)
- [Leaderboards and Rankings](../domains/platform/leaderboards-and-rankings.md)
- [Social and Community Systems](../domains/platform/social-and-community-systems.md)
- [Inventory and Hangar](../domains/gameplay/inventory-and-hangar.md)
- [Progression and Rewards](../domains/gameplay/progression-and-rewards.md)
- [Shop Commerce and Economy](../domains/gameplay/shop-commerce-and-economy.md)
- [Website and Web Presence](../domains/web/website-and-web-presence.md)

## Notes

This doc is a planned-surface map, not an endpoint design.
When a planned surface becomes real, move its current mapping into ../../protocol/api-product-surface.md and keep exact HTTP shape in OpenAPI / HTTP API Contracts.
