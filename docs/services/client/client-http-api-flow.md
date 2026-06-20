## Client HTTP API Flow

Parent index: [Client](./!INDEX.md)

## Purpose

This document describes the client implementation responsibility for shared HTTP API consumption.

It covers the Godot HTTP helper layer, endpoint path configuration, request/result shaping, bearer-token header injection, auth API consumption, player-data profile reads, and local profile management calls.

## Overview

The client HTTP API flow is the Godot-side request boundary for JSON HTTP calls.

The flow has three layers:

```text
Endpoint configuration
= ApiConfig builds concrete HTTP URLs for Rails and game-server data-handler endpoints.

Shared HTTP transport
= ApiHttpClient sends JSON GET/POST/PUT/DELETE requests and returns ApiRequestResult.

Feature API clients
= AuthApiClient, PlayerDataProfileApiClient, and LocalPilotApiClient expose feature-specific calls to auth, profile, and local pilot flows.
````

`ApiHttpClient` owns common request mechanics only. It creates a temporary `HTTPRequest`, attaches JSON headers, optionally attaches a bearer token, serializes non-GET request bodies, parses dictionary JSON responses, and normalizes success or failure into `ApiRequestResult`.

Feature flows do not call `HTTPRequest` directly. They call feature-specific API clients, and those clients depend on the shared HTTP helper.

The client does not own HTTP contract source-of-truth. Request and response shapes are owned by the HTTP contract documentation and backend implementations. The client consumes those contracts through typed wrapper methods and validates only enough structure to keep presentation flows safe.

## Code root

```text
client/
```

Primary implementation areas:

```text
client/scripts/api/
client/scripts/auth/
client/scripts/profile/
client/scripts/ui/menu_flow/
```

## Responsibilities

The client HTTP API flow owns:

* building current client-facing HTTP endpoint URLs
* routing Rails API calls to the Rails API base URL
* routing player-data facade calls to the game-server data-handler base URL
* sending JSON `GET`, `POST`, `PUT`, and `DELETE` requests
* adding `Accept: application/json`
* adding `Content-Type: application/json`
* adding `Authorization: Bearer <token>` when a caller supplies a token
* serializing non-GET request bodies as JSON dictionaries
* accepting empty successful response bodies as successful empty dictionaries
* parsing JSON dictionary responses
* converting request failures into `ApiRequestResult.failure`
* converting non-`2xx` HTTP responses into `ApiRequestResult.failure`
* preserving HTTP status code on success and failure
* exposing feature-specific API methods for auth, profile readout, and local pilot management
* allowing tests to inject fake API clients or fake HTTP clients

## Does not own

The client HTTP API flow does not own:

* HTTP request/response contract source-of-truth
* OpenAPI contract enforcement
* Rails route definitions
* game-server data-handler route definitions
* player-data runtime routing
* local profile persistence
* embedded SQLite storage
* Rails/Postgres account persistence
* OAuth provider secrets
* bearer-token issuance, revocation, expiry, or digest storage
* websocket packet transport
* realtime packet source-of-truth files
* match-result authority
* profile stat mutation
* backing-store selection for Guest, Local Profile, or Authenticated Account identities

## Domain roles

The client HTTP API flow participates in multiple player-facing and platform-facing flows, but only as a client request boundary.

### Auth session role

Auth flows use `AuthApiClient` for Rails auth endpoints:

```text
GET    /api/auth/me
DELETE /api/auth/logout
POST   /api/auth/discord/login_sessions
POST   /api/auth/discord/login_sessions/{login_session_id}/exchange
```

The client uses these calls to validate saved bearer tokens, clear remote sessions on logout, begin Discord browser login, and poll login-session exchange.

The client stores and supplies the bearer token, but Rails owns token validity and account identity.

### Profile readout role

Profile readout uses `PlayerDataProfileApiClient` for:

```text
POST /api/player-data/profile
```

The request includes:

```text
play_mode
identity_kind
local_profile_id
```

Authenticated account reads include the active bearer token. Guest and Local Profile reads do not.

The client consumes the returned profile payload and normalizes stats for display. It does not mutate stats or choose the backing store.

### Local pilot role

Local pilot management uses `LocalPilotApiClient` for:

```text
GET    /api/player-data/local-profiles
GET    /api/player-data/local-profiles/default
POST   /api/player-data/local-profiles
PUT    /api/player-data/local-profiles/{local_profile_id}
DELETE /api/player-data/local-profiles/{local_profile_id}
PUT    /api/player-data/local-profiles/default
```

`LocalPilotFlow` owns the client-side selector behavior and calls these API methods. The player-data service owns persistence and default profile storage.

## Protocols and APIs

### Endpoint configuration

`ApiConfig` currently defines two base URLs:

```text
RAILS_API_BASE_URL = http://localhost:3000
DATA_HANDLER_API_BASE_URL = http://localhost:8080
```

Rails API paths are used for auth and account-facing HTTP calls.

Game-server data-handler paths are used for player-data profile readout and local profile management.

### Shared request behavior

`ApiHttpClient` exposes:

```text
get_json(url, bearer_token = "")
post_json(url, body = {}, bearer_token = "")
put_json(url, body = {}, bearer_token = "")
delete_json(url, body = {}, bearer_token = "")
```

Each method delegates to the shared internal request path.

Request flow:

```text
create HTTPRequest
get SceneTree root
attach request to root
enable request threading
set timeout to 5 seconds
build JSON headers
append bearer Authorization header when supplied
serialize body for non-GET requests
send request
await request_completed
free request node
parse result
return ApiRequestResult
```

### Success behavior

A request returns `ApiRequestResult.success(status_code, body)` when:

* the Godot HTTP request succeeds
* the HTTP status code is `2xx`
* the body is empty, or the body parses as a JSON dictionary

Empty successful response bodies become:

```json
{}
```

This is relevant for endpoints such as successful delete operations that may return no response body.

### Failure behavior

A request returns `ApiRequestResult.failure(status_code, error_message)` when:

* the scene tree or root is unavailable
* `HTTPRequest.request()` fails before sending
* Godot reports a network/request failure
* a non-empty response body is not valid JSON
* a parsed response body is not a dictionary
* the HTTP status code is outside `2xx`

Failure messages use these sources:

```text
scene_tree_unavailable
request_failed
network_failure_<result_code>
invalid_json
parsed error field
parsed message field
http_<status_code>
```

For non-`2xx` JSON responses, `ApiHttpClient` prefers `error`, then `message`, then `http_<status_code>`.

### Bearer-token usage

Bearer tokens are caller-supplied.

`ApiHttpClient` does not load tokens from storage and does not know whether a token is valid. It only appends the header when a non-empty token is passed:

```text
Authorization: Bearer <token>
```

Current token callers include:

* `AuthApiClient.get_current_user(token)`
* `AuthApiClient.logout(token)`
* `PlayerDataProfileApiClient.load_profile(..., token)`
* optional token parameters on `LocalPilotApiClient` methods

## Data ownership

The client HTTP API flow owns transient request and response shaping only.

Client-owned data in this layer:

```text
request URL
request method
request headers
request body dictionary
response status code
response body dictionary
response error message
```

Feature-owned or service-owned data remains outside this layer.

Auth-owned client data:

```text
saved bearer token
in-memory auth session
current signed-in display state
```

Profile-owned client data:

```text
selected identity context
profile readout display payload
normalized display stats
cached authenticated-account stats fallback
```

Player-data-owned backend data:

```text
Guest transient stats
Local Profile records
Local Profile default selection
Local Profile stats
Authenticated Account stats
match-result persistence
```

The client must not infer backing-store ownership from an endpoint URL. Guest, Local Profile, and Authenticated Account routing remains a player-data/runtime responsibility.

## Code map

### Shared HTTP layer

```text
client/scripts/api/api_config.gd
client/scripts/api/api_http_client.gd
client/scripts/api/api_request_result.gd
```

### Auth HTTP consumers

```text
client/scripts/auth/auth_api_client.gd
client/scripts/auth/auth_session_controller.gd
client/scripts/auth/auth_session.gd
client/scripts/auth/auth_token_store.gd
```

### Profile and local pilot HTTP consumers

```text
client/scripts/profile/player_data_profile_api_client.gd
client/scripts/profile/local_pilot_api_client.gd
client/scripts/profile/profile_stats_provider.gd
client/scripts/profile/profile_context_provider.gd
client/scripts/ui/menu_flow/local_pilot_flow.gd
```

### App composition

```text
client/scripts/shell/app_entry.gd
```

`AppEntry` creates one `ApiHttpClient`, adds it to the scene tree, and passes it to auth and profile API clients. `LocalPilotFlow` currently constructs its own `LocalPilotApiClient`, which constructs an HTTP client unless one is injected.

### Backend route boundaries

These files are implementation boundaries consumed by the client, not owned by the client:

```text
services/api-server/config/routes.rb
services/game-server/cmd/game-server/main.go
services/game-server/cmd/game-server/player_data_http.go
services/player-data/httpapi/profile_handler.go
services/player-data/httpapi/local_profiles_handler.go
shared/contracts/http/openapi.yaml
```

### Important non-ownership boundaries

```text
services/api-server/
services/game-server/
services/player-data/
docs/protocol/
docs/data/
```

## Tests

Relevant client tests include:

```text
client/tests/unit/test_auth_session_controller.gd
client/tests/unit/profile/test_player_data_profile_api_client.gd
client/tests/unit/profile/test_profile_stats_provider.gd
client/tests/unit/profile/test_profile_context_provider.gd
client/tests/unit/api/test_api_config.gd
```

Covered behavior includes:

* saved-token validation through the auth API client
* logout clearing local auth state before remote logout completes
* profile API request body construction
* profile API bearer-token forwarding
* Guest profile reads omitting bearer tokens
* local profile context readout behavior
* authenticated profile stat normalization
* exclusion of unrelated or sensitive response fields from profile stats
* fallback to cached authenticated stats after API failure
* zero-stat fallback when profile data is missing

Local pilot API wrapper behavior should get focused tests when local profile API call shapes change.

## Related docs

* [Client](./!INDEX.md)
* [auth-session-flow.md](auth-session-flow.md)
* [local-pilot-flow.md](pregame-menu-flow/local-pilot-flow.md)
* [profile-flow.md](pregame-menu-flow/profile-flow.md)
* [Networking Flow](networking-flow/!INDEX.md)
* [App Shell And Session](app-shell-and-session/!INDEX.md)
* [API-server auth and OAuth](../api-server/auth-and-oauth.md)
* [API-server internal API surface](../api-server/internal-api-surface.md)
* [Player Data](../player-data/!INDEX.md)
* [HTTP Contract Enforcement](../../protocol/http-contract-enforcement.md)

## Notes

Legacy migration material correctly identified that HTTP request and response shapes are not owned by the client, that `shared/contracts/http/openapi.yaml` owns HTTP contracts, and that player-data profile/local-profile calls route through the game-server data-handler facade rather than direct client database or Rails stats access.

`ApiHttpClient` is intentionally small. Feature-specific meaning belongs in `AuthApiClient`, `PlayerDataProfileApiClient`, `LocalPilotApiClient`, or the owning flow docs.

The client currently builds endpoint URLs from constants in `ApiConfig`. Runtime environment selection beyond those constants should be documented where client configuration ownership is formalized.

The `.uid` files adjacent to Godot scripts are not implementation ownership documents and should not be listed as code-map sources.
