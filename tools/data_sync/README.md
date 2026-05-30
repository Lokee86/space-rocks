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
- Go game server files
- GDScript Godot client files
- TypeScript API server files, later

For constants, the tool updates only marked generated blocks. For packet files,
the current generated outputs are fully generated files, so packet push rewrites
the configured packet files as whole files.

Current active scope:

```text
constants -> Go and GDScript
packets -> Go and GDScript
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

The split constants SoT files under `shared/constants/` contain constants only. Obsolete packet reference data was removed when the packet TOML pipeline was adopted. Packet schema changes should be made under `shared/packets/`.
Client constants use nested subcategory sections under `constants.client.presentation.*`, `constants.client.shell.*`, and `constants.client.lobby.*`.

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
python tools/data_sync/main.py -push -constants -go
python tools/data_sync/main.py -push -constants -go -gds
python tools/data_sync/main.py -pull -constants -go
python tools/data_sync/main.py -diff -constants -go -gds
python tools/data_sync/main.py -check -constants -go -gds
python tools/data_sync/main.py -validate -packets
python tools/data_sync/main.py -diff -packets -go -gds
python tools/data_sync/main.py -push -packets -go -gds
python tools/data_sync/main.py -check -packets -go -gds
python tools/data_sync/main.py -validate
python tools/data_sync/main.py -validate -constants
```

`-push`, `-pull`, `-diff`, and `-check` require at least one domain and one language. `-pull` accepts only one language at a time.

## Operation Behavior

`-push` reads TOML and generates canonical language output. Constants replace configured `data-sync` blocks. Packets rewrite configured generated packet files.

`-diff` does the same generation as `-push`, prints a unified diff, and writes nothing.

`-check` writes nothing and exits `0` when generated blocks are current, or `1` when files differ.

`-validate` checks config, TOML integrity, supported values/types, ownership rules, configured file existence, and required managed blocks.

`-pull` is intentionally restricted. Constants pull reads owned generated blocks and updates existing TOML values only.

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

[constants.go]
files = ["services/game-server/internal/game/constants.go"]
sections = ["constants.gameplay", "constants.network"]
owns = ["constants.gameplay", "constants.network"]

[constants.gds]
files = ["client/scripts/constants.gd"]
sections = ["constants.gameplay", "constants.client"]
owns = ["constants.client"]

[constants.ts]
enabled = false
files = []
sections = []
owns = []

[packets.go]
files = [
  "services/game-server/internal/game/entities/packets_generated.go",
  "services/game-server/internal/game/packets.go",
]
sections = ["packets"]
owns = []

[packets.gds]
files = ["client/scripts/packets.gd"]
sections = ["packets"]
owns = []
```

Constants and packets have separate SoT paths. `-constants` commands read/write only the constants SoT, and `-packets` commands read/write only the packet SoT files.

`sections` controls what a language receives during `-push`, `-diff`, and `-check`.

`owns` controls what a language may update during `-pull`.

Constants ownership overlap is invalid per section. Packet ownership is coarse for now; packet-level ownership may be added later.

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
structs = ["ClientPacket", "EventState", "StatePacket"]

[outputs.imports]
entities = "github.com/Lokee86/space-rocks/server/internal/game/entities"

[[structs]]
id = "StatePacket"

[[structs.fields]]
name = "players"
json = "players"
type = "map"
key_type = "string"
value_type = "ShipState"
go_value_type = "entities.ShipState"

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
```

## Active Constants Workflow

1. Edit the needed constants SoT file under `shared/constants/` (`server_constants.toml`, `server_entities.toml`, `client/presentation.toml`, `client/shell.toml`, or `client/lobby.toml`).
2. Run `python tools/data_sync/main.py -validate -constants`.
3. Run `python tools/data_sync/main.py -diff -constants -go -gds`.
4. Run `python tools/data_sync/main.py -push -constants -go -gds`.
5. Run `python tools/data_sync/main.py -check -constants -go -gds`.

## Active Packet Workflow

1. Edit packet schema files under `shared/packets/` (`outputs.toml`, `gameplay.toml`, `debug.toml`, and `lobby.toml`).
2. Run `python tools/data_sync/main.py -validate -packets`.
3. Run `python tools/data_sync/main.py -diff -packets -go -gds`.
4. Review the diff.
5. Run `python tools/data_sync/main.py -push -packets -go -gds`.
6. Run `python tools/data_sync/main.py -check -packets -go -gds`.
