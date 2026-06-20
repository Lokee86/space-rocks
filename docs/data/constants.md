# Constants Pipeline

Parent index: [Data](./!INDEX.md)

## Purpose

This document describes the current constants source-of-truth files, generated outputs, pipeline behavior, consumers, validation workflow, and implementation map for Space Rocks constants.

## Overview

Constants are authored in TOML files under `shared/constants/` and synchronized into generated Go and GDScript outputs through `data-sync`.

The constants pipeline keeps runtime tuning, shared world dimensions, server gameplay values, client presentation values, shell/session strings, lobby UI text, weapon tuning, pickup definitions, and weapon-pickup tuning aligned across the game server and Godot client.

The TOML files are the source of truth. Generated language files are outputs and should not be hand-edited except when intentionally diagnosing generated drift.

The active generated outputs are:

```text
services/game-server/internal/constants/*.go
client/scripts/generated/constants/constants.gd
```

Constants are routed by `data-sync` managed block markers. A TOML section is generated into a destination only when a generated destination file contains a matching block such as:

```go
// data-sync:start constants.shared.world
// data-sync:end constants.shared.world
```

or:

```gdscript
# data-sync:start constants.shared.world
# data-sync:end constants.shared.world
```

## Source files

The active constants source files are configured in `tools/data_sync/config.toml`.

### Server and shared runtime constants

* `shared/constants/server_constants.toml`

Current sections:

```text
constants.server.runtime
constants.shared.world
constants.server.scoring
constants.server.damage
```

These sections own server tick rate, shared world dimensions, base score, and baseline damage/health constants.

### Server entity and simulation constants

* `shared/constants/server_entities.toml`

Current sections:

```text
constants.server.player_movement
constants.server.player_session
constants.server.player_lifecycle
constants.server.asteroids
constants.server.collision
```

These sections own player movement tuning, respawn/session values, lifecycle values, asteroid spawn/despawn tuning, and collision despawn delay.

### Weapon constants

* `shared/constants/weapons.toml`

Current sections:

```text
constants.server.weapons.basic_cannon
constants.server.weapons.torpedo
constants.shared.weapons.torpedo
```

Server weapon sections own projectile speed, projectile lifetime, spawn offset, cooldown, impact damage, radial damage, and radial timing.

The shared torpedo section is generated to both Go and GDScript because the server needs radial gameplay values and the client needs matching presentation/readout values such as radial zone count, zone width, and cooldown display.

### Pickup constants

* `shared/constants/pickups.toml`

Current sections:

```text
constants.server.powerups.one_up
constants.server.powerups.torpedo
```

The current generated server powerup output emits the one-up powerup block. Torpedo weapon-pickup constants are currently generated from `shared/constants/weapon_pickups.toml`.

### Weapon pickup constants

* `shared/constants/weapon_pickups.toml`

Current sections:

```text
constants.server.weapon_pickups.torpedo
```

This section owns torpedo pickup type, class, health, lifespan, weapon id, target weapon slot, and pickup ammo value.

### Client presentation constants

* `shared/constants/client/presentation.toml`

Current sections:

```text
constants.client.presentation.background
constants.client.presentation.rendering
constants.client.presentation.player_visuals
constants.client.presentation.effects
constants.client.presentation.viewport
```

These sections own client-only visual and presentation values: parallax, drift, z-indexes, player hue policy, effect cleanup timing, pickup end-of-life flash timing, viewport bounds, and offscreen indicator margins.

### Client shell constants

* `shared/constants/client/shell.toml`

Current sections:

```text
constants.client.shell.network
constants.client.shell.room_states
constants.client.shell.shell_states
constants.client.shell.session
constants.client.shell.game_menu
```

These sections own client websocket target URLs/origin, network polling priority, room-state strings, shell-state strings, session-mode strings, boot request strings, connect result strings, and gameplay menu primary-action identifiers.

### Client lobby constants

* `shared/constants/client/lobby.toml`

Current sections:

```text
constants.client.lobby.status
constants.client.lobby.dialog
constants.client.lobby.buttons
```

These sections own lobby status text, join-dialog status text, and ready-button text.

## Configuration

The active config file is:

```text
tools/data_sync/config.toml
```

The constants source paths are listed under:

```toml
[sot.constants]
paths = [...]
```

The active constants destination scan is:

