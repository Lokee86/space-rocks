# Platform And Progression Roadmap

## Opening Context

The menu/profile/local-pilot/match-results/stats-refresh vertical slice is complete and green. That is an important milestone, but it also marks a shift in how the project needs to be planned.

We have reached the point where isolated feature slices are no longer enough on their own. Future work will overlap more across auth, observability, packet strategy, progression, enemies, bullet hell, ship variants, unlocks, and account surfaces.

Because those systems now intersect more tightly, seams and ownership boundaries need to be planned before more feature growth lands. The goal is to keep the codebase scalable and avoid coupling the next wave of systems into the wrong layers.

## Roadmap Purpose

This document coordinates the larger phases that follow the completed menu-flow vertical slice. It frames the work that needs to happen next, highlights the dependencies between the major areas, and sets up the order in which the platform and progression systems should be planned.

## Major Planning Pillars

- Network budget and runtime observability
- Realtime protocol architecture and state delivery strategy
- Auth and account surface completion
- Player experience structure and system boundaries
- Progression, rewards, unlocks, and player-data contracts
- Gameplay expansion

## Current Completed Baseline

- Discord auth works through the Godot browser handoff.
- Rails owns bearer token issuance and `/api/auth/me`.
- The Go game-server verifies tokens through Rails for websocket auth.
- Local Profile and Authenticated Account stats route through the player-data runtime.
- Match results and stats refresh are complete.
- The World Telemetry Overlay exists behind devtools.
- Gameplay packets are known to exceed 4KB at times before enemies or bullet hell exist.

## Phase Order

- Phase A - Network Budget + Runtime Observability Foundation
- Phase B - Realtime Protocol Architecture
- Phase C - Player Experience Foundation
- Phase D - Progression Foundation
- Phase E - Gameplay Expansion

Phase A determines the priority and order of Phase B work, not whether launch-grade realtime protocol work is expected.

## Phase A

Phase A answers whether the current architecture can safely support more entities, enemies, bullet hell patterns, progression events, and online play without flying blind. Phase A is measurement and diagnostics, not optimization. Phase A should make later optimization choices evidence-based.

### Existing Baseline

- `services/game-server/internal/networking/outbound/gameplay_state_metrics.go` already warns on gameplay presentation packets over 4KB.
- The same outbound path warns on gameplay presentation writes slower than 20ms.
- `services/game-server/internal/networking/outbound/` owns outbound gameplay presentation helpers.
- `services/game-server/internal/protocol/packetcodec/` owns JSON packet encode/decode.
- `client/scripts/devtools/telemetry/` owns client-side telemetry models.
- `client/scenes/devtools/world_telemetry_overlay.tscn` is the devtools-only overlay.
- `docs/server/logging.md` and `docs/client/logging.md` already define logging rules.

### Ownership Rules

- Server networking owns encoded outbound packet size, write duration, packet type/category, room ID, player ID, and presentation-state diagnostics.
- Server gameplay owns authoritative state and entity counts before serialization.
- Server logging owns threshold warnings and structured fields.
- Client devtools owns packet/network metrics display.
- Client HUD does not own packet observability.
- Documentation owns packet-budget policy.

### Goals

- Define an initial gameplay packet budget.
- Measure outbound gameplay packet byte size.
- Identify contributor counts for large gameplay packets.
- Surface packet byte pressure in devtools telemetry.
- Keep observability separate from gameplay behavior.
- Preserve JSON encoding until measurements identify the bottleneck.
- Provide evidence for later packet strategy work.

### Non-Goals

- No packet compression.
- No binary protocol migration.
- No delta-state protocol.
- No gameplay packet lane split.
- No enemies.
- No bullet hell mechanics.
- No progression rewards or live grants.
- No auth expansion.
- No website work.
- No player-facing telemetry.
- No raw full-payload packet dumps by default.
- No gameplay behavior changes.

### Initial Packet Budget

