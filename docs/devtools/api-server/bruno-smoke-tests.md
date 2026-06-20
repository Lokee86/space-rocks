# Bruno Smoke Test

Parent index: [API Server](./!INDEX.md)

## Purpose

This document describes the Bruno collection used for local API-server smoke testing.

It covers the collection root, local environment variables, request order, token capture behavior, OAuth redirect inspection, runtime gates, diagnostics, and the Rails implementation paths exercised by the collection.

## Overview

The Bruno smoke-test collection lives at:

```text
bruno-api/
```

It is a local development and diagnostics tool for exercising the Rails API server over real HTTP requests. The collection targets the API server through `base_url`, which defaults to:

```text
http://localhost:3000
```

The collection currently exercises:

```text
GET    /health
POST   /api/auth/register
POST   /api/auth/login
GET    /api/auth/me
DELETE /api/auth/logout
GET    /api/auth/discord/start
```

The email/password requests use the same Rails controllers and service objects as runtime clients. Bruno does not bypass authentication, write directly to the database, or use a debug-only API path. It sends normal HTTP requests and displays the response status, body, headers, redirects, and script output.

The register and login requests include after-response scripts that require a `token` field in the response body. When the response includes a token, the script stores the raw token into the collection variable `auth_token`. Authenticated requests then attach it as:

```text
Authorization: Bearer {{auth_token}}
```

`auth_token` must contain only the raw Space Rocks bearer token. It must not include the `Bearer ` prefix.

## Debug-only scope

The Bruno collection is a local smoke-test tool.

It is useful for:

```text
manual endpoint checks
local auth verification
token lifecycle checks
response-shape inspection
OAuth redirect inspection
quick regression checks while editing Rails auth code
```

It is not a production monitoring surface, a contract source, a database seed mechanism, or a replacement for Rails tests.

The collection does not define API authority. HTTP request and response shapes are owned by the shared OpenAPI contract, while endpoint behavior is owned by the Rails implementation under `services/api-server/`.

## Server authority

The Rails API server remains authoritative for all behavior observed by the collection.

Bruno exercises the real API-server routes:

```text
GET    /health
POST   /api/auth/register
POST   /api/auth/login
GET    /api/auth/me
DELETE /api/auth/logout
GET    /api/auth/discord/start
```

Rails owns:

```text
health response data
user creation
password credential validation
email normalization
bearer-token issuance
bearer-token digest storage
bearer-token verification
last-used timestamp updates
logout token revocation
Discord OAuth state creation
Discord authorization URL construction
```

Bruno owns only request construction, local variables, request order, and local display of responses.

## Client presentation

The Bruno collection is not the Godot client.

It presents API responses inside Bruno, including response JSON, HTTP status, headers, redirect metadata, and console output from request scripts. For the current collection, the only scripted presentation behavior is token capture from register and login responses.

The `discord oauth start` request is configured with redirects disabled so the Discord authorization `Location` header can be inspected. The browser-driven callback is not completed by that Bruno request.

## Local environment

The local Bruno environment is:

```text
bruno-api/environments/local.yml
```

Current variables:

```text
base_url=http://localhost:3000
email=test@example.com
password=password123
display_name=Test Pilot
auth_token=
```

`base_url` should point at the running Rails API server.

`email`, `password`, and `display_name` are used by the register and login requests.

`auth_token` is written by the register and login after-response scripts and read by the me and logout requests.

If register is run repeatedly with the same email, Rails returns the duplicate-email validation error. Use login after the initial registration, or change the local `email` variable before registering again.

## Smoke-test flow

A normal local smoke check uses this order:

```text
1. health
2. register or login
3. me
4. logout
5. me again with the same token
6. discord oauth start when OAuth environment variables are configured
```

### `health`

Request:

```text
GET {{base_url}}/health
```

Expected success response:

```json
{
  "status": "ok",
  "service": "space-rocks-api"
}
```

This verifies that the Rails API server is reachable at the configured `base_url`.

### `register`

Request:

```text
POST {{base_url}}/api/auth/register
```

Body:

```json
{
  "display_name": "{{display_name}}",
  "email": "{{email}}",
  "password": "{{password}}"
}
```

Expected success behavior:

```text
HTTP 201
response includes token
response includes user
after-response script stores token in auth_token
```

The script throws a Bruno-side error if the response body is missing `token`.

Expected duplicate-email behavior:

```text
HTTP 422
response includes error
auth_token is not updated
```

### `login`

Request:

```text
POST {{base_url}}/api/auth/login
```

Body:

```json
{
  "email": "{{email}}",
  "password": "{{password}}"
}
```

Expected success behavior:

```text
HTTP 200
response includes token
response includes user
after-response script stores token in auth_token
```

The script throws a Bruno-side error if the response body is missing `token`.

Expected bad-credential behavior:

```text
HTTP 401
response includes error
auth_token is not updated
```

### `me`

Request:

```text
GET {{base_url}}/api/auth/me
Authorization: Bearer {{auth_token}}
```

Expected success behavior before logout:

```text
HTTP 200
response includes user
user includes account_id
```

Expected failure behavior for missing, malformed, expired, revoked, or unknown bearer tokens:

```text
HTTP 401
response includes error
```

### `logout`

Request:

```text
DELETE {{base_url}}/api/auth/logout
Authorization: Bearer {{auth_token}}
```

Expected success behavior:

```text
HTTP 204
current bearer token is revoked
```

Logout revokes only the token used for that request. Other active tokens for the same user remain active.

### `me` after logout

Run `me` again with the same `auth_token`.

Expected behavior:

```text
HTTP 401
response includes error
```

This verifies that the logout request revoked the current bearer token and that bearer-token verification rejects revoked tokens.

