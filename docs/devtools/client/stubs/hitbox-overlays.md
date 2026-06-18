# Hitbox Overlays
Parent index: [Client Devtools](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe client hitbox overlay documentation.

## Overview

TODO: summarize the client hitbox overlays and how they present hitbox/shape information.
Stub note: keep this focused on tooling and diagnostics.

## Debug-only scope

- TODO: define which hitbox overlays are debug-only and what runtime areas they observe.
- Stub note: do not blur into production gameplay UI.

## Server authority

- TODO: describe which server-owned data or commands may feed the overlays.
- Stub note: keep authority rules conceptual.

## Client presentation

- TODO: describe the overlay visuals, labels, or readouts shown by the client.
- Stub note: keep presentation details separate from backend behavior.

## Commands or controls

- TODO: describe the commands, hotkeys, or controls that enable or operate hitbox overlays.
- Stub note: this is intentionally incomplete.

## Telemetry

- TODO: describe the shape IDs, catalogs, or debug readouts surfaced by the overlays.
- Stub note: only note verified telemetry surfaces later.

## Build/runtime gates

- `client/scripts/devtools/dev_tools_build_flags.gd`
- `client/scripts/devtools/hitboxes/devtools_server_hitbox_overlay.gd`
- `client/scripts/devtools/hitboxes/debug_shape_catalog_store.gd`
- TODO: describe any other build or runtime gates when they are confirmed.

## Code map

- `client/scripts/devtools/hitboxes/`
- `client/scripts/devtools/hitboxes/devtools_server_hitbox_overlay.gd`
- `client/scripts/devtools/hitboxes/debug_shape_catalog_packet_reader.gd`
- `client/scripts/devtools/hitboxes/debug_shape_catalog_store.gd`
- `client/scripts/devtools/hitboxes/debug_shape_id_resolver.gd`
- TODO: add narrower code links when they are confirmed.

## Tests

- TODO: add client hitbox overlay test references if they are confirmed.
- Stub note: only list verified tests here.

## Related docs

- [Client Devtools](../!README.md)
- TODO: add hitbox-overlay-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future client hitbox overlay documentation.
Do not treat it as canonical source material.
