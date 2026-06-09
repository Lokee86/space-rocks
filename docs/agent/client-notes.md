# Client Notes

Current client auth flow:

- Main menu signed out state shows `Not Signed In`, hides `LogoutButton`, and shows `Sign-in` on `MultiplayerButton`.
- Main menu signed in state shows the display name, shows `LogoutButton`, and shows `Multi-player` on `MultiplayerButton`.
- Sign-in opens the Discord browser login-session flow.
- The Rails API creates a short-lived login session and returns a poll secret plus login URL.
- The client exchanges the authenticated login session for the normal Space Rocks bearer token.
- The Space Rocks bearer token is stored locally and validated with `GET /auth/me` on startup.
- Logout clears the local token and signed-in state.

Limits and boundaries:

- Single-player stays unchanged and does not require auth.
- Websocket token authentication is future work.
- Non-Discord in-game account creation UI is deferred.
- Do not treat the client as fully authenticated over websocket yet.

Keep this note short and update it when the auth flow changes.
