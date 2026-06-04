# Architecture

Space Rocks is an Asteroids-inspired game with a Godot client and a Go server. The current direction is server-authoritative for gameplay state where networking is involved: player input is sent to the server, the server advances simulation, and clients render the state they receive.

The project is still in development, so this document describes the architecture that exists now and calls out future notes separately.

## Repository Layout

- `client/`: Godot project. Contains scenes, scripts, assets, audio, shaders, and client-side tools.
- `services/game-server/`: Go module for the real-time game server. The current entrypoint is `services/game-server/cmd/game-server`.
- `services/api-server/`: empty placeholder for a planned Node.js/TypeScript NestJS API server for business/backend systems. It is intentionally separate from real-time simulation.
- `shared/`: source data shared across client and server generation, including TOML constants, TOML packet definitions, and JSON collision shape data.
- `docs/`: Project documentation.
- `tools/data_sync/`: Python sync/generation tool used to generate constants and packet code from `shared/`.

## Client Architecture

The Godot client is responsible for presentation, local input collection, UI, audio/effects, interpolation, and websocket communication with the Go server. It does not own authoritative gameplay outcomes like scoring, lives, asteroid collision results, match lifecycle, or respawn validity.

The configured Godot main scene is:

```text
client/scenes/game.tscn
```

Current client runtime seams:

- `client/scripts/session/`: session-level coordinators, including gameplay, room, config, and client session context.
- `client/scripts/shell/gameplay_shell_flow.gd`: narrow gameplay lifecycle/composition shell. It should stay mostly as orchestration and delegation; gameplay state application, per-frame processing, and server hitbox overlay ticking are delegated to focused flows.
- `client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd`: server hitbox overlay lookup, reset, and overlay ticking seam for gameplay debug presentation.
- `client/scripts/gameplay/state/gameplay_state_apply_flow.gd`: gameplay state application seam for packet reading and normalized state application order.
- `client/scripts/gameplay/runtime/gameplay_process_flow.gd`: per-frame gameplay processing seam for the runtime/input/spectate order.
- `client/scripts/gameplay/gameplay_composition.gd`: gameplay flow construction and fanout only. `GameplaySessionController` keeps packet gating and outer lifecycle consequences. `GameplayComposition` should not own packet parsing, connection shutdown, session clearing, menu show/hide, or gameplay rules.
- `client/scripts/gameplay/runtime/`: gameplay runtime composition/delegation context. `GameplayRuntimeContext` stays focused on wiring focused runtime seams rather than acting as a read-model passthrough bucket.
- `client/scripts/gameplay/state/`: gameplay packet/state readers and normalized state helpers.
- `client/scripts/gameplay/input/`: local gameplay input polling/routing, including movement, pause/menu, respawn, spectate input routes, and devtools input ownership.
- HUD/UI mouse input gating is owned by `client/scripts/gameplay/input/hud_input_policy.gd`, registered as the `HudInputPolicy` autoload. `GameplaySessionController` keeps the top-level input priority order and delegates the HUD/UI hover gate to `HudInputPolicy`.
- `client/scripts/gameplay/hud/`: gameplay HUD flow and runtime HUD ticking.
- `client/scripts/gameplay/background/`: gameplay background/parallax shader scroll presentation.
- `client/scripts/devtools/context/`: devtools coordination contexts for state caching, command delegation, placement routing, overlay coordination, gameplay-state fanout, window signal wiring, and DevToggle routing.
- `client/scripts/devtools/telemetry/`: devtools telemetry seam for debug-only world metrics, overlay flow, RTT tracking, and packet-age display plumbing.
- `client/scripts/devtools/dev_tools_session_flow.gd`: devtools gameplay session seam for runtime wiring. `GameplaySessionController` delegates devtools input, per-frame processing, and placement routing to this flow.
- Server hitbox rendering is owned by client devtools. The overlay scene lives under `client/scenes/devtools/`, the drawing/template code lives under `client/scripts/devtools/hitboxes/`, and `WorldSync` exposes read-only draw-entry data only. `GameplayRuntimeContext` does not own or expose that draw-entry data. Normal gameplay entities do not draw their own debug collision outlines.
- `client/scripts/gameplay/menu/`: gameplay menu flow and semantic menu lifecycle signal routing.
- `client/scripts/gameplay/respawn/`: respawn request and confirmation state.
- `client/scripts/gameplay/spectate/`: spectate state, menu requests, and view target selection/cycling; it does not own remote camera nodes.
- `client/scripts/gameplay/spectate/spectate_session_flow.gd`: spectate session wiring. `SpectateSessionFlow` owns `SpectateMenuState` creation, menu/shell configuration, gameplay-state application, and reset delegation.
- `client/scripts/gameplay/presentation/`: client-side presentation policy, including player hue application and OS indicator hue matching.
- `client/scripts/gameplay/events/`: server event lane and death/game-over consequences.
- `client/scripts/gameplay/effects/`: gameplay effects helper used by event/effects flows.
- `client/scripts/lobby/`: lobby shell/presenter/network action flows.
- `client/scripts/boot/`: boot flow and pending boot request.
- `client/scripts/config/`: client config flows.
- `client/scripts/networking/network_client.gd`: websocket transport owner for connect, poll, raw send, raw receive, graceful close, and packet codec use.
- `client/scripts/networking/inbound/`: server packet classification and dispatch.
- `client/scripts/networking/outbound/`: client packet send helpers grouped by packet family.
- `client/scripts/networking/client_connection_service.gd`: public connection facade and signal bridge; it no longer owns packet-family construction or packet-family routing.
- `client/scripts/networking/packets/packet_codec.gd`: client packet wire encode/decode wrapper around JSON parsing and `JSON.stringify`. It owns wire parsing and envelope validation only; packet-specific readers validate payload details.
- `client/scripts/world/world_sync.gd`: coordinator for server-state rendering. It delegates player/render-origin work to `client/scripts/world/player_render/player_render_api.gd` and delegates bullet/asteroid node ownership, packet application, cleanup, and interpolation to the focused sync owners. For targeting, it only exposes `target_source()`; target selection orchestration lives above it in `GameplayTargetingContext`.
- `client/scripts/entities/player.gd`: local player node and packet-facing movement/shoot input state.
- `client/scripts/ui/`: UI nodes/controllers.
- `client/scripts/generated/networking/packets/packets.gd` and `client/scripts/generated/constants/constants.gd`: generated/shared client packet helpers and constants.

