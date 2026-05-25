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
- `client/scripts/networking/network_client.gd`: wraps Godot `WebSocketPeer`, handles connect, poll, send, graceful close, and packet signals.
- `client/scripts/networking/packet_codec/packet_codec.gd`: JSON-only client packet wire encode/decode wrapper around `JSON.stringify` and `JSON.parse_string`.
- `client/scripts/networking/world_sync.gd`: thin coordinator for server-state rendering. It configures sync dependencies, preserves apply/interpolation ordering, forwards bullet-spawn signals, and exposes compatibility accessors to `game.gd`.
- `client/scripts/networking/local_visual_sync.gd`: owns local server position, continuous local visual position, and server-position-to-visual-position conversion for wrapped rendering.
- `client/scripts/networking/player_sync.gd`: owns player nodes, remote hues, target positions/rotations, pause visibility, remote visual positions, cleanup, packet application, and interpolation.
- `client/scripts/networking/bullet_sync.gd`: owns bullet nodes, target positions/rotations, cleanup, packet application, interpolation, and bullet-spawn signal decisions.
- `client/scripts/networking/asteroid_sync.gd`: owns asteroid nodes, scale warning state, target/server/visual positions, visual continuity, cleanup, packet application, and interpolation.
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
- `services/game-server/internal/game/motion`: per-entity movement integration and advance-with-wrap helpers for ships, asteroids, and bullets.
- `services/game-server/internal/game/rules`: match/mode policy evaluation from plain snapshots. It currently owns game-over outcome evaluation and per-player participation classification.
- `services/game-server/internal/game/scoring`: pure score policy evaluation. It converts scoring events into awards without mutating game sessions.
- `services/game-server/internal/game/entities`: game entities and generated packet state structs.
- `services/game-server/internal/game/physics`: collision shapes, collision detection, vectors, and shared collision shape loading.
- `services/game-server/internal/game/space`: gameplay spatial helpers for wrapped distance, direction, shortest delta, and position normalization.
- `services/game-server/internal/constants`: generated Go constants from `shared/game_data.toml`.
- `services/game-server/internal/logging`: structured `slog` wrapper with categories and environment-controlled levels.
- `services/game-server/internal/protocol/packetcodec`: JSON-only packet wire encode/decode helpers for server packets.

### Game Loop And Simulation

Each `game.Game` owns its own simulation state:

- players
- bullets/projectiles
- asteroids
- player sessions
- camera views
- pending presentation events

`Game.Start()` launches a simulation loop at `constants.ServerTickRate`. Each tick advances player sessions, advances moving entities through the motion seam, updates camera views, spawns asteroids, removes expired/far objects, and resolves collisions.

The server currently owns:

- player movement simulation from input
- bullet spawning
- asteroid spawn scheduling, planning, application, and visibility removal
- projectile/asteroid collision facts for current bullet hits
- ship/asteroid collision
- entity damage/destruction resolution
- asteroid splitting
- scoring
- lives, death, and respawn state
- match-over policy evaluation
- safe initial spawn/respawn placement
- state packet generation

`Game.Step()` is now a small same-package simulation coordinator in `services/game-server/internal/game/simulation.go`. It preserves the authoritative phase order while routing player/session, asteroid, bullet, and collision phases through focused same-package helpers:

```text
Game.Step()
  -> stepPlayerSessions
  -> stepPlayers
  -> removeReadyPlayers
  -> stepAsteroidSpawning
  -> stepAsteroids
  -> stepBullets
  -> stepCollisions
```

The focused simulation helpers still mutate `Game` state under the `Game` mutex; mutex ownership has not moved out of `Game`. Per-entity movement integration and wrapping live in `services/game-server/internal/game/motion`:

```text
Game.Step()
  -> motion.AdvanceShip / AdvanceAsteroid / AdvanceBullet
  -> motion.StepShip / StepAsteroid / StepBullet
  -> space.WrapPosition through motion's bounds-aware helper
```

The motion package imports entity types and spatial helpers, but it does not import `internal/game`, mutate `game.state` maps, spawn entities, delete entities, award score, resolve collisions, or write packets. `entities` remains the home for entity state, pause/resume/input capability methods, collision bodies, and packet-facing state conversion.

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

Timed asteroid scheduling belongs to the same-package `stepAsteroidSpawning()` helper in `simulation_asteroids.go`, and `spawnAsteroidBatch()` still owns the timed batch count. `spawnAsteroid()` still selects the target camera position and offscreen wrapped spawn position, then asks `spawning.Spawner` to build the timed `AsteroidSpawnPlan`. Combat still decides when fragments are needed; `spawnAsteroidFragments()` keeps the split log in `Game` and asks `spawning.Spawner` for fragment plans. `applyAsteroidSpawn()` remains in `Game` as the bridge that requests an asteroid ID from `spawning.Spawner`, constructs the entity, and mutates `game.state.Asteroids`.

