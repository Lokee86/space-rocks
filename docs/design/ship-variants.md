# Ship Variants Plan

This is a future implementation plan for supporting different ship scenes with different server collision maps.

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

## Current Assumption

The server currently treats ships as one collision shape through the collision catalog.

Relevant areas:

```text
services/game-server/internal/game/entities/ship.go
services/game-server/internal/game/physics/collision_shapes.go
shared/collisions/collision_shapes.json
client/scripts/world_sync.gd
shared/packets/packets.toml
```

## Execution Plan

### 1. Define Ship Type IDs

Add stable ship type IDs.

Possible location:

```text
shared/game_data.toml
```

Example:

```json
DEFAULT_SHIP_TYPE: "v_wing"
```

If ship types grow beyond a few constants, consider a dedicated shared ship config later.

Regenerate constants after changes:

```bash
python3 tools/data_sync/main.py -push -constants -go -gds
```

### 2. Add Ship Type To Server State

Add a field to the server ship/session state:

```go
ShipType string
```

Likely touch:

```text
services/game-server/internal/game/entities/state.go
services/game-server/internal/game/session.go
```

New sessions should default to the current ship type.

### 3. Add Ship Type To Shared Snapshots

Update:

```text
shared/packets/packets.toml
```

Add `ship_type` to `ShipState`.

Regenerate packets:

```bash
python3 tools/data_sync/main.py -push -packets -go -gds
```

Generated outputs:

```text
services/game-server/internal/game/entities/packets_generated.go
services/game-server/internal/game/packets.go
client/scripts/packets.gd
```

### 4. Expand Collision Shape Catalog

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

### 5. Select Ship Collision Shape By Type

Update:

```text
services/game-server/internal/game/entities/ship.go
```

Instead of always using the one ship shape, use:

```go
catalog.ShipShape(ship.ShipType)
```

or:

```go
catalog.ShipShapeForType(ship.ShipType)
```

If the type is empty or unknown, fall back to the default ship type so malformed state does not crash gameplay.

### 6. Add Client Scene Mapping

Update:

```text
client/scripts/world_sync.gd
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

### 7. Add Selection Later

Do not build ship selection until the variant plumbing works.

Possible future selection inputs:

- main menu ship picker
- profile/account setting
- room option
- debug packet

The first implementation can hardcode every player to the default ship type.

### 8. Add Tests

Server tests should cover:

- default player gets default ship type
- `ShipState` includes ship type
- collision catalog returns the right shape for a known ship type
- unknown ship type falls back safely
- two ship types can produce different collision bodies

Run:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

### 9. Smoke Test In Godot

Manual checks:

- default ship still renders correctly
- remote players render the correct ship scene
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

For now, keep the first pass to scene selection plus collision shape selection.

## Design Rule

Do not let the client decide collision behavior. The client can request or display a ship type, but the server must own the selected type and collision map used for gameplay.
