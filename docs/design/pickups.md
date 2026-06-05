# Pickup System

## Status Summary

- `1_up` exists.
- Pickups are server-authoritative.
- Pickups are targetable.
- Pickups are devtools-spawnable.
- Pickup health is current health only.
- `pickup_collected` and `pickup_effect_applied` are separate events.
- Bullet/pickup collision damage is not enabled.
- Lifespan and expiry are separate work.
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
- `client/scenes/pickups/1_up.tscn`

## Server Entity Model

- Pickup fields: `id`, `type`, `x`, `y`, `health`.
- Definition fields: `type`, scene path, `health`.
- `CollisionBody` uses `CollisionShapeCatalog`.
- Collision shape comes from exported Godot scene data.
- Health is current health only.
- There is no pickup `max_health` field.

## Packet Model

- `StatePacket.pickups`.
- `PickupState` fields: `id`, `type`, `x`, `y`, `health`.
- `pickup_collected` event fields.
- `pickup_effect_applied` event fields.
- Devtools state wrapper must copy new `StatePacket` fields.

## Client Rendering Model

- `world_sync` coordinates pickup sync.
- `pickup_sync` owns pickup scene instantiation, state application, cleanup, interpolation, and target positions.
- Pickup scene visuals are client presentation only.
- Pickup z-index uses the generated `PICKUP_Z_INDEX` constant.
- Pickup glow and pulse are scene-local presentation.

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
- `internal/game/pickups.ResolveCollection` classifies pickup type and returns effect intent.
- Stage 1 removes the pickup entity and emits `pickup_collected`.
- Stage 2 applies the effect intent and emits `pickup_effect_applied`.
- Collision and contact itself is not currently a public event.

## Effect Intent Flow

- `internal/game/pickups` decides what pickup effects should happen.
- `1_up` maps to `add_lives` amount `1`.
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
- Future damage should use the existing damage seam and update `Pickup.Health`.
- Destroyed pickup behavior should be a later explicit decision.

## Lifespan And Expiry

- Lifespan and expiry is separate work unless already implemented.
- Future lifecycle should include age, lifetime, step, `ReadyForRemoval` or equivalent, and state removal.

## Normal Spawning

- Devtools spawn is distinct from normal gameplay spawning.
- Normal spawning should consume the sealed pickup APIs.
- Normal spawning should not add rules directly to root game if a focused spawning seam is appropriate.

## Adding A New Pickup Type

- [ ] Add scene.
- [ ] Export collision shape.
- [ ] Add constants.
- [ ] Add pickup definition.
- [ ] Update client scene mapping if needed.
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
- `python3 tools/data_sync/main.py -push -collisions`
- `go test ./...` in `services/game-server`
- `cd client; godot --headless -s res://addons/gut/gut_cmdln.gd -gdir=res://tests -ginclude_subdirs`
- `Select-String -Path docs,services,client,shared -Pattern 'pickup_collected'`
- `Select-String -Path docs,services,client,shared -Pattern 'pickup_effect_applied'`
- `Select-String -Path docs,services,client,shared -Pattern 'internal/game/pickups'`
- `Select-String -Path docs,services,client,shared -Pattern 'internal/game/entities/pickups'`
- `Select-String -Path docs,services,client,shared -Pattern 'StatePacket.pickups'`
- `Select-String -Path docs,services,client,shared -Pattern 'PickupState'`
