# Player Dev Labels

Parent index: [Client](./!INDEX.md)

## Purpose

This document describes the client devtools player-label overlay used to inspect remote players during development.

It covers client-side presentation, label mode routing, telemetry readback, and the implementation seams that keep player labels out of normal gameplay rendering ownership.

## Overview

Player dev labels are client-only devtools overlays attached to remote player nodes.

They are used to inspect synchronized remote player state and current client telemetry while the game is running. They are not HUD content, not production gameplay UI, and not server-authoritative gameplay state.

The current label system supports three modes:

```text
off
basic
network
```

`basic` mode shows per-remote-player gameplay readback derived from the latest gameplay state.

`network` mode shows network and packet timing telemetry. The network label text is computed once from the latest telemetry snapshot and displayed on each remote player label.

The modes are mutually exclusive. Enabling one active mode replaces the other; toggling the active mode again disables labels and frees existing label nodes.

## Debug-only scope

Player dev labels are development/debug tooling.

They are allowed to observe:

```text
remote player nodes
latest gameplay state
player session read models
network telemetry metrics
world packet timing metrics
```

They must not own:

```text
player lifecycle
remote player creation
remote player removal
player interpolation
player presentation state
gameplay authority
server mutation
production HUD behavior
```

The devtools label system may attach a child node to a remote player node for display, but it does not mutate the player’s gameplay state, movement state, target state, score, lives, or lifecycle.

## Server authority

Player dev labels do not send packets and do not request server mutations.

The server remains authoritative for the gameplay state that labels read:

```text
players
player_sessions
x
y
ship_type
score
lives
```

The client label system treats incoming gameplay state as read-only diagnostic input.

Network label data is also diagnostic readback. It derives from client-side telemetry calculations and packet timing fields, not from a gameplay command surface.

## Client presentation

The label scene is:

```text
client/scenes/devtools/player_dev_label.tscn
```

The scene root is a `Node2D` with two panel-backed labels:

```text
PlayerInfoLabel
PlayerNetworkLabel
```

Only one panel is visible at a time.

`PlayerDevLabel` configures itself as a child of a remote player node:

```text
top_level = false
visible = false
position = Vector2(60, -70)
rotation = 0.0
```

While visible, the label forces `global_rotation` back to `0.0` during processing so label text stays screen-readable instead of inheriting the player’s rotation.

Basic mode shows:

```text
ID
Score
Lives
Ship
X
Y
```

Formatting behavior:

```text
player IDs longer than 8 characters are shortened with an ellipsis
missing values render as —
x and y are rounded to integer text
```

Network mode shows:

```text
rtt_ms
packet_interval_ms
jitter_ms
packet_staleness_ms
packet_age_ms
```

Missing or negative numeric metrics render as `—`.

## Label lifecycle

`PlayerDevLabelsContext` owns label lifecycle.

It receives a remote-player-node provider from gameplay composition. The provider currently comes from:

```text
WorldSync.remote_player_nodes()
```

That route exposes a dictionary of remote player nodes without making world sync responsible for label formatting or label lifecycle.

During each devtools process pass, the label context:

1. Clears labels when mode is off.
2. Clears labels when no valid remote-player-node provider exists.
3. Clears labels when the provider does not return a dictionary.
4. Removes label nodes for remote players that no longer exist.
5. Instantiates missing labels for valid remote player nodes.
6. Updates existing labels for the active mode.

The context stores labels by player ID:

```text
labels_by_player_id
```

When labels are cleared or become stale, label nodes are `queue_free()`d and removed from the map.

## Label modes

The mode state is stored in `DevtoolsStateContext` as:

```text
player_dev_label_mode
```

Valid label-context modes are:

```text
""
"basic"
"network"
```

Any other mode is normalized to `""`.

Mode routing is coordinated through:

```text
DevtoolsHotkeyContext
DevtoolsOverlayContext
PlayerDevLabelsContext
```

The state context stores the selected mode. The overlay context applies that mode to the label context. The label context owns what happens to actual label nodes when the mode changes.

## Commands or controls

Player dev labels are controlled by devtools input actions.

Current controls:

```text
DevToggle8
= toggle basic remote-player labels

Shift + DevToggle8
= toggle network telemetry labels
```

Control behavior:

```text
DevToggle8 when basic is active -> off
DevToggle8 when basic is not active -> basic
Shift + DevToggle8 when network is active -> off
Shift + DevToggle8 when network is not active -> network
```

Basic and network labels are mutually exclusive because both controls write the same mode state.

The labels do not have devtools-window controls in the current implementation. They are hotkey-driven overlays.

## Telemetry

Basic label data comes from the latest gameplay state applied into the devtools overlay context.

