# Developer Toggles

This document covers the developer/debug toggles that currently exist in the project.

These tools are for local development and testing. They are not a player-facing feature and they should stay server-authoritative when they affect gameplay.

## Current Toggles

### Invincibility

Invincibility prevents the player from dying when colliding with asteroids.

Current behavior:

- `1` is a self-targeting hotkey from the Godot client; it toggles the local/requesting player.
- The devtools window can target any listed player.
- The client sends a `toggle_debug_invincible` packet.
- Targeted UI sends `toggle_debug_invincible` with `target_player_id`.
- If `target_player_id` is omitted, the server falls back to the requesting player.
- The Go server toggles the flag for the targeted player instance.
- Ship/asteroid collision skips players with debug invincibility enabled.
- Movement, shooting, and scoring still work normally.
- Pressing `1` again disables invincibility.

There is no in-game developer console yet. This is currently a hardcoded client hotkey.

### Infinite Lives

Infinite lives lets the player die normally without reducing their lives count.

Current behavior:

- `2` is a self-targeting hotkey from the Godot client; it toggles the local/requesting player.
- The devtools window can target any listed player.
- The client sends a `toggle_debug_infinite_lives` packet.
- Targeted UI sends `toggle_debug_infinite_lives` with `target_player_id`.
- If `target_player_id` is omitted, the server falls back to the requesting player.
- The Go server toggles the flag for the targeted player session.
- Ship/asteroid collision still kills and despawns the player.
- The death event still fires and the respawn delay still applies.
- The player's lives count does not decrease while the toggle is enabled.
- The toggle persists across respawns for the same connection/player session.
- Pressing `2` again disables infinite lives.

### World Freeze

World freeze pauses hostile/world simulation while leaving the player able to move.

Current behavior:

- Triggered from the Godot client with `3`.
- The client sends a `toggle_debug_freeze_world` packet.
- The Go server toggles world-freeze state on the current game room.
- The toggle is room-wide. Every player in that room is affected.
- Asteroid spawning stops.
- Existing asteroids stop moving.
- New bullets do not spawn.
- Existing bullets stop moving and their lifetime stops ticking down.
- Ship/asteroid collisions stop running.
- Bullet/asteroid collisions stop running, so bullet impacts, score awards, and asteroid splits are paused.
- Player movement and input still work.
- Player respawn/session timers still work.
- Existing ready-for-removal cleanup can still run.
- Pressing `3` again resumes world simulation.

### Player Freeze

Player freeze suspends one player for debugging through the same ship capability path used by pause.

Current behavior:

- `4` is a self-targeting hotkey from the Godot client; it toggles the local/requesting player.
- The devtools window can target any listed player.
- The client sends a `toggle_debug_freeze_player` packet.
- Targeted UI sends `toggle_debug_freeze_player` with `target_player_id`.
- If `target_player_id` is omitted, the server falls back to the requesting player.
- The Go server toggles the freeze flag for the targeted player instance.
- The toggle blocks input, movement, shooting, and asteroid collision damage through `Ship.IsSuspended()` and related capability helpers.
- Pause and dev freeze are separate suspension causes. Dev freeze does not call `Pause()` or `Resume()`.
- Calling `Resume()` does not unfreeze a dev-frozen player.
- Unfreezing does not resume a paused player.
- Pressing `4` again disables player freeze.

## DevToggle0-9 Map

Current number-key map:

- `0`: window
- `1`: invincible (self-targeting hotkey)
- `2`: infinite lives (self-targeting hotkey)
- `3`: world freeze
- `4`: player freeze (self-targeting hotkey)
- `5`: kill local player
- `6`: spawn new player
- `7`: force respawn local player
- `8`: reserved
- `9`: reserved

Current `6` modifier behavior:

- `Shift+6`: spawn asteroid
- `Alt+6`: spawn bullet

## Devtools Window Targeting

Devtools window actions use player-select controls populated from current gameplay state for:

- Kill Player
- Respawn Player
- Spawn Player
- Invincibility
- Infinite Lives
- Freeze Player

