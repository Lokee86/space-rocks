# Drop Table Pipeline
Parent index: [Data](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe drop-table source files and generated outputs.

## Overview

TODO: summarize how drop-table TOML becomes generated gameplay drop-table code.
Stub note: keep this focused on drop-table flow.

## Source files

- `shared/drop_tables/basicasteroids.toml`
- TODO: add any other confirmed drop-table source files.

## Configuration

- `tools/data_sync/config.toml`
- `tools/data_sync/!README.md`
- TODO: add any other drop-table pipeline configuration roots when they are confirmed.

## Generated outputs

- `services/game-server/internal/game/drops/drop_tables.go`
- TODO: add any other generated drop-table outputs when they are confirmed.

## Consumers

- Game-server drop-table gameplay logic.
- TODO: add any other confirmed consumers.

## Pipeline usage

- `tools/data_sync -push -drop-tables -go`
- `tools/data_sync -check -drop-tables -go`
- `tools/data_sync -validate`
- TODO: add any other verified drop-table pipeline commands.

## Validation commands

- `tools/data_sync -validate`
- `tools/data_sync -check -drop-tables -go`
- `tools/data_sync -diff -drop-tables -go`
- TODO: add any other verified validation commands.

## Failure modes

- Missing or conflicting drop-table source files.
- Stale generated drop-table code after source edits.
- Invalid drop-table schema or unsupported routing configuration.
- TODO: add any other verified failure modes.

## Code or source map

- `shared/drop_tables/`
- `tools/data_sync/data_sync/drop_tables_toml.py`
- `tools/data_sync/data_sync/drop_tables_sync.py`
- `tools/data_sync/data_sync/generators/go_drop_tables.py`
- `tools/data_sync/data_sync/model/drop_tables.py`
- `services/game-server/internal/game/drops/`
- TODO: add narrower drop-table source-map entries when they are confirmed.

## Related docs

- [Data](../!README.md)
- [Documentation procedure](../../documentation-procedure.md)
- [Data Sync](../../../tools/data_sync/!README.md)
- TODO: add drop-table-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future drop-table pipeline documentation.
Do not treat it as canonical source material.
