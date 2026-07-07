## Hot-Lane Stall Fix Skill

Use this skill when fixing the Space Rocks realtime regression where `sr.asteroids` and `sr.bullets` stop updating unless player/world state changes.

## Goal

Restore smooth asteroid and bullet hot-lane updates when the player is completely still, not shooting, and the reliable `sr.world` lane has no gameplay movement changes.

The fix must make asteroid and bullet movement updates emit based on hot-lane update presence, not on whether the `world` lane is dirty.

## Core diagnosis

The unreliable WebRTC transport policy is not the primary bug.

Keep this policy:

* `sr.world`: reliable/ordered
* `sr.overlay`: reliable/ordered
* `sr.session`: reliable/ordered
* `sr.event`: reliable/ordered
* `sr.asteroids`: unordered/unreliable, zero retransmits
* `sr.bullets`: unordered/unreliable, zero retransmits

The bug is in the server realtime planner. Hot asteroid and bullet deltas are gated by `worldDeltaHasChanges` and by cadence based on `worldSequence`. When the player is still and not shooting, `world` may not produce changes, so asteroid/bullet hot packets can stall even though asteroid/bullet positions changed.

## Files

Primary file:

* `services/game-server/internal/protocol/realtime/planner.go`

Expected test file:

* `services/game-server/internal/protocol/realtime/planner_test.go`

Do not modify unless a test proves it is required:

* client rendering files
* client packet application files
* WebRTC transport setup
* docs

## Required behavior

Asteroid and bullet movement updates must be sent on their hot lanes whenever hot updates are present.

Lifecycle ownership must remain unchanged:

* asteroid creates stay in `sr.world`
* asteroid deletes stay in `sr.world`
* asteroid movement updates go to `sr.asteroids`
* bullet creates stay in `sr.world`
* bullet deletes stay in `sr.world`
* bullet movement updates go to `sr.bullets`

Do not move lifecycle creates/deletes into hot lanes.

## Implementation steps

1. Open `services/game-server/internal/protocol/realtime/planner.go`.

2. Find the hot-lane eligibility logic near the world delta split.

   The broken shape is:

   ```
   asteroidHotAllowed := asteroidHotPresent && (worldDeltaHasChanges || hotPacketCadenceAllows(split.CohortState.AsteroidMode, worldSequence))
   bulletHotAllowed := bulletHotPresent && (worldDeltaHasChanges || hotPacketCadenceAllows(split.CohortState.BulletMode, worldSequence))
   ```

3. Replace that eligibility with direct hot-update presence:

   ```
   asteroidHotAllowed := asteroidHotPresent
   bulletHotAllowed := bulletHotPresent
   ```

4. Preserve projection advancement.

   If the planner currently appends a world delta candidate when hot-lane updates are emitted, keep that behavior for now so the stored world projection advances to the current full world state.

   The desired condition shape is:

   ```
   if worldDeltaHasChanges || asteroidHotAllowed || bulletHotAllowed {
       chainedWorldProjection := quantizedWorldFull
       chainedWorldProjection.Metadata = split.WorldDelta.Metadata
       candidates = append(candidates, RealtimeLaneCandidate{
           Lane:       LaneWorld,
           Kind:       RealtimeLaneCandidateKindDelta,
           Delta:      split.WorldDelta,
           Projection: chainedWorldProjection,
       })
   }
   ```

5. Keep asteroid and bullet hot candidates appended independently:

   ```
   if asteroidHotAllowed {
       candidates = append(candidates, RealtimeLaneCandidate{Lane: LaneAsteroids, Kind: RealtimeLaneCandidateKindDelta, Delta: *split.AsteroidDelta})
   }
   if bulletHotAllowed {
       candidates = append(candidates, RealtimeLaneCandidate{Lane: LaneBullets, Kind: RealtimeLaneCandidateKindDelta, Delta: *split.BulletDelta})
   }
   ```

