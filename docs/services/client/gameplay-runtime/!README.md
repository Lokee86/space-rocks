# Gameplay Runtime

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

- [gameplay-session-lifecycle.md](gameplay-session-lifecycle.md) - Client gameplay packet acceptance, reset, replay, and session-exit behavior.
- [gameplay-state-application.md](gameplay-state-application.md) - Client gameplay packet normalization and state fanout flow.
- [runtime-composition.md](runtime-composition.md) - Client gameplay runtime wiring and composition ownership.
- [runtime-processing.md](runtime-processing.md) - Client per-frame gameplay processing order and runtime tick behavior.

## Stub Files

- None.

## Direct Folders

- None.

## Related Docs

- [Client](../!README.md)
- [World Sync](../world-sync/!README.md)
- [Gameplay packets](../../../protocol/stubs/gameplay-packets.md) - Stub: gameplay realtime packet documentation.
- [Documentation policy](../../../documentation-policy.md)
- [Documentation procedure](../../../documentation-procedure.md)

## Notes

Runtime docs describe client presentation and orchestration after packets are classified, while protocol and data docs own packet schema authority.
