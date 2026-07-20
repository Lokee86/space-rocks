---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-723f-a4f6-13d5c1e309c6
document_type: general
policy_exempt: false
summary: Process documentation for the game server lives here.
---
# Game Server Process

Parent index: [Game Server](../!INDEX.md)

Process documentation for the game server lives here.

## Ownership

This folder owns executable startup, route composition, dependency construction, and process lifecycle documentation for the game server.

## Does Not Belong

- WebSocket transport details.
- Room rules or lifecycle rules.
- Simulation mechanics.
- External integration internals.
- Logging policy detail beyond the service boundary.

## Direct Files
<!-- doc-ledger:files:start -->

- [route-composition.md](route-composition.md) - Route Composition documentation.
- [service-shutdown.md](service-shutdown.md) - Graceful game-server process shutdown, HTTP draining, service closure, and room cleanup documentation.
- [service-startup.md](service-startup.md) - Service Startup documentation.
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Game Server](../!INDEX.md)
- [Services index](../../!INDEX.md)

## Notes

This folder is intentionally narrow and covers process startup, route composition, dependency construction, and shutdown behavior only.