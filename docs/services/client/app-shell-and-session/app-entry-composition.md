# App Entry Composition

Parent index: [App Shell And Session](./!README.md)

## Purpose

This document describes the client app-entry composition boundary implemented by `client/scenes/game.tscn` and `client/scripts/shell/app_entry.gd`.

It explains how the Godot client root scene wires the app shell, session controllers, menu flow, auth flow, profile providers, networking, room state, gameplay session, background presentation, shutdown handling, and startup camera state.

## Overview

`client/scenes/game.tscn` is the configured client runtime scene. Its root node uses `app_entry.gd`, which acts as the client composition root for app-shell and session ownership.

`AppEntry` does not own the detailed behavior of the systems it wires. It creates and connects the major runtime controllers, provides shared dependencies, and routes high-level intents between menu, session, room, gameplay, auth, and shutdown seams.

The app-entry boot sequence currently performs this composition:

```text
Game scene root
-> AppEntry._ready()
-> SessionBootController and ClientConfigController
-> AppShutdownController
-> API and auth clients
-> AuthSessionController
-> Profile stats provider
-> BackgroundController
-> GameplaySessionController
-> SessionNetworkController
-> RoomSessionController
-> MainMenuSessionController
-> MenuFlowController
-> MultiplayerEntryFlow
-> main-menu and auth signal wiring
-> saved-token initialization
-> ViewAnchor camera activation
```

The root scene owns the stable node anchors used by the composed flows:

```text
UserInterface
GameplayUserInterface
MainMenu
HUD
Player
ViewAnchor
Bullets
Asteroids
Pickups
RepeatedBackground
RepeatedForegroundBackground
RepeatedPlanetBackground
```

`AppEntry` passes those anchors to focused controllers. It does not directly parse packets, simulate gameplay, own room authority, own profile persistence, or build menu internals.

## Code root

* `client/scenes/game.tscn`
* `client/scripts/shell/app_entry.gd`
* `client/scripts/boot/`
* `client/scripts/session/`
* `client/scripts/main_menu/`
* `client/scripts/ui/menu_flow/`

## Responsibilities

* Act as the Godot client composition root.
* Disable automatic quit acceptance so shutdown can route through the app shutdown controller.
* Create and configure the session boot controller.
* Create and configure the client config sender.
* Create and configure graceful shutdown handling.
* Create shared API, auth, and profile-provider dependencies.
* Connect auth state changes into main-menu presentation.
* Attach auth session state to the connection service for websocket authentication.
* Configure background presentation against the `ViewAnchor`.
* Configure gameplay session composition with scene node anchors and session dependencies.
* Configure network-session routing for connection, room, gameplay, debug, pause, and auth packets.
* Configure room-session state and room-to-gameplay providers.
* Configure main-menu session requests into session boot.
* Configure menu-flow routing and pregame callbacks.
* Configure multiplayer entry gating through auth state.
* Connect top-level main-menu signals.
* Initialize saved auth token state during startup.
* Make the `ViewAnchor/Camera2D` camera current at boot.
* Route gameplay replay and return-to-pregame requests back into menu/session entry points.

## Does not own

* Gameplay simulation.
* Gameplay packet parsing.
* Gameplay runtime state application.
* World synchronization.
* HUD presentation internals.
* Gameplay menu behavior.
* Match-end orchestration internals.
* Match-result row building.
* Lobby room authority.
* Room membership authority.
* Websocket transport internals.
* Packet codec behavior.
* HTTP contract ownership.
* Durable auth/account state.
* Local profile persistence.
* Player-data persistence.
* Backend token validation.
* Server-side admission policy.
* Devtools command authority.

## Domain roles

### Composition root

`AppEntry` is the root composition seam for the running Godot client.

It creates controller instances, adds node-owned controllers where needed, configures `RefCounted` controllers, and wires signals between the runtime flows.

### Scene anchor provider

`game.tscn` provides the long-lived scene nodes that runtime flows need. `AppEntry` resolves them through `@onready` fields and passes them to focused systems.

Important anchors include:

```text
MainMenu
UserInterface
GameplayUserInterface
RepeatedBackground
RepeatedForegroundBackground
RepeatedPlanetBackground
Player
ViewAnchor
Bullets
Asteroids
Pickups
HUD
```

### Session boot owner handoff

`AppEntry` creates `SessionBootController`, then delegates single-player, create-room, and join-room requests through `MainMenuSessionController`.

The actual boot request state and websocket target selection belong to the boot/session flow, not to `AppEntry`.

### Menu route executor

Menu-facing scripts emit intent. `AppEntry` wires those intents into session and menu actions.

Examples:

