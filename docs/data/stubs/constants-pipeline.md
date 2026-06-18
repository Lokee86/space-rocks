# Constants Pipeline
Parent index: [Data](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe constants TOML and data-sync pipeline documentation.

## Overview

TODO: summarize how constants move from TOML source files through data-sync into generated language outputs.
Stub note: keep this focused on the active constants pipeline.

## Source files

- `shared/constants/server_constants.toml`
- `shared/constants/server_entities.toml`
- `shared/constants/client/presentation.toml`
- `shared/constants/client/shell.toml`
- `shared/constants/client/lobby.toml`
- TODO: add any other confirmed constants source files.

## Configuration

- `tools/data_sync/config.toml`
- `tools/data_sync/!README.md`
- TODO: add any other constants pipeline configuration roots when they are confirmed.

## Generated outputs

- `client/scripts/generated/constants/constants.gd`
- `services/game-server/internal/constants/constants.go`
- TODO: add any other generated constants outputs when they are confirmed.

## Consumers

- Client GDScript runtime.
- Game-server Go runtime.
- TODO: add any other confirmed consumers.

## Pipeline usage

- `tools/data_sync -push -constants -go`
- `tools/data_sync -push -constants -go -gds`
- `tools/data_sync -check -constants -go -gds`
- `tools/data_sync -validate -constants`
- TODO: add any other verified constants pipeline commands.

## Validation commands

- `tools/data_sync -validate -constants`
- `tools/data_sync -check -constants -go -gds`
- `tools/data_sync -diff -constants -go -gds`
- TODO: add any other verified validation commands.

## Failure modes

- Duplicate or conflicting constants sections across source files.
- Missing generated output after source changes.
- Stale generated code after TOML edits.
- TODO: add any other verified failure modes.

## Code or source map

- `shared/constants/`
- `tools/data_sync/data_sync/constants_sync.py`
- `tools/data_sync/data_sync/constants_store.py`
- `tools/data_sync/data_sync/generators/go_constants.py`
- `tools/data_sync/data_sync/generators/gds_constants.py`
- `tools/data_sync/data_sync/generators/ts_constants.py`
- TODO: add narrower constants pipeline source-map entries when they are confirmed.

## Related docs

- [Data](../!README.md)
- [Documentation procedure](../../documentation-procedure.md)
- [Data Sync](../../../tools/data_sync/!README.md)
- TODO: add constants-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future constants pipeline documentation.
Do not treat it as canonical source material.
