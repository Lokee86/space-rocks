# Data Sync

`tools/data_sync/` is a reusable Python CLI for syncing shared game data between:

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
- Go game server files
- GDScript Godot client files
- TypeScript API server files, later

For constants, the tool updates only marked generated blocks. Constants outputs
can be declared on arbitrary top-level language subtables such as
`[constants.go]`, `[weapons.go]`, `[constants.gds]`, or `[weapons.gds]` as long
as they are constants outputs for a supported language and only list
`constants.*` sections. Constants sync is a bidirectional many-source/many-output
pipeline: multiple constants TOML files are supported, multiple generated
constants files per language are supported, `-push` writes source sections to
every configured output target that lists them, and `-pull` reads only owned
generated sections and writes each one back to the TOML file that already
contains it.

Current active scope:

```text
constants -> Go, GDScript, and TypeScript when enabled
packets -> Go and GDScript
drop_tables -> Go only
```

Deferred scope:

```text
TypeScript output
```

## Source Of Truth

The split constants files under `shared/constants/` are the canonical source for active constants.

The canonical sources for active packets are:

- `shared/packets/outputs.toml`
- `shared/packets/gameplay.toml`
- `shared/packets/debug.toml`
- `shared/packets/lobby.toml`

The canonical sources for active drop tables are the TOML files under `shared/drop_tables/`, including `shared/drop_tables/basicasteroids.toml`.

Debug/devtools packet schema lives in `shared/packets/debug.toml`. Data-sync generates server devtools packet types into `services/game-server/internal/devtools/packets_generated.go` through the `server_devtools_packets` output id.

The split constants SoT files under `shared/constants/` contain constants only.
Obsolete packet reference data was removed when the packet TOML pipeline was
adopted. Packet schema changes should be made under `shared/packets/`.
Client constants use nested subcategory sections under
`constants.client.presentation.*`, `constants.client.shell.*`, and
`constants.client.lobby.*`.

New constants and packet schema changes should be made in TOML. Language files are generated from TOML through `-push`.

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
configured `data-sync` blocks. Every selected constants language processes all
configured constants outputs for that language. Packets rewrite configured
generated packet files. Drop tables generate the server Go file only.

`-diff` does the same generation as `-push`, prints a unified diff, and writes
nothing.

`-check` writes nothing and exits `0` when generated blocks are current, or `1` when files differ.

`-validate` checks config, TOML integrity, supported values/types, ownership rules, configured file existence, and required managed blocks.

`-pull` is intentionally restricted. Constants pull reads owned generated blocks from all constants outputs for the selected language, updates existing TOML values only, and writes each section back to the SoT file that already contains it.

Pull fails if a source section is missing from all TOML files, if a source section appears in more than one TOML file, or if a generated section is owned by more than one output target.

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

[constants.go]
files = ["services/game-server/internal/constants/constants.go"]
sections = ["constants.gameplay", "constants.network"]
owns = ["constants.gameplay", "constants.network"]

[constants.gds]
files = ["client/scripts/generated/constants/constants.gd"]
sections = ["constants.gameplay", "constants.client"]
owns = ["constants.client"]

[constants.ts]
enabled = false
files = []
sections = []
owns = []

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

Constants and packets have separate SoT paths. `-constants` commands read/write only the constants SoT, and `-packets` commands read/write only the packet SoT files.
Drop tables have their own SoT path set under `shared/drop_tables/`, and `-drop-tables -go` reads and writes only the server Go output.

`sections` controls what a language receives during `-push`, `-diff`, and `-check`.

`owns` controls what a language may update during `-pull`.

Constants ownership overlap is invalid per section. Packet ownership is coarse for now; packet-level ownership may be added later.

Example constants layout:

```toml
[sot.constants]
paths = [
  "shared/constants/server_constants.toml",
  "shared/constants/weapons.toml",
]

[constants.go]
files = ["services/game-server/internal/constants/constants.go"]
sections = ["constants.gameplay"]
owns = ["constants.gameplay"]

[weapons.go]
files = ["services/game-server/internal/constants/weapons.go"]
sections = ["constants.server.weapons.basic_cannon", "constants.server.weapons.torpedo"]
owns = ["constants.server.weapons.basic_cannon", "constants.server.weapons.torpedo"]
```

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

[weapons.go]
files = ["services/game-server/internal/constants/weapons.go"]
sections = ["constants.server.weapons.basic_cannon", "constants.server.weapons.torpedo"]
owns = ["constants.server.weapons.basic_cannon", "constants.server.weapons.torpedo"]

[weapons.gds]
files = ["client/scripts/generated/constants/weapons.gd"]
sections = ["constants.server.weapons.basic_cannon"]
owns = ["constants.server.weapons.basic_cannon"]
```

Example pull layout:

```toml
[sot]
paths = [
  "shared/constants/server_constants.toml",
  "shared/constants/weapons.toml",
]

[constants.go]
files = ["services/game-server/internal/constants/constants.go"]
sections = ["constants.gameplay"]
owns = ["constants.gameplay"]

[weapons.go]
files = ["services/game-server/internal/constants/weapons.go"]
sections = ["constants.server.weapons.basic_cannon", "constants.server.weapons.torpedo"]
owns = ["constants.server.weapons.basic_cannon", "constants.server.weapons.torpedo"]
```

Packets:

```toml
[[outputs]]
language = "go"
path = "services/game-server/internal/game/packets.go"
package = "game"
packet_types = true
structs = ["ClientPacket", "EventState", "StatePacket"]

[outputs.imports]
runtime = "github.com/Lokee86/space-rocks/server/internal/game/runtime"

[[structs]]
id = "StatePacket"

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
id = "state"
value = "state"

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
