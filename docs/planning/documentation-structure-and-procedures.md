## Documentation Structure And Governance
Parent index: [Planning](./!INDEX.md)

## Purpose

This plan defines the documentation restructuring work for Space Rocks.

The goal is to create a durable documentation system that clearly separates:

* cross-system domain flows
* runtime/service implementation responsibilities
* protocol behavior
* data/source-of-truth pipelines and configuration
* systems-design concepts and invariants
* devtools behavior
* future planning
* current limits
* agent/editing rules
* temporary legacy migration material

The current problem is not simply that documentation is missing. The larger problem is that existing documentation does not consistently explain what kind of documentation it is, where new information belongs, how planned systems graduate into implemented documentation, or how documentation relates back to implementation.

This restructuring should prevent the project from needing another broad documentation reorganization later.

## Current Problems

Current documentation has several structural issues:

* Planning docs are becoming more organized than implemented/current-system docs.
* Implemented documentation is uneven, especially for the game server.
* Some existing docs mix conceptual design, implementation details, data/source-of-truth facts, planning residue, and known limits.
* Old docs under `docs/legacy/` are useful as source material, but are not current authority.
* There is no top-level documentation rulebook explaining where new docs belong.
* Folder indexes are inconsistent or missing.
* Planned systems do not have a formal graduation path into implemented docs.
* Code-to-doc linking is not consistently handled.
* Data pipeline usage and configuration are not documented as first-class current-system documentation.

## Documentation Types

The new documentation structure should be organized by documentation type.

### Domains

Domain docs describe broad cross-system flows and integration.

Domains are not one implementation module. They describe how game-facing, player-facing, platform-facing, or technical flows move through multiple systems.

Initial domain groups:

* Player Experience
* Platform
* Technical

Domain docs answer:

* What cross-system flow exists?
* Which systems participate?
* Which system owns authority?
* Which systems are durable, runtime, or presentation participants?
* What data moves through the flow?
* What consumes the flow output?
* What is explicitly out of scope?
* Which implementation docs explain the participating systems?

Domain docs should not map directly to code files. They should link to associated technical system documentation and implementation docs.

Links should mask file names behind the system name.

Example:

```markdown
- [Game Server](../../services/game-server/!INDEX.md)
- [Client](../../services/client/!INDEX.md)
- [Player Data](../../services/player-data/!INDEX.md)
- [Realtime Protocol](../../protocol/realtime-websocket-protocol.md)
- [Data Pipeline](../../data/!INDEX.md)
```

### Services

Service docs describe executable/runtime implementation responsibilities.

Service docs answer:

* What executable, package group, or runtime boundary owns this?
* What APIs or runtime surfaces does it expose?
* What data does it store or mutate?
* What domain responsibilities does it implement?
* What does it trust?
* What does it reject?
* What tests verify it?
* What code paths implement it?

Initial service areas:

* game-server
* client
* player-data
* api-server
* web

Service docs should include code maps where they cover implementation.

### Protocol

Protocol docs are first-class and separate from domain docs.

Protocols are similar to domains because they describe cross-system orchestration, but they also need implementation detail, message structure, timing assumptions, compatibility rules, and service-specific responsibilities.

Protocol docs answer:

* What systems communicate?
* What messages, packets, requests, or responses are involved?
* What sequence or lifecycle does the protocol follow?
* What service owns each part of the protocol?
* What data/source-of-truth defines the protocol?
* What validation or compatibility expectations exist?
* What implementation docs cover the runtime paths?

Protocol docs may include code maps when they directly cover implementation paths, but should prefer linking to service and data docs for detailed ownership.

### Data

Data docs describe source-of-truth files, generated outputs, schema/data contracts, persistence contracts, and the pipelines used to generate, validate, and synchronize data across services.

Data docs include extensive documentation for:

* data-sync usage
* data-sync configuration
* source-of-truth files
* generated outputs
* validation/check commands
* pull/push/diff workflow
* packet generation
* constants generation
* player-data schema contracts
* collision-shape export/import
* drop-table source and generation
* persistence contracts

Data docs answer:

* Where is the source data?
* What generates from it?
* Which services consume it?
* How is it updated?
* How is it validated?
* Which commands are required?
* What configuration affects it?

Data docs should include code/source maps where they cover implementation or pipeline behavior.

### Systems Design

