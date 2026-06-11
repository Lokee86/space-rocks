# Source Of Truth Map

This doc maps where truth lives, what consumes it, how drift is enforced, and what each source does not own.

## Definitions

- Source of Truth: the authoritative editable input that owns a contract or data shape.
- Generated Output: a file produced from a source of truth and not hand-edited.
- Implemented Contract: a live consumer or adapter that fulfills the source contract at runtime.
- Logical Schema: the meaning and shape of data, independent of storage layout.
- Physical Schema: the concrete storage layout used by a database or file format.
- Enforcement: the command or test path that detects drift between source and consumer.

## Summary

| Domain | Source of Truth | Generated Output or Implemented Contract | Enforcement | Status | Does not own |
| --- | --- | --- | --- | --- | --- |
| Gameplay constants | `shared/constants/*.toml` | `client/scripts/generated/constants/constants.gd`, `services/game-server/internal/constants/constants.go` | `python3 tools/data_sync/main.py -check -constants -go -gds` | Active | Packet shapes, drop tables, physical schemas |
| WebSocket packets | `shared/packets/*.toml` | `client/scripts/generated/networking/packets/packets.gd`, `services/game-server/internal/game/packets.go`, `services/game-server/internal/game/runtime/packets_generated.go`, `services/game-server/internal/devtools/packets_generated.go` | `python3 tools/data_sync/main.py -check -packets -go -gds` | Active | Constants, drop tables, database schemas |
| Drop tables | `shared/drop_tables/*.toml` | `services/game-server/internal/game/drops/drop_tables.go` | `python3 tools/data_sync/main.py -check -drop-tables -go` | Active | Pickup collection, pickup effects, packet schema |
| Player-data logical schema | `shared/player_data/*.toml` | `services/player-data/playerdata/*.go`, `services/api-server/app/controllers/internal/player_data/*` | `go test ./...` in `services/player-data` and `bundle exec rails test test/controllers/internal/player_data` | Active | Rails/Postgres physical schema, embedded DB physical schema |
| HTTP API contracts | `shared/contracts/http/openapi.yaml` | Rails controllers and controller tests in `services/api-server/` | `bundle exec rails test test/contracts/openapi_contract_test.rb` | Active | Rails migration layout, runtime middleware, data-sync domains |
| Collision shapes | `shared/collisions/collision_shapes.json` | `client/tools/export_collision_shapes.gd` generates it from `client/scenes/*.tscn`; server collision loading consumes it | `godot --headless --path client -s res://tools/export_collision_shapes.gd` | Active | Gameplay rules, packet shapes, database schemas |
| Rails database schema | `services/api-server/db/migrate/*.rb` | Rails/Postgres physical schema and `schema.rb` | `bundle exec rails db:migrate` and `bundle exec rails test` | Active | HTTP contract source, gameplay state, client scenes |
| Godot scenes/node structure | `client/scenes/*.tscn` | Godot client runtime and scene/script consumers under `client/scripts/` | `godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit` | Active | Server simulation, packet schema, database schema |
## Ownership Rules

- `shared/constants/*.toml` owns gameplay, client, and server constants.
- `shared/packets/*.toml` owns WebSocket packet shapes.
- `shared/drop_tables/*.toml` owns drop-table definitions.
- `shared/player_data/*.toml` owns logical player-data schema only.
- `shared/contracts/http/openapi.yaml` owns HTTP request and response contracts.
- Rails migrations own the Rails/Postgres physical schema.
- HTTP OpenAPI is not part of `tools/data_sync` yet.
- Rails controllers implement HTTP contracts but are not generated from them.
- Generated files are not hand-edit sources.
