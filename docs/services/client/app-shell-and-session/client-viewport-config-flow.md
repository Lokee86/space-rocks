# Client Viewport Config Flow

Parent index: [App Shell And Session](./!INDEX.md)

## Purpose

This document describes the client viewport config flow owned by the app-shell/session boundary.

It covers how the Godot client reports its visible viewport size to the realtime server through the `client_config` packet.

## Overview

The client viewport config flow sends the current visible viewport size to the game server after session boot and when the viewport size changes.

The current flow is:

```text
AppEntry._setup_boot_and_config()
-> ClientConfigController.configure(connection_service, viewport)
-> ClientViewportConfigFlow.configure(connection_service, viewport)
-> Viewport.size_changed connects to ClientViewportConfigFlow.send_client_config

ShellBootFlow sends a pending boot request
-> boot_request_sent
-> ClientConfigController.send_client_config()
-> ClientViewportConfigFlow.send_client_config()
-> Packets.client_config_packet(width, height)
-> ClientConnectionService.send_packet(packet)
```

Room entry also sends config when a room snapshot enters `InGame`:

```text
RoomSessionController.handle_room_snapshot(packet)
-> room_state == InGame
-> client_config_sender.call()
-> ClientConfigController.send_client_config()
```

The packet payload is the viewport's visible rectangle size:

```text
visible_world_width
visible_world_height
```

The client sends config only when connected. If there is no connection service or the server is not connected, `ClientViewportConfigFlow.send_client_config()` returns without sending.

The server uses the config as runtime visibility/camera information for the current game player. It is not generic client settings, account preferences, graphics settings, or app configuration.

## Code root

```text
client/scripts/session/client_config_controller.gd
client/scripts/config/client_viewport_config_flow.gd
client/scripts/shell/app_entry.gd
client/scripts/session/room_session_controller.gd
```

## Responsibilities

The client viewport config flow owns:

* creating the viewport config flow from the session/app-shell boundary
* storing the active connection-service reference
* storing the active Godot viewport reference
* connecting viewport resize notifications to config sending
* reading `Viewport.get_visible_rect().size`
* building the generated `client_config` packet
* sending the packet through `ClientConnectionService.send_packet()`
* sending config after a boot request is sent
* sending config when a room snapshot enters `InGame`
* ignoring sends while disconnected
* logging successful config sends through client shell debug logging

## Does not own

The client viewport config flow does not own:

* WebSocket URL selection
* session mode selection
* pending boot request type
* room membership authority
* lobby state authority
* server camera behavior
* world rendering
* background presentation
* window min/max size constants
* graphics settings
* user preferences
* auth/account settings
* local profile settings
* app shutdown
* packet schema source-of-truth
* packet codec behavior
* networking transport internals

Those responsibilities belong to boot/session, room/lobby, networking, protocol/data, rendering/presentation, auth/profile, or shutdown docs.

## Domain roles

### App-shell config sender

`AppEntry` creates `ClientConfigController` during boot/config setup.

It configures the controller with:

```text
SessionBootController.get_connection_service()
get_viewport()
```

`ClientConfigController` is a thin session-facing wrapper. It creates `ClientViewportConfigFlow` and forwards `send_client_config()` calls to it.

### Viewport resize observer

`ClientViewportConfigFlow.configure()` connects the Godot viewport's `size_changed` signal to `send_client_config()`.

That makes runtime window/viewport resizing resend the latest visible size to the server.

The flow does not calculate presentation layout. It only reports the viewport size used by server-side visibility/camera logic.

### Boot-time config sender

`AppEntry` connects `ShellBootFlow.boot_request_sent` to `ClientConfigController.send_client_config()`.

That means config is sent after a pending session boot request is consumed and sent to the server.

The boot flow owns pending request state. The viewport config flow only sends the current viewport size after boot request emission.

### Room-state config sender

`RoomSessionController` accepts a `client_config_sender` callable.

When `handle_room_snapshot()` applies a snapshot and the resulting room state is `InGame`, it calls the configured sender.

This gives the server viewport/camera config at the room-to-gameplay transition.

### Connection-guarded packet sender

`ClientViewportConfigFlow.send_client_config()` sends nothing unless:

```text
connection_service != null
connection_service.is_server_connected()
```

When connected, it builds:

```text
Packets.client_config_packet(viewport_size.x, viewport_size.y)
```

and sends it through:

```text
connection_service.send_packet(packet)
```

## Protocols and APIs

The viewport config flow consumes the generated client packet builder:

```text
Packets.client_config_packet(visible_world_width, visible_world_height)
```

The generated packet shape is sourced from:

```text
shared/packets/gameplay.toml
```

Current packet shape:

```text
type = "client_config"
config.visible_world_width = <float>
config.visible_world_height = <float>
```

