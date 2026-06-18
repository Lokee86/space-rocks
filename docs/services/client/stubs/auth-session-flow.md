# Auth Session Flow
Parent index: [Client](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe client auth token, session state, and sign-in handoff ownership.

## Overview

TODO: summarize how the client stores auth state and hands off into session-aware runtime flows.
Stub note: keep this focused on client-side session preparation, not backend identity policy.

## Code root

- `client/scripts/`

## Responsibilities

- TODO: describe auth token storage, auth session state, sign-in handoff, and session bootstrap responsibilities.

## Does not own

- Server identity, admission, or account policy.
- Lobby or gameplay authority.
- TODO: any other boundaries that belong outside client auth/session ownership.

## Domain roles

- TODO: define the auth and session roles that participate in sign-in and handoff flow.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe auth API calls, session bootstrap flow, and any token handoff surfaces used by the client.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what auth token or session state the client owns locally.
- Stub note: do not assume encryption, persistence, or account details here.

## Code map

- `client/scripts/auth/`
- `client/scripts/auth/auth_session.gd`
- `client/scripts/auth/auth_session_controller.gd`
- `client/scripts/auth/auth_token_store.gd`
- `client/scripts/boot/session_boot_controller.gd`
- `client/scripts/session/`
- `client/scripts/session/client_session_context.gd`
- `client/scripts/session/session_network_controller.gd`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- TODO: add auth-session and bootstrap test references if they are confirmed.
- Stub note: only list verified tests here.

## Related docs

- [Client](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add auth/session-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future client auth session flow documentation.
Do not treat it as canonical source material.