`GameplayDevtoolsContext` is a composition facade for devtools coordination. It constructs the devtools contexts, delegates reset/process/public wrapper methods, and keeps the outer API stable while the owned responsibilities live under `client/scripts/devtools/context/`. It is not owned by gameplay HUD flows and not owned by `gameplay_shell_flow.gd`.

Client runtime flow:

1. Session controllers under `client/scripts/session/` coordinate boot, config, room/lobby, and gameplay session flows.
2. `gameplay_shell_flow.gd` stores references and delegates to focused gameplay lanes rather than owning input, HUD, menu, respawn, spectate, events, effects, and state application directly.
3. Local gameplay input routes through `client/scripts/gameplay/input/`. Player movement/shooting packet data still comes from `client/scripts/entities/player.gd`.
4. Input/event handling flows `InputEvent` -> `GameplayInputContext` -> `MouseActionFlow` -> `GameplayTargetingContext` -> candidate source / picker / packet send for target selection.
5. `network_client.gd` sends and receives websocket text through the client packet codec.
6. Incoming gameplay state is normalized by `client/scripts/gameplay/state/` and applied by `client/scripts/gameplay/runtime/`.
7. `client/scripts/world/world_sync.gd` updates renderable player/render-origin state through `PlayerRenderApi` and updates bullet and asteroid state through their focused sync owners.
8. HUD, menu, respawn, spectate, event, death, and effects presentation updates flow through the focused gameplay seams under `client/scripts/gameplay/`.

Rendering is scene/node based in Godot. The client renders the ship, asteroids, bullets, background, UI, animations, and audio. Normal gameplay follows the active ViewAnchor render origin after initial spawn. Spectate keeps viewport/camera ownership local/client-owned, uses the selected active player as the current view reference only through the ViewAnchor seam, and requires the background/parallax to sample the same view reference as the camera.

### ViewAnchor Render Origin

ViewAnchor is the single render origin for gameplay world presentation.

`Camera2D` lives under `ViewAnchor`.

Background follows `ViewAnchor`.

Player is not the camera carrier.

Player identity is not automatically the render origin.

### Legacy PlayerRender Quarantine

`client/legacy/player_render` contains quarantined legacy implementation details.

