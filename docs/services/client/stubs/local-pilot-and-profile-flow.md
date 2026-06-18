# Local Pilot And Profile Flow
Parent index: [Client](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe local pilot selection, profile readout, guest profile, and profile API client ownership.

## Overview

TODO: summarize how the client handles local pilot selection and profile-related readouts.
Stub note: keep this focused on client-side profile flow and pilot selection.

## Code root

- `client/scripts/`

## Responsibilities

- TODO: describe local pilot selection, profile readout, guest profile handling, and profile API client responsibilities.

## Does not own

- Server profile storage or account policy.
- Gameplay simulation authority.
- TODO: any other boundaries that belong outside client profile ownership.

## Domain roles

- TODO: define the profile and pilot roles that participate in local pilot and profile flow.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe profile API client calls, readout updates, and any profile handoff surfaces used by the client.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what local pilot state, profile readout state, and guest profile state the client owns locally.
- Stub note: do not assume persistence or account details here.

## Code map

- `client/scripts/profile/`
- `client/scripts/profile/profile_flow.gd`
- `client/scripts/profile/profile_readout.gd`
- `client/scripts/profile/local_pilot_api_client.gd`
- `client/scripts/profile/player_data_profile_api_client.gd`
- `client/scripts/profile/guest_transient_stats_provider.gd`
- `client/scripts/ui/local_pilots/`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- TODO: add local pilot and profile flow test references if they are confirmed.
- Stub note: only list verified tests here.

## Related docs

- [Client](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add profile-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future local pilot and profile flow documentation.
Do not treat it as canonical source material.
