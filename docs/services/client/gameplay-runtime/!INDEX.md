# Gameplay Runtime

Parent index: [Client](../!INDEX.md)

Client gameplay runtime documentation lives here.

## Ownership

- Client gameplay runtime composition.
- Normalized gameplay-state application.
- Gameplay-session lifecycle and reset.
- Per-frame runtime processing.

## Does Not Belong

- Server simulation authority.
- Packet schema ownership.
- World entity rendering and interpolation details.
- HUD, menu, input, targeting, match-end, and profile detail docs.
- Future planning.

## Direct Files
<!-- doc-ledger:files:start -->

- [gameplay-session-lifecycle.md](gameplay-session-lifecycle.md) - Client gameplay packet acceptance, reset, replay, and session-exit behavior.
- [gameplay-state-application.md](gameplay-state-application.md) - Client gameplay packet normalization and state fanout flow.
- [runtime-composition.md](runtime-composition.md) - Client gameplay runtime wiring and composition ownership.
- [runtime-processing.md](runtime-processing.md) - Client per-frame gameplay processing order and runtime tick behavior.
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Client](../!INDEX.md)
- [World Sync](../world-sync/!INDEX.md)
- [Gameplay packets](../../../protocol/gameplay-packets.md) - Stub: gameplay realtime packet documentation.

## Notes

Runtime docs describe client presentation and orchestration after packets are classified, while protocol and data docs own packet schema authority.