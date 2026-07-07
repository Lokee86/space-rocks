## Purpose

Use this skill when moving asteroid and bullet entity lifecycle traffic out of `sr.world` into dedicated reliable lifecycle lanes.

The goal is to keep `sr.world` focused on player, match, and world-level state while high-churn entity families get their own lifecycle ownership.

## Target lane model

Use this lane ownership model:

| Lane | Reliability | Ordered | Responsibility |
|---|---:|---:|---|
| `sr.world` | yes | yes | players, match state, world-level state |
| `sr.asteroids.lifecycle` | yes | yes | asteroid create/delete/split lifecycle |
| `sr.bullets.lifecycle` | yes | yes | bullet, torpedo, projectile create/delete lifecycle |
| `sr.asteroids` | no | no | asteroid hot deltas only |
| `sr.bullets` | no | no | bullet hot deltas only |

Core rule:

> Lifecycle lanes define existence. Hot lanes update known existing entities only.

## Non-goals

Do not move player lifecycle or player state out of `sr.world`.

Do not make hot lanes create entities implicitly.

Do not merge lifecycle data into hot delta packets.

Do not add enemy lifecycle implementation yet.

Do not change asteroid, bullet, torpedo, or projectile simulation behavior unless required to preserve existing behavior after rerouting.

Do not remove unused tuple support for future projectile rotation or homing behavior just because current bullets do not change rotation.

## Required lifecycle ownership

Move asteroid lifecycle facts out of `sr.world` and into `sr.asteroids.lifecycle`.

Asteroid lifecycle includes:

- asteroid create
- asteroid delete/despawn
- asteroid split lifecycle facts, if represented as delete-parent/create-children
- asteroid variant, size, and initial state needed to spawn the correct client entity

Move bullet/projectile lifecycle facts out of `sr.world` and into `sr.bullets.lifecycle`.

Bullet/projectile lifecycle includes:

- bullet create
- bullet delete/despawn
- torpedo create
- torpedo delete/despawn
- projectile kind/type identity
- owner/source identity
- initial transform
- initial velocity, if the current spawn path depends on it
- TTL or expiration metadata, if currently spawn-owned

Keep hot movement/update data in the existing unreliable lanes:

- `sr.asteroids`
- `sr.bullets`

Hot lanes should only carry compact transient updates for entities that already exist on the client.

## Server requirements

Add explicit server lane constants/specs for:

- `sr.asteroids.lifecycle`
- `sr.bullets.lifecycle`

Register both lifecycle lanes as reliable and ordered.

Keep these existing hot lanes unreliable and unordered:

- `sr.asteroids`
- `sr.bullets`

Update server packet planning, projection, or classification so:

- asteroid creates route to asteroid lifecycle
- asteroid deletes route to asteroid lifecycle
- bullet creates route to bullet lifecycle
- bullet deletes route to bullet lifecycle
- torpedo creates route to bullet lifecycle
- torpedo deletes route to bullet lifecycle
- player state remains in world
- match/world state remains in world
- asteroid hot deltas remain in asteroid hot lane
- bullet hot deltas remain in bullet hot lane

Lifecycle packets must be treated as reliable required packets, not supersedable hot packets.

Do not allow lifecycle packets to be dropped by hot-lane budget rules.

## Client requirements

Add explicit client channel specs for:

- `sr.asteroids.lifecycle`
- `sr.bullets.lifecycle`

Route asteroid lifecycle packets into the same asteroid spawn/despawn behavior currently reached through world packets.

Route bullet lifecycle packets into the same bullet/projectile spawn/despawn behavior currently reached through world packets.

Preserve projectile type identity. A torpedo lifecycle create must spawn a torpedo node, not a fallback bullet node.

Hot lanes must not create entities.

Client hot update rules:

- hot update for unknown asteroid ID: ignore
- hot update for unknown bullet/projectile ID: ignore
- hot update for deleted asteroid ID: ignore
- hot update for deleted bullet/projectile ID: ignore
- hot update must not resurrect a deleted entity
- hot update must not create a fallback/default entity

## Cross-lane race rule

Reliable lifecycle lanes and unreliable hot lanes do not provide global ordering across lanes.

The client must tolerate:

- hot update arriving before lifecycle create
- hot update arriving after lifecycle delete
- lifecycle create arriving after earlier hot packets were ignored
- lifecycle delete arriving while newer hot packets are still in flight

Do not fix these races by making hot lanes reliable.

The correct first-pass behavior is to ignore hot packets that cannot safely apply.

## Entity ID reuse rule

Check whether asteroid and bullet/projectile entity IDs can be reused within a match.

If IDs are not reused within a match, document that as the current safety assumption.

If IDs can be reused quickly, add or preserve a generation/spawn-version guard so stale hot traffic for a previous entity cannot affect a new entity with the same ID.

Do not guess. If reuse behavior is unclear and cannot be verified locally, stop and report the uncertainty.

## Suggested implementation order

1. Add lifecycle lane constants/specs on server and client.
2. Register lifecycle WebRTC channels as reliable/ordered.
3. Update server routing/classification for asteroid lifecycle.
4. Update client routing for asteroid lifecycle.
5. Move asteroid lifecycle creates/deletes out of world.
6. Update server routing/classification for bullet/projectile lifecycle.
7. Update client routing for bullet/projectile lifecycle.
8. Move bullet, torpedo, and projectile lifecycle creates/deletes out of world.
9. Add or confirm client race guards.
10. Add or update focused tests.
11. Update protocol, networking, and packet-ownership docs.

## Test requirements

Add or update server tests proving:

- asteroid creates route to `sr.asteroids.lifecycle`
- asteroid deletes route to `sr.asteroids.lifecycle`
- bullet creates route to `sr.bullets.lifecycle`
- bullet deletes route to `sr.bullets.lifecycle`
- torpedo creates route to `sr.bullets.lifecycle`
- torpedo deletes route to `sr.bullets.lifecycle`
- player/world state still routes to `sr.world`
- asteroid hot deltas still route to `sr.asteroids`
- bullet hot deltas still route to `sr.bullets`
- lifecycle packets are reliable/required, not supersedable hot packets

Add or update client tests proving:

- asteroid lifecycle lane spawns asteroids
- asteroid lifecycle lane despawns asteroids
- bullet lifecycle lane spawns bullets
- bullet lifecycle lane despawns bullets
- torpedo lifecycle create preserves torpedo identity
- hot delta for unknown asteroid is ignored
- hot delta for unknown bullet/projectile is ignored
- hot delta for deleted asteroid is ignored
- hot delta for deleted bullet/projectile is ignored
- hot deltas do not create fallback bullet nodes

## Documentation requirements

Update docs that describe:

- lane ownership
- WebRTC channel specs
- WebSocket parity, if documented
- gameplay packet routing
- server outbound message flow
- client inbound packet routing
- lane packet projection
- packet budgeting
- compact wire mapping, if lane IDs or ownership are listed

Required documentation statement:

> Entity lifecycle ownership is split by entity family. The world lane owns player and world/match state. Asteroid lifecycle packets use `sr.asteroids.lifecycle`. Bullet/projectile lifecycle packets use `sr.bullets.lifecycle`. Hot asteroid and bullet lanes are unreliable movement/update lanes only and must not create entities implicitly.

Required stable limitation statement:

> Cross-lane ordering is not guaranteed between reliable lifecycle lanes and unreliable hot lanes. Clients must tolerate hot updates arriving before lifecycle create packets and after lifecycle delete packets.

Remove or update stale references that say:

- `sr.world` owns asteroid lifecycle
- `sr.world` owns bullet lifecycle
- hot lanes create entities
- bullet/asteroid hot updates are reliable lifecycle delivery
- all entity lifecycle traffic is world-lane traffic

## Verification expectations

Before reporting completion, verify behavior at the implementation level and report what was checked.

Required verification points:

- server lifecycle lane routing works
- client lifecycle lane routing works
- hot lanes remain unreliable/unordered
- lifecycle lanes are reliable/ordered
- hot packets do not create entities
- stale hot packets do not resurrect deleted entities
- torpedoes spawn as torpedoes under bullet/projectile lifecycle routing
- player and match/world state remain on `sr.world`
- no stale docs still describe asteroid/bullet lifecycle as world-lane owned

## Report format

Report:

- changed files
- lifecycle lanes added
- server routing changes
- client routing changes
- tests added or updated
- docs updated
- verification performed
- any remaining risks or assumptions, especially around entity ID reuse