# Auth Verifier Integration

Parent index: [Game Server Integrations](./!INDEX.md)

## Purpose

This document describes the current game-server integration with API-server bearer-token verification.

It covers how the Go game-server builds the auth verifier, consumes Rails internal token verification, upgrades WebSocket session identity, gates multiplayer room admission, and adapts the same verifier for game-server-hosted player-data profile reads.

## Overview

The game-server does not own authenticated accounts. Rails owns authenticated-account identity, OAuth, access-token issuance, token expiry, token revocation, token digest storage, and internal token verification.

The game-server owns the runtime integration point that consumes Rails verification.

Current flow:

```text
client stores Space Rocks bearer token
-> websocket connects to game-server
-> client sends authenticate_request
-> game-server auth verifier calls Rails /internal/auth/verify-token
-> Rails returns valid account identity or invalid result
-> game-server stores authenticated account identity on the websocket session
-> multiplayer create/join admission checks authenticated session identity
```

WebSocket sessions start as Guest identity. A successful `authenticate_request` can upgrade the session to Authenticated Account identity. The session then carries the Rails user id, canonical `account_id`, and display name returned by Rails.

Multiplayer room creation and join requests require Authenticated Account identity. If the auth verifier is unavailable, the game-server rejects create/join with `auth_unavailable`. If a verifier exists but the session has not authenticated successfully, the game-server rejects create/join with `auth_required`.

Single-player startup does not require auth verification. The current single-player path remains Guest or Local Profile oriented and can run without an auth verifier.

The same game-server verifier is also adapted into the game-server-hosted player-data profile HTTP handler. Authenticated-account profile reads require a valid bearer token and resolve to `account_id` before loading player-data stats.

## Code root

* `services/game-server/`

## Responsibilities

* Build a `networking.TokenVerifier` from game-server environment configuration.
* Return no verifier when auth-verifier configuration is missing.
* Log auth-verifier initialization errors without crashing the process.
* Pass the verifier into WebSocket session construction.
* Route `authenticate_request` packets before lobby and gameplay packet handling.
* Submit the client bearer token to Rails internal token verification.
* Convert a valid Rails verification response into game-server session identity.
* Send `authenticate_result` packets back to the client.
* Gate multiplayer create-room and join-room requests on authenticated session identity.
* Reject multiplayer admission when verification is unavailable.
* Attach verified `account_id` to room members for authenticated multiplayer.
* Adapt the same verifier for player-data profile HTTP reads.
* Keep Rails account authority behind explicit HTTP/service boundaries.

## Does not own

* Rails authenticated-account records.
* OAuth provider login.
* Discord OAuth state.
* Password credentials.
* Bearer-token issuance.
* Bearer-token expiry or revocation.
* Access-token digest storage.
* Direct Rails auth table access.
* Client token persistence.
* Client sign-in presentation.
* Local Profile identity.
* Guest identity policy beyond game-server runtime behavior.
* Player-data backing-store selection.
* Room lifecycle rules beyond admission checks.
* Gameplay player identity.
* Match scoring, match facts, or match-result persistence.
* HTTP contract source ownership.

## Domain roles

### Auth verifier

The auth verifier is the game-server-side object that can verify a Space Rocks bearer token through Rails.

It is configured from:

```text
API_SERVER_BASE_URL
GAME_SERVER_INTERNAL_TOKEN
```

If either value is missing, `buildAuthVerifierFromEnv()` returns `nil`.

If the verifier cannot be constructed, the game-server logs `auth verifier initialization failed` and continues with no verifier.

### TokenVerifier interface

The networking layer depends on this interface:

```go
type TokenVerifier interface {
    VerifyToken(ctx context.Context, rawToken string) (authclient.VerifyResult, error)
}
```

`services/game-server/internal/authclient.Client` is the current implementation.

### WebSocket session identity

Each WebSocket session starts with Guest identity.

```text
guest
```

A successful auth verification changes the session identity to:

```text
authenticated_account
```

Authenticated session identity stores:

```text
AccountUserID
AccountID
DisplayName
```

`AccountUserID` is the Rails internal user id. `AccountID` is the canonical cross-system authenticated-account UUID and is the identity used for account-routed player-data and match-result flows.

### Room admission

The game-server admission check is centralized in `requireAuthenticatedAccount()`.

Current behavior:

| State | Admission result |
| --- | --- |
| nil session | reject without packet |
| no auth verifier | `room_error` with `auth_unavailable` |
| verifier exists but session is not authenticated | `room_error` with `auth_required` |
| session identity is Authenticated Account | allow create/join |

This check is applied to:

```text
create_room_request
join_room_request
```

It is not applied to:

```text
start_single_player_request
```

### Room member identity attachment

When an authenticated session creates or joins a room, the game-server attaches the verified `account_id` to the room member.

The room member keeps account identity separate from gameplay player identity.

```text
websocket session id
!= room member account_id
!= gameplay player id
```

### Player-data profile verifier adapter

The game-server-hosted player-data profile route uses an adapter around `networking.TokenVerifier`.

The adapter converts:

```text
authclient.VerifyResult
```

to:

```text
httpapi.AuthVerificationResult
```

The profile handler uses this only for authenticated-account profile reads. Guest and Local Profile profile reads do not require bearer verification.

## Protocols and APIs

### WebSocket authenticate request

The client sends:

```json
{
  "type": "authenticate_request",
  "token": "<space-rocks-bearer-token>"
}
```

The packet shape is generated from:

```text
shared/packets/lobby.toml
```

The game-server routes this through:

```text
client_packet_router
-> inbound.HandleAuthPacket
-> inboundSessionAdapter.HandleAuthenticateRequest
-> webSocketSession.handleAuthenticateRequest
```

### WebSocket authenticate result

Successful verification sends:

```json
{
  "type": "authenticate_result",
  "authenticated": true,
  "user_id": 1,
  "display_name": "Pilot"
}
```

Failure sends:

```json
{
  "type": "authenticate_result",
  "authenticated": false,
  "error_code": "invalid_token"
}
```

or:

```json
{
  "type": "authenticate_result",
  "authenticated": false,
  "error_code": "token_verification_unavailable"
}
```

Current failure cases:

| Condition | Result |
| --- | --- |
| empty token | `invalid_token` |
| no verifier configured | `token_verification_unavailable` |
| verifier call returns error | `token_verification_unavailable` |
| verifier returns `Valid: false` | `invalid_token` |

The `authenticate_result` packet does not expose `account_id` to the client. The game-server keeps `account_id` internally for room member identity and downstream account-routed flows.

### Rails internal token verification

The game-server auth client calls:

```http
POST /internal/auth/verify-token
Authorization: Bearer <GAME_SERVER_INTERNAL_TOKEN>
Content-Type: application/json
```

Request body:

```json
{
  "token": "<user-bearer-token>"
}
```

Valid response:

```json
{
  "valid": true,
  "user": {
    "id": 1,
    "account_id": "account-uuid",
    "display_name": "Pilot"
  }
}
```

Invalid user-token response:

```json
{
  "valid": false
}
```

The internal bearer token authenticates the game-server as a service caller. The JSON `token` field is the user bearer token being verified.

The auth client trims a trailing slash from the configured base URL and appends:

```text
/internal/auth/verify-token
```

Non-2xx Rails responses become verifier errors. Invalid user tokens are not verifier errors when Rails returns `200` with `valid: false`.

### Timeout behavior

The auth client default HTTP timeout is two seconds when no explicit timeout is configured.

`webSocketSession.handleAuthenticateRequest()` also wraps the verification call in a two-second context timeout.

### Player-data profile HTTP verification

The game-server process hosts:

```text
POST /api/player-data/profile
```

For authenticated-account profile reads, the handler requires:

```http
Authorization: Bearer <user-bearer-token>
```

