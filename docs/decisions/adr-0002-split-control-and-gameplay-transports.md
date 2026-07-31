---
author: brian
created: "2026-07-31"
document_id: 3b13cb0e-ff7d-44fd-91f9-17ac694d57f0
document_type: general
policy_exempt: false
summary: Records the decision to keep session, control, lobby, auth, and signaling on WebSocket while active gameplay state uses lane-specific WebRTC DataChannels.
---
# ADR-0002: Split Control and Gameplay Transports

Parent index: [Architecture Decisions](!INDEX.md)

## Purpose

This record explains why Space Rocks separates reliable session/control/signaling traffic from lane-specific active gameplay delivery.

## Overview

WebSocket provides the stable control path. WebRTC DataChannels let reliable lifecycle/state and supersedable high-frequency movement use different delivery semantics without changing server gameplay authority.

## Status

Accepted.

## Context

Session setup, authentication, lobby state, room control, signaling, and one-off responses need reliable ordered delivery and straightforward connection lifecycle. Active gameplay contains several different traffic classes: reliable lifecycle and state, transient events, supersedable high-frequency movement, and recipient-specific projections. Sending every class through one ordered WebSocket creates head-of-line coupling and prevents traffic-specific reliability and cadence policy.

## Decision

WebSocket remains the session, authentication, room/lobby, control, signaling, and fallback connection path. After signaling succeeds, active gameplay delivery uses lane-specific WebRTC DataChannels.

Reliable ordered lanes carry world, overlay, session, event, and entity-lifecycle traffic. Unordered unreliable hot lanes carry supersedable ship, asteroid, and bullet movement. Packet-family sequence, baseline, chunk assembly, interest, prioritization, and recovery rules remain explicit per owning lane; transport ordering does not imply cross-lane ordering.

## Consequences

### Ownership and dependencies

- WebSocket session code owns connection, authentication, room/control routing, and WebRTC signaling.
- Gameplay lane policy owns channel assignment, reliability, cadence, candidate construction, and recovery semantics.
- Simulation remains transport-independent and projects authoritative state through networking adapters.
- Client routing and lane gates validate each packet family before mutating presentation read models.

### State, lifecycle, and operations

- An open WebSocket does not imply ready gameplay DataChannels.
- Reliable ordering is per channel only; world, lifecycle, and hot packets can arrive in different orders.
- Clients must assemble chunked sequences atomically and request authoritative recovery after invalid or incomplete required state.
- Hot-lane loss is tolerated through superseding sequences; reliable lifecycle loss is not silently treated as movement loss.

### Compatibility and migration

- Packet schemas and compact-wire descriptors remain shared source contracts across client and server.
- A binary transport, merged channel, WebTransport, or changed reliability layout requires a superseding ADR and measured migration plan.
- WebSocket fallback for active gameplay is not implied unless explicitly designed and tested.

## Alternatives Considered

- **One WebSocket for all traffic:** rejected because reliable ordering couples stale hot movement to important control and lifecycle traffic.
- **One reliable WebRTC channel:** rejected because it recreates head-of-line coupling between unrelated packet families.
- **All gameplay traffic unreliable:** rejected because entity lifecycle, baseline, session, and match-facing state require reliable recovery semantics.

## Verification

- Protocol generator and compact-wire contract checks.
- Server lane routing, cadence, chunking, and candidate-budget tests.
- Client lane acceptance, atomic assembly, stale-sequence, match-identity, and recovery tests.
- Hosted connectivity smoke tests proving WebSocket signaling and DataChannel readiness.

## Risks and Debt

- Multiple channels increase lifecycle, instrumentation, and cross-lane recovery complexity.
- NAT traversal and relay availability can fail after WebSocket connection succeeds; user-visible connection failure remains required.

## Related docs

- [Realtime client-server flow](../domains/technical/realtime-client-server-flow.md)
- [Realtime WebRTC gameplay transport](../protocol/realtime-webrtc-gameplay-transport.md)
- [Realtime WebSocket protocol](../protocol/realtime-websocket-protocol.md)
- [Lane packet projection](../services/game-server/simulation/runtime/lane-packet-projection.md)

## Notes

Transport reliability does not replace packet-family identity, sequence, baseline, assembly, and recovery contracts.
