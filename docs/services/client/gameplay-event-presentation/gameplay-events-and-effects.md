# Gameplay Events And Effects

Parent index: [Gameplay Event Presentation](./!INDEX.md)

## Purpose

This document describes the client gameplay event and visual-effects presentation flow.

It covers how server presentation events enter the Godot client, how event positions are converted into visual coordinates, how short-lived effect scenes are spawned and cleaned up, and how self-death events hand off to local death and match-end presentation.

## Overview

Gameplay events are authoritative server facts carried inside gameplay state packets.

The client does not decide that a bullet blast, ship death, pickup collection, radial effect, damage result, pickup expiry, score award, death, respawn, or match end happened. The client only reads event facts emitted by the server and converts the supported subset into local presentation.

The active client path is:

```text
StatePacket.events
-> GameplayStatePacketReader
-> GameplayStateApplyFlow
-> GameplayEventLifecycleFlow
-> GameplayEventFlow
-> GameplayEventController
-> GameplayEffects
-> effect scene / audio / local death handoff
```

Server event coordinates are server-space positions. Before a visual effect is spawned, `GameplayEventController` converts those coordinates through the world-sync visual-coordinate seam:

```text
server event x/y
-> WorldSync.visual_position_for_server_position(...)
-> visual effect global_position
```

This keeps event effects visually continuous across toroidal wrap boundaries and aligned with the active ViewAnchor.

The current event/effects path presents:

```text
bullet_blast            -> bullet blast animation and sound
ship_death              -> ship death animation, sound, and local self-death handoff when player_id == self_id
radial_effect_started   -> torpedo explosion animation and sound
pickup_collected        -> pickup collection particles and sound
pickup_effect_applied   -> received, but no visual effect is currently spawned
```

Other server event names may exist in the server event vocabulary, but this client flow only presents the events explicitly routed by `GameplayEventController`.

## Code root

```text
client/
```

## Responsibilities

The gameplay event/effects presentation flow owns:

* Receiving normalized `server_events` from gameplay state application.
* Routing supported server event types to local presentation handlers.
* Emitting the local `self_death_event` signal when a `ship_death` event belongs to the local player.
* Converting server-space event coordinates into visual coordinates before spawning effects.
* Spawning short-lived visual effect scenes for supported events.
* Starting effect animations.
* Starting event-local audio through the gameplay audio flow.
* Cleaning up effect nodes after animation and/or sound completion.
* Delegating non-final local self-death to HUD dead/respawn presentation.
* Delegating final local elimination to match-end orchestration.
* Resetting game-over sound one-shot state when the gameplay lifecycle resets.
* Keeping presentation events separate from server simulation authority.

## Does not own

The gameplay event/effects presentation flow does not own:

* Server event authority.
* Collision decisions.
* Damage resolution.
* Pickup collection validity.
* Pickup effect application.
* Pickup expiry.
* Score, lives, or respawn authority.
* Match-over decisions.
* Match-result creation or persistence.
* Room lifecycle.
* Packet schema source-of-truth files.
* WebSocket transport.
* Raw packet decoding.
* Persistent world entity sync.
* Pickup node lifecycle while a pickup exists in `StatePacket.pickups`.
* HUD widget implementation.
* Gameplay menu implementation.
* Result-window presentation.
* Durable player data.
* Devtools-only event behavior.

## Domain roles

### Event lifecycle flow

`GameplayEventLifecycleFlow` wires event presentation into the gameplay runtime.

It receives the gameplay owner node, HUD, HUD flow, player node, visual coordinate converter, and optional match-end flow. It creates or accepts an event flow and death flow, configures them, connects local self-death handling, and exposes a narrow `apply_server_events(state)` method for state application.

Current lifecycle path:

```text
GameplayFlowComposer.configure(...)
-> GameplayEventLifecycleFlow.configure(...)
-> GameplayEventFlow.configure(...)
-> GameplayEventController.configure(...)
-> GameplayEffects.configure(...)
```

### Event flow

`GameplayEventFlow` owns the public event-presentation API used by runtime callers.

It configures `GameplayEffects`, configures `GameplayEventController`, forwards server event arrays, exposes game-over sound requests to match-end orchestration, and resets game-over sound state on lifecycle reset.

It emits:

```text
self_death_event(event)
```

when the event controller reports a local `ship_death` event.

### Event controller

`GameplayEventController` owns event type routing.

It reads event dictionaries, checks `type`, converts event positions into visual coordinates, and calls the appropriate `GameplayEffects` spawn method.

Current routed event behavior:

```text
bullet_blast
-> spawn_bullet_blast(visual_position)

ship_death
-> if player_id == self_id, call self-death handler
-> spawn_ship_death(visual_position)

radial_effect_started
-> spawn_torpedo_explosion(visual_position)

pickup_collected
-> spawn_pickup_collected(visual_position)

pickup_effect_applied
-> no current visual effect
```

Events without a supported route are ignored by this client presentation flow.

### Effects presenter

`GameplayEffects` owns visual effect node creation, animation startup, sound startup, and effect cleanup.

It instantiates effect scenes under the configured gameplay owner node and places them in visual coordinates.

Current effect scenes:

```text
bullet_blast          -> res://scenes/animations/bullet_blast.tscn
ship_death            -> res://scenes/animations/ship_death.tscn
radial_effect_started -> res://scenes/animations/torpedo_explosion.tscn
pickup_collected      -> res://scenes/pickups/pickup_collect.tscn
```

### Local death flow

`GameplayDeathFlow` owns the client presentation response to local self-death events.

For lives above zero, it applies the remaining lives to the HUD and sets dead/respawn presentation using the event respawn delay.

For lives equal to zero, it delegates to `MatchEndFlow.handle_local_player_eliminated(event)` when match-end flow is configured.

This keeps final local elimination presentation separate from authoritative room match-over presentation.

### Match-end collaborator

`MatchEndFlow` may request game-over audio through `GameplayEventFlow.play_game_over_sound_after_delay()`.

The event/effects path owns game-over sound delay and one-shot gating. Match-end orchestration does not play audio directly.

## Protocols and APIs

### State event input

Gameplay events enter the client as the `events` array on a gameplay state packet.

`GameplayStatePacketReader` normalizes that packet field into:

```text
server_events
```

If the packet event field is missing or is not an array, the normalized value is an empty array.

### Runtime application order

Gameplay state application routes server events after world state and alive/respawn restoration have been applied.

Current application order in `GameplayStateApplyFlow`:

```text
1. Apply gameplay state to devtools context.
2. Mark gameplay input as having received gameplay state.
3. Apply gameplay-state summary to HUD flow.
4. Apply world state.
5. Apply alive/respawn restoration.
6. Apply server events.
7. Mark gameplay state as received.
```

This order matters because event effects need current world-sync visual coordinate state before converting server event positions.

### Event coordinate conversion

Event positions are read from packet fields:

```text
x
y
```

The event controller builds a server-space position:

```text
Vector2(event.x, event.y)
```

Then it calls the configured converter:

```gdscript
visual_position_for_server_position.call(event_position)
```

The active runtime configures this converter from:

```text
WorldSync.visual_position_for_server_position(...)
```

If no converter is configured, the client logs a warning and skips the effect.

If event coordinates contain a `Callable`, the client logs a warning and skips the effect.

### Supported event fields

The current event/effects flow reads only the fields it needs for presentation routing.

Common fields:

```text
type
x
y
```

Local death fields:

```text
player_id
lives
respawn_delay
```

Pickup collection fields may be present on the event, but the visual collection effect currently only requires position after routing:

```text
pickup_id
pickup_type
x
y
```

Radial effect fields may be present on the event, but the current torpedo explosion presentation only requires position after routing:

```text
source_id
effect_type
x
y
```

### Game-over sound request

`GameplayEventFlow.play_game_over_sound_after_delay()` delegates to `GameplayEffects`.

`GameplayEffects` uses generated constants for game-over delay and tracks a local token so delayed sound playback can be invalidated by reset or stop behavior.

The game-over sound is played at most once per event/effects lifecycle until reset.

### HTTP APIs

This flow does not expose HTTP APIs.

It consumes realtime gameplay state that has already entered the client gameplay runtime.

## Data ownership

Gameplay event/effects presentation owns transient client presentation state only.

Current local state includes:

```text
GameplayEventFlow.effects
GameplayEventFlow.gameplay_event_controller
GameplayEffects.owner_node
GameplayEffects.audio_flow
GameplayEffects.game_over_sound_played
GameplayEffects.game_over_sound_token
short-lived effect nodes
effect cleanup timers
local signal connections
```

This state is not durable.

It is not authoritative.

It is reset when the gameplay lifecycle resets or when effect nodes complete their animation/sound cleanup.

## Visual effects

### Bullet blast

`bullet_blast` events spawn:

```text
client/scenes/animations/bullet_blast.tscn
```

The effect is positioned at the converted visual event position and uses `Constants.EFFECT_Z_INDEX`.

The effect expects:

```text
AnimatedSprite2D
AsteroidDestroyed
```

The sprite plays the `bullet_blast` animation. The sound is played through `GameplayAudioFlow.play_bullet_blast_sound(...)`.

