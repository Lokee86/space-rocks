# Space Rocks

Space Rocks is an Asteroids-inspired game with a Godot client and a Go game server. A separate API server is planned for business/backend concerns.

## Status

The project is in active development. Current work includes a playable Godot client, a Go websocket game server, room support, shared packet/constants generation, server-authoritative scoring/lives/respawn logic, asteroid collisions/splitting, HUD updates, audio/effects, and structured server logging.

Expect incomplete docs and rough edges while systems are still moving.

## Prerequisites

Install these before running or developing Space Rocks locally:

- **Godot 4.6** for the client project.
  - Open/import the `client/` folder as the Godot project.
  - The configured main scene is `res://scenes/game.tscn`.

- **Go 1.26.3** for the real-time game server.
  - The Go module is in `services/game-server/`.
  - The server entrypoint is `services/game-server/cmd/game-server`.

- **Python 3.10+** for repo tooling and static checks.
  - The data-sync tool uses modern Python typing syntax and requires `tomlkit`.
  - The client constants-boundary test uses `pytest`.

- **Git LFS** for binary/source asset files.
  - The repo tracks asset patterns such as PNG, WEBP, WAV, and MP3 through Git LFS.
  - After cloning, run:

```bash
git lfs install
git lfs pull
```

## Repository Structure

- `client/`: Godot project, scenes, scripts, assets, audio, shaders, and client-side tools.
- `services/game-server/`: Go game server module. The current game server entrypoint is `services/game-server/cmd/game-server`.
- `services/api-server/`: empty placeholder for a planned API server service for business/backend concerns. It is intended to be separate from real-time game simulation.
- `shared/`: sources shared by client/server generation, including TOML constants, TOML packets, and collision shape data.
- `docs/`: Project documentation, including architecture, developer workflow, API plans, devtools, notes, and server logging docs.
- `tools/`: Python tools for syncing shared constants and generating packet code.

## Run Locally

Start the Go game server:

```bash
cd services/game-server
go run ./cmd/game-server
```

Or use Air hot reload if `air` is installed:

```bash
cd services/game-server
air
```

Open the Godot client:

1. Open Godot.
2. Import/open the `client/` folder as a Godot project.
3. Run the project. The main scene is configured as `res://scenes/game.tscn`.

If the `godot` command is available locally, this may also work:

```bash
godot --path client
```

## Development Commands

Run the game server:

```bash
cd services/game-server
go run ./cmd/game-server
```

Run server tests:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Build the game server binary:

```bash
cd services/game-server
go build -buildvcs=false -o ./tmp/game-server ./cmd/game-server
```

Validate shared constants:

```bash
python3 tools/data_sync/main.py -validate -constants
```

Preview shared constants:

```bash
python3 tools/data_sync/main.py -diff -constants -go -gds
```

Apply shared constants:

```bash
python3 tools/data_sync/main.py -push -constants -go -gds
```

Validate shared packets:

```bash
python3 tools/data_sync/main.py -validate -packets
```

Preview shared packets:

```bash
python3 tools/data_sync/main.py -diff -packets -go -gds
```

Apply shared packets:

```bash
python3 tools/data_sync/main.py -push -packets -go -gds
```

## Documentation

- [Architecture](docs/design/architecture.md)
- [NestJS API server plan](docs/api/nestjs-api-server.md)
- [Server logging](docs/server/logging.md)
- [Developer toggles](docs/devtools/toggles.md)

## Assets And Git LFS

Source assets and binary game assets are part of the repo workflow. `.gitattributes` configures Git LFS for asset patterns including PNG, WEBP, WAV, and MP3.

Generated recordings and build artifacts should not be committed. `.gitignore` excludes paths/patterns such as:

- `tmp/`
- `client/.godot/`
- `client/.export/`
- `*.avi`
- Python cache files

## License

No license file is currently present. The project license is not yet specified.