The handler verifies the token through the same game-server auth verifier adapter. Valid verification provides:

```text
account_id
display_name
```

The profile handler then loads stats for authenticated-account identity through the player-data runtime.

If the verifier is unavailable, the bearer token is missing, verification fails, verification returns invalid, or the returned `account_id` is empty, authenticated-account profile reads return `401 unauthorized`.

Guest and Local Profile profile reads do not use the auth verifier.

## Data ownership

The game-server does not persist auth data.

The game-server temporarily stores verified identity on the WebSocket session:

```text
SessionIdentity.State
SessionIdentity.AccountUserID
SessionIdentity.AccountID
SessionIdentity.DisplayName
```

The game-server may attach `account_id` to room members after authenticated create/join:

```text
RoomMember.AccountID
```

The game-server does not store:

```text
raw bearer tokens
token digests
token expiry
token revocation state
OAuth provider ids
OAuth access tokens
OAuth refresh tokens
password credentials
Rails auth table rows
```

Bearer tokens are runtime proof material only. They are not gameplay player IDs, room member IDs, or durable game identity in the game-server.

## Security and failure behavior

The game-server treats Rails as the auth authority.

Important rules:

* Do not log user bearer tokens.
* Do not log `GAME_SERVER_INTERNAL_TOKEN`.
* Do not persist user bearer tokens in game-server state.
* Do not use Rails `users.id` as the cross-system authenticated account id.
* Use Rails `account_id` for account-routed game-server/player-data flows.
* Do not read Rails auth tables directly.
* Do not let the client choose authenticated-account identity by packet fields other than the bearer token.
* Do not allow multiplayer create/join to bypass `requireAuthenticatedAccount()`.

Verifier unavailable behavior is intentionally explicit:

```text
authenticate_request -> authenticate_result token_verification_unavailable
create_room_request  -> room_error auth_unavailable
join_room_request    -> room_error auth_unavailable
```

This separates a missing service integration from an invalid user token.

Invalid user-token behavior:

```text
authenticate_request -> authenticate_result invalid_token
create_room_request  -> room_error auth_required
join_room_request    -> room_error auth_required
```

## Code map

### Process composition

* `services/game-server/cmd/game-server/main.go`
* `services/game-server/cmd/game-server/auth_config.go`
* `services/game-server/cmd/game-server/player_data_http.go`

### Auth client

* `services/game-server/internal/authclient/client.go`
* `services/game-server/internal/authclient/types.go`

### WebSocket auth and identity

* `services/game-server/internal/networking/websocket.go`
* `services/game-server/internal/networking/websocket_session.go`
* `services/game-server/internal/networking/session_auth.go`
* `services/game-server/internal/networking/session_identity.go`
* `services/game-server/internal/networking/session_admission.go`

### Packet routing

* `services/game-server/internal/networking/client_packet_router.go`
* `services/game-server/internal/networking/inbound/auth.go`
* `services/game-server/internal/networking/inbound_adapter.go`
* `services/game-server/internal/game/packets.go`

### Room admission and member identity

* `services/game-server/internal/networking/room_handlers.go`
* `services/game-server/internal/networking/room_sessions.go`
* `services/game-server/internal/rooms/member.go`
* `services/game-server/internal/rooms/room_members.go`

### Player-data profile adapter

* `services/game-server/cmd/game-server/player_data_http.go`
* `services/player-data/httpapi/profile_handler.go`

### Rails internal auth authority

* `services/api-server/app/controllers/internal/base_controller.rb`
* `services/api-server/app/controllers/internal/auth/verify_tokens_controller.rb`
* `services/api-server/app/services/auth/verify_access_token.rb`
* `services/api-server/app/models/access_token.rb`

### Source and generated contracts

* `shared/packets/lobby.toml`
* `shared/contracts/http/openapi.yaml`
* `services/game-server/internal/game/packets.go`

Important non-ownership boundaries:

* `services/game-server/internal/authclient/` consumes Rails token verification but does not own Rails auth persistence.
* `services/game-server/internal/networking/` owns WebSocket session identity but not account identity authority.
* `services/player-data/httpapi/profile_handler.go` consumes the verifier adapter for authenticated-account profile reads but owns profile response behavior.
* `services/api-server/` owns token verification authority and Rails auth persistence.

## Tests

### Auth client tests

* `services/game-server/internal/authclient/client_test.go`

Coverage includes:

* required config validation
* trailing slash trimming
* default timeout
* request method/path/body/header shape
* valid verification response decoding
* Rails `401` and `500` error handling
* malformed JSON handling
* request context cancellation

### Networking auth tests

* `services/game-server/internal/networking/session_auth_test.go`
* `services/game-server/internal/networking/session_identity_test.go`
* `services/game-server/tests/networking/auth_test.go`
* `services/game-server/tests/networking/auth_admission_test.go`
* `services/game-server/tests/networking/rooms_test.go`

Coverage includes:

* `authenticate_request` success response
* invalid token rejection
* missing verifier response
* authenticated-account identity storage
* account id preserved separately from Rails user id
* create-room requires authentication
* join-room requires authentication
* missing verifier returns `auth_unavailable`
* single-player can start without authentication
* authenticated create-room attaches `account_id` to the room member
* authenticated join-room can join an existing lobby room

### Rails authority tests

* `services/api-server/test/controllers/internal/auth/verify_tokens_controller_test.rb`
* `services/api-server/test/services/auth/verify_access_token_test.rb`
* `services/api-server/test/contracts/openapi_contract_test.rb`

Suggested verification:

```bash
cd services/game-server && go test -buildvcs=false ./internal/authclient ./internal/networking ./tests/networking
cd services/api-server && bundle exec rails test test/controllers/internal/auth/verify_tokens_controller_test.rb test/services/auth/verify_access_token_test.rb test/contracts/openapi_contract_test.rb
```

## Active issues

* `start_single_player_request` does not currently reject an already-authenticated WebSocket session at the server boundary. The intended identity model is still Guest or Local Profile for local single-player, and player-data mode validation rejects `single_player + authenticated_account`, but the WebSocket start-single-player path does not enforce that rejection directly yet. See [Current System Limits](../../../../limits/current-system-limits.md#architecture--networking).

## Related docs

* [Game Server Integrations](./!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Game Server Networking](../../networking/!INDEX.md)
* [Game Server Rooms](../../rooms/!INDEX.md)
* [Client Auth Session Flow](../../../client/auth-session-flow.md)
* [API-server auth and OAuth](../../../api-server/auth-and-oauth.md)
* [API-server internal API surface](../../../api-server/internal-api-surface.md)
* [Player Data](../../../player-data/!INDEX.md)
* [HTTP Contract Enforcement](../../../protocol/http-contract-enforcement.md)
* [Account And Identity Current State](../../../domains/platform/account-and-identity-current-state.md)
* [Account And Identity Systems planning](../../../planning/domains/platform/account-and-identity-systems.md)
* [Current System Limits](../../../../limits/current-system-limits.md)

## Notes

This document intentionally stays on the game-server integration boundary. Rails auth internals belong in API-server docs. Client token storage and sign-in presentation belong in client docs. Broad account, identity, and player-data routing policy belongs in platform domain docs.

Legacy source material identified the correct high-level boundary: the game-server consumes Rails token verification through an explicit service call and must not read Rails auth tables directly. This document rewrites that fact from current game-server, API-server, and player-data code.

The current verifier is optional at process startup because missing `API_SERVER_BASE_URL` or `GAME_SERVER_INTERNAL_TOKEN` produces a nil verifier instead of a process failure. Multiplayer create/join then fails closed with `auth_unavailable`.

The WebSocket `authenticate_result` currently returns Rails `user_id` and `display_name`, but not `account_id`. The game-server still stores `account_id` internally after successful verification.
