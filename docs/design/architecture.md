# Architecture

Space Rocks is an Asteroids-inspired game with a Godot client and a Go server. The current direction is server-authoritative for gameplay state where networking is involved: player input is sent to the server, the server advances simulation, and clients render the state they receive.

The project is still in development, so this document describes the architecture that exists now and calls out future notes separately.

## Repository Layout

- `client/`: Godot project. Contains scenes, scripts, assets, audio, shaders, and client-side tools.
- `services/game-server/`: Go module for the real-time game server. The current entrypoint is `services/game-server/cmd/game-server`.
- `services/api-server/`: empty placeholder for a planned Node.js/TypeScript NestJS API server for business/backend systems. It is intentionally separate from real-time simulation.
- `shared/`: source data shared across client and server generation, including TOML constants, TOML packet definitions, and JSON collision shape data.
- `docs/`: Project documentation.
- `tools/`: Python scripts used to generate constants and packet code from `shared/`.

## Client Architecture

The Godot client is responsible for presentation, local input collection, UI, audio/effects, and websocket communication with the Go server. It does not own authoritative gameplay outcomes like scoring, lives, asteroid collision results, or respawn validity.

The configured Godot main scene is:

```text
client/scenes/game.tscn
```

Key client pieces:

- `client/scripts/ui/game_shell.gd`: top-level shell for menu/game-loop scene switching and always-on parallax background scrolling.
- `client/scenes/ui/main_menu.tscn` and `client/scripts/ui/main_menu.gd`: main menu controls for single-player, multiplayer dialog launch, and quit.
- `client/scenes/game_loop.tscn` and `client/scripts/game.gd`: active gameplay scene/controller. Creates the network client, world sync, HUD controller, and effects controller.
- `client/scripts/networking/network_client.gd`: wraps Godot `WebSocketPeer`, handles connect, poll, send, graceful close, and packet parsing.
- `client/scripts/networking/world_sync.gd`: applies server state to local/remote player, bullet, and asteroid nodes. It tracks local server coordinates separately from continuous visual coordinates so rendering can cross wrapped world edges without snapping.
- `client/scripts/world_wrap.gd`: client-side toroidal wrap math using generated world-size constants.
- `client/scripts/entities/player.gd`: collects input into packet data, plays local laser audio, and toggles local afterburner visuals.
- `client/scripts/effects.gd`: spawns local visual/audio effects for bullet impacts, ship death, and game over sound timing.
- `client/scripts/ui/hud_controller.gd`: updates score, lives, room ID, death overlay, respawn state, and game-over UI.
- `client/scripts/networking/packets.gd` and `client/scripts/constants/constants.gd`: generated/shared client packet helpers and constants.

Rendering is scene/node based in Godot. The client renders the ship, asteroids, bullets, background, UI, animations, and audio. The background has local auto-scroll in `game_shell.gd`; gameplay scroll offset follows the local player's continuous visual position after initial spawn.

Input is collected locally every frame and sent to the server as an input packet when connected. Respawn requests are sent as explicit packets. The client also sends visible viewport configuration so the server can tie spawning/visibility to the player's camera view.

Current limitations:

- The client expects a Go server at `ws://localhost:8080/ws` unless a room ID is supplied.
- There is no implemented client-side prediction beyond interpolation/render smoothing.
- Local server launch from the Godot client is not implemented in the inspected code.

## Game Server Architecture

The game server is a Go module under `services/game-server/`.

The main entrypoint is:

```text
services/game-server/cmd/game-server/main.go
```

`main.go` currently:

- configures server logging from environment variables
- creates an HTTP mux
- creates a room manager
- registers `GET /health`
- registers `GET /ws`
- starts HTTP on `:8080`

Core server packages:

- `services/game-server/internal/networking`: websocket transport, packet dispatch, session registry, and outbound writes.
- `services/game-server/internal/rooms`: room state, room membership, lifecycle orchestration, and cleanup policy.
- `services/game-server/internal/game`: game loop, state packets, combat, spawning, scoring, respawn/session logic, visibility.
- `services/game-server/internal/game/entities`: game entities and generated packet state structs.
- `services/game-server/internal/game/physics`: collision shapes, collision detection, vectors, and shared collision shape loading.
- `services/game-server/internal/game/space`: gameplay spatial helpers for wrapped distance, direction, shortest delta, and position normalization.
- `services/game-server/internal/constants`: generated Go constants from `shared/game_data.toml`.
- `services/game-server/internal/logging`: structured `slog` wrapper with categories and environment-controlled levels.

### Game Loop And Simulation

