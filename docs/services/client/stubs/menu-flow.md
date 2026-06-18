# Menu Flow
Parent index: [Client](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe client main menu, pregame menu, lobby menu, and navigation ownership.

## Overview

TODO: summarize how the client routes between main menu, pregame menu, lobby menu, and related navigation surfaces.
Stub note: keep this focused on client presentation flow, not server authority.

## Code root

- `client/scripts/`

## Responsibilities

- TODO: describe main menu flow, pregame menu flow, lobby menu flow, and navigation state ownership.

## Does not own

- Server-authoritative room state.
- Packet schema or transport behavior.
- TODO: any other boundaries that belong outside client menu ownership.

## Domain roles

- TODO: define the menu roles and presenters that participate in navigation flow.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe any menu-related signals, view models, or packet handoffs that the client uses.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what menu state or navigation state the client owns locally.
- Stub note: do not assume persistence or account details here.

## Code map

- `client/scripts/main_menu/`
- `client/scripts/ui/menus/`
- `client/scripts/ui/menu_flow/`
- `client/scripts/lobby/`
- `client/scripts/ui/lobby/`
- `client/scripts/shell/gameplay_menu_flow.gd`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- TODO: add menu flow and navigation test references if they are confirmed.
- Stub note: only list verified tests here.

## Related docs

- [Client](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add menu-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future client menu flow documentation.
Do not treat it as canonical source material.
