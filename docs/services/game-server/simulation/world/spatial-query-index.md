---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7515-b79f-825951135c72
document_type: general
policy_exempt: false
summary: This document is the canonical description of the implemented toroidal spatial-query seam used by the Go game-server simulation.
---
# Toroidal Spatial Query Index

Parent index: [Game Server Simulation World](./!INDEX.md)

## Purpose

This document is the canonical description of the implemented toroidal spatial-query seam used by the Go game-server simulation.

## Overview

The spatial-query seam provides a reusable broad-phase index for bounded toroidal world queries. The generic contract lives in `internal/game/spatial`; the current implementation lives in `internal/game/spatial/grid` and is a uniform grid.

The index returns candidate references. It does not own collision consequences, entity maps, or network presentation. Entries use conservative bounding circles so the broad phase can safely find possible overlaps; exact physics collision detection remains authoritative.

## Code root

```text
services/game-server/internal/game/
```

Relevant packages and integration files:

```text
internal/game/spatial/
internal/game/spatial/grid/
internal/game/collision_spatial_index.go
internal/game/collision_candidates.go
internal/game/simulation.go
```

## Responsibilities

The seam owns:

- The generic `spatial.Index` contract for rebuilding entries and querying circles or rectangles.
- Spatial reference types and kind-mask filtering.
- Toroidal normalization and wrapped grid-cell traversal in the current implementation.
- Broad-phase candidate filtering using conservative entry bounding circles.
- Per-query deduplication when an entry occupies multiple cells or is encountered through a wrapped seam.
- Rebuilding the current index contents from the entries supplied by `Game`.

## Does not own

The seam does not own:

- Runtime entity maps or entity lifecycle.
- Lexical actor ordering, candidate ordering, or collision consequences.
- Exact physics collision detection, body placement, damage, destruction, scoring, drops, pickup collection, or respawn effects.
- World bounds source data or generated constants.
- HTTP, WebSocket, packet, client, or realtime interest-management APIs.
- Persistent storage or API/business-server concerns.

`Game` owns lexical actor ordering, wrapped-distance/ID candidate ordering, and all consequences after candidate lookup.

## Domain roles

`internal/game/spatial` is the generic domain seam. It defines the implementation-independent `spatial.Index` contract and the value types exchanged with callers.

`internal/game/spatial/grid` is the current toroidal uniform-grid implementation. It owns bucket layout, cell traversal, insertion, and query bookkeeping while satisfying the generic contract.

`Game` is the orchestration and policy owner. It projects eligible asteroid or pickup collision bodies into entries, rebuilds at the required simulation boundaries, orders candidates deterministically, and performs exact physics and consequences.

## Protocols and APIs

This is internal Go only. No HTTP, WebSocket, packet, or client API exists.

The generic contract is:

```go
type Index interface {
    Rebuild(entries []Entry)
    QueryCircle(dst []Ref, center physics.Vector2, radius float64, mask KindMask) []Ref
    QueryRect(dst []Ref, rect Rect, mask KindMask) []Ref
}
```

The destination slice is append-oriented: callers may provide reusable storage, and an empty mask (or a negative circle radius) returns it unchanged.

### Ownership-level types

- `Entry` is the rebuild input: a `Ref`, a world position, and a conservative non-negative bounding `Radius`.
- `Ref` is the stable candidate identity composed of a spatial `Kind` and string `ID`.
- `KindMask` selects eligible kinds. `AllKinds` includes the currently defined player, projectile, asteroid, enemy, and pickup kinds.
- `QueryCircle` finds entries whose conservative circles overlap the query circle under wrapped distance.
- `QueryRect` finds entries whose conservative circles overlap the query rectangle under wrapped distance. Negative half-extents are clamped to zero.

These are broad-phase operations; their returned references are not proof of an exact physics collision.

## Data ownership

The index owns transient buckets and query-generation bookkeeping only. `Rebuild` replaces the prior indexed contents and clears stale entries.

`Game` supplies transient entries from active runtime entities. The asteroid rebuild skips nil, pending-despawn, and missing-shape entries. The pickup rebuild skips nil and missing-shape entries. The entry radius is obtained from `physics.BoundingRadius` for the entity collision shape.

The index does not retain ownership of runtime entities and does not mutate them.

## Broad-phase model

Each indexed entry is represented by one conservative bounding circle. The grid inserts that circle into every cell it may touch. A circle query first visits potentially covered cells, then applies wrapped circle-overlap testing using the query radius plus the entry radius. A rectangle query visits potentially covered cells, then applies wrapped circle-versus-rectangle testing.

The broad phase may return conservative false positives. Consumers must resolve candidates against the authoritative runtime state and exact physics detector.

## Toroidal grid behavior

The current grid is constructed with world bounds and a requested cell size of `256` units (`defaultSpatialCellSize`). It normalizes every rebuild entry position and every query center into the active bounds. Negative entry radii are clamped to zero.

