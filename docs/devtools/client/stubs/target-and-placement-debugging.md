# Target And Placement Debugging
Parent index: [Client](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe client target and placement debug tooling documentation.

## Overview

TODO: summarize the client debug tools for target selection and placement-related debugging.
Stub note: keep this focused on tooling and diagnostics.

## Debug-only scope

- TODO: define which target and placement tools are debug-only and what they expose.
- Stub note: do not blur into production gameplay UI.

## Server authority

- TODO: describe which server-owned data or commands the tools may observe or request.
- Stub note: keep authority rules conceptual.

## Client presentation

- TODO: describe the client-visible target and placement panels, readouts, or overlays.
- Stub note: keep presentation details separate from backend behavior.

## Commands or controls

- TODO: describe the commands, hotkeys, or controls that open or operate target and placement debugging.
- Stub note: this is intentionally incomplete.

## Telemetry

- TODO: describe any target status, placement state, or debug output surfaced by the tools.
- Stub note: only note verified telemetry surfaces later.

## Build/runtime gates

- `client/scripts/devtools/dev_tools_build_flags.gd`
- `client/scripts/devtools/devtools_target_resolver.gd`
- `client/scripts/devtools/context/devtools_placement_context.gd`
- TODO: describe any other build or runtime gates when they are confirmed.

## Code map

- `client/scripts/devtools/devtools_target_resolver.gd`
- `client/scripts/devtools/gameplay_debug_flow.gd`
- `client/scripts/devtools/context/devtools_placement_context.gd`
- `client/scripts/devtools/context/devtools_command_context.gd`
- `client/scripts/devtools/context/devtools_overlay_context.gd`
- TODO: add narrower code links when they are confirmed.

## Tests

- TODO: add client target and placement debugging test references if they are confirmed.
- Stub note: only list verified tests here.

## Related docs

- [Client Devtools](../!README.md)
- TODO: add target/placement-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future client target and placement debugging documentation.
Do not treat it as canonical source material.
