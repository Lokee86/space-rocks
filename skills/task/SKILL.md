## Hot Lane Subtractive Ownership Skill

Use this skill when changing Space Rocks realtime hot-lane ownership so regular asteroid and bullet movement updates are removed from `sr.world` and owned by their dedicated hot lanes.

This skill is specifically for the current transition from mixed world/hot-lane overflow to subtractive hot-lane ownership. It is not a general WebRTC, packet-codec, or unreliable-channel skill.

## Goal

Make hot movement ownership subtractive:

```text
sr.world keeps lifecycle and resync-safe world state.
sr.asteroids owns asteroid movement/update records.
sr.bullets owns bullet movement/update records.
```

The intended result is that regular asteroid and bullet `Updates` no longer inflate `sr.world`. Dedicated hot lanes should carry those movement updates instead.

## Ownership rules

Preserve this distinction exactly:

```text
Move out of sr.world:
- worldDelta.Asteroids.Updates
- worldDelta.Bullets.Updates

Keep in sr.world:
- worldDelta.Asteroids.Creates
- worldDelta.Asteroids.Deletes
- worldDelta.Bullets.Creates
- worldDelta.Bullets.Deletes
- worldDelta.Ships
- worldDelta.Pickups
- world metadata / baseline / resync-safe state
```

Do not move creates or deletes to `sr.asteroids` or `sr.bullets` in this slice. The current dedicated hot packets only carry update arrays:

```go
type AsteroidWireDeltaPacket struct {
    Type            string
    Sequence        int
    ServerSentMsec  int
    AsteroidUpdates []map[string]any
}

type BulletWireDeltaPacket struct {
    Type           string
    Sequence       int
    ServerSentMsec int
    BulletUpdates  []map[string]any
}
```

## Primary files

Expected implementation and test files:

```text
services/game-server/internal/protocol/realtime/hot_lane_allocator.go
services/game-server/internal/protocol/realtime/hot_lane_allocator_test.go
```

Only inspect or change additional files if the allocator tests expose a real dependency.

Likely related file if planner behavior must be checked:

```text
services/game-server/internal/protocol/realtime/planner.go
```

## Implementation rule

The allocator should stop selecting a retained subset of asteroid/bullet updates for `sr.world`.

Replace this behavior:

```text
sr.world keeps N asteroid/bullet movement updates.
sr.asteroids receives asteroid overflow.
sr.bullets receives bullet overflow.
```

With this behavior:

```text
sr.world keeps zero regular asteroid movement updates.
sr.world keeps zero regular bullet movement updates.
sr.asteroids receives asteroid movement updates.
sr.bullets receives bullet movement updates.
```

A correct split should set:

```go
result.WorldDelta.Asteroids.Updates = nil
result.WorldDelta.Bullets.Updates = nil
```

when those updates are emitted on dedicated hot-lane packets.

Create asteroid and bullet hot-lane deltas from the original world update arrays:

```go
result.AsteroidDelta = &AsteroidWireDeltaPacket{
    Type:            PacketFamilyAsteroidDelta,
    Sequence:        worldDelta.Metadata.Sequence,
    ServerSentMsec:  worldDelta.Metadata.ServerSentMsec,
    AsteroidUpdates: worldDelta.Asteroids.Updates,
}

result.BulletDelta = &BulletWireDeltaPacket{
    Type:           PacketFamilyBulletDelta,
    Sequence:       worldDelta.Metadata.Sequence,
    ServerSentMsec: worldDelta.Metadata.ServerSentMsec,
    BulletUpdates:  worldDelta.Bullets.Updates,
}
```

Only emit a dedicated delta when its update array is non-empty.

## Budget and cadence rules

Keep the existing lane budget constants. They should still control hot-lane pressure modes and cadence:

```text
DefaultAsteroidHotLaneEntityBudget
DefaultBulletHotLaneEntityBudget
```

They should no longer mean “how many asteroid/bullet movement updates stay in `sr.world`.”

Use existing mode logic where practical:

```text
<= budget * 2: full-owned 30 Hz behavior
<= budget * 3: full-owned 20 Hz behavior
>  budget * 3: needs chunking behavior
```

Preserve existing cadence gating in the planner unless a test proves it is incompatible with subtractive ownership.

Skipped movement updates should not become reliable backlog in `sr.world`. The next allowed hot-lane send should carry fresh latest-state movement updates.

## Route state rules

Route state should reflect ownership:

```text
asteroid update IDs route to HotUpdateRouteAsteroids
bullet update IDs route to HotUpdateRouteBullets
```

Do not preserve stale `HotUpdateRouteWorld` entries for active asteroid or bullet movement updates once subtractive ownership is active.

Continue removing missing asteroid and bullet route entries so deleted entities do not leave stale route state behind.

## Required tests

Add or update focused tests in:

```text
services/game-server/internal/protocol/realtime/hot_lane_allocator_test.go
```

Required coverage:

```text
Low pressure:
- 1 asteroid update and 1 bullet update
- world asteroid Updates length is 0
- world bullet Updates length is 0
- asteroid delta is emitted with 1 update
- bullet delta is emitted with 1 update
```

```text
Stress pressure:
- 80 asteroid updates and 80 bullet updates
- world asteroid Updates length is 0
- world bullet Updates length is 0
- asteroid delta has 80 updates
- bullet delta has 80 updates
- asteroid and bullet route state points to dedicated hot lanes
```

```text
Lifecycle safety:
- asteroid Creates remain in sr.world
- asteroid Deletes remain in sr.world
- bullet Creates remain in sr.world
- bullet Deletes remain in sr.world
- asteroid/bullet movement Updates are removed from sr.world
```

If existing tests assert that low-pressure asteroid or bullet updates remain in `sr.world`, update those tests to match subtractive ownership. Do not keep old mixed-overflow expectations.

## Expected log result

After implementation, stress logs should show:

```text
sr.world packet size drops sharply.
sr.asteroids packet size rises.
sr.bullets packet size rises.
summary encoded bytes may remain similar because it sums lanes.
the largest actual single datachannel packet should no longer usually be sr.world.
```

Primary validation target:

```text
sr.world should no longer be large because of regular asteroid/bullet movement updates.
```

If `sr.world` still has large packets after this change, diagnose what remains in `world_delta` instead of reworking asteroid/bullet hot-lane ownership again.

## Out of scope

Do not combine this change with:

```text
- unreliable/unordered datachannel changes
- WebRTC channel configuration changes
- packet codec rewrites
- protobuf work
- create/delete migration into hot-lane packets
- client interpolation changes
- broad scheduler rewrites
- docs-wide cleanup
```

Unreliable hot lanes are the next likely step, but this skill is only for ownership correction while lanes remain otherwise unchanged.

## Stop conditions

Stop and report instead of expanding the change if any of these are true:

```text
- Client requires create/delete data from hot-lane packets before movement updates can apply.
- Dedicated hot-lane packet types must be expanded beyond update arrays.
- Planner behavior requires a broad redesign instead of a small cadence/mode adjustment.
- The change touches unrelated protocol families or unrelated gameplay logic.
```

## Report requirements

Report:

```text
- Changed files
- Whether asteroid/bullet Updates are removed from sr.world
- Whether Creates/Deletes remain in sr.world
- Tests added or updated
- Any planner behavior touched
```
