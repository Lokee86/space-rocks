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

## Current Client Runtime Areas

Common starting points:

- `client/scripts/ui/game_shell.gd`
- `client/scripts/game.gd`
- `client/scripts/networking/network_client.gd`
- `client/scripts/networking/packet_codec/packet_codec.gd`
- `client/scripts/networking/world_sync.gd`
- `client/scripts/entities/player.gd`
- `client/scripts/ui/hud_controller.gd`

## Spectate / Lifecycle Rules

Keep packet-facing player lifecycle status in `StatePacket.player_lifecycle`, beside `players`.

Do not put match lifecycle on `ShipState`; pending-respawn and eliminated players may not have active ship state.

Client spectate/view-cycle eligibility must use authoritative lifecycle status (`active`) plus visual availability.

Do not infer active eligibility solely from remote player positions or ship presence.

## Implemented Developer Toggles

Current hardcoded Godot hotkeys:

- `F1`: toggle debug invincibility for the player
- `F2`: toggle debug infinite lives for the player
- `F3`: toggle room-wide debug world freeze

These are server-authoritative toggles sent through generated packets. See `docs/devtools/toggles.md`.

Devtools must route through real gameplay seams. Do not create parallel debug-only gameplay logic that bypasses damage, lives, spawning, scoring, movement, room/session, or modifier systems.

## Pause State Context

Pause plumbing exists:

- packets: `pause_player`, `resume_player`
- server player fields include paused/invulnerability state
- paused players should ignore input, not shoot/score, not take asteroid damage, and be hidden by client world sync
- resume starts a short invulnerability window
- menu UI has been in flux

If pause behavior seems wrong, inspect current Godot scenes/scripts before changing code. The HUD/menu scenes have been changed multiple times.

If gameplay or input looks broken, first confirm the Go server is running and the Godot client is connected. This caused a false pause-feature debugging path before.

## Toroidal / Wrapped World

Client wrap math lives under the Godot client and should align with server world rules.

Future client rendering should use unwrapped visual positions relative to the local player so border crossing is invisible.

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

Run GUT when the `godot` CLI is available:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

For client constants-boundary changes, run:

```bash
python3 -m pytest tools/tests/test_client_constants_boundary.py
```

## Known Client Gaps

- Pause/menu UI still needs smoke testing and may still be evolving.
- Window/gameplay balance should move away from raw OS max window pixels toward a logical gameplay viewport cap.
- Collision shape export/import should be verified after the Godot 4.6 upgrade.
- Generated Godot constants/packet files may eventually move into a generated folder, but they currently live under `client/scripts/`.
- Ship variants are planned but not implemented.
- Toroidal wrapping is implemented and still needs manual gameplay smoke testing after related changes.