```toml
[constants.scan]
include = ["services/game-server/internal/constants/*.go", "client/scripts/generated/constants/*.gd"]
exclude = [".git/**", "**/.godot/**", "**/node_modules/**"]
```

The current config does not use explicit `[constants.go]` or `[constants.gds]` target section lists. Instead, `data-sync` discovers generated constants files through `constants.scan`, reads the managed block markers inside those files, and updates each discovered block from the TOML section with the same name.

## Generated outputs

### Game server

Generated Go constants live under:

```text
services/game-server/internal/constants/
```

Current generated files:

```text
services/game-server/internal/constants/constants.go
services/game-server/internal/constants/weapons.go
services/game-server/internal/constants/powerups.go
services/game-server/internal/constants/weapon_pickups.go
```

Go constant names are generated from snake_case TOML keys into PascalCase names.

Example:

```toml
world_width = 17200.0
```

generates:

```go
const WorldWidth = 17200.0
```

### Client

Generated GDScript constants live in:

```text
client/scripts/generated/constants/constants.gd
```

GDScript constant names are generated from snake_case TOML keys into upper snake case names.

Example:

```toml
single_player_ws_url = "ws://localhost:8080/ws"
```

generates:

```gdscript
const SINGLE_PLAYER_WS_URL := "ws://localhost:8080/ws"
```

GDScript also supports two-number TOML arrays as `Vector2` values.

Example:

```toml
window_min_size = [1280.0, 720.0]
```

generates:

```gdscript
const WINDOW_MIN_SIZE := Vector2(1280.0, 720.0)
```

### TypeScript

The data-sync tool has TypeScript constants generator and parser support, but there is no active TypeScript constants destination in the current project config.

## Data role

Constants data owns scalar tuning and shared named values.

Good constants candidates include:

```text
numeric tuning values
string identifiers used by multiple runtime files
client presentation tuning
server gameplay tuning
shared world dimensions
shared weapon values needed by both server and client
```

Constants should not own structured catalogs when another data source is the better owner.

Examples:

```text
Asteroid variant catalog -> shared/asteroids/variants.toml
Drop tables -> shared/drop_tables/*.toml
Packet schemas -> shared/packets/*.toml
Collision bodies -> shared/collisions/collision_shapes.json
Player-data schema -> shared/player_data/*.toml
```

The legacy `constants.AsteroidVariants` style should not be reintroduced. Asteroid variant count and ordering are owned by the asteroid variant catalog, not constants.

## Pipeline usage

`data-sync` supports these operations:

```text
-push
-pull
-diff
-check
-validate
```

For constants, the normal workflow is:

```bash
data-sync -validate -constants
data-sync -diff -constants -go -gds
data-sync -push -constants -go -gds
data-sync -check -constants -go -gds
```

### Push

`-push -constants` reads the configured constants TOML paths, discovers generated destination blocks for the requested language, regenerates each block from the matching TOML section, and writes changed generated files.

`-push` does not create new managed blocks. New sections need matching `data-sync:start` / `data-sync:end` markers in the generated destination file before they can be populated.

### Diff

`-diff -constants` performs the same generation planning as `-push`, prints a unified diff, and writes nothing.

Use this before push when reviewing generated changes.

### Check

`-check -constants` performs the same generation planning as `-push`, writes nothing, and exits non-zero when generated output is stale.

This is the drift-detection command for CI-style verification.

### Validate

`-validate -constants` checks config and constants integrity.

Current validation includes:

```text
configured source files exist
TOML can be parsed
configured generated files exist
managed blocks are present where configured
duplicate constants sections across source files are rejected
discovered output sections must have matching TOML source sections
constant names must be valid snake_case
constant values must use supported value types
```

### Pull

`-pull -constants` exists for reverse synchronization, but it is intentionally restricted.

Rules:

```text
only one language may be pulled at a time
only existing TOML values may be updated
constant keys may not be added, removed, renamed, or reordered through pull
conflicting generated blocks for the same section fail
missing source sections fail
non-canonical generated constants fail
```

Use pull only when intentionally back-porting generated constant value edits into TOML. Normal changes should be made in TOML and pushed forward.

## Supported values and naming

Supported constant value types are:

```text
bool
int
float
string
list[float] with exactly two numeric values for GDScript Vector2 output
```

Names must be snake_case.

Invalid names include:

```text
not-snake
_leading
trailing_
double__underscore
MixedCase
```

