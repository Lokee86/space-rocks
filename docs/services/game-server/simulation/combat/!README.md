# Game Server Simulation Combat

Parent index: [Game Server Simulation](../!README.md)

Combat documentation for the game server lives here.

## Ownership

This folder owns weapons and projectile fire, collision-to-damage flow, damage resolution, and radial effects documentation for the game server simulation.

## Does Not Belong

- Room membership or lifecycle rules.
- WebSocket transport details.
- External integration internals.
- Process startup or shutdown.
- Target state ownership or world motion rules.

## Direct Files
<!-- doc-ledger:files:start -->

- [collision-to-damage-flow.md](collision-to-damage-flow.md) - Collision to damage flow documentation.
- [damage-resolution.md](damage-resolution.md) - Damage resolution documentation.
- [radial-effects.md](radial-effects.md) - Radial effects documentation.
- [weapons-and-projectile-fire.md](weapons-and-projectile-fire.md) - Weapons And Projectile Fire documentation.
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Game Server Simulation](../!README.md)
- [Game Server](../../!README.md)
- [Scoring](../scoring/!README.md)
- [Services index](../../../!README.md)

## Notes

This boundary stays on authoritative combat behavior and not player or room ownership.