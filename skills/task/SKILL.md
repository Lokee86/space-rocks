
Yesterday 10:59 PM
Alright, sounds like we do this in stages, yeah?

Yes. This should be staged because transport, packet application, readiness, and presentation timing are coupled. Moving all four boundaries at once would make regressions hard to isolate.

A sensible order is:

Extract RealtimeTransportSession
Move WebRTC lifecycle, signaling, polling, smoke diagnostics, and failure handling out of ClientConnectionService.
Preserve packet flow exactly.

Introduce RealtimePacketPipeline
Move router ownership, reset, readiness, packet application, and resync state behind one stable API.
Keep the current generic presentation notification temporarily.

Remove router reach-through
Stop GameplaySessionController from calling get_realtime_router() and get_gameplay_readiness().
Give it a stable pipeline or presentation-facing dependency instead.

Extract PresentationBridge
Move lane fanout, devtools adaptation, alive-state restoration, and _presentation_dirty ownership out of GameplaySessionController.
Preserve once-per-frame coalescing.

Replace the generic packet signal
Remove gameplay_packet_received(packet) as the presentation trigger.
Replace it with a semantic pipeline event such as presentation_pending.

Thin the remaining connection service
Review what remains after the extractions and decide whether the outbound façade and inbound non-gameplay signal fanout need further separation.

Each stage should preserve behavior and have its own verification boundary. The transport extraction is the safest first stage because it has the clearest ownership boundary and does not require changing presentation semantics.

Branched from Client Server Architecture Refactor


Today 2:43 AM
alright, what's the plan for #2.

Stage 2 plan — Introduce RealtimePacketPipeline
Goal
Create one owner for realtime protocol state and packet application without changing presentation behavior yet.

After this stage:

ClientConnectionService
  → RealtimePacketPipeline.apply_packet(packet)
  → RealtimeRouter mutates state
  → existing gameplay packet notification still fires
  → GameplaySessionController still marks presentation dirty
  → existing frame-delayed presentation fanout remains unchanged
This stage changes ownership, not presentation semantics.

New boundary
Add a concrete RealtimePacketPipeline under the client networking/realtime area.

It should own:

RealtimeRouter

gameplay readiness state

router reset/replacement

compact packet expansion

realtime packet validation and routing

baseline/resync state already held by the router

the notification that a gameplay packet was successfully applied

Its initial public surface should remain small:

apply_packet(packet)
reset()
get_readiness()
get_router()
get_router() is temporarily acceptable in this stage because removing router reach-through is Stage 3. The important change is that the connection service no longer owns or replaces the router itself.

Work sequence
1. Establish the pipeline class
Create the pipeline with a router constructed internally.

The pipeline becomes the only code allowed to:

instantiate RealtimeRouter,

replace it during reset,

route realtime lane packets,

determine whether application succeeded,

emit the post-application notification.

No behavior should move out of RealtimeRouter itself unless it is currently implemented directly in ClientConnectionService.

2. Move realtime packet application
Move the current logic that takes an inbound gameplay/realtime packet and sends it into the router.

The complete synchronous operation should live together:

receive packet
  → expand if compact
  → validate supported packet/lane form
  → route into router
  → mutate lane state
  → emit packet-applied notification
The notification must occur only after successful state application, matching current behavior.

3. Move reset ownership
Move reset_realtime_protocol_state() or its equivalent into the pipeline.

The pipeline should replace or reset its own router and readiness state. ClientConnectionService should only request:

realtime_pipeline.reset()
It must not instantiate RealtimeRouter directly afterward.

4. Delegate from ClientConnectionService
Replace its direct realtime responsibilities with delegation.

Conceptually:

_on_gameplay_packet_received(packet):
    realtime_pipeline.apply_packet(packet)
The service may continue exposing compatibility methods temporarily:

get_realtime_router()
get_gameplay_readiness()
reset_realtime_protocol_state()
But those methods should delegate directly to the pipeline and contain no protocol logic.

This avoids changing every caller during the same stage.

5. Preserve the existing signal contract
Keep the current gameplay packet signal externally visible for now.

The pipeline should emit a packet-applied signal, and ClientConnectionService can forward it under the existing signal name.

That keeps this path unchanged from the perspective of SessionNetworkController and GameplaySessionController:

pipeline packet applied
  → connection service forwards gameplay_packet_received
  → session controller receives it
  → presentation dirty flag is set