The generator preserves TOML value order inside each section.

Go output supports bool, int, float, and string constants.

GDScript output supports bool, int, float, string, and two-number `Vector2` values.

TypeScript output supports bool, int, float, and string constants, but it is not an active configured output.

## Consumers

### Game server consumers

The game server imports generated constants from:

```text
services/game-server/internal/constants
```

Current consumers include:

```text
services/game-server/internal/game/runtime/
services/game-server/internal/game/physics/
services/game-server/internal/game/space/
services/game-server/internal/game/spawning/
services/game-server/internal/game/scoring/
services/game-server/internal/game/entities/pickups/
services/game-server/internal/game/weapons/
services/game-server/internal/game/
services/game-server/internal/networking/
```

Important runtime uses include:

```text
server tick rate
world bounds
player movement tuning
player lives and respawn timing
pause/resume invulnerability timing
asteroid spawn and despawn tuning
collision despawn timing
score value
damage and health values
weapon projectile/radial tuning
pickup definition values
network gameplay tick cadence
```

### Client consumers

The Godot client preloads:

```text
res://scripts/generated/constants/constants.gd
```

Current consumers include:

```text
client/scripts/boot/
client/scripts/session/
client/scripts/shell/
client/scripts/networking/
client/scripts/lobby/
client/scripts/ui/
client/scripts/entities/
client/scripts/gameplay/
client/scripts/presentation/
client/scripts/world/
```

Important client uses include:

```text
single-player and multiplayer websocket target selection
multiplayer websocket Origin header
network polling priority
room-state and shell-state string comparisons
boot request state
connect result values
gameplay menu action values
lobby and join-dialog text
world wrap dimensions
background parallax and drift
render z-indexes
player hue policy
effect timing and cleanup
pickup warning flash timing
weapon HUD cooldown display
```

## Shared world constants

`constants.shared.world` is generated to both server and client.

Current values:

```text
WORLD_WIDTH / WorldWidth
WORLD_HEIGHT / WorldHeight
```

The server uses these through `space.DefaultBounds()` for authoritative wrapped-space behavior.

The client uses the same values through `world_wrap.gd` for visual wrapping and shortest-delta rendering.

These values must stay aligned. Regenerating only Go or only GDScript after changing world size can create server/client wrap disagreement.

## Shared weapon constants

`constants.shared.weapons.torpedo` is generated to both server and client.

Current shared values:

```text
torpedo_radial_zone_count
torpedo_radial_zone_width
torpedo_cooldown
```

The server uses these values for torpedo radial effect behavior and weapon cooldown policy.

The client uses these values for effect presentation sizing and weapon HUD cooldown display.

Server-only torpedo damage and radial timing remain in `constants.server.weapons.torpedo`.

## Failure modes

Known constants pipeline failure modes include:

```text
stale generated output after TOML edits
missing managed block for a new TOML section
managed block references a TOML section that does not exist
duplicate constants section across multiple TOML files
unsupported TOML value type
invalid constant key name
manual edits inside generated blocks
regenerating only one language for shared constants that must stay aligned
adding semantically duplicate sections without a destination block or consumer
expecting data-sync to create generated blocks automatically
```

For shared constants, the most important operational failure mode is partial regeneration. Shared world and shared weapon constants should usually be checked and pushed for both Go and GDScript together.

## Code or source map

### Source data

* `shared/constants/server_constants.toml`
* `shared/constants/server_entities.toml`
* `shared/constants/weapons.toml`
* `shared/constants/pickups.toml`
* `shared/constants/weapon_pickups.toml`
* `shared/constants/client/presentation.toml`
* `shared/constants/client/shell.toml`
* `shared/constants/client/lobby.toml`

### Configuration and entrypoint

* `tools/data_sync/config.toml`
* `tools/data_sync/main.py`
* `tools/data_sync/data_sync/cli.py`
* `tools/data_sync/data_sync/config.py`

### Constants pipeline implementation

* `tools/data_sync/data_sync/constants_store.py`
* `tools/data_sync/data_sync/constants_sync.py`
* `tools/data_sync/data_sync/discovery.py`
* `tools/data_sync/data_sync/block_io.py`
* `tools/data_sync/data_sync/pull.py`
* `tools/data_sync/data_sync/validate.py`
* `tools/data_sync/data_sync/toml_store.py`
* `tools/data_sync/data_sync/model/constants.py`

