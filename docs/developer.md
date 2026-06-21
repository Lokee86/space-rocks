## Developer Onboarding

Parent index: [Documentation](./!INDEX.md)

## Purpose

This document is the Space Rocks developer onboarding and handoff guide.

It gives new or returning contributors enough context to set up the repo, understand the local development tools, run the main services, verify changes, and find the canonical documentation that owns implementation details.

This document is not the source of truth for detailed system behavior. Detailed facts belong in the owning service, protocol, data, devtools, domain, systems-design, planning, or limits docs.

## Overview

Space Rocks is an Asteroids-inspired game project with these major parts:

```text
client/                 Godot client project
services/game-server/   Go realtime game server
services/api-server/    Ruby/Rails API server
shared/                 Shared source-of-truth data
tools/data_sync/        Constants, packet, and generated-output sync tooling
bruno-api/              Bruno collection for local API smoke tests
docs/                   Project documentation
```

The current development model is server-authoritative for gameplay state. The client collects input and presents world state. The game server owns authoritative gameplay outcomes. Backend/account and player-data concerns are split into service-specific docs.

Use this onboarding guide to get productive. Use the documentation map below to find the current authority for detailed behavior.

## Basic Setup

Clone the repository and enter the repo root.

```bash
git clone <repo-url>
cd space-rocks
```

Install Git LFS before opening or running the project.

```bash
git lfs install
git lfs pull
```

Install the local development runtimes used by the repo:

```text
Godot 4.6.x       client project and GUT tests
Go 1.26.x         realtime game server
Ruby / Rails      API server
Python 3.10+      repo tooling and checks
Git LFS           binary/source asset files
```

Install Python development dependencies from the repo root:

```bash
python -m pip install -r requirements-dev.txt
```

Open the Godot project from:

```text
client/
```

The configured main scene is:

```text
res://scenes/game.tscn
```

The repo path used in many local commands is:

```text
/mnt/d/!bin/space-rocks
```

When writing shell commands for that path, escape the exclamation mark:

```bash
cd /mnt/d/\!bin/space-rocks
```

## Repository Shape

`client/` contains the active Godot client project: scenes, scripts, assets, tests, generated client packet helpers, and generated constants.

`services/game-server/` contains the Go realtime game server. It owns the live gameplay simulation, websocket handling, room lifecycle, authoritative state snapshots, and server-side gameplay decisions.

`services/api-server/` contains the Rails API server. It owns account/auth and backend HTTP behavior that belongs outside the realtime game server.

`shared/` contains shared source files used by generators and multiple runtimes, including constants, packet schemas, collision data, and related source-of-truth material.

`tools/data_sync/` contains the data-sync tooling used to validate, diff, and regenerate shared outputs.

`bruno-api/` contains the Bruno collection used for local API smoke testing.

`docs/` contains current, planning, limits, agent, and legacy documentation. Current documentation uses `!INDEX.md` files as folder indexes.

`SourceAssets/` contains local source art material and is ignored by Git.

## Development Tools

### data-sync

`data-sync` is the repo generation and synchronization tool for shared data.

It is used for packet schemas, constants, generated Go outputs, generated GDScript outputs, and related shared-source validation.

The implementation lives under:

```text
tools/data_sync/
```

Use data-sync when changing shared packet or constants sources. Do not hand-edit generated outputs as the source of truth.

Common validation commands:

```bash
data-sync -check -packets -go -gds
data-sync -check -constants -go -gds
```

Common diff commands:

```bash
data-sync -diff -packets -go -gds
data-sync -diff -constants -go -gds
```

Common regeneration commands:

```bash
data-sync -push -packets -go -gds
data-sync -push -constants -go -gds
```

Use `data-sync` for generated outputs. Do not use it as a replacement for understanding ownership. The owning data docs explain the source files and generated outputs.

### doc-ledger

`doc-ledger` maintains the generated sections in documentation indexes.

It owns the marked blocks inside `!INDEX.md` files, such as:

```markdown
<!-- doc-ledger:files:start -->
<!-- doc-ledger:files:end -->

<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->

<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
```

Use doc-ledger when documentation files or folders are added, moved, renamed, graduated from stubs, or removed.