| Threshold | Policy |
| --- | --- |
| Gameplay packet warning: 4KB | Structured warning with contributor counts. |
| Gameplay packet danger: 8KB | Treat as a blocker before entity-heavy feature growth. |
| Slow gameplay write: 20ms | Structured warning with packet size and route context. |
| Target steady-state gameplay packet: under 4KB | Preferred normal gameplay state. |

These thresholds are provisional. Phase A measures whether they are realistic.

### Required Large-Packet Diagnostics

- Encoded byte size
- Packet type
- Room ID
- Player ID
- Remote address if already available in the outbound path
- Room state
- Players count
- `player_sessions` count
- `player_lifecycle` count
- Asteroid count
- Bullet count
- Pickup count
- Enemy count
- Event count
- Total spawned asteroid count
- Build duration where cheap and localized
- Encode duration where cheap and localized
- Write duration where cheap and localized

Raw packet payloads should not be logged by default.

### Devtools Display Requirements

- The World Telemetry Overlay should show latest gameplay packet bytes.
- The World Telemetry Overlay should show max gameplay packet bytes.
- The World Telemetry Overlay may show optional average gameplay packet bytes.
- The World Telemetry Overlay should show large packet warning count if cheap to track.
- Existing entity counts and timing values should remain.
- This remains devtools-only and must not affect gameplay.

### Likely Phase A Workstreams

1. Document packet budget policy.
2. Extend server contributor metrics in the outbound gameplay presentation path.
3. Measure client-side inbound raw message byte length by packet type.
4. Surface packet byte metrics in the world telemetry overlay.
5. Update telemetry/logging docs and add a smoke checklist.

### Phase A Completion Criteria

- Packet budget policy is documented.
- Large gameplay packets include contributor-count diagnostics.
- Slow writes include useful context.
- Devtools overlay exposes packet byte pressure.
- Manual smoke can demonstrate packet size changes as bullets and asteroids increase.
- No packet format has changed.
- No gameplay behavior has changed.
- No feature work is mixed in.

### Post-Phase-A Decision Gate

Phase A does not automatically choose the Phase B emphasis. Phase A exists to decide what the next major route inside Phase B should be. Network optimization and related protocol work is the most likely next route if Phase A confirms current packet pressure, because gameplay packets are already known to exceed 4KB at times before enemies or bullet hell mechanics exist.

Route 1 - Network optimization immediately

- Choose this if normal gameplay packets are often over 4KB.
- Choose this if packets spike toward or past 8KB.
- Choose this if packet size grows predictably with bullets, asteroids, pickups, or players.
- Choose this if write times or jitter correlate with packet size.
- Choose this if entity-heavy features would clearly make packet pressure worse.
- This is the likely route if Phase A confirms the current concern.

Route 2 - More observability hardening before optimization

- Choose this if packet size is measured but contributors are unclear.
- Choose this if client overlay and server logs disagree.
- Choose this if slow writes happen without large packets.
- Choose this if packet size is acceptable but tick, build, or write timing is not.
- Choose this if instrumentation is too noisy or incomplete to justify a protocol change.

Route 3 - Move to auth and account identity planning before optimization

- Choose this only if normal gameplay packets stay below 4KB.
- Choose this only if spikes are rare and explainable.
- Choose this only if write timing and jitter show no packet-size pressure.
- Choose this only if packet size is not blocking enemies, bullet hell, or progression soon.

Likely optimization families under Route 1, without choosing one:

- Compact wire shape or generated short field names, if JSON key overhead dominates.
- Delta snapshots, if repeated full entity state dominates.
- Fast/slow packet lane split, if all data is being sent at the same frequency.
- Event queue trimming or acknowledgement, if events accumulate or repeat too long.
- Debug lane separation, if debug or devtools data leaks into normal gameplay packets.
- Shared room snapshot plus per-client overlay, if most state is duplicated per client but only small portions are player-specific.

The next planning work after Phase A should be selected by evidence from the decision gate, not by feature visibility alone.