### Generators and parsers

* `tools/data_sync/data_sync/generators/go_constants.py`
* `tools/data_sync/data_sync/generators/gds_constants.py`
* `tools/data_sync/data_sync/generators/ts_constants.py`
* `tools/data_sync/data_sync/parsers/go_constants.py`
* `tools/data_sync/data_sync/parsers/gds_constants.py`
* `tools/data_sync/data_sync/parsers/ts_constants.py`

### Generated outputs

* `services/game-server/internal/constants/constants.go`
* `services/game-server/internal/constants/weapons.go`
* `services/game-server/internal/constants/powerups.go`
* `services/game-server/internal/constants/weapon_pickups.go`
* `client/scripts/generated/constants/constants.gd`

### Key server consumers

* `services/game-server/internal/game/space/space.go`
* `services/game-server/internal/game/weapons/profiles.go`
* `services/game-server/internal/game/entities/pickups/definitions.go`
* `services/game-server/internal/game/runtime/`
* `services/game-server/internal/game/spawning/`
* `services/game-server/internal/game/scoring/`
* `services/game-server/internal/game/physics/`
* `services/game-server/internal/networking/`

### Key client consumers

* `client/scripts/world/world_wrap.gd`
* `client/scripts/boot/session_network_target.gd`
* `client/scripts/networking/network_client.gd`
* `client/scripts/networking/client_connection_service.gd`
* `client/scripts/gameplay/effects/gameplay_effects.gd`
* `client/scripts/ui/hud/weapon_display_registry.gd`
* `client/scripts/presentation/background/background_flow.gd`
* `client/scripts/world/world_sync.gd`

### Tests

* `tools/data_sync/tests/test_constants_generators.py`
* `tools/data_sync/tests/test_constants_store.py`
* `tools/data_sync/tests/test_constants_sync.py`
* `tools/data_sync/tests/test_constants_pull.py`
* `tools/data_sync/tests/test_discovery.py`
* `tools/data_sync/tests/test_validate.py`
* `tools/data_sync/tests/test_cli.py`
* `tools/data_sync/tests/test_block_io.py`
* `tools/data_sync/tests/test_config.py`
* `tools/data_sync/tests/test_toml_store.py`

## Validation commands

Standard constants verification:

```bash
data-sync -validate -constants
data-sync -diff -constants -go -gds
data-sync -check -constants -go -gds
```

Standard constants update:

```bash
data-sync -validate -constants
data-sync -diff -constants -go -gds
data-sync -push -constants -go -gds
data-sync -check -constants -go -gds
```

Focused server-only verification is allowed for server-only constants:

```bash
data-sync -validate -constants
data-sync -diff -constants -go
data-sync -check -constants -go
```

Do not use server-only verification for shared constants that also affect client behavior.

## Related docs

* [Data](./!INDEX.md)
* [Data Sync And SSOT Pipeline](data-sync-and-ssot-pipeline.md)
* [Source Of Truth Map](source-of-truth-map.md)
* [Packet Schemas](packet-schemas.md)
* [Drop Tables](drop-tables.md)
* [Asteroid Variants Data](asteroid-variants-data.md)
* [Collision Shape Data](collision-shape-data.md)
* [Client session boot and network target](../services/client/app-shell-and-session/session-boot-and-network-target.md)
* [Client background and viewport presentation](../services/client/presentation-flow/background-and-viewport-presentation.md)
* [Client entity sync owners](../services/client/world-sync/entity-sync-owners.md)
* [Server toroidal space and motion](../services/game-server/simulation/world/toroidal-space-and-motion.md)
* [Server weapons and projectile fire](../services/game-server/simulation/combat/weapons-and-projectile-fire.md)
* [Server pickup entity lifecycle](../services/game-server/simulation/pickups/pickup-entity-lifecycle.md)

## Notes

The constants pipeline is block-based, not file-rewrite-based. It updates only the content inside matching managed blocks for constants outputs.

Packet files and drop-table files have different generation rules and should not be inferred from constants behavior.

The current local development single-player and multiplayer websocket URLs are identical. They still belong in constants because client session mode selects a named target, and those targets can diverge later without changing the server route model.

The current generated server constants include torpedo weapon-pickup constants from `shared/constants/weapon_pickups.toml`. The torpedo section in `shared/constants/pickups.toml` is not the current emitted torpedo weapon-pickup source.