The client sends the packet through the normal outbound packet path:

```text
ClientViewportConfigFlow
-> ClientConnectionService.send_packet(packet)
-> ClientPacketSender.send_packet(packet)
-> NetworkClient.send_raw_packet(packet)
-> PacketCodec.encode(packet)
-> WebSocketPeer.send_text(...)
```

Server-side routing treats `client_config` as a gameplay packet type that can be handled after room/game player routing exists.

At the server summary level:

```text
inbound.HandleGameplayPacket()
-> room.GameInstance().HandlePacket(currentGamePlayerID, packet)
-> Game.HandlePacket()
-> session.Config = packet.Config
-> cameraView.SetConfig(packet.Config)
-> player.SetConfig(packet.Config), when the active player exists
```

The detailed server implementation is owned by game-server service docs and realtime protocol docs.

## Data ownership

The client viewport config flow does not own durable data.

It reads transient runtime data from:

```text
Viewport.get_visible_rect().size
```

It sends that size to the server as packet data.

The source-of-truth for packet shape is:

```text
shared/packets/gameplay.toml
```

Generated client output includes:

```text
client/scripts/generated/networking/packets/packets.gd
```

Generated server/runtime output includes:

```text
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
```

The server owns the authoritative runtime storage and use of `ClientConfig` for player session, camera view, and active ship config.

## Code map

### Primary client implementation files

```text
client/scripts/session/client_config_controller.gd
client/scripts/config/client_viewport_config_flow.gd
```

### Client composition and trigger files

```text
client/scripts/shell/app_entry.gd
client/scripts/boot/shell_boot_flow.gd
client/scripts/session/room_session_controller.gd
```

### Client networking files

```text
client/scripts/networking/client_connection_service.gd
client/scripts/networking/outbound/client_packet_sender.gd
client/scripts/networking/network_client.gd
client/scripts/networking/packets/packet_codec.gd
```

### Generated and source files

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
client/scripts/generated/networking/packets/packets.gd
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
```

### Server consumption files

```text
services/game-server/internal/networking/inbound/gameplay.go
services/game-server/internal/game/input.go
services/game-server/internal/game/runtime/camera.go
services/game-server/internal/game/runtime/ship.go
services/game-server/internal/game/session.go
```

### Important non-ownership boundaries

```text
client/scripts/boot/session_network_target.gd
client/scripts/session/app_shutdown_controller.gd
client/scripts/presentation/background/background_controller.gd
client/scripts/presentation/background/background_flow.gd
client/scripts/world/world_sync.gd
```

These files may use session, viewport, presentation, or world concepts, but they do not own the viewport config packet send flow.

## Tests

There are no focused client unit tests for:

```text
ClientConfigController
ClientViewportConfigFlow
```

Related tests cover adjacent trigger and session behavior:

```text
client/tests/unit/test_shell_boot_flow.gd
client/tests/unit/test_session_network_controller.gd
client/tests/unit/test_room_session_controller.gd
client/tests/unit/boot/test_session_network_target.gd
```

Server-side related tests include:

```text
services/game-server/internal/networking/gameplay_packets_test.go
services/game-server/tests/game/spawning_test.go
services/game-server/tests/game/visibility_test.go
services/game-server/tests/game/movement_test.go
services/game-server/tests/game/collision_test.go
```

Those server tests verify gameplay or visibility behavior that depends on client camera/config values. They do not replace a focused client test for viewport config packet sending.

## Related docs

* [App Shell And Session](./!INDEX.md)
* [App Entry Composition](app-entry-composition.md)
* [Session Boot And Network Target](session-boot-and-network-target.md)
* [Room Session State](room-session-state.md)
* [App Shutdown Flow](app-shutdown-flow.md)
* [Client](../!INDEX.md)
* [Networking Flow](../networking-flow/!INDEX.md)
* [Lobby Flow](../lobby-flow/!INDEX.md)
* [Gameplay Runtime](../gameplay-runtime/!INDEX.md)
* [World Sync](../world-sync/!INDEX.md)

## Notes

This flow is named around viewport config rather than generic client config because the current implementation only reports visible viewport width and height.

`ClientConfigController` is currently a thin wrapper around `ClientViewportConfigFlow`. That wrapper is still useful as the app-shell/session-facing seam because it keeps `AppEntry` and room/session code from depending directly on the lower-level viewport packet implementation.

The client sends viewport config after boot request send and again when a room snapshot reaches `InGame`. This duplication is intentional in the current implementation: boot-time send covers early session setup, while room-state send covers the room-to-gameplay transition.

The config sender silently no-ops while disconnected. That means viewport resize before connection does not queue a later send. The next boot/request or `InGame` room snapshot send is expected to provide the current viewport size.