## Phase B - Realtime Protocol Architecture

Phase B establishes the end-state realtime protocol seam for authoritative multiplayer state delivery. Phase B replaces the current full-state-per-tick model with a governed realtime protocol boundary.

The central rule is:

- `networking/outbound` owns delivery mechanics.
- `protocol/realtime` owns delivery policy.
- `protocol/packetcodec` owns byte representation.

Server placement:

- `services/game-server/internal/protocol/realtime/`
- `services/game-server/internal/protocol/packetcodec/`
- Keep `services/game-server/internal/networking/outbound/` as delivery mechanics.

Client placement:

- `client/scripts/protocol/realtime/`
- `client/scripts/protocol/packetcodec/`
- Current client packet codec files should move from `client/scripts/networking/packets/` into `client/scripts/protocol/packetcodec/` during Phase B.

Rails/API has no realtime gameplay connection. Rails/API remains auth, account/profile, website/API, and durable player-data persistence. Realtime snapshots, deltas, lanes, replication, and transport stay between the game-server and Godot client.

### Protocol Vocabulary

- Full snapshot
- Delta snapshot
- Baseline ID
- Sequence number
- Create/update/delete records
- Lane
- Priority
- Reliability class
- Resync request
- Forced resync
- Stale update discard

### Lanes And Delivery Policy

Lanes are protocol concepts, not an immediate transport commitment.

- Reliable control lane
- Realtime state lane
- Event lane
- Slow world lane
- Debug/telemetry lane

Transport architecture is a game-time decision. The protocol lanes should be able to map later to WebSocket, WebRTC DataChannel, UDP, or a hybrid, but this phase does not need to choose transport immediately.

### Snapshot And Delta Model

- Full snapshot on join, start, or resync
- Delta snapshots after baseline
- Per-session baseline tracking
- Monotonically increasing sequence numbers
- Entity create/update/delete records
- Missing-baseline recovery
- Stale update discard
- Explicit resync path

### State Projection And Priority Policy

- Critical
- High
- Medium
- Low
- Debug

`protocol/realtime` decides what should be sent and how often, including local player state, active player session state, nearby threats, bullets and projectiles, asteroids, pickups, future enemies, future bullet hell entities, and debug-only data.

### Outbound Collaboration

- `networking/outbound` owns delivery cadence, active sessions and channels, write calls, write failures, and backpressure.
- `protocol/realtime` owns what messages are due, what state they contain, lane assignment, priority, full and delta snapshot rules, sequence and baseline rules, and resync behavior.
- `protocol/packetcodec` owns encode and decode representation only.

Flow:

1. Authoritative game state
2. `protocol/realtime` projection
3. Priority and lane policy
4. Full snapshot or delta snapshot
5. Quantized protocol message shape
6. `protocol/packetcodec` encoding
7. `networking/outbound` delivery mechanics
8. Transport channel
9. Client `protocol/realtime` application
10. World/gameplay presentation

### Quantization And Bit Packing

Quantization and bit packing come before protobuf in Phase B.

Likely candidates:

- Positions as quantized integers
- Velocities as quantized integers
- Rotation as a fixed-range integer
- Enum strings as numeric IDs
- Booleans as flags
- Runtime entity IDs as compact IDs
- Omitted default values

Quantization rules belong to the realtime protocol schema. Encoding mechanics belong to packetcodec. Gameplay continues using normal semantic values.

### Protobuf Target

Protobuf is the final step of Phase B after lanes, snapshots and deltas, priority policy, and quantization rules are defined.

Protobuf should encode the new realtime protocol model, not the old full-state packet.

### Phase B Completion Criteria

