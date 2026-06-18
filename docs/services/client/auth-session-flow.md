# Auth Session Flow

Parent index: [Client](./!README.md)

## Purpose

This document describes the current client-side auth session and session boot flow implemented under `client/`.

It documents how the Godot client stores the Space Rocks bearer token, validates saved session state, starts the Discord browser handoff, updates menu/auth presentation, sends websocket authentication, and gates multiplayer boot requests.

## Overview

The client owns auth session presentation and local token handling. It does not own account identity, bearer-token verification, OAuth provider secrets, multiplayer admission, or persistent account data.

The current client auth flow is centered on three layers:

```text
Auth session state
= in-memory signed-in state plus the bearer token loaded from local storage

HTTP auth client
= Rails API auth endpoints consumed by Godot

Session boot flow
= websocket connection, optional authenticate_request, and pending room/game request dispatch
```

On startup, `AppEntry` creates an `AuthSessionController`, wires it into the main menu, profile providers, and connection service, then calls `initialize_from_saved_token()`.

If no saved token exists, the controller clears auth state and emits `auth_state_changed`. If a saved token exists, the controller validates it through Rails `GET /api/auth/me`. A valid response repopulates `AuthSession`. An invalid or failed response clears the saved token and signs the client out.

Current user-facing sign-in is Discord-only. The sign-in window has disabled manual email/password and Google controls. Pressing the Discord button starts a Rails login-session handoff: the client asks Rails for a login session, opens the returned browser URL, polls the exchange endpoint, receives the normal Space Rocks bearer token, stores it in `user://auth_token.json`, and updates the in-memory session.

Multiplayer entry is gated by client signed-in state. If the player requests multiplayer while signed out, the menu routes to the sign-in screen. If the player is already signed in, or becomes signed in while on the sign-in screen, the menu routes to multiplayer pregame.

When a websocket connects, the client sends `authenticate_request` if the current auth session has a token. Multiplayer create/join boot requests wait for websocket auth success before being sent. Single-player boot requests do not wait for websocket auth and remain Guest/Local Profile oriented at the client flow level.

## Code root

* `client/`
* `client/scripts/`

## Responsibilities

* Store and clear the current Space Rocks bearer token locally.
* Load a saved token during app startup.
* Validate saved tokens through `GET /api/auth/me`.
* Maintain in-memory signed-in state for menu and profile presentation.
* Start Discord browser-assisted login sessions.
* Open the browser login URL returned by Rails.
* Poll the Discord login-session exchange endpoint until authenticated, failed, canceled, or timed out.
* Save the bearer token returned by a successful login-session exchange.
* Clear local auth state on auth failure.
* Clear local auth state immediately on logout.
* Call Rails logout when a token existed at logout time.
* Route signed-out multiplayer entry to the sign-in screen.
* Route signed-in multiplayer entry to multiplayer pregame.
* Provide auth state to profile context and stats providers.
* Attach bearer tokens to HTTP requests when the caller provides one.
* Send websocket `authenticate_request` when a connected websocket has a current auth token.
* Hold multiplayer boot requests until websocket auth succeeds or token verification is unavailable.
* Leave single-player startup independent of signed-in state.

## Does not own

* Rails authenticated account records.
* Password credentials.
* OAuth provider identity records.
* OAuth state verification.
* OAuth client secrets.
* Discord access tokens or refresh tokens.
* Bearer-token issuance, expiry, revocation, or digest storage.
* Backend token verification.
* JWT or future token format decisions.
* Go game-server websocket session identity.
* Multiplayer admission policy.
* Room membership authority.
* Gameplay player identity.
* Match result authority.
* Persistent account stats.
* Local Profile persistence.
* Guest transient player-data persistence.
* OpenAPI contract ownership.
* Realtime packet source-of-truth ownership.

## Domain roles

### Auth session

`AuthSession` is the client's in-memory view of signed-in state.

It currently stores:

```text
signed_in
token
user_id
display_name
email
```

`user_id` is the Rails user id returned in the auth payload. The client session does not currently store `account_id`.

### Auth token store