Cell traversal wraps independently on each axis. Queries crossing horizontal, vertical, or corner seams visit the corresponding wrapped cells without creating ghost entities. Entries larger than one cell can occupy multiple cells; entries spanning a full axis are not duplicated across repeated wrapped cells.

The requested cell size is a target, not an assumption that bounds divide evenly. The implementation uses `ceil(bounds / 256)` cells per axis and derives the exact per-axis cell extents as `bounds.Width / cellsX` and `bounds.Height / cellsY`. Thus non-divisible bounds tile the complete extent exactly rather than leaving a remainder or extending beyond the world.

Each query advances a generation stamp. A `Ref` already observed in the current generation is skipped, so multi-cell entries and seam traversal produce at most one result. Rebuild clears the seen map and resets the generation state; generation wrap also clears the map before continuing.

## Current simulation use

`Game` rebuilds the asteroid index before the ship/projectile collision families. This gives those collision checks a current broad-phase projection of active asteroid bodies.

After projectile consequences have been applied, `Game` rebuilds the pickup index before pickup collection. Pickup candidates therefore reflect the post-projectile state and are not a stale continuation of the asteroid index.

Collision candidate helpers query by kind, remove references that no longer resolve to eligible runtime entities, and sort by wrapped distance with ID as the tie-breaker. Game-owned lexical actor ordering controls the actor iteration order. Exact wrapped body placement and physics detection remain outside the index, followed by Game-owned consequences.

## Invariants

- One `Ref` identifies one indexed candidate for a query result.
- Rebuild removes stale entries from the prior projection.
- Indexed positions are normalized to the configured toroidal bounds.
- Indexed radii are non-negative conservative bounds.
- Queries traverse wrapped cells and use wrapped spatial distance.
- Broad-phase results never replace exact physics collision detection.
- Candidate ordering and all gameplay consequences remain Game-owned.
- The index has no network or client-facing contract.

## Code map

```text
internal/game/spatial/index.go
    spatial.Index contract: Rebuild, QueryCircle, QueryRect.

internal/game/spatial/types.go
    Kind, KindMask, Ref, Entry, and Rect.

internal/game/spatial/grid/index.go
    Uniform-grid state, construction, and rebuild.

internal/game/spatial/grid/rebuild.go
    Wrapped cell insertion and rebuild support for conservative entry circles.

internal/game/spatial/grid/query_circle.go
    Wrapped circle traversal and overlap query.

internal/game/spatial/grid/query_rect.go
    Wrapped rectangle traversal and circle/rectangle overlap query.

internal/game/spatial/grid/query_common.go
    Generation-stamped reference deduplication.

internal/game/collision_spatial_index.go
    Game-owned asteroid and pickup projections and rebuilds.

internal/game/collision_candidates.go
    Kind filtering, stale-reference compaction, and wrapped-distance/ID ordering.
```

## Tests and benchmarks

The contract and implementation are covered by Go tests under:

```text
internal/game/spatial/
internal/game/spatial/grid/
internal/game/collision_spatial_index_test.go
internal/game/collision_spatial_integration_test.go
internal/game/collision_candidates_test.go
```

Coverage includes kind filtering, circle and rectangle queries, seam wrapping, large entries, stale-entry replacement, normalization, exact non-divisible tiling, deduplication, destination reuse, and allocation behavior. `collision_spatial_benchmark_test.go` covers the integrated rebuild/query path.

Run the game-server tests from `services/game-server` with:

```bash
go test -buildvcs=false ./...
```

## Network-interest boundary

Recipient-specific network interest is implemented outside this seam under `internal/protocol/realtime/network_interest.go`. It operates on immutable `GameplayPresentationSnapshot` data and uses the shared wrap-aware camera-region predicate under `internal/game/visibility`.

The current interest implementation does not query or repurpose this mutable authoritative collision broad phase. That separation is intentional:

```text
simulation spatial index
= mutable Game-owned broad phase for authoritative collision candidates

network interest
= recipient-specific policy over published presentation snapshots
```

A future performance optimization may add a presentation-owned immutable lookup/index, but it must preserve this ownership split and must not read mutable simulation buckets from realtime projection.

Quadtree replacement remains future-only. The current uniform grid is the implemented simulation backend behind the generic contract; no quadtree exists in this boundary and no replacement is implied by the interface.

## Related docs

- [Toroidal Space And Motion](toroidal-space-and-motion.md)
- [Physics](physics.md)
- [Collision Shapes](collision-shapes.md)
- [Game Server Simulation](../!INDEX.md)
- [Game Server](../../!INDEX.md)
- [Network Interest](../../networking/network-interest.md)

## Notes

This document describes the implemented simulation spatial index. Network interest is also implemented, but remains a separate presentation/realtime boundary and is linked rather than duplicated here. Quadtree replacement remains non-current.
