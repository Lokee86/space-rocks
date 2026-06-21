## Documentation Procedure
Parent index: [Documentation](./!INDEX.md)

## Purpose

This procedure defines the standard process for creating, updating, moving, and removing Space Rocks documentation.

Use this procedure for all documentation work, including current docs, planning docs, stubs, `!INDEX.md` indexes, limits, notes, and legacy cleanup.

## Procedure Summary

Follow these steps in order:

```text
1. Classify the documentation type.
2. Choose the owning folder.
3. Decide whether a new file or folder is needed.
4. Create or update all relevant `!INDEX.md` indexes.
5. Apply the stub rule if the doc is incomplete.
6. Write or update the doc using the required shape for its type.
7. Add related docs, code maps, active issues, and notes.
8. Update planning docs when planned work becomes current.
9. Clean up stale, duplicated, or legacy documentation.
10. Run a final verification pass.
```

## 1. Classify the Documentation Type

Before writing, moving, or updating a doc, identify what type of documentation it is.

Use these documentation types:

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

Classify by what the information **is**, not by where it was first discussed.

Use these rules:

```text
Domain          = cross-system flow and integration
Service         = runtime or executable implementation responsibility
Protocol        = communication/message flow
Data            = source-of-truth, schema, generated output, pipeline usage/config
Systems Design  = conceptual mechanics, boundaries, invariants
Devtools        = debug/development tooling
Planning        = future, unresolved, proposed, or not-yet-current work
Limits          = temporary issues, blockers, dev-blocked work
Agent           = editing/testing/workflow rules
Notes           = scratchpad
Legacy          = temporary migration source only
```

If the type is unclear, put the information in `docs/notes.md` until it can be classified.

## 2. Choose the Owning Folder

Place documentation where the information is owned, not where it is merely used.

Use this ownership mapping:

```text
Domain          -> docs/domains/
Service         -> docs/services/
Protocol        -> docs/protocol/
Data            -> docs/data/
Systems Design  -> docs/systems-design/
Devtools        -> docs/devtools/
Planning        -> docs/planning/
Limits          -> docs/limits/
Agent           -> docs/agent/
Notes           -> docs/notes.md
Legacy          -> docs/legacy/ temporarily, then delete when deprecated
```

Planning docs should mirror the current documentation structure inside `docs/planning/`.

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

If a planning equivalent folder does not exist yet, create it.

## 3. Decide Whether a New File or Folder Is Needed

Before creating anything new, check:

```text
Does an existing doc already own this?
Does an existing folder already own this type of information?
Is this a stable concern, or just a temporary note?
Would a new folder represent a durable boundary that will have multiple related docs?
```

Create a new file when:

```text
the topic has a clear owner
the topic has enough substance
the content does not fit cleanly into an existing doc
```

Create a new folder only when:

```text
it is a durable boundary that will have multiple related docs
```

Use `docs/notes.md` when:

```text
the information is temporary
the information is unclear
the information is not ready to classify
```

Use `stubs/` when:

```text
the information has a clear eventual home
the document is not complete enough to be canonical
```

Do not create vague folders such as:

```text
misc
common
general
stuff
gameplay
```

## 4. Create or Update All Relevant `!INDEX.md` Indexes

Before writing the doc body, make the location discoverable.

Every documentation folder must contain:

```text
!INDEX.md
```

Every `!INDEX.md` must index:

```text
every markdown file directly in that folder
every direct subfolder
```

Markdown files are linked directly:

```markdown
- [networking.md](networking.md) - Game-server networking responsibilities and runtime flow.
```

Subfolders are linked by folder name to the subfolder `!INDEX.md`:

```markdown
- [Game Server](game-server/!INDEX.md) - Go realtime server implementation docs.
- [Random Subfolder](Random Subfolder/!INDEX.md) - Example subfolder index link.
```

When adding a new file:

```text
1. Add the file to the parent `!INDEX.md`.
2. Use the file name as the link text.
3. Add a one-line description.
```

When adding a new folder:

```text
1. Create the folder `!INDEX.md`.
2. Add the folder to the parent `!INDEX.md`.
3. Use the folder name as the link text.
4. Link to folder-name/!INDEX.md.
5. Add a one-line description.
```

Update `docs/!INDEX.md` only when:

```text
a top-level documentation type changes
a new top-level docs folder is added
the documenting procedure changes
the index policy changes
```

Do not update `docs/!INDEX.md` for every ordinary doc addition.

## 5. Apply the Stub Rule If the Doc Is Incomplete

If the doc is incomplete, place it in the nearest appropriate `stubs/` folder.