- Realtime protocol boundary exists on server and client.
- Client codec files are moved under `client/scripts/protocol/packetcodec/`.
- Client realtime protocol code lives under `client/scripts/protocol/realtime/`.
- Server realtime protocol code lives under `services/game-server/internal/protocol/realtime/`.
- Outbound delivery mechanics and realtime protocol delivery policy are separate.
- Full snapshot and delta snapshot semantics exist.
- Per-session baseline and sequence tracking exists.
- Lane and reliability classes are defined.
- High-frequency state is no longer tied to one full reliable packet every tick.
- Quantization and bit-packing rules are defined before protobuf.
- Protobuf target is staged as the final Phase B encoding step.
- Rails/API remains outside the realtime connection path.

## Phase C - Player Experience Foundation

Phase C plans the player-facing game structure before progression persistence. Phase C starts with preset-driven room mode configuration.

### Purpose

Phase C defines the player-facing game structure: what kind of room or match the player creates, what rules govern that match, what options are configurable, what systems are affected by those rules, and what later progression must consume.

### Step 1 - Preset-Driven Room Mode Foundation

Step 1 is an implementation plan, not pure foundation. The foundation must be proven through two real baseline modes.

- `ModePreset` = named preset or template for a room or match ruleset
- `RoomModeConfig` = concrete options selected when creating a room
- `ResolvedMatchRules` = server-validated rules consumed by gameplay

Mode is not the same thing as single-player or multiplayer. Single-player and multiplayer are session or hosting context, while mode governs the match rules.

### Preset-Driven Mode Model

- `ModePreset` is the named ruleset template selected by the player.
- `RoomModeConfig` is the selected preset plus allowed player-configurable options.
- `ResolvedMatchRules` is the authoritative server-resolved rules used by the match.

Players configure room options through presets, not arbitrary free-form toggles. Presets define policy-heavy groups.

Preset-owned policy groups:

- Scoring policy
- Match-end policy
- Objective policy
- Spawn profile
- Damage policy
- Team policy
- Progression eligibility
- Result policy
- Difficulty/scaling profile

Likely player-configurable option groups:

- Lives, within preset limits
- Target score, when the preset supports it
- Time limit, when the preset supports it
- Max players, when the preset supports it
- Difficulty tier, when the preset supports it
- Hazards and pickups toggles only when the preset exposes them

Implementation flow:

1. Client selects `ModePreset`
2. Client configures exposed options
3. Client sends requested `RoomModeConfig`
4. Server validates preset and options
5. Room stores validated `RoomModeConfig`
6. Room locks mode config when match starts
7. Rules resolve config into `ResolvedMatchRules`
8. Game simulation consumes `ResolvedMatchRules`
9. Match result includes mode-aware result data

### Two-Mode Baseline

One mode only proves naming, while two modes prove the ruleset seam can vary behavior.

Baseline mode 1: `survival_arcade`

- Describes the current play behavior made explicit.
- `scoring_policy`: current asteroid scoring
- `match_end_policy`: all players eliminated
- `objective_policy`: survive / score freely
- `spawn_policy`: current asteroid spawning
- `lives_policy`: configured lives
- `result_policy`: score + deaths
- Configurable option: lives 1-5

Baseline mode 2: `score_attack`

- Describes the proof mode that uses currently available systems.
- `scoring_policy`: current asteroid scoring
- `match_end_policy`: score target reached OR all players eliminated
- `objective_policy`: reach score target
- `spawn_policy`: current asteroid spawning
- `lives_policy`: configured lives
- `result_policy`: won/lost + score + deaths + target_score
- Configurable options: lives 1-5 and `target_score` from preset-approved values

Score Attack is preferred because it uses existing score, asteroid destruction, lives and death, match-over evaluation, match results, and room lifecycle.

Score Attack does not require enemies, waves, bosses, teams, PvP damage, new pickups, campaign state, progression grants, or new objective entities.

Mission support is preparatory and can be implemented before campaign, while campaign itself remains a late future wrapper over missions.

### Affected Systems

Shared contracts / SSoT:

- Mode preset IDs and option vocabularies become shared client/server language.
- Likely fields include `preset_id`, `lives`, `target_score`, mode summary, and mode identity in match results.

Client room creation / pregame:

- Presents presets.
- Presents allowed options.
- Sends requested `RoomModeConfig`.
- Replaces hardcoded Play Endless behavior with the selected mode config path.

Rooms:

- Store validated `RoomModeConfig`.
- Lock mode config when match starts.
- Expose selected mode summary in room snapshot if needed.
- Pass config into match start.
- Rooms do not define what the mode means.

Game rules / modes:

- Define preset registry.
- Validate config.
- Construct `ResolvedMatchRules`.
- Select match-end, scoring, objective, respawn and lives, damage, team, spawn, progression, and result policies.
- Likely starts near `services/game-server/internal/game/rules`, with exact package split as a gametime decision.

Game simulation / player lifecycle:

- Consumes resolved lives count.
- Consumes match-over and objective rules.
- Should not parse raw room config throughout simulation.

Scoring:

- Reuses current asteroid scoring for both baseline modes.
- Score Attack reads current score as objective progress.
- Scoring package remains policy-focused.

Spawning:

- Both baseline modes use current asteroid spawning.
- Spawn profile support is reserved for later mode presets.

Damage / targeting / collision:

- Current baseline keeps existing damage rules.
- PvP and team damage policy are future affected behavior.

Teams:

- Not implemented in Step 1.
- Must be treated as an affected future system.
- Mode policy must leave room for none, free-for-all, co-op, fixed teams, friendly fire, team spawn rules, team result summaries, and team scoring later.

Match Results:

- Result payload should include mode identity.
- Score Attack should carry `target_score` and success/failure.
- Visible UI can remain small at first.

Player-data / progression:

- Not implemented in Step 1.
- Future progression needs trusted mode-aware results.

Client lobby/session state:

- Room snapshots may expose selected preset, option summary, mode locked state, and display name.

Devtools:

- Future diagnostics should inspect `preset_id`, resolved rules summary, objective state, match-end condition, spawn profile, and scoring policy.

### Step 1 Completion Criteria

- `survival_arcade` exists as an explicit preset.
- `score_attack` exists as a second explicit preset.
- `CreateRoomRequest` or an equivalent room creation path can carry selected mode config.
- Server validates requested preset and options.
- Room stores validated `RoomModeConfig`.
- Room snapshot exposes selected mode summary if needed by client or lobby.
- Match start resolves `ResolvedMatchRules`.
- Configured lives affect both modes.
- `target_score` affects only Score Attack.
- Survival Arcade ends on elimination.
- Score Attack ends on target score or elimination.
- Match result includes mode identity.
- Score Attack result includes `target_score` and success or failure.
- Existing current play flow still works through Survival Arcade.
- Current multiplayer create and start flow is not broken by mode config.

### Step 2 - Player Build And Loadout Foundation

Step 2 is the implementation plan for the player build and readiness package that enters a match.

Core build model:

- `ShipVariant + LoadoutSelection -> ResolvedPlayerBuild -> runtime ship/session setup`

Definitions:

- `ShipVariant` = base chassis type
- `LoadoutSelection` = selected pre-match ship/weapons/modules package
- `ResolvedPlayerBuild` = server-validated effective build used by runtime gameplay

This step includes:

- Ship variants
- Weapon hardpoints and softpoints
- Module slots
- Full shield support
- Starting ammunition
- Server-side loadout validation
- Cleanup of stale ship-side weapon fields

Cosmetics are excluded from this step.

### Ship And Weapon Ownership Cleanup

Current `ShipStats` still contains stale weapon-like fields that conflict with the newer weapons seam.

Stale ship-side fields:

- `BulletCooldown`
- `BulletDamage`
- `BulletSpeed`
- `BulletLifetime`
- `BulletSpawnOffset`

Weapon profiles already own:

- cooldown
- projectile speed
- projectile lifetime
- projectile spawn offset
- damage
- impact effects
- projectile metadata

Intended `ShipStats` shape:

- `RotationSpeed`
- `ThrustForce`
- `MaxSpeed`
- `Damping`
- `MaxHealth`
- `MaxShield`
- `CollisionShapeID`

