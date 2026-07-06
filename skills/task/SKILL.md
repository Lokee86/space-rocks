# Client Bullet Hot-Lane Ordering and Presentation Fanout Skill

Use this skill when fixing Space Rocks client stress behavior where high-volume bullet updates cause bullets to disappear, pop in late, or trigger noticeable frame drops.

## Goal

Fix remaining client-side stress behavior without changing WebRTC transport reliability.

Targets:

1. Bullets should not disappear or pop in late when creates and updates arrive through different lanes or different client frames.
2. Presentation fanout should not run once per gameplay packet under bullet/asteroid packet bursts.
3. Existing WebRTC ordered/reliable channels, bounded receive draining, and hot-lane packet routing should remain intact.

## Scope

This skill covers two narrow fixes:

1. Buffer bullet hot-lane updates that arrive before their matching bullet create.
2. Defer realtime presentation fanout so it runs at most once per client frame.

Do not switch WebRTC hot lanes to unreliable or unordered in this skill.

Do not add packet dropping or coalescing in this skill.

Do not add full metrics in this skill unless explicitly requested. Metrics should be a follow-up slice after behavior is fixed.

Do not broadly refactor world presentation, projectile rendering, or realtime routing.

## Background

Bullets are currently split across lanes:

```text
world_delta  = bullet creates and deletes
bullet_delta = bullet position updates
```

This is correct protocol ownership, but it introduces a client-side ordering hazard.

When the client receives and drains packets from multiple WebRTC channels with per-frame budgets, a `bullet_delta` can be applied before the matching `bullet_create` from `world_delta`.

Current state behavior ignores updates for unknown bullet IDs. That was safe when creates and updates were in one ordered lane. It is not safe once bullet updates are on a dedicated hot lane.

Failure shape:

```text
Frame N:
- bullet_delta arrives first for bullet-123
- client has not processed bullet-123 create yet
- update is ignored

Frame N+1:
- world_delta create for bullet-123 arrives
- bullet appears late or at stale/create position

Frame N+2:
- later bullet_delta arrives
- bullet suddenly renders correctly
```

This can look like bullets not rendering from some positions, then suddenly rendering fully.

The remaining frame-rate problem is separate: presentation fanout currently appears to be packet-coupled. Under hot bullet traffic, many packets can cause many full presentation passes in a single frame.

Desired shape:

```text
packet received -> route into realtime state -> mark presentation dirty
_process() -> fan out presentation once -> clear dirty flag
```

## Phase 1: Buffer pending bullet updates

### Objective

Make bullet hot-lane updates resilient when they arrive before their matching bullet create.

### Files

Expected files:

* `client/scripts/protocol/realtime/world_lane_state.gd`
* `client/scripts/protocol/realtime/world_lane_applier.gd`
* `client/tests/unit/protocol/realtime/test_world_lane_applier.gd`

Optional only if needed:

* `client/tests/unit/protocol/realtime/test_lane_protocol_routing.gd`

### Required behavior

If a bullet update arrives for an unknown bullet ID:

```text
- Do not create the bullet.
- Store the latest pending update for that bullet ID.
```

If another pending update arrives for the same unknown bullet ID:

```text
- Replace the previous pending update.
- Keep only the latest update.
```

When the matching bullet create arrives:

```text
- Apply the bullet create.
- Immediately apply the pending update on top of the create.
- Clear the pending update for that bullet ID.
```

When a bullet delete arrives:

```text
- Delete the bullet.
- Clear any pending update for that bullet ID.
```

When a full world state is applied or the world is cleared:

```text
- Clear all pending bullet updates.
```

### Suggested implementation

In `world_lane_state.gd`, add:

```text
var pending_bullet_updates := {}
```

Add concrete methods:

```text
merge_or_buffer_bullet_update(record: Dictionary) -> void
apply_pending_bullet_update(id) -> void
clear_pending_bullet_update(id) -> void
clear_pending_bullet_updates() -> void
```

Suggested behavior:

```text
merge_or_buffer_bullet_update(record):
- Get id from record.
- If id is null, return.
- If bullets has id, merge_bullet_update(record).
- Else store a narrowed update in pending_bullet_updates[id].
```

