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

## Implementation

The devtools state lives in:

```text
server/internal/game/devtools/player_options.go
```

Player entities store their debug options here:

```text
server/internal/game/entities/state.go
```

The packet source of truth is:

```text
shared/packets/packets.json
```

Generated packet files include:

```text
server/internal/game/packets.go
client/scripts/packets.gd
```

The client hotkey is currently handled in:

```text
client/scripts/game.gd
```

The server toggle handling is in:

```text
server/internal/game/game.go
```

The collision gate is in:

```text
server/internal/game/combat.go
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

## Logging

Toggling invincibility logs through the custom server logger:

```go
logging.Game.Info("debug invincibility toggled",
	logging.FieldPlayerID, playerID,
	"enabled", enabled,
)
```

See [server logging](../server/logging.md) for logging configuration.

## Testing

Server tests live in:

```text
server/internal/game/game_devtools_test.go
```

Run:

```bash
cd server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

Current coverage checks:

- an invincible player does not die from asteroid collision
- toggling invincibility once enables it
- toggling invincibility twice disables it

## Design Notes

Keep debug gameplay effects server-side. The client may request a toggle, but the server should own whether the toggle is active and how it affects simulation.

Keep devtools isolated. New debug-only gameplay state should live behind `server/internal/game/devtools` where practical, so the game can ignore or remove it cleanly later.

Avoid scattering one-off debug booleans through core logic. Prefer small methods like `CanTakeDamage()` so collision/combat code only asks a simple gameplay question.

## Future Options

Likely future devtools:

- infinite lives
- collision polygon display
- spawn asteroid near player
- clear asteroids
- force game over
- debug HUD or developer menu
- developer console command layer

If a real dev console is added later, it should call the same packet path instead of bypassing server authority.
