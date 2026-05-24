# Developer Toggles

This document covers the developer/debug toggles that currently exist in the project.

These tools are for local development and testing. They are not a player-facing feature and they should stay server-authoritative when they affect gameplay.

## Current Toggles

### Invincibility

Invincibility prevents the player from dying when colliding with asteroids.

Current behavior:

- Triggered from the Godot client with `F1`.
- The client sends a `toggle_debug_invincible` packet.
- The Go server toggles the flag for that player instance.
- Ship/asteroid collision skips players with debug invincibility enabled.
- Movement, shooting, and scoring still work normally.
- Pressing `F1` again disables invincibility.

There is no in-game developer console yet. This is currently a hardcoded client hotkey.

### Infinite Lives

Infinite lives lets the player die normally without reducing their lives count.

Current behavior:

- Triggered from the Godot client with `F2`.
- The client sends a `toggle_debug_infinite_lives` packet.
- The Go server toggles the flag for that player session.
- Ship/asteroid collision still kills and despawns the player.
- The death event still fires and the respawn delay still applies.
- The player's lives count does not decrease while the toggle is enabled.
- The toggle persists across respawns for the same connection/player session.
- Pressing `F2` again disables infinite lives.

### World Freeze

World freeze pauses hostile/world simulation while leaving the player able to move.

Current behavior:

- Triggered from the Godot client with `F3`.
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
- Pressing `F3` again resumes world simulation.

### Player Freeze

Player freeze suspends one player for debugging through the same ship capability path used by pause.

Current behavior:

- Triggered from the Godot client with `F4`.
- The client sends a `toggle_debug_freeze_player` packet.
- The Go server toggles the freeze flag for that player instance.
- The toggle blocks input, movement, shooting, and asteroid collision damage through `Ship.IsSuspended()` and related capability helpers.
- Pause and dev freeze are separate suspension causes. Dev freeze does not call `Pause()` or `Resume()`.
- Calling `Resume()` does not unfreeze a dev-frozen player.
- Unfreezing does not resume a paused player.
- Pressing `F4` again disables player freeze.

## Implementation

The devtools state lives in:

```text
services/game-server/internal/game/devtools/player_options.go
```

Player entities store their debug options here:

```text
services/game-server/internal/game/entities/state.go
```

The packet source of truth is:

```text
shared/packets/packets.toml
```

Generated packet files include:

```text
services/game-server/internal/game/packets.go
services/game-server/internal/game/entities/packets_generated.go
client/scripts/networking/packets.gd
```

The client hotkey is currently handled in:

```text
client/scripts/game.gd
```

The server toggle handling is in:

```text
services/game-server/internal/game/game.go
```

The world-freeze collision pass gate is in:

```text
services/game-server/internal/game/game.go
```

Pair collision fact helpers are in `services/game-server/internal/game/collisions.go`, and combat consumes those facts in `services/game-server/internal/game/combat.go`.

World-freeze gates are in:

```text
services/game-server/internal/game/game.go
```

## Packet Flow

When `F1` is pressed:

1. `client/scripts/game.gd` checks for `KEY_F1`.
2. If connected, the client sends `Packets.toggle_debug_invincible_packet()`.
3. The generated packet builder emits:

```gdscript
{
	"type": "toggle_debug_invincible"
}
```

4. The server receives `PacketTypeToggleDebugInvincible`.
5. The server toggles `player.DevTools.Invincible`.
6. Collision handling checks `player.DevTools.CanTakeDamage()`.

When `F2` is pressed:

1. `client/scripts/game.gd` checks for `KEY_F2`.
2. If connected, the client sends `Packets.toggle_debug_infinite_lives_packet()`.
3. The generated packet builder emits:

```gdscript
{
	"type": "toggle_debug_infinite_lives"
}
```

4. The server receives `PacketTypeToggleDebugInfiniteLives`.
5. The server toggles `player.DevTools.InfiniteLives`.
6. The player session stores the updated devtools options so the toggle survives respawn.
7. Death handling checks `player.DevTools.CanLoseLives()` before decrementing lives.

When `F3` is pressed:

1. `client/scripts/game.gd` checks for `KEY_F3`.
2. If connected, the client sends `Packets.toggle_debug_freeze_world_packet()`.
3. The generated packet builder emits:

```gdscript
{
	"type": "toggle_debug_freeze_world"
}
```

4. The server receives `PacketTypeToggleDebugFreezeWorld`.
5. The server toggles `game.worldDevTools`.
6. `Game.Step()` checks `worldDevTools` before asteroid spawning, asteroid advancing, bullet advancing, and collision passes.

When `F4` is pressed:

1. `client/scripts/game.gd` checks for `KEY_F4`.
2. If connected, the client sends `Packets.toggle_debug_freeze_player_packet()`.
3. The generated packet builder emits:

```gdscript
{
	"type": "toggle_debug_freeze_player"
}
```

4. The server receives `PacketTypeToggleDebugFreezePlayer`.
5. The server toggles `player.DevTools.FreezePlayer`.
6. Ship capability helpers check `Ship.IsSuspended()` before accepting input, moving, shooting, or taking asteroid collision damage.

## Logging

Toggling invincibility logs through the custom server logger:

```go
logging.Game.Info("debug invincibility toggled",
	logging.FieldPlayerID, playerID,
	"enabled", enabled,
)
```

Toggling infinite lives logs similarly:

```go
logging.Game.Info("debug infinite lives toggled",
	logging.FieldPlayerID, playerID,
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

TODO: add focused server tests for world freeze:

- frozen world does not spawn asteroids
- frozen world does not move asteroids
- frozen world does not move or expire bullets
- frozen world does not spawn bullets
- frozen world does not run ship/asteroid collisions
- frozen world does not run bullet/asteroid collisions, scoring, or asteroid splits

## Design Notes

Keep debug gameplay effects server-side. The client may request a toggle, but the server should own whether the toggle is active and how it affects simulation.

Keep devtools isolated. New debug-only gameplay state should live behind `services/game-server/internal/game/devtools` where practical, so the game can ignore or remove it cleanly later.

Avoid scattering one-off debug booleans through core logic. Prefer small methods like `CanTakeDamage()` so collision/combat code only asks a simple gameplay question.

## Future Options

Likely future devtools:

- collision polygon display
- spawn asteroid near player
- clear asteroids
- force game over
- debug HUD or developer menu
- developer console command layer

If a real dev console is added later, it should call the same packet path instead of bypassing server authority.
