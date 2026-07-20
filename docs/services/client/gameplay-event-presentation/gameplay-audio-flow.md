---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-77a6-a3eb-9d976f729a40
document_type: general
policy_exempt: false
summary: This document describes the Godot client gameplay audio flow.
---
# Gameplay Audio Flow

Parent index: [Gameplay Event Presentation](./!INDEX.md)

## Purpose

This document describes the Godot client gameplay audio flow.

It covers local audio playback helpers, gameplay event sound routing, entity-creation sounds, game-over sound gating, pickup audio, projectile firing audio, afterburner audio, and the boundary between client presentation audio and authoritative gameplay state.

## Overview

Gameplay audio is client presentation only.

The server decides gameplay facts such as projectile existence, pickup collection, ship death, radial effect start, local elimination, and room match-over. The client converts those facts into local Godot audio playback.

The event and authoritative lifecycle paths may both request local-elimination handling, but `MatchEndFlow` accepts immediate local-elimination audio consequences only in multiplayer. Single-player waits for authoritative room `GameOver` before requesting match-over audio through the room match-over path.

The main audio helper is:

```text
client/scripts/gameplay/audio/gameplay_audio_flow.gd
```

`GameplayAudioFlow` is intentionally thin. It does not decide when gameplay happened. It receives existing `AudioStreamPlayer` or `AudioStreamPlayer2D` nodes from scenes and starts or stops playback.

Current gameplay audio enters through these paths:

```text
event_batch
-> client/scripts/protocol/realtime/event_batch_applier.gd
-> EventPresentationAdapter
-> GameplayEventLifecycleFlow
-> GameplayEventController
-> GameplayEffects
-> GameplayAudioFlow
-> effect-scene AudioStreamPlayer2D
```

```text
`world_lane_state` bullet/projectile records
-> WorldSync
-> ProjectileSync
-> GameplayAudioFlow
-> projectile FiringSound
```

```text
world_lane_state pickup records
-> WorldSync
-> PickupSync
-> pickup.gd
-> GameplayAudioFlow
-> pickup spawn sound
```

```text
local player presentation
-> Player.set_afterburner_active()
-> GameplayAudioFlow
-> afterburner AudioStreamPlayer2D
```

```text
MatchEndFlow
-> GameplayEventFlow
-> GameplayEffects
-> GameplayAudioFlow
-> HUD GameOverSound
```

Local game-over audio can be requested after either immediate event handling or authoritative lifecycle reconstruction:

```text
ship_death event with lives == 0
-> GameplayDeathFlow
-> MatchEndFlow.handle_local_player_eliminated(lives)

authoritative eliminated lifecycle with integer lives
-> GameplayLocalLifecycleFlow
-> MatchEndFlow.handle_local_player_eliminated(lives)
```

Background music is a separate scene-level `AudioStreamPlayer` flow. It starts from the root `BackgroundMusic` node in `client/scenes/game.tscn` and is not driven by server gameplay events.

## Code root

```text
client/
```

## Responsibilities

Gameplay audio owns:

* Playing gameplay sound nodes supplied by gameplay scenes.
* Stopping afterburner and game-over audio when requested.
* Detaching projectile firing sounds so short-lived projectile nodes do not cut off launch audio.
* Playing effect sounds for bullet blasts, pickup collection, ship death, and torpedo explosion.
* Playing pickup spawn sounds when pickup nodes are first created.
* Playing local afterburner sound while the local player thrusts.
* Restarting local afterburner audio while afterburner presentation remains active.
* Looking up HUD `%GameOverSound` during gameplay event flow configuration.
* Playing delayed game-over sound requested by `MatchEndFlow`.
* Participating in repeated-game-over-sound suppression owned by the match-end/effects lifecycle seam.
* Invalidating pending delayed game-over sound timers when game-over audio is reset or stopped.
* Keeping audio playback local and non-authoritative.

## Does not own

Gameplay audio does not own:

* Server gameplay authority.
* Projectile spawn authority.
* Pickup spawn, collection, expiry, or effect authority.
* Ship death authority.
* Radial effect authority.
* Match-over decisions.
* Local elimination decisions.
* Score, lives, winner, or match-result calculation.
* Packet schema source-of-truth files.
* Asset authoring.
* Audio mixing policy beyond per-scene node settings.
* HUD layout.
* World-sync entity ownership.
* Visual effect animation ownership.
* App-level route execution.
* Durable settings or player profile audio preferences.

## Domain roles

### Audio playback helper

`GameplayAudioFlow` owns the direct audio playback calls.

Most methods are null-safe wrappers around scene-provided audio nodes:

```text
play_bullet_blast_sound(sound)
play_pickup_collected_sound(sound)
play_pickup_spawned_sound(sound)
play_ship_death_sound(sound)
play_torpedo_explosion_sound(sound)
play_afterburner_sound(sound)
stop_afterburner_sound(sound)
play_game_over_sound()
stop_game_over_sound()
```

The helper does not inspect packet data and does not decide whether an effect should happen. Compact wire aliases are transport details, not audio API inputs.

### Gameplay event audio

`GameplayEventFlow` constructs `GameplayEffects` and `GameplayEventController`.

`EventPresentationAdapter` forwards applied `event_batch` output into `GameplayEventLifecycleFlow`, which then hands those events to `GameplayEventFlow`. `event_batch` reaches this flow through compact wire output, but it is expanded before gameplay audio consumes it, so audio code should work with readable long-key event dictionaries rather than compact aliases.

`GameplayEventController` reads server events, converts event server coordinates into visual coordinates, and asks `GameplayEffects` to spawn presentation effects.

Current event audio paths include:

```text
bullet_blast
-> GameplayEffects.spawn_bullet_blast()
-> AsteroidDestroyed AudioStreamPlayer2D
```

```text
ship_death
-> GameplayEffects.spawn_ship_death()
-> ShipDeath AudioStreamPlayer2D
```

`ship_death` audio is immediate best-effort event presentation. It is useful for death animation and sound timing, but it is not durable lifecycle truth and does not replace authoritative `player_lifecycle`, `player_sessions`, or world-state reconstruction.

```text
radial_effect_started
-> GameplayEffects.spawn_torpedo_explosion()
-> TorpedoExplosionSound AudioStreamPlayer2D
```

```text
pickup_collected
-> GameplayEffects.spawn_pickup_collected()
-> pickup_collect AudioStreamPlayer2D
```

`pickup_effect_applied` is currently received by the event controller but does not trigger a gameplay audio path.

### Game-over audio

`MatchEndFlow` may request game-over audio, but it does not play audio directly.

Current request paths are:

```text
ship_death event with lives == 0
-> GameplayDeathFlow
-> MatchEndFlow.handle_local_player_eliminated(lives)
-> GameplayEventFlow.play_game_over_sound_after_delay()
```

```text
authoritative eliminated lifecycle with integer lives
-> GameplayLocalLifecycleFlow
-> MatchEndFlow.handle_local_player_eliminated(lives)
-> GameplayEventFlow.play_game_over_sound_after_delay()
```

```text
room state GameOver
-> MatchEndFlow.handle_room_match_over()
-> GameplayEventFlow.play_game_over_sound_after_delay()
```

`MatchEndFlow` owns idempotent multiplayer local-elimination orchestration and suppresses duplicate delayed-sound requests across event and authoritative lifecycle entry paths. Its session-mode guard rejects immediate local-elimination consequences in single-player. `GameplayEventFlow` and `GameplayEffects` own the delay, timer invalidation, and playback mechanics after a request is accepted. Single-player receives game-over audio through the authoritative room match-over path instead.

The delay uses:

```text
Constants.GAME_OVER_SOUND_DELAY
```

The current generated value is `0.4`.

`game_over_sound_played` prevents repeated playback in the same lifecycle. `game_over_sound_token` invalidates pending delayed playback when the flow resets or stops the sound.

`GameplayAudioFlow.configure(hud)` resolves the HUD sound node from:

```text
%GameOverSound
```

That node currently lives in:

```text
client/scenes/ui/hud.tscn
```

### Projectile firing audio

