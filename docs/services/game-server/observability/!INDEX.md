---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7f47-b6c6-8d333a5fb78f
document_type: general
policy_exempt: false
summary: Observability documentation for the game server lives here.
---
# Game Server Observability

Parent index: [Game Server](../!INDEX.md)

Observability documentation for the game server lives here.

## Ownership

This folder owns canonical service diagnostics and logging documentation for the game server.

## Does Not Belong

- Process startup or shutdown.
- WebSocket transport details.
- Room rules or simulation mechanics.
- External integration internals.
- General product or domain notes.

## Direct Files
<!-- doc-ledger:files:start -->

- [logging-and-diagnostics.md](logging-and-diagnostics.md) - Game-server canonical observability emission, trace/reason-code rules, and shared servicelog runtime integration.
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

The game-server call-site rollout is complete and production emission is canonical-only through `logging.Emit`. Repository-wide bridge retirement remains separate follow-up work for services that still retain compatibility adapters. Runtime scenarios are implemented under `tools/runtime_scenarios/` and documented in the runtime-performance plan; this boundary stays focused on canonical diagnostics and service-level visibility.
