# Current System Limits
Parent index: [Limits](!README.md)

## Architecture / Networking

- No prediction/reconciliation layer is implemented beyond interpolation.
- The server is expected to be running separately for the Godot client.
- Local server launch from the Godot client is not implemented.
- The current client expects a local Go server target during development.
- Vertical despawn behavior is limited by the relationship between world height, visible viewport height, and despawn margin.

## Combat Systems

- The client should not calculate damage locally.
- Client rendering from damage events is not fully implemented in the damage design path.
- `damage_over_time_started` and `damage_over_time_tick` have adapters/mapping, but active DoT gameplay ownership is not fully wired here unless the code says otherwise.
- Presentation concepts such as `shield_absorbed`, `damage_immune`, and `damage_area_applied` are not implemented unless wired elsewhere.
- Radial client visuals are not fully implemented in the radial design path.
- Torpedo radial currently targets asteroids and enemies only; players, projectiles, and pickups are excluded.
- Radial knockback is not implemented.
- Radial status effects are not implemented.
- Enemy death consequences are not fully wired yet.
- Only `basicasteroids` drop tables exist today.
- There is no minimum drop count policy yet.
- All current asteroid variants use `collision_shape = "asteroid:0"`.
- All current asteroid variants use `stats_profile = "standard"`.
- All current asteroid variants use `drop_table = "basicasteroids"`.
- Pickup health is current health only.
- Pickups have no `max_health` field.
- Bullet/pickup collision damage is not enabled.

## Player Data

- Matchmaking and leaderboards are not implemented.
- The Rails API/auth path exists, but broader account product surfaces and durable progression systems are incomplete.
- Player-data schema generation is not fully implemented as a separate pipeline domain.
- Generated migration skeletons are not implemented.
- Player-data contract tests for schema drift enforcement are not implemented.
- Live progression grants are not implemented.
- Currency, ship parts, unlocks, achievements, and loadout persistence are not implemented.
- V1 stats payloads do not include currency, ship parts, unlocks, loadouts, achievements, or match history yet.

## Client Presentation

- See [Player Build Limits](player-build-limits.md) for current ship-variant and player-build constraints.
- Weapon UI and equip presentation are not fully implemented yet.

### Client Menu Flow

- Options is not implemented.
- Campaign is disabled in the single-player pregame menu.
- Loadout is disabled in the single-player pregame menu.
- Provisioner is disabled in the single-player pregame menu.
- Buy Scrap is disabled in the single-player pregame menu.
- Rankings are disabled in the single-player pregame menu.
- Manual login is disabled.
- Google login is disabled.