```text
MainMenu.single_player_requested
-> AppEntry._on_single_player_requested()
-> MenuFlowController.show_single_player_pregame()

Pregame Play Endless
-> MenuFlowController callback
-> AppEntry._start_single_player_from_pregame()
-> MainMenuSessionController.request_single_player()

Pregame Create Room
-> MenuFlowController callback
-> AppEntry._request_create_room_from_pregame()
-> MainMenuSessionController.request_create_room()

Pregame Join Room
-> MenuFlowController callback
-> AppEntry._request_join_room_from_pregame()
-> MainMenuSessionController.request_join_room()
```

### Auth composition owner

`AppEntry` creates:

```text
ApiHttpClient
AuthApiClient
AuthSessionController
PlayerDataProfileApiClient
ProfileStatsProvider
```

It configures `AuthSessionController`, connects auth signals, gives the connection service access to auth state, and shares `ProfileStatsProvider` with menu/profile presentation.

Auth state persistence and API behavior remain owned by the auth/API flows.

### Gameplay session handoff

`AppEntry` creates and configures `GameplaySessionController`.

It passes gameplay scene anchors and shared session dependencies to the gameplay session, then listens for high-level gameplay-session route requests:

```text
return_to_pregame_requested
replay_requested
```

The gameplay session controller owns gameplay packet gating, gameplay composition, input routing, reset behavior, and gameplay route consequences after a session starts.

### Room/session handoff

`AppEntry` creates `RoomSessionController` and connects it to:

```text
MainMenu
UserInterface
ClientSessionContext
ClientConnectionService
ShellBootFlow
ClientConfigController
```

It also exposes room state, match result, and room max-player provider callables to `GameplaySessionController`.

Room state caching and lobby presentation belong to `RoomSessionController` and lobby flows.

### Network signal router composition

`AppEntry` creates `SessionNetworkController`, configures it with the connection service and shell boot flow, then connects connection, room, and gameplay signals.

The network controller owns dispatching received connection-service signals to room and gameplay session owners. `AppEntry` only installs the wiring.

### Shutdown composition

`AppEntry` disables automatic quit acceptance and routes `NOTIFICATION_WM_CLOSE_REQUEST` into `AppShutdownController`.

The shutdown controller owns graceful-close request behavior and final tree quit.

## Protocols and APIs

`AppEntry` does not define a wire protocol or HTTP API. It composes the runtime surfaces that consume those protocols.

The composed protocol/API participants are:

```text
AuthApiClient
ApiHttpClient
PlayerDataProfileApiClient
ClientConnectionService
SessionNetworkController
RoomSessionController
GameplaySessionController
```

### Startup auth flow

During startup, `AppEntry` calls:

```text
AuthSessionController.initialize_from_saved_token()
```

That call triggers saved-token loading and `/api/auth/me` validation through the auth flow.

`AppEntry` only starts the flow and updates main-menu signed-in/signed-out presentation when auth state changes.

### Websocket authentication handoff

`AppEntry` attaches the auth session controller to `ClientConnectionService`:

```text
connection_service.set_auth_session_controller(auth_session_controller)
```

When the websocket connects, the connection service can send `authenticate_request` if an auth token exists. Multiplayer boot gating remains owned by `SessionNetworkController` and `ShellBootFlow`.

### Boot request handoff

`AppEntry` routes user intent to `MainMenuSessionController`, which delegates to `SessionBootController`.

The boot/session flow owns pending request state for:

```text
single-player start
create room
join room
```

`AppEntry` does not send these packets directly.

## Data ownership

`AppEntry` owns composition-time references, not durable data.

Client app-entry state includes references to:

```text
session_boot_controller
main_menu_session_controller
session_network_controller
room_session_controller
gameplay_session_controller
client_config_controller
app_shutdown_controller
auth_session_controller
api_http_client
player_data_profile_api_client
profile_stats_provider
auth_api_client
background_controller
menu_flow_controller
multiplayer_entry_flow
```

`AppEntry` does not persist or authoritatively own:

```text
auth tokens
account identity
local profiles
profile stats
room membership
room state
match results
gameplay state
websocket identity
server player identity
client config contracts
generated packet constants
```

The only data-shaped decision `AppEntry` currently extracts directly is the selected single-player local profile id before requesting a single-player boot:

```text
MenuFlowController.get_single_player_context()
-> identity_kind == "local_profile"
-> local_profile_id
-> MainMenuSessionController.request_single_player(local_profile_id)
```

That is a handoff from menu/profile presentation into session boot. Local profile persistence remains outside `AppEntry`.

## Code map

### App entry and root scene

* `client/scenes/game.tscn`
* `client/scripts/shell/app_entry.gd`

### Boot and network target composition

* `client/scripts/boot/session_boot_controller.gd`
* `client/scripts/boot/shell_boot_flow.gd`
* `client/scripts/boot/pending_boot_request.gd`
* `client/scripts/boot/session_network_target.gd`

### Session controllers

* `client/scripts/session/client_session_context.gd`
* `client/scripts/session/session_network_controller.gd`
* `client/scripts/session/room_session_controller.gd`
* `client/scripts/session/gameplay_session_controller.gd`
* `client/scripts/session/client_config_controller.gd`
* `client/scripts/session/app_shutdown_controller.gd`