6. Do not create fake world changes to force hot-lane sending.

7. Do not introduce a new cadence system in this fix. Correctness comes first. A proper hot-lane cadence can be added later using an independent simulation/send tick or per-lane sequence.

## Test changes

Update `services/game-server/internal/protocol/realtime/planner_test.go`.

### Remove or rewrite the bad cadence test

Delete or rewrite the test named:

* `TestAssembleRealtimeLaneCandidatesSkipsHotPacketsOnCadenceAndUsesLatestSnapshot`

That test preserves the broken world-sequence coupling. Do not keep an expectation that hot lanes skip merely because the world sequence is not on an eligible cadence tick.

### Add asteroid regression test

Add a test named:

* `TestAssembleRealtimeLaneCandidatesEmitsAsteroidHotDeltaWhenOnlyAsteroidsMove`

Required setup:

* Previous snapshot has:

  * one stable player
  * many asteroids at old positions
* Current snapshot has:

  * same stable player with no movement/state changes
  * same asteroid IDs at new positions
* Realtime session state has:

  * final synced `LaneWorld` baseline metadata
  * `LaneWorld` marked baseline-ready
  * stored world projection from the previous snapshot

Required assertions:

* `LaneAsteroids` candidate exists.
* Candidate kind is `RealtimeLaneCandidateKindDelta`.
* Candidate delta is `AsteroidWireDeltaPacket`.
* `AsteroidUpdates` is non-empty.
* If a `LaneWorld` candidate exists, its world delta must not contain asteroid movement updates in `worldDelta.Asteroids.Updates`.

### Add bullet regression test

Add a test named:

* `TestAssembleRealtimeLaneCandidatesEmitsBulletHotDeltaWhenOnlyBulletsMove`

Required setup:

* Previous snapshot has:

  * one stable player
  * many bullets at old positions
* Current snapshot has:

  * same stable player with no movement/state changes
  * same bullet IDs at new positions
* Realtime session state has:

  * final synced `LaneWorld` baseline metadata
  * `LaneWorld` marked baseline-ready
  * stored world projection from the previous snapshot

Required assertions:

* `LaneBullets` candidate exists.
* Candidate kind is `RealtimeLaneCandidateKindDelta`.
* Candidate delta is `BulletWireDeltaPacket`.
* `BulletUpdates` is non-empty.
* If a `LaneWorld` candidate exists, its world delta must not contain bullet movement updates in `worldDelta.Bullets.Updates`.

## Existing behavior to preserve

Do not break these test expectations:

* Hot asteroid and bullet schedule records use `DeliveryClassHotSupersedable`.
* Hot asteroid and bullet schedule records use high priority.
* Creates/deletes remain in the world delta under pressure.
* Full world projection remains complete after hot-lane splitting.
* Hot-lane packets are real candidates, not debug-only records.

## Non-goals

Do not do these in this fix:

* Do not revert unreliable/unordered WebRTC channels.
* Do not change client rendering.
* Do not change projectile scene resolution.
* Do not change client packet routing.
* Do not add a new hot-lane cadence abstraction.
* Do not move lifecycle events out of `sr.world`.
* Do not perform unrelated cleanup.
* Do not edit docs unless explicitly requested separately.

## Acceptance criteria

The implementation is complete when:

* Asteroid hot-lane deltas emit when only asteroid positions changed.
* Bullet hot-lane deltas emit when only bullet positions changed.
* The tests prove hot-lane emission does not depend on player movement, shooting, or unrelated world-lane changes.
* Existing lifecycle ownership remains unchanged.
* Existing realtime planner tests pass after updating the stale cadence expectation.
* Manual playtest shows smooth asteroid and bullet updates while the player is still and not firing.

## Report format

When done, report:

* Changed files
* Tests added or updated
* Whether hot-lane emission is now independent of `worldDeltaHasChanges`
* Whether lifecycle creates/deletes still remain in `sr.world`
* Any tests run and their result