Player initial spawn and respawn planning now use `PlayerSpawnPlan`. `planInitialPlayerSpawn()` preserves the existing `playerIndex`-based preferred position plus safe-spawn fallback. `planPlayerRespawn()` receives the already-gated `*playerSession` and preserves the existing `safeRespawnPosition()` behavior. Player lifecycle still owns session lookup, `CanRespawn()` gating, lives, death, respawn cooldowns, ship creation, and camera view attachment. Match-over policy is evaluated through `services/game-server/internal/game/rules`.

This is still a partial seam. Bullet construction now lives in `spawning.Spawner`, but `spawnBullet()` remains the `Game` adapter that inserts the projectile into `game.state.Projectiles`. Do not add enemies, powerups, waves, spawn packets, or client behavior through this seam until those systems exist.

### Match Rules

`services/game-server/internal/game/rules` owns match/mode policy decisions from plain facts. It does not import `internal/game` or `internal/rooms`, does not mutate `Game`, and does not inspect game storage directly.

The current seam is intentionally small:

```text
Game/session state
  -> rules.MatchSnapshot
  -> rules.EvaluateMatch
  -> rules.MatchDecision
```

`Game.matchSnapshot()` is the adapter from `Game` internals to rules facts. It creates one `rules.PlayerSnapshot` per player session and maps:

- `ID`
- `HasRemainingLives` from `session.Lives > 0`
- `HasActiveShip` from whether `game.state.Players` contains that session ID

`Game.MatchDecision()` is the public game-facing API for richer match decisions. It locks the game, evaluates the rules snapshot, and returns `rules.MatchDecision`. `Game.IsGameOver()` remains available for existing callers and delegates through the same locked decision path.

`MatchDecision` currently reports:

- `IsOver`
- one player decision per session
- `PlayerActive` for players with an active ship
- `PlayerPendingRespawn` for players with remaining lives and no active ship
- `PlayerEliminated` for players with no active ship and no remaining lives

Current match-over behavior is preserved:

- no player sessions means not game over
- any session with remaining lives means not game over
- any active ship means not game over
- otherwise the match is over

This policy is based on remaining lives, not `session.CanRespawn()`. `CanRespawn()` includes respawn cooldown eligibility and remains a session/player-lifecycle gate for respawn requests, not a match outcome rule. Rooms still own room lifecycle transitions such as `InGame` to `GameOver`; they should use game-facing APIs such as `Game.MatchDecision().IsOver` or `Game.IsGameOver()` rather than importing rules.

`Game.statePacket()` projects `MatchDecision.Players` into `StatePacket.player_lifecycle`, a map from player ID to lifecycle status string. This field sits beside `StatePacket.players`; it is not part of `ShipState` because pending-respawn and eliminated players may have no active ship. `StatePacket.players` remains active ship/render state only.

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

`combat.go` keeps the scan and consequence phases explicit. Projectile/asteroid handling scans projectile and asteroid maps, detects overlap, builds a projectile damage request, resolves damage, and records confirmed hits during the scan. It applies consequences only after the nested scan: score awards, bullet despawn marking, asteroid despawn marking, then fragment spawning. Player/asteroid handling likewise scans player and asteroid maps, gates with `CanTakeCollisionDamage()`, detects overlap, builds a collision damage request, resolves damage, records fatal players in `hitPlayers`, and applies death/session/event consequences only after the nested scan.

The collision helpers only answer whether a projectile/asteroid or player/asteroid pair overlapped in the current wrapped world, plus the preserved impact position. The projectile helper still accepts `*entities.Bullet` because bullets are the current projectile entity. They do not mutate entities, build damage requests, resolve damage, award score, spawn fragments, decrement lives, set respawn cooldowns, emit events, log, or write packets.

The damage resolver only answers what happened to the target from the damage request. It does not mutate lives, respawn players, award score, spawn fragments, emit events, log, or write packets. Those lifecycle effects remain with combat/session/scoring/spawning and the domain event seam.

The initial model already carries general target/source identity and future fields for shields, invulnerability bypass, health, and shield absorption, but no shield or health storage mechanics are active yet. Keep future durability work routed through this seam without moving scoring, spawning, or player lifecycle ownership into the resolver.

### Scoring Policy

`services/game-server/internal/game/scoring` owns pure score policy. It receives scoring facts as `scoring.Event` values and returns `scoring.Award` values. It does not import `internal/game`, inspect game storage, mutate players or sessions, check pause/invulnerability state, log, emit packets, or write events.

Current scoring flow:

```text
combat confirms asteroid destroyed by projectile
  -> scoring.Event{PlayerID, TargetID, AsteroidSize}
  -> scoring.Policy.Evaluate()
  -> []scoring.Award
  -> game.awardScore()
  -> player/session score mutation and logging
```

