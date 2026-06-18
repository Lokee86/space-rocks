# HTTP API Contracts
Parent index: [Protocol](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe OpenAPI-backed HTTP contract documentation.

## Overview

TODO: summarize how HTTP contracts are defined, shared, and enforced across services.
Stub note: keep this focused on contract behavior, not implementation detail.

## Participating systems

- API server HTTP surfaces.
- Player-data HTTP surfaces.
- TODO: any other HTTP participants that are confirmed later.

## Authority

- TODO: describe which service owns the contract source and which services consume it.
- Stub note: do not invent authority details beyond confirmed OpenAPI-backed flows.

## Message or request flow

- TODO: describe request/response contract flow, schema validation, and contract test flow.
- TODO: document any public versus internal HTTP contract distinctions if they are actually used.

## Source-of-truth files

- `services/api-server/test/contracts/openapi_contract_test.rb`
- `services/api-server/test/support/openapi_contract_assertions.rb`
- TODO: add the canonical HTTP contract source files when they are confirmed.

## Service responsibilities

- TODO: describe HTTP contract ownership, documentation, and enforcement responsibilities.
- Stub note: keep transport or business logic details out of this doc unless they are confirmed.

## Validation and testing

- `services/api-server/test/contracts/openapi_contract_test.rb`
- TODO: add any additional contract tests or checks if they are confirmed.
- Stub note: only list verified tests or checks here.

## Related docs

- [Protocol](../!README.md)
- [Documentation procedure](../../documentation-procedure.md)
- TODO: add HTTP-contract-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future HTTP API contract documentation.
Do not treat it as canonical source material.
