# Server Editing
Parent index: [Agent](./!INDEX.md)

## Purpose

This doc guides safe edits to the Go game server.

## Overview

Canonical game-server documentation lives under `docs/services/game-server/`.

## Rules

- Game server owns authoritative simulation outcomes.
- Keep reusable simulation under `services/game-server/internal/game`.
- Do not put reusable simulation in `cmd/game-server/main.go`.
- Networking transports, decodes/routes packets, manages websocket session state, and writes responses.
- Room lifecycle policy belongs in rooms.
- Gameplay simulation belongs in game.
- Packet wire JSON must go through the packet codec.
- Do not add raw packet-path `encoding/json` calls.
- Game server must not directly write Rails/Postgres or embedded SQLite tables.
- Devtools mutations must route through real gameplay seams.
- Use the structured logging wrapper instead of raw logging.
- Server tests belong under `services/game-server/tests/<area>/`.

## Related docs

- [Game Server](../services/game-server/!INDEX.md)
- [Game Server Networking](../services/game-server/networking/!INDEX.md)
- [Game Server Rooms](../services/game-server/rooms/!INDEX.md)
- [Game Server Simulation](../services/game-server/simulation/!INDEX.md)
- [Protocol](../protocol/!INDEX.md)
- [Data](../data/!INDEX.md)
- [Systems Design](../systems-design/!INDEX.md)
- [Devtools Server](../devtools/server/!INDEX.md)
- [Testing](./testing.md)

## Notes

Current runtime facts belong in service, protocol, data, and devtools docs, not this agent guide.
