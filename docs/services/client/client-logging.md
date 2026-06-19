# Client Logging

Parent index: [Client](./!README.md)

## Purpose

This document describes the client logging helper behavior for the client service.

## Overview

`client/scripts/logging/logger.gd` provides the client-side logging helper used by client runtime code.

It defines log levels, category names, default and category-specific log-level controls, and helper methods for emitting formatted log lines. The helper routes warnings through `push_warning`, errors through `push_error`, and informational or debug output through `print`.

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
* enable/disable behavior for the client default level
* routing debug and info messages through `print`
* routing warnings through `push_warning`
* routing errors through `push_error`
* formatting log lines with category and level metadata

## Does not own

The client logging helper does not own:

* server logging policy
* telemetry packet routing
* durable observability storage
* packet schema authority
* gameplay logging semantics
* devtools logging policy

## Behavior

### Log levels

The helper defines these client log levels:

```text
LEVEL_DEBUG
LEVEL_INFO
LEVEL_WARN
LEVEL_ERROR
LEVEL_OFF
```

The default client log level is `LEVEL_INFO`.

### Logger categories

The helper defines category names for common client logging areas:

```text
default
shell
lobby
network
game
world_sync
hud
input
packets
```

These category names let client code emit category-specific logs without re-creating the category string at each call site.

### Level controls

`set_default_level()` changes the default client log level.

`set_category_level()` changes the log level for one category.

`set_all_categories_level()` changes the default level and applies the same level to all known category overrides.

`enable_debug()` sets the default level to debug.

`disable()` sets the default level to off.

### Output routing

`debug()` and `info()` emit formatted lines through `print` when the active level allows them.

`warn()` emits through `push_warning` when the active level allows it.

`error()` emits through `push_error` when the active level allows it.

Category-specific helper methods such as `shell_debug()` and `network_error()` are convenience wrappers over the shared logging methods.

## Code map

Primary implementation files:

```text
client/scripts/logging/logger.gd
```

## Related docs

* [Client](./!README.md)
* [Client Networking Flow](networking-flow/!README.md)
* [Gameplay Runtime](gameplay-runtime/!README.md)
* [Input And Targeting](input-and-targeting.md)
* [Game Server Observability](../game-server/observability/!README.md)
* [Telemetry And Packet Routing](../game-server/networking/telemetry-packet-routing.md)

## Notes

This document captures the current client logging helper behavior only. It does not define server logging policy or telemetry transport behavior.
