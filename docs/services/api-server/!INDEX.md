---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7457-8db8-a6e2496ab4fa
document_type: general
policy_exempt: false
summary: API server documentation lives here.
---
# API Server

Parent index: [Services](../!INDEX.md)

API server documentation lives here.

## Ownership

This folder owns docs for the API service runtime and its implementation responsibility.

## Does Not Belong

- Domain flow docs.
- Planning docs.
- Direct code maps outside this service index.
- Stub content as canonical service authority.

## Direct Files
<!-- doc-ledger:files:start -->

- [auth-and-oauth.md](auth-and-oauth.md) - API-server auth, OAuth, bearer-token, and internal token-verification responsibilities.
- [internal-api-surface.md](internal-api-surface.md) - API-server internal service-to-service HTTP surface.
- [observability-and-logging.md](observability-and-logging.md) - Canonical API request, auth, player-stat, match-result, Puma, and rolling-file observability.
- [player-stats-and-match-results.md](player-stats-and-match-results.md) - API-server player stats and match results documentation.
- [runtime-and-health.md](runtime-and-health.md) - API-server runtime, health checks, database config, Puma port, and CI surface documentation.
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Services index](../!INDEX.md)

## Notes

This index stays at the API service boundary and does not expand into unrelated product or domain planning detail.