# Input And Targeting
Parent index: [Client](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe client gameplay input, mouse action, and visual targeting ownership.

## Overview

TODO: summarize how gameplay input is captured, mapped, and converted into targeting and mouse-action behavior.
Stub note: keep this focused on client-side input and visual targeting.

## Code root

- `client/scripts/`

## Responsibilities

- TODO: describe gameplay input capture, mouse action mapping, targeting candidate selection, and visual targeting responsibilities.

## Does not own

- Server-side targeting authority.
- Packet routing or simulation resolution.
- TODO: any other boundaries that belong outside client input and targeting ownership.

## Domain roles

- TODO: define the input and targeting roles that participate in gameplay input handling.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe input flows, mouse action flows, and targeting APIs used by the client.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what local input state, cursor state, and visual targeting state the client owns.
- Stub note: do not assume persistence or simulation authority here.

## Code map

- `client/scripts/gameplay/input/`
- `client/scripts/gameplay/targeting/`
- `client/scripts/gameplay/input/gameplay_input_flow.gd`
- `client/scripts/gameplay/input/mouse_action_flow.gd`
- `client/scripts/gameplay/input/mouse_action_mapper.gd`
- `client/scripts/gameplay/input/target_visual_picker.gd`
- `client/scripts/gameplay/targeting/gameplay_target_candidate_flow.gd`
- `client/scripts/gameplay/targeting/target_request_flow.gd`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- TODO: add input and targeting test references if they are confirmed.
- Stub note: only list verified tests here.

## Related docs

- [Client](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add input-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future client input and targeting documentation.
Do not treat it as canonical source material.
