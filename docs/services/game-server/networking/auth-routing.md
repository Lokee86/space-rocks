# Auth Routing

Parent index: [Game Server Networking](./!README.md)

## Purpose

This document describes the game-server networking auth-routing boundary.

It covers how an inbound WebSocket `authenticate_request` packet is classified, routed, verified, converted into session identity, and used by later room admission checks.

## Overview

Game-server auth routing is a WebSocket-session concern.

Every new WebSocket session starts with Guest identity. If the client has a Space Rocks bearer token, it may send an `authenticate_request` packet after WebSocket connection. The game server routes that packet through the inbound packet router, delegates token verification to the configured auth verifier, and upgrades the session to Authenticated Account identity when verification succeeds.

Auth routing does not issue bearer tokens. It does not own OAuth, password login, logout, token persistence, token revocation, or Rails auth storage. Those are API-server responsibilities. The game server only consumes the Rails internal verification boundary through its `TokenVerifier` seam.

Current multiplayer room creation and join requests require the WebSocket session to already hold Authenticated Account identity. If auth verification is not configured, multiplayer create/join returns `auth_unavailable`. If verification is configured but the session has not authenticated, multiplayer create/join returns `auth_required`.

Local single-player still uses Guest or Local Profile identity. Authenticated account routing is for multiplayer admission and authenticated-account player-data routing, not for gameplay player IDs.

## Code root

`services/game-server/`

## Responsibilities

Auth routing owns:

* recognizing inbound `authenticate_request` packets
* routing auth packets before telemetry, lobby, and gameplay packets
* extracting the submitted bearer token from the decoded client packet
* calling the WebSocket session auth handler through the inbound adapter seam
* enforcing a short auth verification timeout at the session boundary
* converting successful verification into session Authenticated Account identity
* sending `authenticate_result` packets back to the client
* preserving Guest identity when auth is missing, unavailable, invalid, or failed
* making authenticated session identity available to room admission
* attaching `account_id` to room membership after authenticated room create/join

## Does not own

Auth routing does not own:

* Rails auth tables
* Rails bearer-token issuance
* Rails bearer-token digest storage
* Rails bearer-token expiry or revocation
* Discord OAuth
* email/password login
* client token persistence
* client signed-in UI state
* Local Profile persistence
* embedded SQLite storage
* player-data store selection
* gameplay player IDs
* room lifecycle rules beyond admission checks
* telemetry ping/pong handling

## Runtime flow

Current WebSocket auth routing flow:

```text
WebSocket connection opens
-> game server creates webSocketSession
-> session identity defaults to Guest
-> client sends authenticate_request
-> readClientInput reads raw WebSocket message
-> inbound envelope decode succeeds
-> handleClientPacket builds inbound session adapter
-> inbound.RouteClientPacket decodes ClientPacket
-> HandleAuthPacket recognizes authenticate_request
-> adapter calls session.handleAuthenticateRequest(token)
-> session calls authVerifier.VerifyToken(...)
-> valid token upgrades session identity
-> session enqueues authenticate_result
```

Auth packets are routed after devtools envelope-only handlers and after full client-packet decode. They are routed before telemetry, lobby, and gameplay handlers.

That ordering keeps authentication independent from room and gameplay packet handling. A client may connect before authentication, but multiplayer create and join paths still reject unauthenticated sessions.

## Packet surface

Auth routing currently uses these realtime packet types from `shared/packets/lobby.toml` and generated game-server packet output:

| Packet type            | Direction             | Purpose                                                       |
| ---------------------- | --------------------- | ------------------------------------------------------------- |
| `authenticate_request` | client to game-server | Submit a Space Rocks bearer token for session authentication. |
| `authenticate_result`  | game-server to client | Report whether the session authenticated successfully.        |

`authenticate_request` contains:

```json
{
  "type": "authenticate_request",
  "token": "<space-rocks-bearer-token>"
}
```

Successful `authenticate_result` contains:

```json
{
  "type": "authenticate_result",
  "authenticated": true,
  "user_id": 123,
  "display_name": "Ada"
}
```

Failed `authenticate_result` contains:

```json
{
  "type": "authenticate_result",
  "authenticated": false,
  "error_code": "invalid_token"
}
```

Current auth failure error codes emitted by the game-server auth handler are:

| Error code                       | Meaning                                                                      |
| -------------------------------- | ---------------------------------------------------------------------------- |
| `invalid_token`                  | The submitted token is empty or Rails verification returned invalid.         |
| `token_verification_unavailable` | The auth verifier is missing or verification failed at the service boundary. |

## Token verification boundary

The game server verifies user bearer tokens through `TokenVerifier`.

The current concrete verifier is `authclient.Client`. It calls the Rails API-server internal endpoint:

```text
POST /internal/auth/verify-token
```

The game server authenticates itself to Rails with:

```text
Authorization: Bearer <GAME_SERVER_INTERNAL_TOKEN>
```

The submitted user bearer token is sent in the JSON request body:

```json
{
  "token": "<user-access-token>"
}
```

Rails returns:

```json
{
  "valid": true,
  "user": {
    "id": 123,
    "account_id": "11111111-2222-3333-4444-555555555555",
    "display_name": "Ada"
  }
}
```

or:

```json
{
  "valid": false
}
```

The game server stores the returned `account_id`, Rails `user_id`, and display name as session identity data when the result is valid.

The bearer token itself is not stored in gameplay state and should not be logged.

## Startup configuration

The game-server process builds the auth verifier from environment variables:

| Environment variable         | Purpose                                                                               |
| ---------------------------- | ------------------------------------------------------------------------------------- |
| `API_SERVER_BASE_URL`        | Base URL for the Rails API-server.                                                    |
| `GAME_SERVER_INTERNAL_TOKEN` | Internal service token used by the game server when calling Rails internal endpoints. |

