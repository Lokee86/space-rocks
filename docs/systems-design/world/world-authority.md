# World Authority

Parent index: [World](./!INDEX.md)

## Purpose

This document defines the systems-design authority model for world state in Space Rocks.

It explains which systems own world truth, which systems may request changes, which systems may only present projected state, and which invariants must hold across server simulation, realtime packets, client world sync, devtools, data pipelines, and future gameplay systems.

## Overview

World authority is server-owned.

The authoritative world is the active match-local simulation owned by the game server. It includes live entity state, match-local player session state, world-space positions, movement, spawning, collisions, damage consequences, pickup collection, target state, presentation-event production, and lane-native realtime projection inputs.

The current authority split is:

```text
Game server simulation
= authoritative world mutation and state projection

Rooms
= match lifecycle and game-instance ownership

Networking
= transport, session routing, packet classification, and outbound writes

Client
= input collection, request emission, interpolation, visual coordinates, UI, audio, effects, and presentation

Data pipeline
= source files and generated outputs for constants, packets, collision shapes, variants, and other shared data

Devtools
= debug command requests and presentation, with gameplay mutation routed through server-owned adapters
```

The client does not own gameplay truth. It sends intent and renders state returned by the server. A client-side value becomes confirmed gameplay fact only when the authoritative server accepts it and projects the result through realtime state, room state, pause-state output, or server-produced presentation events.

## Conceptual model

The world authority model is built around one active `Game` aggregate per running match.

```text
room starts match
-> room creates or reuses Game
-> Game starts simulation loop
-> networking activates players into the Game
-> clients send input and request intent
-> Game mutates authoritative world state
-> Game.Step advances simulation phases
-> protocol/realtime projects read-only lane state
-> networking sends projected lane packets to clients
-> clients render and interpolate the projected state
```

The authoritative world state is not the whole product state. It is the match-local runtime state needed to simulate and present a match.

Current world-authoritative state includes:

```text
active ship/avatar entities
active projectiles
active asteroids
active pickups
reserved enemy entity storage
player sessions
camera views
spawn counters
radial effects
pending packet-facing presentation events
world simulation options
collision shape catalog
match-local scoring and lifecycle facts
```

World state does not include:

```text
room membership authority
websocket connection authority
local pilot profile persistence
account persistence
inventory ownership
durable progression
HTTP API contracts
client scene/node ownership
client visual interpolation state
client audio/effect playback state
```

Those systems may feed or consume world state, but they are not the authoritative world simulation.

## Authority rules

The game server owns authoritative world mutation.

The server owns:

```text
player activation into active world state
safe initial spawn placement
safe respawn placement
player input acceptance
movement simulation
world wrapping for authoritative positions
projectile spawning
asteroid spawning
pickup spawning and expiry
collision detection participation
damage application consequences
ship death and pending despawn
score and lives mutation
target selection validation
pause and suspension effects on simulation
match-over evaluation inputs
lane packet projection
server-produced presentation events
```

The room system owns match lifecycle around the world.

Rooms own:

```text
when a match can start
which Game instance belongs to a room
when the Game simulation is started
when room state moves to in-game
when room state moves to game-over
when match result summary is resolved
when a game instance is stopped and cleared on return to lobby
```

Rooms do not own entity movement, collisions, scoring, damage, spawning, or lane packet projection.

Networking owns transport and routing.

Networking owns:

```text
websocket upgrade
connection session state
current room routing context
current game-player routing context
packet-family classification
decoded gameplay request handoff
pause-state packet enqueueing
lane packet write timing
server_sent_msec stamping
packet encoding and write errors
```

Networking does not own world rules. It routes requests to the authoritative game instance and sends projected output back to clients.

The client owns presentation and local request construction.

The client owns:

```text
local input collection
outbound packet construction
pointer and target candidate presentation
visual-to-server coordinate conversion for requests
server-to-visual coordinate conversion for presentation
WorldSync application of normalized server state
ViewAnchor and render-anchor presentation
interpolation
entity scene/node creation and cleanup
HUD, UI, audio, and visual effects
spectate camera presentation
debug overlays and readouts
```

