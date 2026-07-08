## Purpose

Use this skill to restore documentation consistency after implementing dedicated asteroid and bullet lifecycle lanes.

This documentation pass treats the current implementation as the full implementation of:

```text
sr.asteroids.lifecycle
sr.bullets.lifecycle
```

The goal is to eliminate stale documentation that still says asteroid and bullet lifecycle creates/deletes remain on `sr.world`.

## Source of truth

Treat the implementation diff as the current source of truth.

The implemented model is:

```text
sr.asteroids.lifecycle
= reliable / ordered / required / critical
= asteroids_lifecycle
= asteroid_creates + asteroid_deletes

sr.bullets.lifecycle
= reliable / ordered / required / critical
= bullets_lifecycle
= bullet_creates + bullet_deletes

sr.asteroids
= unreliable / unordered / hot-supersedable
= asteroid_delta
= asteroid_updates only

sr.bullets
= unreliable / unordered / hot-supersedable
= bullet_delta
= bullet_updates only

sr.world
= reliable / ordered
= ships, pickups, player/world presentation state, bootstrap/full snapshots, compatibility/resync-safe support where still implemented
= no longer the active owner for asteroid/bullet lifecycle creates/deletes
```

Core rule:

```text
Lifecycle defines existence. Hot lanes update known entities only.
```

## Scope

Update existing documentation only unless an index or documentation policy explicitly requires a new file.

No new documentation file is expected.

No document deletions are expected.

This pass should update protocol, transport, outbound routing, inbound routing, lane projection, compact wire mapping, client world sync, entity docs, devtools docs, and related planning summaries where stale references remain.

## Required documentation statements

Use this statement, or equivalent wording, wherever the lane model is summarized:

```text
Entity lifecycle ownership is split by entity family. The world lane owns player, pickup, world, and full/bootstrap presentation state. Asteroid lifecycle packets use sr.asteroids.lifecycle. Bullet/projectile lifecycle packets use sr.bullets.lifecycle. Hot asteroid and bullet lanes are unreliable movement/update lanes only and must not create entities implicitly.
```

Use this statement, or equivalent wording, wherever cross-lane ordering is described:

```text
Cross-lane ordering is not guaranteed between reliable lifecycle lanes and unreliable hot lanes. Clients must tolerate hot updates arriving before lifecycle create packets and after lifecycle delete packets.
```

Use this statement where client hot-lane behavior is described:

```text
Hot movement packets must never create entities. Unknown hot asteroid updates are ignored. Unknown hot bullet updates are buffered only where the client explicitly supports waiting for lifecycle create; hot updates after delete are ignored and must not resurrect removed entities.
```

## Remove stale claims

Remove or correct docs that say:

```text
lifecycle creates/deletes remain on sr.world
world lane owns asteroid/bullet lifecycle creates/deletes
hot lanes create entities
asteroid_delta/bullet_delta are lifecycle delivery
all entity lifecycle traffic is world-lane traffic
```

Remaining `world lane asteroid records` or `world lane bullet records` references must either be removed or clearly limited to full/bootstrap/compatibility context.

## Primary stale files

Update these first.

### docs/protocol/realtime-webrtc-gameplay-transport.md

Update the physical channel table to eight channels:

```text
logical lane             | physical channel         | negotiated id | packet families
world                    | sr.world                 | 1             | world_full, world_delta
overlay                  | sr.overlay               | 2             | overlay_full, overlay_delta
session                  | sr.session               | 3             | session_full, session_delta
event                    | sr.event                 | 4             | event_batch
asteroids                | sr.asteroids             | 5             | asteroid_delta
bullets                  | sr.bullets               | 6             | bullet_delta
asteroids.lifecycle      | sr.asteroids.lifecycle   | 7             | asteroids_lifecycle
bullets.lifecycle        | sr.bullets.lifecycle     | 8             | bullets_lifecycle
```

Replace the channel policy with:

```text
sr.world, sr.overlay, sr.session, sr.event, sr.asteroids.lifecycle, and sr.bullets.lifecycle are negotiated ordered/reliable channels.
sr.asteroids and sr.bullets are negotiated unordered/unreliable channels with maxRetransmits=0.
sr.asteroids carries supersedable asteroid_updates only.
sr.bullets carries supersedable bullet_updates only.
sr.asteroids.lifecycle carries asteroid_creates and asteroid_deletes.
sr.bullets.lifecycle carries bullet_creates and bullet_deletes.
Lifecycle channels define entity existence. Hot lanes update known existing entities only.
```

Add the cross-lane race note:

```text
Reliable lifecycle lanes and unreliable hot lanes do not provide global ordering across lanes. The client must tolerate hot updates arriving before lifecycle creates and after lifecycle deletes. Hot packets must not create missing entities or resurrect deleted entities.
```

Update send and receive boundary descriptions so lifecycle packets flow through the same WebRTC/DataChannel path.

### docs/protocol/realtime-websocket-protocol.md

Update overview text so current active gameplay WebRTC channels include:

```text
sr.world
sr.overlay
sr.session
sr.event
sr.asteroids
sr.bullets
sr.asteroids.lifecycle
sr.bullets.lifecycle
```

Update current gameplay lanes list to include:

```text
asteroids.lifecycle
bullets.lifecycle
```

Update active packet families to include:

```text
asteroids_lifecycle
bullets_lifecycle
```

Update compact/envelope documentation to include:

```text
asteroids_lifecycle -> al
bullets_lifecycle -> bl
asteroids.lifecycle -> al
bullets.lifecycle -> bl
```

Update field-delta semantics to include:

```text
asteroids_lifecycle asteroid_creates/asteroid_deletes = id
bullets_lifecycle bullet_creates/bullet_deletes = id
asteroid_delta asteroid_updates = id
bullet_delta bullet_updates = id
```

Update the world lane section:

* Remove active `world_delta` ownership of asteroid/bullet creates/deletes.
* Keep compatibility wording only if framed as legacy/bootstrap/resync-safe support.
* Do not say active lifecycle remains on `sr.world`.

Add dedicated sections:

```text
### Asteroid lifecycle lane packets
### Bullet lifecycle lane packets
```

Each section must define:

* Creates/deletes.
* Reliable ordered delivery.
* Required/critical scheduling.
* Lifecycle defines entity existence.
* Hot lanes update known entities only.
* Cross-lane ordering is not global.

Update scheduling:

```text
asteroids_lifecycle and bullets_lifecycle = required / critical
asteroid_delta and bullet_delta = hot-supersedable / high priority
```

State that lifecycle lanes are not hot-supersedable and are not split by hot-lane chunking.

Update client inbound routing:

```text
asteroids_lifecycle_received
bullets_lifecycle_received
```

Add that `RealtimeRouter` handles these through:

```text
WorldLaneApplier.apply_asteroids_lifecycle
WorldLaneApplier.apply_bullets_lifecycle
```

Update client-recognized inbound packet types to include both lifecycle packet families.

Update delivery/failure semantics so ordered/reliable active gameplay lanes include:

```text
sr.world
sr.overlay
sr.session
sr.event
sr.asteroids.lifecycle
sr.bullets.lifecycle
```

### docs/protocol/gameplay-packets.md

Update active server-to-client gameplay packet families:

```text
world_full / world_delta
asteroids_lifecycle
bullets_lifecycle
asteroid_delta
bullet_delta
overlay_full / overlay_delta
session_full / session_delta
event_batch
player_pause_state
resync_request / resync_required
```

Replace lane ownership with:

```text
world lane
= ships, pickups, player/world presentation state, and full/bootstrap world snapshots

asteroids lifecycle lane
= asteroid creates/deletes and initial asteroid presentation identity, including variant/size/scale

bullets lifecycle lane
= bullet/projectile creates/deletes and initial projectile identity, including owner, weapon_id, projectile_type, and torpedo identity

asteroids hot lane
= regular asteroid movement updates on unordered/unreliable sr.asteroids

bullets hot lane
= regular bullet/projectile movement updates on unordered/unreliable sr.bullets
```

Add the hot packet race rule:

```text
Hot movement packets must never create entities. Unknown hot asteroid updates are ignored. Unknown hot bullet updates are buffered only where the client explicitly supports waiting for lifecycle create; hot updates after delete are ignored and must not resurrect removed entities.
```

### docs/domains/technical/realtime-client-server-flow.md

Update packet family summaries to include:

```text
world_full/world_delta
asteroids_lifecycle
bullets_lifecycle
asteroid_delta
bullet_delta
overlay_full/overlay_delta
session_full/session_delta
event_batch
```

Replace stale world lifecycle wording with:

```text
Asteroid and bullet lifecycle creates/deletes are subtractively removed from active world_delta and emitted as reliable ordered lifecycle packets on sr.asteroids.lifecycle and sr.bullets.lifecycle. Dedicated unordered hot lanes carry movement updates only.
```

Update lane policy:

```text
ordered/reliable:
sr.world
sr.overlay
sr.session
sr.event
sr.asteroids.lifecycle
sr.bullets.lifecycle

unordered/unreliable:
sr.asteroids
sr.bullets
```