It must be treated as a black box.

New code must use `client/scripts/world/player_render`.

Current limitations:

- The client expects a Go server at `ws://localhost:8080/ws`.
- There is no implemented client-side prediction beyond interpolation/render smoothing.
- Local server launch from the Godot client is not implemented.

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

- `services/game-server/internal/networking`: websocket upgrade, sessions, read/write loops, transport logging, and adapter wiring.
- `services/game-server/internal/rooms`: room state, room membership ownership, lifecycle ownership, and cleanup policy.
- `services/game-server/internal/game`: game loop, state packets, combat, spawning, scoring, respawn/session logic, visibility.
- `services/game-server/internal/game/motion`: per-entity movement integration and advance-with-wrap helpers for ships, asteroids, and bullets.
- `services/game-server/internal/game/rules`: match/mode policy evaluation from plain snapshots. It currently owns game-over outcome evaluation and per-player participation classification.
- `services/game-server/internal/game/scoring`: pure score policy evaluation. It converts scoring events into awards without mutating game sessions.
- `services/game-server/internal/game/entities`: game entities and generated packet state structs.
- `services/game-server/internal/game/physics`: collision shapes, collision detection, vectors, and shared collision shape loading.
- `services/game-server/internal/game/space`: gameplay spatial helpers for wrapped distance, direction, shortest delta, and position normalization.
- `services/game-server/internal/constants`: generated Go constants from split constants SoT files under `shared/constants/`.
- `services/game-server/internal/logging`: structured `slog` wrapper with categories and environment-controlled levels.
- `services/game-server/internal/protocol/packetcodec`: JSON-only packet wire encode/decode helpers for server packets.

State packet fields may pass through adapter or wrapper paths before reaching the client. Those paths should intentionally copy any new `StatePacket` fields, such as `total_asteroids`, instead of assuming the base packet shape passes through unchanged.

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

`Game.Step()` is a same-package simulation coordinator in `services/game-server/internal/game/simulation.go`. It preserves authoritative phase order while routing player/session, asteroid, bullet, and collision phases through focused same-package helpers.

Per-entity movement integration and wrapping live in `services/game-server/internal/game/motion`. The motion package imports entity types and spatial helpers, but it does not import `internal/game`, mutate `game.state` maps, spawn entities, delete entities, award score, resolve collisions, or write packets.

### Spawn Planning

`services/game-server/internal/game/spawn_types.go` defines player spawn vocabulary currently used by the game package. `services/game-server/internal/game/spawning` owns asteroid/projectile spawning policy, including asteroid plans, bullet construction, and ID allocation.

Player initial spawn and respawn planning use `PlayerSpawnPlan`. Player lifecycle still owns session lookup, `CanRespawn()` gating, lives, death, respawn cooldowns, ship creation, and camera view attachment. Match-over policy is evaluated through `services/game-server/internal/game/rules`.

The spawn seam is still partial. Bullet construction lives in `spawning.Spawner`, but `spawnBullet()` remains the `Game` adapter that inserts the projectile into `game.state.Projectiles`.

Continuous bullet stream runtime state is owned by `services/game-server/internal/devtools/streamruntime`. `internal/game` does not import `internal/devtools` or own devtools runtime state. Game-owned debug operations are exposed through narrow `export_devtools_*.go` hooks, including debug bullet spawning and generic simulation step observer registration.

### Match Rules

`services/game-server/internal/game/rules` owns match/mode policy decisions from plain facts. It does not import `internal/game` or `internal/rooms`, does not mutate `Game`, and does not inspect game storage directly.

`Game.MatchDecision()` is the public game-facing API for richer match decisions. `Game.IsGameOver()` remains available for existing callers and delegates through the same locked decision path.

`Game.statePacket()` projects `MatchDecision.Players` into `StatePacket.player_lifecycle`, a map from player ID to lifecycle status string. `StatePacket.players` (from `game.state.Players`) is active ship/render state only. Durable player identity/status/position/readout state is carried in `StatePacket.player_world_states` (from `player.WorldState`). Pending-respawn players may be absent from `StatePacket.players` while still present in `StatePacket.player_world_states`; in that lifecycle state they are not targetable, damageable, or collidable.

### Entity Damage Resolution

`services/game-server/internal/game/collisions.go` defines a narrow collision detection seam for the destructive collision pairs currently used by combat. It reports concrete projectile/asteroid and player/asteroid collision facts.

