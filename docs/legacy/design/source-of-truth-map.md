# Source Of Truth Map

Parent index: [Design Legacy](./!INDEX.md)

This doc maps where truth lives, what consumes it, how drift is enforced, and what each source does not own.

For a compact player-data routing reference, see [Player-Data Routing Reference](player-data-routing.md).

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
| Asteroid variants | `shared/asteroids/variants.toml` | `client/scripts/generated/asteroids/asteroid_variants.gd`, `services/game-server/internal/game/asteroids/variants.go`; see [Asteroid Variant Contract](asteroid-variants.md) | `go test ./internal/game/asteroids ./...` and GUT coverage for `client/tests/unit/entities/test_asteroid_variants.gd` | Active | Collision shape geometry, packet field names, and asteroid size/health mechanics |
| WebSocket packets | `shared/packets/*.toml` | `client/scripts/generated/networking/packets/packets.gd`, `services/game-server/internal/game/packets.go`, `services/game-server/internal/game/runtime/packets_generated.go`, `services/game-server/internal/devtools/packets_generated.go` | `python3 tools/data_sync/main.py -check -packets -go -gds` | Active | Constants, drop tables, database schemas, logical player-data schema, HTTP contracts |
| Drop tables | `shared/drop_tables/*.toml` | `services/game-server/internal/game/drops/drop_tables.go` | `python3 tools/data_sync/main.py -check -drop-tables -go` | Active | Pickup collection, pickup effects, packet schema |
| Player-data logical schema | `shared/player_data/*.toml` | `services/player-data/playerdata/*.go`, `services/api-server/app/controllers/internal/player_data/*` | `go test ./...` in `services/player-data` and `bundle exec rails test test/controllers/internal/player_data` | Active | Rails/Postgres physical schema, embedded DB physical schema, player-data runtime packets, HTTP contracts |
| HTTP API contracts | `shared/contracts/http/openapi.yaml` | Rails controllers and controller tests in `services/api-server/`; game-server-hosted data-handler profile reads for `POST /api/player-data/profile`; RailsStore-backed internal stats reads for `POST /api/internal/player-data/stats` | `bundle exec rails test test/contracts/openapi_contract_test.rb` | Active | Rails migration layout, runtime middleware, data-sync domains, player-data runtime packets |
| Collision shapes | `shared/collisions/collision_shapes.json` | `client/tools/export_collision_shapes.gd` generates it from `client/scenes/*.tscn`; server collision loading consumes it | `godot --headless --path client -s res://tools/export_collision_shapes.gd` | Active | Gameplay rules, packet shapes, database schemas |
| Rails database schema | `services/api-server/db/migrate/*.rb` | Rails/Postgres physical schema and `schema.rb` | `bundle exec rails db:migrate` and `bundle exec rails test` | Active | HTTP contract source, gameplay state, client scenes |
| Godot scenes/node structure | `client/scenes/*.tscn` | Godot client runtime and scene/script consumers under `client/scripts/` | `godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit` | Active | Server simulation, packet schema, database schema |
## Ownership Rules

- `shared/constants/*.toml` owns gameplay, client, and server constants.
- `shared/asteroids/variants.toml` owns asteroid variant metadata. Variant indexes are zero-based runtime values, ids like `asteroid_1` are stable labels, and spawn weights are float values owned by the variant contract.
- `constants.AsteroidVariants` must not be reintroduced; the asteroid variant catalog owns the list and count.
- `SINGLE_PLAYER_WS_URL` and `MULTIPLAYER_WS_URL` are client target URLs owned by `shared/constants/*.toml`, not server route definitions.
- `shared/packets/*.toml` owns WebSocket packet shapes.
- `shared/drop_tables/*.toml` owns drop-table definitions.
- `shared/player_data/*.toml` owns logical player-data schema only.
- `shared/packets/player_data.toml` owns the player-data runtime packet protocol.
- `shared/contracts/http/openapi.yaml` owns HTTP request and response contracts, including `POST /api/player-data/profile`, `POST /api/internal/player-data/stats`, and the other Rails HTTP request/response shapes.
- The Go server WebSocket route remains `/ws`; client target selection belongs to client boot/session networking code, not scene/menu scripts.
- The embedded SQLite local-profile store exists in the standard no-tag development build; `-tags noembeddedsqlite` deployment/restricted builds omit that package and its `modernc.org/sqlite` dependency.
- Rails migrations own the Rails/Postgres physical schema.
- HTTP OpenAPI is not generated by `tools/data_sync`.
- Rails controllers implement HTTP contracts but are not generated from them.
- Generated files are not hand-edit sources.

## Player-data Routing Rule

- Gameplay and write facts enter through the player-data runtime.
- Profile and read requests enter through the HTTP data-handler facade.
- Clients do not choose Rails, SQLite, or guest memory directly.
- Guest transient stats stay separate from local-profile persistence, and authenticated account stats remain Rails-backed.
- Stores are selected by identity kind after mode and identity validation.

## Documentation Ownership

- `docs/design` describes implemented behavior, current contracts, ownership, and source-of-truth boundaries.
- `docs/limits` describes factual current limitations, unavailable features, incomplete integrations, hardcoded fallbacks, and current constraints.
- `docs/planning` describes future implementation plans, roadmap phases, unresolved design, and backlog direction.
- `docs/design` should not keep detailed current-limit bullets.
- `docs/design` may include a short `Related Limits` link when a limitation is important to prevent misuse.
- Detailed limit wording belongs in `docs/limits`.
- Future implementation steps, remaining-work sections, roadmap phases, and speculative design belong in `docs/planning`.