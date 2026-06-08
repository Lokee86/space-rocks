# Pickup System

## Status Summary

- `1_up` exists.
- `torpedo` exists.
- Pickups are server-authoritative.
- Pickups have server-authoritative age, lifespan, and expiry.
- Pickups are targetable.
- Pickups are devtools-spawnable.
- Pickup health is current health only.
- `pickup_collected` and `pickup_effect_applied` are separate events.
- Bullet/pickup collision damage is not enabled.
- Pickup end-of-life blink is client presentation.
- Normal spawning is separate work.

## Ownership

- `internal/game/entities/pickups` owns pickup entity, type, definition, collision, and current health.
- `internal/game/pickups` owns pickup collection rules and effect intents.
- Root `internal/game` owns entity maps, session mutation, removal, and event recording.
- Client world sync owns rendering and sync of pickup nodes.
- Client gameplay event flow owns pickup event presentation routing.

## Data Sources

- `shared/constants/server_entities.toml`
- `shared/constants/client/presentation.toml`
- `shared/packets/gameplay.toml`
- `shared/collisions/collision_shapes.json`

Pickup collision JSON should use class keys such as `powerup` and `weapon`, not per-type keys such as `1_up` or `torpedo`.

## Server Entity Model

- Pickup fields: `id`, `type`, `x`, `y`, `health`, `age_seconds` / `AgeSeconds`, `lifespan_seconds` / `LifespanSeconds`.
- Definition fields: `type`, `pickup_class`, `health`, `lifespan`.
- `CollisionBody` uses `CollisionShapeCatalog`.
- Server collision shape lookup uses `pickup_class`, not pickup type.
- Scene paths do not belong in server pickup definitions.
- Health is current health only.
- There is no pickup `max_health` field.

## Packet Model

- `StatePacket.pickups`.
- `PickupState` fields: `id`, `type`, `pickup_class`, `x`, `y`, `health`, `age_seconds`, `lifespan_seconds`.
- `remaining_lifespan` is not sent; the client derives it.
- `pickup_collected` event fields.
- `pickup_effect_applied` event fields.
- `pickup_expired` event fields.
- Devtools state wrapper must copy new `StatePacket` fields.

## Client Rendering Model

- `world_sync` coordinates pickup sync.
- `pickup_sync` forwards pickup age and lifespan to the pickup node and owns pickup scene instantiation, state application, cleanup, interpolation, and target positions.
- `pickup_class` chooses the generic pickup scene family; `type` chooses the `Badge` child icon.
- Detailed client rules live in [docs/client/pickup-rendering.md](../client/pickup-rendering.md).
- Pickup scene visuals are client presentation only.
- Pickup z-index uses the generated `PICKUP_Z_INDEX` constant.
- Pickup glow and pulse are scene-local presentation.
- `pickup.gd` computes end-of-life blink locally.

## Targeting

- Pickup is a canonical target kind.
- Client candidate flow includes pickups.
- Server validates pickup targets authoritatively.
- Pickup targets are valid for telemetry and readout.
- Pickup targets are not valid for player-only devtools commands requiring `target_player_id`.

## Devtools Spawn Flow

- Devtools requests pickup spawn.
- Server creates the authoritative pickup.
- Client renders from `StatePacket.pickups`.
- Client must not instantiate authoritative pickups locally.

## Collection Flow

- Player/pickup collision is detected.
- Collision/contact removes the pickup entity and emits `pickup_collected`.
- `internal/game/pickups.ResolveCollection` classifies pickup type and returns effect intent after collection.
- Stage 2 applies the effect intent and emits `pickup_effect_applied`.
- Unknown pickup types resolve to a no-op effect.
- Collision and contact itself is not currently a public event.

## Effect Intent Flow

- `internal/game/pickups` decides what pickup effects should happen.
- `1_up` maps to `add_lives` amount `1`.
- `torpedo` maps to `equip_weapon` secondary `torpedo` with ammo delta `1`.
- Weapon pickup ammo is additive, not replacement.
- Root game applies the mutation because it owns player sessions.
- Durable lives mutate `player_sessions`, not live ship or avatar state.

## Event Semantics

- `pickup_collected` means the pickup entity was consumed or removed.
- `pickup_effect_applied` means the gameplay mutation succeeded.
- `pickup_collected` should drive world and presentation feedback.
- `pickup_effect_applied` should drive HUD, result, and telemetry feedback.
- Collection and effect application are intentionally separate.

## Audio And Effects Guidance

- Spawn sound may live in the pickup scene.
- Collected sound should use `pickup_collected` through gameplay event and effects flow.
- Collection sound should not be owned by the pickup node because sync removal can free the node before sound finishes.
- `pickup_effect_applied` is better for HUD and feedback sounds tied to the gameplay result.

## Health And Future Damage

- Pickup health exists as current health only.
- No `max_health` exists for pickups.
- Bullet/pickup collision damage is not enabled.
- Future pickup damage should use the damage seam in [docs/design/damage.md](damage.md), build `DamageTarget` values in game-owned adapters, and apply `DamageResult` back to `Pickup.Health` through game-owned adapters.
- Destroyed pickup behavior should be a later explicit decision.

## Lifespan And Expiry

- Lifespan and expiry are implemented on the server.
- Server tracks age and lifespan, then expires pickups when lifespan is reached.
- Client end-of-life blinking starts near the end of life and accelerates.
- `pickup_expired` exists as a domain and presentation event.
- The client derives remaining lifespan from age and lifespan.

## Normal Spawning

- Devtools spawn is distinct from normal gameplay spawning.
- Normal spawning should consume the sealed pickup APIs.
- Normal spawning should not add rules directly to root game if a focused spawning seam is appropriate.

## Pickup Drops

- Pickups may enter gameplay from generated drop tables.
- Drop tables decide whether a destroyed source produces a pickup.
- The drop-table seam lives in [docs/design/drop-tables.md](drop-tables.md).
- Pickup collection and pickup effects remain owned by the existing pickup seam.

## Adding A New Pickup Type

- [ ] Add or verify the class scene family.
- [ ] Add a `Badge` child named exactly like the pickup type.
- [ ] Export collision shape.
- [ ] Add constants.
- [ ] Add pickup definition.
- [ ] Add effect intent rule.
- [ ] Add packet and event tests if needed.
- [ ] Update docs.

## Testing And Verification

- `python3 tools/data_sync/main.py -validate -packets`
- `python3 tools/data_sync/main.py -check -packets -go -gds`
- `python3 tools/data_sync/main.py -validate -constants`
- `python3 tools/data_sync/main.py -check -constants -go -gds`
- `python3 tools/data_sync/main.py -push -constants -go -gds`
- `python3 tools/data_sync/main.py -push -packets -go -gds`
- `cd /mnt/d/!bin/space-rocks`
- `godot --headless --path client -s res://tools/export_collision_shapes.gd`
- `go test ./...` in `services/game-server`
- `cd client; godot --headless -s res://addons/gut/gut_cmdln.gd -gdir=res://tests -ginclude_subdirs`
- `Select-String -Path docs,services,client,shared -Pattern 'pickup_collected'`
- `Select-String -Path docs,services,client,shared -Pattern 'pickup_effect_applied'`
- `Select-String -Path docs,services,client,shared -Pattern 'internal/game/pickups'`
- `Select-String -Path docs,services,client,shared -Pattern 'internal/game/entities/pickups'`
- `Select-String -Path docs,services,client,shared -Pattern 'StatePacket.pickups'`
- `Select-String -Path docs,services,client,shared -Pattern 'PickupState'`
