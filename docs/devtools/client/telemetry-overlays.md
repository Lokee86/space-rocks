## Telemetry Overlays

Parent index: [Client](./!INDEX.md)

## Purpose

This document describes the client-side devtools telemetry overlay surface.

It covers the world telemetry overlay, the metrics it displays, how it receives gameplay and network timing data, how it stays separate from production HUD behavior, and which client and server paths participate in the diagnostic loop.

## Overview

Telemetry overlays are debug-only client presentation for live runtime diagnostics. In this context, telemetry means local development readouts and packet timing diagnostics, not analytics, tracing, or durable metrics storage.

The current client telemetry overlay is the world telemetry overlay. It is a `CanvasLayer` panel that shows glanceable world, client, and network metrics while a gameplay session is running. It is separate from:

* player-facing HUD
* match result UI
* gameplay menu UI
* raw devtools-window state packet readouts
* server-authoritative devtools commands

The overlay observes normalized gameplay state after the client has read the authoritative `state` packet. It counts server-owned entity dictionaries, adds local client frame timing, and optionally sends diagnostic `telemetry_ping` packets while visible so the server can respond with `telemetry_pong`.

The high-level flow is:

```text
state packet
-> GameplayStatePacketReader
-> GameplayDevtoolsContext
-> DevtoolsOverlayContext
-> WorldTelemetryContext
-> WorldTelemetryOverlayFlow
-> WorldTelemetryMetrics
-> WorldTelemetryOverlay label

overlay visible
-> WorldTelemetryContext sends telemetry_ping
-> game server replies telemetry_pong
-> ClientConnectionService emits telemetry_pong_received
-> NetworkTelemetryMetrics updates RTT and server clock offset
-> overlay displays network timing fields
```

The overlay does not mutate gameplay. All gameplay facts shown in the overlay come from server-owned state or client-observed packet timing.

## Debug-only scope

Telemetry overlays are development tooling.

They are allowed to:

* display live world counts from normalized gameplay state
* display local client frame timing
* display network timing derived from packet arrival and telemetry ping/pong
* use generated packet constants for diagnostic packets
* attach debug-only UI under the root viewport
* feed network timing snapshots into other devtools presentation, such as network player labels

They are not allowed to:

* replace HUD presentation
* drive gameplay decisions
* infer authoritative gameplay outcomes
* mutate server state
* persist telemetry history
* report analytics
* bypass the normal packet codec or connection service
* become a debug-only gameplay simulation path

The overlay is intentionally glanceable. Raw packet inspection belongs in the devtools window readout surfaces, not in the world telemetry overlay.

## Server authority

The server remains authoritative for gameplay facts displayed by the overlay.

World counts derive from the normalized gameplay state packet. The client does not decide how many players, asteroids, bullets, pickups, or enemies exist. It only counts dictionaries that came from the server-owned state projection.

The server also participates in packet timing diagnostics through `telemetry_ping` and `telemetry_pong`.

`telemetry_ping` is a client-originated diagnostic packet. The server receives it through the normal WebSocket packet route, builds a `telemetry_pong`, and sends the response back through the same connected session. The response preserves the ping sequence and client send timestamp, and adds server receive/send timestamps.

This packet pair does not require room membership and does not mutate gameplay state.

The server also stamps gameplay state packets with `server_sent_msec` before outbound encoding. The client uses that value together with the estimated server clock offset from telemetry pong packets to calculate `packet_age_ms` in local monotonic-clock space.

The authority boundary is:

```text
server owns gameplay facts and server timestamps
client owns overlay presentation and local timing calculations
ping/pong owns diagnostics only
```

## Client presentation

The world telemetry overlay scene is:

```text
client/scenes/devtools/world_telemetry_overlay.tscn
```

The scene is a `CanvasLayer` with a bottom-right `PanelContainer` and a single label. The panel ignores mouse input so it does not become a gameplay or menu input target.

