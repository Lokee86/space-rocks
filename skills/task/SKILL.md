## Skill: Hot-Lane Chunking Documentation Update

Use this skill when updating Space Rocks documentation after the realtime protocol change that added real candidate-level chunking for asteroid and bullet hot lanes.

## Purpose

Restore documentation accuracy after the hot-lane chunking change.

The current implementation truth is:

```text
- sr.asteroids and sr.bullets are unordered/unreliable WebRTC hot-update lanes.
- asteroid_delta and bullet_delta can now split oversized hot movement update lists into multiple real candidate-level chunks.
- Chunks use the existing Metadata envelope: sequence, chunk_index, chunk_count, and is_final_chunk.
- All chunks from one original hot movement delta share the same lane-local sequence and differ by chunk metadata.
- The active result/write path can emit multiple encoded packets on the same lane in one tick.
- The WebRTC writer sends each EncodedLanePacket, not one packet per lane map entry.
- The scheduler no longer owns fake chunking. It schedules real candidates that already exist.
- Hot packets still run encoded-size classification before send.
- Packets above the hard-cap/MTU class are still rejected.
- The client rejects lower hot-lane sequence values as stale.
- Same-sequence hot packets are valid when they are chunks from one asteroid_delta or bullet_delta sequence.
- Lifecycle creates/deletes remain on sr.world.
- event_batch remains ordered/reliable, batched, and non-chunked.
- The project still does not have general record/entity-level prioritization or all-lane fragmentation.
```

## Required ownership framing

Use this wording distinction consistently:

```text
Implemented:
focused hot-lane chunking for asteroid_delta and bullet_delta

Not implemented:
general-purpose payload fragmentation for all lane families
general record/entity-level prioritization
interest filtering
binary/protobuf representation
all-lane record or field sub-packet prioritization
```

Do not describe the change as generic fragmentation unless clearly qualified as **focused asteroid/bullet hot-lane chunking**.

## Files to update

Update these docs when stale references exist:

```text
docs/protocol/realtime-websocket-protocol.md
docs/protocol/realtime-webrtc-gameplay-transport.md
docs/services/game-server/networking/outbound-message-flow.md
docs/services/game-server/simulation/runtime/lane-packet-projection.md
docs/services/game-server/networking/realtime-compact-wire-mapping.md
docs/services/client/networking-flow/inbound-packet-routing.md
docs/protocol/gameplay-packets.md
docs/domains/technical/realtime-client-server-flow.md
docs/planning/protocol/realtime-protocol-architecture.md
docs/planning/domains/technical/network-observability-and-packet-budget.md
docs/planning/development-roadmap.md
```

Do not create a new document for this change unless the existing docs cannot reasonably own the fact. They currently can.

Do not delete documents. The stale material is embedded in otherwise valid docs.

## Stale phrases to remove or correct

Search for these phrases and update them if they describe current behavior:

```text
does not claim full fragmentation
does not claim full fragmentation or payload-splitting
payload fragmentation
asteroid_delta = one candidate
bullet_delta = one candidate
cadence allows it
does not currently split state-lane deltas
late asteroid_delta and bullet_delta packets are rejected by monotonic sequence
```

Also inspect nearby uses of:

```text
hot-packet encoded-size guards
monotonic sequence
one candidate
chunk metadata exists
```

These phrases may remain only if the surrounding text clearly reflects the current focused hot-lane chunking behavior.

## Canonical replacement language

Use these replacement statements where appropriate.

### Chunk metadata and real chunking

```text
Chunk metadata exists in the wire shape and scheduler records. For asteroid_delta and bullet_delta, oversized hot movement update lists are split into multiple real same-sequence lane candidates before scheduling and encoding. Each chunk is encoded and written as its own WebRTC DataChannel message on sr.asteroids or sr.bullets. This is focused hot-lane chunking, not general fragmentation for all lane families.
```

### Scheduling and candidate count

