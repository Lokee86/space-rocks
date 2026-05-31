# Ship Variants

This page tracks the thin ship-variant foundation that exists today and the remaining work for full ship variants.

## Goal

Allow players to use different ship types while keeping the server authoritative for collision behavior.

Desired behavior:

- each player/session has a ship type
- the server uses the correct collision shape for that ship type
- the client renders the correct ship scene for each player
- the existing ship remains the default so current behavior does not change

## Why This Is More Than A Client Scene Swap

The client can render different scenes fairly easily, but the server currently owns collision outcomes. If the server still assumes one ship collision shape, different sprites/scenes would only be cosmetic.

To make ship variants real, the server must know which collision shape each player is using.

## Current State

The server now has a thin runtime foundation:

- `ShipTypeID` on runtime ships
- session-level `ShipTypeID` for respawn continuity
- `ship_type` in `ShipState`
- resolved `ShipStats` on sessions and ships
- `ShipStatModifiers` as neutral per-ship modifiers over base game constants
- stats-driven movement, shooting cooldown, bullet speed, bullet lifetime, and bullet spawn offset
- `CollisionShapeID` in resolved stats
- `ShipShapeByID` with safe fallback to the current default ship shape

The default ship type and collision shape ID are currently:

```text
v_wing
```

The server still has only one imported ship collision shape in `shared/collisions/collision_shapes.json`. The ID seam exists, but a keyed multi-ship collision catalog has not been added yet.

The client still renders all players with the current player scene/sprite. It receives `ship_type`, but does not select visuals from it yet.

Relevant areas:

```text
services/game-server/internal/game/entities/ship.go
services/game-server/internal/game/physics/collision_shapes.go
shared/collisions/collision_shapes.json
client/scripts/world/world_sync.gd
shared/packets/gameplay.toml
client/scripts/networking/packets/packets.gd
```

## Implemented Foundation

### Ship Type Identity

Runtime ships and player sessions carry the current ship type.

Current default:

```text
entities.DefaultShipTypeID = "v_wing"
```

`ShipState` includes:

```text
ship_type
```

This field is passive on the client today.

### Stats

`ShipStats` is the resolved effective runtime value used by movement, shooting, bullets, and collision lookup.

`ShipStatModifiers` is the per-ship profile layer over base constants. The default modifiers are neutral `1.0` values, so current gameplay is unchanged.

Resolution flow:

```text
ship_type
-> ResolveShipStatModifiers
-> ResolveShipStats
-> playerSession.Stats
-> Ship.Stats
-> movement/shooting/collision
```

### Collision Shape ID

Resolved stats include a `CollisionShapeID`. Live ship collision and respawn safety use that ID through `ShipShapeByID`.

Runtime fallback policy:

- the default ID returns the current single ship shape
- unknown IDs safely fall back to the current default ship shape

## Remaining Work

### 1. Add Real Ship Definitions

When a second ship is needed, add a new modifier profile in `ResolveShipStatModifiers`.

Example shape:

```go
case "heavy":
	return ShipStatModifiers{
		RotationSpeed:     0.75,
		ThrustForce:       0.85,
		MaxSpeed:          0.8,
		Damping:           1.0,
		BulletCooldown:    1.25,
		BulletSpeed:       1.0,
		BulletLifetime:    1.0,
		BulletSpawnOffset: 1.0,
		CollisionShapeID:  "heavy",
	}
```

Do not add variant-specific conditionals in movement, shooting, or collision code.

### 2. Expand Collision Shape Catalog

Change the collision shape data/model from one ship shape to keyed ship shapes.

Possible target shape:

```go
Ships map[string]ImportedCollisionShape `json:"ships"`
```

Likely touch:

```text
services/game-server/internal/game/physics/collision_shapes.go
shared/collisions/collision_shapes.json
```

Keep backward compatibility or migrate the current single ship shape into:

```json
"ships": {
  "v_wing": { ... }
}
```

`ShipShapeByID` already exists and should become the lookup point for the keyed catalog. Preserve safe fallback for malformed or unknown IDs.

### 3. Add Client Scene Mapping

Update:

```text
client/scripts/world/player_sync.gd
```

Add a mapping from `ship_type` to scene path or preload.

Example:

```gdscript
const SHIP_SCENES := {
	"v_wing": preload("res://scenes/player.tscn")
}
```

When a player appears, instantiate the scene for that player's `ship_type`.

If the type is missing or unknown, use the default scene.

### 4. Add Selection Later

Do not build acquisition or inventory until the variant plumbing works.

Possible future selection inputs:

- main menu ship picker
- profile/account setting
- room option
- debug packet

The current implementation hardcodes every player to the default ship type.

### 5. Add Tests For Real Variants

Existing server tests cover the foundation:

- default player gets default ship type
- `ShipState` includes ship type
- default modifiers are neutral
- default ship resolves to baseline effective stats
- unknown ship type and collision shape ID fall back safely
- live collision and respawn safety use the collision shape ID seam

When a real second ship is added, add tests for:

- the new ship type resolves different modifiers
- the new collision shape ID resolves through the keyed catalog
- two ship types can produce different collision bodies

Run:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

### 6. Smoke Test In Godot

Manual checks:

- default ship still renders correctly
- remote players still render
- player movement and shooting feel unchanged
- collision shape matches the selected ship type
- existing gameplay works when no explicit ship selection is present

## What Makes This More Complex Later

The basic version should stay moderate in scope. It gets harder if ship variants include:

- different movement stats
- different weapons
- different health/lives
- different scoring rules
- team/faction restrictions
- unlock/account ownership
- client prediction with different physics per ship

Keep acquisition, ownership, unlocks, purchases, and persistence outside the real-time `Ship` entity. Those belong to a future account/API layer.

## Design Rule

Do not let the client decide collision behavior. The client can request or display a ship type, but the server must own the selected type, resolved stats, and collision map used for gameplay.
