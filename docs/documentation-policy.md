## Documentation Policies
Parent index: [Documentation](./!README.md)

## Purpose

This document defines the documentation policies for Space Rocks.

These policies govern where documentation belongs, how documentation is classified, how folders and `!README.md` indexes are maintained, how draft documentation is handled, how documentation connects to implementation, and how stale documentation is removed.

The formal documenting procedure is defined separately. These policies describe the rules that procedure must follow.

## Core Policy

Documentation must be organized by documentation type, not only by topic name.

A “system” in Space Rocks can mean more than one thing:

```text
Domain system
= cross-system game, player, platform, or technical flow

Technical/service system
= executable, package, protocol, pipeline, or implementation boundary
```

Documentation must make that distinction explicit.

## Documentation Types

Space Rocks documentation uses these documentation types:

```text
Domain
Service
Protocol
Data
Systems Design
Devtools
Planning
Limits
Agent
Notes
Legacy
```

Each documentation type has a different purpose and a different ownership rule.

## Domain Documentation Policy

Domain documentation describes broad cross-system flows and integration.

Initial broad domain areas are:

```text
Player Experience
Platform
Technical
```

Domain docs explain how systems participate in a larger flow. They do not own implementation detail.

Domain docs should cover:

```text
participating systems
authority boundaries
durable roles
runtime roles
presentation roles
flow summary
inputs and outputs
integration points
out of scope
```

Domain docs must not map code directly.

Domain docs must link to associated technical systems by `!README.md` index.

Example:

```markdown
- [Game Server](../../services/game-server/!README.md)
- [Client](../../services/client/!README.md)
- [Player Data](../../services/player-data/!README.md)
- [Realtime Protocol](../../protocol/realtime/!README.md)
- [Data Pipeline](../../data/!README.md)
```

## Service Documentation Policy

Service documentation describes runtime, executable, package-group, or implementation responsibility.

Initial service areas include:

```text
game-server
client
player-data
api-server
web
```

Service docs should explain:

```text
what owns the runtime behavior
what the service implements
what the service does not own
what prose summary explains the runtime/API/protocol surface
what APIs or protocols it exposes or consumes
what data it persists or mutates
what domain flows it participates in
what tests verify it
what code paths implement it
```

Service docs should include code maps when they document implementation.

When a service doc covers an API, protocol, or other runtime surface, it must include a prose summary that explains:

```text
what the surface is for
who calls or consumes it
who owns authority behind it
what data crosses the boundary
what the surface explicitly does not own
```

Endpoint tables, packet lists, wrapper method lists, and code maps are supporting detail, not a replacement for explanatory text.

Service docs should not describe the entire cross-system domain concept unless that concept is fully owned by the service.

## Protocol Documentation Policy

Protocol documentation is first-class and separate from domain documentation.

Protocol docs describe communication, message orchestration, transport behavior, compatibility expectations, and service responsibilities.

Protocol docs should explain:

```text
which systems communicate
what messages, packets, requests, or responses are involved
what sequence or lifecycle the protocol follows
which service owns each part
what source-of-truth files define it
what validation or compatibility rules exist
what implementation docs cover the runtime paths
```

Protocol docs may include code maps when they cover implementation paths.

When a protocol doc covers a request, message, packet, or transport surface, it must include a prose summary that explains:

```text
what the surface is for
who calls or consumes it
who owns authority behind it
what data crosses the boundary
what the surface explicitly does not own
```

Endpoint tables, packet lists, wrapper method lists, and code maps are supporting detail, not a replacement for explanatory text.

Protocol docs should link to related service docs and data docs for detailed ownership and source-of-truth information.

Protocol docs should not become broad domain docs. They explain how systems communicate, not why the product/domain flow exists.

## Data Documentation Policy

Data documentation covers source-of-truth material, generated outputs, schema/data contracts, persistence contracts, and data pipelines.

Data docs must document pipeline usage and configuration where relevant.

Data docs may cover:

```text
data-sync usage
data-sync configuration
source files
generated outputs
validation commands
pull/push/diff workflows
packet generation
constant generation
collision-shape export/import
drop-table generation
player-data schemas
persistence contracts
common pipeline failure modes
```

Data docs should explain both the data and the operational pipeline around that data.

Data docs should include code maps or source maps when they document implementation, generation, or pipeline behavior.

## Systems Design Documentation Policy

`systems-design/` is the home for conceptual mechanics, authority boundaries, and invariants.

Systems-design replaces the old broad `design/` category.

Initial systems-design subareas are:

```text
world
combat
entities
```

Top-level systems-design keeps broad docs such as:

```text
!README.md
architecture.md
authority-boundaries.md
```

Systems-design docs should explain:

```text
conceptual model
authority rules
invariants
mechanics
participating systems
related implementation docs
related data or protocol docs
```

Permanent architectural constraints, intentional boundaries, and design invariants belong in systems-design docs.

Systems-design docs may include implementation links when useful, but they should not become exhaustive code indexes.

## Devtools Documentation Policy

Devtools documentation is separate from production gameplay and normal service documentation.

Devtools are development/debug tooling.

Devtools documentation should be split into:

```text
design
server
client
```

Devtools docs should cover:

```text
debug-only scope
server authority
client presentation
commands and controls
telemetry
build/runtime gates
relationship to real gameplay implementation areas
```

Devtools must not document or encourage parallel debug-only gameplay logic that bypasses real game systems.

Devtools implementation docs should include code maps.

## Planning Documentation Policy

Planning docs describe future, unresolved, proposed, or not-yet-current work.

Planning docs should not pretend future work already exists.

Planning docs should not remain the permanent home for implemented facts.

Planning docs should link to relevant existing docs whenever possible.

Planning docs should cover:

```text
purpose
overview
current status
decisions made
open decisions
expected ownership
implementation sequence
related docs
notes
```

Planning docs should clearly distinguish:

```text
decided direction
open questions
future work
current implementation facts
temporary blockers
```

Planning docs should follow the same placement logic inside `docs/planning/` that current docs follow outside it.

Examples:

```text
current domain docs     -> docs/domains/...
planning domain docs    -> docs/planning/domains/...

current service docs    -> docs/services/...
planning service docs   -> docs/planning/services/...

current protocol docs   -> docs/protocol/...
planning protocol docs  -> docs/planning/protocol/...

current data docs       -> docs/data/...
planning data docs      -> docs/planning/data/...

current systems design  -> docs/systems-design/...
planning systems design -> docs/planning/systems-design/...

current devtools docs   -> docs/devtools/...
planning devtools docs  -> docs/planning/devtools/...
```

When planned work becomes implemented, documentation updates should follow the standard documenting procedure. There is no separate graduation policy.

## Limits Documentation Policy

`docs/limits/` is for temporary or active problems.

Limits docs are not for intended architectural limitations.

Limits docs may cover:

```text
temporary implementation gaps
known bugs
dev-blocked issues
blocking issues
incomplete transitional behavior
current constraints that should be fixed later
```

Permanent design constraints, intentional boundaries, and architecture rules belong in `docs/systems-design/`.

Completed systems should not routinely have “Known limits” sections.

If a current doc needs to reference an active problem, it should use an `Active issues` section and link to the relevant sorted limits backlog heading.

Limits backlogs should be sorted inside `docs/limits/`.

Example categories:

```text
gameplay-backlog.md
platform-backlog.md
technical-backlog.md
dev-blockers.md
```

## Agent Documentation Policy

Agent docs describe editing rules, testing expectations, architecture guardrails, and workflow instructions for agents.

Agent docs may link to current docs, but they should not become the main home for system facts.

Agent docs should not duplicate implementation documentation that belongs under services, protocol, data, systems-design, domains, or devtools.

## Notes Documentation Policy

`docs/notes.md` persists as a non-authoritative scratchpad.

It is for temporary, unclear, uncategorized, or not-yet-classified notes.

Rules:

```text
notes.md is allowed to exist
notes.md is not authoritative
stable notes should move into the correct docs folder
obsolete notes should be deleted
notes.md should be periodically triaged
notes.md should not be used to avoid creating a proper doc when ownership is clear
```

## Legacy Documentation Policy

`docs/legacy/` is temporary migration source material only.

Legacy docs are not current authority.

Legacy docs should be used to extract useful facts while rebuilding current documentation.

Once useful facts have been mined, rewritten, split, or intentionally discarded, the legacy source doc should be deleted.

Do not keep stale legacy docs indefinitely.

Current docs should not link to legacy docs as authority.

## Folder Creation Policy

Create a new folder only when it is a durable boundary that will have multiple related docs.

Do not create folders for vague buckets such as:

```text
misc
common
general
stuff
gameplay
```

If the information is temporary, unclear, or not ready to classify, use `docs/notes.md`.

If the information is incomplete but has a clear eventual home, use a nearby `stubs/` folder.
Empty folders named exactly `stubs/` may remain in place as reserved draft locations. They are exempt from empty-folder cleanup, do not need their own `!README.md`, and do not need to be listed as a Direct Folder in the parent `!README.md` when there are no stub files to index.

This exception applies only to folders named exactly `stubs/`. All other empty documentation folders remain noncompliant unless they are intentionally populated with a valid `!README.md` and indexed by their parent.

## `!README.md` Index Policy

Every documentation folder must contain a `!README.md`.

`stubs/` folders are exempt from this index requirement.

