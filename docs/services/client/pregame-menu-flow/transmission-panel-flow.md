## Transmission Panel Flow

Parent index: [Pregame Menu Flow](./!README.md)

## Purpose

This document describes the client implementation responsibility for the pregame transmission panel flow.

## Overview

The transmission panel flow owns how the pregame menu mounts and clears transmission content in the primary screen and the subpanel screen.

`TransmissionFlow` mounts primary transmissions into the primary screen display, mounts subpanel transmissions into the subpanel screen display, and keeps the active transmission references in sync with what is shown.

When a subpanel is active, the primary transmission input is locked so the primary screen does not continue accepting focus or pointer interaction underneath the subpanel.

Clearing behavior is split between primary and subpanel state. Clearing the primary transmission removes primary content and resets the active primary state. Clearing the subpanel removes subpanel content and releases the primary input lock state that was in effect while the subpanel was active.

Back routing follows the active transmission state already mounted in the panel. The current transmission flow keeps the active transmission references aligned with the mounted primary or subpanel content so the pregame menu can route Back behavior through the active panel state.

## Code root

```text
client/
```

Primary implementation areas:

```text
client/scripts/ui/menu_flow/
client/scenes/ui/elements/windows/
```

## Responsibilities

The client transmission panel flow owns:

* mounting primary transmissions into the primary screen display
* mounting subpanel transmissions into the subpanel screen display
* keeping active transmission references aligned with mounted panel content
* locking primary input while a subpanel is active
* clearing primary transmission content and active primary state
* clearing subpanel transmission content and releasing primary input locks
* routing Back behavior through the currently active transmission state

## Does not own

The client transmission panel flow does not own:

* local pilot selection policy
* profile readout shaping
* input action definitions
* menu navigation policy beyond active transmission routing
* server state
* gameplay runtime

## Code map

Primary implementation files:

```text
client/scripts/ui/menu_flow/transmission_flow.gd
```

Primary scenes:

```text
client/scenes/ui/elements/windows/transmission_screen.tscn
client/scenes/ui/elements/windows/transmission_screen.2.tscn
```

## Tests

No focused test is documented yet for this flow.

The closest verification boundary is the transmission flow implementation in:

```text
client/scripts/ui/menu_flow/transmission_flow.gd
```

## Related docs

* [Pregame Menu Flow](./!README.md)
* [Local Pilot Flow](local-pilot-flow.md)
* [Profile Flow](profile-flow.md)
* [Client](../!README.md)
* [Services](../../!README.md)

## Notes

This document captures current transmission-panel behavior only. It does not describe future panel routing or procedural menu setup steps.