Asteroid score behavior is unchanged:

```text
constants.BaseScore / asteroid.Size
```

`game.awardScore` remains the game-owned application seam. It applies awards to active game state and keeps missing-player, paused-player, invulnerable-player, session sync, and score logging behavior outside the scoring policy package.

### Domain Gameplay Events

`services/game-server/internal/game/events` defines the small domain event vocabulary for gameplay facts that already produce client-visible packet events. The first supported facts are bullet blasts and ship deaths.

Gameplay systems such as combat still own gameplay decisions: collision outcomes, scoring, lives, death, respawn, and spawning rules remain in their existing files. Current producers record `events.Event` values. Root `services/game-server/internal/game/events.go` remains the package-local presentation adapter that translates domain events to the generated packet-facing `EventState` output where needed.

Current event flow:

```text
combat/scoring/spawning/lives code
  -> events.Event
  -> game/events.go packet-facing EventState adapter
  -> per-player pendingPresentationEvents queue
  -> StatePacket.Events
```

The packet queue is intentionally named `pendingPresentationEvents` because it stores generated packet `EventState` values for client effects. It is not a domain event queue.

The seam should stay narrow. It is not an achievements system, stats system, API integration point, persistence layer, logging policy, pub/sub system, listener registry, async dispatcher, match history, or replacement home for gameplay rules. Future systems may consume `events.Event` or richer domain facts, but the event seam itself should only define, record, drain, and translate events needed by existing behavior.

### Toroidal World Wrap

The server stores bounded wrapped world coordinates using `constants.WorldWidth` and `constants.WorldHeight`. `Game.Step()` chooses the world bounds with `space.DefaultBounds()` and passes them to `motion.AdvanceShip`, `motion.AdvanceAsteroid`, and `motion.AdvanceBullet`. Those advance helpers step one entity and wrap its position with `space.WrapPosition`.

Spatial rules flow through `services/game-server/internal/game/space`:

- spawning aim uses wrapped direction
- visibility/despawn uses wrapped delta
- respawn safety uses wrapped distance
- ship/asteroid and projectile/asteroid collision helpers place temporary asteroid bodies in wrapped-local space before collision checks

The client renders continuous visual coordinates. `local_visual_sync.gd` tracks the local server position and continuous local visual position; `player_sync.gd`, `asteroid_sync.gd`, and `bullet_sync.gd` render entities relative to the local player with shortest wrapped deltas. `world_sync.gd` coordinates the update order and exposes visual-position conversion for server-driven effects. The camera and background follow the local player node, so they inherit the continuous visual position.

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

Server packet wire serialization goes through `services/game-server/internal/protocol/packetcodec`. The seam is intentionally JSON-only for now and exposes generic `Encode(packet any)` and `Decode(data []byte, packet any)` helpers. It must not import `internal/game`; generated packet structs stay in their current packages. Networking uses it for websocket client packet decode, state packet encode, and room snapshot/error encode. Collision-shape JSON loading and tests that inspect generated JSON tags are not packet wire serialization.

Client packet wire serialization goes through `client/scripts/networking/packet_codec/packet_codec.gd`. It is also intentionally JSON-only and only wraps `JSON.stringify`/`JSON.parse_string`; `network_client.gd` remains the websocket owner for polling, signals, and `send_text`. Generated GDScript packet builders remain in `client/scripts/networking/packets.gd`, and the codec should not grow packet validation, typed packet objects, protobuf references, or format switching without an explicit migration.

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
2. `game.gd` sends input/client-config/respawn packets through `network_client.gd`; outbound packet dictionaries are JSON-encoded through the client packet codec.
3. The Go websocket read path decodes client packet JSON through the server `packetcodec` and passes packets to room/game handlers.
4. The game simulation applies input and advances authoritative state.
5. The server encodes `StatePacket` JSON through `packetcodec` and writes it back to the client.
6. `network_client.gd` decodes inbound websocket text through the client packet codec, then `game.gd` receives the packet, stores packet-level player lifecycle status, and passes renderable state to `world_sync.gd`.
7. `world_sync.gd` delegates rendered node creation, removal, packet application, and interpolation to `PlayerSync`, `BulletSync`, `AsteroidSync`, and `LocalVisualSync`.
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
- per-player match lifecycle status through `StatePacket.player_lifecycle`
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
- Do not infer player lifecycle from `StatePacket.players` or client-side ship presence. Use `StatePacket.player_lifecycle`; pending-respawn and eliminated players can be absent from active ship state.
- Keep network transport separate from core game simulation. Websocket code should live in `services/game-server/internal/networking`; reusable simulation should live in `services/game-server/internal/game`; match/mode policy evaluation should live in `services/game-server/internal/game/rules`.
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
