# App Shutdown Flow

Parent index: [App Shell And Session](./!README.md)

## Purpose

This document describes the client app shutdown flow owned by the app-shell/session boundary.

It covers how the Godot client handles OS/window close requests, delegates graceful WebSocket close, and hands final process exit to `SceneTree.quit()`.

## Overview

The client shutdown flow is intentionally narrow.

`AppEntry` disables Godot's automatic quit acceptance during startup, then routes OS/window close notifications through `AppShutdownController`.

The current shutdown sequence is:

```text
AppEntry._ready()
-> get_tree().set_auto_accept_quit(false)
-> AppShutdownController.new()
-> AppShutdownController.configure(connection_service, scene_tree)

OS/window close request
-> AppEntry._notification(NOTIFICATION_WM_CLOSE_REQUEST)
-> AppShutdownController.request_shutdown()
-> ClientConnectionService.begin_graceful_close()
-> NetworkClient.begin_graceful_close()
-> WebSocketPeer.close(1000, "client closed")
-> SceneTree.quit()
```

Shutdown is a client process-exit path. It is not the same as returning from gameplay to the main menu, returning to the lobby, replaying a match, leaving a room, clearing session state, or navigating between menu screens.

The shutdown controller requests a graceful WebSocket close only when the connection service reports an open server connection. It then asks the Godot scene tree to quit.

## Code root

```text
client/scripts/shell/app_entry.gd
client/scripts/session/app_shutdown_controller.gd
client/scripts/networking/client_connection_service.gd
client/scripts/networking/network_client.gd
```

## Responsibilities

The app shutdown flow owns:

* disabling Godot automatic quit acceptance at app-entry startup
* catching OS/window close requests at the root app-entry node
* configuring shutdown with the active client connection service and scene tree
* checking whether the client is connected before requesting WebSocket close
* delegating graceful WebSocket close to the networking layer
* handing final process exit to `SceneTree.quit()`
* providing a fallback direct `get_tree().quit()` path if the shutdown controller is unavailable

## Does not own

The app shutdown flow does not own:

* main-menu route behavior
* main-menu quit button wiring
* gameplay menu quit-to-main-menu behavior
* match-results quit-to-main-menu behavior
* lobby leave behavior
* return-to-lobby behavior
* return-to-pregame behavior
* session context clearing
* shell boot request clearing
* room membership state
* gameplay cleanup/reset behavior
* server room cleanup policy
* packet parsing or packet routing
* client viewport config
* WebSocket URL selection
* auth logout behavior
* profile persistence

Those behaviors are owned by focused menu, lobby, gameplay, boot/session, networking, auth, or profile flows.

## Domain roles

### App-entry shutdown route

`AppEntry` is the root owner of the OS/window close notification route.

During `_ready()`, it disables automatic quit acceptance:

```text
get_tree().set_auto_accept_quit(false)
```

That lets `AppEntry._notification()` receive `NOTIFICATION_WM_CLOSE_REQUEST` and route shutdown through the configured shutdown controller.

### Shutdown controller

`AppShutdownController` owns the shutdown request procedure.

It stores:

```text
connection_service
tree
```

When shutdown is requested, it:

1. checks whether a connection service exists and is connected
2. calls `connection_service.begin_graceful_close()` when connected
3. calls `tree.quit()` when a scene tree was configured

The controller does not await full socket closure. It begins the close handshake, then exits through the scene tree.

### Networking close handoff

`ClientConnectionService.begin_graceful_close()` delegates to `NetworkClient.begin_graceful_close()`.

`NetworkClient.begin_graceful_close()` closes the underlying `WebSocketPeer` with:

```text
normal close code: 1000
reason: "client closed"
```

It also marks the connection as gracefully closing so normal connection-closed signal handling is not emitted as an unexpected disconnect path.

### Fallback quit path

If `AppEntry` receives a close notification before `app_shutdown_controller` is available, it calls:

```text
get_tree().quit()
```

That fallback does not request WebSocket close. It exists only as a defensive exit path for startup or configuration failure cases.

## Protocols and APIs

The app shutdown flow does not define a packet protocol.

It uses the client networking API surface:

```text
ClientConnectionService.is_server_connected()
ClientConnectionService.begin_graceful_close()
NetworkClient.begin_graceful_close()
WebSocketPeer.close()
```

The shutdown close uses WebSocket normal close code `1000`.

The current shutdown path does not send a gameplay, lobby, auth, or leave-room packet before quitting. Room leave and room cleanup are separate lobby/server concerns.

## Data ownership

The shutdown flow owns no durable data.

`AppShutdownController` only stores runtime references to:

```text
connection_service
SceneTree
```

Networking owns socket state. Room/session flows own room and boot state. Auth/profile flows own auth and profile state.

## Code map

### Primary implementation files

```text
client/scripts/shell/app_entry.gd
client/scripts/session/app_shutdown_controller.gd
```

### Networking handoff files

```text
client/scripts/networking/client_connection_service.gd
client/scripts/networking/network_client.gd
```

### Related app/session files

```text
client/scripts/boot/session_boot_controller.gd
client/scripts/session/client_session_context.gd
client/scripts/session/room_session_controller.gd
client/scripts/session/gameplay_session_controller.gd
```

These related files are not shutdown owners. They are listed because shutdown is composed beside session boot, room session, and gameplay session in `AppEntry`.

### Important non-ownership boundaries

```text
client/scripts/ui/menus/main_menu.gd
client/scripts/shell/gameplay_menu_flow.gd
client/scripts/gameplay/match_end/match_end_flow.gd
client/scripts/ui/match_results/match_results_flow.gd
client/scripts/lobby/lobby_return_flow.gd
```

These files own menu, match-end, or lobby route behavior. They should not be documented as app shutdown owners.

## Tests

There are no focused unit tests for `AppShutdownController` in the current client test suite.

Related tests cover adjacent session and route behavior:

```text
client/tests/unit/test_shell_boot_flow.gd
client/tests/unit/test_session_network_controller.gd
client/tests/unit/test_room_session_controller.gd
client/tests/unit/shell/test_gameplay_menu_flow.gd
client/tests/unit/gameplay/match_end/test_match_end_flow.gd
client/tests/unit/ui/match_results/test_match_results_flow.gd
```

These tests do not directly verify OS/window close shutdown or graceful close delegation.

## Related docs

* [App Shell And Session](./!README.md)
* [App Entry Composition](app-entry-composition.md)
* [Session Boot And Network Target](session-boot-and-network-target.md)
* [Room Session State](room-session-state.md)
* [Client Viewport Config Flow](client-viewport-config-flow.md)
* [Client](../!README.md)
* [Menu Flow](../menu-flow.md)
* [Networking Flow](../networking-flow/!README.md)
* [Lobby Flow](../lobby-flow/!README.md)
* [Gameplay Runtime](../gameplay-runtime/!README.md)
* [Gameplay Menu Flow](../gameplay-menu-flow/!README.md)
* [Match End Flow](../match-end-flow/!README.md)

## Notes

The current main-menu quit button calls `get_tree().quit()` directly from `client/scripts/ui/menus/main_menu.gd`. That is a separate direct UI quit path, not the OS/window close path documented here.

If the project later requires all quit paths to request graceful WebSocket close, the main-menu quit button should route through an app-level shutdown callback instead of calling the scene tree directly.

`NetworkClient.close_gracefully()` exists as an awaitable close path with a short timeout, but `AppShutdownController` currently uses `begin_graceful_close()` and does not await socket closure before quitting.
