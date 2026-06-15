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

`services/api-server/` exists as a Ruby/Rails API-only scaffold.

Do not put these concerns into the Go game server unless the user explicitly changes that direction:

- account logic
- persistence
- matchmaking metadata
- leaderboard
- other business/backend concerns

The planned API service should own business/backend concerns, not real-time simulation.

See:

```text
docs/api/ruby-api-server.md
```

Rails/Godot auth session handoff now exists, and Go websocket auth now has an explicit Rails token-verification boundary plus a Go authclient seam. `services/game-server/internal/authclient` verifies Space Rocks bearer tokens through Rails and returns minimal identity for the game server to use at the websocket boundary, not by reading Rails tables directly.
The current server-side auth seams are the Go authclient, then the game-server session identity seam, then multiplayer websocket auth/admission.
The websocket auth handshake uses `authenticate_request` and `authenticate_result`, with `SessionIdentity` remaining separate from `AccountUserID` so the game server can keep session identity distinct from account identity.

Server-side account and local-profile routing must follow [docs/design/cross-mode-routing-and-player-data.md](../design/cross-mode-routing-and-player-data.md): `services/player-data` now has the Phase 4 routes for `authenticated_account` through the Rails adapter, `local_profile` through the SQLite-backed route in the standard no-tag development build, and `guest` through singleton memory; `cmd/game-server` hosts the configured player-data runtime in-process for now, gameplay is still not wired to emit player-data packets, and `services/player-data` must not import `services/game-server/internal`. The game-server must still not directly write SQLite or Rails/Postgres tables.
Account-shaped player data must also follow [docs/design/player-data-schema-ssot.md](../design/player-data-schema-ssot.md): logical contract work must keep Local Profile and Authenticated Account concepts aligned, Rails/Postgres and embedded DB may differ physically but must satisfy the same logical contract, gameplay code must not depend directly on Rails tables or embedded DB tables, and the data-sync pipeline remains the packet/schema source of truth.
Player-data runtime extraction is later work; the Rails adapter exists now, and the embedded SQLite backing store is present in the standard no-tag development build.
Embedded SQLite now lives under `services/player-data/playerdata/embeddedsqlite` and is compiled by default, while `-tags noembeddedsqlite` deployment/restricted builds exclude it. `-tags noembeddedsqlite` builds must not import or depend on `modernc.org/sqlite`, the core `playerdata` package must not import the embedded SQLite package, and local-store construction is injected from the game-server composition root.

## Networking / Rooms / Game Ownership

- Keep websocket/session transport and adapter wiring in `services/game-server/internal/networking`.
- Keep inbound packet-family handlers in `services/game-server/internal/networking/inbound`.
- Keep outbound gameplay presentation/write helpers in `services/game-server/internal/networking/outbound`.
- Networking should transport, decode/route packets, manage websocket session state, and write responses.
- Room lifecycle policy belongs in rooms.
- Gameplay simulation belongs in game.
- `rooms` owns room creation, joining, leaving, readiness, lifecycle transitions, cleanup policy, and game instance ownership.
- Room start now uses a deliberate `Lobby -> Starting -> InGame` transition. `Starting` is an admission-closed handoff state for pre-game coordination, including future slow-client handling, final readiness or sync steps, and other pre-match server work before the room becomes `InGame`.
- Multiplayer start validation is centralized in the room lifecycle path, and `RoomManager` should resolve the room, session, and player identity before delegating to `Room` for the actual transition.
- If the game-server auth verifier is configured, multiplayer create/join requires an Authenticated Account identity; if it is not configured, local/no-auth multiplayer can still proceed because server-side admission remains authoritative.
- `networking` may retain websocket session activation/deactivation when it mutates websocket session fields.
- `game` owns authoritative gameplay simulation, gameplay state mutation, and adapters from game storage into narrower gameplay seams.
- `game` must not know SQLite table names or schema details for local-profile storage.
- Asteroid variant metadata lives in `shared/asteroids/variants.toml`. Variant indexes are zero-based runtime values, stable ids like `asteroid_1` are presentation names, and spawn code must choose variants through the asteroid catalog rather than through `constants.AsteroidVariants`.
- Damage resolution lives in `services/game-server/internal/game/damage/`; it owns pure resolution only. `game` owns the entity mutation adapters, and devtools must route damage through the same real damage seam rather than a parallel debug-only path. See `docs/design/damage.md`.
- Weapons live in `services/game-server/internal/game/weapons` and radial effects live in `services/game-server/internal/game/effects/radial`. Weapon profiles may carry impact effects, torpedo uses a radial impact effect, radial effects emit hit intents, and Game applies radial hits through the damage seam. See `docs/design/weapons.md` and `docs/design/radial-effects.md`.
- Match/mode policy evaluation belongs in `services/game-server/internal/game/rules`, which should receive plain snapshots/facts and return decisions/status.
- `game` should not own websocket transport, API persistence, account/auth concerns, or lobby UI flow.

