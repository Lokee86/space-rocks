## Local Pilot Flow

Parent index: [Pregame Menu Flow](./!README.md)

## Purpose

This document describes the client implementation responsibility for single-player local pilot selection and local profile management in the pregame menu.

## Overview

The local pilot flow lets the player choose whether single-player runs as Guest or as a saved local profile.

`PregameMenuFlow` owns the outer pregame menu routing. It creates `LocalPilotFlow`, passes in the transmission flow, passes the pregame callsign update callback, and routes pilot-selection requests only while the pregame menu is in single-player mode.

`LocalPilotFlow` owns the selector flow after that request is accepted. It mounts the local pilot selector, connects selector intent signals, calls the player-data local profile API, updates the active single-player profile context, refreshes selector state, and updates the visible callsign.

The local pilot UI scenes emit intent only. They do not own API calls, persistence, or identity policy.
The selector helper components handle list scrolling, row focus, row styling, and row selection feedback inside the local pilot selection UI.

## Code root

```text
client/
```

Primary implementation areas:

```text
client/scripts/ui/menu_flow/
client/scripts/ui/local_pilots/
client/scripts/ui/menus/elements/
client/scripts/profile/
client/scripts/api/
client/scenes/ui/transmission_displays/
client/scenes/ui/transmission_displays/sub-transmissions/
client/scenes/ui/elements/
```

## Responsibilities

The client local pilot flow owns:

* opening the local pilot selector from the single-player pregame menu
* mounting selector and subpanel scenes through the transmission flow
* wiring selector load, create, edit, and delete intent
* applying the saved default profile on single-player entry
* falling back to Guest when the saved default is unavailable or invalid
* persisting selected Guest/local-profile defaults through player-data
* creating local profiles
* updating local profile display names
* deleting local profiles
* refreshing the selector after successful mutations
* updating `ProfileContextProvider` for single-player identity
* updating the pregame callsign indicator

## Does not own

The client local pilot flow does not own:

* local profile persistence internals
* player-data schema or storage decisions
* authenticated account identity
* multiplayer account/session state
* profile stat aggregation
* match-result persistence
* progression, achievements, rewards, unlocks, or inventory
* broad player identity domain policy

## Domain roles

The local pilot flow participates in the player identity/profile portion of the player experience domain.

Its domain role is presentation and client-side selection state:

* Guest is always available as the single-player fallback.
* Local profile identity is represented by `local_profile_id`.
* Display names are callsign presentation data, not durable identity.
* Single-player profile context is selected before starting a local match.

The authoritative persistence boundary for local profiles remains the player-data service.

## Protocols and APIs

The local pilot flow uses client HTTP helpers to call player-data local profile endpoints.

Client API wrapper:

```text
client/scripts/profile/local_pilot_api_client.gd
```

Configured endpoint paths:

```text
GET    /api/player-data/local-profiles
GET    /api/player-data/local-profiles/default
POST   /api/player-data/local-profiles
PUT    /api/player-data/local-profiles/{local_profile_id}
DELETE /api/player-data/local-profiles/{local_profile_id}
PUT    /api/player-data/local-profiles/default
```

Runtime usage:

* list profiles for selector population
* get default profile when entering single-player
* create profile from a callsign
* update profile display name
* delete profile
* set Guest or local profile as default

The selector and subpanel scene scripts do not call these APIs directly.

## Data ownership

The client does not persist local profiles.

Client-owned state:

* selected single-player identity kind
* selected local profile id
* selected display name/callsign
* mounted selector/subpanel UI state

Player-data-owned state:

* local profile records
* local profile stats
* default local profile selection
* Guest/local profile persistence behavior

The client sends `seed_from_guest_stats` when creating local profiles. It sets this to true only when the active single-player identity is Guest.

## Code map

Primary implementation files:

```text
client/scripts/ui/menu_flow/pregame_menu_flow.gd
client/scripts/ui/menu_flow/local_pilot_flow.gd
client/scripts/profile/profile_context_provider.gd
client/scripts/profile/local_pilot_api_client.gd
client/scripts/profile/profile_identity_kind.gd
client/scripts/api/api_config.gd
```

Primary scene scripts:

```text
client/scripts/ui/local_pilots/select_pilot_readout.gd
client/scripts/ui/local_pilots/enter_pilot_id.gd
client/scripts/ui/local_pilots/confirm_delete.gd
client/scripts/ui/menus/elements/discrete_list_view.gd
client/scripts/ui/menus/elements/pilot_select_row.gd
```

Primary scenes:

```text
client/scenes/ui/transmission_displays/select_pilot_readout.tscn
client/scenes/ui/transmission_displays/sub-transmissions/enter_pilot_id.tscn
client/scenes/ui/transmission_displays/sub-transmissions/confirm_delete.tscn
client/scenes/ui/elements/discrete_list_view.tscn
client/scenes/ui/elements/pilot_select_row.tscn
```

Related tests:

```text
client/tests/unit/profile/test_profile_context_provider.gd
```

Important non-ownership boundaries:

```text
services/player-data/
docs/services/player-data/
docs/data/
docs/protocol/
```

## Tests

Relevant test coverage:

```text
client/tests/unit/profile/test_profile_context_provider.gd
```

Covered behavior includes:

* single-player defaults to active Guest
* selecting a local profile returns active local profile context
* multiplayer signed-in context uses authenticated account identity
* multiplayer signed-out context falls back to offline Guest

The selector and subpanel UI behavior should be covered by focused UI/unit tests when those flows are changed.

## Related docs

* [pregame-menu-flow](./!README.md)
* [profile-flow.md](profile-flow.md)
* [Client](../!README.md)
* [Services](../../!README.md)

## Notes

This document covers client local pilot selection and local profile management in the canonical pregame menu flow.