`AuthTokenStore` persists the bearer token to:

```text
user://auth_token.json
```

The file contains:

```json
{ "token": "..." }
```

The store only loads, saves, and clears the token. It does not encrypt the file, rotate tokens, validate tokens, or store provider credentials.

### Auth API client

`AuthApiClient` wraps the Rails auth HTTP calls used by the Godot client:

```text
GET    /api/auth/me
DELETE /api/auth/logout
POST   /api/auth/discord/login_sessions
POST   /api/auth/discord/login_sessions/{id}/exchange
```

It depends on `ApiHttpClient` for JSON request behavior.

### Auth session controller

`AuthSessionController` coordinates saved-token validation, Discord login-session polling, token persistence, session updates, logout, and auth signals.

It emits:

```text
auth_state_changed
auth_error(message)
```

### Sign-in UI flow

The sign-in UI owns presentation and button routing only.

Manual email/password controls and Google login are currently disabled in the client. Discord login is the implemented sign-in path.

### Multiplayer entry flow

`MultiplayerEntryFlow` checks current auth session state before multiplayer pregame routing.

Signed-out players are sent to the sign-in screen. Signed-in players are sent to multiplayer pregame.

### Websocket auth flow

The websocket auth flow proves an already-stored bearer token to the game server.

The client sends:

```json
{
  "type": "authenticate_request",
  "token": "<space-rocks-bearer-token>"
}
```

The game server responds with:

```json
{
  "type": "authenticate_result",
  "authenticated": true,
  "user_id": 1,
  "display_name": "Pilot"
}
```

or a failed result with `authenticated: false` and an `error_code`.

The client stores the websocket auth result as connection state. The game server owns the verified authenticated-account session identity.

## Protocols and APIs

### Rails auth HTTP flow

The client consumes Rails auth endpoints through `AuthApiClient` and `ApiHttpClient`.

`ApiHttpClient` behavior:

* Creates a temporary `HTTPRequest`.
* Adds `Accept: application/json`.
* Adds `Content-Type: application/json`.
* Adds `Authorization: Bearer <token>` when the caller provides a token.
* Serializes non-GET request bodies as JSON.
* Parses JSON dictionary responses.
* Returns `ApiRequestResult.success` for `2xx` responses.
* Returns `ApiRequestResult.failure` for request errors, network failures, invalid JSON, or non-`2xx` HTTP responses.

### Saved-token validation

Startup validation flow:

```text
AppEntry._ready()
AuthSessionController.initialize_from_saved_token()
AuthTokenStore.load_token()
if token exists:
    GET /api/auth/me with bearer token
    if valid:
        AuthSession.set_signed_in(token, user)
    else:
        AuthTokenStore.clear_token()
        AuthSession.clear()
emit auth_state_changed
```

Invalid saved tokens are cleared locally.

### Discord login-session handoff

Discord sign-in flow:

```text
LoginWindow.discord_login_requested
SignInFlow calls request_discord_sign_in
AuthSessionController.request_discord_sign_in
POST /api/auth/discord/login_sessions
OS.shell_open(login_url)
repeat until success, failure, cancel, or timeout:
    POST /api/auth/discord/login_sessions/{id}/exchange
    body: { "poll_secret": "<poll-secret>" }
```

The controller polls once per second for up to 120 seconds.

`202` means the browser login session is still pending. A successful exchange must return a non-empty `token` and a dictionary `user` payload. On success, the client saves the token, updates `AuthSession`, and emits `auth_state_changed`.

Begin-session failure, malformed response, failed exchange, or timeout clears local auth state and emits `auth_error`.

### Logout flow

Logout flow:

```text
AuthSessionController.logout()
load current saved token
clear local token
clear in-memory auth session
emit auth_state_changed
if token existed:
    DELETE /api/auth/logout with bearer token
```

Remote logout is best-effort from the client perspective. Local state is cleared immediately before the remote call finishes.

### Websocket authentication

Websocket auth flow:

```text
ClientConnectionService._on_connected()
NetworkClient.send_authenticate_request(token) if session token exists
emit connected
SessionNetworkController handles pending boot request
```

