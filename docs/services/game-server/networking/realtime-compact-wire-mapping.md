# Realtime Compact Wire Mapping

Parent index: [Game Server Networking](./!INDEX.md)

This file is hand-authored because it defines the compact wire alias contract.
It is not generated from packet SSoT.
Do not reconcile compact aliases from raw packet struct names.

## Implementation Rule

- Internal server structs keep readable field names.
- `WireLanePacket` keeps producing readable long-key maps.
- `CompactWirePacket` owns the final outbound alias conversion at the encode boundary.
- The client expands compact packets back to readable long-key dictionaries before existing lane appliers process them.
- Both compact and legacy long-key packets must remain accepted by the client during this transition.
- Aliases must be globally unambiguous so the client can recursively expand compact keys without needing entity-kind context.
- Compacting field names is separate from omitting empty delta sections.
- Sparse delta omission happens before compact aliases are applied.

## Runtime Metadata Inference

For active realtime world, overlay, and session packet families, the preferred outbound wire shape now omits runtime metadata the client can infer.

The client derives:

- `lane` from packet `type`.
- `snapshot_kind` from packet `type`: `_full` -> `full`, `_delta` -> `delta`.
- `snapshot_id` from lane, packet kind, and sequence.
  Full packets infer `<lane>-baseline-<sequence>`.
  Delta packets infer `<lane>-snapshot-<sequence>`.
- Full-packet `baseline_id` from lane and sequence as `<lane>-baseline-<sequence>`.
- Delta-packet `baseline_id` from readable `baseline_sequence` or compact `bq` as `<lane>-baseline-<baseline_sequence>`.
- `is_final_chunk` from `chunk_index` and `chunk_count`.

For active realtime world, overlay, and session packet families, the server now emits `chunk_index` and `chunk_count` only when `chunk_count > 1`.
When those fields are absent, the client treats the packet as a single final chunk.

This inference rule is runtime-lane specific. It does not imply compacted event/control packet behavior beyond what the current implementation actually emits.

## Compact Packet Type Values

- `world_full` -> `wf`
- `world_delta` -> `wd`
- `overlay_full` -> `of`
- `overlay_delta` -> `od`
- `session_full` -> `sf`
- `session_delta` -> `sd`

## Compact Lane Values

These remain documented for backward-compatible decode support and for non-runtime contexts where `lane` is still explicitly present. They are not the preferred active runtime output for world/overlay/session gameplay lanes.

- `world` -> `w`
- `overlay` -> `o`
- `session` -> `s`

## Compact Snapshot Kind Values

These remain documented for backward-compatible decode support and for contexts where `snapshot_kind` is still explicitly present. They are not the preferred active runtime output for world/overlay/session gameplay lanes.

- `full` -> `f`
- `delta` -> `d`

## Metadata Keys

Preferred active runtime output for world/overlay/session gameplay lanes:

- `type` -> `t`
- `sequence` -> `q`
- `baseline_sequence` -> `bq`
- `server_sent_msec` -> `ms`
- `chunk_index` -> `ci` when `chunk_count > 1`
- `chunk_count` -> `cc` when `chunk_count > 1`

Legacy or backward-compatible decode support still accepted by the client:

- `lane` -> `l`
- `baseline_id` -> `b`
- `snapshot_id` -> `sid`
- `snapshot_kind` -> `k`
- `is_final_chunk` -> `fc`

For active runtime world/overlay/session packets, `baseline_id` is only emitted when a delta packet cannot express its dependency as a numeric baseline sequence or when full-packet metadata cannot be represented by the current inferred format safely.

## World Delta Section Keys

- `ship_creates` -> `sc`
- `ship_updates` -> `su`
- `ship_deletes` -> `sx`
- `bullet_creates` -> `bc`
- `bullet_updates` -> `bu`
- `bullet_deletes` -> `bx`
- `asteroid_creates` -> `ac`
- `asteroid_updates` -> `au`
- `asteroid_deletes` -> `ax`
- `pickup_creates` -> `pc`
- `pickup_updates` -> `pu`
- `pickup_deletes` -> `px`

