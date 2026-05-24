# Project Notes

This file is project memory for future Codex sessions. It is not a full architecture document. For that, see [docs/design/architecture.md](design/architecture.md). For practical development commands and conventions, see [docs/developer.md](developer.md).

Always prefer cleaner, scalable design choices.

## Current Development Status

Space Rocks is playable in development form with:

- Godot client
- Go websocket game server
- main menu and game loop scene shell
- packet-based multiplayer room lifecycle over `/ws`
- server-authoritative player movement, bullets, asteroid collisions, scoring, lives, death, respawn, and safe spawn placement
- generated shared constants and packet helpers
- structured server logging
- HUD score/lives/room/death/game-over display
- audio/effects for shooting, asteroid impacts, ship death, and game over
- pause-state server plumbing and client toggle plumbing, but no real pause menu scene yet
- bounded toroidal world wrapping with continuous client visual coordinates

The project is still moving quickly. Treat recent systems as subject to refinement.

## Recently Implemented Systems

### Client GUT Test Suite

A lightweight Godot client test suite now lives under:

```text
client/tests/
```

Current layout:

- `client/tests/unit/` for focused GUT unit-style tests.
- `client/tests/fixtures/` for small world-state and scene/data fixtures.
- `client/tests/helpers/` for reusable test-only helpers.

The current suite covers smoke testing, generated packet builders and field constants, HUD lives/death/respawn behavior, `world_sync` create/update/remove behavior, asteroid packet scale handling, and missing-field safety for current client behavior.

Run GUT with:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

Expected warnings may appear for tests that intentionally verify safe missing-field handling. The run should still end with all tests passing.

Keep test-only fixtures/helpers out of `client/scripts/`. Full gameplay/network smoke testing is still manual for now: opening the game scene, websocket connection, spawning asteroids, shooting/effects, pause/debug flow, and the full gameplay loop.

There is also a Python static boundary scan for forbidden client references to server-owned constants:

```bash
python3 -m pytest tools/tests/test_client_constants_boundary.py
```

### Server Test Layout

Go server tests have been moved out of production package folders and now live under:

```text
services/game-server/tests/
```

Current subareas:

- `game`
- `networking`
- `physics`
- `rooms`
- `space`

Future server tests should stay under `services/game-server/tests/<area>/`, not beside production packages under `services/game-server/internal/`.

Game simulation tests use the shared harness in:

```text
services/game-server/tests/game/helpers_test.go
```

The harness exists to keep tests readable while still allowing precise server-authoritative setup for collisions, respawn, devtools, pause, scoring, spawning, and similar behavior. Keep new harness helpers intent-level and avoid exposing raw private maps directly to tests.

### Shared Constants

`shared/game_data.toml` is the active source of truth for generated Go and GDScript constants.

Generated outputs:

- `services/game-server/internal/constants/constants.go`
- `client/scripts/constants/constants.gd`

Constants are synced through:

```text
tools/data_sync/
```

Recent additions include:

- player starting lives
- scoring base score
- respawn buffer/delay
- game-over sound delay
- window min/max size constants
- player resume invulnerability duration

Note: `tools/data_sync/` updates only marked `data-sync` blocks. Do not use the old `tools/scripts/generate_constants.py` path for active constants changes.

Boundary note: server-owned constants live under `constants.server.*`. World size is generated to both Go and GDScript because client visual wrapping must match server bounds. `player_starting_lives` and `player_respawn_delay` live under `constants.server.player_lifecycle`; `asteroid_size_scale` lives under `constants.server.asteroids`. The client should receive lives through player/state data, respawn delay through death events, and asteroid scale through asteroid state instead of importing those constants.

### Toroidal World Wrap

World wrapping is implemented.

Server behavior:

- world bounds come from `constants.WorldWidth` and `constants.WorldHeight`
- `services/game-server/internal/game/space` owns wrapped delta, distance, direction, and normalization
- `Game.Step()` centrally wraps players, asteroids, and bullets after movement
- spawning, visibility/despawn, respawn safety, and collision checks use wrap-aware spatial helpers

Client behavior:

- `client/scripts/world_wrap.gd` uses generated `Constants.WORLD_WIDTH` and `Constants.WORLD_HEIGHT`
- `client/scripts/networking/world_sync.gd` tracks `local_server_position` and continuous `local_visual_position`
- remote players, asteroids, bullets, and server-driven effects render relative to the local player's visual position
- camera/background follow the continuous local player node