```text
apply_pending_bullet_update(id):
- If pending_bullet_updates does not have id, return.
- If bullets has id, merge_bullet_update(pending_bullet_updates[id]).
- Erase pending_bullet_updates[id].
```

```text
clear_pending_bullet_update(id):
- Erase pending_bullet_updates[id].
```

```text
clear_pending_bullet_updates():
- Clear pending_bullet_updates.
```

Keep pending update records narrowed through existing bullet field rules. Do not store unrelated fields.

In `world_lane_applier.gd`:

* Bullet creates should call `upsert_bullet(decoded)` and then `apply_pending_bullet_update(decoded["id"])`.
* Bullet updates from `bullet_delta` should call `merge_or_buffer_bullet_update(decoded)`.
* Bullet updates inside `world_delta` can also use `merge_or_buffer_bullet_update(decoded)` for consistency.
* Bullet deletes should call `delete_bullet(id)` and then `clear_pending_bullet_update(id)`.
* Full-world apply and clear paths should clear pending bullet updates.

Avoid adding a generic pending system for all entity kinds in this slice. Fix bullets first.

### Tests

Add tests in `test_world_lane_applier.gd`.

Required coverage:

1. Unknown bullet hot update is buffered but does not create a bullet.
2. Bullet create later applies the pending update immediately.
3. Multiple pending updates for the same bullet keep only the latest.
4. Bullet delete clears pending update.
5. Full world apply clears all pending bullet updates.

Test data should be small and direct:

```text
bullet id: bullet-1
create x/y: 10, 20
pending update x/y: 30, 40
newer pending update x/y: 50, 60
```

Assert that after create, `world_lane_state.bullets["bullet-1"]` has the pending update coordinates, not the stale create coordinates.

## Phase 2: Defer presentation fanout to once per frame

### Objective

Stop running full realtime presentation fanout once per gameplay packet.

### Files

Expected file:

* `client/scripts/session/gameplay_session_controller.gd`

Expected test file depends on existing coverage. Use the smallest existing relevant test file. If no suitable test file exists, create one under:

* `client/tests/unit/session/test_gameplay_session_controller_realtime_fanout.gd`

### Current problem

Gameplay packets are already routed by `ClientConnectionService` before `gameplay_packet_received` is emitted.

`GameplaySessionController.handle_gameplay_packet(packet)` should not perform full presentation fanout immediately for every packet during bursts. Packet receive should mark presentation dirty. `_process()` should present at most once per frame.

### Required behavior

On gameplay packet:

```text
- If gameplay packets are not accepted, return.
- If gameplay readiness is missing or not ready, return.
- Preserve first-ready logging behavior.
- Preserve event diagnostics behavior for event_batch packets.
- Mark presentation dirty.
- Do not fan out immediately.
```

On `_process(delta)`:

```text
- Run existing gameplay_composition.process(...) behavior.
- If presentation is dirty and gameplay is ready, fan out presentation once.
- Clear dirty flag after fanout.
```

If multiple gameplay packets arrive in the same frame:

```text
- Mark dirty once.
- Fan out once during _process.
```

If a later frame receives another packet:

```text
- Mark dirty again.
- Fan out once in that later frame.
```

### Suggested implementation

In `gameplay_session_controller.gd`, add:

```text
var _presentation_dirty := false
var _pending_event_fanout := false
```

Optional:

```text
var _last_dirty_packet_type := ""
```

Extract the current fanout block from `handle_gameplay_packet(packet)` into a concrete helper:

```text
func _fanout_realtime_presentation_once() -> void:
```

This helper should contain the current logic that:

* logs `"Gameplay presentation fanout started"` once
* gets `world_sync`
* gets `gameplay_hud_flow`
* gets `event_lifecycle_flow`
* calls `gameplay_presentation_adapter.fanout_lane_states(...)`
* builds/applies devtools lane state
* restores alive presentation
* marks presentation fanned out once

Do not create an abstract wrapper. This helper is a direct movement of existing behavior.

In `handle_gameplay_packet(packet)`:

```text
- Keep the acceptance/readiness guards.
- Keep event_batch diagnostics.
- Set _presentation_dirty = true.
- Return without calling fanout.
```

In `_process(delta)`:

```text
- Keep existing gameplay_composition.process(delta, required_lane_baselines_synced).
- After that, if _presentation_dirty is true and gameplay is ready:
  - call _fanout_realtime_presentation_once()
  - set _presentation_dirty = false
```

Reset should clear:

```text
_presentation_dirty = false
_pending_event_fanout = false
```

### Event batch handling

Preserve event fanout.

If the existing fanout block needs `event_lifecycle_flow`, compute it inside `_fanout_realtime_presentation_once()` from `gameplay_composition`.

If event diagnostics are currently packet-specific, keep them in `handle_gameplay_packet(packet)`.

If multiple event batches arrive in one frame, it is acceptable for the one presentation fanout to use the latest accumulated event state from the router.

## Phase 3: Tests for once-per-frame fanout

### Objective

Prove presentation fanout is no longer packet-coupled.

### Test requirements

Add or update focused tests.

Required coverage:

1. Multiple gameplay packets before `_process()` do not cause multiple presentation fanouts.
2. `_process()` performs one fanout when dirty and gameplay ready.
3. Dirty flag clears after fanout.
4. A later packet marks dirty again and allows one more fanout on a later `_process()`.
5. If gameplay is not ready, dirty state does not cause fanout.

Use fakes rather than full scenes where possible.

Recommended fake shape:

```text
FakePresentationAdapter:
- can_fanout() returns configurable bool
- fanout_lane_states(...) increments fanout_count
- mark_fanned_out() records call
```

```text
FakeGameplayComposition:
- has gameplay_shell_flow/runtime_context/world_sync shape if needed
- has gameplay_hud_flow if needed
- process(...) increments process_count
- apply_devtools_gameplay_state(...) records calls
- restore_alive_presentation_from_realtime_router(...) records calls
```

Keep the test focused on controller behavior, not actual rendering.

## Phase 4: Stress validation

After Phase 1 and Phase 2:

Run targeted tests first:

```bash
cd /mnt/d/\!bin/space-rocks
{
  echo "== packet sync check =="
  data-sync -check -packets -go -gds

  echo
  echo "== realtime state and presentation tests =="
  cd client
  godot --headless -s addons/gut/gut_cmdln.gd \
    -gtest=res://tests/unit/protocol/realtime/test_world_lane_applier.gd \
    -gtest=res://tests/unit/protocol/realtime/test_lane_protocol_routing.gd \
    -gtest=res://tests/unit/networking/test_webrtc_transport.gd \
    -gtest=res://tests/unit/test_client_connection_service_webrtc.gd

  echo
  echo "== server smoke =="
  cd /mnt/d/\!bin/space-rocks/services/game-server
  go test ./internal/networking ./internal/protocol/realtime
} 2>&1 | tee /dev/tty | clip.exe
```

Then run the stress scenario.

Expected improvements:

```text
- Bullets no longer vanish before popping in.
- Bullets that receive hot updates before creates appear at the latest known update position when created.
- Frame drops reduce during high bullet volume.
- No stale bullets stick around after delete or clear.
- No growing pending bullet update map after deletes/full-world resets.
```

## Phase 5: Metrics after behavior fixes

Do not add metrics before the two behavior fixes unless explicitly requested.

After this skill lands, useful metrics are:

```text
packets_drained_per_frame
packets_left_queued_by_lane
packets_routed_by_type
presentation_fanout_count_per_frame
world_bullet_count
projectile_node_count
pending_bullet_update_count
fanout_time_msec
```

Success criteria for later metrics:

```text
presentation_fanout_count_per_frame <= 1
drained_packets_per_frame <= MAX_PACKETS_PER_POLL
pending_bullet_update_count does not grow without bound
frame time stabilizes under high bullet stress
```

## Anti-patterns

Do not:

* Create bullets from unknown `bullet_delta` updates.
* Ignore unknown bullet hot updates without buffering.
* Store all pending updates forever.
* Forget to clear pending bullet updates on delete/full-world reset.
* Fan out world presentation once per packet.
* Add packet dropping in this slice.
* Change WebRTC channel reliability or ordering in this slice.
* Add broad generic entity buffering before bullet behavior is fixed.
* Add wrappers or abstract presentation layers.
* Move gameplay simulation ownership.
* Change server protocol packet ownership.
* Change bullet creates/deletes to the bullet hot lane.
