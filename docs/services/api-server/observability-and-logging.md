---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7d90-b7ea-c206094c832b
document_type: general
policy_exempt: false
summary: The API server owns a canonical observability emitter boundary for request, auth, player-stat, and match-result workflows. It consumes generated Ruby contract metadata and writes bounded local rolling JSONL without sharing active...
---
# API Server Observability And Logging

Parent index: [API Server](./!INDEX.md)

## Purpose

The API server owns a canonical observability emitter boundary for request, auth, player-stat, and match-result workflows. It consumes generated Ruby contract metadata and writes bounded local rolling JSONL without sharing active writers across Puma workers or forked processes.

## Overview

This document describes the current API Server Observability And Logging behavior, ownership boundaries, state flow, failure behavior, implementation owners, and verification surfaces.

## Runtime architecture

```text
Rails request concern
  -> Observability::PumaHooks.emit
  -> Observability::WorkerRuntime
  -> Observability::Emitter
  -> per-process RollingJsonlWriter
  -> ArchiveStore + coordinated retention
```

`services/api-server/app/lib/observability/contract_generated.rb` is generated contract metadata. `Observability::Emitter` validates the canonical event, context, fields, redaction, size, and trace rules before serialization. `Observability::WorkerRuntime` owns boot, recovery, status, degraded behavior, and shutdown for one process/worker. `Observability::PumaHooks` owns Puma lifecycle registration and rejects inherited runtime state after fork.

The request concern `ObservesApiRequest` creates/continues request trace context and calls the hook. Controllers and domain services add auth, authorization, player-stat, and match-result events through the same boundary.

## Puma process and worker ownership

Single-process Puma registers `after_booted` and `after_stopped` hooks. Clustered Puma registers `before_worker_boot` and `before_worker_shutdown`; each worker resolves a worker identity and PID before creating its writer. Active paths include service instance, worker identity, and PID:

```text
<API_OBSERVABILITY_LOG_ROOT>/active/api-server-<instance>-<worker>-pid-<pid>.jsonl.open
```

A runtime records its creating PID. `PumaHooks` drops inherited runtime state when the current PID differs, so a forked worker cannot write through a parent writer or a shared active file. `WriterFactory` refuses to overwrite a surviving active path. Active-file ownership is therefore one writer per worker/process.

## Rolling files, recovery, retention, and degraded behavior

`RollingJsonlWriter` rotates by configured byte or age limit. `ArchiveStore` moves completed `.jsonl.open` content into `archive/`, optionally gzip-compresses it, and removes the uncompressed source only after the compressed file is finalized. Startup recovery finalizes stale active segments and applies retention before opening the new active file.

Retention removes archives older than the configured age and then removes oldest archives until the byte budget is satisfied. `RetentionLock` uses an exclusive lock under `<log-root>/state/retention.lock`, coordinating archive cleanup across clustered workers/processes. Writer, archive, compression, recovery, and close failures set runtime degraded status, report a bounded stderr message, and leave request handling non-blocking; observability failure never becomes an API request failure by itself.

## Configuration

`Observability::ConfigurationFactory` consumes:

```text
API_OBSERVABILITY_ENABLED
API_OBSERVABILITY_LOG_ROOT
API_SERVICE_INSTANCE_ID
BUILD_VERSION
RAILS_ENV
API_OBSERVABILITY_SEGMENT_BYTES
API_OBSERVABILITY_SEGMENT_AGE
API_OBSERVABILITY_RETENTION_AGE
API_OBSERVABILITY_RETENTION_BYTES
API_OBSERVABILITY_COMPRESSION
```

When enabled, `API_SERVICE_INSTANCE_ID` must be a UUID. The generated contract supplies default segment age, retention age, and compression policy. `API_OBSERVABILITY_*` values are configuration inputs, not contract sources.

## Request lifecycle

`ObservesApiRequest` establishes:

- a Rails request ID, with a UUID fallback;
- a trace ID continued from a valid `X-Trace-ID` or generated when absent/invalid;
- a response `X-Trace-ID` containing the resolved trace;
- route and elapsed duration context.

Health controller paths (`health` and `rails/health`) are excluded from request lifecycle events. Other requests emit:

```text
api_request_started
api_request_completed
api_request_failed
```

An action that emits a specific accepted failure marks the request as covered, suppressing the generic duplicate failure in the around-action ensure path. Unhandled exceptions receive one generic `api_request_failed` event. Emitter rejection or writer failure does not replace the HTTP response or raise a second application error.

## Domain event ownership

- **Auth:** auth controllers/services emit canonical authentication success/failure and token/session events with request trace and account context.
- **Authorization:** policy/eligibility owners emit authorization decisions and stable reason codes; raw bearer tokens and private policy inputs are excluded.
- **Player stats:** player-stat reads/mutations own canonical stat request/result/failure events and bounded account context.
- **Match results:** match-result submission/processing owns canonical acceptance, duplicate, validation, persistence, and failure events with result/match identifiers.

The API request concern owns request lifecycle only. It does not duplicate a specific domain failure that was accepted through `PumaHooks.emit`.

## Canonical safety and fallback

The generated Ruby contract defines allowed context fields, redaction rules, free-form limits, stable rejection codes, and write-failure status. `Observability::Emitter` never writes rejected or unsafe raw payloads. Warnings are rate-limited and bounded. The request concern catches emitter failures so the application has a bounded stderr fallback rather than a recursive logging failure.

## Code map

```text
services/api-server/app/lib/observability/emitter.rb
services/api-server/app/lib/observability/worker_runtime.rb
services/api-server/app/lib/observability/puma_hooks.rb
services/api-server/app/lib/observability/contract_generated.rb
services/api-server/app/lib/observability/configuration_factory.rb
services/api-server/app/lib/observability/process_identity.rb
services/api-server/app/lib/observability/writer_factory.rb
services/api-server/app/lib/observability/rolling_jsonl_writer.rb
services/api-server/app/lib/observability/archive_store.rb
services/api-server/app/lib/observability/retention_lock.rb
services/api-server/app/controllers/concerns/observes_api_request.rb
services/api-server/config/puma.rb
```

## Test map

```text
services/api-server/test/lib/observability_emitter_test.rb
services/api-server/test/lib/observability_real_emitter_unit_test.rb
services/api-server/test/lib/observability_configuration_test.rb
services/api-server/test/lib/worker_runtime_test.rb
services/api-server/test/lib/puma_hooks_test.rb
services/api-server/test/lib/writer_factory_test.rb
services/api-server/test/lib/rolling_jsonl_writer_test.rb
services/api-server/test/lib/archive_store_test.rb
services/api-server/test/lib/retention_lock_test.rb
services/api-server/test/lib/observes_api_request_unit_test.rb
```

Run the focused Rails tests from `services/api-server` with the repository's normal bundle/test environment. The observability contract itself is verified by the data-sync validation and drift commands in [Observability contract](../../data/observability-contract.md).

## Related docs

- [API Server](./!INDEX.md)

## Notes

Changes to this boundary should update its canonical owner, code map or source map, verification evidence, and related documentation in the same change.