Do not manually fight generated index blocks. Manual index text belongs outside the doc-ledger marker blocks. If an index is wrong, fix the underlying file/folder placement, doc-ledger configuration, or generated description source rather than hand-editing only the generated block.

Doc-ledger expects normal documentation folders to use:

```text
!INDEX.md
```

`stubs/` folders are exempt from requiring their own `!INDEX.md`.

### Bruno

Bruno is the local API smoke-test tool for the Rails API server.

The collection root is:

```text
bruno-api/
```

Use Bruno to exercise real local API-server routes during development. Bruno does not replace Rails tests, does not own HTTP contracts, does not bypass application behavior, and must not contain committed secrets.

The local Bruno environment is:

```text
bruno-api/environments/local.yml
```

Typical smoke order:

```text
Health
Register or Login
Me
Logout
Me should fail with the same token after logout
```

For API auth and Bruno details, use:

```text
docs/devtools/api-server/bruno-smoke-tests.md
docs/devtools/api-server/local-auth-smoke-flow.md
```

Keep real Discord secrets, bearer tokens, callback codes, and local credentials out of committed Bruno files.

### Git LFS

Git LFS is required for binary and source asset files.

The repo tracks asset patterns such as images and audio through Git LFS. Run LFS setup after cloning:

```bash
git lfs install
git lfs pull
```

Do not assume assets are valid if Git LFS has not been pulled. Broken or pointer-only assets can cause confusing Godot import and runtime behavior.

### Godot And GUT

Godot is the client editor and runtime.

Open the project at:

```text
client/
```

Run the editor from the repo root if the `godot` command is available:

```bash
godot --path client
```

GUT is the Godot unit test framework used by the client tests.

Run client unit tests with:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

A passing GUT run can still print Godot cleanup or ObjectDB warnings. Treat the run as passing when GUT reports that all tests passed.

Use Godot and the EngineForge bridge to inspect scene and UI state before guessing about node paths, tree shape, scene ownership, or editor state.

### Go

Go is used for the realtime game server.

The module root is:

```text
services/game-server/
```

Run the server from the repo root with:

```bash
cd /mnt/d/\!bin/space-rocks
{ (cd services/game-server && go run ./cmd/game-server); }
```

Run all game-server tests with:

```bash
cd /mnt/d/\!bin/space-rocks
{ (cd services/game-server && go test -buildvcs=false ./...); } 2>&1 | tee /dev/tty | clip.exe
```

Use an explicit Go cache if the shell environment has cache or permission problems:

```bash
cd /mnt/d/\!bin/space-rocks
{ (cd services/game-server && env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...); } 2>&1 | tee /dev/tty | clip.exe
```

Do not put broad test catalogs in this onboarding doc. Use agent testing guidance and service docs for detailed test placement rules.

### Rails

Rails is used for the API server.

The service root is:

```text
services/api-server/
```

Start from:

```bash
cd services/api-server
set -a && source ../../.env && set +a
bundle install
bundle exec rails db:create
bundle exec rails db:migrate
bundle exec rails test
bundle exec rails server
```

The API server listens on port `3000` by default.

Use Rails tests for automated API verification. Use Bruno or curl for local smoke checks.

Do not require Rails for local single-player client work unless the work specifically touches auth, account, backend player-data persistence, or API-server flows.

### Python Checks

Python supports repo tooling and static checks.

Install development dependencies from the repo root:

```bash
python -m pip install -r requirements-dev.txt
```

Run the client constants-boundary scan with:

```bash
python3 -m pytest tools/tests/test_client_constants_boundary.py
```

Use Python tooling for repo validation and generation support. Do not use ad hoc Python scripts to mutate source-of-truth data when an existing repo tool owns the workflow.

### Air

Air is used for optional Go hot reload when installed.

Run it from the game-server service root:

```bash
cd /mnt/d/\!bin/space-rocks
{ (cd services/game-server && air); }
```

Use Air for local convenience only. A normal `go run` or `go test` path should still work without Air.

## MCP Servers And Godot Bridge

Space Rocks uses local MCP servers and the EngineForge/Godot bridge to support planning, inspection, and bounded implementation workflows.