`services/game-server/internal/game/damage/` defines the internal damage resolution seam for authoritative entity damage and destruction decisions. Combat/game code is the adapter: it consumes collision facts, builds a `DamageRequest`, calls the damage package resolver, and then existing systems consume the `DamageResult`.

The damage resolver only answers what happened to the target from the damage request. It does not mutate lives, respawn players, award score, spawn fragments, emit events, log, or write packets. Those lifecycle effects remain with combat/session/scoring/spawning and the domain event seam.

### Scoring Policy

`services/game-server/internal/game/scoring` owns pure score policy. It receives scoring facts as `scoring.Event` values and returns `scoring.Award` values. It does not import `internal/game`, inspect game storage, mutate players or sessions, check pause/invulnerability state, log, emit packets, or write events.

Score/lives mutation is centralized in game-owned player counter seams under `services/game-server/internal/game`. The scoring policy boundary remains pure: policy computes awards only, and game-owned adapters apply mutations.

`game.awardScore` remains the game-owned application seam for scoring awards. It applies score changes through the score counter seam and keeps missing-player, paused-player, invulnerable-player, session sync, and score logging behavior outside the scoring policy package.

Lives and death paths mutate authoritative lives through the lives counter seam. The counter seam keeps persistent player session values and active ship values synchronized.

Future gameplay and devtools adapters should use the same counter mutation seam instead of directly changing session or active-player fields.

### Domain Gameplay Events

`services/game-server/internal/game/events` defines the small domain event vocabulary for gameplay facts that already produce client-visible packet events. The first supported facts are bullet blasts and ship deaths.

Gameplay systems such as combat still own gameplay decisions: collision outcomes, scoring, lives, death, respawn, and spawning rules remain in their existing files. Current producers record `events.Event` values. Root `services/game-server/internal/game/events.go` remains the package-local presentation adapter that translates domain events to the generated packet-facing `EventState` output where needed.

The packet queue is intentionally named `pendingPresentationEvents` because it stores generated packet `EventState` values for client effects. It is not a domain event queue.

### Toroidal World Wrap

The server stores bounded wrapped world coordinates using `constants.WorldWidth` and `constants.WorldHeight`. `Game.Step()` chooses the world bounds with `space.DefaultBounds()` and passes them to `motion.AdvanceShip`, `motion.AdvanceAsteroid`, and `motion.AdvanceBullet`. Those advance helpers step one entity and wrap its position with `space.WrapPosition`.

Spatial rules flow through `services/game-server/internal/game/space`:

- spawning aim uses wrapped direction
- visibility/despawn uses wrapped delta
- respawn safety uses wrapped distance
- ship/asteroid and projectile/asteroid collision helpers place temporary asteroid bodies in wrapped-local space before collision checks

The client renders continuous visual coordinates. `local_visual_sync.gd` tracks the local server position and continuous local visual position; `player_sync.gd`, `asteroid_sync.gd`, and `bullet_sync.gd` render entities relative to the active anchor position with shortest wrapped deltas. `world_sync.gd` coordinates the update order and delegates player/render-origin work to `client/scripts/world/player_render/player_render_api.gd`. `player_render_api.gd` coordinates player meaning and ViewAnchor/render-anchor mapping. `bullet_sync` and `asteroid_sync` receive the active anchor visual/server positions from `world_sync`. The camera and background follow `ViewAnchor`, not the local player node.

Targeting ownership sits outside `WorldSync`:

- `MouseActionFlow` stays the lowest-level mouse/input action coordinator.
- `GameplayTargetingContext` owns target selection orchestration.
- `GameplayTargetCandidateFlow` builds target candidates.
- `TargetPositionSource` owns targetable position read models.
- `WorldSync` only exposes `target_source()` for targeting access.

See [toroidal wrap](toroidal-wrap.md).

### Rooms And Networking

Room/domain ownership lives in `services/game-server/internal/rooms`. That package owns room state, room membership ownership, room-code/default-room helpers, create/join/leave/ready/start-game/single-player/return-to-lobby/game-over/cleanup decisions, and each room's `*game.Game` lifecycle while simulation rules stay in `internal/game`.

Concrete room ownership is split inside the package:

- `roomMembership` owns members and owner selection.
- `roomMatch` owns the room game instance and active-player count.
- `roomCleanup` owns the cleanup timer and cleanup version.
- `Room` remains the aggregate root for room ID, top-level state, joinability, locking, and coordination.

