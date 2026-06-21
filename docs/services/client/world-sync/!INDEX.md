# World Sync

Parent index: [Client](../!INDEX.md)

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
<!-- doc-ledger:files:start -->

- [asteroid-variant-presentation.md](asteroid-variant-presentation.md) - Asteroid Variant Presentation documentation.
- [entity-sync-owners.md](entity-sync-owners.md) - Projectile, asteroid, and pickup sync owners and scene-node synchronization.
- [pickup-presentation.md](pickup-presentation.md) - Pickup presentation ownership, sync handoff, and visual behavior.
- [view-anchor-and-visual-coordinates.md](view-anchor-and-visual-coordinates.md) - ViewAnchor, render-anchor, toroidal wrap, and server/visual coordinate conversion.
- [world-sync-coordinator.md](world-sync-coordinator.md) - WorldSync coordinator ownership, apply order, delegation, interpolation, and read-model exposure.
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Client](../!INDEX.md)
- [Gameplay Runtime](../gameplay-runtime/!INDEX.md)
- [Toroidal wrap](../../../systems-design/world/toroidal-wrap.md) - toroidal world design documentation.
- [Gameplay packets](../../../protocol/gameplay-packets.md) - gameplay realtime packet documentation.

## Notes

World sync starts after runtime normalization and stays focused on rendering authoritative state, not deciding gameplay outcomes.