---
author: brian
created: "2026-08-04"
document_id: 4e2d7b2e-9334-431e-bc57-6114fd1ed6b6
document_type: general
policy_exempt: false
summary: Maps critical Space Rocks behavioral invariants to focused tests, runtime scenarios, and release gates.
---
# Behavioral-Contract Matrix

Parent index: [Development](./!INDEX.md)

## Purpose

This document maps critical behavioral invariants to the tests and release gates that protect them.

## Overview

Static architecture and documentation checks cannot prove runtime behavior. This matrix identifies the focused behavioral evidence expected to fail when a critical ownership, lifecycle, persistence, protocol, or recovery invariant regresses.

## Contracts

| Critical invariant | Primary evidence | Additional gate |
| --- | --- | --- |
| The Go game server is authoritative for gameplay outcomes; client state application does not create authority | `services/game-server/tests/game/`, client gameplay-state tests | Server/client integration scenarios |
| Seeded gameplay execution is deterministic for identical inputs | `services/game-server/tests/game/determinism_test.go`, `internal/game/game_seed_test.go`, RNG tests | Game-server test suite |
| Toroidal distance, motion, visibility, and collision use wrap-aware space rules | `services/game-server/tests/space/`, game visibility/collision tests | Client view-anchor tests |
| Durable player counters remain in player-session state while ships are replaceable active avatars | Player session, counter, death, despawn, and respawn tests | Match-result and room snapshot tests |
| Damage requests resolve through the damage owner before death and presentation effects | Damage package tests, collision-to-damage tests, radial-effect tests | Game-server test suite |
| Room membership, match start, cleanup, and result reporting preserve lifecycle ownership | `services/game-server/internal/rooms/*_test.go`, networking room tests | Runtime scenarios |
| WebSocket/WebRTC admission binds authenticated identity without reading Rails tables directly | Networking auth tests and `internal/authclient` tests | Rails internal auth controller tests |
| Realtime baseline, resync, lane ordering, chunking, prioritization, and quantization remain compatible | `internal/protocol/realtime/*_test.go`, shared fixtures, client realtime tests | Data-sync realtime-wire check |
| Hot lanes never become authoritative for entity existence | Lifecycle/hot-lane server tests and client lifecycle gate tests | Protocol fixture tests |
| Generated constants and packet bindings match source-of-truth files | `tools/data_sync/tests/`, generated-file checks | Repository CI data-sync checks |
| Rails stores only bearer-token digests and enforces expiry/revocation | API-server auth model/service/controller tests | OpenAPI contract tests |
| Player-data routing keeps guest, local, and account-backed persistence behind explicit stores | `services/player-data/playerdata/*_test.go`, HTTP API tests | Game-server player-data integration tests |
| Local-profile SQLite persistence remains isolated from authenticated Rails persistence | Embedded SQLite and store-router tests | Packaged local-alpha gate |
| Diagnostic submissions reject unsafe or invalid material before durable storage | Diagnostic API, validation, redaction, and report tests | Bruno diagnostic smoke tests |
| Diagnostic JSONL storage recovers active files and enforces rotation, quarantine, and retention | `services/diagnostic-aggregator/internal/storage/jsonlstore/*_test.go` | Standalone/hosted service tests |
| Service logging failures do not alter gameplay, HTTP, or persistence outcomes | Service observability/logging tests | Runtime scenarios and log inspection |
| Production images build from repository-owned Dockerfiles and the Compose payload resolves before publication | `publish-hosted-images.yml` Compose validation | `deploy/production/verify.py` after deployment |
| Client package presets and environment selection do not rewrite development source values | Client configuration and export tests | Local and multiplayer release gates |
| Devtools mutate real game seams rather than maintaining parallel debug gameplay systems | Game control and devtools controller tests | Pitlord devtools dependency rules |
| Documentation indexes, links, required owners, coverage, and change-impact mappings remain valid | `.standards/docs_policy/check.py` | Pitlord documentation policy and repository CI |

## Maintenance rules

Update this matrix when a critical invariant changes, a protecting test moves, a new recovery seam is introduced, or a release gate becomes the primary evidence. A row should point to focused behavior, not merely a broad test command.

## Related docs

- [Documentation coverage](documentation-coverage.md)
- [Maintainer map](../maintainer-map.md)
- [Verification and quality gates planning](../planning/domains/technical/verification-and-quality-gates.md)
- [Developer guide](../developer.md)

## Notes

This matrix does not claim that every test listed is sufficient in isolation. Cross-service and deployed behavior still requires runtime scenarios and release verification.
