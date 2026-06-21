# Space Rocks API Server

## Purpose

This folder is the code seam for the Ruby/Rails API-only service that owns Space Rocks API-server runtime behavior and implementation responsibility.

## What this folder owns

This folder owns the API-server code that runs the Rails service, serves HTTP requests, connects to PostgreSQL, and exposes the service-level health and contract surfaces.

It also owns the local service setup needed to boot, test, and smoke-check the API server.

## What this folder does not own

- Go game-server simulation, room flow, or websocket state.
- Auth, OAuth, or player-stats service details as the canonical documentation source.
- Planning docs or broader product/domain documentation.
- Devtools procedures as the primary home for smoke-flow and Bruno collection details.

## Important files and subfolders

- `config/application.rb` - Rails API-only application configuration.
- `config/routes.rb` - `/health`, `/up`, auth, player-data, and internal route wiring.
- `config/database.yml` - PostgreSQL configuration and `SPACE_ROCKS_API_DATABASE_*` overrides.
- `config/puma.rb` - Puma port configuration and `PORT` override.
- `config/ci.rb` - CI entrypoint and security/test step sequence.
- `app/controllers/health_controller.rb` - `GET /health` implementation.
- `test/controllers/health_controller_test.rb` - `GET /health` coverage.
- `bruno-api/` - Local API smoke collection. Keep usage notes short here; see devtools docs for the working smoke-flow.

## Related documentation

- [API Server docs index](../../docs/services/api-server/!INDEX.md)
- [Auth and OAuth](../../docs/services/api-server/auth-and-oauth.md)
- [Internal API Surface](../../docs/services/api-server/internal-api-surface.md)
- [Player Stats and Match Results](../../docs/services/api-server/player-stats-and-match-results.md)
- [Runtime and Health](../../docs/services/api-server/runtime-and-health.md)
- [HTTP contract enforcement](../../docs/protocol/http-contract-enforcement.md)
- [API Server devtools](../../docs/devtools/api-server/!INDEX.md)
- [Documentation policy](../../docs/documentation-policy.md)
- [Documentation procedure](../../docs/documentation-procedure.md)

## Related tests

- `test/controllers/health_controller_test.rb`
- `test/contracts/openapi_contract_test.rb`
- `test/controllers/api/auth/*`
- `test/controllers/api/player/*`
- `test/controllers/internal/*`

## Notes

Local setup is usually:

- `bundle install`
- `bundle exec rails db:create` if the local database is new
- `bundle exec rails db:migrate`
- `bundle exec rails test`
- `bundle exec rails server`

Discord OAuth secrets live outside git in `.secrets/api-server.env`, usually loaded through `direnv`.

`GET /health` is the service-specific health check. `GET /up` is the Rails boot/load-balancer health route.

For smoke-flow steps and Bruno collection usage, prefer the devtools docs instead of expanding this seam index.