`services/game-server/internal/networking` owns websocket upgrade, sessions, read/write loops, transport logging, and adapter wiring. It upgrades `/ws`, reads generated packets, calls room-domain methods, attaches or clears websocket session player IDs, and sends or broadcasts generated packets such as `RoomSnapshot` and `RoomError`.

`services/game-server/internal/networking/inbound` owns pure inbound packet family handlers.

`services/game-server/internal/networking/outbound` owns pure outbound gameplay presentation/write helpers.

Server packet wire serialization goes through `services/game-server/internal/protocol/packetcodec`. It remains JSON-only for now. Client packet wire serialization goes through `client/scripts/networking/packets/packet_codec.gd`. Callers consume `PacketEncodeResult` and `PacketDecodeResult` so `NetworkClient` does not depend on raw JSON return types.

Diagnostic telemetry packets are part of networking transport behavior:

- `telemetry_ping` and `telemetry_pong` are diagnostic packets.
- The client sends `telemetry_ping` only while the world telemetry overlay is visible.
- The server replies with `telemetry_pong` only to the same websocket session.
- Ping/pong handling does not require room membership or active gameplay state.
- Ping/pong does not mutate gameplay state.

The websocket connection itself is session-only. Room membership ownership lives in `services/game-server/internal/rooms`; networking routes the membership packets:

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

#### Devtools Command Handling Boundary

Devtools command handling is a networking-routed server boundary, not a normal gameplay packet path.

- inbound networking routing detects devtools command types and forwards them through the inbound seam
- devtools command packets route to `services/game-server/internal/devtools.HandleCommand`
- devtools commands do not route through `Game.HandlePacket`
- `nodevtools` builds ignore/reject devtools command handling through the existing devtools gate

Mutation ownership for devtools remains explicit:

- `services/game-server/internal/devtools` owns devtools command handlers
- `services/game-server/internal/game/export_devtools_*.go` exposes narrow game-owned adapters used by devtools
- score/lives devtools adapters delegate to the shared player counter mutation seam
- clear bullets/asteroids mutate authoritative server state only; clients observe changes through normal state/world sync

#### Server Identity Policy

- `PlayerID` is the permanent player-facing identity.
- `PlayerID` uses readable values such as `Player-1`, `Player-2`, `Player-3`.
- `PlayerID` must not be converted to UUID.
- Player-facing room/lobby packet identity fields use `PlayerID`: `player_id`, `local_player_id`, and `owner_id`.
- `SessionID` is server-internal websocket/session identity and is a target for the upcoming UUID upgrade.
- `MemberID` is server-internal room-membership identity, currently UUID v4, reserved as the future disconnect/reconnect seam.
- `MemberID` should not be exposed in normal room snapshot packets.
- Server-internal identity values may migrate to UUID, but player-facing identity values must not be swept into that migration.
- `currentGamePlayerID` is networking-owned active game routing state only; it is not room membership identity and not player-facing identity.

### Physics

The physics package provides collision primitives and collision detection for circles, capsules, rectangles, and polygons. Collision shapes are loaded from:

```text
shared/collisions/collision_shapes.json
```

This JSON is generated by:

```text
client/tools/export_collision_shapes.gd
```

The exporter reads Godot collision nodes from:

- `client/scenes/bullet.tscn`
- `client/scenes/player.tscn`
- `client/scenes/asteroid.tscn`

The Go server uses the generated JSON for authoritative collision bodies. Do not hand-edit the JSON unless intentionally patching generated output.

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

1. Godot input is collected by gameplay input seams and `player.gd`.
2. Outbound packets route through `network_client.gd` and the client packet codec.
3. The Go websocket read path decodes client packet JSON through the server `packetcodec` and passes packets to room/game handlers.
4. The game simulation applies input and advances authoritative state.
5. The server encodes `StatePacket` JSON through `packetcodec` and writes it back to the client.
6. `network_client.gd` decodes inbound websocket text through the client packet codec result types, not raw JSON values.
7. Gameplay shell/session code normalizes state through `client/scripts/gameplay/state/`, applies runtime state through `client/scripts/gameplay/runtime/`, and updates world sync plus presentation lanes.
8. Devtools telemetry observes normalized gameplay state for debug metrics and telemetry overlay plumbing after state-reader normalization.
9. `world_sync.gd` delegates rendered node creation, removal, packet application, and interpolation to `PlayerSync`, `BulletSync`, `AsteroidSync`, and `LocalVisualSync`.
10. HUD/menu/respawn/spectate/events/effects update from state and events through focused gameplay seams.