Renaming or removing that signal belongs in Stage 5.

6. Preserve reconnect and lifecycle ordering
Audit every place that currently resets or retrieves the router:

initial connection setup,

disconnect,

reconnect,

room leave,

gameplay session start,

match teardown,

test setup.

Every reset must now target the pipeline.

Because Stage 3 has not happened yet, callers retrieving the router must still receive the pipeline’s current router instance after a reset.

Expected file impact
Likely primary files:

new realtime_packet_pipeline.gd

client_connection_service.gd

realtime router-related tests

connection service tests

possibly composition/bootstrap code where the connection service is constructed

GameplaySessionController and presentation adapters should require little or no behavioral change during this stage.

Explicit non-goals
Do not yet:

extract presentation fanout,

move _presentation_dirty,

remove frame coalescing,

remove get_realtime_router(),

remove get_gameplay_readiness(),

rename the generic gameplay packet signal,

redesign RealtimeRouter,

split individual lane stores,

alter packet formats,

alter WebSocket or WebRTC transport behavior.

Those changes would blur the verification boundary.

Required invariants
The stage is complete only if these remain true:

Every accepted packet mutates realtime state exactly once.

The packet notification fires only after mutation.

Presentation still updates on the later frame, not immediately.

Multiple packets in one frame still coalesce into one presentation fanout.

Reset creates clean protocol state.

Reconnect callers receive the new active router, not a stale instance.

Readiness and baseline behavior remain unchanged.

Both WebSocket-delivered and WebRTC-delivered gameplay packets use the same pipeline.

ClientConnectionService contains no direct router mutation or construction.

Verification strategy
Structural checks
Confirm:

only RealtimePacketPipeline constructs RealtimeRouter,

only the pipeline invokes the router’s packet-routing entry point,

ClientConnectionService delegates reset, router access, and readiness access,

no new duplicate realtime state exists.

Automated behavior checks
Cover:

applying a valid packet mutates the expected lane,

applying a packet emits one post-application notification,

rejected packets do not emit it,

reset clears readiness and lane state,

reset exposes a fresh active router,

sequential packets preserve order,

compact and expanded packet forms behave identically,

both transport entry paths reach the pipeline.

Integration check
Run a client session and verify:

connect
→ enter gameplay
→ receive baseline
→ receive realtime updates
→ presentation updates normally
→ disconnect/reconnect
→ receive a new baseline
→ presentation resumes without stale state
End state
ClientConnectionService will still look somewhat broad, but its realtime protocol ownership will be gone.

The key result is:

Before:
ClientConnectionService owns transport + router + readiness + packet application

After:
ClientConnectionService owns connection coordination
RealtimePacketPipeline owns realtime protocol state and application
That gives Stage 3 a clean target: remove the remaining router/readiness reach-through without also having to establish their owner at the same time.


write this out in a SKILL.md

SKILL.md 


Library
/
SKILL.md


# Realtime Packet Pipeline Extraction

## Purpose

Introduce a `RealtimePacketPipeline` that becomes the single owner of client realtime protocol state and packet application.

This skill applies when extracting realtime packet ownership from `ClientConnectionService` while preserving all existing presentation timing and externally visible behavior.

The change is an ownership refactor, not a presentation redesign.

## Target State

After this work, the packet path should be:

```text
ClientConnectionService
  → RealtimePacketPipeline.apply_packet(packet)
  → RealtimeRouter mutates state
  → existing gameplay packet notification still fires
  → GameplaySessionController still marks presentation dirty
  → existing frame-delayed presentation fanout remains unchanged
```

`RealtimePacketPipeline` becomes the only owner of:

- `RealtimeRouter`
- gameplay readiness state
- router reset or replacement
- compact packet expansion
- realtime packet validation and routing
- baseline and resync state already held by the router
- the notification that a gameplay packet was successfully applied

## Public Surface

Keep the initial API small and concrete:

```text
apply_packet(packet)
reset()
get_readiness()
get_router()
```

`get_router()` is temporary. It exists only to preserve current callers until the later router reach-through cleanup.

Do not expand the API unless an existing caller requires it.

## Required Work

### 1. Create the pipeline

Add a `RealtimePacketPipeline` class under the existing client networking or realtime area.

The pipeline must construct and own its `RealtimeRouter`.

