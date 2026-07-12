# Client Logging

Parent index: [Client](./!INDEX.md)

## Purpose

This document is the canonical reference for the current client logging helper implementation.

## Overview

`client/scripts/logging/logger.gd` provides the client-side logging helper used by client runtime code.

Existing text helper calls still emit console-compatible lines, but they do so through structured records with event `log_message`. Structured logging is the core behavior, not a wrapper around a separate text-only logger.

The helper supports structured event logging, category-specific convenience wrappers, and optional local JSONL file output for diagnostics.

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

### Levels

The implemented levels are `debug`, `info`, `warn`, `error`, and `off`.

`LEVEL_DEBUG`, `LEVEL_INFO`, `LEVEL_WARN`, and `LEVEL_ERROR` map to the matching lowercase level names. `LEVEL_OFF` is treated as disabled by `_should_log(...)` and `level_name(LEVEL_OFF)` currently returns `unknown`.

The logger compares the requested level against the active level for the category. Records are emitted when the requested level is at least the active level and the active level is not `LEVEL_OFF`.

### Level control

`default_level` is the global fallback threshold.

`category_levels` stores per-category overrides.

`set_default_level(level)` updates the fallback threshold.

`set_category_level(category, level)` updates one category override.

`set_all_categories_level(level)` updates the fallback threshold and all currently stored category overrides.

`enable_debug()` sets the fallback threshold to `LEVEL_DEBUG`.

`disable()` sets the fallback threshold to `LEVEL_OFF`.

`_should_log(category, level)` checks the category override when present, otherwise the fallback threshold. `LEVEL_OFF` suppresses output when it is the active threshold.

### Categories

The implemented categories are:

* `default`
* `shell`
* `lobby`
* `network`
* `game`
* `world_sync`
* `hud`
* `input`
* `packets`

If a category does not have an override, it uses the current default level.

### Current gameplay diagnostic usage

Gameplay and presentation diagnostics route through structured `ClientLogger` events rather than direct client `print(...)` calls. Event-batch diagnostics use the `packets` category, gameplay death diagnostics use `game`, HUD lifecycle diagnostics use `hud`, and pickup presentation failures use `world_sync`. These diagnostics use stable event names and structured fields; high-frequency event-batch activity remains debug-level and targeted.

### Text helpers

Use the existing text helpers for ordinary human-readable status lines. They still build structured records with event `log_message`.

Example:

```gdscript
Logger.network_info("realtime protocol state reset")
```

The category helpers such as `shell_info(...)` and `network_warn(...)` still produce console-compatible lines. `warn(...)` routes through `push_warning(...)`, `error(...)` routes through `push_error(...)`, and the other helper levels print through the normal console path.

### Structured event logging

Use `event(...)` for category-specific structured logs, `network_event(...)` for network-category structured logs, and `packets_event(...)` for packet-category structured logs.

Use stable event names and fields when the log represents a recurring diagnostic, threshold warning, or failure with structured context. Keep high-frequency packet logs gated or targeted rather than emitted for every routine packet path.

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

### Record schema

Structured records currently contain these fields:

* `timestamp_unix_ms`
* `level`
* `category`
* `event`
* `message`
* `fields`

`fields` is stored as a deep duplicate of the caller-provided dictionary, so later caller mutation does not mutate the stored record.

### Console format

Console lines are formatted with category and level brackets first, for example `[network][info]`.

For non-`log_message` events, the console line includes an event bracket such as `[packet_decode_failed]` before the message.

Field output is deterministic because dictionary keys are sorted before formatting. Dictionaries and arrays are JSON-stringified for display.

### JSONL output

`format_json_line(...)` returns one JSON object per line. This is local diagnostic output only.

### Local JSONL output

The logger can optionally mirror emitted records to a sequential local JSONL file. This is local diagnostic output, not server logging, telemetry transport, or durable observability storage.

`configure_file_output(base_dir, prefix)` creates numbered files such as `client-000001.jsonl` and `client-000002.jsonl` in the chosen base directory. The current default base directory is `user://logs` and the current default prefix is `client`.

`configure_file_output(...)` closes any existing file output before opening a new file. It creates the base directory when possible, then picks the first available sequential filename from `prefix-000001.jsonl` upward.

`current_file_output_path()` returns the active path when file output is enabled, and `close_file_output()` flushes and closes the handle, then clears the output state.

`configure_file_output(...)` returns `false` when directory creation or file creation fails. When file output is active, emitted records are written as JSONL and flushed after each stored line.

`AppEntry._ready()` currently calls `ClientLogger.configure_file_output("user://logs", "client")`, so normal client startup attempts to enable structured JSONL file logging under Godot user data. On success, `AppEntry` logs the active path with `ClientLogger.shell_info(...)`; on failure, it logs that client structured log file output is unavailable with `ClientLogger.shell_warn(...)`. `current_file_output_path()` is the source for the active path while file output is enabled.

### Test coverage

`client/tests/unit/test_client_logger.gd` covers:

* level-name mapping
* record construction
* field duplication
* JSONL formatting
* console formatting for text helpers and named events
* deterministic field sorting
* category override behavior
* numbered filename generation
* file-output configuration
* JSONL emission
* file-output close/reset behavior


## Boundary

Client logging is not server logging.

Client JSONL is not telemetry transport.

Client JSONL is not durable observability aggregation.

High-frequency packet logs should remain gated or targeted.

## Code map

Primary implementation files:

```text
client/scripts/logging/logger.gd
```

## Related docs

* [Client](./!INDEX.md)
* [Client Networking Flow](networking-flow/!INDEX.md)
* [Gameplay Runtime](gameplay-runtime/!INDEX.md)
* [Developer Guide](../../developer.md)
* [Agent Testing](../../agent/testing.md)
* [Input And Targeting](input-and-targeting.md)

## Notes

This document captures the current client logging helper behavior only. It does not define server logging policy, packet metrics, or telemetry transport behavior. For practical log-file lookup and focused test guidance, see [Developer Guide](../../developer.md) and [Agent Testing](../../agent/testing.md).
