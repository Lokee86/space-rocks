# Stub: Toroidal Space And Motion

Parent index: [Game Server Simulation World](../!README.md)

## Purpose

This stub is incomplete and non-canonical. It points to server-side world bounds, position wrapping, and motion integration support.

## Overview

This stub tracks default world bounds, toroidal position wrapping, per-entity motion stepping helpers, ship motion stepping through the motion package, asteroid motion stepping, and projectile motion and lifetime stepping.

Move-policy gating is an input from player pause/suspension, not pause ownership.

## Code root

`services/game-server/internal/game/motion/`

## Expected ownership

Server-side world bounds, wrapping, and motion integration support.

## Related docs

- [Game Server Simulation World](../!README.md)
- [Game Server Simulation](../../!README.md)
- [Player Pause And Suspension](../../players/stubs/player-pause-and-suspension.md)

## Notes

Player docs may link here for active avatar movement, but this doc owns motion mechanics.
This is a scaffold only.