Ownership rule:

- `ShipVariant` owns chassis, handling, survivability, collision, and slot capability.
- `WeaponProfile` owns firing, projectile, damage, ammo, and impact behavior.
- `LoadoutSelection` chooses equipment.
- `ResolvedPlayerBuild` validates and combines them.

### Weapon Points, Hardpoints, And Softpoints

The weapon point model caps each ship at four weapon points:

- `primary_1`
- `primary_2`
- `secondary_1`
- `secondary_2`

Mount kinds:

- `hardpoint` = can be equipped before match
- `softpoint` = cannot be equipped before match; pickup-capable during run
- `none` = ship does not have that weapon point

Invariant:

- Every valid ship has `primary_1` as a hardpoint.
- `primary_1` cannot be empty at match start.

`ShipVariant` defines which weapon points exist and whether they are hardpoints, softpoints, or unavailable.

`LoadoutSelection` fills hardpoints only.

Softpoints are runtime capacity for weapon pickups, not pre-match equipment slots.

### Weapon Profile Classification

Weapon profiles need classification fields for loadout validation and pickup compatibility.

Classification fields:

- `slot`
- `size`
- `delivery_class`
- `targeting_policy`
- `effect_flags`
- ammo policy support

Sizes:

- `light`
- `standard`
- `heavy`

Delivery classes:

- `ballistic`
- `missile`
- `beam`
- `mine`
- `drone`
- `self`

Targeting policies:

- `skill_shot`
- `auto_target`
- `target_lock`
- `self_target`
- `area_placed`

Effect flags:

- `direct`
- `area`
- `radial`
- `over_time`

Effect flags are composable, and at least one must be present.

Examples:

- `BasicCannon`: standard, ballistic, skill_shot, direct
- `Torpedo`: heavy, missile, target_lock or skill_shot as a gametime decision, direct + radial

Ammo policy must be loaded with the weapon selection/profile because infinite ammo for permanently equipped weapons may become a balance issue.

### Loadout Ammunition And Runtime State

`LoadoutWeaponSelection`:

- `weapon_id`
- `ammo_policy`
- `starting_ammo`

Loadouts track starting ammunition, not live ammunition.

Runtime weapon state owns:

- current ammo
- cooldown remaining
- temporary pickup changes
- temporary overwrite state

Infinite ammo should not be assumed for permanently equipped weapons because it may become a major balance issue.

`LoadoutSelection` is the pre-match starting package.

`RuntimeEquipmentState` is the mutable match state.

Runtime pickup changes must not mutate the saved loadout.

### Ship Variants And Shield Support

Ship variants define:

- `ship_type`
- display name
- movement stats
- `MaxHealth`
- `MaxShield`
- `CollisionShapeID`
- weapon point layout
- module slot availability
- default loadout

Baseline ship variants:

- `v_wing` = balanced baseline/current default
- `heavy` = slower, more health, more shield, different slot layout

Full shield support requirements:

- `ShipStats.MaxShield`
- runtime ship initializes shields from resolved build
- respawn restores shields from resolved build
- `ShipState` carries health and shields
- HUD/readout displays shields
- live combat consumes shield before health
- tests prove shield behavior through gameplay path, not only damage resolver unit tests

Optional shield regen may be added only if selected during implementation; the required baseline is max shield, shield damage absorption, state visibility, and respawn restore.

### Module Slots And Module Profiles

The system supports all four module slot types:

- `shield_mod`
- `armor_mod`
- `engine_mod`
- `utility_mod`

Ship variants define which module slots are available or restricted.

Provisional `ModuleProfile` shape:

- `module_id`
- `module_slot`
- `effect_category`
- `stat_modifiers`
- `penalties`
- `activation_policy`, if active module

Likely module meanings:

- `shield_mod`: shield amount or shield behavior
- `armor_mod`: health, resistance, or collision durability
- `engine_mod`: movement or handling
- `utility_mod`: pickup, ammo, targeting, cooldown, or special behavior

