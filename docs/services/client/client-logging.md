# Client Logging

Parent index: [Client](./!INDEX.md)

## Purpose

This document describes the client logging helper behavior for the client service.

## Overview

`client/scripts/logging/logger.gd` provides the client-side logging helper used by client runtime code.

Existing text helpers still work. Calls such as `Logger.network_info("realtime protocol state reset")` continue to emit the same console-friendly lines while routing internally through structured records.

The helper also supports structured event logging, category-specific convenience wrappers for immediate packet and networking work, and optional local JSONL file output for diagnostics.

## Code root

```text
client/
```

Primary implementation area:

```text
client/scripts/logging/
```

## Responsibilities

The client logging helper owns:

* client logging levels
* client logger categories
* default log-level control
* category-specific log-level control
* structured log record construction
* console log formatting
* optional local JSONL diagnostic file output

## Behavior

### Text helpers

Use the existing text helpers for ordinary human-readable status lines:

```gdscript
Logger.network_info("realtime protocol state reset")
```

### Structured event logging

Use structured event logging when the log should carry stable event identity plus fields.

Good fits:

* recurring diagnostics
* network packet send and receive summaries
* threshold warnings
* failures with extra fields

Examples:

```gdscript
Logger.network_event(
	Logger.LEVEL_WARN,
	"packet_decode_failed",
	"Packet decode failed",
	{
		"error": "Invalid JSON",
		"raw_bytes": 42,
	}
)
```

```gdscript
Logger.packets_event(
	Logger.LEVEL_INFO,
	"packet_sent",
	"Packet sent",
	{
		"packet_type": "world_delta",
		"bytes": 384,
	}
)
```

Do not emit high-frequency packet logs by default. Use them for targeted diagnostics, narrow summaries, or temporary investigation windows rather than every routine packet path.

### Local JSONL output

The logger can optionally mirror emitted records to a sequential local JSONL file. This is local diagnostic output, not server logging, telemetry transport, or durable observability storage.

```gdscript
if Logger.configure_file_output("user://logs", "client"):
	Logger.network_event(
		Logger.LEVEL_INFO,
		"realtime_connected",
		"Realtime connected"
	)
```

Output files use sequential names such as `client-000001.jsonl` and `client-000002.jsonl`.

## Code map

Primary implementation files:

```text
client/scripts/logging/logger.gd
```

## Related docs

* [Client](./!INDEX.md)
* [Client Networking Flow](networking-flow/!INDEX.md)
* [Gameplay Runtime](gameplay-runtime/!INDEX.md)
* [Input And Targeting](input-and-targeting.md)

## Notes

This document captures the current client logging helper behavior only. It does not define server logging policy, packet metrics, or telemetry transport behavior.