Cleanup uses both sound completion and a timer fallback based on sound length plus generated cleanup padding.

### Ship death

`ship_death` events spawn:

```text
client/scenes/animations/ship_death.tscn
```

The effect is positioned at the converted visual event position and uses `Constants.EFFECT_Z_INDEX`.

The effect expects:

```text
AnimatedSprite2D
ShipDeath
```

The sprite plays the default animation. The sound is played through `GameplayAudioFlow.play_ship_death_sound(...)`.

Cleanup is guarded so the node is queued for deletion only once.

### Pickup collected

`pickup_collected` events spawn:

```text
client/scenes/pickups/pickup_collect.tscn
```

The effect is positioned at the converted visual event position and uses:

```text
Constants.PICKUP_Z_INDEX + 1
```

The effect expects:

```text
GPUParticles2D
AudioStreamPlayer2D
```

The particles restart and emit locally. The sound is played through `GameplayAudioFlow.play_pickup_collected_sound(...)`.

Cleanup waits for the longer of particle lifetime and sound length, plus a small padding interval.

This path is separate from persistent pickup node sync. Pickup node creation, interpolation, lifespan presentation, and removal belong to world-sync pickup presentation.

### Radial effect started

`radial_effect_started` events currently spawn the torpedo explosion scene:

```text
client/scenes/animations/torpedo_explosion.tscn
```

The effect is positioned at the converted visual event position and uses `Constants.EFFECT_Z_INDEX`.

The effect expects:

```text
AnimatedSprite2D
TorpedoExplosionSound
```

The sprite is scaled to the generated torpedo radial diameter:

```text
TORPEDO_RADIAL_ZONE_COUNT * TORPEDO_RADIAL_ZONE_WIDTH * 2
```

The source frame used for scale measurement is frame `5` of the `torpedo_explosion` animation.

Sound is optional. If present, it is played through `GameplayAudioFlow.play_torpedo_explosion_sound(...)`.

## Data ownership boundaries

### Packet source of truth

Gameplay event packet structure is generated from shared packet definitions.

Current relevant source and generated files:

```text
shared/packets/gameplay.toml
client/scripts/generated/networking/packets/packets.gd
services/game-server/internal/game/packets.go
```

The client event/effects flow consumes generated packet constants. It does not own packet schema.

### Server event source

The game server owns gameplay event production and event-to-packet adaptation.

Current relevant server files:

```text
services/game-server/internal/game/events/events.go
services/game-server/internal/game/packets.go
```

The server-side queue is `pendingPresentationEvents`. It stores packet-facing `EventState` values for client presentation. It is not client-owned state.

### Constants source of truth

Effect z-indexes, game-over sound delay, torpedo radial size, and cleanup padding values are generated constants.

Current relevant generated client file:

```text
client/scripts/generated/constants/constants.gd
```

The underlying constants source belongs to shared data and the data-sync pipeline, not this service doc.

## Code map

### Primary event/effects implementation

* `client/scripts/gameplay/events/gameplay_event_lifecycle_flow.gd` - Wires event flow, death flow, visual coordinate conversion, and match-end collaboration into gameplay runtime.
* `client/scripts/gameplay/events/gameplay_event_flow.gd` - Public event flow wrapper for applying server events, requesting game-over sound, and resetting game-over sound state.
* `client/scripts/gameplay/events/gameplay_event_controller.gd` - Routes server event dictionaries to supported presentation handlers.
* `client/scripts/gameplay/events/gameplay_death_flow.gd` - Handles local self-death presentation and final local elimination delegation.
* `client/scripts/gameplay/effects/gameplay_effects.gd` - Instantiates visual effect scenes, starts animations/sounds, scales torpedo explosions, and cleans up effect nodes.
* `client/scripts/gameplay/audio/gameplay_audio_flow.gd` - Plays event-local sounds and game-over sound requests.

### Runtime callers

* `client/scripts/gameplay/state/gameplay_state_packet_reader.gd` - Normalizes packet events into `server_events`.
* `client/scripts/gameplay/state/gameplay_state_apply_flow.gd` - Applies server events after world state and alive/respawn restoration.
* `client/scripts/gameplay/runtime/gameplay_flow_composer.gd` - Constructs and configures `GameplayEventLifecycleFlow`.
* `client/scripts/shell/gameplay_shell_flow.gd` - Owns gameplay runtime shell delegation and reset.
* `client/scripts/gameplay/gameplay_composition.gd` - Constructs gameplay shell, match-end flow, and surrounding gameplay collaborators.

### Coordinate conversion collaborators