`systems-design/` replaces `design/`.

Systems-design docs describe conceptual mechanics, authority boundaries, rules, and invariants that should survive implementation changes.

Top-level systems-design should keep:

* README
* architecture
* authority boundaries

Initial subareas:

* world
* combat
* entities

Systems-design docs answer:

* Why does this system work this way?
* What conceptual boundary exists?
* What invariant must be preserved?
* What owns authority?
* Which domains, services, protocols, or data docs implement the concept?

Systems-design docs may include implementation links when useful, but should not become exhaustive code indexes.

### Devtools

Devtools are a separate documentation domain.

Devtools are not normal gameplay docs and not production environment docs. They document debug and development tooling that spans client and server.

Devtools should be split into:

* design
* server
* client

Devtools docs answer:

* What debug/development capability exists?
* What is client-only presentation?
* What is server-authoritative mutation?
* What commands or controls exist?
* What telemetry exists?
* What is disabled or excluded in non-dev builds?
* How do devtools route through real gameplay implementation areas?

Devtools implementation docs should include code maps.

### Planning

Planning docs describe future, unresolved, or not-yet-current systems.

Planning docs answer:

* What is intended?
* What has been decided?
* What remains open?
* What sequence or dependency exists?
* What current docs will receive the implemented facts later?

Planning docs should not remain the long-term home for implemented system facts.

### Limits

Limits docs describe known current constraints, intentionally incomplete behavior, temporary limitations, and non-final implementation facts.

Limits docs answer:

* What is currently constrained?
* What is known incomplete?
* What behavior should not be mistaken for final design?
* What planning docs describe the intended future?

### Agent

Agent docs describe editing rules, testing expectations, architecture guardrails, and workflow instructions for agents.

Agent docs are not normal system documentation. They may link to service, domain, protocol, data, systems-design, or devtools docs, but should not become the main source of current-system facts.

### Legacy

`docs/legacy/` is a temporary migration source only.

Legacy docs are not current authority. They exist only to preserve old documentation while useful facts are mined, rewritten, split, or intentionally discarded.

Once a legacy document has been fully replaced or intentionally deprecated, it should be deleted. Stale legacy docs should not be kept indefinitely.

### Notes

`docs/notes.md` should persist.

It is a fallback scratchpad for uncategorized notes, temporary observations, or items that do not yet have a clear documentation home.

Rules for `docs/notes.md`:

* It is allowed to exist.
* It is not authoritative current documentation.
* It should be periodically triaged.
* Stable or repeated notes should graduate into the correct docs folder.
* It should not be used to avoid creating a proper doc when ownership is already clear.

## README And Index Rules

The top-level `docs/!INDEX.md` should be both:

* the documentation rulebook
* the top-level documentation index

Do not create a separate top-level index unless the top-level README becomes too large to serve both purposes.

Every documentation folder must contain a `!INDEX.md`.

Every `!INDEX.md` must, at minimum:

* explain what the folder owns
* explain what does not belong there
* index every markdown file directly in the folder
* index every direct subfolder
* link to each direct subfolder `!INDEX.md`

This rule applies at every folder level.

Folder README files are the primary navigation mechanism for documentation.

## Domain Documentation Rules

Domain docs describe cross-system flows and integration.

Domain docs should not include required code maps.

Domain docs must link to associated system documentation by `!INDEX.md` index. The link text should be the name of the technical system, not the file name.

Domain docs should link to:

* associated domain docs
* associated service docs
* associated protocol docs
* associated data docs
* associated systems-design docs
* associated devtools docs when relevant
* associated planning docs
* associated limits docs

Domain docs should use terms consistently:

* domain system
* authority boundary
* service responsibility
* runtime role
* durable role
* presentation role
* integration points

## Service Documentation Rules

Service docs describe implementation responsibility inside one executable/runtime boundary.

Service docs should include:

* purpose
* code root
* responsibilities
* explicit non-responsibilities
* implemented domain roles
* APIs or protocol surfaces
* data ownership
* related domain docs
* related protocol docs
* related data docs
* related systems-design docs
* tests and verification
* code map

Service docs should not describe the whole cross-system product/domain concept unless that concept is fully owned by the service.

## Protocol Documentation Rules

Protocol docs describe communication behavior and message orchestration between systems.

Protocol docs should include:

* purpose
* participating systems
* protocol authority
* message/request/response flow
* lifecycle or sequence
* source-of-truth files
* generated outputs where relevant
* service responsibilities
* compatibility or versioning expectations where relevant
* validation/testing
* related domain docs
* related service docs
* related data docs

Protocol docs should not become broad domain docs. They should explain how systems communicate, not why the product/domain flow exists.

## Data And Pipeline Documentation Rules

Data docs should treat pipeline usage and configuration as first-class content.

Data docs should include:

* source files
* generated outputs
* consuming services
* update commands
* validation commands
* configuration files
* expected workflow
* common failure modes
* related service docs
* related protocol docs
* related systems-design docs

Data docs should explain both the data and the operational pipeline around that data.

## Systems-Design Documentation Rules

Systems-design docs describe conceptual boundaries and invariants.

Systems-design docs should include:

* purpose
* conceptual model
* authority rules
* invariants
* participating domains
* service implementations
* related protocol docs
* related data docs
* known limits
* related planning docs

Systems-design docs should not become exhaustive implementation maps. Service docs own detailed implementation maps.

## Devtools Documentation Rules

Devtools docs are split from production gameplay and normal service docs.

Devtools docs should include:

* debug/development purpose
* server authority rules
* client presentation rules
* command/control behavior
* telemetry behavior
* build/runtime gates
* relationship to real gameplay implementation areas
* code maps for implementation docs

Devtools must not document or encourage parallel debug-only gameplay logic that bypasses real game systems.

## Code And Documentation Linking Rules

### Documentation To Code

Service, data, protocol, and devtools implementation docs should include code maps for the implementation they cover.

A code map should list:

* primary implementation files or folders
* related generated/source files
* related tests
* important non-ownership boundaries

Domain docs should not map code directly.

Systems-design docs may include implementation links where useful, but should not become exhaustive code indexes.

Source comments should be used only for unusual or easily misunderstood implementation areas where a doc link is not enough.

Do not add documentation links to every file, every function, or routine implementation details.

## Planning-To-Implemented Graduation Procedure

Graduation from planning to current documentation is a real procedure.

When a planned system becomes implemented or partially implemented:

1. Identify which planning facts are now implemented.
2. Classify each implemented fact by documentation type:

   * cross-system flow -> domains
   * communication contract -> protocol
   * runtime responsibility -> services
   * data/source/pipeline/schema -> data
   * conceptual mechanics/invariants -> systems-design
   * dev/debug tooling -> devtools
   * known current constraint -> limits
3. Create or update the target folder `!INDEX.md` index before adding new docs.
4. Create or update the current-system documentation.
5. Add links from domain docs to associated technical systems by `!INDEX.md`.
6. Add code maps only in service, data, protocol, and devtools implementation docs where appropriate.
7. Remove or reduce duplicated implemented facts from the planning doc.
8. Add an `Implemented references` section in the planning doc linking to the new current docs.
9. Leave unresolved decisions, future variants, and sequencing in the planning doc.
10. Update parent `!INDEX.md` indexes.
11. Update `docs/!INDEX.md` only if the documentation taxonomy itself changed.

Planning docs should include a `Graduation targets` section before or during implementation.

A `Graduation targets` section should state where implemented facts will move:

* cross-system flow facts -> domains
* protocol facts -> protocol
* runtime implementation facts -> services
* data/source/pipeline facts -> data
* conceptual mechanics/invariants -> systems-design
* dev/debug facts -> devtools
* known limitations -> limits
* remaining future work -> the planning doc

## Rebuild And Migration Strategy

Migration should be rebuild-first, not move-first.

The new documentation should be built from:

* current code analysis
* existing planning docs
* useful facts extracted from legacy docs
* known upcoming refactors
* current limitations

Legacy docs are source material, not authoritative migration units.

Do not blindly move a legacy file if it mixes:

* domain flow
* service implementation
* protocol behavior
* data/source-of-truth facts
* conceptual systems design
* current limits
* future planning residue

Instead, split or rewrite the content into the correct documentation type.

Current-system docs should document what exists now. If a major planned refactor will change the area later, the current doc should link to the relevant planning doc rather than pretending the future structure already exists.

## Known Timing Constraints

Some major refactors are already planned and may not be represented in current implementation docs yet.

The rebuild should account for this by separating:

* current implemented facts
* known current limits
* planned changes
* future intended ownership