Update client routing lists to include:

```text
asteroids_lifecycle
bullets_lifecycle
```

### docs/services/client/networking-flow/inbound-packet-routing.md

Update overview channel list to include lifecycle DataChannels.

Add classified inbound packet types:

```text
asteroids_lifecycle
bullets_lifecycle
```

Add dispatcher outputs:

```text
asteroids_lifecycle_received(packet)
bullets_lifecycle_received(packet)
```

Update the connection-service gameplay packet path:

```text
ServerPacketDispatcher
-> ClientConnectionService._route_gameplay_packet
-> RealtimeRouter.route_lane_packet
-> WorldLaneApplier lifecycle method
-> gameplay_packet_received
```

Update current inbound routes so lifecycle packets are included with other lane packets.

In gameplay handoff, state that by the time `GameplaySessionController` receives the packet, lifecycle packet state has already been applied to `WorldLaneState`.

Add relevant tests:

```text
client/tests/unit/networking/test_server_packet_dispatcher.gd
client/tests/unit/world/test_projectile_sync.gd
client/tests/unit/world/test_asteroid_sync.gd
client/tests/unit/world/test_world_sync.gd
```

### docs/services/game-server/networking/outbound-message-flow.md

Update physical channel overview so current active gameplay channels include:

```text
sr.world
sr.overlay
sr.session
sr.event
sr.asteroids
sr.bullets
sr.asteroids.lifecycle
sr.bullets.lifecycle
```

Update responsibilities to include reliable/ordered WebRTC delivery for lifecycle lanes.

Update active realtime lane packets list:

```text
world_full
world_delta
asteroids_lifecycle
bullets_lifecycle
asteroid_delta
bullet_delta
overlay_full
overlay_delta
session_full
session_delta
event_batch
resync_request
resync_required
```

Update lane roles:

```text
world = ships, pickups, world/match presentation state, full/bootstrap snapshots
asteroids.lifecycle = asteroid creates/deletes
bullets.lifecycle = bullet/projectile creates/deletes
asteroids = asteroid movement updates
bullets = bullet movement updates
```

Replace any stale lifecycle wording with:

```text
Lifecycle creates/deletes use required/critical reliable lifecycle lanes.
```

Update ticker-driven active lane write descriptions to mention that planner can emit lifecycle candidates before hot movement candidates.

Update observability so `lane protocol gameplay wire packet written` may show:

```text
sr.asteroids.lifecycle
sr.bullets.lifecycle
```

But chunked hot-lane logs only apply to:

```text
sr.asteroids
sr.bullets
```

Ensure the code map includes:

```text
services/game-server/internal/protocol/realtime/lanes.go
services/game-server/internal/protocol/realtime/planner.go
services/game-server/internal/protocol/realtime/wire_packets.go
services/game-server/internal/protocol/realtime/compact_wire_packet.go
services/game-server/internal/networking/webrtc_transport.go
```

### docs/services/game-server/simulation/runtime/lane-packet-projection.md

Update active flow:

```text
authoritative game state
-> realtime projection / planning
-> raw lane records
-> numeric wire quantization into wire-shaped records
-> delta comparison
-> split asteroid/bullet movement to hot lanes
-> split asteroid/bullet creates/deletes to reliable lifecycle lanes
-> sparse readable wire-map serialization
-> compact alias mapping
-> packetcodec JSON encoding
-> WebRTC gameplay lane write using per-lane reliability policy
```

Update lane ownership:

```text
world lane
= ships, pickups, player/world presentation state, and full/bootstrap world snapshots

asteroids.lifecycle lane
= asteroid creates/deletes

bullets.lifecycle lane
= bullet/projectile creates/deletes

asteroids lane
= regular asteroid movement updates

bullets lane
= regular bullet/projectile movement updates
```

Remove or replace:

```text
Asteroid and bullet creates/deletes remain on the world lane.
```

### docs/services/game-server/networking/realtime-compact-wire-mapping.md

Add compact packet type values:

```text
asteroids_lifecycle -> al
bullets_lifecycle -> bl
```

Add compact lane values:

```text
asteroids.lifecycle -> al
bullets.lifecycle -> bl
```

Add lifecycle delta section keys:

```text
asteroids_lifecycle.asteroid_creates -> ac
asteroids_lifecycle.asteroid_deletes -> ax
bullets_lifecycle.bullet_creates -> bc
bullets_lifecycle.bullet_deletes -> bx
```

Update asteroid tuple mapping:

```text
world_full.asteroids uses full tuple records.
asteroids_lifecycle.ac uses tuple records for asteroid creates.
asteroid_delta.au uses tuple records for regular movement updates.
asteroids_lifecycle.ax uses compact numeric IDs for deletes.
world_delta.au remains compatibility/resync-safe support only.
```

Update bullet tuple mapping:

```text
world_full.bullets uses full tuple records.
bullets_lifecycle.bc uses tuple records for bullet/projectile creates.
bullet_delta.bu uses tuple records for regular movement updates.
bullets_lifecycle.bx uses compact numeric IDs for deletes.
world_delta.bu remains compatibility/resync-safe support only.
```

Keep existing optional bullet rotation support:

```text
Current straight bullet movement normally emits [id, x, y]. The optional trailing rotation slot remains supported for future projectile types, such as homing or turning projectiles, that may change rotation during flight.
```

## Secondary stale files

Update these after the primary protocol/service docs.

### docs/services/client/world-sync/world-sync-coordinator.md

Do not rewrite it as a transport doc.

Narrowly update wording from:

```text
RealtimeRouter applies world lane packets into world_lane_state
```

To:

```text
RealtimeRouter applies world, lifecycle, and hot movement packets into WorldLaneState.
```

Update active input path:

```text
RealtimeRouter.route_lane_packet(packet)
-> world/lifecycle/hot lane appliers update world_lane_state
-> WorldPresentationAdapter.apply_world_lane_state(...)
-> WorldSync.apply_world_lane_state(world_lane_state)
```

Mention that lifecycle-created dirty bullets/asteroids are allowed to create render nodes during world-sync fanout.

### docs/services/client/world-sync/entity-sync-owners.md

Replace:

```text
WorldSync receives world lane state
```

With:

```text
WorldSync receives accumulated WorldLaneState from world, lifecycle, and hot movement lane appliers.
```

Replace:

```text
removes rendered nodes when absent from latest world lane dictionary
```

With:

```text
removes rendered nodes when lifecycle/full-state no longer includes the entity.
```

For `ProjectileSync`, add:

```text
Unknown hot projectile updates must not create nodes. Hot updates after lifecycle delete are ignored. Lifecycle-created projectiles, including torpedoes, are allowed to create render nodes and preserve projectile_type.
```

For `AsteroidSync`, add:

```text
Unknown hot asteroid updates must not create nodes. Hot updates after lifecycle delete are ignored. Lifecycle-created asteroids are allowed to create render nodes and preserve variant/scale.
```

### docs/services/client/gameplay-runtime/gameplay-state-application.md

Update packet routing to include lifecycle packet families.

Add:

```text
WorldLaneState is populated by world_full/world_delta, asteroids_lifecycle, bullets_lifecycle, asteroid_delta, and bullet_delta. Lifecycle packets define existence; hot packets update existing entities only.
```

Update the `world_lane_applier.gd` code-map note so it mentions lifecycle methods.

### docs/domains/technical/local-singleplayer-routing-flow.md

Replace stale world-lane lifecycle wording with:

```text
world lane: ships, pickups, and world/full-state presentation
asteroids.lifecycle lane: asteroid creates/deletes
bullets.lifecycle lane: bullet/projectile creates/deletes
asteroids/bullets hot lanes: movement updates only
```

### docs/services/game-server/simulation/runtime/simulation-loop-and-phase-order.md

Update the world-lane projection summary so:

* Player/avatar and pickup state remain on world.
* Asteroid/bullet lifecycle moves to lifecycle lanes.
* Asteroid/bullet movement stays on hot lanes.

### docs/systems-design/world/world-authority.md

Replace broad world-lane lifecycle ownership with:

```text
World authority owns the authoritative entities. Realtime projection exposes them through family-specific packet lanes: world for player/pickup/world presentation, asteroid/bullet lifecycle lanes for existence, and asteroid/bullet hot lanes for movement updates.
```

### docs/services/game-server/simulation/world/asteroid-spawning-and-variants.md

Replace “world lane asteroid records” where it refers to active asteroid lifecycle with:

```text
Asteroid creation/variant identity is exposed through asteroid lifecycle creates and full/baseline snapshots.
Regular asteroid movement is exposed through sr.asteroids hot movement updates.
```

Keep server authority over variant selection.

### docs/protocol/asteroid-variant-contract.md

Update so variant readback is through:

```text
asteroids_lifecycle asteroid_creates
world_full/bootstrap snapshots when needed
```

State:

```text
Hot asteroid_delta updates do not carry variant and do not create asteroids.
```

### docs/systems-design/entities/asteroids.md

Replace stale projection language with:

```text
The server projects asteroid existence through asteroid lifecycle packets and full/baseline state, and projects regular movement through asteroid_delta hot updates.
```

