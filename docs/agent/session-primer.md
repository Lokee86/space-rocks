# Session Primer

Use this as the short orientation layer for new sessions. It should stay stable enough to be useful, but narrow enough to avoid becoming a second architecture manual.

## Active Architecture State

- Gameplay is server-authoritative.
- The Godot client handles rendering, UI, audio/effects, local input collection, interpolation, and presentation-facing devtools.
- The Go game server owns simulation outcomes, including movement, bullets, collisions, scoring, lives, death, respawn, pause safety, rooms, and websocket state.
- API/business/backend concerns are intentionally kept out of the Go game server until the planned API server is real.

## Major Completed Seams

- Room membership and owner state moved behind a room membership owner seam.
- `websocket_write.go` is outbound/presentation only and no longer advances game-over lifecycle.
- Targeting orchestration now sits above `MouseActionFlow`; `GameplayTargetingContext` owns selection orchestration and `WorldSync` only exposes `target_source()`.
- The upgraded damage seam now lives in `services/game-server/internal/game/damage/`; `ResolveSingle` handles modifiers, shields, area damage, and DoT at a high level while `game` owns adapters and entity mutation.
- Weapons live in `services/game-server/internal/game/weapons` and radial effects live in `services/game-server/internal/game/effects/radial`; weapon profiles may carry impact effects, torpedo uses a radial impact effect, radial effects emit hit intents, and Game applies radial hits through the damage seam. See [docs/design/weapons.md](../design/weapons.md) and [docs/design/radial-effects.md](../design/radial-effects.md).
- Rails internal token verification, Go authclient, websocket session identity, and websocket auth packets now form the completed auth/admission seam for multiplayer admission.
- Devtools coordination moved under `client/scripts/devtools/context/` with `GameplayDevtoolsContext` as the facade/composition seam.
- Continuous bullet stream runtime state was isolated in `services/game-server/internal/devtools/streamruntime`.
- Pickup entity/drop/collection/lifespan/expiry work is complete through the pickup seam.
- Pickup presentation blink is client-side and derives from age/lifespan packet state.

## Fragile Or Moving Areas

- Dev-readiness item 11 is still open: local-player camera piggybacking must be replaced with a dedicated camera target/controller.
- Generated Godot constants and packet files still live under `client/scripts/` for now, even though that is not the ideal long-term shape.
- The API server is planned but not scaffolded.
- Local Profile, embedded DB, player-data routing, and player-data SSoT are still future work.
- Ship variants are planned but not implemented.
- Client packet codec callers already use `PacketEncodeResult` and `PacketDecodeResult`; the codec itself should stay focused on JSON parsing and envelope validation only.

## Near-Term Direction

- Keep docs and seams aligned with the server-authoritative split.
- Prefer small ownership moves over broad gravity-well edits.
- Keep gameplay-owned logic on the server and presentation/UI/effects on the client.
- Continue pushing future business/backend concerns toward the planned API server instead of growing the Go game server.
- Keep packet and constants changes flowing through the source-of-truth TOML plus data-sync path.
- Keep pickup presentation blink client-side from age/lifespan packet state.

## Common Mistakes To Avoid

- Do not treat volatile session memory as a replacement for the permanent architecture docs.
- Do not expand `docs/notes.md`; keep it as a small parking lot only.
- Do not add new behavior to gravity-well files when a seam already exists or a smaller seam can be introduced.
- Do not blur `target_kind` + `target_id` back into legacy `target_player_id` for new gameplay work.
- Do not assume generated files are safe to hand-edit as a convenience.
- Do not mix unrelated refactors into a docs-only or seam-specific prompt.
- Do not add damage math to `combat.go`, and do not bypass the real damage seam from devtools.