* `client/scripts/world/world_sync.gd` - Exposes `visual_position_for_server_position(...)`.
* `client/scripts/world/player_render/player_render_api.gd` - Routes coordinate conversion through the active player-render API.
* `client/scripts/world/player_render/view_anchor_sync.gd` - Wraps ViewAnchor visual/server coordinate conversion.

### Match-end and HUD collaborators

* `client/scripts/gameplay/match_end/match_end_flow.gd` - Handles final local elimination and authoritative room match-over presentation orchestration.
* `client/scripts/shell/gameplay_hud_flow.gd` - Owns HUD lives, dead, respawn, game-over, and match-over visibility behavior.
* `client/scripts/shell/gameplay_menu_flow.gd` - Owns gameplay menu and match-over overlay menu presentation.

### Effect scenes

* `client/scenes/animations/bullet_blast.tscn`
* `client/scenes/animations/ship_death.tscn`
* `client/scenes/animations/torpedo_explosion.tscn`
* `client/scenes/pickups/pickup_collect.tscn`

### Generated and source-of-truth boundaries

* `client/scripts/generated/networking/packets/packets.gd`
* `client/scripts/generated/constants/constants.gd`
* `shared/packets/gameplay.toml`
* `services/game-server/internal/game/events/events.go`
* `services/game-server/internal/game/packets.go`

### Non-owning boundaries

* `client/scripts/networking/` - Owns WebSocket transport, packet decode, packet classification, and signal dispatch.
* `client/scripts/world/` - Owns persistent world entity sync, ViewAnchor, visual coordinates, interpolation, and pickup node presentation.
* `client/scripts/ui/` - Owns mounted UI widgets and result-window controls.
* `services/game-server/` - Owns simulation, event production, event packet adaptation, room state, and authoritative match lifecycle.
* `shared/` - Owns packet and constants source-of-truth files.

## Tests

Relevant tests include:

* `client/tests/unit/test_gameplay_state_packet_reader.gd`
* `client/tests/unit/test_gameplay_state_apply_flow.gd`
* `client/tests/unit/gameplay/test_gameplay_event_controller.gd`
* `client/tests/unit/gameplay/events/test_gameplay_death_flow.gd`
* `client/tests/unit/gameplay/effects/test_gameplay_effects.gd`
* `client/tests/unit/gameplay/test_gameplay_flow_composer.gd`
* `client/tests/unit/gameplay/match_end/test_match_end_flow.gd`
* `client/tests/unit/gameplay/test_gameplay_alive_restore_flow.gd`
* `client/tests/unit/test_world_sync.gd`
* `client/tests/unit/world/player_render/test_view_anchor_sync.gd`

Use the normal client GUT verification flow when changing event/effects presentation behavior.

## Related docs

* [Gameplay Event Presentation](./!README.md)
* [Gameplay Audio Flow](gameplay-audio-flow.md)
* [Gameplay Runtime](../gameplay-runtime/!README.md)
* [Gameplay State Application](../gameplay-runtime/gameplay-state-application.md)
* [Match End Flow](../match-end-flow/!README.md)
* [Match End Orchestration](../match-end-flow/match-end-orchestration.md)
* [Pickup Presentation](../world-sync/pickup-presentation.md)
* [View Anchor And Visual Coordinates](../world-sync/view-anchor-and-visual-coordinates.md)
* [HUD And Gameplay UI](../hud-and-gameplay-ui.md)
* [Gameplay packets](../../../protocol/stubs/gameplay-packets.md) - Stub: gameplay realtime packet documentation.
* [Realtime websocket protocol](../../../protocol/stubs/realtime-websocket-protocol.md) - Stub: realtime websocket protocol documentation.
* [Radial effects](../../../systems-design/combat/stubs/radial-effects.md) - Stub: radial effects design documentation.
* [Pickup entities](../../../systems-design/entities/stubs/pickup-entities.md) - Stub: pickup entity design documentation.

## Notes

Legacy docs correctly identified the event/effects/audio path as the owner of game-over sound playback, delay, and one-shot gating. Current implementation keeps that ownership in `GameplayEffects` and `GameplayAudioFlow`.

`pickup_collected` collection effects belong to gameplay event/effects presentation even though pickup node rendering belongs to world sync. This separation lets a pickup collection particle and sound outlive the pickup node that may be removed by the next world-sync state application.

The current client event controller routes `pickup_effect_applied` but does not spawn a visual effect for it. Do not document that event as a visible client effect unless implementation changes.

The event/effects path should stay presentation-only. New event types should be added by extending server event production, packet/schema documentation, and client presentation routing without moving gameplay authority into the client.
