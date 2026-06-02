# Agent Godot Notes

Use this when changing the Godot client, scenes, UI, GDScript, GUT tests, assets, imports, or developer toggles.

## Client Responsibilities

The Godot client owns:

- rendering
- UI
- audio/effects
- local input collection
- interpolation
- client presentation controllers

The Go server owns authoritative simulation outcomes.

## Scene and Import Safety

Be careful with Godot scene diffs. Godot may rewrite:

- `uid`
- `unique_id`
- offsets
- imports
- scene metadata

Do not revert user/editor changes unless explicitly requested.

Godot was upgraded to 4.6 recently. Scene/import diffs may be noisy.

There are unrelated Godot/editor asset changes in the worktree sometimes. Do not clean or revert them casually.

Generated recordings and build artifacts should not be committed. In particular, avoid committing:

- `*.avi`
- `tmp/`
- `*/tmp/`
- `client/.godot/`

The older `space-rocks-(4.3)/` project copy is ignored and should not be used as the active project.

## Client Packet Codec

Route client packet wire JSON through:

```text
client/scripts/networking/packet_codec/packet_codec.gd
```

Do not add direct `JSON.stringify` or `JSON.parse_string` calls in websocket packet paths.

The client codec is intentionally JSON-only and thin. Do not add validation, format switching, typed packet objects, protobuf references, or generator changes unless explicitly requested.

`network_client.gd` still owns websocket behavior.

## World Script Paths

Current world sync/wrap ownership paths:

- `client/scripts/world/world_sync.gd`
- `client/scripts/world/world_wrap.gd`
- `client/scripts/world/local_visual_sync.gd`
- `client/scripts/world/player_sync.gd`

## Spectate / Lifecycle Rules

Keep packet-facing player lifecycle status in `StatePacket.player_lifecycle`, beside `players`.

Do not put match lifecycle on `ShipState`; pending-respawn and eliminated players may not have active ship state.

Client spectate/view-cycle eligibility must use authoritative lifecycle status (`active`) plus visual availability.

Do not infer active eligibility solely from remote player positions or ship presence.

## HUD-first mouse input routing

- HUD/UI controls get first priority for mouse clicks.
- Gameplay mouse handling should run only after visible HUD controls have had a chance to receive the click.
- Broad HUD layout containers should pass/ignore mouse events so they do not swallow clicks.
- Actual interactive controls, such as GameMenu TextureButtons, should keep normal button mouse handling.
- Gameplay target selection should only consume left-click when a target/action is actually selected.
- If no target/action is available, gameplay should return `false` so the click can continue through normal Godot routing.

Why this rule exists:

- `_unhandled_input` alone was not reliable because other controls/layers can swallow clicks before gameplay sees them.
- Pure gameplay `_input` consumption was also wrong because it could prevent menu buttons from receiving clicks.
- The working model is HUD priority first, then world/gameplay fallback.

## Implemented Developer Toggles

Current hardcoded dev toggles use number keys (`DevToggle0` through `DevToggle9`). Use the canonical map in `docs/devtools/toggles.md`.

Pause/menu is separate from dev toggles and should route through `OpenMenu`, not `DevToggle4`.

These are server-authoritative toggles sent through generated packets where applicable.

Devtools must route through real gameplay seams. Do not create parallel debug-only gameplay logic that bypasses damage, lives, spawning, scoring, movement, room/session, or modifier systems.

Devtools UI/controller/read-model code belongs under `client/scripts/devtools/`. Devtools scenes belong under `client/scenes/devtools/`.

Do not place devtools presentation/read models under `client/scripts/gameplay/` just because they consume gameplay state.

Player-targeting `OptionButton` nodes in the devtools window should use `Select` naming, not `Option` naming. Current select node names:

- `InvincibleStatusSelect`
- `InfiniteLivesSelect`
- `PlayerFrozenSelect`

Keep `docs/devtools/toggles.md` as the canonical behavior reference.

## Pause / Menu Context

Pause plumbing exists:

- packets: `pause_player`, `resume_player`
- server player fields include paused/invulnerability state
- paused players should ignore input, not shoot/score, not take asteroid damage, and be hidden by client world sync
- resume starts a short invulnerability window
- pause/menu UI exists but is still evolving
- input routing should treat pause/menu as `OpenMenu` behavior, not a devtool toggle

Pause/menu behavior still needs smoke testing, especially active-game pause, GameOver menu behavior, ReturnToLobby, and websocket preservation.

If pause behavior seems wrong, inspect current Godot scenes/scripts before changing code. The HUD/menu scenes have been changed multiple times.

If gameplay or input looks broken, first confirm the Go server is running and the Godot client is connected. This caused a false pause-feature debugging path before.

## Toroidal / Wrapped World

Client wrap math lives under the Godot client and should align with server world rules.

Future/current client rendering should use unwrapped visual positions relative to the local player so border crossing is invisible.

See:

- `docs/design/toroidal-wrap.md`
- `services/game-server/internal/game/space`

## Ship Variants

Future ships may use different client scenes and server collision maps.

See:

- `docs/design/ship-variants.md`

## Client Testing

Godot client tests use GUT and live under:

```text
client/tests/
```

Unit tests go under:

```text
client/tests/unit/
```

Fixtures go under:

```text
client/tests/fixtures/
```

Reusable test-only helpers go under:

```text
client/tests/helpers/
```

Do not put test helpers in `client/scripts/`.

The GUT command is normally human-run unless the prompt explicitly allows terminal commands:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

For client constants-boundary changes, the pytest boundary scan is normally human-run unless the prompt explicitly allows terminal commands:

```bash
python3 -m pytest tools/tests/test_client_constants_boundary.py
```

## Known Client Gaps

- Pause/menu UI is functional but still evolving; smoke-test game-over, spectate, return-to-lobby, and websocket preservation after menu/input changes.
- Window/gameplay balance should move away from raw OS max window pixels toward a logical gameplay viewport cap.
- Collision shape export/import should be verified after the Godot 4.6 upgrade.
- Generated Godot constants/packet files may eventually move into a generated folder, but they currently live under `client/scripts/`.
- Ship variants are planned but not implemented.
- Toroidal wrapping is implemented and still needs manual gameplay smoke testing after related changes.