Only the pipeline may:

- instantiate `RealtimeRouter`
- replace or reset the active router
- route realtime lane packets
- determine whether packet application succeeded
- emit the post-application notification

Do not move behavior out of `RealtimeRouter` unless that behavior currently lives directly in `ClientConnectionService`.

### 2. Move realtime packet application

Move the full realtime packet application path into the pipeline.

The operation must remain synchronous and ordered:

```text
receive packet
  → expand compact form when required
  → validate supported packet or lane form
  → route into the active router
  → mutate lane state
  → emit packet-applied notification
```

The notification must fire only after successful state application.

Rejected or invalid packets must not emit the notification.

### 3. Move reset ownership

Move realtime protocol reset ownership into the pipeline.

The pipeline must reset or replace its own router and readiness state.

`ClientConnectionService` may request:

```text
realtime_pipeline.reset()
```

It must not construct or replace `RealtimeRouter` directly.

### 4. Delegate from `ClientConnectionService`

Replace direct realtime ownership in `ClientConnectionService` with delegation.

The service may temporarily retain compatibility methods such as:

```text
get_realtime_router()
get_gameplay_readiness()
reset_realtime_protocol_state()
```

These methods must delegate directly to the pipeline and contain no realtime protocol logic.

Existing inbound gameplay packet handlers should delegate to:

```text
realtime_pipeline.apply_packet(packet)
```

### 5. Preserve the existing signal contract

Do not remove or rename the current gameplay packet signal in this stage.

The pipeline should emit a post-application signal.

`ClientConnectionService` may forward that signal under the existing public signal name.

The external path must remain behaviorally unchanged:

```text
pipeline packet applied
  → connection service forwards gameplay_packet_received
  → session controller receives it
  → presentation dirty flag is set
```

Signal cleanup belongs to a later stage.

### 6. Preserve lifecycle ordering

Audit all places that currently reset, retrieve, or depend on realtime protocol state, including:

- initial connection setup
- disconnect
- reconnect
- room leave
- gameplay session start
- match teardown
- test setup

Every reset must target the pipeline.

Current callers that retrieve the router must receive the pipeline's active router after a reset.

Do not leave stale router or readiness references active.

## Constraints

- Preserve behavior.
- Preserve packet ordering.
- Preserve synchronous packet application.
- Preserve frame-delayed presentation.
- Preserve presentation coalescing.
- Preserve WebSocket and WebRTC delivery behavior.
- Preserve the existing gameplay packet signal contract.
- Do not redesign `RealtimeRouter`.
- Do not split lane stores.
- Do not alter packet formats.
- Do not move `_presentation_dirty`.
- Do not extract presentation fanout.
- Do not remove `get_realtime_router()`.
- Do not remove `get_gameplay_readiness()`.
- Do not rename the gameplay packet signal.
- Do not mix Stage 3 or later cleanup into this work.
- Avoid wrappers and helper layers unless required to preserve an existing public contract.
- Prefer direct ownership transfer and concrete delegation.

## Expected File Scope

Likely files include:

- a new `realtime_packet_pipeline.gd`
- `client_connection_service.gd`
- realtime router or pipeline tests
- connection service tests
- composition or bootstrap code if dependency construction changes

`GameplaySessionController` and presentation adapters should require little or no behavioral change.

If another file is required, touch only the smallest necessary additional file and report why.

## Completion Invariants

The work is complete only when all of the following are true:

1. Every accepted packet mutates realtime state exactly once.
2. The packet notification fires only after successful mutation.
3. Rejected packets do not emit the notification.
4. Presentation still updates on a later frame.
5. Multiple packets in one frame still coalesce into one presentation fanout.
6. Reset produces clean realtime protocol state.
7. Callers receive the new active router after reset.
8. Readiness and baseline behavior remain unchanged.
9. WebSocket-delivered and WebRTC-delivered gameplay packets use the same pipeline.
10. `ClientConnectionService` contains no direct router construction.
11. `ClientConnectionService` contains no direct router mutation.
12. Only `RealtimePacketPipeline` owns the active `RealtimeRouter`.

## Verification

### Structural verification

Confirm that:

- only `RealtimePacketPipeline` constructs `RealtimeRouter`
- only `RealtimePacketPipeline` invokes the router packet-routing entry point
- `ClientConnectionService` delegates reset, router access, and readiness access
- no duplicate realtime state exists
- no stale router reference survives reset