Each `game.Game` owns its own simulation state:

- players
- bullets/projectiles
- asteroids
- player sessions
- camera views
- pending events

`Game.Start()` launches a simulation loop at `constants.ServerTickRate`. Each tick applies player input, moves entities, wraps moving entities into world bounds, handles cooldowns, spawns asteroids, removes expired/far objects, and resolves collisions.

The server currently owns:

- player movement simulation from input
- bullet spawning
- asteroid spawn scheduling, planning, application, and visibility removal
- projectile/asteroid collision facts for current bullet hits
- ship/asteroid collision
- entity damage/destruction resolution
- asteroid splitting
- scoring
- lives, death, game-over, and respawn rules
- safe initial spawn/respawn placement
- state packet generation

### Spawn Planning

`services/game-server/internal/game/spawn_types.go` defines the player spawn vocabulary currently used by the game package:

- `SpawnEntityType`
- `SpawnReason`
- `PlayerSpawnPlan`

`services/game-server/internal/game/spawning` owns asteroid/projectile spawning policy:

- `Spawner`
- `AsteroidSpawnPlan`
- bullet ID allocation and bullet construction
- asteroid ID allocation
- timed asteroid plan construction
- asteroid fragment plan construction

The vocabulary and plans stay entity-specific. There is no universal optional-field spawn request/plan object.

Current implemented seam:

```text
Game.Step/combat/session decides when spawn is needed
  -> entity-specific planner selects spawn facts
  -> entity-specific apply/lifecycle code mutates game state
```

Timed asteroid scheduling still belongs to `Game.Step()`, and `spawnAsteroidBatch()` still owns the timed batch count. `spawnAsteroid()` still selects the target camera position and offscreen wrapped spawn position, then asks `spawning.Spawner` to build the timed `AsteroidSpawnPlan`. Combat still decides when fragments are needed; `spawnAsteroidFragments()` keeps the split log in `Game` and asks `spawning.Spawner` for fragment plans. `applyAsteroidSpawn()` remains in `Game` as the bridge that requests an asteroid ID from `spawning.Spawner`, constructs the entity, and mutates `game.state.Asteroids`.

Player initial spawn and respawn planning now use `PlayerSpawnPlan`. `planInitialPlayerSpawn()` preserves the existing `playerIndex`-based preferred position plus safe-spawn fallback. `planPlayerRespawn()` receives the already-gated `*playerSession` and preserves the existing `safeRespawnPosition()` behavior. Player lifecycle still owns session lookup, `CanRespawn()` gating, lives, death, respawn cooldowns, ship creation, camera view attachment, and game-over/session state.

This is still a partial seam. Bullet construction now lives in `spawning.Spawner`, but `spawnBullet()` remains the `Game` adapter that inserts the projectile into `game.state.Projectiles`. Do not add enemies, powerups, waves, spawn packets, or client behavior through this seam until those systems exist.

### Entity Damage Resolution

`services/game-server/internal/game/collisions.go` defines a narrow collision detection seam for the destructive collision pairs currently used by combat. It reports concrete collision facts:

- `ProjectileAsteroidCollision`
- `PlayerAsteroidCollision`

The helpers answer only which pair collided and which impact position preserves the current event behavior. Current projectile/asteroid impact position remains the bullet position because bullets are the only projectile implementation. Player/asteroid impact position remains the player position. The seam intentionally does not use `physics.Collision.ContactPoint` for these event positions.

`services/game-server/internal/game/damage.go` defines the internal damage resolution seam for authoritative entity damage and destruction decisions. Combat consumes collision facts, builds a `DamageRequest`, calls `resolveDamage`, and then existing systems consume the `DamageResult`.

Current behavior is intentionally unchanged:

```text
bullet hits asteroid -> projectile damage resolves Destroyed -> existing asteroid despawn, fragmentation, scoring, and bullet blast flow
ship hits asteroid -> collision damage resolves Fatal for player -> existing death, lives, respawn cooldown, logging, and ship death flow
```

The collision helpers only answer whether a projectile/asteroid or player/asteroid pair overlapped in the current wrapped world, plus the preserved impact position. The projectile helper still accepts `*entities.Bullet` because bullets are the current projectile entity. They do not mutate entities, build damage requests, resolve damage, award score, spawn fragments, decrement lives, set respawn cooldowns, emit events, log, or write packets.

The damage resolver only answers what happened to the target from the damage request. It does not mutate lives, respawn players, award score, spawn fragments, emit events, log, or write packets. Those lifecycle effects remain with combat/session/scoring/spawning and the domain event seam.

