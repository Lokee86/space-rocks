# Client Logging

Parent index: [Client Legacy](./!INDEX.md)

The Godot client has a lightweight logging helper at:

```text
client/scripts/logging/logger.gd
```

Use this helper for new client lifecycle, UI, networking, packet, HUD, input, and world-sync diagnostics instead of adding raw `print()` calls. Existing temporary `print()` debugging can be cleaned up opportunistically when touching the same path, but avoid broad logging-only churn.

## Categories

Current categories:

- `shell`
- `lobby`
- `network`
- `game`
- `world_sync`
- `hud`
- `input`
- `packets`
- `default`

Convenience methods exist for each category and level, for example:

```gdscript
const ClientLogger = preload("res://scripts/logging/logger.gd")

ClientLogger.shell_debug("room entered InGame; showing multiplayer game loop")
ClientLogger.lobby_debug("LeaveButton pressed")
ClientLogger.network_debug("ReturnToLobbyRequest sent")
ClientLogger.game_warn("unexpected game menu state")
```

## Levels

Supported levels:

- `debug`
- `info`
- `warn`
- `error`
- `off`

`warn` logs through `push_warning()`, `error` logs through `push_error()`, and lower levels use `print()` behind the helper.

The current default level is `debug` while multiplayer lifecycle work is active. If logs become noisy, prefer lowering the level through the helper rather than deleting useful lifecycle events:

```gdscript
ClientLogger.set_default_level(ClientLogger.LEVEL_WARN)
ClientLogger.set_category_level(ClientLogger.CATEGORY_NETWORK, ClientLogger.LEVEL_DEBUG)
ClientLogger.disable()
ClientLogger.enable_debug()
```

## Guidance

Good client log events:

- create/join/leave button actions
- websocket open/close events
- generated lifecycle request sends
- room snapshot/state transitions that move UI between lobby/gameplay/game-over
- unexpected packet or UI states that are recoverable

Avoid:

- per-frame logs
- per-entity world-sync logs during normal gameplay
- logging every input packet
- raw `print()` in new lifecycle or networking code

Logging should remain observational. It should not change UI behavior, packet sends, or gameplay state.