Update despawn language:

```text
Once removed, the asteroid is deleted through the asteroid lifecycle lane; later hot movement updates for that id must not resurrect it.
```

### docs/systems-design/entities/variants.md

Update asteroid variant references:

```text
For asteroids, the server-selected variant index is stored on the runtime asteroid and sent in asteroid lifecycle creates or full snapshot state. Hot movement updates do not own variant identity.
```

### docs/systems-design/entities/projectiles.md

Replace “projected through world lane bullet records” with:

```text
Projectile existence and identity are projected through bullets_lifecycle creates/deletes and full/baseline state. Regular movement is projected through bullet_delta hot updates.
```

Add:

```text
Projectile type identity is lifecycle-owned. Torpedo creates must preserve projectile_type so the client spawns a torpedo node instead of a fallback bullet node.
```

Replace disappearance language:

```text
A projectile delete on bullets_lifecycle means the authoritative server no longer presents that projectile as live. Later hot updates must not recreate it.
```

### docs/services/game-server/simulation/combat/weapons-and-projectile-fire.md

Update projectile readback:

```text
Weapon fire creates runtime projectiles server-side. Projectile lifecycle creates carry owner_id, weapon_id, projectile_type, initial transform, and other spawn-owned presentation identity over sr.bullets.lifecycle. Movement after spawn travels over sr.bullets.
```

### Devtools docs

Update stale world-lane bullet/asteroid references in:

```text
docs/devtools/client/debug-status-and-target-readmodels.md
docs/devtools/server/clear-entity-tools.md
docs/devtools/server/continuous-bullet-streams.md
docs/devtools/server/spawn-tools.md
docs/devtools/server/telemetry.md
```

Use this principle:

```text
Debug/devtools entity creation and clearing are reflected to clients through normal realtime entity-family lanes: asteroid/bullet lifecycle lanes for existence and hot lanes for movement.
```

For continuous bullet streams, use:

```text
Spawned bullets are not emitted as a stream-specific telemetry packet. Their lifecycle appears as normal bullets_lifecycle traffic, and their movement appears as normal bullet_delta traffic.
```

## Planning doc updates

Update these summaries if they still describe only six WebRTC lanes or say lifecycle remains on `sr.world`:

```text
docs/planning/protocol/realtime-protocol-architecture.md
docs/planning/domains/technical/network-observability-and-packet-budget.md
docs/planning/development-roadmap.md
```

Planning docs should not become the detailed implementation authority. They should summarize the current state and link to protocol/service docs.

Use this summary:

```text
Dedicated reliable/ordered lifecycle lanes now carry asteroid and bullet/projectile creates/deletes. Unreliable hot lanes carry movement only. Lifecycle lanes are required/critical traffic; hot lanes are high-priority supersedable traffic.
```

## Verification

After edits, run stale-reference searches.

```bash
cd /mnt/d/\!bin/space-rocks

{
  echo "== stale world-lifecycle phrases =="
  grep -R "lifecycle creates/deletes remain on sr.world" docs || true
  grep -R "asteroid/bullet lifecycle creates/deletes" docs || true
  grep -R "world lane bullet records" docs || true
  grep -R "world lane asteroid records" docs || true

  echo
  echo "== lifecycle lane coverage =="
  grep -R "sr.asteroids.lifecycle" docs || true
  grep -R "sr.bullets.lifecycle" docs || true
  grep -R "asteroids_lifecycle" docs || true
  grep -R "bullets_lifecycle" docs || true
} 2>&1 | tee /dev/tty | clip.exe
```

Expected result:

* No stale “remain on sr.world” claims.
* No broad “world lane owns asteroid/bullet lifecycle” claims.
* Remaining `world lane asteroid records` or `world lane bullet records` references are either gone or explicitly limited to full/bootstrap/compatibility context.
* `sr.asteroids.lifecycle`, `sr.bullets.lifecycle`, `asteroids_lifecycle`, and `bullets_lifecycle` appear in canonical protocol, transport, outbound, inbound, compact wire, gameplay packet, and lane projection docs.

Then run normal documentation/index verification used by the project.

## Report

Report:

```text
Changed files
New files, if any
Deleted files, if any
Stale sr.world lifecycle references removed
Lifecycle channel docs updated
Inbound routing docs updated
Outbound routing docs updated
Compact wire docs updated
World sync/entity sync docs updated
Devtools/entity docs updated
Planning summaries updated
Verification result
Remaining risks or assumptions
```

Specifically report whether any remaining `world lane asteroid records` or `world lane bullet records` references are intentional compatibility/full-snapshot wording.
