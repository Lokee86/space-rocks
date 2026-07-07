## Hot-lane size estimator cleanup skill

Use this skill to remove synchronous hot-path waste from realtime hot-lane packet construction.

The implementation goal is:

* Remove runtime raw-float reflection assertions from packet encoding.
* Replace hot-lane greedy JSON trial chunking with conservative compact-JSON byte estimation.
* Preserve current packet semantics.
* Preserve bullet movement rotation support for future projectile types, while estimating the actual tuple shape present in each update.

## Files

Expected files:

```text
services/game-server/internal/protocol/realtime/active.go
services/game-server/internal/protocol/realtime/hot_lane_chunker.go
services/game-server/internal/protocol/realtime/hot_lane_chunker_test.go
```

Likely new file:

```text
services/game-server/internal/protocol/realtime/hot_lane_size_estimate.go
```

Possible removal if unused:

```text
services/game-server/internal/protocol/realtime/quantize/assert.go
services/game-server/internal/protocol/realtime/quantize/assert_test.go
```

Possible test touch:

```text
services/game-server/internal/protocol/realtime/active_test.go
```

## Step 1: Remove runtime raw-float reflection

In:

```text
services/game-server/internal/protocol/realtime/active.go
```

Remove the `quantize.AssertNoRawFloats(...)` call from `encodeLanePacketUnchecked()`.

Remove the `quantize` import from `active.go` if it becomes unused.

If these files become unused, delete them:

```text
services/game-server/internal/protocol/realtime/quantize/assert.go
services/game-server/internal/protocol/realtime/quantize/assert_test.go
```

Do not remove or change actual quantization logic. The reflection assertion is only a diagnostic leak detector.

## Step 2: Add compact hot-lane byte estimation

Create:

```text
services/game-server/internal/protocol/realtime/hot_lane_size_estimate.go
```

Add helpers for estimating compact JSON bytes without calling `json.Marshal`.

Required helper shape:

```go
func estimateCompactJSONTupleBytes(items []any) int
func estimateJSONValueBytes(value any) int
func estimateJSONIntBytes(value int64) int
func estimateJSONStringBytes(value string) int
```

Use these rules:

* tuple brackets cost `2` bytes
* tuple commas cost `len(items)-1` bytes
* `nil` costs `4` bytes for `null`
* string estimates must include quote bytes
* integer estimates must count decimal digits and include a negative sign
* zero costs `1` byte
* unknown types must use a conservative fallback estimate

Prefer over-estimation to under-estimation.

## Step 3: Estimate bullet update tuples from actual shape

Add:

```go
func estimateCompactBulletMovementUpdateBytes(update map[string]any) int
```

The estimator must mirror the logical shape produced by `compactWirePackBulletMovementUpdate()`.

Supported bullet movement tuple shapes:

```text
[id]
[id,x]
[id,null,y]
[id,x,y]
[id,x,y,rotation]
```

Do not remove bullet rotation support from the generic wire path. Current straight bullets normally update `x/y`, but future projectiles may update rotation.

The estimator must inspect the actual update map and estimate the tuple it would produce. Do not hardcode all bullet updates as `[id,x,y,rotation]`.

## Step 4: Estimate asteroid update tuples from actual shape

Add:

```go
func estimateCompactAsteroidMovementUpdateBytes(update map[string]any) int
```

Supported asteroid movement tuple shapes:

```text
[id]
[id,x]
[id,null,y]
[id,x,y]
```

The estimator must inspect the actual update map and estimate the tuple it would produce.

## Step 5: Estimate hot packet envelope bytes

Add packet-level estimators:

```go
func estimateBulletDeltaPacketBytes(packet BulletWireDeltaPacket, updates []map[string]any) int
func estimateAsteroidDeltaPacketBytes(packet AsteroidWireDeltaPacket, updates []map[string]any) int
```

Estimate final compact packet shape, including packet metadata and update array overhead.

Bullet compact shape is approximately:

```json
{"t":"bd","q":123,"ms":0,"bq":122,"ci":0,"cc":3,"bu":[...]}
```