### Match Result Summaries

- The existing room/game-over lifecycle remains the match-over authority.
- Rooms store one resolved `MatchResultSummary` on `game_over`.
- `Game` exposes `PlayerMatchFacts`.
- Rooms assign stable match IDs per started match.
- `matchreporting` maps each player summary to `player_data_record_match_result` and sends those commands through the hosted `services/player-data` runtime.
- `services/player-data` owns account/local/guest store routing.
- `result_id` is `match_id + ":" + game_player_id`.
- Successful reports are marked and not repeated.
- Failed reports are not marked successful.
- The Go server does not write Rails/Postgres or SQLite directly.

### Server Identity Policy

- `PlayerID` is permanent and player-facing.
- `PlayerID` values are readable labels like `Player-1`, `Player-2`, `Player-3`.
- `PlayerID` must not be converted to UUID.
- `SessionID` is server-internal websocket/session identity and should move to UUID during the internal UUID upgrade.
- `MemberID` is server-internal room-membership identity, currently UUID v4, and is the future reconnect seam.
- `MemberID` should not be added back to normal room snapshot packets.
- `currentGamePlayerID` is networking active-game routing state, not a room membership or public identity.
- Keep `SessionID` / `GamePlayerID` separate from `AccountUserID`; game-server session identity and account identity are different seams.
- Single-player must not require Rails auth.
- Online multiplayer now authenticates through the Go websocket boundary, and when the auth verifier is configured multiplayer admission requires an Authenticated Account.

## Server Packet Codec

Server-facing packet schema source of truth is split under `shared/packets/`:

- `shared/packets/outputs.toml`
- `shared/packets/gameplay.toml`
- `shared/packets/debug.toml`
- `shared/packets/lobby.toml`

Route server packet wire JSON through:

```text
services/game-server/internal/protocol/packetcodec
```

Do not add direct `encoding/json` calls in server packet wire paths.

The codec is intentionally JSON-only and generic. Do not add format switching, protobuf references, or an interface unless explicitly requested.

Non-packet JSON such as collision-shape data-file parsing may still use `encoding/json` directly.

- `shared/collisions/collision_shapes.json` is generated by `client/tools/export_collision_shapes.gd`.
- It is non-packet JSON and may be parsed directly by server collision code.
- Prefer changing Godot scene collision nodes and rerunning the exporter instead of hand-editing the JSON.

### Server Devtools Packets

- Devtools packet schema belongs in `shared/packets/debug.toml`.
- Devtools output routing belongs in `shared/packets/outputs.toml`.
- Generated server devtools packet output belongs in `services/game-server/internal/devtools/packets_generated.go`.
- Devtools packet routing goes through `services/game-server/internal/networking/inbound`, then dispatches mutation work to `services/game-server/internal/devtools`.
- Server-supported debug actions include `debug_kill_player`, `debug_spawn_entity`, `debug_respawn_player`, and freeze/invincible/infinite-lives toggles.

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
