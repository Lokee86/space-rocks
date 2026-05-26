# Packet Schema Change Skill

Use this skill when changing packet schemas, generated packet files, packet builders, or packet codec boundaries.

## When to use

Use this skill for work involving:

- `shared/packets/packets.toml`
- `client/scripts/networking/packets.gd`
- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/entities/packets_generated.go`
- `services/game-server/internal/protocol/packetcodec/`
- `client/scripts/networking/packet_codec/packet_codec.gd`
- websocket packet encode/decode paths
- packet-facing lifecycle/status fields

## Source of truth

Packet source of truth:

```text
shared/packets/packets.toml
```

Generated packet outputs:

```text
client/scripts/networking/packets.gd
services/game-server/internal/game/packets.go
services/game-server/internal/game/entities/packets_generated.go
```

Do not hand-edit generated packet outputs unless the user explicitly asks for a temporary/manual intervention.

Packet pull is intentionally unsupported. Packet schema changes should be made in `shared/packets/packets.toml` and pushed with `tools/data_sync`.

## Codec rules

- Route server packet wire JSON through `services/game-server/internal/protocol/packetcodec`.
- Do not add direct `encoding/json` packet calls outside the codec path.
- Non-packet JSON, such as collision-shape data-file parsing, may still use `encoding/json` directly.
- Route client packet wire JSON through `client/scripts/networking/packet_codec/packet_codec.gd`.
- Do not add direct `JSON.stringify` or `JSON.parse_string` calls in websocket packet paths.
- The codecs are intentionally JSON-only and thin/generic.
- Do not add validation, format switching, typed packet objects, protobuf references, or generator changes unless explicitly requested.
- `network_client.gd` still owns websocket behavior.

## Lifecycle packet rules

- Keep packet-facing player lifecycle status in `StatePacket.player_lifecycle`, beside `players`.
- Do not put match lifecycle on `ShipState`; pending-respawn and eliminated players may not have active ship state.
- Client spectate/view-cycle eligibility must use authoritative lifecycle status (`active`) plus visual availability.
- Do not infer active eligibility solely from remote player positions or ship presence.

## Workflow

1. Edit `shared/packets/packets.toml` when the schema changes.
2. Do not hand-edit generated packet files unless explicitly requested.
3. Keep generated Go/GDS outputs together when applying generation.
4. Update packet codec or call sites only if required.
5. Add/update focused Go or GUT tests only when the prompt asks for test edits.
6. Leave broad validation and generated diff checks for human-run checkpoint commands unless the prompt explicitly allows terminal commands.

## Human-run commands

Validate shared packets:

```bash
python3 tools/data_sync/main.py -validate -packets
```

Preview shared packet generated output:

```bash
python3 tools/data_sync/main.py -diff -packets -go -gds
```

Apply shared packet generated output:

```bash
python3 tools/data_sync/main.py -push -packets -go -gds
```

Run server tests:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

Run client GUT tests when Godot CLI is available:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

Do not run these commands by default as the agent.

## Stop conditions

Stop and report instead of continuing if:

- Generated diffs include unexpected unrelated churn.
- Packet changes require behavior changes in multiple unrelated systems.
- A change would add packet lifecycle policy to `ShipState`.
- A change would bypass either packet codec boundary.