There are two MCP server roles:

```text
Info MCP    read/search repo plus read-only Godot bridge diagnostics
Write MCP   bounded repo writes, allowlisted commands, and Godot bridge mutations
```

The normal local ports are:

```text
Info MCP    8789
Write MCP   8788
```

Use Info MCP for planning, read-only repo inspection, Godot project info, scene-tree inspection, editor logs, and bridge diagnostics.

Use Write MCP only for intentional implementation work. Keep it local. Do not expose Write MCP through ngrok or other public tunnels.

The EngineForge/Godot bridge runs inside the local Godot project and exposes editor/project/scene capabilities to the MCP servers.

Do not guess bridge command names. The bridge command set comes from the installed plugin capabilities. Read-only bridge diagnostics should use the available MCP tools before making assumptions about Godot scenes, node paths, or editor state.

For detailed MCP server usage, bridge command shape, startup commands, and troubleshooting, use:

```text
docs/agent/mcp-servers.md
```

## Running The Project

### Run The Game Server

From the repo root:

```bash
cd /mnt/d/\!bin/space-rocks
{ (cd services/game-server && set -a && source ../../.env && set +a && go run ./cmd/game-server); }
```

The local game server listens on port `8080`.

Normal local runs include devtools.

To run with server devtools disabled:

```bash
cd /mnt/d/\!bin/space-rocks
{ (cd services/game-server && go run -tags nodevtools ./cmd/game-server); }
```

### Run The API Server

From the repo root:

```bash
cd /mnt/d/\!bin/space-rocks
{ (cd services/api-server && set -a && source ../../.env && set +a && bundle exec rails server); }
```

The API server listens on port `3000` by default.

Run migrations and tests before relying on the API server after schema or auth changes:

```bash
cd services/api-server
set -a && source ../../.env && set +a
bundle exec rails db:migrate
RAILS_ENV=test bundle exec rails db:test:prepare
bundle exec rails test
```

### Open The Godot Client

Open or import:

```text
client/
```

Or launch with:

```bash
godot --path client
```

The client expects the Go game server to be running for gameplay flows.

## Verifying Changes

Use the smallest verification path that covers the files changed.

For Go game-server changes:

```bash
cd /mnt/d/\!bin/space-rocks
{ (cd services/game-server && go test -buildvcs=false ./...); } 2>&1 | tee /dev/tty | clip.exe
```

For Godot client unit tests:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

For Rails API changes:

```bash
cd services/api-server
set -a && source ../../.env && set +a
bundle exec rails test
```

For packet or constants generation checks:

```bash
data-sync -check -packets -go -gds
data-sync -check -constants -go -gds
```

For generated-output diffs before pushing:

```bash
data-sync -diff -packets -go -gds
data-sync -diff -constants -go -gds
```

For client constants-boundary checks:

```bash
python3 -m pytest tools/tests/test_client_constants_boundary.py
```

For manual API smoke checks, use Bruno from:

```text
bruno-api/
```

Use the detailed testing rules in:

```text
docs/agent/testing.md
```

## Documentation Map

Use this section to find the owning documentation area. Do not duplicate detailed system facts in this onboarding doc.

* [Agent](agent/!INDEX.md) - Agent workflow guidance, testing expectations, MCP usage, session memory, and implementation guardrails.
* [Data](data/!INDEX.md) - Source-of-truth files, generated outputs, schemas, persistence contracts, and data-sync pipeline documentation.
* [Devtools](devtools/!INDEX.md) - Debug and development tooling, including API-server smoke tooling, client/server devtools, telemetry, controls, and debug-only boundaries.
* [Domains](domains/!INDEX.md) - Cross-system player, platform, and technical flows. Domain docs explain how multiple services participate in larger project behavior.
* [Limits](limits/!INDEX.md) - Temporary blockers, known bugs, dev-blocked issues, active gaps, and transitional limitations.
* [Planning](planning/!INDEX.md) - Future, unresolved, proposed, or not-yet-current work. Planning docs should not be treated as current implementation authority.
* [Protocol](protocol/!INDEX.md) - HTTP, WebSocket, packet, and message-flow contracts between systems.
* [Services](services/!INDEX.md) - Runtime and implementation documentation for the client, game server, API server, player-data service, and web service.
* [Systems Design](systems-design/!INDEX.md) - Conceptual mechanics, authority boundaries, durable invariants, and design rules.
* [Documentation policy](documentation-policy.md) - Rules for where documentation belongs and how documentation types are classified.
* [Documentation procedure](documentation-procedure.md) - Workflow for creating, updating, moving, graduating, and deleting docs.