## Overlay Delta Section Keys

- `receiver_creates` -> `rc`
- `receiver_updates` -> `ru`
- `receiver_deletes` -> `rx`

## Session Delta Section Keys

- `players` -> `pl`
- `player_session_updates` -> `psu`
- `player_session_deletes` -> `psx`
- `player_lifecycle` -> `plc`
- `player_lifecycle_updates` -> `plu`
- `player_lifecycle_deletes` -> `plx`
- `total_asteroids` -> `ta`

Delta section aliases are only present when the corresponding readable delta section is present.
Missing compact delta section aliases mean empty or no-op.

## Shared Record Keys

- `id` -> `i`
- `player_id` -> `pid`
- `self_id` -> `self`
- `type` -> `t`
- `status` -> `stat`
- `x` -> `x`
- `y` -> `y`
- `rotation` -> `r`
- `health` -> `h`
- `score` -> `sco`
- `lives` -> `lv`
- `respawn_cooldown` -> `rcd`

## World Record Keys

- `ship_type` -> `st`
- `shields` -> `sh`
- `thrusting` -> `th`
- `target_kind` -> `tk`
- `target_id` -> `tid`
- `owner_id` -> `oi`
- `weapon_id` -> `wid`
- `projectile_type` -> `pt`
- `size` -> `sz`
- `scale` -> `sl`
- `variant` -> `v`
- `pickup_class` -> `pcl`
- `age_seconds` -> `age`
- `lifespan_seconds` -> `life`

## Overlay And Session Weapon And Loadout Keys

- `primary_weapon_id` -> `pwid`
- `primary_ammo_policy` -> `pap`
- `primary_cooldown_remaining` -> `pcr`
- `primary_ammo_remaining` -> `par`
- `secondary_weapon_id` -> `swid`
- `secondary_ammo_policy` -> `sap`
- `secondary_cooldown_remaining` -> `scr`
- `secondary_ammo_remaining` -> `sar`
- `spawn_x` -> `spx`
- `spawn_y` -> `spy`

## Implemented Boundary

- Server readable lane maps are still built by `WireLanePacket`.
- `CompactWirePacket` applies aliases only at the final outbound encode boundary.
- Active outbound compacting currently applies to world, overlay, and session realtime gameplay lanes.
- event_batch and control-lane resync packet families are not compacted in this pass unless implementation changes.
- `PacketCodec.decode` performs the first compact expansion before packet envelope validation. `RealtimeRouter` may defensively normalize already-expanded packets, but it is not the first decode boundary.
- Legacy long-key packets remain accepted during the transition.
- Empty delta section omission is implemented by the readable delta serializers before CompactWirePacket applies aliases. CompactWirePacket only aliases keys that remain present. The current generated control-lane recovery packet families are resync_request and resync_required; there is no separate generated packet family named control.

## Code Paths

- `services/game-server/internal/protocol/realtime/wire_packets.go`
- `services/game-server/internal/protocol/realtime/compact_wire_packet.go`
- `services/game-server/internal/protocol/realtime/active.go`
- `client/scripts/networking/packets/packet_codec.gd`
- `client/scripts/protocol/realtime/compact_lane_packet.gd`
- `client/scripts/protocol/realtime/realtime_router.gd` - Defensive/idempotent normalization after decode, if still present in implementation.

## Tests

- `services/game-server/internal/protocol/realtime/compact_wire_packet_test.go`
- `client/tests/unit/protocol/realtime/test_compact_lane_packet.gd`
- PacketCodec compact decode coverage in `client/tests/unit/test_packet_codec.gd`

## Observed Development Run

Recent compact three-lane observed development samples include quantization, sparse delta omission, and compact aliases. The sample sizes are approximately:

- sparse three-lane sample: `world_delta` ~412 bytes, `overlay_delta` ~135 bytes, `session_delta` ~132 bytes, total ~679 bytes/tick
- sparse world-only sample: ~577-587 bytes/tick
- sparse 8-player world-only sample: ~3.1-3.6 KB/tick

These are observed development samples, not guaranteed budgets.