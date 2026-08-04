---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-78a4-9f4f-7abb266ec5e3
document_type: general
policy_exempt: false
summary: This document describes the runtime measurement path used by the Godot client and game server during a live gameplay room.
---
## Runtime Measurement

Parent index: [Client](./!INDEX.md)

## Purpose

This document describes the runtime measurement path used by the Godot client and game server during a live gameplay room.

## Overview

This document describes the debug-only this boundary surface, its authority boundary, controls, telemetry, runtime gates, implementation owners, and tests.

## Lifecycle

A measurement run is started, sampled, reset, and stopped through request/response packets on the reliable `sr.tooling` channel:

```text
measurement_start -> measurement_started
measurement_snapshot_request -> measurement_snapshot
measurement_reset
measurement_stop -> measurement_stopped
```

The client refuses to start, stop, reset, or request a snapshot while `sr.tooling` is unavailable. This prevents a request from being silently dropped while the coordinator remains stuck in a pending state.

The server owns the run identifier, authoritative tick/entity/process samples, and server packet-write observations. The client owns frame timing, presentation/lane timing, local entity/node/resource samples, lifecycle churn, and client transport observations.

## Client capture boundary

`ClientMeasurementContext` captures the current client transport counters when the server acknowledges the start. Reported cumulative transport counters are then expressed relative to that baseline rather than to process or connection startup.

Run-relative counters include:

```text
packets_in / packets_out
bytes_in / bytes_out
decode_failures / encode_failures / send_failures
bullet-delta receive and timestamp-quality counters
nested WebSocket and per-WebRTC-lane counters
```

Last-value and maximum-value fields remain live values rather than deltas.

The client records a sample immediately at start, periodically while recording, and immediately before stop. This gives the overlay and final report useful boundary samples even for short runs.

## Server capture boundary

The server attaches the measurement observer to the active game when the run starts. Tick summaries, periodic entity/process samples, and packet writes are therefore scoped to that run. Process samples include Go heap/runtime memory, current and peak resident memory, cumulative user/system CPU time, interval CPU utilization expressed as cores consumed, goroutine/GC counts, cumulative GC pause time, interval GC pause time, and the latest GC pause. A disconnect or shutdown finalizes a partial report.

The default one-second sample ring retains 3,600 entries, covering one hour before older samples are overwritten and counted.

## Presentation

The world telemetry overlay shows a compact Measurement section while a run is starting, recording, stopping, or in an error state. It includes elapsed time, client frame average/maximum, server tick average/maximum, authoritative entity counts, client/server packet and byte totals, and server snapshot age.

The devtools window owns the start, stop, reset, scenario-label, status, active-run, and export controls.

## Reports

Stopping a run combines the final client and server reports and writes the client copy beneath:

```text
user://measurements
```

The game server may also persist its server report through its configured report writer. Export failures are reported separately from measurement collection.

## Code map

```text
client/scripts/devtools/measurement/
client/scripts/gameplay/runtime/gameplay_flow_composer.gd
client/scripts/devtools/telemetry/world_telemetry_overlay.gd
services/game-server/internal/measurement/
services/game-server/internal/tooling/controller.go
services/game-server/internal/networking/tooling/router.go
```

## Tests

```text
client/tests/unit/devtools/measurement/
client/tests/unit/devtools/telemetry/test_world_telemetry_overlay.gd
services/game-server/internal/measurement/run_test.go
services/game-server/internal/tooling/controller_test.go
```

## Remaining work

The capture and presentation path is implemented. Repeatable scripted/synthetic scenario orchestration remains a separate runtime-performance slice; scenario labels currently annotate manually controlled runs but do not alter gameplay by themselves.

## Related docs

- [Client](./!INDEX.md)

## Notes

Changes to this boundary should update its canonical owner, code map or source map, verification evidence, and related documentation in the same change.
