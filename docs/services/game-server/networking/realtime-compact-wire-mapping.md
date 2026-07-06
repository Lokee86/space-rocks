# Realtime Compact Wire Mapping

Parent index: [Game Server Networking](./!INDEX.md)

This file is hand-authored because it defines the compact wire alias contract.
It is not generated from packet SSoT.
Do not reconcile compact aliases from raw packet struct names.

## Implementation Rule

- Internal server structs keep readable field names.
- `WireLanePacket` keeps producing readable long-key maps.
- `CompactWirePacket` owns the final outbound compact-key conversion, compact-value conversion, shared ID compaction, and tuple packing at the encode boundary.
- The client compact lane packet expands compact wire data back to readable long-key dictionaries before existing lane appliers process them.
- The client expands tuple packets back into readable dictionaries before existing lane appliers process them.
- For tuple-packed records, the client rehydrates compact numeric IDs back into readable string IDs before world, session, and event appliers process them.
- Both compact and legacy long-key packets must remain accepted by the client during this transition.
- Aliases must be globally unambiguous so the client can recursively expand compact keys without needing entity-kind context.
- Compacting field names is separate from omitting empty delta sections.
- Sparse delta omission happens before compact aliases are applied.
## Shared ID Compaction

Shared ID compaction applies to tuple-packed records and compact value slots when the field position or context determines the prefix safely.

The current compact tuple ID rule is:

- bare numeric suffix when tuple context determines the prefix
- tagged compact ID when the prefix is known but not tuple-determined
- original string when the prefix is unknown or the suffix is malformed

Server-side bare numeric helpers currently apply to:

- `asteroid-N` -> `N`
- `bullet-N` -> `N`
- `player-N` -> `N`
- `pickup-N` -> `N`
- `ship-N` -> `N`
- `table-N` -> `N`
- `presentation-event-N` -> `N`
- `event-batch-N` -> `N`

Tagged compact IDs currently use these shapes:

- `["p", N]` -> `player-N`
- `["b", N]` -> `bullet-N`
- `["a", N]` -> `asteroid-N`
- `["pk", N]` -> `pickup-N`
- `["s", N]` -> `ship-N`
- `["tbl", N]` -> `table-N`
- `["pe", N]` -> `presentation-event-N`
- `["eb", N]` -> `event-batch-N`

The compacting helper only removes or tags a prefix when the numeric suffix is valid.
Non-string IDs remain unchanged.
Unknown prefixes and malformed suffixes remain unchanged.
String fields that are only loosely associated with an entity or record keep their readable string form unless a tuple slot has a safe specific context.

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
- `asteroid_delta` -> `ad`
- `bullet_delta` -> `bd`
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

`world_delta.bu` and `world_delta.au` remain compact aliases for serializer compatibility, bootstrap, lifecycle, and resync-safe world deltas. In the active subtractive hot-lane path, regular bullet and asteroid movement updates are removed from `world_delta` and emitted as `bullet_delta.bu` and `asteroid_delta.au`.

## Dedicated Hot Movement Delta Section Keys

- `asteroid_delta.asteroid_updates` -> `au`
- `bullet_delta.bullet_updates` -> `bu`

## Asteroid Tuple Mapping

This contract applies to world lane asteroid records and dedicated asteroid movement deltas.
`world_full.asteroids` uses tuple records.
`world_delta.ac` uses tuple records for asteroid creates.
`asteroid_delta.au` uses tuple records for regular asteroid movement updates.
`world_delta.au` remains supported for compatibility and resync-safe world deltas, but regular active movement updates are split to `asteroid_delta`.
`world_delta.ax` uses numeric suffix IDs for asteroid deletes.

Tuple order for full/create asteroid records:

- `[id_number, x, y, size, health, scale, variant]`

Tuple order for update asteroid records:

- `[id_number, x, y]`

Tuple order for x-only updates:

- `[id_number, x]`

Y-only updates use:

- `[id_number, null, y]`

Identity-only updates, when no x/y is present, use:

- `[id_number]`

Deletes use:

- `ax: [id_number, id_number]`

Server ID compaction converts `"asteroid-123"` to JSON number `123`.
This is not string `"123"`.
The byte saving is 11 bytes per occurrence from removing the 9-byte `asteroid-` prefix plus 2 JSON quote bytes.
Malformed asteroid IDs such as `"asteroid-not-a-number"` remain unchanged.
Non-string IDs are left unchanged by the server helper.

