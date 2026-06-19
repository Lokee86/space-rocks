# Telemetry And Packet Routing

Parent index: [Game Server Networking](./!README.md)

## Purpose

This document describes the game-server networking boundary for telemetry-facing realtime packets.

It explains how telemetry and diagnostic packet requests enter the game server, how responses leave the server, what runtime state may be exposed, and what this boundary does not own.

## Overview

Telemetry packet routing is a narrow realtime surface on the game server.

The current implementation does not send telemetry to a durable external telemetry backend. Instead, telemetry-facing packets are routed through the same WebSocket packet path used by normal client/server communication. The server decodes inbound packets, routes telemetry or diagnostic requests to the appropriate runtime handler, and writes structured response packets back to the connected client or devtool consumer.

This boundary exists so liveness, latency, and development diagnostics can use the real packet codec and outbound writer path without becoming gameplay authority or a separate debug-only transport.

The networking-facing behavior is:

```text
client or devtool
-> WebSocket packet
-> networking decode
-> packet router
-> telemetry or diagnostic handler
-> outbound server packet
-> client or devtool consumer
```

## Code root

```text
services/game-server/
```

## Responsibilities

Telemetry packet routing owns the game-server side of:

* Accepting telemetry-facing packet requests after the networking layer decodes them.
* Routing telemetry and diagnostic packet types to the correct server-side handler.
* Writing telemetry responses through the normal outbound packet writer.
* Keeping telemetry responses structured as normal server packets.
* Reporting only runtime facts the server owns or can observe.
* Preserving the same codec and transport assumptions used by normal gameplay packets.
* Keeping diagnostic packet behavior separate from simulation authority.

## Does not own

Telemetry packet routing does not own:

* WebSocket connection lifecycle.
* General inbound packet dispatch.
* General outbound packet fanout.
* Room lifecycle, membership, or ownership rules.
* Gameplay simulation state mutation.
* Match result reporting.
* Player-data persistence.
* Auth token verification.
* Client UI presentation of telemetry or diagnostics.
* A durable metrics, tracing, or observability backend.

## Packet surface

The telemetry-facing packet surface currently uses realtime packets, not HTTP.

Relevant packet families include:

```text
telemetry_ping
telemetry_pong
debug_status
debug_shape_catalog
```

`telemetry_ping` is a client-originated liveness or latency probe. The server response is `telemetry_pong`.

`debug_status` exposes server-side runtime status for development consumers.

`debug_shape_catalog` exposes diagnostic shape/catalog data used by debug tooling.

These packet families should remain read-only from the perspective of gameplay authority. They may observe server state, but they should not create a second path for simulation mutation.

## Routing flow

Inbound telemetry routing follows the normal packet path:

```text
WebSocket receives bytes
-> packet codec decodes client message
-> networking packet router identifies packet type
-> telemetry or debug handler builds response
-> outbound writer serializes server message
-> WebSocket sends bytes
```

The important ownership rule is that telemetry does not bypass networking. Even when a packet is diagnostic, it still uses the production codec and outbound writer.

## Data ownership

Telemetry packet routing does not own durable data.

It may read:

* connection/session state known by the networking layer
* room or server runtime state exposed through existing service seams
* generated packet definitions shared with the client
* debug/devtool catalog data prepared by server-side tooling

It must not persist:

* player profile state
* match result state
* gameplay progression state
* long-term telemetry history

If durable telemetry storage is added later, that should be documented as a separate integration boundary rather than folded into packet routing.

## Trust and validation

The server should treat telemetry requests as client input.

Telemetry packet routing should trust only:

* decoded packet structure after codec validation
* server-owned runtime state
* room/session state owned by existing game-server seams
* generated packet contracts

Telemetry packet routing should reject or ignore:

* malformed packets
* unknown telemetry packet types
* requests that require unavailable runtime state
* client-submitted telemetry values that claim authoritative game facts
* diagnostic requests that are not enabled for the current runtime mode

Telemetry packets must not become a way for clients to mutate authoritative state outside normal gameplay or devtool command rules.

## Runtime boundaries

Telemetry packet routing sits between three runtime boundaries:

```text
networking
  owns WebSocket session state, packet decode, packet dispatch, and packet writes

telemetry/debug handlers
  own response construction for diagnostic packet families

rooms/simulation/devtools
  own the authoritative runtime state being observed
```

Telemetry routing should stay thin. If a response needs game state, the routing layer should ask the owning game-server seam for that state instead of duplicating ownership.

## Tests and verification

Relevant verification should cover:

* telemetry request decode and route selection
* telemetry response packet construction
* outbound write behavior for telemetry responses
* rejection or no-op behavior for unknown packet types
* debug packet behavior when devtool state is unavailable
* packet codec compatibility with generated client/server packet definitions

Useful test areas include:

```text
services/game-server/internal/networking/
services/game-server/internal/protocol/packetcodec/
services/game-server/internal/devtools/
services/game-server/internal/rooms/
```

## Code map

Primary implementation folders:

```text
services/game-server/internal/networking/
services/game-server/internal/networking/outbound/
services/game-server/internal/protocol/packetcodec/
services/game-server/internal/devtools/
services/game-server/internal/rooms/
```

Related generated or shared packet contract paths:

```text
shared/packets/
```

Important non-ownership boundaries:

```text
services/player-data/
client/
docs/services/game-server/networking/
docs/devtools/
```

`services/player-data/` is not part of telemetry packet routing unless a future durable telemetry sink is added.

`client/` consumes telemetry responses but does not own game-server routing behavior.

`docs/services/game-server/networking/` owns the broader inbound and outbound packet routing documentation.

`docs/devtools/` owns debug-tool behavior that consumes or presents diagnostic packet output.

## Related docs

* [Game Server Integrations](../integrations/!README.md)
* [Game Server](../!README.md)
* [Game Server Networking](./!README.md)
* [Inbound Packet Routing](./inbound-packet-routing.md)
* [Outbound Packet Routing](./outbound-message-flow.md)
* [Room Network Adapter](./room-network-adapter.md)
* [Game Server Observability](../observability/!README.md)
* [Protocol](../../../protocol/!README.md)
* [Devtools](../../../devtools/!README.md)

## Notes

The word telemetry in this document means realtime diagnostic packet traffic, not a full observability pipeline.

If Space Rocks later adds metrics export, tracing, structured log aggregation, or a durable telemetry service, that should become a separate integration document. This document should remain focused on packets that enter and leave the game server through the realtime protocol.
