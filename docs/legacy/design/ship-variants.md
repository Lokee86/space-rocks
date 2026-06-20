# Ship Variants

Parent index: [Design Legacy](./!INDEX.md)

## Status Summary

Ship variants currently exist as an implemented runtime contract for ship identity, resolved stats, and collision lookup.

## Ownership

- The Go game server owns ship type resolution, resolved ship stats, and authoritative collision behavior.
- The client consumes `ship_type` as part of the gameplay state contract.

## Data Sources

- `services/game-server/internal/game/runtime/ship.go`
- `services/game-server/internal/game/physics/collision_shapes.go`
- `shared/collisions/collision_shapes.json`
- `shared/packets/gameplay.toml`
- `client/scripts/world/world_sync.gd`

## Current Runtime Model

- Runtime ships carry `ShipTypeID`.
- Player sessions preserve `ShipTypeID` for respawn continuity.
- Resolved `ShipStats` exists on sessions and ships.
- `ShipStatModifiers` exists as the current profile layer over base constants.
- `CollisionShapeID` exists in resolved stats.
- Server collision behavior is authoritative.

Current default ship type:

```text
v_wing
```

## Ship Type Identity

`ShipState` includes:

```text
ship_type
```

This value is part of the live gameplay packet contract and is preserved across respawn through the player session.

## Resolved ShipStats

`ShipStats` is the resolved effective runtime value used for ship movement, chassis behavior, and collision lookup.
Weapon firing, projectile, damage, ammo, and impact ownership belongs to the weapon profile seam.

`ShipStatModifiers` is the per-ship profile layer over base game constants. The default modifiers remain neutral, so current gameplay stays unchanged for the default ship.

## Collision Shape ID

Resolved stats include a `CollisionShapeID`, and `ShipShapeByID` safely falls back to the current default ship shape when the ID is unknown or missing.

## Design Rule

Do not let the client decide collision behavior. The server owns the selected ship type, resolved stats, and collision map used for gameplay.

## Verification

- `ShipTypeID` is carried by runtime ship/session state.
- `ShipState` includes `ship_type`.
- `ShipStatModifiers` resolves into `ShipStats` for live ship setup.
- `ShipShapeByID` provides safe default fallback behavior.
- Server collision handling remains authoritative.

## Related Limits

- [Player Build Limits](./../limits/player-build-limits.md)