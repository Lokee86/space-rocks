# Session Primer

Use this as the short orientation layer for new sessions. It should stay stable enough to be useful, but narrow enough to avoid becoming a second architecture manual.

## MCP Tooling Reminder

Space Rocks has two local MCP servers under `tools/space-rocks-mcp`.

- The info MCP server uses `server-info-next.js` on port `8789`. It is read-only and is intended for ChatGPT/planning/diagnosis. It exposes repo read/search tools plus read-only EngineForge/Godot bridge diagnostics.
- The write MCP server uses `server-write.js` on port `8788`. It is write-capable and is intended for Codex implementation. It exposes bounded repo writes, allowlisted commands, and explicit Godot bridge mutation tools.
- For Godot scene/UI diagnosis, use the info server’s bridge tools before guessing from files alone.
- For implementation work that needs repo writes or Godot scene edits, prompt Codex to use `space_rocks_write`.
- Never expose the write MCP server through ngrok. Remote access, if needed, should be read-only info server only.
- Full reference: `docs/agent/mcp-servers.md`.

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
- Phase 5 match-result reporting is complete: resolved `MatchResultSummary` is reported through `services/player-data`, `account_id` routes to `authenticated_account`, `local_profile_id` routes to `local_profile`, and guest/no durable identity routes to guest behavior.
- Profile readout is complete: the client calls `POST /api/player-data/profile` on the game-server data-handler, guest reads hit in-process memory, and authenticated reads flow through `RailsStore` to `POST /api/internal/player-data/stats`.
- Client menu-flow Phase 1 foundation is complete and green: Main Menu is a route launcher, `MenuFlowController` owns scene routing, `pregame_menu.tscn` is the shared shell, `PregameModePresenter` owns mode display, and Pregame Back returns to Main Menu.
- Client menu-flow Phase 2 is complete and green: Pregame Play Endless starts the old single-player flow, `MenuFlowController` clears menu UI when gameplay starts, and Pregame Back still returns to Main Menu.
- Client menu-flow Phase 3 is complete and green: signed-out Multiplayer opens `LoginWindow`, Discord login works from the Sign In screen, signed-in Multiplayer routes to Pregame, and successful Discord auth returns to Multiplayer Pregame.
- Client menu-flow Phase 4 is complete and green: Multiplayer Pregame Create/Join/Logout work, and Lobby Leave returns to Multiplayer Pregame without logging out.
- Client menu-flow Phase 5 is complete and green: profile readout transmission mounts `profile_readout.tscn` under `TransmissionScreen/ScreenDisplay`, and the readout fills callsign plus stat labels for guest and authenticated account contexts.
- Next near-term work is Match Results plus the small `GameMenuFlow` fix, then Local Pilot / Guest selector, then final stats refresh smoke.
- Devtools coordination moved under `client/scripts/devtools/context/` with `GameplayDevtoolsContext` as the facade/composition seam.
- Continuous bullet stream runtime state was isolated in `services/game-server/internal/devtools/streamruntime`.
- Pickup entity/drop/collection/lifespan/expiry work is complete through the pickup seam.
- Pickup presentation blink is client-side and derives from age/lifespan packet state.
- Player-data foundation is now complete as a sibling `services/player-data` module with shared packet SSoT/generated protocol, independent codec, in-process game-server runtime hosting, Rails adapter for authenticated_account, SQLite adapter for local_profile, and singleton memory stats for guest.

## Fragile Or Moving Areas

- Dev-readiness item 11 is still open: local-player camera piggybacking must be replaced with a dedicated camera target/controller.
- Generated Godot constants and packet files still live under `client/scripts/` for now, even though that is not the ideal long-term shape.
- Godot stats UI and profile readout are implemented through the data-handler route; save guest profile, live progression grants, currency, ship parts, unlocks, and achievements are still future work.
- Ship variants are planned but not implemented.
- Client packet codec callers already use `PacketEncodeResult` and `PacketDecodeResult`; the codec itself should stay focused on JSON parsing and envelope validation only.

## Near-Term Direction

- Keep docs and seams aligned with the server-authoritative split.
- Prefer small ownership moves over broad gravity-well edits.
- Keep gameplay-owned logic on the server and presentation/UI/effects on the client.
- Continue pushing future business/backend concerns toward the planned API server instead of growing the Go game server.
- Keep packet and constants changes flowing through the source-of-truth TOML plus data-sync path.
- Keep pickup presentation blink client-side from age/lifespan packet state.
- Profile readout transmission is complete; do not re-open it as the next slice.
- Match Results plus the small `GameMenuFlow` fix is the next active slice.
- Local Pilot / Guest selector is deferred until after Match Results.
- Stats refresh / final smoke remains the last slice.
- Active-game and personal-death menu behavior should not change, and multiplayer Lobby Leave should return to Multiplayer Pregame without logging out.

## Common Mistakes To Avoid

- Do not treat volatile session memory as a replacement for the permanent architecture docs.
- Do not expand `docs/notes.md`; keep it as a small parking lot only.
- Do not add new behavior to gravity-well files when a seam already exists or a smaller seam can be introduced.
- Do not blur `target_kind` + `target_id` back into legacy `target_player_id` for new gameplay work.
- Do not assume generated files are safe to hand-edit as a convenience.
- Do not mix unrelated refactors into a docs-only or seam-specific prompt.
- Do not add damage math to `combat.go`, and do not bypass the real damage seam from devtools.
- Do not reintroduce client-side profile stat mutation.
- Do not route profile readout directly to Rails `/api/player/stats`.
- Do not require `PLAYER_DATA_RAILS_BEARER_TOKEN`.
