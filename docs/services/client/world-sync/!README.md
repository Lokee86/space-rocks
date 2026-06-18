# World Sync

Client world sync documentation lives here.

## Ownership

- Client server-state rendering.
- Visual sync.
- ViewAnchor and render-anchor behavior.
- Continuous visual coordinates.
- Entity sync ownership.

## Does Not Belong

- Server simulation authority.
- Gameplay packet schema ownership.
- Gameplay runtime fanout before world state reaches WorldSync.
- HUD, menu, input, and targeting orchestration.
- Pickup gameplay rules.
- Future planning.

## Direct Files

- [world-sync-coordinator.md](world-sync-coordinator.md) - WorldSync coordinator ownership, apply order, delegation, interpolation, and read-model exposure.
- [entity-sync-owners.md](entity-sync-owners.md) - Projectile, asteroid, and pickup sync owners and scene-node synchronization.
- [view-anchor-and-visual-coordinates.md](view-anchor-and-visual-coordinates.md) - ViewAnchor, render-anchor, toroidal wrap, and server/visual coordinate conversion.
- [pickup-presentation.md](pickup-presentation.md) - Pickup presentation ownership, sync handoff, and visual behavior.

## Stub Files

- None.

## Direct Folders

- None.

## Related Docs

- [Client](../!README.md)
- [Gameplay Runtime](../gameplay-runtime/!README.md)
- [Toroidal wrap](../../../systems-design/world/stubs/toroidal-wrap.md) - Stub: toroidal world design documentation.
- [Gameplay packets](../../../protocol/stubs/gameplay-packets.md) - Stub: gameplay realtime packet documentation.

## Notes

World sync starts after runtime normalization and stays focused on rendering authoritative state, not deciding gameplay outcomes.
