---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7465-8b09-bce729abb38e
document_type: general
policy_exempt: false
summary: This document describes the current client observability architecture. The canonical envelope is the primary record model; ClientLogger is the compatibility-facing facade around the canonical emitter and the local rolling writer.
---
# Client Logging

Parent index: [Client](./!INDEX.md)

## Purpose

This document describes the current client observability architecture. The canonical envelope is the primary record model; `ClientLogger` is the compatibility-facing facade around the canonical emitter and the local rolling writer.

## Primary record model

Client production diagnostics are canonical observability records, not the former six-field text record. A record contains the generated contract envelope fields (`timestamp`, `level`, `event`, `event_id`, `service`, `environment`, `build_version`, `schema_version`, `service_instance_id`, `category`, and `retention_tier`) plus validated context and fields when present. Context carries identifiers such as `trace_id`, `request_id`, `session_id`, `room_id`, `player_id`, `account_id`, `match_id`, `route`, `packet_type`, and `duration_ms` according to the generated contract.

`client/scripts/logging/observability_emitter.gd` is the canonical validation and serialization boundary. It consumes `client/scripts/generated/observability/contract_generated.gd` and owns service/event eligibility, required traces, UUID and type checks, field/key/count/size limits, redaction, unsafe-field rejection, canonical JSON serialization, and bounded status. Rejected values are not written. Writer failures return a stable `write_failed` result and increment emitter status rather than raising into gameplay.

## Emission API

Use `ClientLogger.emit_canonical(event_name, message, context, fields)` for new semantic events. It calls the emitter, reports rejection/write-failure status through the client logging surface, and keeps event ownership with the workflow that has the relevant context.

`emit_legacy` remains a compatibility path for existing text/category helpers. `Logger.info`, `network_info`, `shell_warn`, and similar helpers may still route through the bridge event so old callers continue to work, but they are not recommended for new semantic events. Do not add new uses of `network_event(...)` or `packets_event(...)`; meaningful network, packet, boot, room, gameplay, and devtools events should use a generated canonical event with an owning trace.

## ClientOperationTrace ownership

`client/scripts/observability/client_operation_trace.gd` creates a UUID-backed operation trace and exposes its operation name and `trace_id`. A trace is created at the boundary that starts the operation and is passed through its complete workflow:

| Workflow | Trace owner and continuation |
| --- | --- |
| Auth | Auth/session flow creates or continues the auth operation trace across token and result handling. |
| Connection | `ClientConnectionService` owns the connection trace for connect, open, failure, close, and reset events. |
| Boot | Boot flow owns the boot trace and passes it through pending multiplayer/single-player startup. |
| Room | Room/session flow owns create, join, leave, and room-operation traces; request sends and results use that trace. |
| Gameplay | Gameplay/session owners continue the active room/match trace for lifecycle, presentation, and match-boundary diagnostics. |
| Devtools | Devtools command/session flow owns command traces and passes them through server command and result/error handling. |

A lower-level helper must use the supplied owner trace rather than create a second operation identity. `ClientOperationTrace` carries correlation only; event semantics remain with the workflow owner.

## Local rolling output

`ClientLogger.configure_file_output(base_dir, prefix)` creates an active/archive layout:

```text
<base_dir>/active/<prefix>.jsonl.open
<base_dir>/active/<prefix>.jsonl.clean
<base_dir>/archive/<timestamped completed segment>.jsonl
<base_dir>/archive/<timestamped completed segment>.jsonl.gz
```

`client/scripts/logging/rolling_jsonl_writer.gd` owns active writes, age/size rotation, startup recovery of an interrupted `.jsonl.open` file, and retention. Completed segments are compressed by `client/scripts/logging/gzip_archive_compressor.gd`; the uncompressed archive is removed only after the gzip replacement is finalized. The active file is flushed after stored records.

A clean shutdown writes the small `.jsonl.clean` marker after closing the active handle. The next launch removes that marker and reopens the same active segment in append mode, so normal launches do not create one archive each. An active file without the marker is treated as an interrupted segment and recovered into the archive before a fresh active segment opens.

Retention scans the archive once per pass and removes segments by age, total bytes, and a default 256-file ceiling, oldest first. This bounds startup filesystem work even when segments are unusually small. Configuration, rotation, recovery, compression, retention, and close failures degrade to console-only logging with bounded warning reporting; they do not block the client workflow. `current_file_output_path()` reports the active path while output is enabled.

Normal startup requests `user://logs` with the `client` prefix. GUT processes skip persistent application logging, so repeated test scene construction cannot populate the real client archive directory. The resolved OS path depends on Godot's user-data directory; focused writer tests use a temporary user-data root.

## Status and failure behavior

Canonical validation failures expose rejection codes and the rejected key through emitter status. Redacted records increment the redaction counter while storing only the generated replacement marker. A writer or compression failure increments write/failure status and disables the affected file path without turning a diagnostics failure into a gameplay failure. Console compatibility output remains available.

## Code map

Implementation:

```text
client/scripts/logging/observability_emitter.gd
client/scripts/logging/logger.gd
client/scripts/logging/rolling_jsonl_writer.gd
client/scripts/logging/gzip_archive_compressor.gd
client/scripts/observability/client_operation_trace.gd
client/scripts/generated/observability/contract_generated.gd
```

Tests:

```text
client/tests/unit/test_observability_emitter.gd
client/tests/unit/observability/test_client_operation_trace.gd
client/tests/unit/test_client_logger.gd
client/tests/unit/test_rolling_jsonl_writer.gd
client/tests/unit/test_shell_boot_flow.gd
client/tests/unit/test_pending_boot_request.gd
client/tests/unit/test_client_connection_service.gd
client/tests/unit/test_room_session_controller.gd
client/tests/unit/networking/realtime/test_realtime_packet_pipeline_match_boundary.gd
client/tests/unit/devtools/gameplay_debug_flow_test.gd
```

The first group covers emitter contract behavior and status; the second covers trace construction/ownership; the migrated workflow tests cover boot, auth, connection, room, gameplay, and devtools call sites.

## Boundaries

Client logging owns local canonical record construction, compatibility helpers, console presentation, operation correlation, and local rolling JSONL output. It does not own server logs, telemetry transport, diagnostic aggregation, packet schema, or gameplay authority. High-frequency packet diagnostics remain explicitly gated and should not become a substitute for semantic workflow events.

## Related docs

- [Canonical event emission](../../observability/canonical-event-emission.md)
- [Observability contract](../../data/observability-contract.md)
- [WebSocket connection lifecycle](networking-flow/websocket-connection-lifecycle.md)
- [Inbound packet routing](networking-flow/inbound-packet-routing.md)
- [Agent testing](../../agent/testing.md)
- [Developer guide](../../developer.md)