`stubs/` folders are exempt from the `!INDEX.md` index requirement.
Empty `stubs/` folders may remain in place as reserved draft locations and do not need their own `!INDEX.md` or parent `Direct Folder` listing when no stub files are present.
Only folders named exactly `stubs/` receive this empty-folder exemption.

When a parent `!INDEX.md` indexes a stub file, the link description must start with `Stub:`.

Examples:

```text
docs/planning/services/game-server/stubs/loadout-resolution.md
docs/services/game-server/stubs/networking.md
docs/systems-design/combat/stubs/status-effects.md
```

Example entry:

```text
- `[example.md](stubs/example.md) - Stub: incomplete example documentation.`
```

If the needed `stubs/` folder does not exist:

```text
1. Create the stubs/ folder.
2. Add the stub file to the !INDEX index of its parent folder.
```

If a `stubs/` folder has its own index, it must state that files in the folder are:

```text
drafts
incomplete
non-canonical
expected to graduate or be deleted
```

When a stub becomes canonical:

```text
1. Move it from stubs/ into the parent folder.
2. Update the parent `!INDEX.md` index.
3. Remove it from the stubs index, if one exists.
4. Ensure the doc has the required shape for its type.
5. Add related docs.
6. Add code maps if required.
7. Delete the old stub path.
```

## 6. Write or Update the Doc Using the Required Shape

Every normal documentation file must include, at minimum:

```text
Purpose
Overview
Type-specific sections
Related docs
Notes
```

Use the required shape for the documentation type.

## Domain Doc Shape

Use for cross-system flows and integration.

Required sections:

```text
Purpose
Overview
Participating systems
Authority boundaries
Flow summary
Inputs and outputs
Out of scope
Related docs
Notes
```

Domain docs must not include direct code maps.

Domain docs should link to associated technical systems by `!INDEX.md` index.

## Service Doc Shape

Use for runtime, executable, package-group, or implementation responsibility.

Required sections:

```text
Purpose
Overview
Code root
Responsibilities
Does not own
Domain roles
Protocols and APIs
Data ownership
Code map
Tests
Related docs
Notes
```

The `Protocols and APIs` section must include a prose summary when the doc covers an API, protocol, or runtime surface. The summary must explain:

```text
what the surface is for
who calls or consumes it
who owns authority behind it
what data crosses the boundary
what the surface explicitly does not own
```

Endpoint tables, packet lists, wrapper method lists, and code maps are supporting detail, not a replacement for explanatory text.

## Protocol Doc Shape

Use for communication, message flow, packet flow, request/response flow, or transport behavior.

Required sections:

```text
Purpose
Overview
Participating systems
Authority
Message or request flow
Source-of-truth files
Service responsibilities
Validation and testing
Related docs
Notes
```

Include a code map when implementation paths are covered.

The `Message or request flow` section must include a prose summary when the doc covers a request, message, packet, or transport surface. The summary must explain:

```text
what the surface is for
who calls or consumes it
who owns authority behind it
what data crosses the boundary
what the surface explicitly does not own
```

Endpoint tables, packet lists, wrapper method lists, and code maps are supporting detail, not a replacement for explanatory text.

## Data Doc Shape

Use for source-of-truth files, generated outputs, persistence contracts, schemas, and pipelines.

Required sections:

```text
Purpose
Overview
Source files
Configuration
Generated outputs
Consumers
Pipeline usage
Validation commands
Failure modes
Code or source map
Related docs
Notes
```

## Systems-Design Doc Shape

Use for conceptual mechanics, boundaries, authority rules, and invariants.

Required sections:

```text
Purpose
Overview
Conceptual model
Authority rules
Invariants
Participating systems
Related docs
Notes
```

Implementation links are optional and should not become exhaustive code indexes.

## Devtools Doc Shape

Use for debug and development tooling.

Required sections:

```text
Purpose
Overview
Debug-only scope
Server authority
Client presentation
Commands or controls
Telemetry
Build/runtime gates
Code map
Tests
Related docs
Notes
```

## Planning Doc Shape

Use for future, unresolved, proposed, or not-yet-current work.

Required sections:

```text
Purpose
Overview
Current status
Decisions made
Open decisions
Expected ownership
Implementation sequence
Related docs
Notes
```

Use `Current status` to identify whether the plan is:

```text
stub
active planning
ready for implementation
partially implemented
mostly implemented
superseded
```

Planning docs should link to relevant existing docs whenever possible.

Planning docs should not pretend future work already exists.

## Limits Doc Shape

Use for temporary issues, blockers, dev-blocked work, known bugs, and incomplete transitional behavior.

Required sections:

```text
Purpose
Overview
Issue list or backlog
Affected docs/systems
Status
Related docs
Notes
```

Limits docs are not for intended architectural limits or permanent design constraints.

## Agent Doc Shape

Use for editing, testing, workflow, and architecture rules for agents.

Normal agent docs should include these sections:

```text
Purpose
Overview
Rules
Related docs
Notes
```

Agent docs should link to canonical docs for current system facts instead of duplicating implementation, protocol, data, devtools, domain, service, or systems-design details.

## 7. Add Related Docs, Code Maps, Active Issues, and Notes

After the main content is written, add the required supporting sections.

## Related Docs

Every normal doc should include:

```markdown
## Related docs
```

Use this section for relevant canonical docs.

Related docs may include:

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

Use one `Related docs` section unless the doc becomes large enough to justify grouping.

Domain docs should link to technical system `!INDEX.md` indexes, not code files or implementation files.

## Code Maps

Implementation-facing docs should include code maps.

Required for:

```text
services/
protocol/ when implementation paths are covered
data/
devtools/server/
devtools/client/
```

Not required for:

```text
domains/
planning/
limits/
agent/
```

Optional for:

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

## Active Issues

Completed systems should not routinely have “Known limits” sections.

Use:

```markdown
## Active issues
```

only when the doc needs to reference temporary issues, blockers, dev-blocked work, or incomplete transitional behavior.

Active issues should link to sorted limits backlog headings.

Example:

```markdown
## Active issues

- Loadout validation does not yet perform inventory-backed ownership checks. See [Loadout ownership validation](../../limits/gameplay-backlog.md#loadout-ownership-validation).
```

Permanent constraints and intentional boundaries belong in systems-design docs, not limits docs.

## Notes

Every normal doc should end with:

```markdown
## Notes
```

Use Notes for relevant information that does not fit cleanly elsewhere.

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

Do not use Notes for:

```text
large backlog items
known blockers
future plans
core design rules
implementation details that deserve their own section
```

## 8. Update Planning Docs When Planned Work Becomes Current

When planned work becomes current, update the relevant planning doc using the same standard documentation procedure.

Do not use a separate graduation policy.

When planned work becomes implemented or partially implemented:

```text
1. Identify which planning facts are now current.
2. Classify those facts by documentation type.
3. Move or rewrite current facts into the correct current docs.
4. Leave future work, unresolved decisions, and sequencing in planning.
5. Add an Implemented references section when useful.
6. Remove stale duplicate current facts from the planning doc.
7. Update related `!INDEX.md` indexes.
8. Update related docs as needed.
```

Planning docs should not remain the permanent home for implemented facts.

Planning docs may keep short summaries of implemented work when useful, but should link to the canonical current docs.

Example:

```markdown
## Implemented references

- [Game Server Loadout Resolution](../../services/game-server/loadout-resolution.md) - Runtime validation and build snapshot resolution.
- [Inventory Loadout Flow](../../domains/player-experience/inventory-loadout-flow.md) - Cross-system flow from ownership to match start.
```

## 9. Clean Up Stale, Duplicated, or Legacy Documentation

After adding or updating docs, remove stale or duplicate information.

Check for:

```text
planning docs that still contain current facts as if they are future
legacy docs that have been fully replaced
duplicate sections across current docs
limits that actually describe permanent systems-design rules
notes that now have a proper home
stub docs that should graduate or be deleted
`!INDEX.md` index entries that point to moved or deleted files
empty non-stub documentation folders
```

Legacy docs should be deleted once fully deprecated.

A legacy doc is fully deprecated when:

```text
all useful facts have been migrated, rewritten, or intentionally discarded
current docs no longer depend on it
no `!INDEX.md` index presents it as current authority
```

Do not keep stale legacy documentation indefinitely.

## 10. Final Verification Pass

Before considering the documentation change done, verify:

```text
The documentation type is correct.
The owning folder is correct.
No unnecessary file or folder was created.
All relevant `!INDEX.md` indexes are updated.
Subfolder links point to subfolder `!INDEX.md` files.
Stub policy was followed.
Empty folders named exactly `stubs/` were not flagged as stale or noncompliant.
The doc has Purpose, Overview, Related docs, and Notes.
Agent docs do not duplicate canonical current-system facts.
The doc has the required type-specific sections.
Implementation docs have code maps where required.
Domain docs link to system `!INDEX.md` indexes, not code.
Active issues link to sorted limits backlog headings.
Planning docs were updated if planned work became current.
Legacy docs were deleted if fully deprecated.
No stale duplicate facts remain.
docs/!INDEX.md was updated only if taxonomy or procedure changed.
```
