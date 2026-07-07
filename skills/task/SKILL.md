## Real bullet hot-lane chunking skill

Use this skill when fixing Space Rocks realtime hot-lane overflow where `sr.bullets` or another hot lane stops sending updates after packet size pressure increases.

The goal is to create **real bounded hot-lane packet chunks** before encoding and writing. Do not rely on scheduler-only chunk records unless they map to real encoded packets.

## Problem summary

The current failure mode is:

* Bullet movement updates are offloaded from `sr.world` to `sr.bullets`.
* The bullet hot packet grows as bullet count rises.
* Once the encoded `bullet_delta` exceeds the hard hot-packet cap, `encodeLanePacket()` drops it.
* The server keeps sending `world_delta`, but `bullet_delta` collapses or stops.
* Raising policy constants does not fix this because `size_estimate.go` has separate active constants such as `HardCapBytes = 1200`.
* Scheduler chunking is not enough if it only creates `ScheduleRecord` chunks. Real packets must be split before encoding/writing.

The fix must produce multiple actual `bullet_delta` messages on the same WebRTC lane, each under the hard cap.

## Do not change

Do not change these unless the task explicitly says otherwise:

* Do not raise hot-lane packet size caps as the fix.
* Do not move bullet creates/deletes out of `sr.world`.
* Do not move asteroid creates/deletes out of `sr.world`.
* Do not change WebRTC channel policy.
* Do not change bullet rendering.
* Do not change torpedo spawning.
* Do not make client rendering changes to hide the server issue.
* Do not treat scheduler-only chunking as real packet chunking.

## Required behavior

Hot-lane overflow must degrade like this:

* One small bullet hot packet remains one packet.
* One oversized bullet hot packet becomes multiple real `bullet_delta` packets.
* Every chunk must encode under `HardCapBytes`.
* All chunks from one source hot delta should share the same sequence.
* Chunks must carry `chunk_index`, `chunk_count`, and `is_final_chunk`.
* The writer must send every encoded chunk, even when multiple chunks use `LaneBullets`.
* The client must not reject valid same-sequence hot chunks as stale.
* Duplicate hot updates are acceptable because position updates overwrite by entity ID.

## Files usually involved

Server:

* `services/game-server/internal/protocol/realtime/hot_lane_policy.go`
* `services/game-server/internal/protocol/realtime/size_estimate.go`
* `services/game-server/internal/protocol/realtime/wire_packets.go`
* `services/game-server/internal/protocol/realtime/baseline.go`
* `services/game-server/internal/protocol/realtime/hot_lane_allocator.go`
* `services/game-server/internal/protocol/realtime/planner.go`
* `services/game-server/internal/protocol/realtime/active.go`
* `services/game-server/internal/protocol/realtime/scheduler.go`
* `services/game-server/internal/networking/websocket_write.go`

Client:

* `client/scripts/protocol/realtime/world_lane_state.gd`

Tests:

* `services/game-server/internal/protocol/realtime/hot_lane_chunker_test.go`
* `services/game-server/internal/protocol/realtime/active_test.go`
* `services/game-server/internal/protocol/realtime/baseline_test.go`
* `services/game-server/internal/protocol/realtime/scheduler_test.go`
* Existing client tests near `world_lane_state.gd`, if present.

## Implementation sequence

### 1. Restore sane hot-lane constants

In `hot_lane_policy.go`, keep the policy defaults sane:

```go
DefaultBulletHotLaneEntityBudget   = 48
DefaultTargetEncodedPacketBytes    = 800
DefaultHardEncodedPacketBytes      = 1200
DefaultMTUSafePacketBytes          = 1500
```

Do not solve the bug by setting these to huge values.

### 2. Add hot-packet chunk metadata

In `wire_packets.go`, extend `AsteroidWireDeltaPacket` and `BulletWireDeltaPacket` with:

```go
ChunkIndex   int  `json:"chunk_index,omitempty"`
ChunkCount   int  `json:"chunk_count,omitempty"`
IsFinalChunk bool `json:"is_final_chunk,omitempty"`
```

Update `wireAsteroidWireDeltaPacket()` and `wireBulletWireDeltaPacket()` to include those fields when emitting hot packets.

### 3. Fix hot-lane metadata extraction

In `baseline.go`, update `CandidateMetadata()` so it understands:

```go
AsteroidWireDeltaPacket
*AsteroidWireDeltaPacket
BulletWireDeltaPacket
*BulletWireDeltaPacket
```

Return metadata with the correct lane, sequence, snapshot kind, chunk index, chunk count, final chunk flag, and server sent time.

This prevents diagnostics from reporting hot-lane sequence `0` and lets hot lane state advance correctly.

### 4. Use hot-lane sequences

In `planner.go`, hot-lane deltas should use their own lane sequence, not world-lane sequence.

Before appending hot candidates, derive:

```go
asteroidState, asteroidSynced := state.LaneState(LaneAsteroids)
asteroidSequence := NextLaneSequence(asteroidState, asteroidSynced)

bulletState, bulletSynced := state.LaneState(LaneBullets)
bulletSequence := NextLaneSequence(bulletState, bulletSynced)
```

Assign these to the split hot deltas:

```go
split.AsteroidDelta.Sequence = asteroidSequence
split.BulletDelta.Sequence = bulletSequence
```

All chunks from a single source hot delta should share the same sequence and differ by chunk metadata.

### 5. Add real hot-lane candidate chunking

Create `hot_lane_chunker.go`.

Add a function like:

```go
func ExpandHotLaneCandidateChunks(candidates []RealtimeLaneCandidate) []RealtimeLaneCandidate
```

Rules:

* Non-hot candidates pass through unchanged.
* `LaneBullets` delta candidates are split if their encoded packet exceeds `HardCapBytes`.
* Small bullet packets return one candidate with `ChunkIndex = 0`, `ChunkCount = 1`, and `IsFinalChunk = true`.
* Oversized bullet packets become multiple real `RealtimeLaneCandidate` values.
* Every emitted bullet chunk must contain a non-empty subset of `BulletUpdates`.
* Every emitted chunk must encode under `HardCapBytes` unless a single update is itself pathological.
* Preserve update order.
* Preserve packet type, sequence, and server sent time.
* Set chunk indexes from `0` to `chunk_count - 1`.
* Only the last chunk has `IsFinalChunk = true`.

Preferred implementation:

* Try encoding the full candidate.
* If it fits, return it as one finalized chunk.
* If it does not fit, split the update slice.
* Use binary search or recursive halving to find the largest prefix that fits under `HardCapBytes`.
* Repeat until all updates are emitted.

Do not create scheduler-only chunks here. Create actual candidates.

### 6. Expand candidates before scheduling

In `active.go`, the flow should become:

1. Assemble candidates.
2. Expand hot-lane chunks into real candidates.
3. Build schedule records from the expanded candidates.
4. Select send plan.
5. Select candidates from the expanded candidate list.
6. Encode each selected candidate.
7. Write every encoded packet.

The scheduler must operate on the real chunk candidates.

### 7. Refactor encoded packet storage

The current shape cannot represent multiple packets for the same lane if it stores packets by lane:

```go
EncodedPackets map[Lane][]byte
EncodedBytes   map[Lane]int
```

Replace or supplement it with an ordered list:

```go
type EncodedRealtimeLanePacket struct {
	Candidate     RealtimeLaneCandidate
	Encoded       []byte
	EncodedBytes  int
}
```

Then `ActiveRealtimeResult` should hold:

```go
EncodedPackets []EncodedRealtimeLanePacket
```

Optionally keep an aggregate `EncodedBytesByLane` map only for metrics or tests.

### 8. Write every encoded packet

In `websocket_write.go`, change the writer loop so it iterates encoded packets, not only selected candidates keyed by lane.

Use the candidate stored on each encoded packet:

```go
for _, encoded := range result.EncodedPackets {
	candidate := encoded.Candidate
	encodedPacket := encoded.Encoded
	// send candidate on candidate.Lane
}
```

This allows multiple `sr.bullets` messages in one server tick.

Logging must use the chunk candidate and the chunk’s encoded byte count.

### 9. Remove fake scheduler chunking

In `scheduler.go`, remove or neutralize the block that chunks only `ScheduleRecord` values when `EstimatedBytes > HardCapBytes`.

Scheduler-only chunks are misleading if they do not correspond to real encoded packets.

Chunking ownership should be:

```text
candidate chunker -> scheduler -> encoder -> writer
```

Not:

```text
scheduler fake chunks -> deduplicated candidates -> one oversized encoded packet
```

### 10. Relax client same-sequence stale rejection

In `world_lane_state.gd`, update bullet and asteroid hot-lane sequence acceptance.

Bad shape:

```gdscript
if parsed <= latest_bullet_delta_sequence:
	return false
latest_bullet_delta_sequence = parsed
return true
```

Safer first pass:

```gdscript
if parsed < latest_bullet_delta_sequence:
	return false
latest_bullet_delta_sequence = parsed
return true
```

Apply the same logic for asteroid hot deltas.

Reason: valid chunks from the same source delta share a sequence. Same-sequence chunks must not be rejected. Duplicate hot updates are acceptable because they overwrite entity positions by ID.

## Required tests

### Server: metadata support

Add or update tests proving `CandidateMetadata()` returns proper metadata for:

* `AsteroidWireDeltaPacket`
* `BulletWireDeltaPacket`

Assert sequence, chunk index, chunk count, final chunk, and lane.

### Server: small bullet packet remains one chunk

Create a bullet hot candidate that encodes under `HardCapBytes`.

Assert:

* output candidate count is `1`.
* lane is `LaneBullets`.
* sequence is preserved.
* `ChunkIndex == 0`.
* `ChunkCount == 1`.
* `IsFinalChunk == true`.

### Server: oversized bullet packet splits

Create a bullet hot candidate with enough updates to exceed `HardCapBytes`.

Assert:

* output candidate count is greater than `1`.
* every chunk is `LaneBullets`.
* every encoded chunk is `<= HardCapBytes`.
* every chunk has non-empty `BulletUpdates`.
* all original bullet updates appear exactly once across chunks.
* all chunks share the same sequence.
* chunk indexes are ordered and complete.
* only the final chunk has `IsFinalChunk == true`.

### Server: active result supports multiple packets on same lane

Build a snapshot with enough bullet movement updates to require chunking.

Assert:

* `BuildActiveRealtimeResult()` returns multiple encoded packets for `LaneBullets`.
* no bullet chunk exceeds `HardCapBytes`.
* total encoded packet list includes the world packet plus multiple bullet packets when expected.
* packets are not collapsed by lane.

### Server: scheduler no longer fakes hot chunks

Update scheduler tests so they no longer imply `ScheduleRecord` chunking creates real packet chunks.

If a scheduler chunking test remains, it must only describe scheduler records, not actual packet emission.

### Client: same-sequence chunks

If there is a client test harness, add tests for bullet and asteroid hot sequence acceptance:

* accept sequence `10`
* accept another packet with sequence `10`
* reject sequence `9`
* accept sequence `11`

## Expected log after fix

Before the fix, logs look like this at high bullet count:

```text
bullet_delta 60/s near 1180 bytes
bullet_delta collapses to 35/s
bullet_delta collapses to 7/s
bullet_delta reaches 0/s
world_delta continues 60/s
```

After the fix, logs should look like this:

```text
bullet_delta sequence=123 chunk_index=0 chunk_count=3 encoded_bytes=1050
bullet_delta sequence=123 chunk_index=1 chunk_count=3 encoded_bytes=1080
bullet_delta sequence=123 chunk_index=2 chunk_count=3 encoded_bytes=920
```

The bullet lane should continue writing bounded chunks instead of disappearing.

## Verification

Run targeted Go tests from the Go module root:

```bash
cd /mnt/d/\!bin/space-rocks/services/game-server
{
  go test ./internal/protocol/realtime
} 2>&1 | tee /dev/tty | clip.exe
```

Run broader server tests:

```bash
cd /mnt/d/\!bin/space-rocks/services/game-server
{
  go test ./...
} 2>&1 | tee /dev/tty | clip.exe
```

Then repeat the bullet stream stress test.

## Completion criteria

This task is complete when:

* Bullet hot packets are split into real encoded chunks.
* Multiple `sr.bullets` packets can be sent in one tick.
* No bullet chunk exceeds `HardCapBytes`.
* Hot packet diagnostics show nonzero sequence metadata.
* Same-sequence chunks are accepted by the client.
* Bullet streams no longer collapse to zero packet sends at the old size threshold.
* Creates/deletes remain in `sr.world`.
* Movement updates remain on `sr.bullets`.
