# Client

Client service documentation lives here.

## Ownership

This folder owns docs for the client runtime and its implementation responsibility.

## Does Not Belong

- Domain flow docs.
- Planning docs.
- Direct code maps outside this service index.
- Stub content as canonical service authority.

## Direct Files

- [client-http-api-flow.md](client-http-api-flow.md) - Client shared HTTP API helper, request-result, auth API, profile API, and local profile API flow documentation.
- [auth-session-flow.md](auth-session-flow.md) - Client auth session flow documentation.
- [hud-and-gameplay-ui.md](hud-and-gameplay-ui.md) - Client HUD and gameplay UI documentation.
- [input-and-targeting.md](input-and-targeting.md) - Client input and targeting documentation.
- [menu-flow.md](menu-flow.md) - Client high-level menu flow documentation.

## Direct Folders

- [App Shell And Session](app-shell-and-session/!README.md) - Client app entry, shell/session composition, boot flow, room session state, shutdown, and client config documentation.
- [Gameplay Runtime](gameplay-runtime/!README.md) - Client gameplay runtime composition, state application, lifecycle, and processing documentation.
- [Gameplay Event Presentation](gameplay-event-presentation/!README.md) - Client gameplay event presentation, visual effects, local event handoff, and gameplay audio documentation.
- [Presentation Flow](presentation-flow/!README.md) - Client non-HUD gameplay presentation, local player presentation, background presentation, and viewport presentation documentation.
- [Networking Flow](networking-flow/!README.md) - Client WebSocket connection, packet routing, packet dispatch, and outbound packet sending documentation.
- [Lobby Flow](lobby-flow/!README.md) - Client multiplayer lobby session, room entry, join dialog, and lobby presentation documentation.
- [match-end-flow](match-end-flow/!README.md) - Client match-end orchestration and match-results presentation documentation.
- [gameplay-menu-flow](gameplay-menu-flow/!README.md) - Client gameplay menu and match-over overlay menu documentation.
- [World Sync](world-sync/!README.md) - Client world-state rendering, ViewAnchor, visual coordinates, and entity sync documentation.
- [pregame-menu-flow](pregame-menu-flow/!README.md) - Client pregame menu implementation flows for local pilot selection and profile readout.

## Related Docs

- [Services index](../!README.md)

## Notes

This index stays at the client service boundary and does not try to describe broader domain flows.
