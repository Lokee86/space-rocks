# Go Gameplay Seam Skill

Use this skill when changing authoritative gameplay logic on the Go game server.

## When to use

Use this skill for server-side gameplay work involving:

- movement
- toroidal/wrap-aware spatial math
- bullets
- collisions
- damage
- lives
- death
- respawn
- scoring
- spawning
- match/game-over rules
- pause safety
- domain gameplay events
- game simulation state

## Ownership rules

- The Go game server owns authoritative simulation outcomes.
- Keep reusable gameplay simulation in `services/game-server/internal/game`.
- Keep per-entity movement and advance-with-wrap behavior in `services/game-server/internal/game/motion`.
- Keep gameplay distance, direction, and wrap-aware spatial math in `services/game-server/internal/game/space`.
- Keep match/mode policy evaluation in `services/game-server/internal/game/rules`, using plain snapshots/facts and returning decisions/status.
- Keep room creation, joining, leaving, readiness, lifecycle transitions, cleanup policy, and game instance ownership in rooms.
- Networking should transport, decode/route packets, manage websocket session state, and write responses. It should not own gameplay policy.
- API/business concerns belong in the future `services/api-server/`, not the Go game server.

## Seam rules

- First identify the owning system.
- If no owner exists, stop and report the missing seam.
- Prefer behavior-preserving extraction before behavior change.
- Same-package Go file splits are preferred for reducing god files when no new package boundary is needed.
- New packages/folders are architecture decisions and require a clear domain boundary.
- Do not mix unrelated seams in one prompt. Keep codec, rooms, scoring, spawning, health, domain events, devtools, and lifecycle changes separate unless explicitly instructed.
- Devtools must route through real gameplay seams. Do not create parallel debug-only gameplay logic.

## Line-count guardrails

For hand-written Go production files:

- Prefer under about 250 lines.
- 350+ lines requires responsibility scrutiny.
- 500+ lines should be split unless it is cohesive orchestration.
- Do not add a new responsibility to a 350+ line file unless it clearly belongs there.
- For a 500+ line file, prefer routing/extraction/bug fix over new behavior.

Generated files and tests are exempt from strict line-count rules, though tests may still be split for readability.

## Testing rules

For gameplay rule changes, add or update focused Go tests when the prompt asks for test changes.

Server tests live under:

```text
services/game-server/tests/<area>/
```

Do not add new `*_test.go` files beside production packages under `services/game-server/internal/`.

For game simulation setup, use:

```text
services/game-server/tests/game/helpers_test.go
```

Keep new helpers intent-level, such as placing entities or sending packets, instead of exposing raw private maps.

## Human-run verification

Suggested human-run server verification:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

If read-only `envman` warnings appear but tests pass, treat them as warnings, not failures.

Do not run this command by default as the agent.
