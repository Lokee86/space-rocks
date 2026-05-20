# Data Sync

`tools/data_sync/` is a reusable Python CLI for syncing shared game data between:

- TOML source of truth for active constants: `shared/game_data.toml`
- Go game server files
- GDScript Godot client files
- TypeScript API server files, later

The tool updates only marked generated blocks. It never rewrites whole source files.

Current active scope:

```text
constants -> Go and GDScript
```

Deferred scope:

```text
packets -> Go/GDScript
TypeScript output
```

## Source Of Truth

`shared/game_data.toml` is the canonical source for active constants. It also contains packet data copied during migration for future work, but packet sync is disabled in the default config. TOML is used because it is readable for hand-edited values, supports ordered sections, and can preserve practical round-trip formatting through `tomlkit`.

New constants should be made in TOML. Language files are generated from TOML through `-push`.

Packet schema changes should still be made in `shared/packets/packets.json` and regenerated with `tools/scripts/generate_packets.py`.

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
python tools/data_sync/main.py -validate
python tools/data_sync/main.py -validate -constants
```

`-push`, `-pull`, `-diff`, and `-check` require at least one domain and one language. `-pull` accepts only one language at a time.

## Operation Behavior

`-push` reads TOML, generates canonical language output, and replaces configured `data-sync` blocks.

`-diff` does the same generation as `-push`, prints a unified diff, and writes nothing.

`-check` writes nothing and exits `0` when generated blocks are current, or `1` when files differ.

`-validate` checks config, TOML integrity, supported values/types, ownership rules, configured file existence, and required managed blocks.

`-pull` is intentionally restricted. Constants pull reads owned generated blocks and updates existing TOML values only.

Packet sync and TypeScript output are disabled in the default config.

## Config Format

Default config:

```text
tools/data_sync/config.toml
```

Shape:

```toml
[sot]
path = "shared/game_data.toml"

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
enabled = false
files = []
sections = []
owns = []
```

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
[packets.player_input]
id = 100
direction = "client_to_server"

[packets.player_input.fields]
sequence = "uint32"
turn = "float32"
thrust = "bool"
shoot = "bool"
```

Packet directions:

```text
client_to_server
server_to_client
bidirectional
```

Supported packet field types:

```text
bool
int
uint32
float32
float64
string
```

## Generated Blocks

Go and TypeScript markers:

```go
// data-sync:start constants.gameplay
// data-sync:end constants.gameplay
```

GDScript markers:

```gdscript
# data-sync:start packets
# data-sync:end packets
```

Only content between matching markers is replaced. Missing or duplicate markers are hard failures.

## Formatting Policy

Generated block content is canonical and deterministic. The tool does not preserve old formatting inside generated blocks.

For pull, parsers are strict and accept only canonical generated constants. Added, removed, renamed, reordered, or non-canonical constants are rejected.

## Packet Pull Policy

Full packet schema pull is not supported. Packet schema changes should be edited in `shared/packets/packets.json`, then regenerated through the existing packet generator.

`-pull -packets ...` returns a clear refusal instead of attempting fragile packet parsing.

## JSON Migration

A disposable migration script seeded the TOML source from the old JSON constants source and the active packet JSON source. The old constants JSON source has been retired, so the script is kept only as historical migration scaffolding.

The migration produced:

```text
shared/game_data.toml
```

## Active Constants Workflow

1. Edit `shared/game_data.toml`.
2. Run `python tools/data_sync/main.py -validate -constants`.
3. Run `python tools/data_sync/main.py -diff -constants -go -gds`.
4. Run `python tools/data_sync/main.py -push -constants -go -gds`.
5. Run `python tools/data_sync/main.py -check -constants -go -gds`.
