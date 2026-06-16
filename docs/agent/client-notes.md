# Client Notes

Current client auth flow:

- Menu-flow foundation is implemented and green.
- Single Player pregame Play Endless is implemented and green.
- Main menu keeps the login indicator and `LogoutButton`.
- Main menu Single Player routes to `pregame_menu.tscn` in single-player mode.
- Main menu Multiplayer routes through `MultiplayerEntryFlow`.
- Main menu no longer owns sign-in or multiplayer dialog routing.
- `pregame_menu.tscn` is attached and mode-aware.
- Pregame Back returns to Main Menu.
- Sign In screen is implemented and green.
- Multiplayer pre-lobby Create/Join/Logout routing is implemented and green.
- Pregame Play Endless starts the old single-player flow.
- `PregameMenu` clears when gameplay starts.
- Disabled future Single Player buttons remain disabled.
- See [Client Menu Flow](../client/menu-flow.md) for the canonical menu direction.
- Main Menu Multiplayer routes through `MultiplayerEntryFlow`.
- Signed out opens `LoginWindow`.
- Discord login uses the existing browser login-session flow.
- Signed in opens Multiplayer Pregame.
- Successful Discord auth routes from `LoginWindow` to Multiplayer Pregame.
- Multiplayer Pregame Create uses the existing create-room flow.
- Multiplayer Pregame Join opens `JoinDialog` and uses the existing join-room flow.
- Multiplayer Pregame Logout returns Main Menu signed out.
- Lobby Leave returns to Multiplayer Pregame without logout.
- The Rails API creates a short-lived login session and returns a poll secret plus login URL.
- The client exchanges the authenticated login session for the normal Space Rocks bearer token.
- The Space Rocks bearer token is stored locally and validated with `GET /api/auth/me` on startup.
- After websocket connect, the client sends `authenticate_request` when a Space Rocks bearer token exists.
- `authenticate_result` updates websocket auth state for later multiplayer admission checks.
- Logout clears the local token and signed-in state.

Limits and boundaries:

- Single-player still does not require auth.
- Its launch path now goes through Pregame Menu -> Play Endless.
- Online multiplayer create/join flows are intended for signed-in users, but the server remains the authority for admission.
- Non-Discord in-game account creation UI is deferred.
- `MainMenu` remains dumb.
- `AppEntry` remains wiring/composition only.
- `pregame_menu.gd` is a shell only; flow, controller, and presenter code own the real logic.
- `SessionBootController` chooses the WebSocket target by session mode.
- `SessionNetworkTarget` maps single-player mode to `SINGLE_PLAYER_WS_URL` and multiplayer mode to `MULTIPLAYER_WS_URL`.
- Scene and menu code must not pass raw WebSocket URLs.
- Both targets currently use `/ws` on localhost for development, and the server route remains `/ws`.
- Asteroid variants follow [Asteroid Variant Contract](../design/asteroid-variants.md): use `client/scripts/generated/asteroids/asteroid_variants.gd` for texture lookup, do not reintroduce hardcoded asteroid texture arrays, and keep `index = 0` mapped to `asteroid_1` / `asteroid1.png`.

Keep this note short and update it when the auth flow changes.
