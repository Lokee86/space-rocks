## Implementation plan

This should be one focused slice:

```text id="jp9tnp"
Make sr.asteroids and sr.bullets unordered/unreliable hot-update lanes.
Add stale hot-delta rejection so late packets cannot roll entity positions backward.
Fix scheduler classification so hot-lane deltas are supersedable, not required.
Do not move lifecycle creates/deletes out of sr.world.
```

## 1. Add explicit WebRTC channel policy

### Server

File:

```text id="n96rih"
services/game-server/internal/networking/webrtc_transport.go
```

Change `webRTCGameplayChannelSpec` from:

```go
type webRTCGameplayChannelSpec struct {
    Lane  string
    Label string
    ID    uint16
}
```

To include policy fields:

```go
type webRTCGameplayChannelSpec struct {
    Lane           string
    Label          string
    ID             uint16
    Ordered        bool
    MaxRetransmits *uint16
}
```

Then set:

```go
var noRetransmits uint16 = 0

return []webRTCGameplayChannelSpec{
    {Lane: "world", Label: "sr.world", ID: 1, Ordered: true},
    {Lane: "overlay", Label: "sr.overlay", ID: 2, Ordered: true},
    {Lane: "session", Label: "sr.session", ID: 3, Ordered: true},
    {Lane: "event", Label: "sr.event", ID: 4, Ordered: true},
    {Lane: "asteroids", Label: "sr.asteroids", ID: 5, Ordered: false, MaxRetransmits: &noRetransmits},
    {Lane: "bullets", Label: "sr.bullets", ID: 6, Ordered: false, MaxRetransmits: &noRetransmits},
}
```

In `createNegotiatedGameplayChannels()`, stop using one shared `ordered := true`. Use the spec:

```go
ordered := spec.Ordered
channel, err := p.peer.CreateDataChannel(spec.Label, &webrtc.DataChannelInit{
    Ordered:        &ordered,
    Negotiated:     &negotiated,
    ID:             &channelID,
    MaxRetransmits: spec.MaxRetransmits,
})
```

### Client

File:

```text id="xebk2c"
client/scripts/networking/webrtc/webrtc_transport.gd
```

Expand `GAMEPLAY_CHANNEL_SPECS`:

```gdscript
{"lane": "world", "label": "sr.world", "id": 1, "ordered": true},
{"lane": "overlay", "label": "sr.overlay", "id": 2, "ordered": true},
{"lane": "session", "label": "sr.session", "id": 3, "ordered": true},
{"lane": "event", "label": "sr.event", "id": 4, "ordered": true},
{"lane": "asteroids", "label": "sr.asteroids", "id": 5, "ordered": false, "max_retransmits": 0},
{"lane": "bullets", "label": "sr.bullets", "id": 6, "ordered": false, "max_retransmits": 0},
```

When creating channels, build options from the spec:

```gdscript
var options := {
    "negotiated": true,
    "id": int(spec.get("id")),
    "ordered": bool(spec.get("ordered", true)),
}
if spec.has("max_retransmits"):
    options["maxRetransmits"] = int(spec.get("max_retransmits"))
```

Use Godot’s expected option key `maxRetransmits`, not snake_case.

## 2. Add stale hot-delta rejection on the client

Files:

```text id="hbm4wj"
client/scripts/protocol/realtime/world_lane_state.gd
client/scripts/protocol/realtime/world_lane_applier.gd
```

Add state:

```gdscript
var latest_asteroid_delta_sequence := -1
var latest_bullet_delta_sequence := -1
```

Reset both on `clear_world()` / full world sync.

Add helpers:

```gdscript
func accept_asteroid_delta_sequence(sequence) -> bool:
    var parsed := _parse_hot_delta_sequence(sequence)
    if parsed == null:
        return false
    if parsed <= latest_asteroid_delta_sequence:
        return false
    latest_asteroid_delta_sequence = parsed
    return true

func accept_bullet_delta_sequence(sequence) -> bool:
    var parsed := _parse_hot_delta_sequence(sequence)
    if parsed == null:
        return false
    if parsed <= latest_bullet_delta_sequence:
        return false
    latest_bullet_delta_sequence = parsed
    return true

func _parse_hot_delta_sequence(sequence):
    if sequence == null:
        return null
    if typeof(sequence) != TYPE_INT and typeof(sequence) != TYPE_FLOAT:
        return null
    return int(sequence)
```

Then gate the hot appliers:

```gdscript
func apply_asteroid_delta(world_lane_state: WorldLaneState, lane: String, asteroid_packet: Dictionary) -> void:
    if not world_lane_state.accept_asteroid_delta_sequence(asteroid_packet.get("sequence")):
        return
    _apply_entity_deltas(world_lane_state, [], _array_field(asteroid_packet, "asteroid_updates"), [], "asteroid")

func apply_bullet_delta(world_lane_state: WorldLaneState, lane: String, bullet_packet: Dictionary) -> void:
    if not world_lane_state.accept_bullet_delta_sequence(bullet_packet.get("sequence")):
        return
    _apply_entity_deltas(world_lane_state, [], _array_field(bullet_packet, "bullet_updates"), [], "bullet")
```

Rules:

```text id="u86enm"
Accept sequence 100 then 103.
Reject sequence 101 after 103.
Do not require contiguous sequence numbers.
Drop missing/non-numeric sequence on hot lanes.
```