Projectile firing audio is driven by first-seen projectile nodes in world sync.

Current flow:

```text
`world_lane_state` bullet/projectile records
-> WorldSync.apply_world_lane_state()
-> ProjectileSync.apply()
-> first projectile node creation
-> FiringSound
-> GameplayAudioFlow.play_projectile_firing_sound()
```

`play_projectile_firing_sound()` duplicates the projectile scene's `FiringSound`, adds the duplicate to the projectile layer, starts playback, and queues the duplicate for deletion when playback finishes.

This keeps launch audio from being cut off if the projectile node is removed quickly by later world state.

Current projectile scenes using `FiringSound`:

```text
client/scenes/bullet.tscn
client/scenes/projectiles/torpedo.tscn
```

### Pickup audio

Pickup audio has two separate paths.

Pickup spawn sound is node-creation presentation:

```text
world_lane_state pickup records
-> WorldSync
-> PickupSync
-> pickup node first creation
-> pickup.gd.play_spawn_sound(audio_flow)
-> GameplayAudioFlow.play_pickup_spawned_sound()
```

Pickup collection sound is event presentation:

```text
event_batch output
-> GameplayEventController.apply_pickup_collected()
-> GameplayEffects.spawn_pickup_collected()
-> pickup_collect scene AudioStreamPlayer2D
```

Collection sound should stay event-driven rather than pickup-node cleanup-driven because the pickup node may already be removed by world sync before collection audio finishes.

### Afterburner audio

Local afterburner audio is owned by the local player presentation path.

Current flow:

```text
LocalPlayerPresentationController.process()
-> Player.set_afterburner_active(Input.is_action_pressed(move_forward))
-> GameplayAudioFlow.play_afterburner_sound()
```

When the local afterburner sound finishes while afterburner presentation is still active, `Player._on_afterburner_audio_finished()` starts it again.

Stopping afterburner presentation calls:

```text
GameplayAudioFlow.stop_afterburner_sound()
```

`Player.stop_transient_effects()` also clears afterburner audio by setting afterburner inactive. Local self-death uses that path before HUD or death presentation is applied.

Remote afterburner visual state does not currently play remote afterburner audio.

### Background music

Background music is scene-level audio, not server-event gameplay audio.

Current scene path:

```text
client/scenes/game.tscn
-> BackgroundMusic
-> client/scripts/gameplay/audio/background_music_flow.gd
```

`BackgroundMusicFlow` starts music in `_ready()`, can stop music, and can ensure music is playing. It is not routed through `GameplayEventFlow`, `GameplayEffects`, or `GameplayAudioFlow`.

## Protocols and APIs

### Server event input

Gameplay event sounds consume applied `event_batch` output from the realtime event adapter path.

Relevant event types currently handled by `GameplayEventController`:

```text
bullet_blast
ship_death
radial_effect_started
pickup_collected
pickup_effect_applied
```

Only the first four currently produce visual or audio effects.

Event positions are converted through the configured visual-position converter before spawning effect scenes. Audio itself plays from the spawned scene node.

### World-state input

Projectile and pickup spawn sounds are driven by world-lane entity appearance.

The client does not receive explicit "play projectile sound" or "play pickup spawn sound" packets. It infers those sounds from first local creation of server-authoritative entities.

### Match-end input

Game-over audio requests originate from client match-end presentation state:

```text
local elimination
room match-over
```

Both paths are consequences of server-owned gameplay or room facts. The client only chooses how to present the sound. The local-elimination path has immediate consequences only in multiplayer; single-player waits for room `GameOver`.

The `ship_death` path is supplementary immediate presentation. The authoritative world/session lifecycle path is the reconstructable source for pending respawn, active restoration, and eliminated state; audio does not acknowledge durable lifecycle delivery.

### Outbound APIs

Gameplay audio sends no packets and exposes no HTTP APIs.

## Data ownership

Gameplay audio owns transient client presentation state only.

Current local state includes:

```text
GameplayAudioFlow.game_over_sound
GameplayEffects.game_over_sound_played
GameplayEffects.game_over_sound_token
scene-local AudioStreamPlayer and AudioStreamPlayer2D node playback state
temporary detached projectile firing sound nodes
temporary effect nodes and cleanup timers
Player.afterburner_audio playback state
BackgroundMusic playback state
```

