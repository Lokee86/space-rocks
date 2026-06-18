# Local Profiles HTTP API
Parent index: [Player Data](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe player-data local pilot and profile HTTP API ownership.

## Overview

TODO: summarize how the player-data service serves local profile and pilot HTTP requests.
Stub note: keep this focused on the local profiles API surface.

## Code root

- `services/player-data/`

## Responsibilities

- TODO: describe local profile lookup, local pilot lookup, and HTTP handler responsibilities.

## Does not own

- API-server auth policy.
- Client profile presentation.
- TODO: any other boundaries that belong outside player-data HTTP API ownership.

## Domain roles

- TODO: define the local profile and pilot roles that participate in HTTP requests.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe local-profile HTTP endpoints, request shapes, and response shapes.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what local profile and pilot data the player-data service owns or exposes.
- Stub note: do not assume schema or persistence policy here.

## Code map

- `services/player-data/httpapi/`
- `services/player-data/httpapi/local_profiles_handler.go`
- `services/player-data/httpapi/profile_handler.go`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- `services/player-data/httpapi/local_profiles_handler_test.go`
- TODO: add any additional verified tests here.

## Related docs

- [Player Data](../!README.md)
- TODO: add local-profile-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future local profiles HTTP API documentation.
Do not treat it as canonical source material.
