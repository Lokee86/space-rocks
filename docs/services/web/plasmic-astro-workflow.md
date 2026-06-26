# Plasmic / Astro Workflow

Parent index: [Web](./!INDEX.md)

## Purpose

This document describes the current Plasmic-to-Astro workflow used by the Space Rocks web site.

It defines the implementation boundary between Plasmic visual/layout output, the Astro deployable site, owned wrapper/source files, and the temporary parked Next/Plasmic host app.

## Overview

The current Space Rocks web site uses Astro as the deployable static site and Plasmic as the visual/layout source.

The implemented workflow is:

```text
Plasmic Studio
-> visual layout and responsive page design

temporary parked Next/Plasmic host
-> supports Plasmic Studio workflow

Plasmic codegen/output
-> generated React layout components

owned Astro/React source files
-> routing, content loading, behavior, and integration glue

Astro build output
-> deployable static website
```

The current deployable website lives in:

```text
web-astro/
```

The temporary Plasmic/Next host app is parked under:

```text
tools/parked-plasmic-next-host/
```

The parked host is not the deployed website. It exists to support Plasmic Studio workflow while the current static site is implemented in Astro. It can be recycled, replaced, or promoted when the future interactive website is established.

The hard ownership rule is:

```text
Custom behavior MUST go in owned wrapper/source files, always.
```

Generated Plasmic files are layout output. They must not become customization surfaces.

## Code root

Current deployable site:

```text
web-astro/
```

Temporary parked Plasmic/Next host:

```text
tools/parked-plasmic-next-host/
```

Owned wrapper/source files for the current Astro site include:

```text
web-astro/src/pages/
web-astro/src/components/Homepage.tsx
web-astro/src/components/HomepageClientMount.tsx
web-astro/src/components/Archive.tsx
web-astro/src/components/ArchiveClientMount.tsx
web-astro/src/components/media/
web-astro/src/components/card/
web-astro/src/content/
web-astro/src/compat/
```

Generated Plasmic output includes:

```text
web-astro/src/components/plasmic/
web-astro/src/plasmic/
web-astro/src/components/plasmic-tokens.theo.json
web-astro/plasmic.json
web-astro/plasmic.lock
```

The parked Next/Plasmic host has its own package and generated files under its parked tool root. It should remain clearly separate from the deployable Astro site.

## Responsibilities

This workflow owns the current integration rules for:

```text
using Plasmic as the visual/layout source
keeping Astro as the deployable static site
routing Plasmic-generated layout through Astro pages
mapping repo-owned content into generated layouts
placing custom behavior in owned wrapper/source files
keeping generated Plasmic files free of custom behavior
maintaining the parked Next/Plasmic host as a workflow aid
preserving a future reuse path for the interactive website
```

The workflow exists to keep visual layout work and deployable site behavior connected without turning generated Plasmic output into the project’s behavior layer.

## Does not own

This workflow does not own:

```text
devlog content schema
devlog route behavior
CRT media-frame internals
static-site deployment policy
future launch website product planning
account portal behavior
commerce behavior
Steam ownership verification
CMS backend behavior
support/admin surfaces
leaderboard/profile surfaces
Plasmic visual design decisions
```

The devlog static site implementation is documented separately. The CRT media frame is documented separately. Future launch website planning remains in planning docs until implemented.

## Domain roles

The Plasmic/Astro workflow participates in the current website domain by defining the implementation handoff between visual layout and deployable site behavior.

Role split:

```text
Plasmic Studio
-> visual layout and responsive page composition

parked Next/Plasmic host
-> temporary Studio host/workflow support

generated Plasmic files
-> layout output consumed by React wrappers

owned wrapper/source files
-> custom behavior, content mapping, media wiring, and integration glue

Astro pages
-> route ownership, content loading, static generation

Astro build output
-> static deployable website
```

The workflow is an implementation boundary, not a product surface. Users do not interact with the parked host or generated Plasmic files directly.

## Protocols and APIs

This workflow does not expose a public application API.

It relies on several local/build-time surfaces:

```text
Plasmic Studio / host workflow
-> produces or syncs generated visual layout code

React component props
-> connect owned wrappers to generated Plasmic components

Astro content collection APIs
-> load repo-owned devlog content

Astro route/build pipeline
-> produces static pages and assets
```