The basic path reads:

```text
latest_gameplay_state.server_players
latest_gameplay_state.player_sessions
```

For each labeled remote player:

```text
server_players[player_id]
= position and ship state

player_sessions[player_id]
= score and lives state
```

Network label data comes from the world telemetry context snapshot.

The overlay context copies telemetry into the label context during processing:

```text
WorldTelemetryContext.telemetry_snapshot()
-> PlayerDevLabelsContext.apply_network_metrics()
```

The telemetry snapshot merges network and world-packet metrics from the current client runtime. Network labels do not require per-player telemetry. The same current telemetry snapshot is shown on every remote player label.

## Build/runtime gates

The player dev label hotkey is part of the devtools action list.

In public builds, `DevToolsBuildFlags` removes devtools input action events for:

```text
DevToggle0
DevToggle1
DevToggle2
DevToggle3
DevToggle4
DevToggle5
DevToggle6
DevToggle7
DevToggle8
DevToggle9
```

That disables the configured input route for player dev labels in public builds.

The label implementation itself is still client code. The runtime gate is the devtools input-action removal, not a separate label-specific build tag.

## Relationship to gameplay implementation

Player dev labels intentionally use a narrow observation seam from world rendering.

Current observation path:

```text
GameplayFlowComposer
-> GameplayInputContext
-> GameplayDevtoolsContext
-> DevtoolsOverlayContext
-> PlayerDevLabelsContext

WorldSync.remote_player_nodes()
-> PlayerRenderApi.remote_player_nodes()
-> PlayerMeaningApi.remote_player_nodes()
-> legacy player render sync remote-player node map
```

The important ownership rule is:

```text
World/player rendering exposes remote-player nodes.
Client devtools owns label lifecycle and formatting.
```

`PlayerSync`, `WorldSync`, and player-rendering code should not accumulate label formatting, label mode state, or devtools label lifecycle.

## Code map

Primary implementation paths:

```text
client/scripts/devtools/player_dev_label.gd
client/scripts/devtools/player_dev_label_formatter.gd
client/scripts/devtools/player_labels/player_dev_labels_context.gd
client/scenes/devtools/player_dev_label.tscn
```

Mode and overlay coordination:

```text
client/scripts/devtools/context/devtools_state_context.gd
client/scripts/devtools/context/devtools_hotkey_context.gd
client/scripts/devtools/context/devtools_overlay_context.gd
client/scripts/devtools/gameplay_devtools_context.gd
client/scripts/devtools/dev_tools_build_flags.gd
```

Gameplay composition and observation paths:

```text
client/scripts/gameplay/runtime/gameplay_flow_composer.gd
client/scripts/gameplay/input/gameplay_input_context.gd
client/scripts/gameplay/state/gameplay_state_apply_flow.gd
client/scripts/gameplay/runtime/gameplay_process_flow.gd
client/scripts/world/world_sync.gd
client/scripts/world/player_render/player_render_api.gd
client/scripts/world/player_render/player_meaning_api.gd
client/legacy/player_render/player_sync.gd
```

Telemetry inputs:

```text
client/scripts/devtools/telemetry/world_telemetry_context.gd
client/scripts/devtools/telemetry/world_telemetry_overlay_flow.gd
client/scripts/devtools/telemetry/world_telemetry_metrics.gd
client/scripts/devtools/telemetry/network_telemetry_metrics.gd
```

Generated packet constants used for field names:

```text
client/scripts/generated/networking/packets/packets.gd
shared/packets/
```

## Tests

Related test coverage exists around the surrounding seams:

```text
client/tests/unit/devtools/context/test_devtools_state_context.gd
client/tests/unit/devtools/telemetry/test_world_telemetry_context.gd
client/tests/unit/devtools/telemetry/test_world_telemetry_metrics.gd
client/tests/unit/devtools/telemetry/test_network_telemetry_metrics.gd
client/tests/unit/world/player_render/test_player_render_api.gd
client/tests/unit/test_world_sync.gd
client/tests/unit/test_gameplay_input_context.gd
client/tests/unit/test_gameplay_state_apply_flow.gd
```

No dedicated player-dev-label formatter or player-dev-label context test file is currently present in the client test tree.

## Related docs

* [Client Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Client Service](../../services/client/!INDEX.md)
* [Client input and targeting](../../services/client/input-and-targeting.md)
* [Client world sync](../../services/client/world-sync/!INDEX.md)
* [Realtime protocol](../../protocol/!INDEX.md)
* [Data](../../data/!INDEX.md)

## Notes

Player dev labels should remain diagnostic presentation. Any future label mode should either read existing authoritative state or read existing telemetry seams. It should not add a parallel debug-only gameplay model.
