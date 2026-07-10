## Space Rocks Devtools Boundary Skill

Use this skill when maintaining the current Space Rocks devtools Controller / Target / Control boundary.

This skill describes the implemented architecture as the current state. It is not a proposal to extract the boundary later.

## Goal

Keep devtools ownership split cleanly across the existing seams without changing gameplay behavior, packet schemas, packet transport, or realtime scheduling.

The current ownership split is:

```text
internal/devtools
  owns Controller, command policy, capability interfaces, target resolution,
  stream-runtime coordination, observer registration policy, status projection,
  and debug DTOs

internal/game
  owns authoritative state and exposes narrow capabilities through Control

internal/networking
  constructs Control and Controller and routes existing packets
```

## Hard rules

* Do not change devtools packet transport in this pass.
* Do not add a WebRTC/UDP devtools lane in this pass.
* Do not edit the realtime scheduler.
* Do not change packet schemas.
* Do not redesign locking.
* Do not change gameplay behavior.
* Production files in `services/game-server/internal/devtools` must not import `github.com/Lokee86/space-rocks/server/internal/game`.
* Devtools integration tests may import `internal/game`.
* Production files in `services/game-server/internal/game` must not import `internal/devtools`.
* Production files in `services/game-server/internal/devtools` must not import the root game package.
* Keep files small and focused.
* Preserve current behavior unless a test proves current behavior cannot be preserved.

## Current maintenance checklist

Use this skill for implementation-oriented maintenance of the current devtools boundary.

When changing this boundary:

- add or change the narrow capability interface
- implement it on `game.Control`
- route policy through `Controller`
- preserve packet and gameplay behavior
- update focused tests and canonical docs
- run boundary verification

## Current code map

The main files involved in this boundary are:

```text
internal/devtools/target.go
internal/devtools/target_player_ids.go
internal/devtools/controller.go
internal/devtools/observer_registry.go
internal/devtools/collision_telemetry.go
internal/game/control.go
internal/game/control_*.go
internal/networking/inbound/*.go
internal/networking/outbound/*.go
```

Prefer small capability interfaces:

```go
type StatusTarget interface { ... }
type ToggleTarget interface { ... }
type PlayerTarget interface { ... }
type SpawnTarget interface { ... }
type RespawnTarget interface { ... }
type CounterTarget interface { ... }
type ClearTarget interface { ... }
type CollisionTelemetryTarget interface { ... }
type StreamTarget interface { ... }
```

Compose them into the current `Target` shape, keeping `PlayerTargetSource` separate from `StatusTarget` so command fanout and status membership can follow different sources.

Use current method names, including `ApplyPlayerDefeat`, and avoid spelling removed `Devtools...` method names in new guidance or examples.

`internal/game` owns the `Control` adapter and focused `control_*.go` files. `internal/devtools` owns `Controller`, target selection, observer policy, and DTO projection. `internal/networking` constructs `Control` and `Controller` and routes the existing packets.

`All-player` command fanout uses `Control.TargetPlayerIDs()`.

`Controller.StatusesForAllPlayers()` uses `MatchDecision().Players` for all-player status membership.

`Controller` owns the stream runtime coordination.

The controller-injected stream runtime owns the begin, step, and clear operations for that same controller.

The game adapter may return the underlying `*Game` from `ObserverKey()` as `any`, but devtools must not import `internal/game` to key the registry.

## Dependency rules

- production and test files in `internal/game` must not import `internal/devtools`
- devtools integration tests may import `internal/game`
- production devtools files must not import the root game package

## Network boundary

Keep packet routing unchanged.

Existing WebSocket devtools packets should still be decoded and routed as before.

Only the execution target changes:

```go
control := game.NewControl(room.GameInstance())
controller := devtools.NewController(devtools.Dependencies{
    Target: control,
})
controller.HandleCommand(playerID, command)
```

Do not add `sr.devtools`.
Do not move debug status/catalog to WebRTC.
Do not touch channel negotiation.

## DTO ownership

Debug DTO ownership belongs to `internal/devtools`.

Good devtools-owned names:

```text
StatusSnapshot
CollisionPoint
CollisionBody
```

The game adapter can return raw authoritative or physics data. Devtools should shape that data into debug packets.

## Testing guidance

Add or preserve coverage for:

```text
- session-backed status membership
- separate command fanout membership
- one-way package dependencies
- controller-owned stream runtime
- raw game telemetry plus devtools DTO projection
- devtools commands still execute
- debug status payload stays compatible
- debug shape catalog payload stays compatible
- collision telemetry JSON stays compatible
- continuous bullet stream still registers one observer per game target
- target-player commands still resolve the same player IDs
```

## Stop conditions

Stop and report if:

```text
- the task requires changing devtools packet transport
- the task requires changing realtime lane scheduling
- the task requires changing packet schemas
- preserving behavior requires a locking redesign
- the change grows into unrelated cleanup
- production internal/devtools still needs to import internal/game after interface extraction
```

Report:

```text
- changed files
- remaining legacy game-side debug method references, if any
- remaining production devtools imports of internal/game, if any
- tests run
```