Client ID rehydration converts int `123` to `"asteroid-123"`.
JSON-decoded whole-number float `123.0` also becomes `"asteroid-123"`.
String suffix `"123"` becomes `"asteroid-123"` for transition tolerance.
Already-full `"asteroid-123"` remains unchanged.
Non-integer float values and unsupported values remain unchanged.

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
These keys are still applied where the wire slot remains a map field.
Tuple-packed slots can bypass these readable key aliases when their position already carries the meaning safely.

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
Polymorphic or broadly shared string fields remain strings unless the tuple slot has safe specific context.
Tuple slots with safe context now compact and rehydrate these IDs:

- `owner_id`
- `target_id`
- `source_id`
- `pickup_id`
- `table_id`

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

## Event Batch Compact Mapping

`event_batch` is now compacted through the realtime compact wire path.
It is the presentation-event lane, not a state lane.
It keeps batching: one ordered batch can contain multiple pending presentation events.
It does not become one packet per event.

Known event records are tuple-shaped for the compact wire path.\r\nKnown event records no longer use broad reflected `EventState` output.\r\nUnknown or newly added event types may still fall back to legacy long-key reflected output for compatibility until they are explicitly shaped for compact sparse output.\r\n\r\nUnknown map-shaped event records remain compatibility fallback records and should not be forced into tuple arrays just because they have compact aliases. Known event records are tuple arrays; unknown event records stay map shaped.\r\n\r\n`sd` is not used for `ship_death` because `sd` is already reserved for `session_delta`.

### Canonical Event Batch Alias Tables

#### Event Batch Envelope

| Readable/logical field name | Compact wire key/value | Applies where |
| --- | --- | --- |
| `type = event_batch` | `t = eb` | event_batch envelope |
| `sequence` | `q` | event_batch envelope |
| `server_sent_msec` | `ms` | event_batch envelope |
| `batch_id` | `bid` | event_batch envelope |
| `events` | `ev` | event_batch envelope |

#### Nested Event Record Keys

| Readable/logical field name | Compact wire key/value | Applies where |
| --- | --- | --- |
| `type` | `t` | nested event record |
| `event_id` | `ei` | nested event record |
| `player_id` | `pid` | nested event record |
| `target_id` | `tid` | nested event record |
| `lives` | `lv` | nested event record |
| `x` | `x` | nested event record |
| `y` | `y` | nested event record |
| `source_type` | `srct` | nested event record |
| `source_id` | `src` | nested event record |
| `target_type` | `tt` | nested event record |
| `damage_type` | `dt` | nested event record |
| `damage_cause` | `dc` | nested event record |
| `base_amount` | `ba` | nested event record |
| `modified_amount` | `ma` | nested event record |
| `applied_to_health` | `ah` | nested event record |
| `absorbed_by_shield` | `abs` | nested event record |
| `remaining_health` | `rh` | nested event record |
| `remaining_shield` | `rs` | nested event record |
| `pickup_id` | `pkid` | nested event record |
| `pickup_type` | `pkt` | nested event record |
| `table_id` | `tbl` | nested event record |
| `lives_after` | `lva` | nested event record |
| `effect_type` | `fx` | nested event record |
| `amount` | `amt` | nested event record |
| `respawn_delay` | `rd` | nested event record |

`x` and `y` remain `x` and `y` after compact encoding, but they are quantized on the wire.

#### Event Type Values

| Readable/logical field name | Compact wire key/value | Applies where |
| --- | --- | --- |
| `type = bullet_blast` | `t = bb` | event type value |
| `type = ship_death` | `t = shd` | event type value |
| `type = radial_effect_started` | `t = rfx` | event type value |
| `type = pickup_collected` | `t = pcol` | event type value |
| `type = pickup_effect_applied` | `t = pea` | event type value |
| `type = pickup_expired` | `t = pexp` | event type value |
| `type = pickup_dropped` | `t = pdr` | event type value |
| `type = damage_applied` | `t = dmg` | event type value |
| `type = damage_over_time_started` | `t = dots` | event type value |
| `type = damage_over_time_tick` | `t = dott` | event type value |

## Bullet Tuple Mapping

This contract applies to world lane bullet records and dedicated bullet movement deltas.
`world_full.bullets` uses tuple records.
`world_delta.bc` uses tuple records for bullet creates.
`bullet_delta.bu` uses tuple records for regular bullet movement updates.
`world_delta.bu` remains supported for compatibility and resync-safe world deltas, but regular active movement updates are split to `bullet_delta`.
`world_delta.bx` uses compact numeric delete IDs.

