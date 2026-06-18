# HUD And Gameplay UI
Parent index: [Client](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe HUD, overlays, gameplay UI, and presentation ownership.

## Overview

TODO: summarize how the client owns player-facing HUD, overlay, and gameplay UI presentation.
Stub note: keep this focused on client presentation responsibilities.

## Code root

- `client/scripts/`

## Responsibilities

- TODO: describe HUD, overlay, gameplay UI, and presentation flow responsibilities.

## Does not own

- Server-authoritative simulation.
- Packet schemas and server state.
- TODO: any other boundaries that belong outside client HUD ownership.

## Domain roles

- TODO: define the UI roles and presenters that participate in HUD and overlay flow.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe HUD-facing presenters, view models, and any packet-fed presentation APIs used by the client.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what HUD state, overlay state, and presentation state the client owns locally.
- Stub note: do not assume persistence or simulation authority here.

## Code map

- `client/scripts/ui/hud/`
- `client/scripts/gameplay/presentation/`
- `client/scripts/gameplay/input/hud_input_policy.gd`
- `client/scripts/shell/gameplay_hud_flow.gd`
- `client/scripts/gameplay/presentation/gameplay_presentation_flow.gd`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- TODO: add HUD and gameplay UI test references if they are confirmed.
- Stub note: only list verified tests here.

## Related docs

- [Client](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add HUD-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future HUD and gameplay UI documentation.
Do not treat it as canonical source material.