The initial model already carries general target/source identity and future fields for shields, invulnerability bypass, health, and shield absorption, but no shield or health storage mechanics are active yet. Keep future durability work routed through this seam without moving scoring, spawning, or player lifecycle ownership into the resolver.

### Domain Gameplay Events

`services/game-server/internal/game/events.go` defines a small internal domain event seam for gameplay facts that already produce client-visible packet events. The first supported facts are bullet blasts and ship deaths.

Gameplay systems such as combat still own gameplay decisions: collision outcomes, scoring, lives, death, respawn, and spawning rules remain in their existing files. The event seam records internal facts and translates them to the current packet-facing `EventState` output where needed.

Current event flow:

```text
combat/scoring/spawning/lives code
  -> internal domain event
  -> packet-facing EventState
  -> per-player pending event queue
  -> StatePacket.Events
```

The seam should stay narrow. It is not an achievements system, stats system, API integration point, persistence layer, logging policy, or replacement home for gameplay rules. Future systems may consume richer domain facts, but the event seam itself should only define, record, drain, and translate events needed by existing behavior.

### Toroidal World Wrap

The server stores bounded wrapped world coordinates using `constants.WorldWidth` and `constants.WorldHeight`. `Game.Step()` centrally normalizes moving players, asteroids, and bullets after movement.

Spatial rules flow through `services/game-server/internal/game/space`:

- spawning aim uses wrapped direction
- visibility/despawn uses wrapped delta
- respawn safety uses wrapped distance
- ship/asteroid and projectile/asteroid collision helpers place temporary asteroid bodies in wrapped-local space before collision checks

The client renders continuous visual coordinates. `world_sync.gd` tracks `local_server_position` and `local_visual_position`; remote players, asteroids, bullets, and server-driven effects render relative to the local player with shortest wrapped deltas. The camera and background follow the local player node, so they inherit the continuous visual position.

See [toroidal wrap](toroidal-wrap.md).

### Rooms And Networking

Room/domain ownership lives in `services/game-server/internal/rooms`. That package owns:

- `Room`
- `RoomManager`
- `RoomMember`
- room state and error-code constants
- max room capacity and room-code/default-room helpers
- create, join, leave, ready, start-game, single-player startup, return-to-lobby, game-over transition, and cleanup decisions
- ownership of each room's `*game.Game` lifecycle, while simulation rules stay in `internal/game`

`services/game-server/internal/networking` owns websocket/session/packet transport. It upgrades `/ws`, reads generated packets, calls room-domain methods, attaches or clears websocket session player IDs, and sends or broadcasts generated packets such as `RoomSnapshot` and `RoomError`.

The websocket connection itself is session-only. Room membership happens through packets:

- `CreateRoomRequest`
- `JoinRoomRequest`
- `LeaveRoomRequest`
- `SetReadyRequest`
- `StartGameRequest`
- `ReturnToLobbyRequest`

Important lifecycle rules:

- connection does not imply room membership
- room membership does not imply an active game player
- active ships/game players are created only when `StartGameRequest` succeeds
- websocket session activation/deactivation stays in networking because it mutates per-connection session fields
- `/ws?room_id=...` no longer creates or joins rooms
- empty rooms schedule cleanup after members/active players leave

On websocket connect:

- the server creates a session identity
- one goroutine reads client input packets
- one write path sends queued packets/state

On disconnect:

- disconnect routes through leave-room behavior when the session is in a room
- active game players are removed when needed
- remaining members receive `RoomSnapshot`
- empty rooms schedule cleanup

### Physics

The physics package provides collision primitives and collision detection for circles, capsules, rectangles, and polygons. Collision shapes are loaded from:

```text
shared/collisions/collision_shapes.json
```

The server uses imported collision shapes for ship, bullet, and asteroid collision bodies. Gameplay collision ownership is split deliberately: `physics` owns primitive overlap math, `collisions.go` owns current game-pair collision facts, and `combat.go` owns damage/death/destruction/scoring/event orchestration.

### Logging And Config

Server logging is implemented in:

```text
services/game-server/internal/logging/logger.go
```

It uses `log/slog`, logs to stderr, and supports category loggers:

- `logging.Server`
- `logging.Network`
- `logging.Rooms`
- `logging.Game`

Configuration is environment-variable based. See [server logging](../server/logging.md).

Client logging is implemented in:

```text
client/scripts/logging/logger.gd
```

Use `ClientLogger` for new client lifecycle, UI, networking, packet, HUD, input, and world-sync diagnostics. See [client logging](../client/logging.md).

## NestJS API Server Plan