This shape is expected to evolve, but it is sufficient for the Step 2 seam.

### Loadout Validation And ResolvedPlayerBuild

The server validates the requested build before match start.

Validation inputs:

- selected ship
- selected hardpoint weapons
- selected modules
- starting ammo
- mode rules
- availability/unlocks later

Validation checks:

- ship exists
- weapon exists
- module exists
- weapon point exists on ship
- weapon point is a hardpoint for pre-match equipment
- weapon classification fits the slot rules
- module fits module slot
- `primary_1` is present
- starting ammo is valid for ammo policy
- mode allows the selected build

`ResolvedPlayerBuild` output:

- ship type
- resolved ship stats
- weapon point layout
- equipped weapons
- starting weapon state
- module effects
- max health
- max shield
- collision shape ID

Runtime ship/session setup consumes `ResolvedPlayerBuild`.

Gameplay should not parse raw loadout selection throughout simulation.

### Pickup Interaction

Weapon pickups mutate runtime weapon state, not saved loadout.

Rules:

- Same-weapon pickup increases ammunition.
- Dedicated ammunition pickup increases ammo but does not grant a weapon.
- Weapon pickup fills or overwrites an empty compatible softpoint or hardpoint first.
- If no empty compatible point exists, the pickup temporarily overwrites a filled weapon point.
- Softpoint pickup weapons persist for the run.
- Hardpoint overwrites are temporary for the run.
- Utility mods may modify pickup or ammo behavior later.

### Owned Ship And Hardwired Module Boundary

Hardwiring modules is not part of loadout.

Hardwiring belongs to a future equipment/inventory/progression layer.

`OwnedShip` may later have ship instance ID, base ship variant, and hardwired modules.

Loadout may later select an owned ship.

Loadout does not install or remove hardwired modules.

`ResolvedPlayerBuild` applies hardwired effects when selected owned ships exist.

### Step 2 Baseline Proof

Step 2 must prove two build paths:

- `v_wing` build
- `heavy` build

`v_wing` includes:

- `primary_1` hardpoint
- `basic_cannon`
- baseline health and shield
- current default behavior preserved

`heavy` includes:

- different chassis stats
- different health and shield
- different weapon point layout
- weapon behavior still comes from weapon profiles

Pickup behavior must still prove:

- torpedo pickup fills a compatible runtime weapon point
- same torpedo pickup increases ammo
- saved loadout does not change

### Step 2 Completion Criteria

- `ShipStats` no longer owns weapon projectile tuning.
- Weapon profiles own firing, projectile, damage, ammo, and impact behavior.
- Weapon points support `primary_1`, `primary_2`, `secondary_1`, `secondary_2`.
- Weapon points support `hardpoint`, `softpoint`, and `none`.
- Every valid ship has `primary_1` hardpoint.
- Loadout tracks starting ammo, not live ammo.
- Ship variants define chassis stats, shields, and slot layout.
- `v_wing` and `heavy` exist as real selectable build paths.
- `ResolvedPlayerBuild` exists and drives runtime ship setup.
- Shields initialize, absorb damage, appear in state, display in HUD/readout, and restore on respawn.
- Module slots and `ModuleProfile` foundation exist.
- Server validates selected ship, weapons, modules, and starting ammunition.
- Pickups mutate runtime weapon state without changing saved loadout.
- Current default flow still works.
- Existing torpedo pickup behavior still works through the new model.

### Gametime Decisions

- Exact package split between `game/rules`, `game/rules/modes`, or `game/modes`.
- Exact shared data format for presets.
- Whether first client UI is a full selector or a minimal preset path.
- Exact room snapshot mode-summary shape.
- Exact team policy fields.
- Exact future mission option shape.

Modes are preset-driven room and match configurations. Rooms store selected config. Rules resolve config into authoritative match policy. Gameplay consumes resolved rules. The baseline implementation must prove the seam with two modes.