If either value is missing, the game server starts without an auth verifier.

A missing verifier does not prevent local WebSocket connections or local single-player flow. It does prevent authenticated multiplayer create/join admission because the session cannot prove Authenticated Account identity.

## Session identity

A new WebSocket session starts as:

```text
guest
```

A valid auth result upgrades the session to:

```text
authenticated_account
```

Authenticated Account identity currently stores:

* Rails `user_id`
* cross-system `account_id`
* display name

Identifier separation must be preserved:

* bearer token proves identity, but is not gameplay identity
* `account_id` is the authenticated-account routing identity
* Rails `user_id` remains a Rails database identity
* WebSocket `sessionID` remains session-scoped
* room member identity remains room-scoped
* gameplay player IDs remain gameplay-scoped

Do not replace gameplay player IDs with `account_id` or Rails `user_id`.

## Admission behavior

Auth routing feeds room admission, but it does not own room lifecycle.

Current multiplayer admission behavior:

| Request               | Required identity     | Missing verifier   | Unauthenticated session |
| --------------------- | --------------------- | ------------------ | ----------------------- |
| `create_room_request` | Authenticated Account | `auth_unavailable` | `auth_required`         |
| `join_room_request`   | Authenticated Account | `auth_unavailable` | `auth_required`         |

Room creation uses the authenticated account identity when adding the session member. Room join attaches the authenticated `account_id` after the room manager accepts the join.

Single-player start does not require Authenticated Account identity. Local Profile identity is selected through `local_profile_id` on `start_single_player_request`, not through the auth verifier.

## Data ownership

Auth routing mutates only WebSocket session identity.

It does not directly mutate:

* Rails auth tables
* Rails player-data tables
* embedded SQLite local-profile tables
* player-data runtime stores
* gameplay state

Authenticated Account identity later participates in player-data routing through room membership and match-result reporting, but player-data store selection belongs to `services/player-data`.

## Failure behavior

Auth routing is fail-closed for multiplayer admission.

Failure cases:

| Case                                | Behavior                                                                  |
| ----------------------------------- | ------------------------------------------------------------------------- |
| Empty token                         | Sends failed `authenticate_result` with `invalid_token`.                  |
| Missing verifier                    | Sends failed `authenticate_result` with `token_verification_unavailable`. |
| Verifier error                      | Sends failed `authenticate_result` with `token_verification_unavailable`. |
| Invalid Rails result                | Sends failed `authenticate_result` with `invalid_token`.                  |
| Missing verifier during create/join | Sends room error `auth_unavailable`.                                      |
| Unauthenticated create/join         | Sends room error `auth_required`.                                         |

Auth failure does not close the WebSocket by itself. The session remains connected as Guest unless another flow ends the connection.

## Related docs

* [Game Server Networking](./!README.md)
* [Game Server](../!README.md)
* [Game Server Integrations](../integrations/!README.md)
* [API Server](../../api-server/!README.md)
* [Account And Identity Current State](../../../domains/platform/account-and-identity-current-state.md)
* [Protocol](../../../protocol/!README.md)

## Code map

Primary implementation files:

* `services/game-server/internal/networking/inbound/auth.go`
* `services/game-server/internal/networking/inbound/router.go`
* `services/game-server/internal/networking/client_packet_router.go`
* `services/game-server/internal/networking/inbound_adapter.go`
* `services/game-server/internal/networking/session_auth.go`
* `services/game-server/internal/networking/session_identity.go`
* `services/game-server/internal/networking/session_admission.go`
* `services/game-server/internal/networking/room_handlers.go`
* `services/game-server/internal/networking/room_sessions.go`
* `services/game-server/internal/networking/websocket_session.go`
* `services/game-server/internal/networking/websocket_read.go`

Auth verification integration files:

* `services/game-server/internal/authclient/client.go`
* `services/game-server/internal/authclient/types.go`
* `services/game-server/cmd/game-server/auth_config.go`
* `services/game-server/cmd/game-server/main.go`

Protocol/source files:

* `shared/packets/lobby.toml`
* `services/game-server/internal/game/packets.go`

Related API-server implementation files:

* `services/api-server/config/routes.rb`
* `services/api-server/app/controllers/internal/base_controller.rb`
* `services/api-server/app/controllers/internal/auth/verify_tokens_controller.rb`
* `services/api-server/app/services/auth/verify_access_token.rb`

Related tests:

* `services/game-server/internal/networking/session_auth_test.go`
* `services/game-server/internal/authclient/client_test.go`

Important non-ownership boundaries:

* Rails owns bearer-token issuance, storage, revocation, and verification authority.
* The game server owns WebSocket session identity after verification.
* Rooms own room membership and room lifecycle.
* Gameplay owns game-player IDs and live simulation state.
* services/player-data owns store selection for Guest, Local Profile, and Authenticated Account data.

## Tests and verification

Relevant package tests:

```text
cd services/game-server && go test -buildvcs=false ./internal/networking ./internal/authclient ./cmd/game-server
```

Auth-focused coverage currently includes:

* authclient config validation
* authclient request shape for `POST /internal/auth/verify-token`
* internal bearer header behavior
* valid Rails verification response mapping
* invalid, unauthorized, server-error, malformed JSON, and canceled-context verifier failures
* WebSocket session storage of Rails `account_id`, Rails `user_id`, and display name after successful auth

## Notes

This file is intentionally focused on networking auth routing. The broader product/domain identity model belongs in Account And Identity domain documentation. API auth implementation detail belongs in API-server service documentation. External Rails verification integration detail belongs in game-server integrations documentation.