Single-player behavior:

* `start_single_player_request` is sent when connected.
* The client does not wait for websocket auth before sending single-player boot.
* The single-player context comes from Guest or Local Profile selection.

Multiplayer behavior:

* `create_room_request` and `join_room_request` are held as pending boot requests.
* If the connection is already websocket-authenticated, the pending request is sent.
* If not authenticated yet, the client waits for `authenticate_result`.
* On successful websocket auth, the pending multiplayer request is sent.
* On `token_verification_unavailable`, the pending multiplayer request is still sent so server-side admission can return the appropriate failure.
* On invalid token or other auth failure, the pending multiplayer request remains unsent.

### Realtime packet source

The realtime auth packets are generated from the shared lobby packet source:

```text
shared/packets/lobby.toml
```

Generated Godot packet helpers live at:

```text
client/scripts/generated/networking/packets/packets.gd
```

## Data ownership

The client owns only local auth/session presentation state and the locally saved bearer token.

Client-owned local state:

* `AuthSession.signed_in`
* `AuthSession.token`
* `AuthSession.user_id`
* `AuthSession.display_name`
* `AuthSession.email`
* `user://auth_token.json`

The client does not own durable account identity.

Important identity separation:

* `AuthSession.user_id` is the Rails user id from auth payloads.
* `account_id` is the canonical cross-system authenticated-account UUID.
* The current `AuthSession` does not store `account_id`.
* Bearer tokens prove identity, but they are not gameplay identity.
* Gameplay player IDs remain server/gameplay scoped.

The client must not store:

* Discord access tokens.
* Discord refresh tokens.
* OAuth client secrets.
* OAuth provider secrets.
* OAuth state.
* Login-session poll secrets after the exchange flow.
* Game-server internal service tokens.
* Rails token digests.
* Backend auth table data.

## Security and failure behavior

The saved bearer token is a valid credential while it remains unexpired and unrevoked. Treat `user://auth_token.json` as sensitive local state.

Current failure behavior:

* Missing saved token signs the client out.
* Invalid saved token signs the client out and clears the local token.
* Failed Discord login-session creation clears local auth state.
* Malformed login-session creation response clears local auth state.
* Failed login-session exchange clears local auth state.
* Login-session polling timeout clears local auth state.
* Logout clears local state immediately.
* Websocket `invalid_token` prevents pending multiplayer boot from being sent.
* Websocket `token_verification_unavailable` allows the pending multiplayer boot request to reach server-side admission.

Logging rules:

* Do not log bearer tokens.
* Do not log login-session poll secrets.
* Do not log OAuth codes.
* Do not log OAuth state.
* Do not log provider secrets.

## Code map

### Composition and startup

* `client/scripts/shell/app_entry.gd`
* `client/scripts/main_menu/main_menu_session_controller.gd`

### Auth session

* `client/scripts/auth/auth_session.gd`
* `client/scripts/auth/auth_token_store.gd`
* `client/scripts/auth/auth_api_client.gd`
* `client/scripts/auth/auth_session_controller.gd`

### HTTP API consumption

* `client/scripts/api/api_config.gd`
* `client/scripts/api/api_http_client.gd`
* `client/scripts/api/api_request_result.gd`

### Menu and sign-in UI

* `client/scripts/ui/menus/main_menu.gd`
* `client/scripts/ui/menu_flow/menu_flow_controller.gd`
* `client/scripts/ui/menu_flow/multiplayer_entry_flow.gd`
* `client/scripts/ui/sign_in/login_window.gd`
* `client/scripts/ui/sign_in/sign_in_flow.gd`
* `client/scenes/ui/main_menu.tscn`
* `client/scenes/ui/dialogs/login_window.tscn`

### Profile/session consumers

* `client/scripts/profile/profile_context_provider.gd`
* `client/scripts/profile/profile_identity_kind.gd`
* `client/scripts/profile/profile_stats_provider.gd`
* `client/scripts/profile/player_data_profile_api_client.gd`

### Session boot