The client does not own:

```text
authoritative position
authoritative velocity
spawn validity
respawn validity
collision outcomes
damage outcomes
pickup collection outcomes
score
lives
match lifecycle
target validation
lane packet contents
```

The data pipeline owns shared source material, not runtime decisions.

Shared data and generated outputs define reusable inputs such as:

```text
world dimensions
server tick rate
packet field shapes
packet type strings
collision bodies
asteroid variant catalog
drop tables
constants used by both server and client
```

Generated files are outputs. Runtime authority still belongs to the service that consumes the generated data.

Devtools do not bypass world authority.

Debug commands may request world mutation, but gameplay-affecting effects must route through server-owned command handling and narrow game-owned adapters. Client-only debug presentation can observe and display world state, but it must not create an alternate source of gameplay truth.

## World-state request model

Client-originated gameplay packets are intent, not authority.

Examples:

```text
input
= request to apply movement/fire input

respawn
= request to respawn if server rules allow it

pause_request
= request to toggle server-owned pause state

select_target_at_position_request
= request to select a target candidate at a position

clear_target_request
= request to clear current target state
```

The server may accept, ignore, reject, or transform the request according to current world state and player lifecycle.

The confirmation path is server output:

```text
lane packet
player_pause_state packet
room snapshot
match result summary
server-produced presentation event
```

Client UI state must not be treated as confirmation of world mutation.

## World-state projection model

Lane packets are projections of authoritative state. They are not the authority itself.

Current projected world state is organized by lane as follows:

```text
world lane
= active ship/avatar state, bullets, asteroids, pickups

overlay lane
= receiver-local overlay and HUD state

session lane
= match-local durable player session state and lifecycle read models

event_batch
= transient packet-facing presentation events
```

Important projection boundaries:

```text
world lane ship records
= active ship/avatar state only

session lane player records
= match-local durable player session state

session lane lifecycle records
= active, pending_respawn, or eliminated lifecycle read model

event_batch
= transient packet-facing presentation events

server_sent_msec
= networking write timestamp, not simulation state
```

Clients must not infer lifecycle from active ship presence alone. Pending-respawn and eliminated players may be absent from `players` while still existing in `player_sessions` and `player_lifecycle`.

## World-space authority

The server owns bounded world coordinates.

The current world is toroidal. Server simulation stores one bounded authoritative position per live entity and wraps moved entities inside the current world bounds. Cross-edge distance, direction, collision placement, respawn safety, visibility, despawn, radial coverage, and spawn aiming use wrapped spatial helpers.

The client owns continuous visual coordinates.

The client receives bounded server positions and converts them into visual positions relative to the active ViewAnchor/render anchor. This preserves presentation continuity when a server position wraps at an edge.

The coordinate split is:

```text
server position
= authoritative bounded world coordinate

visual position
= client presentation coordinate relative to the active render anchor
```

Changing ViewAnchor, spectate target, camera focus, interpolation, or visual coordinate conversion does not change authoritative world state.

## Entity authority

Runtime entities are authoritative only inside the game-server simulation aggregate.

Current entity-family authority:

```text
ships
= server-created active avatars backed by player sessions

projectiles
= server-spawned weapon outputs

asteroids
= server-spawned world hazards with variant, size, health, and collision behavior

pickups
= server-spawned collectible entities with lifecycle, collection, and effect intent

enemies
= reserved runtime storage, not a completed gameplay authority model yet
```

The client may create scene nodes for these entities, interpolate their positions, choose presentation scenes, and play effects. It must not decide that an entity exists, was destroyed, collected, hit, damaged, scored, or despawned unless that fact comes from server-owned state or events.

## Player and world authority split

A player is not the same thing as a live world ship.

The current split is:

```text
websocket session
= connection-local networking state

room member
= room membership and lobby identity

playerSession
= match-local durable player state

runtime.Ship
= active live avatar/world entity

client player node
= presentation node for server-owned ship state
```