The relevant boundary for runtime page rendering is:

```text
Astro route
-> loads content
-> passes content into client mount wrapper
-> owned React wrapper maps content into Plasmic overrides
-> generated Plasmic layout renders visual structure
-> owned components render custom behavior
```

Generated Plasmic files may define named override slots and instantiate code components. Owned wrappers should use those seams to pass data and behavior in, rather than editing generated files.

## Data ownership

This workflow does not own durable product data.

It owns implementation placement rules for generated layout output and wrapper/source integration.

Data and source ownership split:

```text
devlog Markdown/content files
-> owned by the static devlog site

Plasmic-generated layout files
-> generated layout output

owned wrappers
-> content mapping and behavior wiring

public assets
-> static web assets used by pages and generated output

parked Next/Plasmic host files
-> temporary Plasmic Studio workflow support
```

The parked host does not own deployable site content or production website state.

## Source-of-truth boundaries

Current source-of-truth rules:

```text
Astro is the deployable site source.
Plasmic is the visual/layout source.
Repo-owned content files are the public content source.
Owned wrappers/source files own custom behavior.
Generated Plasmic files are not customization surfaces.
The parked Next/Plasmic host is a workflow aid, not the deployed site.
```

Generated files may be overwritten by sync/codegen. Any custom logic placed inside generated files is at risk of being lost or causing merge/sync conflicts.

## Custom behavior rule

Custom behavior must go in owned wrapper/source files.

Use owned files for:

```text
content normalization
content-to-layout mapping
route behavior
media source selection
media control behavior
client mount behavior
links and navigation wiring
Astro integration
compatibility shims
site-specific React behavior
```

Do not place custom behavior in:

```text
web-astro/src/components/plasmic/
web-astro/src/plasmic/
generated Plasmic component files
generated Plasmic CSS modules
generated token output
parked host generated Plasmic output
```

Generated files may expose props, overrides, slots, or code-component seams. Those seams should be consumed from owned files.

## Owned wrapper pattern

The current homepage wrapper pattern is:

```text
Astro page
-> loads devlog content
-> passes content to HomepageClientMount
-> mounts Homepage
-> Homepage normalizes content
-> Homepage passes values into PlasmicHomepage overrides
-> generated Plasmic layout renders visual structure
```

Current homepage wrapper:

```text
web-astro/src/components/Homepage.tsx
```

Current archive wrapper:

```text
web-astro/src/components/Archive.tsx
```

Client mount wrappers:

```text
web-astro/src/components/HomepageClientMount.tsx
web-astro/src/components/ArchiveClientMount.tsx
```

This pattern should remain the default integration approach: generated layout stays visual, while owned wrappers provide behavior and data.

## Generated Plasmic files

Generated Plasmic files are implementation output.

They may contain:

```text
visual layout markup
generated style classes
generated component wrappers
named override definitions
Plasmic imports
code-component usage
responsive layout output
```

They should not contain hand-authored site behavior.

Current generated output includes:

```text
web-astro/src/components/plasmic/space_rocks_devlog/PlasmicHomepage.tsx
web-astro/src/components/plasmic/space_rocks_devlog/PlasmicArchive.tsx
web-astro/src/components/plasmic/space_rocks_devlog/*.module.css
web-astro/src/components/plasmic/space_rocks_devlog/plasmic.css
web-astro/src/components/plasmic/space_rocks_devlog/plasmic.tsx
web-astro/src/components/plasmic/space_rocks_devlog/PlasmicStyleTokensProvider.tsx
web-astro/src/components/plasmic/space_rocks_devlog/PlasmicGlobalVariant__Screen.tsx
```

Editing generated files directly should be avoided. If a generated file must be changed to recover or unblock workflow, the change should be treated as temporary and moved into an owned file or Plasmic source as soon as practical.

## Temporary parked Next/Plasmic host

The parked host lives at:

```text
tools/parked-plasmic-next-host/
```

It exists to support Plasmic Studio workflow. It is not the current deployable website.

The parked host may include:

```text
Next.js host app setup
Plasmic host configuration
Plasmic project files
generated Plasmic output for Studio support
local development dependencies
```

The parked host should remain conspicuous and separate from the deployable Astro site.

Operational meaning:

```text
web-astro/
-> current deployable Astro static site

tools/parked-plasmic-next-host/
-> temporary Plasmic/Next workflow host

future interactive website root
-> unresolved until implemented
```

When the future interactive website is established, the parked host can be recycled, replaced, or promoted if it still provides useful structure. Until then, it should not be treated as production web implementation.

## Astro integration

Astro owns the deployable static-site structure.

Current route files include:

```text
web-astro/src/pages/index.astro
web-astro/src/pages/archive/index.astro
web-astro/src/pages/devlog/[slug].astro
```

Astro pages load content and hand off rendering to React client mount wrappers:

```text
HomepageClientMount
ArchiveClientMount
```

This split allows Astro to own routing and content collection loading while React/Plasmic handle the current visual layout.

Astro also owns the static build output. The parked Next/Plasmic host is not part of the static deployment path.

## Code map

Current deployable Astro root:

```text
web-astro/
-> current static web site root
```

Astro route and content files:

```text
web-astro/src/pages/index.astro
-> homepage route; selects latest devlog entry

web-astro/src/pages/archive/index.astro
-> archive route; renders devlog archive data

web-astro/src/pages/devlog/[slug].astro
-> static devlog post routes

web-astro/src/content.config.ts
-> devlog content collection schema

web-astro/src/content/devlog/
-> repo-owned devlog source entries
```

Owned wrapper/source files:

```text
web-astro/src/components/Homepage.tsx
-> maps content into generated Homepage layout and media-frame props

web-astro/src/components/HomepageClientMount.tsx
-> client mount wrapper for Homepage

web-astro/src/components/Archive.tsx
-> maps archive content into generated Archive layout

web-astro/src/components/ArchiveClientMount.tsx
-> client mount wrapper for Archive

web-astro/src/content/homepageContent.ts
-> homepage content type and normalization

web-astro/src/content/archiveContent.ts
-> archive content type and normalization

web-astro/src/components/media/
-> owned media-frame behavior

web-astro/src/components/card/
-> owned card-frame behavior

web-astro/src/compat/
-> compatibility shims for generated Plasmic integration
```

Generated Plasmic files:

```text
web-astro/src/components/plasmic/
-> generated Plasmic React components and styles

web-astro/src/plasmic/
-> generated Plasmic support files

web-astro/plasmic.json
web-astro/plasmic.lock
-> Plasmic project/codegen metadata
```

Temporary parked host:

```text
tools/parked-plasmic-next-host/
-> temporary Next/Plasmic Studio host app
```

## Tests

No dedicated automated test suite is currently documented for the Plasmic/Astro workflow itself.

Current verification is workflow and runtime smoke testing.

Minimum verification:

```text
Astro dev server starts from web-astro/
homepage renders generated Plasmic layout
archive renders generated Plasmic layout
devlog post routes render generated Plasmic layout
owned wrappers pass content into named Plasmic overrides
media frame renders through generated layout instances
custom behavior remains in owned wrapper/source files
generated Plasmic files are not manually customized
parked Next/Plasmic host remains outside the deployable Astro root
Astro build succeeds for deployable static output
```

Relevant commands for the deployable site are run from:

```text
web-astro/
```

Current package scripts include:

```text
npm run dev
npm run build
npm run preview
npm run astro
```

The parked host may have its own local commands, but those are workflow-host commands, not deployable-site commands.

## Related docs

* [Web](./!INDEX.md)
* [Devlog Static Site](devlog-static-site.md)
* [CRT Media Frame](crt-media-frame.md)
* [Website and Web Presence](../../domains/web/website-and-web-presence.md)
* [Future Interactive Website](../../planning/web/stubs/interactive-website.md)
* [Future Website and Web Presence Plan](../../planning/domains/web/website-and-web-presence.md)
* [Documentation Policy](../../documentation-policy.md)
* [Documentation Procedure](../../documentation-procedure.md)

## Notes

The parked Next/Plasmic host exists because Plasmic Studio workflow and the deployable Astro site currently need different support surfaces. This is a workaround, not the product architecture target.

The important durable rule is that custom behavior belongs in owned wrapper/source files. Generated Plasmic output should stay visual/layout-focused and replaceable.

The future interactive website may reuse the parked host, but that should be decided when the interactive website becomes current implementation work.
