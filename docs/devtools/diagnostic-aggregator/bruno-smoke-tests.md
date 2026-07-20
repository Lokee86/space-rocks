---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-72a1-9726-ee88b7e098da
document_type: general
policy_exempt: false
summary: This document describes the Bruno collection used for local diagnostic-aggregator smoke testing.
---
# Bruno Smoke Test

Parent index: [Diagnostic Aggregator](./!INDEX.md)

## Purpose

This document describes the Bruno collection used for local diagnostic-aggregator smoke testing.

It covers the collection root, local environment variables, request order, rejection coverage, runtime gates, diagnostics, and the hosted HTTP paths exercised by the collection.

## Overview

The Bruno smoke-test collection lives at:

```text
bruno-diagnostic-aggregator/
```

It is a local development and diagnostics tool for exercising the diagnostic-aggregator HTTP surface over real requests. The hosted service runs on the game-server HTTP listener at:

```text
http://localhost:8080
```

The collection currently exercises:

```text
POST /v1/diagnostic-reports
GET  /v1/diagnostic-reports/{{diagnostic_report_id}}
```

The collection is a manual smoke-check client. It does not own the HTTP contract, and it does not replace Go tests or server-side contract assertions.

## Local environment

The local Bruno environment is:

```text
bruno-diagnostic-aggregator/environments/local.yml
```

Use the `local` environment when opening the collection.

Current variables:

```text
base_url=http://localhost:8080
bearer_token=local-diagnostic-token
diagnostic_report_id=
```

The hosted service expects these local settings when smoke testing:

```text
DIAGNOSTIC_AGGREGATOR_ENABLED=true
DIAGNOSTIC_AGGREGATOR_TOKEN must match bearer_token
DIAGNOSTIC_AGGREGATOR_STORAGE_ROOT controls the local report storage root
```

`bearer_token` is only a placeholder value for local smoke testing. Committed files must not contain real tokens or secrets.

Example local WSL setup:

```bash
export DIAGNOSTIC_AGGREGATOR_ENABLED=true
export DIAGNOSTIC_AGGREGATOR_TOKEN=local-diagnostic-token
export DIAGNOSTIC_AGGREGATOR_STORAGE_ROOT=/tmp/space-rocks-diagnostic-reports
export BUILD_VERSION=dev
export ENVIRONMENT=development

cd services/game-server
go run ./cmd/game-server
```

## Happy-path flow

Run `create-valid` before `get-created`.

`create-valid` posts a valid diagnostic report and captures the returned `diagnostic_report_id` into the Bruno variable of the same name. `get-created` then reads that captured ID and verifies the stored report can be fetched.

Use this order for the happy-path smoke check:

1. `create-valid`
2. `get-created`

## Rejection coverage

The collection also checks the most important rejection paths and their canonical error codes:

- `create-unauthorized` → HTTP 401 `unauthorized`
- `create-malformed-json` → HTTP 400 `malformed_json`
- `create-trailing-json` → HTTP 400 `trailing_json`
- `create-unsupported-media-type` → HTTP 415 `unsupported_media_type`
- `create-invalid-report` → HTTP 422 `invalid_diagnostic_report`
- `get-invalid-id` → HTTP 400 `invalid_diagnostic_report_id`
- `get-not-found` → HTTP 404 `diagnostic_report_not_found`
- `method-not-allowed` → HTTP 405 `method_not_allowed`

## Debug-only scope

The Bruno collection is a local smoke-test tool.

It is useful for:

```text
manual endpoint checks
local hosted-service verification
report creation and retrieval checks
response-shape inspection
request rejection inspection
quick regression checks while editing diagnostic-aggregator code
```

It is not a production monitoring surface, a contract source, a database seed mechanism, or a replacement for Go tests.

The collection does not define API authority. HTTP request and response shapes are owned by the diagnostic-aggregator Go implementation and its contract tests, while Bruno only exercises those endpoints as a live client.

## Server authority

Diagnostic-aggregator remains authoritative for all behavior observed by the collection.

Bruno exercises the real hosted report routes:

```text
POST /v1/diagnostic-reports
GET  /v1/diagnostic-reports/{{diagnostic_report_id}}
```

The collection-level `GET /v1/diagnostic-reports` request is intentionally kept in `method-not-allowed.yml` as the HTTP 405 test. The collection endpoint accepts `POST` only; retrieval requires a report ID.

The hosted service owns:

```text
report validation
report creation
report retrieval
authorization checks
content-type checks
JSON syntax checks
status code selection
canonical error code selection
```

Bruno owns only request construction, local variables, request order, and local display of responses.

## Client presentation

The Bruno collection is not the game client.

It presents API responses inside Bruno, including response JSON, HTTP status, headers, and console output from request scripts. The only scripted presentation behavior is capture of the created `diagnostic_report_id` from `create-valid`.

## Diagnostics

Useful checks:

```text
HTTP status code
response JSON
request script errors
captured diagnostic_report_id value
local environment values
```

If a request fails, inspect the response body for the canonical error code above and confirm the local hosted-service settings are loaded.

## Build and runtime gates

The Bruno collection depends on a running game-server process hosting diagnostic-aggregator.

Typical local setup assumptions:

```text
game-server dependencies are installed
required diagnostic-aggregator environment variables are set
hosted diagnostic-aggregator is enabled
the server is listening at http://localhost:8080
```

For local smoke testing, the service must be configured with a valid bearer token and `DIAGNOSTIC_AGGREGATOR_STORAGE_ROOT` if persistence is desired.

The collection is not a CI gate by itself. Go tests and contract assertions remain the automated verification path.

## Code map

### Bruno collection

```text
bruno-diagnostic-aggregator/opencollection.yml
bruno-diagnostic-aggregator/environments/local.yml
bruno-diagnostic-aggregator/diagnostic-reports/folder.yml
bruno-diagnostic-aggregator/diagnostic-reports/create-valid.yml
bruno-diagnostic-aggregator/diagnostic-reports/get-created.yml
bruno-diagnostic-aggregator/diagnostic-reports/create-unauthorized.yml
bruno-diagnostic-aggregator/diagnostic-reports/create-malformed-json.yml
bruno-diagnostic-aggregator/diagnostic-reports/create-trailing-json.yml
bruno-diagnostic-aggregator/diagnostic-reports/create-unsupported-media-type.yml
bruno-diagnostic-aggregator/diagnostic-reports/create-invalid-report.yml
bruno-diagnostic-aggregator/diagnostic-reports/get-invalid-id.yml
bruno-diagnostic-aggregator/diagnostic-reports/get-not-found.yml
bruno-diagnostic-aggregator/diagnostic-reports/method-not-allowed.yml
```

### Hosted service paths exercised

```text
services/diagnostic-aggregator/hosted/service.go
services/diagnostic-aggregator/internal/diagnosticapi/handler.go
services/diagnostic-aggregator/internal/diagnosticreports/service.go
services/diagnostic-aggregator/internal/diagnostics/http.go
```

### Contract and behavior tests exercised

```text
services/diagnostic-aggregator/hosted/service_test.go
services/diagnostic-aggregator/internal/diagnosticapi/contract_test.go
services/diagnostic-aggregator/internal/diagnosticapi/handler_test.go
services/diagnostic-aggregator/internal/diagnosticreports/service_test.go
services/diagnostic-aggregator/internal/diagnostics/http_test.go
```