### `discord oauth start`

Request:

```text
GET {{base_url}}/api/auth/discord/start
```

Expected success behavior when Discord OAuth environment variables are configured:

```text
HTTP 302
Location header points at Discord OAuth authorization
```

Redirect following is disabled for this request so the redirect target can be inspected directly.

Required Rails environment variables:

```text
DISCORD_CLIENT_ID
DISCORD_CLIENT_SECRET
DISCORD_REDIRECT_URI
```

Do not place real Discord client secrets, authorization codes, access tokens, refresh tokens, or callback tokens in Bruno files.

## Diagnostics

The collection exposes diagnostics through normal Bruno request output.

Useful checks:

```text
HTTP status code
response JSON
response headers
redirect Location header
after-response script errors
console output from token capture scripts
auth_token environment value
```

Register and login emit console output when they capture an auth token.

Register and login throw a script error when the response body does not include `token`. This is useful because it makes auth-response shape drift visible during a manual smoke check.

For bearer-token failures, inspect the response status and error body from `me` or `logout`.

For OAuth start failures, confirm that the Rails server has the Discord environment variables loaded before the request is sent.

## Build and runtime gates

The Bruno collection depends on a running Rails API server.

Typical local setup for the API server is owned by the service docs, but the smoke test assumes:

```text
services/api-server dependencies are installed
the Rails database exists
migrations have run
Rails is listening at base_url
```

For auth smoke tests, the API-server database must be writable because register, login, token issuance, token verification, and logout all use Rails persistence.

For Discord OAuth start, these environment variables must be present:

```text
DISCORD_CLIENT_ID
DISCORD_CLIENT_SECRET
DISCORD_REDIRECT_URI
```

The collection is not a CI gate by itself. Rails tests and OpenAPI contract assertions remain the automated verification path.

## Code map

### Bruno collection

```text
bruno-api/opencollection.yml
bruno-api/environments/local.yml
bruno-api/api/folder.yml
bruno-api/api/health.yml
bruno-api/api/register.yml
bruno-api/api/login.yml
bruno-api/api/me.yml
bruno-api/api/logout.yml
bruno-api/api/discord-oauth-start.yml
```

### Rails routes and controllers exercised

```text
services/api-server/config/routes.rb
services/api-server/app/controllers/health_controller.rb
services/api-server/app/controllers/api/auth/registrations_controller.rb
services/api-server/app/controllers/api/auth/sessions_controller.rb
services/api-server/app/controllers/api/auth/me_controller.rb
services/api-server/app/controllers/api/auth/discord_controller.rb
```

### Rails auth implementation exercised

```text
services/api-server/app/controllers/concerns/authenticates_bearer_token.rb
services/api-server/app/controllers/concerns/renders_auth_response.rb
services/api-server/app/services/auth/register_user.rb
services/api-server/app/services/auth/login_user.rb
services/api-server/app/services/auth/issue_access_token.rb
services/api-server/app/services/auth/verify_access_token.rb
services/api-server/app/services/auth/oauth_state_issuer.rb
services/api-server/app/services/auth/providers/discord_config.rb
services/api-server/app/services/auth/providers/discord_authorization_url.rb
services/api-server/app/models/user.rb
services/api-server/app/models/password_credential.rb
services/api-server/app/models/access_token.rb
services/api-server/app/models/oauth_state.rb
```

### Contract source and automated verification

```text
shared/contracts/http/openapi.yaml
services/api-server/test/contracts/openapi_contract_test.rb
services/api-server/test/support/openapi_contract_assertions.rb
services/api-server/test/controllers/health_controller_test.rb
services/api-server/test/controllers/api/auth/registrations_controller_test.rb
services/api-server/test/controllers/api/auth/sessions_controller_test.rb
services/api-server/test/controllers/api/auth/me_controller_test.rb
services/api-server/test/controllers/api/auth/discord_controller_test.rb
```

### Important non-ownership boundaries

```text
Bruno does not own HTTP contracts.
Bruno does not own Rails controller behavior.
Bruno does not own token persistence.
Bruno does not own OAuth provider secrets.
Bruno does not replace Rails tests.
Bruno does not exercise the Go game-server websocket protocol.
```

## Tests

The automated coverage for the behavior smoke-tested by Bruno lives in the Rails test suite.

Relevant tests:

```text
services/api-server/test/controllers/health_controller_test.rb
services/api-server/test/controllers/api/auth/registrations_controller_test.rb
services/api-server/test/controllers/api/auth/sessions_controller_test.rb
services/api-server/test/controllers/api/auth/me_controller_test.rb
services/api-server/test/controllers/api/auth/discord_controller_test.rb
services/api-server/test/contracts/openapi_contract_test.rb
```

The controller tests verify response status, response bodies, token behavior, token revocation, and OpenAPI response alignment.

The Bruno collection is useful as a manual local check after the server boots, after auth code changes, or after local environment changes. It should not be treated as the only verification path for API behavior.

## Related docs

* [API Server Devtools](./!INDEX.md)
* [API Server](../../services/api-server/!INDEX.md)
* [API-server runtime and health](../../services/api-server/runtime-and-health.md)
* [API-server auth and OAuth](../../services/api-server/auth-and-oauth.md)
* [HTTP contract enforcement](../../protocol/http-contract-enforcement.md)

## Notes

The collection root is at repo root under `bruno-api/`, not inside `services/api-server/`. The API-server service README intentionally keeps Bruno details short and points devtool procedure back to this devtools area.

Keep committed Bruno files free of real secrets and real bearer tokens. Local values belong in ignored local environment state, not in the collection source.

The current collection covers the API-server health check, email/password auth lifecycle, current-user lookup, logout revocation behavior, and Discord OAuth start redirect inspection.