This state is not durable.

It is not authoritative.

It is reset through gameplay, effects, reset paths, scene teardown, or normal Godot node lifecycle.

## Scene requirements

Gameplay audio depends on conventional sound node names in scenes.

Current required or expected nodes:

```text
client/scenes/ui/hud.tscn
- %GameOverSound

client/scenes/animations/bullet_blast.tscn
- AsteroidDestroyed

client/scenes/animations/ship_death.tscn
- ShipDeath

client/scenes/animations/torpedo_explosion.tscn
- TorpedoExplosionSound

client/scenes/pickups/pickup_collect.tscn
- AudioStreamPlayer2D

client/scenes/pickups/powerup_pickup.tscn
- AudioStreamPlayer2D

client/scenes/pickups/weapon_pickup.tscn
- AudioStreamPlayer2D

client/scenes/bullet.tscn
- FiringSound

client/scenes/projectiles/torpedo.tscn
- FiringSound

client/scenes/animations/blue_afterburner.tscn
- AudioStreamPlayer2D

client/scenes/game.tscn
- BackgroundMusic
```

Scene-local audio settings such as stream, volume, pitch, looping, and polyphony are configured on those scene nodes.

## Code map

### Primary audio implementation

* `client/scripts/gameplay/audio/gameplay_audio_flow.gd` - Gameplay audio helper for scene sound playback, game-over sound lookup, projectile sound detachment, and null-safe play or stop calls.
* `client/scripts/gameplay/audio/background_music_flow.gd` - Scene-level background music start, stop, or ensure helper.

### Event and effect audio callers

* `client/scripts/protocol/realtime/event_batch_applier.gd` - Applies and dedupes `event_batch` payloads before event-presentation fanout.
* `client/scripts/protocol/realtime/event_presentation_adapter.gd` - Forwards applied `event_batch` output into gameplay event presentation.
* `client/scripts/gameplay/events/gameplay_event_flow.gd` - Constructs effects and event controller, exposes game-over audio request and reset methods.
* `client/scripts/gameplay/events/gameplay_event_controller.gd` - Converts server events into effect spawn calls.
* `client/scripts/gameplay/effects/gameplay_effects.gd` - Spawns effect scenes, starts their sounds, manages effect cleanup, and executes accepted game-over sound delay or one-shot mechanics.
* `client/scripts/gameplay/events/gameplay_event_lifecycle_flow.gd` - Wires event flow into gameplay event presentation and reset.
* `client/scripts/gameplay/events/gameplay_death_flow.gd` - Stops transient player effects before local death presentation and delegates final elimination to match-end flow.
* `client/scripts/gameplay/lifecycle/gameplay_local_lifecycle_flow.gd` - Reconstructs local lifecycle presentation and delegates authoritative elimination to match-end flow.
* `client/scripts/gameplay/match_end/match_end_flow.gd` - Requests game-over audio for accepted multiplayer local elimination and authoritative room match-over.

### World-sync audio callers

* `client/scripts/world/projectile_sync.gd` - Plays projectile firing sound on first projectile creation.
* `client/scripts/world/pickup_sync.gd` - Plays pickup spawn sound on first pickup node creation.
* `client/scripts/entities/pickup.gd` - Delegates pickup spawn sound playback to gameplay audio flow.

### Player presentation audio callers

* `client/scripts/entities/player.gd` - Owns local afterburner sound playback, restart, and stop behavior.
* `client/scripts/gameplay/presentation/local_player_presentation_controller.gd` - Activates local afterburner presentation from local movement input.
* `client/scripts/shell/gameplay_menu_flow.gd` - Stops local afterburner presentation when gameplay menu state requires it.

### Scenes and assets