Invincibility, Infinite Lives, and Freeze Player selectors show only feature state wording (`Active`/`Inactive`) for the selected player:

- `InvincibleStatusSelect`
- `InfiniteLivesSelect`
- `PlayerFrozenSelect`

Kill/Respawn selectors may still use lifecycle wording such as `ALIVE`/`DEAD`.

World Freeze remains a global room toggle and does not use a player selector.

## Implementation

Current ownership paths:

- packet schema (devtools): `shared/packets/debug.toml`
- packet output routing: `shared/packets/outputs.toml`
- generated server devtools packets: `services/game-server/internal/devtools/packets_generated.go`
- generated client packets: `client/scripts/networking/packets/packets.gd`
- server devtools behavior: `services/game-server/internal/devtools/`
- controlled game access seam: `services/game-server/internal/game/export_devtools*.go`
- websocket routing: `services/game-server/internal/networking/`
- client devtools window/context: `client/scripts/devtools/`
- client gameplay input routing: `client/scripts/gameplay/input/`
- gameplay shell state routing: `client/scripts/shell/gameplay_shell_flow.gd`

## Server Build Flag

Server devtools are enabled in normal/default builds.

Building or running with the Go build tag `nodevtools` disables server devtools command handling.

```bash
go run -tags nodevtools ./cmd/game-server
go build -tags nodevtools -buildvcs=false -o ./tmp/game-server ./cmd/game-server
```

## Packet Flow

When `1` is pressed:

1. `DevToggle1` routes through client devtools/gameplay input seams.
2. The client sends `Packets.toggle_debug_invincible_packet()`.
3. The generated packet builder emits:

```gdscript
{
	"type": "toggle_debug_invincible"
}
```

4. `internal/networking` classifies packet type first and routes enabled devtools packets to `devtools.HandleCommand(...)`.
5. The server toggles player `DamageOptions.Invincible`.
6. Targeted devtools UI can send `Packets.toggle_debug_invincible_target_player_packet(target_player_id)`, which emits:

```gdscript
{
	"type": "toggle_debug_invincible",
	"target_player_id": "<player-id>"
}
```

7. Outgoing devtools status reports the receiving/local player's state through `debug_status.invincible` and the per-player map through `debug_statuses` for devtools window target/status rows.

When `2` is pressed:

1. `DevToggle2` routes through client devtools/gameplay input seams.
2. The client sends `Packets.toggle_debug_infinite_lives_packet()`.
3. The generated packet builder emits:

```gdscript
{
	"type": "toggle_debug_infinite_lives"
}
```

4. `internal/networking` classifies packet type first and routes enabled devtools packets to `devtools.HandleCommand(...)`.
5. The server toggles session `LifeOptions.InfiniteLives`.
6. Targeted devtools UI can send `Packets.toggle_debug_infinite_lives_target_player_packet(target_player_id)`, which emits:

```gdscript
{
	"type": "toggle_debug_infinite_lives",
	"target_player_id": "<player-id>"
}
```

7. Outgoing devtools status reports the receiving/local player's state through `debug_status.infinite_lives` and the per-player map through `debug_statuses` for devtools window target/status rows.

When `3` is pressed:

1. `DevToggle3` routes through client devtools/gameplay input seams.
2. The client sends `Packets.toggle_debug_freeze_world_packet()`.
3. The generated packet builder emits:

```gdscript
{
	"type": "toggle_debug_freeze_world"
}
```

4. World Freeze is global and has no `target_player_id` field.
5. `internal/networking` classifies packet type first and routes enabled devtools packets to `devtools.HandleCommand(...)`.
6. The server toggles `WorldSimulationOptions`.
7. Simulation gates read `worldSimulationOptions` before asteroid spawning, asteroid advancing, bullet advancing, and collision passes.
8. Outgoing devtools status reports the receiving/local player's view through `debug_status.world_frozen`, while per-player rows still come from `debug_statuses`.

When `4` is pressed:

1. `DevToggle4` routes through client devtools/gameplay input seams.
2. The client sends `Packets.toggle_debug_freeze_player_packet()`.
3. The generated packet builder emits:

```gdscript
{
	"type": "toggle_debug_freeze_player"
}
```

4. `internal/networking` classifies packet type first and routes enabled devtools packets to `devtools.HandleCommand(...)`.
5. The server toggles player/session `Suspension.DevFrozen`.
6. Ship capability helpers use `Ship.IsSuspended()` before accepting input, moving, shooting, or taking asteroid collision damage.
7. Targeted devtools UI can send `Packets.toggle_debug_freeze_player_target_player_packet(target_player_id)`, which emits:

```gdscript
{
	"type": "toggle_debug_freeze_player",
	"target_player_id": "<player-id>"
}
```

8. Outgoing devtools status reports the receiving/local player's state through `debug_status.player_frozen` and the per-player map through `debug_statuses` for devtools window target/status rows.

## Logging

Toggling invincibility logs through the custom server logger:

```go
logging.Game.Info("debug invincibility toggled",
	logging.FieldPlayerID, playerID,
	"target_player_id", targetPlayerID,
	"enabled", enabled,
)
```

Toggling infinite lives logs similarly:

```go
logging.Game.Info("debug infinite lives toggled",
	logging.FieldPlayerID, playerID,
	"target_player_id", targetPlayerID,
	"enabled", enabled,
)
```

Toggling world freeze logs similarly:

```go
logging.Game.Info("debug world freeze toggled",
	logging.FieldPlayerID, playerID,
	"enabled", enabled,
)
```

Toggling player freeze logs similarly:

```go
logging.Game.Info("debug player freeze toggled",
	logging.FieldPlayerID, playerID,
	"target_player_id", targetPlayerID,
	"enabled", enabled,
)
```

See [server logging](../server/logging.md) for logging configuration.

## Testing

Server tests live in:

```text
services/game-server/tests/game/devtools_test.go
```

Run:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

Current coverage checks:

- an invincible player does not die from asteroid collision
- toggling invincibility once enables it
- toggling invincibility twice disables it
- an infinite-lives player dies without losing a life
- infinite lives persists after respawn
- toggling infinite lives once enables it
- toggling infinite lives twice disables it
- toggling world freeze once enables it
- toggling world freeze twice disables it
- player freeze contributes to `Ship.IsSuspended()`
- player freeze blocks ship input and movement capabilities
- paused and frozen players remain suspended until both causes are cleared
- kill player can target another player
- source player remains unchanged when kill-player targets another player

TODO: add focused server tests for world freeze:

- frozen world does not spawn asteroids
- frozen world does not move asteroids
- frozen world does not move or expire bullets
- frozen world does not spawn bullets
- frozen world does not run ship/asteroid collisions
- frozen world does not run bullet/asteroid collisions, scoring, or asteroid splits

TODO: add focused server tests for targeted player toggles:

- invincibility can target another player
- infinite lives can target another player
- freeze player can target another player
- source player remains unchanged when another player is targeted

TODO: if/when a `Game.Start` duplicate-simulation-goroutine guard test exists, list it here.

## Design Notes

Keep debug gameplay effects server-side. The client may request a toggle, but the server should own whether the toggle is active and how it affects simulation.

Keep devtools isolated. Debug packet handling and outgoing debug status wrapping should stay in `internal/devtools`, while gameplay-affecting state should live in the owning gameplay seams via `export_devtools*.go`: `DamageOptions`, `LifeOptions`, `Suspension`, and `WorldSimulationOptions`.

Avoid scattering one-off debug booleans through core logic. Prefer small gameplay-owned capability methods so collision/combat code only asks simple gameplay questions.

## Future Options

Likely future devtools:

- collision polygon display
- spawn asteroid near player
- clear asteroids
- force game over
- debug HUD or developer menu
- developer console command layer

If a real dev console is added later, it should call the same packet path instead of bypassing server authority.