### Automated behavior verification

Cover:

- valid packet application mutates the expected lane
- successful application emits exactly one notification
- rejected application emits no notification
- reset clears readiness and lane state
- reset exposes a fresh active router
- sequential packets preserve order
- compact and expanded packet forms behave identically
- WebSocket and WebRTC entry paths reach the same pipeline

### Integration verification

Verify this lifecycle:

```text
connect
→ enter gameplay
→ receive baseline
→ receive realtime updates
→ presentation updates normally
→ disconnect or reconnect
→ receive a new baseline
→ presentation resumes without stale state
```

## Non-Goals

Do not perform any of the following in this skill:

- remove router reach-through from `GameplaySessionController`
- extract `PresentationBridge`
- move presentation dirty-state ownership
- replace frame coalescing
- remove the generic gameplay packet signal
- split outbound request ownership
- split non-gameplay inbound signal fanout
- redesign transport ownership
- change protocol semantics

Those are separate stages with separate verification boundaries.

## Final Result

Before:

```text
ClientConnectionService
  owns connection coordination
  owns realtime router
  owns readiness
  owns packet application
```

After:

```text
ClientConnectionService
  owns connection coordination
  delegates realtime packet application

RealtimePacketPipeline
  owns realtime protocol state
  owns router lifecycle
  owns readiness
  owns packet application
```

This creates the ownership boundary required for the later removal of router and readiness reach-through.
Library
/
SKILL.md


# Realtime Packet Pipeline Extraction

## Purpose

Introduce a `RealtimePacketPipeline` that becomes the single owner of client realtime protocol state and packet application.

This skill applies when extracting realtime packet ownership from `ClientConnectionService` while preserving all existing presentation timing and externally visible behavior.

The change is an ownership refactor, not a presentation redesign.

## Target State

After this work, the packet path should be:

```text
ClientConnectionService
  → RealtimePacketPipeline.apply_packet(packet)
  → RealtimeRouter mutates state
  → existing gameplay packet notification still fires
  → GameplaySessionController still marks presentation dirty
  → existing frame-delayed presentation fanout remains unchanged
```

`RealtimePacketPipeline` becomes the only owner of:

- `RealtimeRouter`
- gameplay readiness state
- router reset or replacement
- compact packet expansion
- realtime packet validation and routing
- baseline and resync state already held by the router
- the notification that a gameplay packet was successfully applied

## Public Surface

Keep the initial API small and concrete:

```text
apply_packet(packet)
reset()
get_readiness()
get_router()
```

`get_router()` is temporary. It exists only to preserve current callers until the later router reach-through cleanup.

Do not expand the API unless an existing caller requires it.

## Required Work

### 1. Create the pipeline

Add a `RealtimePacketPipeline` class under the existing client networking or realtime area.

The pipeline must construct and own its `RealtimeRouter`.

Only the pipeline may:

- instantiate `RealtimeRouter`
- replace or reset the active router
- route realtime lane packets
- determine whether packet application succeeded
- emit the post-application notification

Do not move behavior out of `RealtimeRouter` unless that behavior currently lives directly in `ClientConnectionService`.

### 2. Move realtime packet application

Move the full realtime packet application path into the pipeline.

The operation must remain synchronous and ordered:

```text
receive packet
  → expand compact form when required
  → validate supported packet or lane form
  → route into the active router
  → mutate lane state
  → emit packet-applied notification
```

The notification must fire only after successful state application.

Rejected or invalid packets must not emit the notification.

### 3. Move reset ownership

Move realtime protocol reset ownership into the pipeline.

The pipeline must reset or replace its own router and readiness state.

`ClientConnectionService` may request:

```text
realtime_pipeline.reset()
```

It must not construct or replace `RealtimeRouter` directly.

### 4. Delegate from `ClientConnectionService`

Replace direct realtime ownership in `ClientConnectionService` with delegation.

The service may temporarily retain compatibility methods such as:

```text
get_realtime_router()
get_gameplay_readiness()
reset_realtime_protocol_state()
```

These methods must delegate directly to the pipeline and contain no realtime protocol logic.

Existing inbound gameplay packet handlers should delegate to:

```text
realtime_pipeline.apply_packet(packet)
```

### 5. Preserve the existing signal contract

Do not remove or rename the current gameplay packet signal in this stage.

