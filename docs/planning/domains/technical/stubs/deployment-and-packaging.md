# Deployment And Packaging
Parent index: [Technical Planning](../!INDEX.md)

## Purpose

This doc plans the deployment and packaging seam for local and hosted play.

## Ownership Boundary

This doc owns planning for local packaged play, launching or bundling the Go game server, hosted multiplayer deployment, and restricted or no-embedded-SQLite builds.

It should stay on deployment shape and packaging constraints rather than account or gameplay policy.

## Current Inputs

- local packaged play inputs
- bundled game-server launch inputs
- hosted multiplayer deployment inputs
- restricted build inputs
- no-embedded-SQLite build inputs

## Planned Outputs

- packaging boundaries for local play and hosted deployment
- build-shape expectations for restricted targets
- deployment questions for future infrastructure work

## Related Docs

- [Planning](../../../!INDEX.md)
- [Website And Web Presence](../../web/website-and-web-presence.md)
- [Matchmaking And Room Discovery](../../platform/stubs/matchmaking-and-room-discovery.md)
- [Development Roadmap](../../../development-roadmap.md)
- [Player Data And Persistence](../../platform/stubs/player-data-and-persistence.md)
- [Shop, Commerce, And Economy](../../gameplay/shop-commerce-and-economy.md)

## Open Planning Questions

- Which packaging mode is the long-term default for local play?
- Which deployment targets need the game server bundled versus separate?
- Which build variants should exclude embedded SQLite support?