The overlay renders three groups.

World metrics:

```text
players
enemies
asteroids
pickups
total_asteroids
bullets
```

Client metrics:

```text
fps
frame_ms
```

Network metrics:

```text
rtt_ms
packet_interval_ms
jitter_ms
packet_staleness_ms
packet_age_ms
```

Unavailable values render as:

```text
—
```

Counts come from server dictionaries after gameplay packet normalization. Timing values use `-1` internally for unavailable state and render as unavailable in the label.

The overlay refresh cadence is separate from gameplay packet arrival. `WorldTelemetryOverlayFlow` refreshes the visible label at most once every 250 ms while the overlay is visible.

## Commands or controls

`DevToggle9` toggles the world telemetry overlay.

The current route is:

```text
GameplayDevtoolsContext.process
-> DevtoolsHotkeyContext.process
-> DevtoolsOverlayContext.toggle_world_telemetry_overlay
-> WorldTelemetryContext.toggle_overlay
-> WorldTelemetryOverlayFlow.toggle_overlay
```

When the overlay is shown, `WorldTelemetryOverlayFlow` instantiates the overlay scene if needed, adds it to the root viewport, marks it visible, and refreshes metrics immediately.

When the overlay is hidden, the node remains available but is not visible. `reset()` clears timing state and frees the overlay node when the devtools context is reset.

While the overlay is visible and the connection service reports an active server connection, `WorldTelemetryContext` sends one `telemetry_ping` packet every 1000 ms. No ping is sent while the overlay is hidden.

## Telemetry

World telemetry uses two metric sources.

The first source is normalized gameplay state:

```text
server_players
server_asteroids
server_pickups
server_bullets
server_enemies
enemies
total_asteroids
server_sent_msec
```

`GameplayStatePacketReader` produces these fields from the authoritative state packet. `WorldTelemetryMetrics` consumes them and stores the current count and packet timing state.

`server_enemies` is preferred when present. If it is not present, the metrics collector falls back to `enemies`.

The second source is network ping/pong data:

```text
rtt_ms
server_clock_offset_ms
```

`NetworkTelemetryMetrics` creates ping packets with:

```text
type = telemetry_ping
sequence
client_sent_msec
```

When a matching pong arrives, it calculates RTT from the local send and receive timestamps. If the pong also contains server receive/send timestamps, it estimates the server clock offset by comparing the local midpoint to the server midpoint.

Packet timing fields mean:

```text
packet_interval_ms
  local elapsed time between the two most recent gameplay state packets

jitter_ms
  absolute difference between the two most recent packet intervals

packet_staleness_ms
  local elapsed time since the latest gameplay state packet was received

packet_age_ms
  estimated age of the latest gameplay packet using server_sent_msec and server_clock_offset_ms
```

`packet_staleness_ms` can be available without a server clock offset. `packet_age_ms` requires both `server_sent_msec` and an estimated server clock offset.

## Build/runtime gates

Client devtools hotkeys are gated by:

```text
client/scripts/devtools/dev_tools_build_flags.gd
```

When `public_build` is `true`, devtools input actions listed in `DEVTOGGLE_ACTIONS` are erased from the `InputMap` during `_ready()`.

Telemetry overlay runtime behavior is also gated by visibility and connection state:

* the overlay only refreshes while its node is visible
* telemetry ping packets are only sent while the overlay is visible
* ping packets are not sent when the connection service is absent
* ping packets are not sent when the connection service reports no server connection

The server-side telemetry ping route is not a gameplay mutation path. It is handled by networking telemetry routing and returns a diagnostic response through the same session.

## Code map

Primary client telemetry overlay files:

```text
client/scenes/devtools/world_telemetry_overlay.tscn
client/scripts/devtools/telemetry/world_telemetry_overlay.gd
client/scripts/devtools/telemetry/world_telemetry_overlay_flow.gd
client/scripts/devtools/telemetry/world_telemetry_context.gd
client/scripts/devtools/telemetry/world_telemetry_metrics.gd
client/scripts/devtools/telemetry/network_telemetry_metrics.gd
```

