# Server Devtools

Focused reference for the current server devtools command surface and boundaries.

## Purpose

Devtools are for controlled gameplay debugging in active sessions. Commands are client-triggered by packet and applied only by server-owned gameplay seams.

## Enable/Disable Boundary

- default server builds handle devtools commands
- `nodevtools` server builds disable devtools command handling through the existing devtools gate
- disabled builds ignore/reject devtools command packets before normal game handling

## Packet/Generation Sources

- packet schema source: `shared/packets/debug.toml`
- packet output routing source: `shared/packets/outputs.toml`
- generated server packet types: `services/game-server/internal/devtools/packets_generated.go`
- generated client packet helpers: `client/scripts/networking/packets/packets.gd`
- regenerate packets: `data-sync -push -packets -go -gds`

## Server Command Path

- `services/game-server/internal/networking/websocket_read.go` detects devtools command packets with `devtools.ShouldHandleCommand`
- matching packets route to `services/game-server/internal/devtools.HandleCommand`
- devtools commands do not route through `Game.HandlePacket`
- `services/game-server/internal/devtools` owns command dispatch and handlers

## Client UI Path

- devtools UI originates from `client/scripts/devtools/`
- UI actions send generated debug packets through the normal networking send path
- client devtools UI does not apply gameplay mutations locally

## Current Command List

- `toggle_debug_invincible`
- `toggle_debug_infinite_lives`
- `toggle_debug_freeze_world`
- `toggle_debug_freeze_player`
- `debug_kill_player`
- `debug_spawn_player`
- `debug_respawn_player`
- `debug_spawn_entity` (existing spawn controls)
- `debug_set_score`
- `debug_add_score`
- `debug_set_lives`
- `debug_add_lives`
- `debug_clear_bullets`
- `debug_clear_asteroids`

Target behavior:

- commands use `target_player_id` where applicable
- score/lives commands target active players
- clear bullets and clear asteroids are room/global commands

## Safety Rules

- server owns all gameplay mutations
- client devtools UI sends packets only
- no client-only HUD or `world_sync` mutation for devtools effects
- `internal/game/export_devtools_*.go` exposes narrow game-owned adapters for devtools
- score/lives devtools adapters delegate to the shared player counter seam
- clear bullets/asteroids mutate authoritative server state only; clients observe changes through normal state/world sync

## Verification Commands

```bash
cd services/game-server
go test -buildvcs=false ./internal/devtools/...
go test -buildvcs=false -tags nodevtools ./internal/devtools/...
```

```bash
data-sync -check -packets -go -gds
data-sync -diff -packets -go -gds
```
