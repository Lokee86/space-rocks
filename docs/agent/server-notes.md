# Agent Server Notes

Use this when changing the Go game server, networking, rooms, gameplay simulation, packet codec, logging, or the planned API boundary.

## Server Responsibilities

The Go game server owns authoritative simulation outcomes:

- movement
- bullets
- collisions
- scoring
- lives
- death
- respawn
- pause safety
- rooms
- websocket state

Keep reusable game simulation in:

```text
services/game-server/internal/game
```

Do not put reusable simulation logic in:

```text
cmd/game-server/main.go
```

## Game Server Layout

The Go game server was moved from `server/` to:

```text
services/game-server/
```

Its Go module path is still:

```text
github.com/Lokee86/space-rocks/server
```

That mismatch is currently intentional. Import paths inside the Go server still use `github.com/Lokee86/space-rocks/server/...`.

## Planned API Boundary

`services/api-server/` exists as an empty placeholder directory for a future Node.js/TypeScript/NestJS API service.

Do not put these concerns into the Go game server unless the user explicitly changes that direction:

- account logic
- persistence
- matchmaking metadata
- leaderboard
- other business/backend concerns

The planned API service should own business/backend concerns, not real-time simulation.

See:

```text
docs/api/nestjs-api-server.md
```

## Networking / Rooms / Game Ownership

- Keep websocket and room transport in `services/game-server/internal/networking`.
- Networking should transport, decode/route packets, manage websocket session state, and write responses.
- Room lifecycle policy belongs in rooms.
- Gameplay simulation belongs in game.
- `rooms` owns room creation, joining, leaving, readiness, lifecycle transitions, cleanup policy, and game instance ownership.
- `networking` may retain websocket session activation/deactivation when it mutates websocket session fields.
- `game` owns authoritative gameplay simulation, gameplay state mutation, and adapters from game storage into narrower gameplay seams.
- Match/mode policy evaluation belongs in `services/game-server/internal/game/rules`, which should receive plain snapshots/facts and return decisions/status.
- `game` should not own websocket transport, API persistence, account/auth concerns, or lobby UI flow.

## Server Packet Codec

Route server packet wire JSON through:

```text
services/game-server/internal/protocol/packetcodec
```

Do not add direct `encoding/json` calls in server packet wire paths.

The codec is intentionally JSON-only and generic. Do not add format switching, protobuf references, or an interface unless explicitly requested.

Non-packet JSON such as collision-shape data-file parsing may still use `encoding/json` directly.

## Spatial / Movement Rules

Use:

```text
services/game-server/internal/game/motion
```

for per-entity movement integration and advance-with-wrap behavior.

Use:

```text
services/game-server/internal/game/space
```

for new gameplay distance, direction, and wrap-aware spatial math for the toroidal world.

Current/future wrapped-world rules:

- server coordinates should be bounded/wrapped
- respawn safety uses wrapped distance
- ship/asteroid and projectile/asteroid collision helpers place temporary asteroid bodies in wrapped-local space before collision checks

See:

```text
docs/design/toroidal-wrap.md
```

## Logging

The server has a custom structured logging wrapper:

```text
services/game-server/internal/logging/logger.go
```

Use it for server logs. Do not add raw `log.Println` or a second logging package.

Category loggers:

- `logging.Server`
- `logging.Network`
- `logging.Rooms`
- `logging.Game`

Environment variables:

- `LOG_LEVEL`
- `LOG_SERVER`
- `LOG_NETWORK`
- `LOG_ROOMS`
- `LOG_GAME`

Supported levels include:

- `debug`
- `info`
- `warn`
- `warning`
- `error`
- `off`

Default is quiet: unset `LOG_LEVEL` resolves to `warn`.

Normal lifecycle logs should usually be `Debug`. Warnings are for unusual recoverable situations. Errors are for real failed operations.

See:

```text
docs/server/logging.md
```

## Server Test Rules

Go server tests live under:

```text
services/game-server/tests/<area>/
```

Current areas include:

- `game`
- `networking`
- `physics`
- `rooms`
- `scoring`
- `space`

Do not add new `*_test.go` files beside production packages under `services/game-server/internal/`.

For game simulation setup, use the shared harness in:

```text
services/game-server/tests/game/helpers_test.go
```

Keep new helpers intent-level, such as placing entities or sending packets, instead of exposing raw private maps.

For server gameplay changes, run:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

If the command prints read-only `envman` warnings but tests pass, those warnings have been harmless in this environment.
