---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-70b6-8566-2d94e5417e63
document_type: general
policy_exempt: false
summary: Client gameplay runtime documentation lives here.
---
# Gameplay Runtime

Parent index: [Client](../!INDEX.md)

Client gameplay runtime documentation lives here.

## Ownership

- Client gameplay runtime composition.
- Lane-native gameplay application.
- Gameplay-session lifecycle and reset.
- Realtime-to-presentation bridge ownership.
- Frame-coalesced presentation fanout.
- Per-frame runtime processing.

## Does Not Belong

- Server simulation authority.
- Packet schema ownership.
- World entity rendering and interpolation details.
- HUD, menu, input, targeting, match-end, and profile detail docs.
- Future planning.

## Direct Files
<!-- doc-ledger:files:start -->

- [gameplay-session-lifecycle.md](gameplay-session-lifecycle.md) - Client gameplay-session activation, presentation lifecycle, reset, replay, and session-exit behavior.
- [gameplay-state-application.md](gameplay-state-application.md) - Lane packet application, readiness, and presentation fanout.
- [presentation-bridge.md](presentation-bridge.md) - Direct applied-packet notification handling, frame-coalesced presentation, readiness-gated flush, and presentation orchestration.
- [runtime-composition.md](runtime-composition.md) - Client gameplay runtime construction, bridge wiring, and composition ownership.
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
- [Gameplay packets](../../../protocol/gameplay-packets.md) - Gameplay packet documentation.

## Notes

Runtime docs describe the client gameplay runtime seam after packets are classified, while protocol and data docs own packet schema authority. Gameplay composition owns the realtime-to-presentation handoff, frame-coalesced fanout, and downstream presentation orchestration, and this index stays at that seam without duplicating ownership.