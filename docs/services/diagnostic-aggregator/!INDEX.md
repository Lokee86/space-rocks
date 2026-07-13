# Diagnostic Aggregator

Parent index: [Services](../!INDEX.md)

Diagnostic-aggregator documentation covers the bounded diagnostic-report service runtime and its implementation responsibility.

## Ownership

This service owns triggered diagnostic-report intake, validation, redaction, report construction, bounded storage, retrieval, health reporting, and graceful shutdown.

The current Stage 2 service is an independently runnable Go service. Its default listener is `127.0.0.1:8091`, and its default diagnostic-report retention is 14 days.

## Does Not Belong

- Ordinary service log ownership or continuous log streaming.
- General-purpose log search or analytics.
- Authoritative audit storage.
- Product-wide planning or domain flow documentation.
- Authentication and rate-limit policy that belongs to a future public upload integration.

## Direct Files
<!-- doc-ledger:files:start -->

- [runtime-and-report-flow.md](runtime-and-report-flow.md) - Diagnostic-report intake, processing, storage, retrieval, health, and shutdown boundaries.
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Services index](../!INDEX.md)
- [Observability planning](../../planning/domains/technical/observability-logging-and-diagnostics.md)

## Notes

This service boundary describes diagnostic reports and bounded diagnostic bundles. It does not turn the service into a continuously fed log sink or an audit-record authority.
