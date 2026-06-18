# Local Auth Smoke Flow
Parent index: [API Server Devtools](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe API-server local auth smoke-flow documentation.

## Overview

TODO: summarize the local email/password and OAuth smoke-flow paths used for API-server verification.
Stub note: keep this focused on smoke-flow tooling and diagnostics.

## Debug-only scope

- TODO: define which auth smoke paths are debug-only and which runtime areas they observe.
- Stub note: do not blur into production API policy.

## Server authority

- TODO: describe which API-server auth systems the smoke flow exercises and what remains authoritative.
- Stub note: keep authority rules conceptual.

## Client presentation

- TODO: describe any client-visible results or auth success readouts shown during the smoke flow.
- Stub note: keep presentation details separate from backend behavior.

## Commands or controls

- TODO: describe the local auth and OAuth smoke-flow commands or controls.
- Stub note: this is intentionally incomplete.

## Telemetry

- TODO: describe any auth status, redirect, or token exchange outputs surfaced by the smoke flow.
- Stub note: only note verified telemetry surfaces later.

## Build/runtime gates

- `services/api-server/config/environments/test.rb`
- `services/api-server/app/controllers/api/auth/`
- `services/api-server/app/services/auth/`
- TODO: describe any other build or runtime gates when they are confirmed.

## Code map

- `services/api-server/app/controllers/api/auth/discord_controller.rb`
- `services/api-server/app/controllers/api/auth/sessions_controller.rb`
- `services/api-server/app/controllers/api/auth/registrations_controller.rb`
- `services/api-server/app/controllers/api/auth/me_controller.rb`
- `services/api-server/app/controllers/api/auth/discord_login_sessions_controller.rb`
- `services/api-server/app/services/auth/`
- `services/api-server/app/services/auth/login_user.rb`
- `services/api-server/app/services/auth/oauth_login_user.rb`
- `services/api-server/app/services/auth/oauth_login_session_issuer.rb`
- `services/api-server/app/services/auth/oauth_state_issuer.rb`
- `services/api-server/app/services/auth/oauth_state_verifier.rb`
- `services/api-server/app/services/auth/verify_access_token.rb`
- TODO: add narrower code links when they are confirmed.

## Tests

- `services/api-server/test/controllers/api/auth/me_controller_test.rb`
- `services/api-server/test/controllers/api/auth/sessions_controller_test.rb`
- `services/api-server/test/controllers/api/auth/registrations_controller_test.rb`
- `services/api-server/test/controllers/api/auth/discord_controller_test.rb`
- `services/api-server/test/controllers/api/auth/discord_login_sessions_controller_test.rb`
- `services/api-server/test/services/auth/oauth_login_user_test.rb`
- `services/api-server/test/services/auth/oauth_login_session_issuer_test.rb`
- `services/api-server/test/services/auth/oauth_state_issuer_test.rb`
- `services/api-server/test/services/auth/oauth_state_verifier_test.rb`
- `services/api-server/test/services/auth/verify_access_token_test.rb`
- TODO: add any additional verified tests here.

## Related docs

- [API Server Devtools](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add local-auth-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future API-server local auth smoke-flow documentation.
Do not treat it as canonical source material.
