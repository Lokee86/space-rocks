# Match End And Gameplay Menu
Parent index: [Client](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe client match result window, game menu, replay, lobby, menu, and quit flow ownership.

## Overview

TODO: summarize how the client handles match-end presentation and gameplay menu transitions.
Stub note: keep this focused on client-side menu and result flow.

## Code root

- `client/scripts/`

## Responsibilities

- TODO: describe match result window behavior, game menu behavior, replay/lobby/menu/quit flow, and transition responsibilities.

## Does not own

- Server match resolution.
- Network room lifecycle.
- TODO: any other boundaries that belong outside client match-end menu ownership.

## Domain roles

- TODO: define the match-end and menu roles that participate in result and transition flow.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe any packet-fed match-end data, menu signals, or flow APIs used by the client.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what result state, menu state, and transition state the client owns locally.
- Stub note: do not assume persistence or simulation authority here.

## Code map

- `client/scripts/gameplay/match_end/`
- `client/scripts/gameplay/game_over/`
- `client/scripts/shell/gameplay_menu_flow.gd`
- `client/scripts/ui/menus/game_menu.gd`
- `client/scripts/ui/menus/pregame_menu.gd`
- `client/scripts/main_menu/`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- TODO: add match-end and menu transition test references if they are confirmed.
- Stub note: only list verified tests here.

## Related docs

- [Client](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add match-end-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future match-end and gameplay menu documentation.
Do not treat it as canonical source material.