Every `!README.md` must index:

```text
every markdown file directly in that folder
every direct subfolder
```

Markdown files are linked directly.

Example:

```markdown
- [networking.md](networking.md) - Game-server networking responsibilities and runtime flow.
```

Subfolders are linked by folder name to the subfolder `!README.md`.

Example:

```markdown
- [Game Server](game-server/!README.md) - Go realtime server implementation docs.
- [Random Subfolder](Random Subfolder/!README.md) - Example subfolder index link.
```

Rules:

```text
No orphan docs.
No folder without a `!README.md`.
Subfolder links must point to the subfolder `!README.md`.
```

The top-level `docs/!README.md` is both:

```text
documentation rulebook
top-level documentation index
```

Files linking to their parent index must use `./!README.md` instead of plain `!README.md` due to relevant pathing issues.

## Stub Policy

Incomplete docs belong in a nearby `stubs/` folder.

Canonical docs belong in the owning parent folder.

A stub is not canonical documentation.

A stub may be incomplete, exploratory, partial, or waiting for enough detail to become official documentation.

`stubs/` folders are exempt from `!README.md` index requirements.
Empty `stubs/` folders are also exempt from empty-folder cleanup.
Only folders named exactly `stubs/` receive this exemption.

Links to stub files in parent `!README.md` indexes must label the description as a stub.

If a `stubs/` folder has its own index, it should explain that docs in the folder are:

```text
drafts
incomplete
non-canonical
expected to graduate or be deleted
```

A stub should move out of `stubs/` only when it is complete enough to be considered canonical documentation.

When a stub becomes canonical:

```text
move it from stubs/ into the parent folder
update the parent `!README.md` index
remove it from the stubs index, if one exists
ensure the doc has the required shape for its type
add related docs
add code maps if required
delete the old stub path
```

## Universal Document Shape Policy

Every normal documentation file should include, at minimum:

```text
Purpose
Overview
Type-specific sections
Related docs
Notes
```

## Purpose Section Policy

The `Purpose` section explains why the document exists.

It should be short and direct.

## Overview Section Policy

The `Overview` section explains the actual thing being documented.

It should describe:

```text
what it does
how it behaves
why it exists
how it fits into the project
```

Technical language is encouraged when useful.

Short code examples, packet examples, data shapes, pseudocode, or flow snippets are encouraged when they clarify behavior.

The overview should be descriptive documentation, not just a vague summary.

## Related Docs Section Policy

Every normal doc should include a single `Related docs` section unless the doc becomes large enough to justify grouping related links.

Related docs should include any relevant:

```text
domain docs
service docs
protocol docs
data docs
systems-design docs
devtools docs
planning docs
limits docs
agent docs
```

Docs should link to the most relevant canonical docs, not duplicate their content.

## Notes Section Policy

Every normal doc should end with:

```markdown
## Notes
```

The Notes section is for relevant information that does not fit cleanly elsewhere.

Good uses:

```text
small caveats
temporary context
naming notes
historical context that still matters
edge cases
minor implementation observations
open but non-blocking questions
```

Bad uses:

```text
large backlog items
known blockers
future plans
core design rules
implementation details that deserve their own section
```

Those belong elsewhere:

```text
blockers/issues -> docs/limits/
future plans -> docs/planning/
permanent design rules -> docs/systems-design/
implementation ownership -> docs/services/
```

## Active Issues Policy

Completed systems should not routinely include “Known limits.”

Use `Active issues` only when the doc needs to reference temporary issues, blockers, dev-blocked work, or incomplete transitional behavior.

Active issues should link to sorted limits backlog headings.

Anything temporary should ideally be kept out of docs and in a work order or to-do list instead.

Example:

```markdown
## Active issues

- Loadout validation does not yet perform inventory-backed ownership checks. See [Loadout ownership validation](../../limits/gameplay-backlog.md#loadout-ownership-validation).
```

## Code Map Policy

Implementation-facing docs should include code maps.

Code maps are required for:

```text
services/
protocol/ when implementation paths are covered
data/
devtools/server/
devtools/client/
```

Code maps are not required for:

```text
domains/
planning/
limits/
agent/
```

Code maps are optional for:

```text
systems-design/
```

A code map should include:

```text
primary implementation files or folders
related generated/source files
related tests
important non-ownership boundaries
```

Do not add documentation comments to every source file.

Use source comments only for unusually easy-to-misunderstand implementation areas where a doc link is not enough.

## Legacy Removal Policy

Legacy docs should be deleted once fully deprecated.

A legacy doc is fully deprecated when:

```text
all useful facts have been migrated, rewritten, or intentionally discarded
current docs no longer depend on it
no `!README.md` index presents it as current authority
```

Stale legacy documentation should not be preserved indefinitely.
