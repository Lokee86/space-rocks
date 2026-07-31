---
author: brian
created: "2026-07-31"
document_id: 7e85cf2b-24a1-45d4-8f5e-b338a96061d1
document_type: general
policy_exempt: false
summary: Index of accepted, proposed, superseded, and rejected Space Rocks architectural decisions.
---
# Architecture Decisions

Parent index: [Documentation](../!INDEX.md)

## Purpose

This index records consequential Space Rocks architectural decisions and links them to current canonical documentation.

## Overview

Decision records explain why an expensive, cross-runtime, persistence, protocol, authority, migration, or standards-exception choice was made. Current service, protocol, data, domain, and systems-design documents remain authoritative for what is implemented now.

## Decisions

| ADR | Status | Decision |
| --- | --- | --- |
| [ADR-0001](adr-0001-server-authoritative-gameplay.md) | Accepted | The Go game server owns authoritative gameplay outcomes |
| [ADR-0002](adr-0002-split-control-and-gameplay-transports.md) | Accepted | WebSocket owns control/signaling while lane-specific WebRTC DataChannels own active gameplay delivery |

## Related docs

- [Maintainer map](../maintainer-map.md)
- [Realtime client-server flow](../domains/technical/realtime-client-server-flow.md)
- [Architecture and seam editing rules](../agent/architecture-rules.md)
- [Behavioral contract matrix](../development/behavioral-contract-matrix.md)

## Notes

Accepted ADRs are not rewritten to describe a different decision. A materially changed choice receives a new ADR that supersedes the old record and updates all affected current-state documents.