| Record family | Tuple shape |
| --- | --- |
| `world_full.bullets` | `[id, owner_id, x, y, rotation, weapon_id, projectile_type]` |
| `world_delta.bullet_creates` | `[id, owner_id, x, y, rotation, weapon_id, projectile_type]` |
| `bullet_delta.bullet_updates` | `[id, x, y, rotation]` |
| `world_delta.bullet_updates` | `[id, x, y, rotation]` for compatibility/resync-safe world deltas only |
| `world_delta.bullet_deletes` | `[id]` |

Sparse placeholder rules:

- Missing trailing fields are omitted.
- Missing middle fields use `null` placeholders.
- Zero values are preserved.
- `false` booleans are preserved.

Deletes use compact numeric IDs such as `bullet-123 -> 123`.
Malformed bullet IDs remain unchanged.

## World Ship/Player Tuple Mapping

This contract applies to world lane ship and player records.
`world_full.ships` uses tuple records.
`world_delta.sc` uses tuple records for ship creates.
`world_delta.su` uses tuple records for ship updates.
`world_delta.sx` uses compact numeric delete IDs.

| Record family | Tuple shape |
| --- | --- |
| `world_full.ships` | `[id, ship_type, x, y, rotation, health, shields, thrusting, target_kind, target_id]` |
| `world_delta.ship_creates` | `[id, ship_type, x, y, rotation, health, shields, thrusting, target_kind, target_id]` |
| `world_delta.ship_updates` | `[id, x, y, rotation, thrusting]` |
| `world_delta.ship_deletes` | `[id]` |

Sparse placeholder rules:

- Missing trailing fields are omitted.
- Missing middle fields use `null` placeholders.
- Zero values and `false` booleans are preserved.

Deletes use compact player IDs when the ship ID is player-scoped.
Legacy non-player ship IDs remain unchanged.

## Session Player Tuple Mapping

This contract applies to session lane player and lifecycle records.
`session_full.pl` and `session_delta.pl` use tuple records for session players.
`session_delta.psu` uses pair tuples for partial session player updates.
`session_delta.psx` uses compact numeric delete IDs.
`session_full.plc`, `session_delta.plc`, `session_delta.plu`, and `session_delta.plx` cover lifecycle records.

| Record family | Tuple shape |
| --- | --- |
| `session_full.pl` | `[id, ship_type, score, lives, respawn_cooldown, primary_weapon_id, primary_ammo_policy, secondary_weapon_id, secondary_ammo_policy, spawn_x, spawn_y]` |
| `session_delta.pl` | `[id, ship_type, score, lives, respawn_cooldown, primary_weapon_id, primary_ammo_policy, secondary_weapon_id, secondary_ammo_policy, spawn_x, spawn_y]` |
| `session_delta.player_session_updates` | `[id, field_alias, value, field_alias, value]` |
| `session_delta.player_session_deletes` | `[id]` |
| `session_full.plc` | `[player_id, status]` |
| `session_delta.plc` | `[player_id, status]` |
| `session_delta.plu` | `[player_id, status]` |
| `session_delta.plx` | `[player_id]` |

Sparse placeholder rules:

- Missing trailing fields are omitted.
- Missing middle fields use `null` placeholders.
- Zero values and `false` booleans are preserved.

Deletes use compact player IDs.

## Event Tuple Mapping

This contract applies to `event_batch` records.
`event_batch.ev` uses tuple records for known event types.
Known event tuples expand back into readable dictionaries on the client before event appliers run.
Legacy map-shaped or long-key event records remain accepted during the transition.

| Event type | Tuple shape |
| --- | --- |
| `bullet_blast` | `[type, event_id, x, y]` |
| `ship_death` | `[type, event_id, player_id, lives, respawn_delay, x, y]` |
| `damage_applied` | `[type, event_id, source_type, source_id, effect_type, amount, x, y]` |
| `damage_over_time_started` | `[type, event_id, source_type, source_id, effect_type, amount]` |
| `damage_over_time_tick` | `[type, event_id, source_type, source_id, effect_type, amount, x, y]` |
| `radial_effect_started` | `[type, event_id, source_type, source_id, effect_type, x, y]` |
| `pickup_collected` | `[type, event_id, player_id, pickup_id, pickup_type, x, y]` |
| `pickup_effect_applied` | `[type, event_id, player_id, pickup_id, pickup_type, effect_type, amount, lives_after]` |
| `pickup_expired` | `[type, event_id, pickup_id, pickup_type, x, y]` |
| `pickup_dropped` | `[type, event_id, pickup_id, pickup_type, source_type, source_id, table_id, x, y]` |

