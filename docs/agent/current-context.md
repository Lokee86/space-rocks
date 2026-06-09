# Agent Current Context

This file is volatile project memory. Read it only when the task depends on current refactor status, dirty worktree notes, recent Godot/editor changes, or known gaps.

Keep this file shorter than permanent docs. Remove stale notes aggressively.

For stable architecture/runtime maps, use [docs/design/architecture.md](../design/architecture.md).
For stable current-session orientation, use [docs/agent/session-primer.md](session-primer.md).
For current devtool toggle behavior and hotkeys, use [docs/devtools/toggles.md](../devtools/toggles.md).

## Current Context

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
- `services/player-data` exists as a sibling Go module with an independent codec, generated protocol packets, runtime/dispatcher, memory account/local stores, and guest no-op store.
- `cmd/game-server` can host the player-data runtime in-process through composition.
- SQLite, Rails adapter, real match resolution, gameplay wiring, and client UI remain later work.
- World Telemetry Overlay is implemented behind the devtools seam and toggled by `DevToggle9` / `9`.
- Overlay scene: `client/scenes/devtools/world_telemetry_overlay.tscn`; telemetry scripts live under `client/scripts/devtools/telemetry/`.
- Devtools coordination now lives under `client/scripts/devtools/context/` with `GameplayDevtoolsContext` acting as the facade/composition seam.
- Network telemetry uses `telemetry_ping` / `telemetry_pong`; gameplay state packets include `server_sent_msec`.
- `packet_age_ms` depends on server clock offset estimated from telemetry ping/pong, not raw wall-clock subtraction.
- Client auth checkpoint is working: Main Menu `Sign-in` opens Discord browser OAuth, Rails login-session exchange returns the normal Space Rocks bearer token, Godot stores the token, validates it through `/auth/me`, shows the display name in the menu, and clears local token plus signed-in state on logout.
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
