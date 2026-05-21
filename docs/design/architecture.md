# Architecture

Space Rocks is an Asteroids-inspired game with a Godot client and a Go server. The current direction is server-authoritative for gameplay state where networking is involved: player input is sent to the server, the server advances simulation, and clients render the state they receive.

The project is still in development, so this document describes the architecture that exists now and calls out future notes separately.

## Repository Layout

- `client/`: Godot project. Contains scenes, scripts, assets, audio, shaders, and client-side tools.
- `services/game-server/`: Go module for the real-time game server. The current entrypoint is `services/game-server/cmd/game-server`.
- `services/api-server/`: empty placeholder for a planned Node.js/TypeScript NestJS API server for business/backend systems. It is intentionally separate from real-time simulation.
- `shared/`: JSON source data shared across client and server generation, including constants, packet definitions, and collision shape data.
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
- `client/scripts/networking/world_sync.gd`: applies server state to local/remote player, bullet, and asteroid nodes. It also interpolates rendered nodes toward server positions.
- `client/scripts/entities/player.gd`: collects input into packet data, plays local laser audio, and toggles local afterburner visuals.
- `client/scripts/effects.gd`: spawns local visual/audio effects for bullet impacts, ship death, and game over sound timing.
- `client/scripts/ui/hud_controller.gd`: updates score, lives, room ID, death overlay, respawn state, and game-over UI.
- `client/scripts/networking/packets.gd` and `client/scripts/constants/constants.gd`: generated/shared client packet helpers and constants.

Rendering is scene/node based in Godot. The client renders the ship, asteroids, bullets, background, UI, animations, and audio. The background has local auto-scroll in `game_shell.gd`; gameplay scroll offset follows the local player after initial spawn.

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

- `services/game-server/internal/networking`: websocket handler and room manager.
- `services/game-server/internal/game`: game loop, state packets, combat, spawning, scoring, respawn/session logic, visibility.
- `services/game-server/internal/game/entities`: game entities and generated packet state structs.
- `services/game-server/internal/game/physics`: collision shapes, collision detection, vectors, and shared collision shape loading.
- `services/game-server/internal/game/space`: gameplay spatial helpers for distance, direction, and position normalization. Current behavior is flat/infinite; this package is the intended seam for future wrapped-world support.
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

`Game.Start()` launches a simulation loop at `constants.ServerTickRate`. Each tick applies player input, moves entities, handles cooldowns, spawns asteroids, removes expired/far objects, and resolves collisions.

The server currently owns:

- player movement simulation from input
- bullet spawning
- asteroid spawning and visibility removal
- bullet/asteroid collision
- ship/asteroid collision
- asteroid splitting
- scoring
- lives, death, game-over, and respawn rules
- safe initial spawn/respawn placement
- state packet generation

### Rooms And Networking

`services/game-server/internal/networking/rooms.go` manages rooms. Each room owns its own `*game.Game`. Blank room IDs map to the default room. Non-blank room IDs create or join separate rooms. Empty rooms are cleaned up after a grace period.

`services/game-server/internal/networking/websocket.go` upgrades `/ws` connections. It accepts an optional query parameter:

```text
/ws?room_id=abc
```

On connect:

- the connection joins a room
- the room's game adds a player
- one goroutine reads client input packets
- the write loop sends server state packets at the server tick rate

On disconnect:

- the player is removed from the room's game
- the room leave function runs
- empty rooms schedule cleanup

### Physics

The physics package provides collision primitives and collision detection for circles, capsules, rectangles, and polygons. Collision shapes are loaded from:

```text
shared/collisions/collision_shapes.json
```

The server uses imported collision shapes for ship, bullet, and asteroid collision bodies.

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

Server-owned constants live under `constants.server.*` and may be omitted from client generated constants. In particular, `player_starting_lives` and `player_respawn_delay` live under `constants.server.player_lifecycle`, while `asteroid_size_scale` lives under `constants.server.asteroids`. The client receives lives through player state, respawn delay through death events, and asteroid scale through asteroid state instead of importing those constants.

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
- Use `shared/` JSON plus generation scripts for packet and constant data that must stay aligned across Go and Godot.
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
- Invisible toroidal/wrapped playfield is planned as a future option. See [toroidal wrap plan](toroidal-wrap.md).
- A thin server-side ship variant foundation exists: runtime ship type, resolved ship stats/modifiers, `ship_type` snapshots, and collision shape ID lookup. Full variants with client scene mapping and keyed collision catalogs remain future work. See [ship variants plan](ship-variants.md).
