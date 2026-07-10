## Realtime Transport Session Extraction Skill

Use this skill when extracting WebRTC transport ownership from `client/scripts/networking/client_connection_service.gd` into a dedicated `RealtimeTransportSession`.

This skill covers Stage 1 only. Preserve all observable behavior and leave realtime packet application, router ownership, readiness, and presentation flow unchanged.

## Goal

Move all WebRTC-specific lifecycle, signaling, polling, failure handling, and diagnostics into one focused transport boundary without changing the existing downstream packet path.

The intended ownership split after this stage is:

```text
ClientConnectionService
  owns WebSocket connection lifecycle, outbound request coordination,
  inbound non-realtime packet coordination, and temporary compatibility forwarding

RealtimeTransportSession
  owns WebRTC peer/channel lifecycle, signaling, polling, readiness,
  failure handling, smoke-test diagnostics, and received packet delivery
```

The existing downstream path must remain:

```text
WebRTC packet
  -> RealtimeTransportSession
  -> ClientConnectionService existing realtime packet handling
  -> RealtimeRouter
  -> existing gameplay packet signal
  -> existing presentation dirty flow
```

## Hard rules

* Preserve behavior.
* Do not redesign the realtime protocol.
* Do not move `RealtimeRouter` ownership.
* Do not change gameplay readiness ownership.
* Do not remove `get_realtime_router()` or `get_gameplay_readiness()` in this stage.
* Do not change `gameplay_packet_received` semantics.
* Do not move `_presentation_dirty`.
* Do not create `RealtimePacketPipeline`.
* Do not create `PresentationBridge`.
* Do not change packet formats or signaling formats.
* Do not change packet scheduling or presentation timing.
* Do not split WebSocket inbound and outbound ownership further in this stage.
* Preserve existing logs, diagnostics, failure behavior, and smoke-test behavior.
* Keep files small and focused.
* Avoid unrelated cleanup or broad renaming.

## Stage plan

### 1. Inventory the current WebRTC surface

Identify every WebRTC-related item in `ClientConnectionService`:

```text
- fields and constants
- signals
- public methods
- private callbacks
- polling logic
- offer handling
- ICE handling
- signaling translation
- reset and cleanup paths
- readiness and failure state
- smoke-test state
- logs and diagnostics
- tests and scene wiring
```

Classify each item as:

```text
- move entirely into RealtimeTransportSession
- remain temporarily as a forwarding facade
- leave for a later stage
```

### 2. Create the transport boundary

Create:

```text
client/scripts/networking/realtime_transport_session.gd
```

`RealtimeTransportSession` owns:

```text
- WebRTC peer/session creation
- WebRTC data-channel lifecycle
- transport startup and teardown
- WebRTC polling
- offer handling
- ICE candidate handling
- signaling callbacks
- channel readiness
- transport failure state
- smoke-test diagnostics
- delivery of received realtime packets
```

It must not own:

```text
- realtime lane state
- packet expansion
- packet validation beyond current transport-level handling
- packet application
- gameplay readiness
- resync state
- presentation scheduling
- WebSocket lifecycle
- the general outbound request facade
```

### 3. Define narrow dependencies

The transport session should receive only what it needs, such as:

```text
- a signaling-send callback or narrow sender dependency
- a callback or signal for received realtime packets
- existing logging or diagnostic dependencies where required
```

Prefer direct concrete dependencies over generic service locators, broad interfaces, or wrapper layers.

A conceptual surface may include:

```text
start(...)
stop()
poll()
handle_offer(...)
handle_ice_candidate(...)
handle_signaling_packet(...)
is_ready()
```

Possible outward signals include:

```text
realtime_packet_received
transport_ready
transport_failed
diagnostic_updated
```

Use existing project terminology and naming conventions where they already exist.

### 4. Move implementation without redesign

Move the current WebRTC implementation as directly as possible.

Preserve:

```text
- control flow
- state transitions
- polling cadence
- signaling order
- packet decoding and forwarding
- readiness behavior
- teardown behavior
- failure behavior
- logs
- diagnostics
- smoke-test behavior
```

Do not combine this extraction with protocol or presentation cleanup.

### 5. Keep ClientConnectionService as coordinator

`ClientConnectionService` should instantiate or receive the transport session and:

```text
- call its poll method from the existing update point
- pass inbound signaling messages to it
- provide the existing signaling-send path
- forward any public signals still required by callers
- invoke transport teardown during disconnect and reset
```

Temporary forwarding methods are allowed only where needed to preserve the current external API. Do not introduce forwarding helpers merely for convenience.

### 6. Preserve lifecycle behavior

The transport lifetime must remain aligned with the connection lifecycle:

```text
connection established
  -> transport becomes available when currently required

disconnect or reset
  -> transport stops
  -> peer and channel state clear
  -> callbacks cannot target stale state

reconnect
  -> a clean transport state is created
```

Use one clear owner. Do not make the transport session a global singleton unless the existing architecture already requires that lifetime.

### 7. Preserve the packet path

After extraction, received WebRTC packets must still enter the same existing `ClientConnectionService` realtime handling path.

Do not route packets directly into new protocol or presentation abstractions during this stage.

### 8. Update focused tests

Add or preserve focused coverage for:

```text
- transport creation
- offer handling
- ICE handling
- signaling-send requests
- polling
- received packet forwarding
- ready signals
- failure signals
- smoke-test behavior
- teardown
- reconnect after teardown
- no stale callbacks after reset
```

Existing connection-service behavior and tests should continue to pass through any temporary compatibility facade.

## Completion criteria

This stage is complete when:

```text
- ClientConnectionService contains no WebRTC peer internals
- ClientConnectionService contains no WebRTC data-channel internals
- ClientConnectionService contains no offer or ICE implementation details
- ClientConnectionService contains no smoke-test state
- WebRTC polling delegates to RealtimeTransportSession
- teardown and reconnect behavior remain unchanged
- received packets follow the existing downstream path
- router ownership remains unchanged
- gameplay readiness ownership remains unchanged
- presentation timing remains unchanged
```

## Expected code map

Primary files are expected to include:

```text
client/scripts/networking/client_connection_service.gd
client/scripts/networking/realtime_transport_session.gd
```

Tests should follow the current client networking test organization. Do not create a new test hierarchy solely for this extraction.

## Stop conditions

Stop and report if:

```text
- preserving behavior requires changing packet formats
- preserving behavior requires changing signaling formats
- the extraction requires moving RealtimeRouter ownership
- the extraction requires changing gameplay readiness
- the extraction requires changing presentation scheduling
- teardown semantics cannot be preserved without a larger lifecycle redesign
- the task grows into unrelated networking cleanup
```

Report:

```text
- changed files
- WebRTC responsibilities moved
- temporary forwarding surface left in ClientConnectionService
- tests changed or added
- tests run
- remaining WebRTC implementation details in ClientConnectionService, if any
```
