---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7b03-9fda-29dd122f7df8
document_type: general
policy_exempt: false
summary: Client service documentation lives here.
---
# Client

Parent index: [Services](../!INDEX.md)

Client service documentation lives here.

## Ownership

This folder owns docs for the client runtime and its implementation responsibility.

## Does Not Belong

- Domain flow docs.
- Planning docs.
- Direct code maps outside this service index.
- Stub content as canonical service authority.

## Direct Files
<!-- doc-ledger:files:start -->

- [auth-session-flow.md](auth-session-flow.md) - Client auth session flow documentation.
- [client-http-api-flow.md](client-http-api-flow.md) - Client shared HTTP API helper, request-result, auth API, profile API, and local profile API flow documentation.
- [client-logging.md](client-logging.md) - Canonical client envelope, emitter validation/status, operation traces, compatibility emission, and rolling local JSONL output.
- [hud-and-gameplay-ui.md](hud-and-gameplay-ui.md) - Client HUD and gameplay UI documentation.
- [input-and-targeting.md](input-and-targeting.md) - Client input and targeting documentation.
- [menu-flow.md](menu-flow.md) - Client high-level menu flow documentation.
- [team-presentation-and-configuration.md](team-presentation-and-configuration.md) - Client team configuration controls and resolved team presentation.
<!-- doc-ledger:files:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->

- [app-shell-and-session](app-shell-and-session/!INDEX.md) - App Shell And Session documentation.
- [gameplay-event-presentation](gameplay-event-presentation/!INDEX.md) - Gameplay Event Presentation documentation.
- [gameplay-menu-flow](gameplay-menu-flow/!INDEX.md) - Gameplay Menu Flow documentation.
- [gameplay-runtime](gameplay-runtime/!INDEX.md) - Gameplay Runtime documentation.
- [lobby-flow](lobby-flow/!INDEX.md) - Lobby Flow documentation.
- [match-end-flow](match-end-flow/!INDEX.md) - Match End Flow documentation.
- [networking-flow](networking-flow/!INDEX.md) - Networking Flow documentation.
- [pregame-menu-flow](pregame-menu-flow/!INDEX.md) - Pregame Menu Flow documentation.
- [presentation-flow](presentation-flow/!INDEX.md) - Presentation Flow documentation.
- [spectate-flow](spectate-flow/!INDEX.md) - Spectate Flow documentation.
- [world-sync](world-sync/!INDEX.md) - World Sync documentation.
<!-- doc-ledger:folders:end -->

## Stub Files

<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Related Docs

- [Services index](../!INDEX.md)

## Notes

This index stays at the client service boundary and does not try to describe broader domain flows. Client observability ownership is documented in [Client Logging](client-logging.md); presentation consumers within the client branch read `RealtimePresentationState` after runtime routing, and the service index only summarizes that seam at a high level.