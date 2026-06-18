# Pickup Rendering
Parent index: [Client](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe client pickup presentation, collection effects, and pickup scene resolution ownership.

## Overview

TODO: summarize how pickups are rendered, animated, and presented when collected.
Stub note: keep this focused on client-side pickup presentation.

## Code root

- `client/scripts/`

## Responsibilities

- TODO: describe pickup presentation, collection effects, and pickup scene resolution responsibilities.

## Does not own

- Pickup spawn authority or server-side collection rules.
- Shared collision policy.
- TODO: any other boundaries that belong outside client pickup presentation ownership.

## Domain roles

- TODO: define the pickup rendering and presentation roles that participate in pickup flow.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe pickup presentation APIs, catalog lookups, and any packet-fed state surfaces used by the client.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what pickup presentation state, effect state, and scene resolution state the client owns locally.
- Stub note: do not assume persistence or simulation authority here.

## Code map

- `client/scripts/world/pickups/`
- `client/scripts/world/pickup_sync.gd`
- `client/scripts/world/pickup_sync_state.gd`
- `client/scripts/entities/pickup.gd`
- `client/scripts/world/pickups/pickup_presentation_catalog.gd`
- `client/scripts/world/projectiles/projectile_scene_resolver.gd`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- TODO: add pickup presentation test references if they are confirmed.
- Stub note: only list verified tests here.

## Related docs

- [Client](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add pickup-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future pickup rendering documentation.
Do not treat it as canonical source material.