Shared packet structures are sourced from:

```text
shared/packets/outputs.toml
shared/packets/gameplay.toml
shared/packets/debug.toml
shared/packets/lobby.toml
```

Generated packet files include:

- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/entities/packets_generated.go`
- `services/game-server/internal/devtools/packets_generated.go`
- `client/scripts/generated/networking/packets/packets.gd`

Shared constants are sourced from:

```text
shared/constants/server_constants.toml
shared/constants/server_entities.toml
shared/constants/client/presentation.toml
shared/constants/client/shell.toml
shared/constants/client/lobby.toml
```

Generated constants include:

- `services/game-server/internal/constants/constants.go`
- `client/scripts/generated/constants/constants.gd`

Server-owned constants live under `constants.server.*` and may be omitted from client generated constants. Client constants use nested subcategory sections under `constants.client.presentation.*`, `constants.client.shell.*`, and `constants.client.lobby.*`. World size is intentionally generated to both Go and GDScript because client visual wrapping must use the same bounds as the server.

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

### Devtools / Debug Boundary

Devtools are client-triggered and server-authoritative. Client input requests debug actions through generated packets, while gameplay effects are applied by server-owned systems.

- packet schema lives in `shared/packets/debug.toml`
- output routing lives in `shared/packets/outputs.toml`
- generated server devtools packets live in `services/game-server/internal/devtools/packets_generated.go`
- `internal/networking` classifies incoming packet type first
- default/disabled builds ignore devtools command packets before game handling
- `internal/devtools` owns command handling and devtools status wrapping
- devtools state wrapping must preserve normal gameplay state fields, including `PlayerWorldStates`
- `internal/game/export_devtools*.go` exposes only controlled operations needed by devtools
- `internal/game` must not import `internal/devtools`
- generated game packets must not own devtools command packet constants or devtools-only command fields
- debug actions route through real gameplay seams, not parallel debug-only gameplay logic
- current actions include invincible, infinite lives, world freeze, player freeze, kill, spawn, and respawn

Targeting seam notes:

- canonical target is gameplay/server state
- canonical game target identity is separate from local player readout state
- a client selection request is not authoritative until reflected in authoritative state updates
- devtools canonical target readout prefers active players and falls back to `player_world_states` for player targets with no active ship
- target/readout resolution must not depend only on active players; inactive player identity may remain available through `player_world_states`
- inactive player identity/readout does not make the player active, clickable, collidable, damageable, or targetable
- targeting state is not automatically combat behavior

Telemetry seam notes:

- HUD remains player-facing
- devtools window owns raw `LocalPlayerTelemetry` and `TargetTelemetry` state inspection
- world telemetry overlay is a separate devtools seam for glanceable world/performance/network metrics

For key mappings and detailed behavior, see [devtool toggles](../devtools/toggles.md).

Telemetry timing ownership notes:

- `StatePacket.server_sent_msec` is stamped by `services/game-server/internal/networking/websocket_write.go` before encode/write. That file now only writes outbound/presentation state and no longer advances game-over lifecycle.
- `services/game-server/internal/devtools/WrapStatePacket` preserves `server_sent_msec` when devtools status wrapping is applied.
- devtools local player readout prefers active players and falls back to `player_world_states` when the local player has no active ship.
- server `player.WorldState` serializes with snake_case JSON field names for client packet compatibility.
- `packet_staleness_ms` is local monotonic age since the last received gameplay state packet.
- `packet_age_ms` is derived in client monotonic clock space using server clock offset estimated from telemetry pong, not raw client wall-clock minus server wall-clock.

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
- Use `shared/constants/server_constants.toml`, `shared/constants/server_entities.toml`, `shared/constants/client/presentation.toml`, `shared/constants/client/shell.toml`, `shared/constants/client/lobby.toml`, `shared/packets/outputs.toml`, `shared/packets/gameplay.toml`, `shared/packets/debug.toml`, `shared/packets/lobby.toml`, and `tools/data_sync/` for packet and constant data that must stay aligned across Go and Godot.
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