```text
asteroid_delta = one or more candidates when asteroid hot movement is present; oversized hot movement update lists are split into bounded same-sequence chunks
bullet_delta = one or more candidates when bullet hot movement is present; oversized hot movement update lists are split into bounded same-sequence chunks
```

### Active path budget behavior

```text
The active path does not implement general record/entity-level prioritization or arbitrary field-level packet splitting. It does implement focused hot-lane chunking for asteroid_delta and bullet_delta: oversized hot movement update lists become multiple real candidates before SelectSendPlan and before encoding.
```

### Encoded-size guards

```text
Hot asteroid and bullet movement packets are first split into bounded real lane candidates when needed, then encoded and classified by encoded size before send. Packets above the hard-cap/MTU class are still rejected, but normal oversized hot movement lists should be chunked before that gate.
```

### Client hot sequence behavior

```text
The client rejects lower hot-lane sequence values as stale. Same-sequence hot packets are valid when they are chunks from one asteroid_delta or bullet_delta sequence and may apply independently. Sequence gaps remain valid because unordered/unreliable hot packets may be dropped.
```

### Per-tick packet count

```text
packet_count counts encoded packets actually written, not unique logical lanes. Under hot-lane stress it may include multiple asteroid_delta or bullet_delta packets for the same lane in one tick.
```

## File-specific edit map

## `docs/protocol/realtime-websocket-protocol.md`

This is the canonical protocol doc. Update it first.

Required edits:

* Update lane metadata language so active asteroid and bullet packets are covered where Metadata envelope behavior applies.
* State that `chunk_index` and `chunk_count` are emitted when `chunk_count > 1`.
* State that same-sequence chunks are valid for `asteroid_delta` and `bullet_delta`.
* Replace any “chunk metadata exists but no payload splitting” claim.
* Replace `asteroid_delta = one candidate` and `bullet_delta = one candidate`.
* Replace “active path does not currently split state-lane deltas” with the focused hot-lane chunking distinction.
* Clarify that general record/entity-level prioritization remains future work.

Required current behavior summary:

```text
For asteroid_delta and bullet_delta, oversized hot movement update lists are split into multiple real same-sequence lane candidates before scheduling and encoding. The active writer can emit multiple encoded packets on sr.asteroids or sr.bullets in one tick. The client rejects lower sequence values but accepts same-sequence chunks.
```

## `docs/protocol/realtime-webrtc-gameplay-transport.md`

This doc owns physical WebRTC lane behavior.

Required edits:

* Replace strict monotonic sequence wording with lower-sequence rejection plus same-sequence chunk acceptance.
* Update the server send boundary to include expanded asteroid/bullet hot chunks and an encoded packet list.
* Replace “payload fragmentation” in the non-support list with “general-purpose payload fragmentation for all lane families.”
* Add that focused hot-lane chunking is implemented for `asteroid_delta` and `bullet_delta`.

Suggested server send boundary:

```text
BuildActiveRealtimeResultForGame
-> realtime lane candidates, including expanded asteroid/bullet hot chunks when needed
-> selected realtime lane candidates
-> encoded lane packet list
-> SendEncodedLaneJSON(candidate.Lane, encodedPacket) for each encoded packet
-> physical WebRTC gameplay DataChannel
```

## `docs/services/game-server/networking/outbound-message-flow.md`

This doc owns the service write path.

Required edits:

* Update the active write steps so they mention expanded hot chunks and `EncodedLanePackets`.
* Replace any statement that chunk metadata exists but payload splitting is not implemented.
* Clarify that networking writes already-encoded packets and does not own general fragmentation.
* Clarify that multiple wire logs can appear for the same hot lane in one tick.
* Clarify that `packet_count` counts encoded packets written, not unique lanes.
* Add `hot_lane_chunker.go` to the code map.
* Update `active.go` ownership to mention the encoded packet list.
* Update `scheduler.go` ownership to say scheduler schedules already-built candidates and does not own fake packet chunking.

Suggested code-map additions:

```text
- services/game-server/internal/protocol/realtime/hot_lane_chunker.go - expands oversized asteroid_delta and bullet_delta hot movement candidates into bounded real candidate chunks before scheduling and encoding.
- services/game-server/internal/protocol/realtime/active.go - active lane packet encoding path, EncodedLanePackets list construction, raw-float assertion/compact/packetcodec boundary, and encoded-byte accounting.
- services/game-server/internal/protocol/realtime/scheduler.go - include/defer planning for already-built candidates; real hot-lane chunks are created before scheduling.
```

## `docs/services/game-server/simulation/runtime/lane-packet-projection.md`

This doc owns projection-side behavior.

Required edits:

* Update the flow to include hot movement chunk expansion after hot movement split and before serialization/encoding.
* Add `hot_lane_chunker.go` to code map.
* Update scheduler ownership to remove fake chunking implications.
* Keep lifecycle ownership unchanged: asteroid/bullet creates/deletes remain on world lane.

Suggested flow insert:

```text
-> oversized asteroid/bullet hot movement update lists expand into real same-sequence candidate chunks
```

## `docs/services/game-server/networking/realtime-compact-wire-mapping.md`

This doc owns compact alias and metadata wire mapping.

Required edits:

* Expand runtime metadata inference wording from world/overlay/session to world/asteroid/bullet/overlay/session where accurate.
* Clarify that `asteroid_delta` and `bullet_delta` emit `chunk_index` and `chunk_count` when split into multiple chunks.
* Clarify all chunks for one original hot movement delta share the same lane-local sequence.
* Keep `event_batch` excluded from chunking.
* If `is_final_chunk` is emitted as compact `fc` for current runtime chunks, move it out of “legacy only” language and describe it as conditional runtime metadata when emitted.

Suggested text:

```text
For asteroid_delta and bullet_delta, chunk_index and chunk_count are emitted when a hot movement update list is split into multiple candidate chunks. All chunks for one original hot-lane delta share the same lane-local sequence and differ by chunk_index.
```

## `docs/services/client/networking-flow/inbound-packet-routing.md`

This doc owns client inbound routing behavior.

Required edits:

* Replace strict monotonic hot rejection wording.
* State that lower sequences are rejected, same-sequence hot chunks are valid.
* State that multiple `asteroid_delta` or `bullet_delta` packets may arrive for the same lane sequence in one poll window.
* Do not describe routing as coalescing chunks.
* Do not move gameplay application ownership into this doc.

Suggested text:

```text
Hot asteroid and bullet packets are routed on unordered/unreliable lanes. The client rejects lower sequence values so late packets cannot roll positions backward. Same-sequence packets are valid for chunked asteroid_delta or bullet_delta output and may apply independently. Sequence gaps are valid because hot packets can be dropped.
```

## `docs/protocol/gameplay-packets.md`

This doc should stay high level.

Required edits:

* Add that `asteroid_delta` and `bullet_delta` may be emitted as multiple same-sequence chunks when a hot movement update list would exceed the encoded packet cap.
* Clarify that chunks only move existing entities.
* Keep lifecycle creates/deletes on `world_delta`.
* Replace stale sequence wording with lower-sequence rejection plus same-sequence chunk acceptance.

Suggested text:

```text
When a hot movement update list would exceed the encoded packet cap, asteroid_delta and bullet_delta may be emitted as multiple same-sequence chunks. These chunks still only move existing entities; lifecycle creates/deletes remain on world_delta.
```

## `docs/domains/technical/realtime-client-server-flow.md`

This domain doc should mention the current fact without becoming a protocol spec.

Required edits:

* Add that oversized asteroid/bullet hot movement update lists are split into multiple real same-sequence hot-lane packets before encoding.
* Update stale sequence wording.
* Update future-work list to include current focused hot-lane chunking as shipped, while keeping deeper prioritization as future.

Suggested text:

```text
Oversized asteroid_delta and bullet_delta movement update lists are split into multiple real same-sequence hot-lane packets before encoding. This keeps hot packets under the encoded hard cap while leaving lifecycle creates/deletes on sr.world.
```