Client devtools composition and routing:

```text
client/scripts/devtools/gameplay_devtools_context.gd
client/scripts/devtools/context/devtools_overlay_context.gd
client/scripts/devtools/context/devtools_hotkey_context.gd
client/scripts/devtools/context/devtools_gameplay_state_context.gd
client/scripts/devtools/dev_tools_build_flags.gd
```

Gameplay state source path:

```text
client/scripts/gameplay/state/gameplay_state_flow.gd
client/scripts/gameplay/state/gameplay_state_packet_reader.gd
client/scripts/gameplay/state/gameplay_state_apply_flow.gd
```

Client networking path for telemetry pong:

```text
client/scripts/networking/client_connection_service.gd
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/generated/networking/packets/packets.gd
```

Server telemetry packet path:

```text
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/inbound/telemetry.go
services/game-server/internal/networking/outbound/gameplay_presentation.go
services/game-server/internal/protocol/packetcodec/
```

Packet source and generated contract boundary:

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
client/scripts/generated/networking/packets/packets.gd
services/game-server/internal/game/packets.go
```

Important non-ownership boundaries:

```text
client/scripts/gameplay/hud/
  owns player-facing HUD presentation, not telemetry overlays

client/scripts/ui/
  owns normal UI surfaces, not telemetry diagnostics

services/game-server/internal/game/
  owns authoritative simulation and state projection

services/game-server/internal/networking/
  owns realtime packet routing and timestamped outbound packet responses

docs/services/game-server/networking/
  owns game-server packet routing documentation

docs/data/
  owns shared packet source-of-truth and generated output documentation
```

## Tests

Focused telemetry tests:

```text
client/tests/unit/devtools/telemetry/test_world_telemetry_metrics.gd
client/tests/unit/devtools/telemetry/test_network_telemetry_metrics.gd
client/tests/unit/devtools/telemetry/test_world_telemetry_context.gd
```

Related client tests:

```text
client/tests/unit/test_gameplay_state_packet_reader.gd
client/tests/unit/test_gameplay_devtools_context.gd
client/tests/unit/test_gameplay_state_apply_flow.gd
client/tests/unit/test_packet_codec.gd
```

The telemetry metrics tests cover world counts, enemy fallback behavior, invalid source handling, packet interval timing, jitter availability, reset behavior, `server_sent_msec`, missing timestamp behavior, ping packet shape, RTT updates, unknown pong handling, reset clearing, and server clock offset calculation.

The telemetry context test verifies that showing the overlay allows processing to send a `telemetry_ping`, and that a matching `telemetry_pong` updates RTT metrics.

## Related docs

* [Client Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Client Networking Flow](../../services/client/networking-flow/!INDEX.md)
* [Inbound Packet Routing](../../services/client/networking-flow/inbound-packet-routing.md)
* [Gameplay State Application](../../services/client/gameplay-runtime/gameplay-state-application.md)
* [Game Server Telemetry And Packet Routing](../../services/game-server/networking/telemetry-packet-routing.md)
* [Game Server Networking](../../services/game-server/networking/!INDEX.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Data Sync And SSoT Pipeline](../../data/data-sync-and-ssot-pipeline.md)
* [Protocol](../../protocol/!INDEX.md)

## Notes

Telemetry overlay data is intentionally shallow. It is for immediate development visibility, not historical analysis.

`packet_staleness_ms` and `packet_age_ms` are different measurements. Staleness is local time since the last state packet arrived. Age estimates how old the packet was by using the server send timestamp and the estimated server clock offset.

The overlay and network player-label mode share the same network telemetry snapshot source, but they are separate presentation surfaces. The overlay owns the fixed world/client/network panel. Player dev labels own per-player label presentation.