## Source-Of-Truth Rules

Generated files are not the source of truth.

Packet fields belong in:

```text
shared/packets/*.toml
```

Packet output routing belongs in:

```text
shared/packets/outputs.toml
```

Constants belong in:

```text
shared/constants/
```

Collision shapes come from the Godot export pipeline and shared collision data.

Use data-sync to check, diff, and regenerate generated outputs.

Protocol docs own message and API contracts.

Service docs own implementation behavior.

Systems-design docs own durable authority rules, invariants, and conceptual mechanics.

Data docs own schemas, persistence contracts, source files, generated outputs, and pipeline behavior.

Devtools docs own debug-only tooling and smoke-test workflows.

Planning docs do not own implemented facts. When planned work becomes current, rewrite or move current facts into the owning current docs.

Limits docs own temporary blockers, known bugs, and transitional issues. They do not own permanent design constraints.

## Development Cautions

Do not hand-edit generated packet or constants outputs as source-of-truth changes.

Do not let the client locally mutate authoritative gameplay state for normal gameplay behavior. The client should present state and send input or request packets; the server owns authoritative outcomes.

Do not use planning docs as current implementation authority.

Do not treat legacy docs as current authority. Legacy docs are migration source material only.

Do not expose the Write MCP server remotely.

Do not guess Godot node paths, scene trees, or bridge command names when the MCP/Godot bridge can inspect the actual project state.

Do not commit real secrets, bearer tokens, Discord OAuth secrets, callback codes, or local credentials.

Do not require Rails for single-player work unless the change explicitly touches account, auth, backend persistence, or API-server integration.

Keep `!INDEX.md` files current when documentation files or folders move. Normal folder indexes use `!INDEX.md`; documentation policy and procedure own the rules and workflow.

## Handoff Notes

Use this section only for short, high-signal warnings that help a returning developer avoid immediate mistakes.

Current handoff points:

* Space Rocks is in active development. Expect rough edges around newer systems.
* The gameplay model is server-authoritative. Preserve authority boundaries when changing client presentation.
* Single-player should remain usable without Rails unless the work explicitly touches backend/account behavior.
* Generated packet and constants files should be regenerated through data-sync, not hand-edited as the source.
* Godot scene and UI work should be inspected through the editor or bridge before changing node paths or assumptions.
* Documentation should route detailed facts to the owning docs section instead of making this onboarding guide a second system reference.
* `docs/!INDEX.md` indexes docs. `documentation-policy.md` defines documentation rules. `documentation-procedure.md` defines documentation workflow.

## Related docs

* [Documentation index](./!INDEX.md)
* [Documentation policy](documentation-policy.md)
* [Documentation procedure](documentation-procedure.md)
* [Agent docs](agent/!INDEX.md)
* [MCP servers](agent/mcp-servers.md)
* [Testing](agent/testing.md)
* [Data docs](data/!INDEX.md)
* [Source-of-truth map](data/source-of-truth-map.md)
* [Data sync and source-of-truth pipeline](data/data-sync-and-ssot-pipeline.md)
* [Devtools docs](devtools/!INDEX.md)
* [API-server Bruno smoke tests](devtools/api-server/bruno-smoke-tests.md)
* [Services docs](services/!INDEX.md)
* [Protocol docs](protocol/!INDEX.md)
* [Systems-design docs](systems-design/!INDEX.md)
* [Limits docs](limits/!INDEX.md)
* [Planning docs](planning/!INDEX.md)

## Notes

This document intentionally repeats only the basic setup and developer workflow information needed for onboarding.

When a detail becomes large enough to explain implementation behavior, move it to the owning service, protocol, data, devtools, domain, systems-design, limits, or planning doc and link to it from here.