This distinction matters because a player can remain in the match without an active ship. Respawn, elimination, scoring, lives, and lifecycle classification are session-owned or rule-owned server facts, not client node facts.

## Simulation authority

`Game.Step(delta)` is the top-level authoritative simulation phase coordinator.

The current active-match flow is:

```text
step player sessions
-> step player weapons
-> step players
-> remove ready players
-> step asteroid spawning
-> step asteroids
-> step bullets
-> step pickups
-> step collisions
-> step radial effects
-> notify simulation step observers
```

When the match is already over, the simulation uses a reduced path. It still advances asteroids, bullets, pickups, radial effects, and simulation observers, but skips player weapons, player movement, player removal, new asteroid spawning, and collisions.

Subsystem packages may own narrow policy or helper behavior, but the aggregate owns mutation of the authoritative world store.

Examples:

```text
motion
= movement integration and wrapping helpers

space
= bounds, wrap, distance, direction, and shortest delta helpers

physics
= collision primitive math and loaded collision shape bodies

damage
= pure damage resolution result calculation

scoring
= pure score award calculation

rules
= pure match decision evaluation

spawning
= spawn planning and projectile/asteroid construction support

radial effects
= radial zone timing and hit-intent production
```

Those packages do not own the full world. The game aggregate adapts their outputs into authoritative state mutation.

## Invariants

World authority must preserve these invariants:

```text
The game server is the source of authoritative gameplay world state.

A client request is intent until confirmed by server output.

The client does not own collision, damage, scoring, lives, respawn, spawn, pickup, or match outcomes.

One active Game aggregate owns one match-local authoritative simulation.

Rooms own Game lifecycle; Game owns simulation mutation.

Networking owns packet transport and routing; Game owns gameplay consequences.

Lane packets project server state; they are not mutable world storage.

Client world sync renders projected state; it does not create gameplay truth.

Server positions are bounded authoritative coordinates.

Client visual positions are presentation coordinates.

World wrapping is authoritative on the server and presentational on the client.

Player lifecycle must not be inferred from active ship presence alone.

A player session may exist without an active ship.

Devtools gameplay mutation must route through server-owned debug command handling and game-owned adapters.

Generated packet and constants files must not be hand-edited as authority.

Shared data can define inputs, but runtime systems own how those inputs are applied.

Future player builds, inventory, enemies, waves, and encounter systems must feed authoritative setup through explicit server-owned seams.
```

## Participating systems

```text
Game server simulation
= authoritative match-local world mutation and projection

Game server rooms
= match lifecycle, game instance ownership, game-over transition, and reset-to-lobby lifecycle

Game server networking
= transport, decoded request routing, session context, and lane packet writes

Realtime protocol
= packet contract for client intent and server state projection

Client gameplay runtime
= normalized state application, runtime processing, presentation fanout, and session consequences

Client world sync
= rendered world state, ViewAnchor mapping, visual coordinates, interpolation, and target read models

Data pipeline
= shared constants, packet schemas, generated outputs, collision shape data, and other source-of-truth inputs

Devtools
= debug requests, overlays, telemetry, and server-routed debug mutation

Player-data and API systems
= durable profile/account/result persistence outside live world authority
```

## Service implementation

The detailed implementation belongs in service docs.

The current authoritative implementation is centered on:

```text
Game Aggregate
Simulation Loop And Phase Order
Runtime Entity Store
Lane Packet Projection
Toroidal Space And Motion
Visibility And Despawn
Asteroid Spawning And Variants
Player Lifecycle Overview
Player Input Routing
Player Respawn
Collision To Damage Flow
Pickup Entity Lifecycle
```

Client-side presentation implementation is centered on:

```text
World Sync Coordinator
View Anchor And Visual Coordinates
Entity Sync Owners
Gameplay State Application
Input And Targeting
Gameplay Events And Effects
```

This systems-design document owns the authority model and invariants. Service docs own code maps, exact implementation paths, tests, and verification commands.

