---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-72a1-9726-ee88b7e098da
document_type: general
policy_exempt: false
summary: This document describes the Bruno collection used for local diagnostic-aggregator HTTP smoke testing.
---
# Bruno Smoke Tests

Parent index: [Diagnostic Aggregator Devtools](./!INDEX.md)

## Purpose

This document describes the Bruno collection used for local diagnostic-aggregator HTTP smoke testing.

## Overview

The collection at `bruno-diagnostic-aggregator/` sends real HTTP requests to a configured diagnostic-aggregator endpoint. It can target either the standalone executable or the optional game-server-hosted adapter because both expose the same report routes.

The committed `local` environment uses `http://localhost:8080`. That address matches the standalone executable's default port and may also match a locally hosted game-server configuration. Production exposes the standalone container through host loopback port `8083`; production credentials are never stored in the Bruno collection.

Bruno is a manual smoke client. It does not own the HTTP contract, report validation, authentication, persistence, or error semantics, and it does not replace the Go test suite.

## Debug-only scope

Use the collection for:

- manual report creation and retrieval;
- response-shape inspection;
- selected validation and not-found checks;
- local standalone or hosted-adapter regression checks.

Do not use it as a production monitor, contract source, data seeder, or secret store.

## Server authority

Diagnostic-aggregator remains authoritative for bearer authentication, request limits, decoding, validation, safety inspection, report construction, storage, retrieval, and error responses. Bruno only submits requests and asserts selected results.

## Client presentation

Bruno presents request and response bodies in its normal client interface. The `create-valid` request captures the returned `diagnostic_report_id` into the collection variable used by `get-created`.

## Commands or controls

Open `bruno-diagnostic-aggregator/opencollection.yml`, select the `local` environment, and configure:

```text
base_url=http://localhost:8080
bearer_token=local-diagnostic-token
diagnostic_report_id=
```

The bearer token must match the running service configuration. The committed value is a local placeholder only.

The current collection contains:

| Request | Expected behavior |
| --- | --- |
| `create-valid` | Authenticated `POST /v1/diagnostic-reports`; stores the returned report ID. |
| `get-created` | `GET /v1/diagnostic-reports/{{diagnostic_report_id}}`; retrieves the accepted report. |
| `create-invalid-report` | Unsupported report version; expects HTTP 422 and `invalid_diagnostic_report`. |
| `get-invalid-id` | Invalid report identifier; expects HTTP 400 and `invalid_diagnostic_report_id`. |
| `get-not-found` | Valid but unknown report identifier; expects HTTP 404 and `diagnostic_report_not_found`. |
| `method-not-allowed` | Unsupported method; expects HTTP 405 and `method_not_allowed`. |

Run `create-valid` before `get-created`. The other requests are independent negative-path checks.

## Telemetry

The collection does not create a separate telemetry stream. The running diagnostic service emits its normal canonical request, rejection, accepted, stored, and storage-failure events. Inspect the service-owned logs when diagnosing a failed smoke request.

## Build/runtime gates

A smoke target must have:

- `DIAGNOSTIC_AGGREGATOR_ENABLED=true` where the hosted configuration seam is used;
- a nonblank `DIAGNOSTIC_AGGREGATOR_TOKEN` matching `bearer_token`;
- `BUILD_VERSION` and `ENVIRONMENT` for service identity;
- a writable storage root when persistence is expected.

Standalone local execution:

```bash
cd services/diagnostic-aggregator
DIAGNOSTIC_AGGREGATOR_ENABLED=true \
DIAGNOSTIC_AGGREGATOR_TOKEN=local-diagnostic-token \
BUILD_VERSION=dev \
ENVIRONMENT=development \
go run ./cmd/diagnostic-aggregator
```

The optional hosted adapter may instead be exercised by running the game-server with matching diagnostic configuration. The collection itself is not an automated CI gate.

## Code map

### Collection

- `bruno-diagnostic-aggregator/opencollection.yml`
- `bruno-diagnostic-aggregator/environments/local.yml`
- `bruno-diagnostic-aggregator/diagnostic-reports/folder.yml`
- `bruno-diagnostic-aggregator/diagnostic-reports/create-valid.yml`
- `bruno-diagnostic-aggregator/diagnostic-reports/get-created.yml`
- `bruno-diagnostic-aggregator/diagnostic-reports/create-invalid-report.yml`
- `bruno-diagnostic-aggregator/diagnostic-reports/get-invalid-id.yml`
- `bruno-diagnostic-aggregator/diagnostic-reports/get-not-found.yml`
- `bruno-diagnostic-aggregator/diagnostic-reports/method-not-allowed.yml`

### Service paths exercised

- `services/diagnostic-aggregator/cmd/diagnostic-aggregator/main.go`
- `services/diagnostic-aggregator/hosted/service.go`
- `services/diagnostic-aggregator/internal/diagnosticapi/handler.go`
- `services/diagnostic-aggregator/internal/diagnosticreports/service.go`
- `services/diagnostic-aggregator/internal/storage/jsonlstore/`

## Tests

Automated authority remains in:

- `services/diagnostic-aggregator/cmd/diagnostic-submit/e2e_test.go`
- `services/diagnostic-aggregator/hosted/service_test.go`
- `services/diagnostic-aggregator/internal/diagnosticapi/*_test.go`
- `services/diagnostic-aggregator/internal/diagnosticreports/service_test.go`
- `services/diagnostic-aggregator/internal/diagnostics/*_test.go`
- `services/diagnostic-aggregator/internal/redaction/*_test.go`
- `services/diagnostic-aggregator/internal/storage/jsonlstore/*_test.go`

## Related docs

- [Diagnostic Aggregator runtime and report flow](../../services/diagnostic-aggregator/runtime-and-report-flow.md)
- [Diagnostic Aggregator service](../../services/diagnostic-aggregator/!INDEX.md)
- [Operations controls and verification](../../operations/operations-controls-and-verification.md)

## Notes

The collection documents only requests that currently exist in the repository. Add or remove entries here when the Bruno collection changes.
