# Bruno Smoke Tests
Parent index: [API Server](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe Bruno collection smoke-test documentation.

## Overview

TODO: summarize how Bruno smoke tests are used to exercise API-server endpoints.
Stub note: keep this focused on smoke-test tooling and diagnostics.

## Debug-only scope

- TODO: define which smoke tests are debug-only and which runtime areas they observe.
- Stub note: do not blur into production API policy.

## Server authority

- TODO: describe which API-server systems the smoke tests exercise and what remains authoritative.
- Stub note: keep authority rules conceptual.

## Client presentation

- TODO: describe any client-visible results or response summaries shown by smoke-test tooling.
- Stub note: keep presentation details separate from backend behavior.

## Commands or controls

- TODO: describe the Bruno collection commands or controls used for smoke tests.
- Stub note: this is intentionally incomplete.

## Telemetry

- TODO: describe any response traces, status outputs, or failure summaries surfaced by smoke tests.
- Stub note: only note verified telemetry surfaces later.

## Build/runtime gates

- `services/api-server/config/environments/test.rb`
- `services/api-server/test/contracts/openapi_contract_test.rb`
- TODO: describe any other build or runtime gates when they are confirmed.

## Code map

- `services/api-server/test/contracts/openapi_contract_test.rb`
- `services/api-server/test/support/openapi_contract_assertions.rb`
- `services/api-server/test/controllers/api/auth/`
- `services/api-server/test/controllers/internal/auth/`
- TODO: add narrower code links when they are confirmed.

## Tests

- `services/api-server/test/contracts/openapi_contract_test.rb`
- `services/api-server/test/controllers/api/auth/me_controller_test.rb`
- `services/api-server/test/controllers/api/auth/sessions_controller_test.rb`
- `services/api-server/test/controllers/api/auth/registrations_controller_test.rb`
- `services/api-server/test/controllers/api/auth/discord_controller_test.rb`
- `services/api-server/test/controllers/api/auth/discord_login_sessions_controller_test.rb`
- `services/api-server/test/controllers/internal/auth/verify_tokens_controller_test.rb`
- TODO: add any additional verified tests here.

## Related docs

- [API Server Devtools](../!README.md)
- TODO: add Bruno-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future API-server Bruno smoke-test documentation.
Do not treat it as canonical source material.
