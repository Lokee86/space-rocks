# API Server

API server documentation lives here.

## Ownership

This folder owns docs for the API service runtime and its implementation responsibility.

## Does Not Belong

- Domain flow docs.
- Planning docs.
- Direct code maps outside this service index.
- Stub content as canonical service authority.

## Direct Files

- [auth-and-oauth.md](auth-and-oauth.md) - API-server auth, OAuth, bearer-token, and internal token-verification responsibilities.
- [internal-api-surface.md](internal-api-surface.md) - API-server internal service-to-service HTTP surface.
- [player-stats-and-match-results.md](player-stats-and-match-results.md) - API-server player stats and match results documentation.
- [runtime-and-health.md](runtime-and-health.md) - API-server runtime, health checks, database config, Puma port, and CI surface documentation.

## Stub Files

- None.

## Direct Folders

- None.

## Related Docs

- [Services index](../!README.md)

## Notes

This index stays at the API service boundary and does not expand into unrelated product or domain planning detail.
