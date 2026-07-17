# Game Server Observability

Parent index: [Game Server](../!INDEX.md)

Observability documentation for the game server lives here.

## Ownership

This folder owns canonical service diagnostics, logging, and compatibility-surface documentation for the game server.

## Does Not Belong

- Process startup or shutdown.
- WebSocket transport details.
- Room rules or simulation mechanics.
- External integration internals.
- General product or domain notes.

## Direct Files
<!-- doc-ledger:files:start -->

- [logging-and-diagnostics.md](logging-and-diagnostics.md) - Game-server canonical observability emission, trace/reason-code rules, compatibility logging surfaces, and shared servicelog runtime integration.
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

The game-server call-site rollout is complete. Repository-wide bridge retirement and P3C runtime scenarios remain separate follow-up work; this boundary stays focused on canonical diagnostics, compatibility surfaces, and service-level visibility.
