## Space Rocks P3 hosted deployment

This directory is the complete server deployment payload. The server does not need a Git clone or project source files.

The GHCR packages are private when first published. Before the first pull, either make the three packages public in GitHub Packages or run `docker login ghcr.io` on the server with a token that has `read:packages`.

Copy the top-level `.env.multiplayer-production.example` to `/opt/space-rocks/.env`, then fill in the real values:

```dotenv
SPACE_ROCKS_IMAGE_TAG=p3-hosted
BUILD_VERSION=p3-hosted
POSTGRES_USER=space_rocks
POSTGRES_PASSWORD=replace-with-a-long-random-password
RAILS_MASTER_KEY=replace-with-the-api-server-rails-master-key
SECRET_KEY_BASE=replace-with-a-long-random-secret
GAME_SERVER_INTERNAL_TOKEN=replace-with-a-long-random-token
DISCORD_CLIENT_ID=
DISCORD_CLIENT_SECRET=
DISCORD_REDIRECT_URI=https://api.space-rocks.laughingskull.ca/api/auth/discord/callback
DIAGNOSTIC_AGGREGATOR_TOKEN=replace-with-a-long-random-token
RAILS_LOG_LEVEL=info
LOG_LEVEL=info
SPACE_ROCKS_WEBRTC_ADVERTISED_IPS=replace-with-playit-assigned-address
SPACE_ROCKS_WEBRTC_UDP_PORT_MIN=50000
SPACE_ROCKS_WEBRTC_UDP_PORT_MAX=50019
PLAYIT_SECRET_KEY=replace-with-the-rotated-playit-secret-key
```

Keep it private with `chmod 600 .env`. The real Playit value is never stored in Git. The Compose service maps `PLAYIT_SECRET_KEY` to the container's required `SECRET_KEY` variable.

Run:

```bash
python3 deploy.py
python3 verify.py
```

HTTP services bind only to loopback for the existing Cloudflare Tunnel routes:

- game/WebSocket: `127.0.0.1:8081`
- Rails API: `127.0.0.1:8082`
- diagnostics: `127.0.0.1:8083`

The WebRTC UDP range must match the Playit tunnel allocation. Set `SPACE_ROCKS_WEBRTC_ADVERTISED_IPS` to the address Playit provides.

To update or roll back, change `SPACE_ROCKS_IMAGE_TAG`, then rerun `deploy.py` and `verify.py`.