### Main menu and route composition

* `client/scripts/main_menu/main_menu_session_controller.gd`
* `client/scripts/ui/menu_flow/menu_flow_controller.gd`
* `client/scripts/ui/menu_flow/multiplayer_entry_flow.gd`
* `client/scripts/ui/menu_flow/pregame_menu_flow.gd`
* `client/scripts/ui/menu_flow/transmission_flow.gd`

### Auth, API, and profile dependencies

* `client/scripts/api/api_http_client.gd`
* `client/scripts/auth/auth_api_client.gd`
* `client/scripts/auth/auth_session_controller.gd`
* `client/scripts/profile/player_data_profile_api_client.gd`
* `client/scripts/profile/profile_stats_provider.gd`
* `client/scripts/profile/profile_context_provider.gd`

### Connection and packet dispatch dependencies

* `client/scripts/networking/client_connection_service.gd`
* `client/scripts/networking/network_client.gd`
* `client/scripts/networking/inbound/server_packet_dispatcher.gd`
* `client/scripts/networking/inbound/server_packet_router.gd`
* `client/scripts/networking/outbound/client_packet_sender.gd`

### Gameplay composition dependencies

* `client/scripts/gameplay/gameplay_composition.gd`
* `client/scripts/gameplay/runtime/gameplay_flow_composer.gd`
* `client/scripts/gameplay/runtime/gameplay_runtime_context.gd`
* `client/scripts/gameplay/state/gameplay_state_flow.gd`
* `client/scripts/shell/gameplay_shell_flow.gd`
* `client/scripts/shell/gameplay_menu_flow.gd`
* `client/scripts/shell/gameplay_hud_flow.gd`
* `client/scripts/shell/gameplay_runtime_tick_flow.gd`

### Background and presentation dependencies

* `client/scripts/presentation/background/background_controller.gd`
* `client/scripts/presentation/background/background_flow.gd`

### Related generated/source files

* `client/scripts/generated/constants/constants.gd`
* `client/scripts/generated/networking/packets/packets.gd`
* `shared/constants/client/shell.toml`
* `shared/constants/client/lobby.toml`
* `shared/packets/lobby.toml`
* `shared/packets/gameplay.toml`
* `shared/packets/outputs.toml`

## Tests

Relevant tests include:

```text
client/tests/unit/ui/menu_flow/test_app_entry_menu_flow.gd
client/tests/unit/ui/menu_flow/test_menu_flow_controller.gd
client/tests/unit/ui/menu_flow/test_multiplayer_entry_flow.gd
client/tests/unit/test_pending_boot_request.gd
client/tests/unit/test_shell_boot_flow.gd
client/tests/unit/test_session_network_controller.gd
client/tests/unit/boot/test_session_network_target.gd
client/tests/unit/test_room_session_controller.gd
client/tests/unit/test_gameplay_session_controller.gd
client/tests/unit/test_auth_session_controller.gd
client/tests/unit/ui/menus/test_main_menu_auth_state.gd
```

These tests verify the app-entry route wiring, shared profile-provider composition, signed-in and signed-out multiplayer routing, pending boot request selection, session-network gating, room-session behavior, gameplay-session handoff, and auth/menu presentation state.

## Related docs

* [App Shell And Session](./!README.md)
* [Client](../!README.md)
* [Client Menu Flow](../menu-flow.md)
* [Auth Session Flow](../auth-session-flow.md)
* [Client HTTP API Flow](../client-http-api-flow.md)
* [Networking Flow](../networking-flow/!README.md)
* [Lobby Flow](../lobby-flow/!README.md)
* [Pregame Menu Flow](../pregame-menu-flow/!README.md)
* [Gameplay Runtime](../gameplay-runtime/!README.md)
* [Gameplay Menu Flow](../gameplay-menu-flow/!README.md)
* [Match End Flow](../match-end-flow/!README.md)
* [World Sync](../world-sync/!README.md)

## Notes

Legacy documentation correctly identified that `AppEntry` and session-level owners execute route changes requested by result-window, game-menu, and menu-flow intents. This document rewrites that fact from the current implementation and keeps `AppEntry` at the composition boundary.

`AppEntry` is intentionally broad because it is the root composition seam. New feature logic should usually live in a focused controller or flow and be wired here only when it needs root-scene anchors, shared dependencies, or top-level route/session callbacks.

`AppEntry` currently creates several `RefCounted` controllers without adding them as child nodes. Node-based controllers that need processing, notifications, or scene-tree ownership are added as children. This distinction is implementation detail, but it matters when adding new runtime behavior.

The app-entry boundary should not become a substitute for service docs covering session boot, room session state, shutdown/config, networking, gameplay runtime, lobby flow, or auth flow. Those docs own their detailed behavior.