The pipeline should emit a post-application signal.

`ClientConnectionService` may forward that signal under the existing public signal name.

The external path must remain behaviorally unchanged:

```text
pipeline packet applied
  → connection service forwards gameplay_packet_received
  → session controller receives it
  → presentation dirty flag is set
```

Signal cleanup belongs to a later stage.

### 6. Preserve lifecycle ordering

Audit all places that currently reset, retrieve, or depend on realtime protocol state, including:

- initial connection setup
- disconnect
- reconnect
- room leave
- gameplay session start
- match teardown
- test setup

Every reset must target the pipeline.

Current callers that retrieve the router must receive the pipeline's active router after a reset.

Do not leave stale router or readiness references active.

## Constraints

- Preserve behavior.
- Preserve packet ordering.
- Preserve synchronous packet application.
- Preserve frame-delayed presentation.
- Preserve presentation coalescing.
- Preserve WebSocket and WebRTC delivery behavior.
- Preserve the existing gameplay packet signal contract.
- Do not redesign `RealtimeRouter`.
- Do not split lane stores.
- Do not alter packet formats.
- Do not move `_presentation_dirty`.
- Do not extract presentation fanout.
- Do not remove `get_realtime_router()`.
- Do not remove `get_gameplay_readiness()`.
- Do not rename the gameplay packet signal.
- Do not mix Stage 3 or later cleanup into this work.
- Avoid wrappers and helper layers unless required to preserve an existing public contract.
- Prefer direct ownership transfer and concrete delegation.

## Expected File Scope

Likely files include:

- a new `realtime_packet_pipeline.gd`
- `client_connection_service.gd`
- realtime router or pipeline tests
- connection service tests
- composition or bootstrap code if dependency construction changes

`GameplaySessionController` and presentation adapters should require little or no behavioral change.

If another file is required, touch only the smallest necessary additional file and report why.

## Completion Invariants

The work is complete only when all of the following are true:

1. Every accepted packet mutates realtime state exactly once.
2. The packet notification fires only after successful mutation.
3. Rejected packets do not emit the notification.
4. Presentation still updates on a later frame.
5. Multiple packets in one frame still coalesce into one presentation fanout.
6. Reset produces clean realtime protocol state.
7. Callers receive the new active router after reset.
8. Readiness and baseline behavior remain unchanged.
9. WebSocket-delivered and WebRTC-delivered gameplay packets use the same pipeline.
10. `ClientConnectionService` contains no direct router construction.
11. `ClientConnectionService` contains no direct router mutation.
12. Only `RealtimePacketPipeline` owns the active `RealtimeRouter`.

## Verification

### Structural verification

Confirm that:

- only `RealtimePacketPipeline` constructs `RealtimeRouter`
- only `RealtimePacketPipeline` invokes the router packet-routing entry point
- `ClientConnectionService` delegates reset, router access, and readiness access
- no duplicate realtime state exists
- no stale router reference survives reset

### Automated behavior verification

Cover:

- valid packet application mutates the expected lane
- successful application emits exactly one notification
- rejected application emits no notification
- reset clears readiness and lane state
- reset exposes a fresh active router
- sequential packets preserve order
- compact and expanded packet forms behave identically
- WebSocket and WebRTC entry paths reach the same pipeline

### Integration verification

Verify this lifecycle:

```text
connect
→ enter gameplay
→ receive baseline
→ receive realtime updates
→ presentation updates normally
→ disconnect or reconnect
→ receive a new baseline
→ presentation resumes without stale state
```

## Non-Goals

Do not perform any of the following in this skill:

- remove router reach-through from `GameplaySessionController`
- extract `PresentationBridge`
- move presentation dirty-state ownership
- replace frame coalescing
- remove the generic gameplay packet signal
- split outbound request ownership
- split non-gameplay inbound signal fanout
- redesign transport ownership
- change protocol semantics

Those are separate stages with separate verification boundaries.

## Final Result

Before:

```text
ClientConnectionService
  owns connection coordination
  owns realtime router
  owns readiness
  owns packet application
```

After:

```text
ClientConnectionService
  owns connection coordination
  delegates realtime packet application

RealtimePacketPipeline
  owns realtime protocol state
  owns router lifecycle
  owns readiness
  owns packet application
```

This creates the ownership boundary required for the later removal of router and readiness reach-through.