Asteroid compact shape is approximately:

```json
{"t":"ad","q":123,"ms":0,"bq":122,"ci":0,"cc":3,"au":[...]}
```

Include:

* object braces
* field names
* quotes
* colons
* commas
* metadata integer byte counts
* update array brackets
* commas between update tuples

The estimate may be conservative.

## Step 6: Replace greedy JSON trial chunking

In:

```text
services/game-server/internal/protocol/realtime/hot_lane_chunker.go
```

Replace the greedy per-update trial-encode behavior in:

```go
greedyBulletWireDeltaChunks()
greedyAsteroidWireDeltaChunks()
```

The old behavior to remove is:

```text
append update
encode trial packet
append update
encode trial packet
append update
encode trial packet
...
```

The new behavior should be:

```text
start empty chunk
for each update:
    estimate bytes if this update is added
    if adding update exceeds HardCapBytes and current chunk is not empty:
        close current chunk
        start new chunk
    add update
```

Each chunk must contain at least one update.

A single update that estimates above `HardCapBytes` should still become one chunk. The existing final encoded-size guard will decide whether it can be sent.

## Step 7: Use estimator for whole-packet fast path

For small deltas, use the estimator to keep the packet as one chunk.

Bullet example:

```go
if estimateBulletDeltaPacketBytes(packet, packet.BulletUpdates) <= HardCapBytes {
    return []RealtimeLaneCandidate{
        normalizedBulletWireDeltaCandidate(candidate, packet, packet.BulletUpdates, 0, 1),
    }
}
```

Asteroid equivalent:

```go
if estimateAsteroidDeltaPacketBytes(packet, packet.AsteroidUpdates) <= HardCapBytes {
    return []RealtimeLaneCandidate{
        normalizedAsteroidWireDeltaCandidate(candidate, packet, packet.AsteroidUpdates, 0, 1),
    }
}
```

Do not call JSON encode just to decide the fast path.

## Step 8: Preserve final encode guard

Do not remove the real packet encoding and send validation path.

The estimator is only used to split chunks cheaply. The final encoded packet bytes remain the real authority before send.

Keep existing behavior around:

```go
encodeLanePacket(...)
hotPacketSendAllowed(...)
```

## Step 9: Update tests

Update or add tests in:

```text
services/game-server/internal/protocol/realtime/hot_lane_chunker_test.go
```

Add estimator-vs-actual tests for bullet updates:

```text
[id,x,y]
[id,x]
[id,null,y]
[id,x,y,rotation]
zero values
negative values
```

Add estimator-vs-actual tests for asteroid updates:

```text
[id,x,y]
[id,x]
[id,null,y]
zero values
negative values
```

Each estimator test should compare the estimated byte count against actual encoded bytes from the real compact encoding path.

Required assertion:

```text
estimated >= actual
```

Keep existing chunking behavior coverage:

* small deltas stay one final chunk
* oversized deltas split
* chunk indexes are correct
* chunk counts are correct
* final flags are correct
* all updates are preserved
* chunks normally stay under `HardCapBytes`

## Constraints

Preserve packet semantics.

Preserve actual quantization behavior.

Preserve bullet movement rotation support in generic wire/compact code.

Do not add async queues.

Do not change metadata advance behavior.

Do not change baseline storage behavior.

Do not change event drain behavior.

Do not do unrelated cleanup.

If a change requires touching another file, touch only the smallest necessary additional file and report why.

If the task balloons, stop and report what expanded.

## Verification

After implementation, run:

```bash
cd /mnt/d/\!bin/space-rocks
{
  cd services/game-server && go test ./internal/protocol/realtime
  echo
  cd /mnt/d/\!bin/space-rocks && data-sync -check -packets -go -gds
} 2>&1 | tee /dev/tty | clip.exe
```

Then rerun the peak stress scenario and compare:

```text
main ramp writes/s
peak hold writes/s
encoded packet bytes
hot chunk counts
server CPU/log pressure
```

## Report

Report:

```text
Changed files
Behavior preserved or changed
Estimator test coverage added
Verification result
```