Current docs should not overfit to soon-to-change code, but they also should not document unimplemented architecture as current fact.

When an area is expected to change soon, document the current state clearly and link to planning.

## Initial Implementation Phases

### Phase 1: Write This Planning Document

Create the documentation restructuring plan under technical planning.

The planning document should capture:

* documentation types
* README/index rules
* domain/service/protocol/data/systems-design/devtools distinctions
* code/documentation linking rules
* package/folder documentation index rules
* graduation procedure
* rebuild strategy
* legacy deletion policy
* notes scratchpad policy
* implementation phases
* acceptance criteria

### Phase 2: Create Top-Level Documentation Governance

Create or update `docs/!INDEX.md`.

It should act as both:

* documentation rulebook
* top-level documentation index

It should define all documentation types and where new docs belong.

### Phase 3: Install Minimum Category Scaffolding

Create only category-level folders and README files first.

Initial categories:

* domains
* services
* protocol
* data
* systems-design
* devtools
* planning
* limits
* agent
* legacy

Also create missing README files in existing planning subfolders.

Do not build detailed folder trees until the category rules and README/index conventions are established.

### Phase 4: Create README Template

Create an agent-facing folder README template.

The template should include:

* purpose
* does not own
* files
* folders
* related documentation
* placement rules

### Phase 5: Rebuild Service Documentation

Analyze actual service structure and create accurate service `!INDEX.md` indexes.

Initial services:

* game-server
* client
* player-data
* api-server
* web

Start with high-level service responsibility maps before writing detailed service docs.

### Phase 6: Rebuild Data And Pipeline Documentation

Document:

* data-sync usage
* data-sync configuration
* source-of-truth files
* generated outputs
* packet generation
* constants generation
* collision-shape export/import
* drop-table generation
* player-data schema
* persistence contracts
* validation/check commands

### Phase 7: Rebuild Protocol Documentation

Document protocols separately from domains and services.

Initial protocol areas should cover:

* realtime protocol
* HTTP/API protocol behavior

Protocol docs should link to service docs and data docs.

### Phase 8: Rebuild Systems-Design Documentation

Rebuild systems-design docs from actual behavior and useful legacy material.

Initial systems-design subareas:

* world
* combat
* entities

Top-level systems-design should include architecture and authority boundaries.

### Phase 9: Rebuild Domain Integration Documentation

Create broad domain docs.

Initial domain groups:

* Player Experience
* Platform
* Technical

Domain docs should describe cross-system flows and link to associated technical system `!INDEX.md` indexes.

### Phase 10: Rebuild Devtools Documentation

Document devtools separately from production gameplay docs.

Initial devtools split:

* design
* server
* client

Use legacy devtools docs as extraction material only.

### Phase 11: Update Planning Docs With Graduation Targets

As planning docs are touched, add `Graduation targets`.

Do not mass-edit every planning doc before the target docs exist unless the change is small and mechanical.

### Phase 12: Triage Legacy And Notes

Use `docs/legacy/` as temporary extraction material.

Delete legacy docs once they are fully deprecated.

Keep `docs/notes.md` as a non-authoritative scratchpad and periodically triage it.

## Acceptance Criteria

The restructuring is successful when:

* `docs/!INDEX.md` explains documentation types, folder rules, `!INDEX.md` index rules, graduation procedure, and code-linking policy.
* Every documentation folder has a `!INDEX.md` index.
* Every `!INDEX.md` indexes direct files and direct subfolders.
* Domain docs describe broad cross-system flows and link to associated technical system `!INDEX.md` indexes.
* Domain docs do not directly map code.
* Protocol docs are first-class and separate from domains.
* Data docs include source-of-truth, generated output, pipeline usage, pipeline configuration, and validation procedures.
* `systems-design/` replaces `design/` as the home for conceptual mechanics, authority boundaries, and invariants.
* Devtools docs are separate from production gameplay docs and split into design, server, and client documentation.
* Service, data, protocol, and devtools implementation docs include code maps where appropriate.
* Planning docs have a real graduation procedure and use `Graduation targets` as they are updated.
* Legacy docs are not current authority.
* Legacy docs are deleted once fully deprecated.
* `docs/notes.md` persists as a scratchpad but is not authoritative.
* Current docs distinguish implemented facts from future plans and known limits.
