# Data Sync

`tools/data_sync/` is a reusable Python CLI for syncing data-sync-supported shared game data.

## Scope

`tools/data_sync/` currently owns workflow support for:

- constants
- packets
- drop_tables

`tools/data_sync/` does not currently own:

- HTTP OpenAPI contracts
- Rails/Postgres migrations
- Godot scene/node structure
- collision export source scenes/assets

`player_data` is a logical-schema domain with partial/planned pipeline support. The authoritative overview for project-wide ownership lives in [source-of-truth-map](../../docs/design/source-of-truth-map.md).

HTTP contracts are separate from data-sync and are documented elsewhere.

This README describes data-sync-supported domains only, not every project source of truth.

`tools/data_sync/` works between:

- TOML sources of truth for active constants:
  - `shared/constants/server_constants.toml`
  - `shared/constants/server_entities.toml`
  - `shared/constants/client/presentation.toml`
  - `shared/constants/client/shell.toml`
  - `shared/constants/client/lobby.toml`
- TOML sources of truth for active packets:
  - `shared/packets/outputs.toml`
  - `shared/packets/gameplay.toml`
  - `shared/packets/debug.toml`
  - `shared/packets/lobby.toml`
- TOML sources of truth for active drop tables:
  - `shared/drop_tables/*.toml`
- Planned future TOML sources of truth for player-data schema:
  - `shared/player_data/*.toml`
- Go game server files
- GDScript Godot client files
- TypeScript API server files, later, when enabled

For constants, the tool uses `data-sync` destination blocks discovered through
`[constants.scan]`. `-push` maps each TOML `constants.*` section to matching
destination blocks, `-pull` maps destination blocks back to the matching TOML
section, and the section name is the routing contract. No constants
files/sections/owns config is required. Multiple constants TOML files are
supported, each constants section must exist in exactly one source TOML file,
and duplicate pull blocks must parse to identical values or pull fails.

Current active data-sync scope:

```text
constants -> Go, GDScript, and TypeScript when enabled
packets -> Go and GDScript
drop_tables -> Go only
```

Deferred data-sync scope:

```text
TypeScript output
player_data logical schema domain
migration skeleton generation
```

## Future Player-Data Schema Domain

Player-data schema is a logical schema SSoT, not raw database DDL, and the pipeline support here is still partial/planned.

The likely future domain flag is `-player-data`, but it is not implemented yet.

Future outputs may include Go structs/contracts, schema docs, contract fixtures, Rails migration skeletons, and embedded DB migration skeletons.

See the broader [source-of-truth map](../../docs/design/source-of-truth-map.md) and [player-data schema source of truth](../../docs/design/player-data-schema-ssot.md).

## Supported Source Of Truth

The split constants files under `shared/constants/` are the canonical source for data-sync-supported active constants.

The canonical sources for data-sync-supported active packets are:

- `shared/packets/outputs.toml`
- `shared/packets/gameplay.toml`
- `shared/packets/debug.toml`
- `shared/packets/lobby.toml`

The canonical sources for data-sync-supported active drop tables are the TOML files under `shared/drop_tables/`, including `shared/drop_tables/basicasteroids.toml`.

Debug/devtools packet schema lives in `shared/packets/debug.toml`. Data-sync generates server devtools packet types into `services/game-server/internal/devtools/packets_generated.go` through the `server_devtools_packets` output id.

The split constants SoT files under `shared/constants/` contain constants only.
Obsolete packet reference data was removed when the packet TOML pipeline was
adopted. Packet schema changes should be made under `shared/packets/`.
Client constants use nested subcategory sections under
`constants.client.presentation.*`, `constants.client.shell.*`, and
`constants.client.lobby.*`.

New constants and packet schema changes should be made in TOML. Active gameplay packet output is lane-native, and language files are generated from TOML through `-push`.

## Commands

Exactly one operation is required:

```bash
-push
-pull
-diff
-check
-validate
```

Domains:

```bash
-constants
-packets
-drop-tables
```

Languages:

```bash
-go
-gds
-ts
```

Options:

```bash
-config <path>
-sot <path>
```

Examples:

```bash
data-sync -push -constants -go
data-sync -push -constants -go -gds
data-sync -pull -constants -go
data-sync -diff -constants -go -gds
data-sync -check -constants -go -gds
data-sync -validate -packets
data-sync -diff -packets -go -gds
data-sync -push -packets -go -gds
data-sync -check -packets -go -gds
data-sync -push -drop-tables -go
data-sync -diff -drop-tables -go
data-sync -check -drop-tables -go
data-sync -validate
data-sync -validate -constants
```

`-push`, `-pull`, `-diff`, and `-check` require at least one domain and one language. `-pull` accepts only one language at a time.
`-constants` does not generate drop tables.

## Operation Behavior

`-push` reads TOML and generates canonical language output. Constants replace
matching discovered `data-sync` blocks by section name. Packets rewrite
configured generated packet files. Drop tables generate the server Go file
only.

`-diff` does the same generation as `-push`, prints a unified diff, and writes
nothing.

`-check` writes nothing and exits `0` when generated blocks are current, or `1` when files differ.

`-validate` checks config, TOML integrity, supported values/types, configured
file existence, and required managed blocks.

`-pull` is intentionally restricted. Constants pull reads discovered generated
blocks for the selected language, updates existing TOML values only, and writes
each section back to the SoT file that already contains it.

Pull fails if a source section is missing from all TOML files, if a source
section appears in more than one TOML file, or if duplicate discovered blocks
for one section disagree.

TypeScript output is disabled in the default config.

## Config Format

Default config:

```text
tools/data_sync/config.toml
```

Shape:

```toml
[sot.constants]
paths = [
  "shared/constants/server_constants.toml",
  "shared/constants/server_entities.toml",
  "shared/constants/client/presentation.toml",
  "shared/constants/client/shell.toml",
  "shared/constants/client/lobby.toml",
]

[sot.packets]
paths = [
  "shared/packets/outputs.toml",
  "shared/packets/gameplay.toml",
  "shared/packets/debug.toml",
  "shared/packets/lobby.toml",
]

[sot.drop_tables]
paths = [
  "shared/drop_tables/basicasteroids.toml",
]

[constants.scan]
include = ["services/**/*.go", "client/**/*.gd", "services/**/*.ts"]
exclude = [".git/**", "**/.godot/**", "**/node_modules/**"]

[packets.go]
files = [
  "services/game-server/internal/game/runtime/packets_generated.go",
  "services/game-server/internal/game/packets.go",
  "services/game-server/internal/devtools/packets_generated.go",
]
sections = ["packets"]
owns = []
outputs = ["server_entities_packets", "server_game_packets", "server_devtools_packets"]

[packets.gds]
files = ["client/scripts/generated/networking/packets/packets.gd"]
sections = ["packets"]
owns = []

[drop_tables.go]
files = ["services/game-server/internal/game/drops/drop_tables.go"]
sections = []
owns = []
outputs = ["server_drop_tables"]
```

Constants and packets have separate SoT paths. `-constants` commands read/write
only the constants SoT, and `-packets` commands read/write only the packet SoT
files.
Drop tables have their own SoT path set under `shared/drop_tables/`, and `-drop-tables -go` reads and writes only the server Go output.

## TOML Format

Constants:

```toml
[constants.gameplay]
player_speed = 420.0
bullet_speed = 900.0
asteroid_spawn_interval = 1.5

[constants.network]
tick_rate = 60
max_players_per_room = 2

```

Packets:

```toml
[[outputs]]
language = "go"
path = "services/game-server/internal/game/packets.go"
package = "game"
packet_types = true
structs = ["ClientPacket", "EventState", "WorldFullPacket"]

[outputs.imports]
runtime = "github.com/Lokee86/space-rocks/server/internal/game/runtime"

[[structs]]
id = "WorldFullPacket"

[[structs.fields]]
name = "players"
json = "players"
type = "map"
key_type = "string"
value_type = "ShipState"
go_value_type = "runtime.ShipState"

[[structs.fields]]
name = "events"
json = "events"
type = "array"
item_type = "EventState"

[[packet_types]]
id = "world_full"
value = "world_full"

[[builders]]
id = "input_packet"
args = ["forward", "back", "right", "left", "shoot"]

[builders.body]
type = "input"
```

`packet_type_ids` on an `[[outputs]]` entry restricts which packet type constants that output renders. If `packet_type_ids` is omitted, outputs that render packet types keep legacy behavior and render all schema packet types. When present, the `packet_type_ids` order controls generated constant order.

The packet schema preserves the old rich JSON behavior:

```text
outputs       generated file targets, language, package/base, imports, selected structs/builders
structs       Go/GDScript packet/state shapes and field metadata
packet_types  packet type constant names and values
builders      GDScript packet builder functions and argument references
```

Supported field shapes include primitives, arrays, maps, custom struct references, Go type overrides, and rich type strings where needed:

```text
bool
int
float
string
map<string,ShipState>
array<EventState>
```

## Generated Blocks

Go and TypeScript markers:

```go
// data-sync:start constants.gameplay
// data-sync:end constants.gameplay
```

GDScript markers:

```gdscript
# data-sync:start constants.client.presentation.background
# data-sync:end constants.client.presentation.background
```

Only content between matching markers is replaced for constants. Missing or duplicate markers are hard failures.

Packet files are fully generated outputs and do not require data-sync block markers.

## Formatting Policy

Generated block content is canonical and deterministic. The tool does not preserve old formatting inside generated blocks.

For pull, parsers are strict and accept only canonical generated constants. Added, removed, renamed, reordered, or non-canonical constants are rejected.

## Packet Pull Policy

Full packet schema pull is not supported. Packet schema changes should be edited under `shared/packets/`, then pushed from TOML.

`-pull -packets ...` returns a clear refusal instead of attempting fragile packet parsing.

## JSON Migration

Disposable migration scripts seeded TOML from the old JSON sources. The old constants and packet JSON sources have been retired.

The active TOML sources are:

```text
shared/constants/server_constants.toml
shared/constants/server_entities.toml
shared/constants/client/presentation.toml
shared/constants/client/shell.toml
shared/constants/client/lobby.toml
shared/packets/outputs.toml
shared/packets/gameplay.toml
shared/packets/debug.toml
shared/packets/lobby.toml
shared/drop_tables/basicasteroids.toml
```

## Active Constants Workflow

1. Edit the needed constants SoT file under `shared/constants/` (`server_constants.toml`, `server_entities.toml`, `client/presentation.toml`, `client/shell.toml`, or `client/lobby.toml`).
2. Run `data-sync -validate -constants`.
3. Run `data-sync -diff -constants -go -gds`.
4. Run `data-sync -push -constants -go -gds`.
5. Run `data-sync -check -constants -go -gds`.

## Active Packet Workflow

1. Edit packet schema files under `shared/packets/` (`outputs.toml`, `gameplay.toml`, `debug.toml`, and `lobby.toml`).
2. Run `data-sync -validate -packets`.
3. Run `data-sync -diff -packets -go -gds`.
4. Review the diff.
5. Run `data-sync -push -packets -go -gds`.
6. Run `data-sync -check -packets -go -gds`.

## Active Drop Table Workflow

1. Edit the drop table TOML files under `shared/drop_tables/`.
   The current baseline drop table source is `shared/drop_tables/basicasteroids.toml`.
2. Run `data-sync -validate -drop-tables`.
3. Run `data-sync -diff -drop-tables -go`.
4. Review the diff.
5. Run `data-sync -push -drop-tables -go`.
6. Run `data-sync -check -drop-tables -go`.
