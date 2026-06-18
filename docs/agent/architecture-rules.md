# Agent Architecture Rules
Parent index: [Agent](!README.md)

Read this before adding ownership seams, moving packages/folders, changing lifecycle/networking/game-loop responsibilities, or editing known gravity-well files.

## Core Ownership Rules

- Prefer small, explicit ownership seams over broad god files.
- If a change would add a new responsibility to an already-large file, stop and propose the smallest seam or same-package split first.
- Do not add new behavior to gravity-well files unless the prompt explicitly allows it.
- Known gravity-well candidates include broad lifecycle, networking, sync, shell, and game-loop files.
- A seam must have a concrete responsibility.
- Good seams include rooms, scoring, spawning, damage, player lifecycle, codec, domain events, session flow, logging, and presentation controllers.
- Avoid vague buckets like `utils`, `common`, `manager`, or `helpers` unless the responsibility is specific.
- Keep systems self-contained. A system may emit facts/events, expose narrow methods, or accept policy/config, but it should not reach into unrelated systems to make their decisions.
- Do not let integration seams become god objects. Domain events may define, queue, drain, and translate events, but must not decide scoring, damage, spawning, lives, achievements, API persistence, or other gameplay/business rules.
- When adding a feature, first identify the owning system.
- If no obvious owner exists, stop and report the missing seam instead of placing code in the nearest working file.
- Defer mechanics, not ownership. If a near-future feature clearly needs a home, add the minimal owning seam early even if the first behavior remains unchanged.
- Prefer behavior-preserving extraction before behavior change. First move or route existing behavior through the correct seam, then add new behavior in a later prompt.

## File Size Guardrails

Watch file size as an architecture smell, not as an automatic failure.

For hand-written production files:

- Under 200 lines: preferred target for most files.
- 200–350 lines: acceptable if cohesive, but watch for mixed responsibilities.
- 350–500 lines: yellow zone; avoid adding new responsibility.
- 500–800 lines: gravity-well risk; refactor if the file is still actively growing.
- Over 800 lines: active god-object risk unless generated, declarative, or a special case.

Apply these guardrails mainly to:

- `client/scripts/**/*.gd`
- `services/game-server/internal/**/*.go`
- future `services/api-server/app/**/*`

Do not apply these line-count limits directly to:

- generated files
- Godot `.tscn` scene files
- `.tres` resources
- vendored addons
- snapshots
- fixtures
- generated packet/constants files
- large declarative data files

Large scene/resource files should be judged by ownership and editor safety, not line count alone. Large test files are less risky than large production files, but if a test file becomes hard to navigate, prefer splitting by behavior area.

If a file is already above the watch threshold, new prompts should avoid making it larger unless the change belongs there and no smaller owning seam exists.

If a change adds more than roughly 50 lines to one existing production file, stop and check whether a new seam, helper, or same-package split would be cleaner.

If a change adds more than roughly 100 lines total, report why the slice is that large before continuing.

## Go Server Boundaries

- Same-package Go file splits are preferred for reducing god files when no new package boundary is needed.
- New Go packages/folders are architecture decisions and require a clear domain boundary.
- `rooms` owns room creation, joining, leaving, readiness, lifecycle transitions, cleanup policy, and game instance ownership.
- `networking` transports, decodes/routes packets, manages websocket session state, and writes responses.
- `networking` may retain websocket session activation/deactivation when it mutates websocket session fields.
- Room lifecycle policy belongs in rooms.
- Gameplay simulation belongs in game.
- `game` owns authoritative gameplay simulation, gameplay state mutation, and adapters from game storage into narrower gameplay seams.
- Match/mode policy evaluation belongs in `services/game-server/internal/game/rules`, which should receive plain snapshots/facts and return decisions/status.
- `game` should not own websocket transport, API persistence, account/auth concerns, or lobby UI flow.
- Devtools must route through real gameplay seams. Do not create parallel debug-only gameplay logic that bypasses damage, lives, spawning, scoring, movement, room/session, or modifier systems.
- Constants/config should live with the smallest system that owns the decision.
- Do not globalize local presentation defaults unnecessarily.
- Do not bury gameplay, protocol, lifecycle, or environment policy in random files.

## Godot / GDScript Boundaries

- In Godot/GDScript, folder moves are less important than scene wiring.
- Avoid risky scene/node/path/signal changes unless required.
- Prefer extracting pure/helper/controller logic before changing scene ownership.
- Client presentation, UI, audio/effects, local input collection, and interpolation belong in the Godot client.
- Client spectate/view-cycle eligibility must use authoritative lifecycle status (`active`) plus visual availability.
- Do not infer active eligibility solely from remote player positions or ship presence.
- Client menu flow ownership lives in [docs/client/menu-flow.md](../client/menu-flow.md).
- Scene scripts should emit intent and expose display methods.
- Menu scenes should not own API calls, room create/join logic, profile parsing, local profile persistence, or match result row building.
- `MenuFlowController` owns scene routing.
- Feature flows/controllers own behavior.
- Presenters own visibility, label, and disabled state.
- `pregame_menu.gd` is a wiring shell only.

## Packet / Codec Boundaries

- Packet schema source of truth is split under `shared/packets/`:
  - `shared/packets/outputs.toml`
  - `shared/packets/gameplay.toml`
  - `shared/packets/debug.toml`
  - `shared/packets/lobby.toml`
- Do not recreate `shared/packets/packets.toml`.
- Devtools packet schema belongs in `shared/packets/debug.toml`.
- Devtools generated server packet output belongs in `services/game-server/internal/devtools/packets_generated.go`.
- Route server packet wire JSON through `services/game-server/internal/protocol/packetcodec`.
- Do not add direct `encoding/json` calls in server packet wire paths.
- The server codec is intentionally JSON-only and generic.
- Do not add format switching, protobuf references, or an interface unless explicitly requested.
- Non-packet JSON such as collision-shape data-file parsing may still use `encoding/json` directly.
- Route client packet wire JSON through `client/scripts/networking/packets/packet_codec.gd`.
- Do not add direct `JSON.stringify` or `JSON.parse_string` calls in websocket packet paths.
- The client codec is intentionally JSON-only and thin.
- Do not add validation, format switching, typed packet objects, protobuf references, or generator changes unless explicitly requested.
- `network_client.gd` still owns websocket behavior.
- Devtools must route through real gameplay seams. Do not create parallel debug-only gameplay logic that bypasses damage, lives, spawning, scoring, movement, room/session, or modifier systems.

## Domain Event Rules

- Domain gameplay events should be emitted by owning systems and consumed later by achievements, stats, API summaries, logs, or notifications.
- Do not hardwire future consumers into combat/scoring/spawning/lives code.
- Domain events may define, queue, drain, and translate events.
- Domain events must not decide scoring, damage, spawning, lives, achievements, API persistence, or other gameplay/business rules.

## Prompt Scope Rules

- Do not mix unrelated seams in one prompt.
- Codec, rooms, scoring, spawning, health, domain events, devtools, and client lifecycle should be changed in separate slices unless explicitly instructed otherwise.
- If a prompt cannot be completed with a small reviewable diff, stop and report the smallest next prompt instead.
- If implementation touches more than the named lifecycle/system path, stop and report why before continuing.
- Every architecture/refactor prompt must preserve current behavior unless it explicitly says behavior may change.
- Do not add broad cleanup, formatting-only churn, or opportunistic refactors while implementing a seam.
- Before adding code to a large file, check whether an existing seam already owns the behavior.
- If an existing seam owns the behavior, add the behavior there.
- If no seam owns it, propose the seam first.
