---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7fd3-8c04-a674a7a10752
document_type: general
policy_exempt: false
summary: Player-data service documentation lives here.
---
# Player Data

Parent index: [Services](../!INDEX.md)

Player-data service documentation lives here.

## Ownership

This folder owns docs for the player-data service runtime and its implementation responsibility.

## Permanent Non-Ownership And Dependency Rule

Player-data does not host, import, call, persist for, or reach through to diagnostic-aggregator. It must not import any diagnostic-aggregator package or receive diagnostic handlers, services, stores, or internal types through its constructors. Any future bounded diagnostic report produced by player-data must use an outbound client/transport seam for the diagnostic-report HTTP API; failure on that optional path must not affect player-data behavior or persistence.

## Does Not Belong

- Domain flow docs.
- Planning docs.
- Direct code maps outside this service index.
- Stub content as canonical service authority.

## Direct Files
<!-- doc-ledger:files:start -->

- [local-profiles-http-api.md](local-profiles-http-api.md) - Player-data local profiles HTTP API documentation.
- [match-result-sinks.md](match-result-sinks.md) - Player-data match result sink documentation.
- [observability-and-logging.md](observability-and-logging.md) - Player-data canonical HTTP and dispatcher observability, compatibility layer, and rolling runtime.
- [profile-stats-flow.md](profile-stats-flow.md) - Player-data profile stats flow documentation.
- [runtime-and-store-routing.md](runtime-and-store-routing.md) - Player-data runtime and store routing documentation.
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Services index](../!INDEX.md)
- [Game-server diagnostic-aggregator hosting](../game-server/integrations/diagnostic-aggregator-hosting.md)

## Notes

This index stays at the player-data service boundary and does not attempt to describe broader account or platform planning.