* `client/scripts/boot/session_boot_controller.gd`
* `client/scripts/boot/shell_boot_flow.gd`
* `client/scripts/boot/pending_boot_request.gd`
* `client/scripts/boot/session_network_target.gd`
* `client/scripts/session/client_session_context.gd`
* `client/scripts/session/session_network_controller.gd`

### Websocket client and packet routing

* `client/scripts/networking/client_connection_service.gd`
* `client/scripts/networking/network_client.gd`
* `client/scripts/networking/inbound/server_packet_dispatcher.gd`
* `client/scripts/networking/inbound/server_packet_router.gd`
* `client/scripts/generated/networking/packets/packets.gd`
* `shared/packets/lobby.toml`

### Non-owning backend consumers and authorities

* `services/game-server/internal/networking/session_auth.go`
* `services/game-server/internal/networking/inbound/auth.go`
* `services/game-server/internal/networking/inbound_adapter.go`
* `services/game-server/internal/authclient/client.go`
* `services/game-server/internal/authclient/types.go`
* `services/api-server/config/routes.rb`
* `shared/contracts/http/openapi.yaml`

These backend paths are listed for boundary clarity. The client does not own them.

## Tests

### Auth state and token storage

* `client/tests/unit/test_auth_session.gd`
* `client/tests/unit/test_auth_token_store.gd`
* `client/tests/unit/test_auth_session_controller.gd`

### Sign-in UI and menu routing

* `client/tests/unit/ui/sign_in/test_login_window.gd`
* `client/tests/unit/ui/sign_in/test_sign_in_flow.gd`
* `client/tests/unit/ui/menu_flow/test_multiplayer_entry_flow.gd`
* `client/tests/unit/ui/menus/test_main_menu_auth_state.gd`

### Session boot and websocket auth gating

* `client/tests/unit/test_pending_boot_request.gd`
* `client/tests/unit/test_shell_boot_flow.gd`
* `client/tests/unit/test_session_network_controller.gd`
* `client/tests/unit/boot/test_session_network_target.gd`

### HTTP config and profile consumers

* `client/tests/unit/api/test_api_config.gd`
* `client/tests/unit/profile/test_profile_context_provider.gd`
* `client/tests/unit/profile/test_profile_stats_provider.gd`
* `client/tests/unit/profile/test_player_data_profile_api_client.gd`

## Active issues

* Manual email/password login and Google login are disabled in the current client sign-in UI. See [Current System Limits](../../limits/current-system-limits.md#client-menu-flow).
* `start_single_player_request` does not currently reject an already-authenticated websocket session at the server boundary. The client single-player flow still uses Guest or Local Profile context and does not wait for auth, but the server-side enforcement gap remains active. See [Current System Limits](../../limits/current-system-limits.md#architecture--networking).

## Related docs

* [Client](../!README.md)
* [Account And Identity Current State](../../domains/platform/account-and-identity-current-state.md)
* [API-server auth and OAuth](../api-server/auth-and-oauth.md)
* [API-server internal API surface](../api-server/internal-api-surface.md)
* [Game Server](../game-server/!README.md)
* [Player Data](../player-data/!README.md)
* [HTTP Contract Enforcement](../../protocol/http-contract-enforcement.md)
* [Current System Limits](../../limits/current-system-limits.md)

## Notes

Legacy source material correctly identified the browser-assisted Discord login-session handoff, saved Godot bearer token path, `/api/auth/me` validation, websocket auth packet flow, and the rule that single-player remains independent of Rails auth. This document rewrites those facts from current client code and current canonical API/domain docs.

The current Rails auth response used by login-session exchange does not include `account_id`. `GET /api/auth/me` includes `account_id`, but `AuthSession.set_signed_in()` currently stores only `id`, `display_name`, and `email` from the user payload.

The connection service sends websocket `authenticate_request` whenever the current auth session has a token. Session boot behavior decides whether a pending request waits for auth. Multiplayer waits; single-player does not.

Token hardening beyond the current `user://auth_token.json` store belongs in planning or limits documentation, not in this client service implementation doc.
