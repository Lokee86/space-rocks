# Drop Table System

## Status Summary

- `basicasteroids` exists.
- Drop tables are server-authoritative.
- The `drops` package owns evaluation only.
- Root `internal/game` owns spawning the pickup result.
- Pickup collection and effects stay in the existing pickup seam.
- `drop_mode` is explicit in TOML and generated output.
- Single-drop behavior is preserved for `basicasteroids`.
- Multi-drop support is entry-roll capped, not slot-roll.
- There is no minimum drop count yet.
- Drop tables are separate from constants and packets.

## Ownership

- `shared/drop_tables/*.toml` owns the source-of-truth drop table data.
- `tools/data_sync` owns parsing, validation, and Go generation.
- `internal/game/drops` owns table types and roll evaluation.
- Root `internal/game` owns pickup spawning after a successful drop result.
- `internal/game/pickups` continues to own pickup collection and effect resolution.
- Client gameplay only observes the spawned pickup through the existing state packet path.

## Data Sources

- `shared/drop_tables/basicasteroids.toml`
- `services/game-server/internal/game/drops/drop_tables.go`
- `services/game-server/internal/game/drops/roll.go`
- `services/game-server/internal/game/drops/table.go`
- `shared/packets/gameplay.toml`

## Generated Output

- `services/game-server/internal/game/drops/drop_tables.go`
- `GeneratedTables` is the generated runtime table map.
- The generated file is server Go only.
- Drop tables are not generated into constants outputs.
- Drop tables are not generated into client outputs.

## Server Runtime Model

- `DropMode` is a runtime table field.
- `DropModeSingle` means one successful result at most.
- `DropModeMulti` means multiple successful results are possible.
- `MaxDropsPerSource` is the per-source cap for a table.
- `MaxActivePickups` is the active world pickup cap for the table.
- `Entry` fields remain `pickup_type`, `chance`, `min_source_size`, and `max_source_size`.
- `Roll` carries a sequence of roll values.
- `Result` carries `table_id`, `pickup_type`, `x`, and `y`.

## Drop Evaluation Model

- The `drops` package evaluates a table ID against a source and roll values.
- Evaluation is deterministic for a given roll value sequence.
- Evaluation does not spawn pickups.
- Evaluation does not mutate the game state.
- Evaluation only returns result data for the caller to act on.
- Missing tables return no results.
- Source type mismatches return no results.
- Entries outside source size bounds are skipped.
- A result succeeds when the roll value is below the entry chance.

## Current Basicastroids Table

- Table ID: `basicasteroids`
- Source type: `asteroid`
- Drop mode: `single`
- Max drops per source: `1`
- Max active pickups: `2`
- Entries:
- `pickup_type = "1_up"`
- `chance = 0.01`
- `min_source_size = 1`
- `max_source_size = 4`

## Asteroid Destruction Flow

- Asteroid destruction consequences already run inside root `internal/game`.
- The destruction consequence path calls the drop helper while the game lock is held.
- The drop helper checks the `basicasteroids` table and active pickup cap.
- The helper builds the source from asteroid ID, size, and position.
- The helper rolls the table and spawns pickups from any returned results.
- The helper records the pickup dropped event after each successful spawn.

## Event Semantics

- `pickup_dropped` means a pickup was successfully created from a drop table result.
- The event includes `pickup_id`, `pickup_type`, `source_type`, `source_id`, `table_id`, `x`, and `y`.
- Drop events are separate from pickup collection and pickup effect events.
- `pickup_expired` belongs to pickup lifecycle, not drop-table evaluation.
- `pickup_collected` and `pickup_effect_applied` remain owned by the pickup seam.

## Relationship To Pickups

- Drop tables create pickups, but they do not own pickup collection.
- Drop tables do not own pickup effects.
- Pickup collection still resolves through `internal/game/pickups`.
- Pickup effect application still mutates player sessions through the pickup seam.
- The drop seam ends once the authoritative pickup exists.
- Pickup expiry is owned by the pickup lifecycle, not the drop seam.

## Single vs Multi Drop Behavior

- `DropModeSingle` returns zero or one result.
- `DropModeSingle` returns the first successful matching entry only.
- `DropModeMulti` can return multiple successful results.
- `DropModeMulti` consumes roll values in entry order.
- `DropModeMulti` stops at `MaxDropsPerSource`.
- Missing roll values are treated as failed rolls.
- Multi mode is entry-roll capped multi-drop, not slot-roll.

## Adding A New Drop Table

- Add a TOML file under `shared/drop_tables/`.
- Set `table.id` explicitly.
- Set `source_type` explicitly.
- Set `drop_mode` explicitly.
- Set `max_drops_per_source` explicitly.
- Set `max_active_pickups` explicitly.
- Add one or more `[[entries]]` rows.
- Regenerate or update `services/game-server/internal/game/drops/drop_tables.go` through data sync.
- Add or update focused parser, generator, and runtime tests.

## Testing And Verification

- `python tools/data_sync/main.py -validate -drop-tables`
- `python tools/data_sync/main.py -diff -drop-tables -go`
- `python tools/data_sync/main.py -push -drop-tables -go`
- `python tools/data_sync/main.py -check -drop-tables -go`
- `go test ./...` in `services/game-server`
- `Select-String -Path docs,services,tools,shared -Pattern 'pickup_dropped'`
- `Select-String -Path docs,services,tools,shared -Pattern 'drop_mode'`
- `Select-String -Path docs,services,tools,shared -Pattern 'MaxDropsPerSource'`

## Future Work

- Multi-drop tables with more than one table entry.
- Additional drop table definitions for other source types.
- Minimum drop count policy, if ever needed.
- More explicit per-source source-type routing.
- Client-facing presentation polish for drop events, if needed.
