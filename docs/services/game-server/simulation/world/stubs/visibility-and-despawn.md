# Stub: Visibility And Despawn

Parent index: [Game Server Simulation World](../!README.md)

## Purpose

This stub is incomplete and non-canonical. It points to server-side world despawn behavior driven by camera consumption.

## Overview

This stub tracks world systems consuming camera views, far-from-camera despawn, pending-despawn versus ready-for-removal behavior, and world-entity removal boundaries.

## Code root

`services/game-server/internal/game/`

## Expected ownership

Server-side world visibility and despawn behavior.

## Related docs

- [Game Server Simulation World](../!README.md)
- [Game Server Simulation](../../!README.md)
- [Player Camera View State](../../players/stubs/player-camera-view-state.md)

## Notes

Player camera-view state creation and update belongs in `../../players/stubs/player-camera-view-state.md`.
Asteroid spawning details belong in `asteroid-spawning-and-variants.md`.
Active player avatar lifecycle belongs in `../../players/stubs/active-player-avatar-state.md`.
World despawn behavior consumes camera views but does not own their creation or update.
This is a scaffold only.
