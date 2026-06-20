# Asteroid Variants Data
Parent index: [Data](./!README.md)

## Purpose

This document describes the asteroid variant data source of truth and the generated outputs that consume it.

## Overview

`shared/asteroids/variants.toml` is the asteroid variant source data.

That source data defines the variant entries used to generate client and server asteroid variant data. The generated outputs keep the client and game server aligned on variant ids, catalog order, and variant-count expectations.

## Source files

* `shared/asteroids/variants.toml`

## Generated outputs

* `client/scripts/generated/asteroids/asteroid_variants.gd`
* `services/game-server/internal/game/asteroids/variants.go`
* `services/game-server/internal/game/asteroids/variants_test.go`

## Data role

The source data owns the asteroid variant catalog content.

Generated client and server variant data are derived from the same source and should be treated as outputs, not hand-authored sources.

The client output provides generated variant data for client-side consumption.
The server outputs provide the runtime variant catalog and tests that validate the generated catalog shape.

## Validation and failure modes

Asteroid variant data should fail fast when the generated outputs drift from the source data or when the source data becomes invalid.

Known failure modes include:

* stale generated outputs after source edits
* invalid variant entries in `shared/asteroids/variants.toml`
* missing required fields in the source data
* count mismatches between the source catalog and generated outputs
* index mismatches between source entries and generated catalogs

## Related docs

* [Client entity sync owners](../../services/client/world-sync/entity-sync-owners.md)
* [Server asteroid spawning and variants](../../services/game-server/simulation/world/asteroid-spawning-and-variants.md)
* [Gameplay packets stub](../../protocol/stubs/gameplay-packets.md)
* [Data](../!README.md)

## Notes

This document stays on data ownership and generated outputs rather than runtime spawning or client presentation behavior.
