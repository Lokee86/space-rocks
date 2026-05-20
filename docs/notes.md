# Project Notes

This file is project memory for future Codex sessions. It is not a full architecture document. For that, see [docs/design/architecture.md](design/architecture.md). For practical development commands and conventions, see [docs/developer.md](developer.md).

Always prefer cleaner, scalable design choices.

## Current Development Status

Space Rocks is playable in development form with:

- Godot client
- Go websocket game server
- main menu and game loop scene shell
- room support through `/ws?room_id=...`
- server-authoritative player movement, bullets, asteroid collisions, scoring, lives, death, respawn, and safe spawn placement
- generated shared constants and packet helpers
- structured server logging
- HUD score/lives/room/death/game-over display
- audio/effects for shooting, asteroid impacts, ship death, and game over
- pause-state server plumbing and client toggle plumbing, but no real pause menu scene yet

The project is still moving quickly. Treat recent systems as subject to refinement.

## Recently Implemented Systems

### Server Test Layout

Go server tests have been moved out of production package folders and now live under:

```text
services/game-server/tests/
```

Current subareas:

- `game`
- `networking`
- `physics`
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

Boundary note: server-owned constants live under `constants.server.*`. `player_starting_lives` and `player_respawn_delay` live under `constants.server.player_lifecycle`; `asteroid_size_scale` lives under `constants.server.asteroids`. The client should receive lives through player/state data, respawn delay through death events, and asteroid scale through asteroid state instead of importing those constants.

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

### Server Rooms

`services/game-server/internal/networking` owns websocket handling and rooms.

Current behavior:

- `/ws` joins the default room.
- `/ws?room_id=abc` creates or joins a separate room.
- each room owns its own `*game.Game`.
- empty rooms schedule cleanup after a grace period.

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

## Important Design Decisions

- The server is authoritative for game rules.
- The client is responsible for presentation, UI, audio/effects, interpolation, and input collection.
- Network transport belongs in `services/game-server/internal/networking`, not `cmd/game-server/main.go`.
- Business/API concerns belong in the planned `services/api-server/` service, not in the Go game server.
- Room state owns a separate `*game.Game` per room.
- Shared constants/packets should be generated, not copied by hand. Constants use `shared/game_data.toml` and `tools/data_sync/`; packets use `shared/packets/packets.toml` and `tools/data_sync/`. Output filtering may keep server-owned constants out of client generated files even when they remain in the constants source of truth.
- Collision shapes are shared through JSON and used by the server physics package.
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
- Invisible toroidal/wrapped playfield; see [docs/design/toroidal-wrap.md](design/toroidal-wrap.md).
- Ship variants with different client scenes and server collision maps; see [docs/design/ship-variants.md](design/ship-variants.md).

## Current Short-Term Priorities

1. Smoke test pause/resume in Godot with the server running.
2. Add a real pause/menu overlay scene and wire resume/menu options to the existing pause packets.
3. Revisit window/gameplay balance sizing so large monitors do not change gameplay difficulty.
4. Check current Git status for generated recordings or tmp binaries before committing.
5. Verify collision shape export/import after the Godot 4.6 upgrade.
6. Keep `game.gd` from growing again; move new UI behavior into `client/scripts/ui/` where possible.

## Longer-Term Ideas

- Local server launch/bundling for single-player/local play.
- Hosted online game server.
- Separate backend API server for non-real-time systems.
- Matchmaking, account identity, leaderboards, persistence.
- Enemy and pickup systems integrated with the scoring framework.
- More robust client prediction/reconciliation if networking latency becomes visible.
- Better tooling around collision shape generation and validation.
- Invisible toroidal world wrapping to keep multiplayer players in one arena while preserving endless-feeling flight.
- Ship type variants with server-selected collision maps and client scene mapping.

## Risks / Likely Messy Areas

- `client/scripts/game.gd`: central gameplay coordinator. It has already accumulated networking, HUD, effects, input, and state responsibilities. Continue extracting carefully.
- Pause/menu state: server rules exist, but UI is not complete. Be careful around `OpenMenu`, game-over return-to-menu, and websocket close behavior.
- Window/viewport sizing: raw window pixel limits are not reliable across monitors, DPI, title bars, taskbars, and Godot editor/debug behavior.
- Godot scene diffs: editor upgrades can add `uid`/`unique_id`/offset changes. Inspect scene diffs carefully before reverting.
- Collision shape freshness: server physics depends on shared JSON. Ensure client scene collision changes are reflected in shared collision data.
- Generated files: modifying generated files without changing the TOML source will be overwritten.
- Room cleanup/reconnect timing: room cleanup exists with grace time, but reconnect flows should be tested when multiplayer UX grows.
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
- What final world wrap dimensions should be used? Proposed planning values are currently `1062 x 5250`, but width should be confirmed.
- Should ship variants eventually affect only collision/visuals, or also movement stats, weapons, and gameplay rules?

## Notes For Future Codex Sessions

- Always inspect current files first. The project has been refactored often, and stale assumptions are easy.
- Prefer small, reversible changes. The user is sensitive to unnecessary code growth and wants scalable structure without bloat.
- When asked to “answer” or “report,” do not edit files.
- When changing generated constants, edit `shared/game_data.toml` and run `tools/data_sync`. When changing packets, edit `shared/packets/packets.toml` and run `tools/data_sync`.
- When changing server gameplay rules, add or update focused Go tests under `services/game-server/tests/<area>/`.
- Do not add new Go server `*_test.go` files beside production packages under `services/game-server/internal/`.
- New server gameplay distance/position logic should go through `services/game-server/internal/game/space`; it is flat/infinite today, with no-op normalization, but keeps future wrapped-world work localized.
- When changing Godot scenes, inspect `.tscn` diffs for accidental editor movement/offsets.
- Avoid broad rewrites of `game.gd`; extract only when the boundary is clear.
- Developer/debug toggles are documented in [docs/devtools/toggles.md](devtools/toggles.md).
- Future toroidal wrap and ship-variant plans are documented in [docs/design/toroidal-wrap.md](design/toroidal-wrap.md) and [docs/design/ship-variants.md](design/ship-variants.md).
- For logging, use `services/game-server/internal/logging` category loggers. Do not add raw `log.Println`.
- Before diagnosing gameplay/network issues, confirm the Go server is actually running and the client is connected.
- Current server tests pass with:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

## TODO Snapshot

TODO: add pause menu scene and connect it to the already-existing pause/resume packets.

TODO: smoke test pause/resume with a running server after recent re-enable.

TODO: decide whether to replace OS window max-size enforcement with logical gameplay viewport clamping.

TODO: verify whether generated recordings/build artifacts are present in version control and remove them if tracked.

TODO: validate collision shape export after Godot 4.6.

TODO: consider moving generated Godot constants/packets into a generated scripts folder once references are easy to update.

TODO: document networking/rooms separately if room UX grows beyond the current architecture doc.