## `docs/planning/protocol/realtime-protocol-architecture.md`

This planning doc must not frame shipped hot-lane chunking as future work.

Required edits:

* Add implemented status bullets for focused asteroid/bullet hot-lane chunking.
* Update “hot packets have encoded-size guards” to include focused chunking.
* Update client sequence guard wording.
* Keep record/entity-level prioritization, interest filtering, and deeper budget policy as future work.

Suggested implemented bullets:

```text
- Oversized asteroid_delta and bullet_delta movement update lists are split into real same-sequence hot-lane candidate chunks before scheduling and encoding.
- Active output can emit multiple encoded packets on sr.asteroids or sr.bullets in one tick.
- Hot asteroid and bullet packets have focused candidate-level chunking and encoded-size guards before send.
- Client hot-lane sequence guards reject lower sequence values while accepting same-sequence chunks for split asteroid_delta and bullet_delta packets.
```

## `docs/planning/domains/technical/network-observability-and-packet-budget.md`

This planning doc should reflect the new evidence model.

Required edits:

* Add that focused hot-lane chunking is current for asteroid and bullet deltas.
* State that packet evidence can show multiple packets per tick on `sr.asteroids` or `sr.bullets`.
* State that `packet_count` counts encoded packets, not unique lanes.
* State that bandwidth should be interpreted alongside write cadence.

Suggested current-state additions:

```text
- Focused hot-lane chunking is current for asteroid_delta and bullet_delta. Packet evidence should expect multiple packets per tick on sr.asteroids or sr.bullets under stress.
- packet_count is a count of encoded packets written, not unique lanes.
- Encoded bandwidth evidence should be interpreted with write cadence: under peak stress, bandwidth may drop because write cadence drops even while entity pressure rises.
```

## `docs/planning/development-roadmap.md`

Only update current-status sections if they mention realtime protocol compression, packet budget, or protocol implementation state.

Suggested wording:

```text
Compact JSON aliases, sparse delta omission, tuple packing, lane-native WebRTC channels, focused asteroid/bullet hot-lane chunking, candidate-level scheduling, estimated byte-budget selection, and hot-packet encoded-size guards are implemented. General record/entity-level prioritization, interest filtering, and binary/protobuf representation remain future work.
```

## Files that usually do not need edits

Do not edit these unless an exact stale claim is found:

```text
docs/data/packet-schemas.md
docs/services/game-server/networking/websocket-session-lifecycle.md
docs/services/game-server/simulation/runtime/game-aggregate.md
docs/systems-design/world/world-authority.md
```

Reason:

```text
- No packet TOML/schema SSoT change is implied.
- Hot-lane chunking is protocol/realtime behavior, not game aggregate ownership.
- World authority does not change.
- WebSocket lifecycle does not change.
```

## Consistency rules

Use these rules while editing:

```text
- Do not say asteroid_delta or bullet_delta are always one candidate.
- Do not say hot lanes only have send-time encoded-size guards.
- Do not say same-sequence hot packets are stale.
- Do not say networking owns fragmentation.
- Do not say event_batch chunks.
- Do not imply lifecycle creates/deletes moved out of sr.world.
- Do not imply general record/entity-level prioritization is implemented.
- Do not imply binary/protobuf/compression is implemented.
```

## Final stale-search checklist

After edits, search docs for these phrases:

```text
does not claim full fragmentation
does not claim full fragmentation or payload-splitting
payload fragmentation
asteroid_delta = one candidate
bullet_delta = one candidate
cadence allows it
does not currently split state-lane deltas
late asteroid_delta and bullet_delta packets are rejected by monotonic sequence
```

Allowed remaining concepts:

```text
general record/entity-level prioritization remains future work
general-purpose fragmentation remains not implemented
event_batch remains non-chunked
world/overlay/session are not generalized record-splitting lanes
```

## Completion report

When finished, report:

```text
- Changed files
- Stale claims removed
- Current behavior now documented
- Any remaining matches and why they are acceptable
```
