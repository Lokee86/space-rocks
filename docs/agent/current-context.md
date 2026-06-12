# Agent Current Context

This file is volatile project memory. Read it only when the task depends on current refactor status, dirty worktree notes, recent Godot/editor changes, or known gaps.

Keep this file shorter than permanent docs. Remove stale notes aggressively.

For stable architecture/runtime maps, use [docs/design/architecture.md](../design/architecture.md).
For stable current-session orientation, use [docs/agent/session-primer.md](session-primer.md).
For current devtool toggle behavior and hotkeys, use [docs/devtools/toggles.md](../devtools/toggles.md).

## Current Context

- MCP tooling is available: use the read-only info MCP for ChatGPT/planning and `space_rocks_write` for Codex implementation. See `docs/agent/mcp-servers.md`.
- The repo may be dirty.
- There may be unrelated Godot/editor asset changes in the worktree.
- Do not clean or revert unrelated user/editor changes casually.
- If gameplay or input looks broken, first confirm the Go server is running and the Godot client is connected.
- Godot was upgraded to 4.6 recently. Scene/import diffs may be noisy.
- The older `space-rocks-(4.3)/` project copy is ignored and should not be used as the active project.
- Generated recordings and build artifacts should not be committed. In particular, avoid committing `*.avi`, `tmp/`, `*/tmp/`, and `client/.godot/`.
- `1_up` and `torpedo` pickups exist.
- `torpedo` is pickup-acquired, not a default secondary.
- `pickup_class` drives pickup scene and collision family.
- Pickup type drives `Badge` icon selection and effect identity.
- Pickup weapon ammo is additive.
- Pickup collision export runs through `godot --headless --path client -s res://tools/export_collision_shapes.gd`.
- Pickup expiry emits `pickup_expired`.
- Debug hitboxes use `debug_shape_catalog` plus `GameplayState` transforms; there is no live `DebugOverlayStatePacket`.
- `GameplayRuntimeContext` is runtime wiring only; do not treat it as a read-model passthrough bucket.
- Server hitbox overlay data comes through `WorldSync`/devtools seams, not `GameplayRuntimeContext`.
- Targeting now sits above `MouseActionFlow`; `GameplayTargetingContext` owns target selection orchestration and `WorldSync` only exposes `target_source()`.
- Weapons live in `services/game-server/internal/game/weapons`; see [docs/design/weapons.md](../design/weapons.md).
- Radial effects live in `services/game-server/internal/game/effects/radial`; see [docs/design/radial-effects.md](../design/radial-effects.md).
- Weapon profiles can carry impact effects, torpedo uses a radial impact effect, radial effects emit hit intents, and Game applies radial hits through the damage seam.
- Room membership/owner state is behind the room membership owner seam.
- `websocket_write.go` only writes outbound/presentation state now; it does not advance game-over lifecycle.
- Continuous bullet stream runtime state is owned by `services/game-server/internal/devtools/streamruntime`.

## Current Direction Notes

- The user wants docs/plans to reflect a future Ruby/Rails API-only server separated from the Go game server.
- API/business/backend concerns should remain out of the Go real-time game server unless explicitly redirected.
- The user strongly prefers small implementation prompts and quick reviewable diffs.
- The user prefers scalable structure and useful seams over dumping more behavior into existing large files.
- Server-side account and local-profile work must follow [docs/design/cross-mode-routing-and-player-data.md](../design/cross-mode-routing-and-player-data.md): Local Single-Player allows Guest and Local Profile only, rejects Authenticated Account, Online Multiplayer requires Authenticated Account, and gameplay code must not directly choose embedded DB vs Rails/API.
- Account-shaped player data must also follow [docs/design/player-data-schema-ssot.md](../design/player-data-schema-ssot.md): `shared/player_data` contracts now exist, `shared/packets/player_data.toml` defines player-data packets, and gameplay code must not depend directly on Rails tables or embedded DB tables.
- `services/player-data` exists as a sibling Go module with an independent codec, generated protocol packets, and a configured runtime builder.
- `services/player-data` now has the Phase 4 routes for `authenticated_account` through the Rails adapter, `local_profile` through the SQLite adapter, and `guest` through singleton memory-backed stats.
- `cmd/game-server` can host the configured player-data runtime in-process through composition.
- Phase 4 Go match summary work is complete.
- Rooms now store one resolved `MatchResultSummary` on `game_over`.
- `Game` exposes match facts including score and `ship_deaths`.
- Winner resolution uses the highest multiplayer score.
- Ties produce no winner.
- Single-player produces no win.
- Summaries use `account_id`, `local_profile_id`, or neither for guest.
- Phase 5 match-result reporting is complete.
- Game-server reports resolved `MatchResultSummary` through `services/player-data`.
- `services/player-data` routes `account_id` to `authenticated_account`, `local_profile_id` to `local_profile`, and guest/no durable identity to guest behavior.
- Reporting is triggered from the existing room game-over lifecycle after the resolved summary exists.
- Successful reports are marked reported and not repeated.
- Failed reports do not mark reported.
- Backend player stats/reporting are implemented and committed.
- Client menu-flow Phase 1 / foundation is complete and green.
- Main Menu is a route launcher with login indicator/logout button.
- Single Player routes to Pregame Menu in single-player mode.
- Multiplayer routes to Pregame Menu in multiplayer mode.
- PregameMenu mode presentation works.
- Pregame Back returns to Main Menu.
- Old Main Menu multiplayer dialog/sign-in behavior is removed.
- Remaining client work is the menu-flow/profile/pilot/match-results UI slice sequence:
  - Single-player pregame actions
  - Sign In screen
  - Multiplayer pre-lobby create/join/logout routing
  - Local Pilot / Guest selection
  - Profile readout transmission
  - Match Results window
  - Stats refresh / final smoke
- Godot stats UI, save guest profile, live progression grants, currency, ship parts, unlocks, and achievements remain later work.
- World Telemetry Overlay is implemented behind the devtools seam and toggled by `DevToggle9` / `9`.
- Overlay scene: `client/scenes/devtools/world_telemetry_overlay.tscn`; telemetry scripts live under `client/scripts/devtools/telemetry/`.
- Devtools coordination now lives under `client/scripts/devtools/context/` with `GameplayDevtoolsContext` acting as the facade/composition seam.
- Network telemetry uses `telemetry_ping` / `telemetry_pong`; gameplay state packets include `server_sent_msec`.
- `packet_age_ms` depends on server clock offset estimated from telemetry ping/pong, not raw wall-clock subtraction.
- Client auth checkpoint is working: Main Menu `Sign-in` opens Discord browser OAuth, Rails login-session exchange returns the normal Space Rocks bearer token, Godot stores the token, validates it through `/api/auth/me`, shows the display name in the menu, and clears local token plus signed-in state on logout.
- Rails internal token verification exists for the Go game server.
- Go authclient exists and verifies Space Rocks bearer tokens through Rails.
- Websocket `authenticate_request` / `authenticate_result` exists.
- Multiplayer create/join admission is auth-aware.
- Single-player remains unchanged and does not require auth.
- Local/no-auth game-server mode can still allow multiplayer for dev because server-side admission remains authoritative.
- Non-Discord in-game account creation UI is still deferred.

## Known Gaps / TODOs

- Generated Godot constants/packet files may eventually move into a generated folder, but they currently live under `client/scripts/`.
- API server scaffold exists, but no product features are implemented yet.
- Ship variants are planned but not implemented.
- Client packet codec callers now consume `PacketEncodeResult` and `PacketDecodeResult`; the codec at `client/scripts/networking/packets/packet_codec.gd` owns JSON parsing plus envelope validation only.