That matters because unreliable lanes may drop packets, so gaps are normal.

## 3. Fix scheduler classification for hot lanes

File:

```text id="6qwhal"
services/game-server/internal/protocol/realtime/planner.go
```

Current risk: `LaneAsteroids` / `LaneBullets` are not included in the delta classification branches, so they can fall through as required/critical.

Change `deliveryClassForCandidate()`:

```go
case RealtimeLaneCandidateKindDelta:
    switch candidate.Lane {
    case LaneSession:
        return DeliveryClassDeferrable
    case LaneWorld, LaneOverlay, LaneAsteroids, LaneBullets:
        return DeliveryClassHotSupersedable
    }
```

Change `priorityForCandidate()`:

```go
case RealtimeLaneCandidateKindDelta:
    switch candidate.Lane {
    case LaneSession:
        return PriorityMedium
    case LaneWorld, LaneOverlay, LaneAsteroids, LaneBullets:
        return PriorityHigh
    }
```

Do not classify asteroid/bullet deltas as required. They are disposable hot movement updates.

## 4. Preserve current ownership boundaries

Do not move these:

```text id="n6xxtk"
bullet_creates
bullet_deletes
asteroid_creates
asteroid_deletes
```

They stay in:

```text id="xf24wm"
sr.world / world_delta
```

Only these are on hot lanes:

```text id="0291rd"
asteroid_updates -> sr.asteroids
bullet_updates -> sr.bullets
```

Existing bullet buffering still matters because a bullet update may arrive before its world-lane create. Keep it.

## 5. Test updates

### Server WebRTC transport tests

File:

```text id="b76klm"
services/game-server/internal/networking/webrtc_transport_test.go
```

Update fake channel creation capture to include:

```go
MaxRetransmits *uint16
```

Assert:

```text id="g5q4vx"
sr.world     ordered=true,  maxRetransmits=nil
sr.overlay   ordered=true,  maxRetransmits=nil
sr.session   ordered=true,  maxRetransmits=nil
sr.event     ordered=true,  maxRetransmits=nil
sr.asteroids ordered=false, maxRetransmits=0
sr.bullets   ordered=false, maxRetransmits=0
```

### Client WebRTC transport tests

File:

```text id="qc29c1"
client/tests/unit/networking/test_webrtc_transport.gd
```

Update fake peer/channel capture to assert:

```text id="jut36b"
sr.world ordered true
sr.overlay ordered true
sr.session ordered true
sr.event ordered true
sr.asteroids ordered false and maxRetransmits 0
sr.bullets ordered false and maxRetransmits 0
```

### Client hot stale packet tests

File:

```text id="1ixacv"
client/tests/unit/protocol/realtime/test_world_lane_applier.gd
```

Add:

```text id="fw6plb"
asteroid_delta sequence 10 applies
asteroid_delta sequence 9 is ignored
asteroid_delta sequence 11 applies
bullet_delta sequence 10 applies
bullet_delta sequence 9 is ignored
bullet_delta sequence 12 applies after gap
missing sequence hot packet is ignored
non-numeric sequence hot packet is ignored
```

### Server scheduler tests

Likely file:

```text id="omlm4n"
services/game-server/internal/protocol/realtime/planner_test.go
```

Add/adjust assertions:

```text id="ap4mqh"
LaneAsteroids delta -> DeliveryClassHotSupersedable, PriorityHigh
LaneBullets delta -> DeliveryClassHotSupersedable, PriorityHigh
```

## 6. Docs after code lands

Update only docs that describe current WebRTC lane behavior. The doc language should become:

```text id="5mcukc"
sr.world, sr.overlay, sr.session, and sr.event are ordered/reliable negotiated gameplay channels.

sr.asteroids and sr.bullets are unordered/unreliable negotiated gameplay channels with maxRetransmits=0. They carry supersedable movement/update packets only. Late hot packets are rejected by sequence. Missing packets are tolerated because later packets replace older movement state.

Lifecycle create/delete records remain on sr.world.
```

Remove stale wording like:

```text id="6yj2g5"
all hot lanes start ordered/reliable
later hot lanes may become unordered/unreliable
```

## 7. Verification

```bash id="u8kc3j"
cd /mnt/d/\!bin/space-rocks
{
  echo "== data sync =="
  data-sync -check -packets -go -gds

  echo
  echo "== server targeted tests =="
  go test ./services/game-server/internal/networking ./services/game-server/internal/protocol/realtime

  echo
  echo "== client targeted tests =="
  cd client
  godot --headless -s addons/gut/gut_cmdln.gd \
    -gtest=res://tests/unit/networking/test_webrtc_transport.gd \
    -gtest=res://tests/unit/protocol/realtime/test_world_lane_applier.gd \
    -gtest=res://tests/unit/protocol/realtime/test_lane_protocol_routing.gd
} 2>&1 | tee /dev/tty | clip.exe
```

Main acceptance criteria:

```text id="ace8ep"
Hot WebRTC channels are unordered/unreliable on both server and client.
Reliable lanes remain ordered/reliable.
Late asteroid/bullet hot packets cannot roll state backward.
Hot lane deltas are scheduler-supersedable, not required/critical.
Creates/deletes still route through sr.world.
```

My attempted local MCP check failed with a connector permission error, so this plan is based on the current repo files available through GitHub plus the recent diff context already inspected.
