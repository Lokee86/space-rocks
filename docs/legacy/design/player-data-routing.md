# Player-Data Routing Reference

Parent index: [Design Legacy](./!README.md)

This is the compact reference for player-data data-handler routing.
Use it before touching profile reads, match-result writes, or player-data store selection.

For the full routing architecture, see [Cross-Mode Routing And Player Data](cross-mode-routing-and-player-data.md).
For logical schema ownership, see [Player-Data Schema Source Of Truth](player-data-schema-ssot.md).
For contract ownership, see [Source Of Truth Map](source-of-truth-map.md).
For HTTP shapes, see [HTTP Contracts](../api/http-contracts.md).

Local-profile durable storage is present in the standard no-tag development build.
`-tags noembeddedsqlite` deployment/restricted builds exclude `playerdata/embeddedsqlite` and the `modernc.org/sqlite` dependency.
In `-tags noembeddedsqlite` builds, local profile management returns `local_profiles_unavailable`.

## Core Rule

All player-data reads and writes route through the game-server data-handler and the in-process player-data runtime.

The client does not choose Rails, SQLite, or guest memory directly.
The player-data runtime owns identity-based store selection.
The core playerdata package receives local-store construction through dependency injection and does not import the embedded SQLite package.

## Write Flow

Match-result writes follow this path:

1. The game-server resolves authoritative match facts at game over.
2. The game-server sends `RecordMatchResult` into the player-data runtime.
3. The player-data runtime validates mode and identity.
4. The runtime routes the write to the selected store.

Store outcomes:

- `guest` -> in-memory transient stats
- `local_profile` -> SQLite in the standard no-tag development build, unavailable in `-tags noembeddedsqlite` builds
- `authenticated_account` -> Rails/Postgres through `RailsStore`

`ship_deaths` comes from authoritative server match facts, not client-side counting.

## Local Profile Management Flow

Local profile list, create, and default requests enter through the game-server data-handler.
The player-data runtime/store seam owns local-profile persistence only when the embedded SQLite build is present.
The client never writes SQLite directly.

CREATE routing:

1. The client calls the data-handler create route.
2. The game-server data-handler validates the request and forwards it to the player-data runtime.
3. The runtime creates the local profile in SQLite.
4. Guest stats are copied only when `seed_from_guest_stats` is true.

LOAD/default routing:

1. The client persists the selected default through the data-handler.
2. The game-server data-handler forwards the default request to the player-data runtime.
3. The runtime stores the default by identity kind and `local_profile_id`, not by display name.
4. `display_name` remains presentation only.

## Read Flow

Profile reads follow this path:

1. The client sends `POST /api/player-data/profile` to the game-server.
2. The game-server data-handler validates the request, mode, and identity.
3. The game-server forwards the request to the player-data runtime.
4. The runtime calls `LoadStats`.
5. The runtime routes the read to the selected store.
6. The game-server returns a normalized profile payload.

Store outcomes:

- `guest` -> in-memory transient stats
- `local_profile` -> SQLite in the standard no-tag development build, unavailable in `-tags noembeddedsqlite` builds
- `authenticated_account` -> Rails/Postgres through `RailsStore`

The same logical stats payload is normalized for display regardless of backing store.

## Token Roles

- Authenticated profile requests use user bearer tokens to prove identity to the game-server profile endpoint; guest reads remain unauthenticated.
- Internal bearer tokens authorize service-to-service Rails calls.
- Do not require a static user bearer token environment variable such as `PLAYER_DATA_RAILS_BEARER_TOKEN`.

The game-server verifies the user token before it chooses the authenticated-account store route.
`RailsStore` uses the internal token for Rails calls such as stats reads and match-result writes.

## SSoT Rules

- `shared/contracts/http/openapi.yaml` owns HTTP request/response shapes.
- `shared/packets/player_data.toml` owns the runtime packet shapes.
- `shared/player_data/*.toml` owns the logical player-data schema only.

Keep those boundaries separate:

- OpenAPI is for HTTP contracts.
- Player-data packets are for runtime transport.
- Logical schema is for the shared player-data model.

## Forbidden Bypasses

- Do not route profile readout directly to Rails `/api/player/stats`.
- Do not mutate guest stats from client-side presentation code.
- Do not let the game-server read or write databases directly.
- Do not make `RailsStore.LoadStats` use `/api/player/stats` for data-handler profile reads.
- Do not let the client choose the backing store.
- Do not add a direct client Rails stats path for profile readout.

## Identity Routing

The runtime routes by identity kind after mode and identity validation.

| Identity | Read route | Write route | Backing store |
| --- | --- | --- | --- |
| Guest | transient read | transient write | in-memory stats |
| Local Profile | durable read | durable write | SQLite in the standard no-tag development build |
| Authenticated Account | durable read | durable write | Rails/Postgres |

`account_id` is the authenticated-account UUID identity.
`local_profile_id` is the local-profile identity.
Rails `user_id` stays an internal database foreign key.

## Practical Check

When in doubt, ask:

- Is this a read or a write?
- Does it enter through the game-server data-handler?
- Is the identity guest, local profile, or authenticated account?
- Is the store chosen by the runtime, not by the client?

If any answer is no, the routing is probably off-seam.