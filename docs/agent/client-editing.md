---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-74e0-8c96-47331bb28c6e
document_type: general
policy_exempt: false
summary: This doc guides safe agent edits to the Godot client.
---
# Client Editing
Parent index: [Agent](./!INDEX.md)

## Purpose

This doc guides safe agent edits to the Godot client.

## Overview

Canonical client runtime documentation lives under `docs/services/client/`.

## Rules

- Client owns presentation, UI, audio/effects, local input collection, interpolation, and presentation controllers.
- Do not put authoritative simulation outcomes in the client.
- Do not pass raw WebSocket URLs through scene/menu code.
- Do not hand-edit generated client packet/constants files.
- Route packet schema changes through shared source files and data-sync.
- Keep Main Menu and app-shell code as routing/composition where existing docs say so.
- Treat [`SessionNetworkController`](../services/client/app-shell-and-session/session-boot-and-network-target.md) as a known pressure point: route existing behavior through it as required, but do not add unrelated policy or new ownership there; use a concrete seam when a new responsibility appears.
- Inspect current scenes/scripts before changing UI flow.
- Keep broad HUD containers from swallowing clicks unless they are intentionally interactive.

## Related docs

- [Client Service](../services/client/!INDEX.md)
- [Client Menu Flow](../services/client/menu-flow.md)
- [Client Input And Targeting](../services/client/input-and-targeting.md)
- [Client Networking Flow](../services/client/networking-flow/!INDEX.md)
- [Data](../data/!INDEX.md)
- [Generated Files](./generated-files.md)
- [Godot Editing](./godot-editing.md)

## Notes

Current menu, auth, and runtime facts belong in service or protocol docs, not this agent guide.
