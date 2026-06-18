## Profile Flow

Parent index: [Pregame Menu Flow](!README.md)

## Purpose

This document describes the client implementation responsibility for pregame profile context loading and profile readout display.

## Overview

The profile flow displays the active player profile context in the pregame menu transmission panel.

`PregameMenuFlow` routes profile requests to `ProfileFlow.show_profile(current_mode)`. `ProfileFlow` asks `ProfileContextProvider` for the active context for the current pregame mode, asks `ProfileStatsProvider` to load the matching profile data, shapes the display payload, mounts the profile readout scene, and applies the display values.

This flow is separate from local pilot management. Local pilot selection, creation, editing, deletion, and default selection are owned by `LocalPilotFlow`.

## Code root

```text
client/
```

Primary implementation areas:

```text
client/scripts/profile/
client/scripts/ui/menu_flow/
client/scripts/api/
client/scenes/ui/transmission_displays/
```

## Responsibilities

The client profile flow owns:

* handling pregame profile readout requests
* resolving the current profile context by pregame mode
* loading profile data through `ProfileStatsProvider`
* shaping the profile readout display payload
* mounting the profile readout scene in the transmission panel
* applying callsign, activity status, and stat values to the profile readout UI
* falling back to safe empty/zero profile display data when profile data is unavailable

## Does not own

The client profile flow does not own:

* local pilot selector behavior
* local profile create, edit, delete, or default selection
* player-data profile persistence
* Rails account persistence
* auth session creation or logout
* profile stat mutation
* match result persistence
* progression, achievements, rewards, unlocks, or inventory
* full cross-system identity policy

## Domain roles

The profile flow participates in the pregame identity/profile readout portion of the player experience domain.

Its domain role is client-side presentation:

* single-player Guest contexts display as active Guest profile contexts
* selected single-player local profile contexts display as active local profile contexts
* multiplayer signed-in contexts display authenticated account identity
* multiplayer signed-out contexts fall back to offline Guest
* profile readout displays current stats but does not mutate them

The player-data service remains the persistence and profile data boundary.

## Protocols and APIs

The profile flow uses the player-data profile endpoint through `PlayerDataProfileApiClient`.

Client API wrapper:

```text
client/scripts/profile/player_data_profile_api_client.gd
```

Configured endpoint path:

```text
POST /api/player-data/profile
```

Request fields:

```text
play_mode
identity_kind
local_profile_id
```

Authenticated account requests also pass the active session token.

`ProfileFlow` does not call the endpoint directly. It delegates profile loading to `ProfileStatsProvider`.

## Data ownership

The client owns only display shaping and transient context selection.

Client-owned display payload:

```text
callsign
activity_status
total_score
high_score
games_played
wins
ship_deaths
```

Player-data-owned profile data:

```text
profile identity
profile callsign
profile activity status
profile stats
```

`ProfileStatsProvider` normalizes stats to the client-facing shape:

```text
total_score
high_score
ship_deaths
games_played
wins
```

Extra sensitive or unrelated fields from API responses are ignored by the client-side normalization path.

When profile data is unavailable:

* Guest and local profile paths return zeroed profile data.
* Authenticated account paths can return cached normalized stats after a successful prior load.
* Authenticated account paths return zeroed stats when no token and no cache are available.

## Code map

Primary implementation files:

```text
client/scripts/ui/menu_flow/pregame_menu_flow.gd
client/scripts/profile/profile_flow.gd
client/scripts/profile/profile_context_provider.gd
client/scripts/profile/profile_stats_provider.gd
client/scripts/profile/player_data_profile_api_client.gd
client/scripts/profile/profile_readout.gd
client/scripts/profile/profile_identity_kind.gd
client/scripts/api/api_config.gd
```

Primary scene:

```text
client/scenes/ui/transmission_displays/profile_readout.tscn
```

Related tests:

```text
client/tests/unit/profile/test_profile_context_provider.gd
client/tests/unit/profile/test_profile_stats_provider.gd
```

Important non-ownership boundaries:

```text
client/scripts/ui/menu_flow/local_pilot_flow.gd
services/player-data/
services/api-server/
docs/services/player-data/
docs/services/api-server/
docs/data/
docs/protocol/
```

## Tests

Relevant test coverage:

```text
client/tests/unit/profile/test_profile_context_provider.gd
client/tests/unit/profile/test_profile_stats_provider.gd
```

Covered context behavior includes:

* single-player Guest context
* selected local profile context
* signed-in authenticated account context
* authenticated empty display-name fallback
* signed-out multiplayer Guest fallback

Covered stats behavior includes:

* Guest profile loading
* local profile loading with `local_profile_id`
* authenticated profile loading with token
* stat normalization
* sensitive field exclusion
* missing-stat zero fallback
* authenticated cached-stat fallback after API failure

## Related docs

* [pregame-menu-flow](!README.md)
* [local-pilot-flow.md](local-pilot-flow.md)
* [Client](../!README.md)
* [Services](../../!README.md)

## Notes

This document covers client pregame profile readout implementation only. It intentionally leaves local profile management to `local-pilot-flow.md` and persistence internals to player-data service documentation.