Sparse placeholder rules:

- Missing trailing fields are omitted.
- Missing middle fields use `null` placeholders.
- Zero values and `false` booleans are preserved.

Known tuple event fields use compact presentation-event IDs and compact player IDs where those slots are safe and specific. `event_id` rehydrates to `presentation-event-N` on the client.
`source_id`, `target_id`, `owner_id`, `pickup_id`, and `table_id` compact when a known tuple context can determine the prefix, and they rehydrate back to full string IDs on the client.

### Event Batch Example

Readable logical event_batch example:

```json
{
  "type": "event_batch",
  "sequence": 412,
  "server_sent_msec": 1712345678901,
  "batch_id": "event-batch-412",
  "events": [
    {
      "type": "ship_death",
      "event_id": "presentation-event-100",
      "player_id": "Player-2",
      "lives": 2,
      "respawn_delay": 3,
      "x": 512,
      "y": 384
    },
    {
      "type": "damage_applied",
      "event_id": "presentation-event-101",
      "source_type": "pickup",
      "source_id": "pickup-4",
      "effect_type": "impact",
      "amount": 20,
      "x": 512,
      "y": 384
    }
  ]
}
```

Compact wire event_batch example:

```json
{
  "t": "eb",
  "q": 412,
  "ms": 1712345678901,
  "bid": 412,
  "ev": [
    {
      "t": "shd",
      "ei": 100,
      "pid": 2,
      "lv": 2,
      "rd": 3,
      "x": 512,
      "y": 384
    },
    {
      "t": "dmg",
      "ei": 101,
      "srct": "pickup",
      "src": 4,
      "fx": "impact",
      "amt": 20,
      "x": 512,
      "y": 384
    }
  ]
}
```

Readable/logical docs may show expanded names, while runtime wire sends compact aliases. Domain logs may still show raw x/y before projection.
After client expansion, readable/logical packets still use full string IDs.
The current implementation uses tuple arrays for known compact event records.
The current implementation does not use binary encoding for events.

## Implemented Boundary

- Server readable lane maps are still built by `WireLanePacket`.
- `CompactWirePacket` applies compact keys, compact values, shared ID compaction, and tuple packing only at the final outbound encode boundary.
- Active outbound compacting currently applies to world, asteroid, bullet, overlay, session, and `event_batch` realtime packet families.
- Generated control-lane resync packet families are not compacted in this pass unless implementation changes.
- `PacketCodec.decode` performs the first compact expansion before packet envelope validation. `RealtimeRouter` may defensively normalize already-expanded packets, but it is not the first decode boundary.
- Legacy long-key packets remain accepted during the transition.
- Empty delta section omission is implemented by the readable delta serializers before CompactWirePacket applies aliases. CompactWirePacket only aliases keys that remain present. The current generated control-lane recovery packet families are resync_request and resync_required; there is no separate generated packet family named control.

## Code Paths

- `services/game-server/internal/protocol/realtime/wire_packets.go`
- `services/game-server/internal/protocol/realtime/compact_wire_packet.go`
- `services/game-server/internal/protocol/realtime/compact_wire_asteroids.go`
- `services/game-server/internal/protocol/realtime/compact_wire_bullets.go`
- `services/game-server/internal/protocol/realtime/compact_wire_ships.go`
- `services/game-server/internal/protocol/realtime/compact_wire_players.go`
- `services/game-server/internal/protocol/realtime/compact_wire_events.go`
- `services/game-server/internal/protocol/realtime/active.go`
- `client/scripts/networking/packets/packet_codec.gd`
- `client/scripts/protocol/realtime/compact_lane_packet.gd`
- `client/scripts/protocol/realtime/realtime_router.gd` - Defensive/idempotent normalization after decode, if still present in implementation.

## Tests

- `services/game-server/internal/protocol/realtime/compact_wire_packet_test.go`
- `services/game-server/internal/protocol/realtime/active_test.go`
- `client/tests/unit/protocol/realtime/test_compact_lane_packet.gd`
- `client/tests/unit/protocol/realtime/test_world_lane_applier.gd`
- PacketCodec compact decode coverage in `client/tests/unit/test_packet_codec.gd`



