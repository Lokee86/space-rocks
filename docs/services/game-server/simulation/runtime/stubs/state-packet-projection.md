# Stub: State Packet Projection

Parent index: [Game Server Simulation Runtime](../!README.md)

## Purpose

This stub is incomplete and non-canonical. It points to server-side gameplay state packet construction and projection.

## Overview

This stub tracks gameplay state packet construction and projection for active entities, player sessions, player lifecycle, room-facing state fields, and snapshot-safe runtime read models.

It includes future ownership for `StatePacket.players`, `player_sessions`, and `player_lifecycle` projection.

## Code root

`services/game-server/internal/game/runtime/`

## Expected ownership

Server-side gameplay state packet construction and projection.

This stub does not own presentation-event queue semantics, protocol schema source of truth, or client rendering.

## Related docs

- [Game Server Simulation Runtime](../!README.md)
- [Game Server Simulation](../../!README.md)

## Notes

This is a scaffold only.
