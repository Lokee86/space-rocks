# HTTP Contracts

`shared/contracts/http/openapi.yaml` owns the HTTP request and response shapes for the Rails API.

Rails controllers implement that contract, and Rails integration tests enforce it using `openapi_first`.

This is Level 2 enforcement:

- test-time request/response validation
- runtime OpenAPI middleware is not active yet

## What This Does Not Do

- does not generate Rails controllers
- does not replace Rails strong params
- does not own Rails database schema
- does not generate TypeScript yet
- does not cover WebSocket packet schema

## Update Rule

API payload shape changes must update `shared/contracts/http/openapi.yaml` and the relevant Rails tests in the same change.

## Verification

- `cd services/api-server && bundle exec rails test test/contracts/openapi_contract_test.rb`
- `cd services/api-server && bundle exec rails test`

## Related

- [Project source-of-truth map](../design/source-of-truth-map.md)
