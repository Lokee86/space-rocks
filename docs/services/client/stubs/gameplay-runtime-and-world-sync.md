# Gameplay Runtime And World Sync
Parent index: [Client](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe client gameplay runtime composition, packet application, and visual sync ownership.

## Overview

TODO: summarize how client gameplay runtime pieces compose and how world state is applied for visual sync.
Stub note: keep this focused on client presentation and local runtime composition.

## Code root

- `client/scripts/`

## Responsibilities

- TODO: describe gameplay runtime composition, gameplay state application, world sync, and visual sync responsibilities.

## Does not own

- Server simulation authority.
- Packet schema ownership.
- TODO: any other boundaries that belong outside client runtime and sync ownership.

## Domain roles

- TODO: define the runtime and sync roles that participate in gameplay composition.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe packet application paths, runtime flow inputs, and any sync-facing APIs used by the client.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what runtime state, visual state, and sync state the client owns locally.
- Stub note: do not assume persistence or simulation authority here.

## Code map

- `client/scripts/gameplay/composition/`
- `client/scripts/gameplay/runtime/`
- `client/scripts/gameplay/state/`
- `client/scripts/world/`
- `client/scripts/world/world_sync.gd`
- `client/scripts/gameplay/runtime/gameplay_world_state_apply_flow.gd`
- `client/scripts/gameplay/state/gameplay_state_apply_flow.gd`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- TODO: add runtime and world-sync test references if they are confirmed.
- Stub note: only list verified tests here.

## Related docs

- [Client](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add runtime-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future gameplay runtime and world sync documentation.
Do not treat it as canonical source material.