Important maintenance note: after changing world size in `shared/game_data.toml`, run:

```bash
python3 tools/data_sync/main.py -push -constants -go -gds
```

Running only `-go` can leave the client using stale wrap bounds.

### Shared Packets

`shared/packets/packets.toml` is the active source of truth for packet constants, Go structs, and GDScript packet builders.

Generated outputs:

- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/entities/packets_generated.go`
- `client/scripts/networking/packets.gd`

The packet TOML schema preserves the old rich JSON behavior:

- `outputs`: generated file targets, Go package mapping, imports, base class, and selected structs/builders
- `structs`: packet/state structs and field metadata
- `packet_types`: generated packet type constants
- `builders`: GDScript packet builder functions and `$arg` references
- arrays/maps/custom struct refs and overrides such as `go_type`, `go_item_type`, and `go_value_type`
- rich type strings such as `map<string,ShipState>` and `array<EventState>`

`shared/game_data.toml` contains constants only. Obsolete packet reference sections were removed when the packet TOML pipeline was adopted; packet schema edits belong in `shared/packets/packets.toml`.

Recent packet additions:

- `pause_player`
- `resume_player`

### Data Sync Pipeline

A TOML-based sync tool is active for constants and packets:

```text
tools/data_sync/
```

The active sources of truth are:

```text
shared/game_data.toml
shared/packets/packets.toml
```

The tool supports `-push`, `-pull`, `-diff`, `-check`, and `-validate`. The active paths are `-constants -go -gds` and `-packets -go -gds`. TypeScript output is future/deferred.

Output filtering is controlled by `tools/data_sync/config.toml`:

- `sections`: TOML sections a language receives during push/diff/check.
- `owns`: TOML sections a language may update during pull.

Constants pull is strict and updates existing owned values only. Full packet schema pull is intentionally unsupported; packet schema should be edited directly in `shared/packets/packets.toml`.

Current constants workflow:

1. Edit `shared/game_data.toml`.
2. Validate with `python3 tools/data_sync/main.py -validate -constants`.
3. Preview with `python3 tools/data_sync/main.py -diff -constants -go -gds`.
4. Apply with `python3 tools/data_sync/main.py -push -constants -go -gds`.
5. Check with `python3 tools/data_sync/main.py -check -constants -go -gds`.

Current packet workflow:

1. Edit `shared/packets/packets.toml`.
2. Validate with `python3 tools/data_sync/main.py -validate -packets`.
3. Preview with `python3 tools/data_sync/main.py -diff -packets -go -gds`.
4. Review the diff.
5. Apply with `python3 tools/data_sync/main.py -push -packets -go -gds`.
6. Check with `python3 tools/data_sync/main.py -check -packets -go -gds`.

### Multiplayer Lifecycle And Server Rooms

Multiplayer lifecycle v1 is now packet-based:

```text
Main Menu -> Multiplayer Dialog -> /ws session -> Create/Join packets -> Lobby -> Ready -> Start -> InGame -> GameOver -> Lobby or Leave
```

Important lifecycle boundaries:

- websocket connection is only a session; it does not imply room membership
- room membership is separate from active game players
- ships/game players are created only when `StartGameRequest` succeeds
- `/ws?room_id=...` no longer creates or joins rooms
- `CreateRoomRequest` creates private code rooms
- `JoinRoomRequest` only joins existing Lobby rooms
- all connected members must be ready before start; one ready player can start alone
- `LeaveRoomRequest` and disconnect remove the member and schedule cleanup when rooms become empty
- `ReturnToLobbyRequest` resets GameOver rooms back to Lobby and clears ready/game state
- legacy direct-room compatibility is quarantined: `DefaultRoom()` has been removed, while `GetOrCreate()` and `Join()` remain only for already-started direct game rooms and should not be used by new websocket lifecycle code

Server package ownership:

- `services/game-server/internal/rooms` owns `Room`, `RoomManager`, `RoomMember`, room states, room error constants, room capacity, room code/default-room helpers, room `*game.Game` ownership, and room lifecycle decisions for create/join/leave/ready/start, single-player startup, return-to-lobby, game-over transition, and cleanup.
- `services/game-server/internal/networking` owns websocket/session/outbound queue transport, packet handlers, per-connection player activation/deactivation, and sending/broadcasting `RoomSnapshot`/`RoomError`.
- `services/game-server/internal/game` owns simulation/gameplay rules.

Client lifecycle notes:

- `client/scripts/ui/game_shell.gd` owns explicit session mode and lobby/gameplay transitions.
- `client/scripts/networking/network_client.gd` owns generated packet send helpers.
- `client/scripts/game.gd` can receive an injected `NetworkClient` for multiplayer gameplay and returns it to the shell when a GameOver room returns to Lobby.
- Gameplay packets are ignored/deferred unless the multiplayer room state is `InGame`; single-player and legacy empty-room-state behavior still work.

### Pause / Resume Plumbing

Pause support currently exists in code, but the real UI menu is deferred.

Current implemented behavior:

- client can send `pause_player` / `resume_player`
- server stores per-player `Paused` and `InvulnerabilityRemaining`
- paused players ignore input, cannot move, shoot, score, or take asteroid damage
- resumed players receive a short invulnerability window
- invulnerable players cannot shoot/score during the window
- client `OpenMenu` toggles pause during active gameplay and stops local gameplay input while paused

Important recent context: there was a false debugging path where input looked broken because the server had not been started before testing. Pause support was re-enabled afterward. If pause seems broken again, first confirm the Go server is running and the Godot client is connected.

### Respawn And Safe Spawn

Respawn logic is server controlled. Players start with shared constant lives. Death reduces lives. Respawn is delayed. Safe spawn placement checks asteroids and existing players using a respawn buffer.

Initial spawning has been adjusted to reuse safe spawn logic while staying a separate concern so future initial-spawn-specific rules can be added.

### Spawn Planning Seam

Spawning now has a partial ownership seam in `services/game-server/internal/game`:

- `spawn_types.go` defines generic spawn vocabulary plus `AsteroidSpawnPlan` and `PlayerSpawnPlan`.
- `spawning.go` keeps timed asteroid orchestration in `spawnAsteroid()` and fragment orchestration in `spawnAsteroidFragments()`.
- `planTimedAsteroidSpawn()` chooses timed asteroid spawn facts.
- `planAsteroidFragmentSpawns()` chooses fragment spawn facts.
- `applyAsteroidSpawn()` allocates asteroid IDs and mutates `game.state.Asteroids`.
- `planInitialPlayerSpawn()` chooses the player initial spawn position while preserving `playerIndex` behavior and safe-spawn fallback.
- `planPlayerRespawn()` chooses the player respawn position from the already-gated `*playerSession`.

Current behavior is intended to remain unchanged. `Game.Step()` still owns timed asteroid scheduling, `spawnAsteroidBatch()` still owns batch count, combat still decides when fragments are needed, and player lifecycle still owns session lookup, lives, death, respawn cooldowns, ship creation, camera attachment, and game-over/session state. Bullet spawning is still separate projectile logic.

### Entity Damage Resolution

Server combat now routes current destructive collision outcomes through a small internal damage seam in `services/game-server/internal/game/damage.go`.

Current behavior is unchanged: bullet/asteroid hits still destroy and fragment asteroids, award score, despawn bullets, and record bullet blast events; ship/asteroid hits still kill the player, decrement lives, set respawn cooldown, and record ship death events. The resolver is intentionally side-effect-free: it returns `DamageResult` facts such as `Destroyed` or `Fatal`, while combat/session/scoring/spawning continue to own lifecycle effects.

The seam is general entity damage, not player-only health. Requests carry target/source IDs and entity types for players, asteroids, and projectiles, plus future no-op shape for shield and invulnerability bypass concepts. No client packets, packet schemas, health storage, shield UI, or balance rules were added in this slice.

### Collision Detection Seam

Server combat now routes current bullet/asteroid and player/asteroid pair overlap checks through `services/game-server/internal/game/collisions.go`.

The seam returns concrete facts: `BulletAsteroidCollision` and `PlayerAsteroidCollision`. It preserves current event positions by using the source entity position: bullet position for bullet blasts, player position for ship deaths. It intentionally does not use `physics.Collision.ContactPoint` for those event positions.

The helpers are unexported and side-effect-free. They do not despawn entities, build or resolve damage requests, award score, spawn fragments, decrement lives, set respawn cooldowns, record events, or touch packets/client code. Combat still owns the deferred ordering for scoring, despawn, fragments, death/session updates, game-over logging, and domain event recording.

### Scoring

Scoring is server controlled and tied to player instances. Asteroid hit score is:

```text
BASE_SCORE / asteroid size
```

The scoring code is intentionally modular enough to add future enemies or item pickups.

### Background / Game Shell

The root `game.tscn` owns the always-visible parallax background. `game_shell.gd` controls:

- background auto-scroll
- menu/game-loop scene switching
- minimum/maximum window size calls
- gameplay scroll offset from player position

`game_loop.tscn` and `main_menu.tscn` should not own their own background references.

### Logging

The server has a custom structured logging wrapper in `services/game-server/internal/logging`.

Categories:

- server
- network
- rooms
- game

Default is warn-level. Category overrides exist. See [docs/server/logging.md](server/logging.md).

The Godot client has a lightweight logger in `client/scripts/logging/logger.gd`.

Client categories include:

- shell
- lobby
- network
- game
- world_sync
- hud
- input
- packets

Use `ClientLogger` for new client lifecycle/network/UI diagnostics instead of adding raw `print()` calls. See [docs/client/logging.md](client/logging.md).

## Important Design Decisions

- The server is authoritative for game rules.
- The client is responsible for presentation, UI, audio/effects, interpolation, and input collection.
- Network transport belongs in `services/game-server/internal/networking`, not `cmd/game-server/main.go`.
- Business/API concerns belong in the planned `services/api-server/` service, not in the Go game server.
- Room state owns a separate `*game.Game` per room.
- Shared constants/packets should be generated, not copied by hand. Constants use `shared/game_data.toml` and `tools/data_sync/`; packets use `shared/packets/packets.toml` and `tools/data_sync/`. Output filtering may keep server-owned constants out of client generated files even when they remain in the constants source of truth.
- Collision shapes are shared through JSON and used by the server physics package. Current combat pair checks route through the server collision detection seam in `internal/game/collisions.go`.
- Score, lives, respawn, and collision outcomes should not be duplicated as authoritative client logic.
- Normal lifecycle logs should usually be debug-level; warnings/errors should be reserved for unusual or failed behavior.

## Considered But Deferred

- A real pause menu scene/overlay. Current pause support is functional plumbing without UI.
- A separate API server for accounts, matchmaking, leaderboards, persistence, or other non-gameplay backend concerns.
- Node.js/TypeScript with NestJS is the current planned stack for `services/api-server/`; see [docs/api/nestjs-api-server.md](api/nestjs-api-server.md).
- Packaging or launching the Go game server from the Godot client for local play.
- Client-side prediction/reconciliation beyond interpolation.
- More granular/documented collision shape export workflow.
- Logical gameplay viewport cap instead of raw OS window max size for balance.
- Ship variant foundation exists server-side: runtime ship type, `ship_type` state, resolved ship stats/modifiers, and collision shape ID lookup. Client scene mapping, real keyed collision catalogs, selection, and acquisition remain future work; see [docs/design/ship-variants.md](design/ship-variants.md).

## Current Short-Term Priorities

1. Manual two-client multiplayer smoke test: create room, join by code, ready both clients, start, enter gameplay, reach GameOver, return to lobby, leave, and verify empty-room cleanup.
2. Smoke test pause/resume in Godot with the server running.
3. Add a real pause/menu overlay scene and wire resume/menu options to the existing pause packets.
4. Revisit window/gameplay balance sizing so large monitors do not change gameplay difficulty.
5. Check current Git status for generated recordings or tmp binaries before committing.
6. Verify collision shape export/import after the Godot 4.6 upgrade.
7. Keep `game.gd` from growing again; move new UI behavior into `client/scripts/ui/` where possible.
8. Add focused client GUT coverage alongside packet, HUD, or `world_sync` changes instead of relying only on manual smoke testing.

## Longer-Term Ideas

- Local server launch/bundling for single-player/local play.
- Hosted online game server.
- Separate backend API server for non-real-time systems.
- Matchmaking, account identity, leaderboards, persistence.
- Enemy and pickup systems integrated with the scoring framework.
- More robust client prediction/reconciliation if networking latency becomes visible.
- Better tooling around collision shape generation and validation.
- Complete ship type variants with real alternate definitions, client scene mapping, and keyed server collision maps.

## Risks / Likely Messy Areas

- `client/scripts/game.gd`: central gameplay coordinator. It has already accumulated networking, HUD, effects, input, and state responsibilities. Continue extracting carefully.
- Pause/menu state: server rules exist, but UI is not complete. Be careful around `OpenMenu`, game-over return-to-menu, and websocket close behavior.
- Window/viewport sizing: raw window pixel limits are not reliable across monitors, DPI, title bars, taskbars, and Godot editor/debug behavior.
- Godot scene diffs: editor upgrades can add `uid`/`unique_id`/offset changes. Inspect scene diffs carefully before reverting.
- Collision shape freshness: server physics depends on shared JSON. Ensure client scene collision changes are reflected in shared collision data.
- Generated files: modifying generated files without changing the TOML source will be overwritten.
- Room cleanup timing: empty Lobby/InGame/GameOver rooms clean up through the rooms-domain cleanup helper. Reconnect is intentionally not implemented.
- Audio positioning: some sounds are 2D and can be inaudible if attached to the wrong world-space node. Past fixes moved sounds to player/effect-local nodes.

## Unresolved Questions

- What should the final pause menu UI look like, and should it include resume, main menu, settings, or room info?
- Should resume invulnerability block all scoring/shooting, or only player damage? Current implementation blocks shooting/scoring during invulnerability.
- Should player pause state be included in server snapshots for rendering other players as paused? Currently it is server-internal unless the client needs visual state later.
- What is the right logical viewport cap for spawning/visibility/game balance?
- Should `WINDOW_MAX_SIZE` remain as a real OS window maximum, or should it become a logical gameplay cap?
- Should the packet/constants generated client files move under `client/scripts/generated/` or a similar folder?
- How should the eventual API server share code, if at all, with the game server?
- Is `client/game-clip.avi` tracked or just present locally? It should not be committed.
- What final world wrap dimensions should be used for balance? Current values are in `shared/game_data.toml`.
- Ship variants now have a server-side stats modifier seam. Future work should decide which concrete stats/weapons/rules are allowed to vary for real ship definitions.

## Notes For Future Codex Sessions

- Always inspect current files first. The project has been refactored often, and stale assumptions are easy.
- Prefer small, reversible changes. The user is sensitive to unnecessary code growth and wants scalable structure without bloat.
- When asked to “answer” or “report,” do not edit files.
- When changing generated constants, edit `shared/game_data.toml` and run `tools/data_sync`. When changing packets, edit `shared/packets/packets.toml` and run `tools/data_sync`.
- When changing server gameplay rules, add or update focused Go tests under `services/game-server/tests/<area>/`.
- Do not add new Go server `*_test.go` files beside production packages under `services/game-server/internal/`.
- When changing generated packets, HUD behavior, `world_sync`, or pure client logic, add or update focused GUT tests under `client/tests/unit/`.
- Keep client test fixtures/helpers under `client/tests/fixtures/` and `client/tests/helpers/`, not under production `client/scripts/`.
- New server gameplay distance/position logic should go through `services/game-server/internal/game/space`; it is wrap-aware and keeps toroidal world behavior centralized.
- When changing Godot scenes, inspect `.tscn` diffs for accidental editor movement/offsets.
- Avoid broad rewrites of `game.gd`; extract only when the boundary is clear.
- Developer/debug toggles are documented in [docs/devtools/toggles.md](devtools/toggles.md).
- Toroidal wrap behavior and ship-variant plans are documented in [docs/design/toroidal-wrap.md](design/toroidal-wrap.md) and [docs/design/ship-variants.md](design/ship-variants.md).
- For logging, use `services/game-server/internal/logging` category loggers. Do not add raw `log.Println`.
- Before diagnosing gameplay/network issues, confirm the Go server is actually running and the client is connected.
- Current server tests pass with:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

- Current client GUT tests pass with:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

- Current client constants-boundary scan passes with:

```bash
python3 -m pytest tools/tests/test_client_constants_boundary.py
```

## TODO Snapshot

TODO: add pause menu scene and connect it to the already-existing pause/resume packets.

TODO: smoke test pause/resume with a running server after recent re-enable.

TODO: decide whether to replace OS window max-size enforcement with logical gameplay viewport clamping.

TODO: verify whether generated recordings/build artifacts are present in version control and remove them if tracked.

TODO: validate collision shape export after Godot 4.6.

TODO: consider moving generated Godot constants/packets into a generated scripts folder once references are easy to update.

TODO: document networking/rooms separately if room UX grows beyond the current architecture doc.
