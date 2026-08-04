---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-716c-8ec6-87fe5532cd28
document_type: general
policy_exempt: false
summary: Player-data owns canonical HTTP and dispatcher observability while retaining a transitional category compatibility layer for callers that have not completed semantic migration.
---
# Player-Data Observability And Logging

Parent index: [Player Data](./!INDEX.md)

## Purpose

Player-data owns canonical HTTP and dispatcher observability while retaining a transitional category compatibility layer for callers that have not completed semantic migration.

## Overview

This document describes the current Player-Data Observability And Logging behavior, ownership boundaries, state flow, failure behavior, implementation owners, and verification surfaces.

## Runtime boundary

```text
services/player-data/logging
  -> shared/go/observabilityevent.Emitter
  -> shared/go/servicelog.Runtime
  -> active player-data JSONL writer
```

`logging.Emit` is the canonical logging boundary. The shared Go emitter consumes generated contract metadata and owns validation, redaction, serialization, status, and stable write/rejection outcomes. `servicelog` owns console/file fanout, rolling active output, archive compression, recovery, retention, and degraded writer status.

The process constructs the player-data logging runtime before constructing the player-data runtime itself. The service identity is `player-data`; its identity includes a required name, build version, environment, and instance ID. File output is enabled only after identity validation.

## Compatibility layer

Legacy package-level category methods remain available for existing callers. They route through the legacy compatibility emitter and remain transitional. New semantic events must use `logging.Emit` with a generated event and owning context. Compatibility categories do not define canonical event policy and are not evidence that the canonical boundary is absent.

## HTTP request identity

`services/player-data/httpapi/request_context.go` owns the HTTP boundary. `WithRequestContext` continues a valid UUID `X-Trace-ID`, replaces an absent/invalid trace with a generated UUID, creates one request ID, and stores both values in request context. Only the generated request ID is returned in the `X-Request-ID` response header. Applying the middleware more than once preserves the request identity. Profile and local-profile handlers use these values in canonical HTTP failure events.

The HTTP path excludes raw request bodies, bearer tokens, private profile values, and raw upstream errors from events. It emits stable operation/failure classifications instead.

## Dispatcher match-result ownership

`playerdata.Dispatcher` owns canonical events for packet-dispatched player-data workflows:

- `match_result_report_succeeded` for an accepted new result;
- `match_result_duplicate_suppressed` when the store identifies a duplicate;
- `player_data_read_failed` for stats-load/mode/identity/store failures;
- `player_data_write_failed` for match-result validation/store failures.

These events carry the dispatcher trace, match/result identifiers, packet type, account identity where permitted, and stable error/failure classification. Duplicate suppression is an accepted workflow outcome, not a write failure.

## Local-profile HTTP events

HTTP handlers emit canonical read/write/create failure events for unavailable stores, profile CRUD failures, invalid local-store operations, and profile-stat failures. They include request and trace IDs, operation names, stable error codes, and bounded local-profile identifiers where allowed. They do not include raw errors or private data.

Stable classifications include `invalid_mode_identity`, `store_failure`, `store_unavailable`, `operation_failed`, and the service's stable response error names. `FailureClassOf` is the owner of reusable player-data error classification; callers must not serialize raw error strings into observability.

## Rolling policy and degraded behavior

The shared Go runtime uses an active `.jsonl.open` file and an archive directory. Completed segments follow the shared contract's retention age/size and gzip policy. Startup recovers interrupted active content; rotation is driven by configured size/age limits; retention removes expired or over-budget archives; shutdown flushes and closes the active writer.

Emitter status reports accepted, rejected, redacted, and write-failure counts plus the latest rejection/write outcome. Runtime status reports enabled/degraded state, active path, failure count, and last error. File, archive, compression, recovery, or close failure degrades diagnostics to console/stderr and does not block player-data reads, writes, or HTTP responses.

## Code map

```text
services/player-data/logging/logger.go
services/player-data/observability/generated.go
shared/go/observabilityevent/emitter.go
shared/go/observabilityevent/contract_generated.go
shared/go/servicelog/
services/player-data/httpapi/request_context.go
services/player-data/httpapi/profile_handler.go
services/player-data/httpapi/local_profiles_handler.go
services/player-data/playerdata/dispatcher.go
services/player-data/playerdata/runtime.go
services/player-data/playerdata/configured_runtime.go
```

## Tests

JSONL-backed observability coverage is maintained in:

```text
services/player-data/httpapi/observability_test.go
services/player-data/httpapi/request_context_test.go
services/player-data/playerdata/dispatcher_observability_test.go
services/player-data/logging/logger_test.go
shared/go/observabilityevent/emitter_test.go
```

The focused tests assert canonical event names, trace/request context, duplicate suppression, stable failure fields, accepted/rejected emitter status, JSONL output, and exclusion of raw errors/private payloads.

```bash
cd services/player-data && go test ./...
cd services/player-data && go test -tags noembeddedsqlite ./...
```

## Related docs

- [Observability contract](../../data/observability-contract.md)
- [Canonical event emission](../../observability/canonical-event-emission.md)
- [Runtime and store routing](runtime-and-store-routing.md)
- [Local profiles HTTP API](local-profiles-http-api.md)
- [Match result sinks](match-result-sinks.md)
- [Game-server service startup](../game-server/process/service-startup.md)

## Notes

Changes to this boundary should update its canonical owner, code map or source map, verification evidence, and related documentation in the same change.
