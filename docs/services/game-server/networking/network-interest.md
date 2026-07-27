---
author: brian
created: "2026-07-27"
document_id: 82eaa581-34a5-4ac3-93bf-fca9ecbc819f
document_type: architecture
policy_exempt: false
summary: Recipient-specific realtime interest filtering and coarse player locator behavior.
---
# Network Interest

Parent index: [Game Server Networking](./!INDEX.md)

The game server filters realtime world presentation separately for each receiving player. Simulation authority and entity existence remain global; interest only controls what presentation state is sent to one client.

## Camera Region

Interest uses the shared wrap-aware camera-region functions under `internal/game/visibility`. Spawning and networking consume the same toroidal region math.

The receiver camera is the default anchor. During spectating, the session-selected view target becomes the anchor when that player is active.

Entry and exit margins differ to prevent entities repeatedly entering and leaving interest near the camera boundary. Projectiles use a larger margin than other entities.

## Entity Lifecycle

Interest exit is represented through the existing recipient world projection and lifecycle packet flow. It does not destroy the authoritative entity.

- Entering interest produces a lifecycle create/full record for that recipient.
- Remaining in interest produces ordinary hot updates.
- Leaving interest produces a lifecycle delete for that recipient.
- Re-entering interest produces another lifecycle create/full record.

The receiver's own ship and the active spectate target remain relevant regardless of distance.

## Player Locator

Far players stop receiving full high-frequency ship state, but every client receives a coarse `player_locator` packet at approximately 5 Hz. The packet contains player ID, position, velocity, and active state.

`player_locator` is a separate packet family sent through the existing unordered, unreliable `sr.ships` DataChannel. It does not create another physical channel and does not share the ship-delta sequence/projection state.

The client extrapolates locator positions for at most 0.75 seconds and discards locator-only indicators after two seconds without a fresh packet. Rendered nearby ship positions take precedence over locator positions.

## Spectating

The client sends `set_view_target_request` and `clear_view_target_request` over the control connection. The server stores this selection on the WebSocket session and applies it when constructing recipient-specific presentation snapshots.

The client retries focusing the camera until the selected target's full ship presentation enters interest.
