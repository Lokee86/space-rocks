# Player Inventory Client

Parent index: [Game Server Integrations](./!INDEX.md)

## Purpose

This document describes the game-server adapter that loads hangar inventory and submits inventory grants through the player-data runtime boundary.

## Overview

```text
game or playerbuild service
-> playerinventory.RuntimeClient
-> generated player-data command
-> PlayerDataSink.HandlePlayerDataCommand
-> generated result
-> inventory/build consumer
```

The adapter keeps player-data protocol encoding out of game simulation packages and satisfies the inventory-loader interface used by `playerbuild.Service`.

## Code root

```text
services/game-server/internal/playerinventory/
```

## Responsibilities

This boundary owns:

- validating that a player-data sink is present
- encoding load and grant commands with the shared codec
- assigning the correct generated packet type
- decoding generated result structures
- converting rejected loads and grants into caller-visible errors
- exposing the exact inventory result, including fallback and repair flags
- satisfying the player-build inventory loader contract

## Does not own

This adapter does not own:

- inventory persistence
- grant policy
- starter inventory synthesis
- build eligibility
- room identity resolution
- transport hosting
- client UI

## Domain roles

The runtime client is an in-process service adapter. The game server may currently co-host player-data, but callers depend only on the `PlayerDataSink` interface and generated protocol, not on player-data storage internals.

A load error code is treated as a failed inventory load. A grant must return `Accepted`; duplicate accepted grants remain successful because idempotence is owned by player-data.

## Protocols and APIs

```go
type PlayerDataSink interface {
    HandlePlayerDataCommand(payload []byte) ([]byte, error)
}

func NewRuntimeClient(sink PlayerDataSink) (*RuntimeClient, error)
func (client *RuntimeClient) Load(identity protocol.PlayerDataIdentity, context protocol.PlayerDataRequestContext) (protocol.PlayerDataLoadHangarInventoryResult, error)
func (client *RuntimeClient) ApplyGrant(command protocol.PlayerDataApplyInventoryGrant) (protocol.PlayerDataApplyInventoryGrantResult, error)
```

## Data ownership

The adapter owns no inventory state. It passes generated command/result values across the service boundary. The returned inventory snapshot remains player-data-owned until the build service captures it in a `LoadedBuildContext` for one resolution flow.

## Code map

```text
services/game-server/internal/playerinventory/runtime_client.go
services/player-data/codec/
services/player-data/protocol/packets.go
services/player-data/playerdata/dispatcher.go
services/game-server/internal/game/playerbuild/service.go
```

## Tests

Tests cover missing sinks, command encoding and packet types, successful loads and grants, malformed responses, player-data error codes, and rejected grants.

```text
services/game-server/tests/playerinventory/runtime_client_test.go
services/game-server/internal/game/playerbuild/service_test.go
```

## Related docs

- [Game Server Integrations](./!INDEX.md)
- [Hangar Inventory](../../player-data/hangar-inventory.md)
- [Player Builds And Loadouts](../simulation/players/player-builds-and-loadouts.md)
- [Player Data HTTP Hosting](player-data-http-hosting.md)
- [Inventory And Build Flow](../../../domains/player-experience/inventory-and-build-flow.md)

## Notes

Keeping this adapter narrow allows the build system to depend on an inventory loader rather than on player-data runtime or storage packages.
