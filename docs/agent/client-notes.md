# Client Notes

Current client auth flow:

- Menu-flow foundation is implemented and green.
- Single Player pregame Play Endless is implemented and green.
- Main menu keeps the login indicator and `LogoutButton`.
- Main menu Single Player routes to `pregame_menu.tscn` in single-player mode.
- Main menu Multiplayer routes to `pregame_menu.tscn` in multiplayer mode.
- Main menu no longer owns sign-in or multiplayer dialog routing.
- `pregame_menu.tscn` is attached and mode-aware.
- Pregame Back returns to Main Menu.
- Sign-in moves to a dedicated Sign In screen.
- Pregame Play Endless starts the old single-player flow.
- `PregameMenu` clears when gameplay starts.
- Disabled future Single Player buttons remain disabled.
- See [Client Menu Flow](../client/menu-flow.md) for the canonical menu direction.
- Sign-in opens the Discord browser login-session flow.
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
- `pregame_menu.gd` is a shell only; flow, controller, and presenter code own the real logic.

Keep this note short and update it when the auth flow changes.