## Protocol and data relationships

World authority is exposed to clients through realtime gameplay packets.

The gameplay packet protocol carries:

```text
client intent -> server
server authoritative state -> client
server-owned pause state -> client
server-produced presentation events -> client
```

Packet schema source files define wire shapes and generated helpers. They do not own simulation behavior.

World dimensions are shared source data. The server uses generated world constants for authoritative toroidal wrapping. The client uses matching generated constants for visual shortest-delta presentation. These constants must stay aligned across Go and GDScript outputs.

Collision shape data is exported from Godot scene collision nodes and loaded by the server for authoritative collision bodies. The source scene geometry can originate in the client project, but runtime collision outcomes remain server-owned.

## Planning boundary

Future systems such as loadouts, inventory, enemies, waves, bosses, richer spawn plans, social play, progression rewards, and encounter modes may feed new inputs into world setup.

Those systems must not invert world authority.

The durable rule is:

```text
future durable/player/platform systems
-> provide eligible setup facts or requests

game-server authoritative seams
-> resolve those facts into match-local world state

client
-> presents the resulting state and sends intent
```

For example, a future loadout selection may define which ship and weapons a player is eligible to start with. It should not allow the client to directly mutate active ship stats, weapon cooldowns, projectile damage, collision shape, or spawn position.

## Related docs

* [World](./!INDEX.md)
* [Toroidal Wrap](toroidal-wrap.md)
* [Spawning And Space](spawning-and-space.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Game Aggregate](../../services/game-server/simulation/runtime/game-aggregate.md)
* [Simulation Loop And Phase Order](../../services/game-server/simulation/runtime/simulation-loop-and-phase-order.md)
* [Runtime Entity Store](../../services/game-server/simulation/runtime/runtime-entity-store.md)
* [Lane Packet Projection](../../services/game-server/simulation/runtime/lane-packet-projection.md)
* [Toroidal Space And Motion](../../services/game-server/simulation/world/toroidal-space-and-motion.md)
* [Visibility And Despawn](../../services/game-server/simulation/world/visibility-and-despawn.md)
* [Asteroid Spawning And Variants](../../services/game-server/simulation/world/asteroid-spawning-and-variants.md)
* [Player Lifecycle Overview](../../services/game-server/simulation/players/player-lifecycle-overview.md)
* [Player Input Routing](../../services/game-server/simulation/players/player-input-routing.md)
* [Player Respawn](../../services/game-server/simulation/players/player-respawn.md)
* [Collision To Damage Flow](../../services/game-server/simulation/combat/collision-to-damage-flow.md)
* [Pickup Entity Lifecycle](../../services/game-server/simulation/pickups/pickup-entity-lifecycle.md)
* [Client](../../services/client/!INDEX.md)
* [World Sync Coordinator](../../services/client/world-sync/world-sync-coordinator.md)
* [View Anchor And Visual Coordinates](../../services/client/world-sync/view-anchor-and-visual-coordinates.md)
* [Entity Sync Owners](../../services/client/world-sync/entity-sync-owners.md)
* [Gameplay Packets](../../protocol/gameplay-packets.md)
* [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md)
* [Constants Pipeline](../../data/constants.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Collision Shape Data](../../data/collision-shape-data.md)
* [Realtime Client Server Flow](../../domains/technical/realtime-client-server-flow.md)
* [Gameplay Session Flow](../../domains/player-experience/gameplay-session-flow.md)
* [Player Build And Loadouts](../../planning/domains/gameplay/player-build-and-loadouts.md)

## Notes

Gameplay state is server-authoritative, while the client owns presentation.

The current implementation still keeps many world mutations under the root `game.Game` aggregate. Focused packages provide policy and helper seams, but that does not make them independent world authorities.

The active client ViewAnchor model is intentionally presentation-only. It can change what the player sees and how coordinates are converted, but it does not change world ownership.

World authority should be treated as a permanent design constraint, not a temporary implementation detail.