`services/api-server/` is currently an empty placeholder reserved for a separate business/backend API service. The intended stack is Node.js, TypeScript, and NestJS.

This service is not implemented yet. The purpose of the separate service is to keep business logic physically and technically separate from the real-time Go game server.

Planned API-owned concerns include:

- accounts and authentication
- profiles
- matchmaking or room discovery metadata
- leaderboards
- unlocks/cosmetics
- persistence and database-backed workflows
- admin or moderation endpoints

The API server should not own real-time game simulation. The Go game server should remain responsible for live rooms, websocket gameplay, collisions, scoring during a match, lives, death, respawn, and authoritative state packets.

See [NestJS API server plan](../api/nestjs-api-server.md).

## Data Flow

The current runtime data flow is:

1. Godot collects input in `player.gd`.
2. `game.gd` sends input/client-config/respawn packets through `network_client.gd`.
3. The Go websocket handler reads packets and passes them to the room's `game.Game`.
4. The game simulation applies input and advances authoritative state.
5. The server writes `StatePacket` JSON back to the client.
6. `game.gd` receives the packet and passes state to `world_sync.gd`.
7. `world_sync.gd` creates/removes/interpolates rendered nodes.
8. HUD/effects/audio update from state and events.

Shared packet structures are sourced from:

```text
shared/packets/packets.toml
```

Generated packet files include:

- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/entities/packets_generated.go`
- `client/scripts/networking/packets.gd`

Shared constants are sourced from:

```text
shared/game_data.toml
```

Generated constants include:

- `services/game-server/internal/constants/constants.go`
- `client/scripts/constants/constants.gd`

Server-owned constants live under `constants.server.*` and may be omitted from client generated constants. World size is intentionally generated to both Go and GDScript because client visual wrapping must use the same bounds as the server. In particular, `player_starting_lives` and `player_respawn_delay` live under `constants.server.player_lifecycle`, while `asteroid_size_scale` lives under `constants.server.asteroids`. The client receives lives through player state, respawn delay through death events, and asteroid scale through asteroid state instead of importing those constants.

Authoritative today:

- server simulation state
- player lives/death/game-over state
- score
- asteroid splits and despawns
- safe spawn/respawn placement
- bullet and asteroid collision outcomes

Client-owned today:

- rendering
- interpolation
- menus/UI presentation
- local audio/effects playback
- websocket connection lifecycle

Current limitations:

- No account, matchmaking, leaderboard, or persistent backend API is implemented.
- No prediction/reconciliation layer is implemented beyond interpolation.
- The server is expected to be running separately for the Godot client.

## Design Rules And Conventions

- Keep authoritative gameplay logic on the server unless client prediction/interpolation is explicitly being added.
- Do not duplicate scoring, lives, respawn safety, collision outcomes, or asteroid split rules in the client.
- Keep network transport separate from core game simulation. Websocket code should live in `services/game-server/internal/networking`; game rules should live in `services/game-server/internal/game`.
- Keep reusable simulation code out of `main.go`. The server entrypoint should register routes, configure dependencies, and start the process.
- Use `shared/game_data.toml`, `shared/packets/packets.toml`, and `tools/data_sync/` for packet and constant data that must stay aligned across Go and Godot.
- Do not hand-edit generated files unless the generator/source data is intentionally being bypassed.
- Do not commit generated recordings or build artifacts. `.gitignore` excludes `tmp/`, Godot export/import state, and `*.avi`.
- Do not put secrets in client code. The client should be treated as inspectable.
- Prefer focused tests for game rules that are easy to regress, especially collision, spawning, respawn, rooms, and packet behavior.

## Future Architecture Notes

These are possible directions, not implemented features.

- Local play packaging may eventually launch or bundle the Go game server with the Godot client.
- A hosted online game server may use the same room/websocket structure with deployment-specific process management.
- A separate backend API server may be useful for non-gameplay systems such as accounts, matchmaking, leaderboards, persistence, or purchases.
- Matchmaking/accounts/leaderboards are not current features. If added, they should stay separate from the real-time game simulation unless a clear shared boundary is needed.
- If prediction/reconciliation is added, keep it explicitly separate from authoritative game rules so the client remains a presentation/prediction layer rather than the source of truth.
- Invisible toroidal/wrapped playfield is implemented. See [toroidal wrap](toroidal-wrap.md).
- A thin server-side ship variant foundation exists: runtime ship type, resolved ship stats/modifiers, `ship_type` snapshots, and collision shape ID lookup. Full variants with client scene mapping and keyed collision catalogs remain future work. See [ship variants plan](ship-variants.md).