* `client/scenes/game.tscn`
* `client/scenes/ui/hud.tscn`
* `client/scenes/animations/bullet_blast.tscn`
* `client/scenes/animations/ship_death.tscn`
* `client/scenes/animations/torpedo_explosion.tscn`
* `client/scenes/animations/blue_afterburner.tscn`
* `client/scenes/pickups/pickup_collect.tscn`
* `client/scenes/pickups/powerup_pickup.tscn`
* `client/scenes/pickups/weapon_pickup.tscn`
* `client/scenes/bullet.tscn`
* `client/scenes/projectiles/torpedo.tscn`
* `client/audio/`

### Generated inputs

* `client/scripts/generated/constants/constants.gd` - Generated presentation constants, including game-over sound delay and effect cleanup padding.
* `client/scripts/generated/networking/packets/packets.gd` - Generated event type and packet field constants consumed by event flow.

### Source-of-truth boundaries

* `shared/constants/client/presentation.toml` - Source for generated client presentation constants.
* `shared/packets/gameplay.toml` - Source for gameplay packet and server-event fields.

### Non-owning boundaries

* `services/game-server/` - Owns gameplay authority and emits authoritative world state or events.
* `client/scripts/networking/` - Owns transport and packet routing, not sound playback.
* `client/scripts/world/` - Owns entity sync and creation timing, not sound implementation.
* `client/scripts/gameplay/match_end/` - Owns match-end presentation orchestration, not audio playback or audio gating.
* `client/scripts/ui/` and `client/scripts/shell/gameplay_hud_flow.gd` - Own HUD visibility and widgets, not audio playback policy.
* `client/scripts/ui/` - Owns UI controls and result or menu presentation, not gameplay event audio.

## Tests

Relevant tests include:

* `client/tests/unit/gameplay/effects/test_gameplay_effects.gd`
* `client/tests/unit/gameplay/match_end/test_match_end_flow.gd`
* `client/tests/unit/protocol/realtime/test_event_batch_and_resync.gd`
* `client/tests/unit/test_world_sync.gd`
* `client/tests/unit/test_pickup.gd`
* `client/tests/unit/test_pickup_sync.gd`
* `client/tests/unit/gameplay/events/test_gameplay_death_flow.gd`
* `client/tests/unit/gameplay/test_gameplay_flow_composer.gd`

Current direct coverage verifies:

* Torpedo explosion effect scene creation, scaling, and sound-node presence.
* Multiplayer match-end requests game-over sound for accepted local elimination.
* Single-player local-elimination calls do not request immediate game-over sound.
* Match-end requests game-over sound for authoritative room match-over.
* Repeated room match-over handling does not repeatedly show results.
* Bullet and torpedo projectile nodes expose firing sound nodes on first creation.
* Pickup presentation and world-sync behavior around pickup nodes and collection effects.

Use the normal client GUT verification flow when changing gameplay audio behavior.

## Related docs

* [Gameplay Event Presentation](./!INDEX.md)
* [Client](../!INDEX.md)
* [Gameplay Runtime](../gameplay-runtime/!INDEX.md)
* [Gameplay State Application](../gameplay-runtime/gameplay-state-application.md)
* [Match End Orchestration](../match-end-flow/match-end-orchestration.md)
* [Pickup Presentation](../world-sync/pickup-presentation.md)
* [HUD And Gameplay UI](../hud-and-gameplay-ui.md)
* [Presentation Flow](../presentation-flow/!INDEX.md)
* [Gameplay packets](../../../protocol/gameplay-packets.md) - gameplay realtime packet documentation.
* [Constants pipeline](../../../data/data-sync-and-ssot-pipeline.md) - generated constants documentation.

## Notes

Gameplay audio should remain a presentation adapter, not a gameplay decision point.

Do not move gameplay authority into audio paths. Audio playback should follow server facts, world-sync node creation, local presentation state, or scene lifecycle.

Projectile firing sound is intentionally detached from projectile node lifetime. Pickup collection sound is intentionally detached from pickup node lifetime by using the `pickup_collected` event effect path.

Game-over audio is requested by match-end presentation, with duplicate local-elimination requests suppressed by `MatchEndFlow` and delayed playback handled by the event/effects seam. Keep those responsibilities separate so repeated room snapshots, duplicate event delivery, or repeated local lifecycle state do not directly replay audio.