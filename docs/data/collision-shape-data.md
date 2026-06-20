# Collision Shape Data
Parent index: [Data](./!README.md)

## Purpose

This document describes the collision shape data source of truth and the export/generated-output flow that keeps client and server shape names aligned.

## Overview

`client/tools/export_collision_shapes.gd` is the client-side export tool.

The export tool writes the shared collision shape data output to `shared/collisions/collision_shapes.json`. That shared output is then consumed by the server collision-shape generation and verification outputs.

## Source files

* `client/tools/export_collision_shapes.gd`

## Shared output

* `shared/collisions/collision_shapes.json`

## Generated outputs

* `services/game-server/internal/game/physics/collision_shapes.go`
* `services/game-server/internal/game/physics/collision_shapes_test.go`
* `services/game-server/tests/physics/collision_shapes_test.go`

## Data role

The export output owns the shared collision shape names and shape definitions used to keep client-exported collision data and server-consumed collision data in sync.

Generated client-facing or server-facing collision outputs are derived from the shared JSON output and should be treated as generated data, not hand-authored sources.

The server outputs provide the collision shape catalog and tests that verify the generated collision shape mapping.

## Validation and failure modes

Collision shape data should fail fast when the export output or generated outputs drift from the shared source.

Known failure modes include:

* stale exports after source edits
* missing shapes in the shared collision output
* invalid JSON in `shared/collisions/collision_shapes.json`
* client/server collision mismatch
* shape-name mismatches between the shared output and generated consumers

## Related docs

* [Server collision shapes](../../services/game-server/simulation/world/collision-shapes.md)
* [Server physics](../../services/game-server/simulation/world/physics.md)
* [Gameplay packets stub](../../protocol/stubs/gameplay-packets.md)
* [Data](../!README.md)

## Notes

This document stays on collision shape data ownership and export/generated-output flow rather than physics runtime or collision resolution behavior.
