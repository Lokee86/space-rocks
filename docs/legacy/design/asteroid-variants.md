# Asteroid Variant Contract

Parent index: [Design Legacy](./!INDEX.md)

## Purpose

The asteroid variant contract defines the shared metadata for asteroid shapes, textures, drops, and spawn weighting. It keeps the server and client aligned on the same variant list without making either side own the contract independently.

## Source Of Truth

- `shared/asteroids/variants.toml`

## Implemented Catalogs

- `services/game-server/internal/game/asteroids/variants.go`
- `client/scripts/generated/asteroids/asteroid_variants.gd`

## Runtime Index Contract

- Variant indexes are zero-based runtime values.
- `index = 0` maps to `asteroid_1` and `asteroid1.png`.
- `index = 7` maps to `asteroid_8` and `asteroid8.png`.
- `index = 8` wraps back to index `0` only through safe lookup helpers.
- Variant `id` values like `asteroid_1` are stable presentation IDs, not runtime indexes.

## Contract Fields

- `id`: stable variant identifier used for presentation and matching.
- `index`: zero-based runtime index used by packets, runtime state, and safe lookup helpers.
- `texture`: client texture path for the asteroid appearance.
- `collision_shape`: collision shape key for the asteroid variant.
- `stats_profile`: named stats profile for per-variant runtime tuning.
- `drop_table`: named drop table for per-variant drop behavior.
- `timed_spawn_weight`: weighted eligibility for timed asteroid spawning.
- `fragment_spawn_weight`: weighted eligibility for asteroid fragment spawning.
- `debug_spawn_weight`: weighted eligibility for debug asteroid spawning.

## Spawn Weight Rules

- A weight greater than `0.0` means the variant is eligible for that spawn source.
- A weight of `0.0` excludes the variant from that spawn source.
- Larger weights make a variant more common.
- Rare variants should use lower nonzero weights instead of booleans.

## Current Contract State

- All 8 current variants have `timed_spawn_weight = 1.0`.
- All 8 current variants have `fragment_spawn_weight = 1.0`.
- All 8 current variants have `debug_spawn_weight = 1.0`.
- All current variants use `collision_shape = "asteroid:0"`.
- All current variants use `stats_profile = "standard"`.
- All current variants use `drop_table = "basicasteroids"`.

## Server Ownership

- Server spawn code must call the asteroid catalog random helpers for variant selection.
- Server spawn code must not use raw `rand.Intn` pools for asteroid variants.
- `constants.AsteroidVariants` is not the owner of variant count anymore.
- The asteroid variant catalog owns the server-side variant list and count behavior.

## Client Ownership

- `client/scripts/entities/asteroid.gd` must use `AsteroidVariants.texture_path_for_index()`.
- Client asteroid rendering must not use hardcoded asteroid texture arrays.
- Client lookup helpers must treat the runtime index as zero-based.

## Extension Rules

Planning notes live in [docs/planning/domain-backlog.md](../planning/domain-backlog.md).

## Verification

- Go test:
  - `cd services/game-server && go test -buildvcs=false ./internal/game/asteroids ./internal/game/spawning ./internal/devtools`
- GUT test:
  - `cd client && godot --headless --path . -s addons/gut/gut_cmdln.gd -gtest=res://tests/unit/entities/test_asteroid_variants.gd`
- Drift checks:
  - no `constants.AsteroidVariants`
  - no `shared/constants` asteroid variants entry
  - no `Variant: rand.Intn(...)` in spawning or devtools
  - no hardcoded `asteroid_textures` array in `asteroid.gd`
  - no index `8` in variant catalogs