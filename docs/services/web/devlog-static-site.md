# Devlog Static Site

Parent index: [Web](./!INDEX.md)

## Purpose

This document describes the current Space Rocks devlog static-site implementation.

It defines the Astro content model, MDX post format, homepage/archive routing, Plasmic layout integration, media conventions, deployment path, and current verification expectations for the public devlog site.

## Overview

The current public web site is a static Astro site under:

```text
web-astro/
```

The site uses repo-owned devlog content files as the source of published posts. Astro owns content loading, route generation, and static build output. Plasmic owns the visual layout source. Owned React wrappers connect Astro content data to generated Plasmic layouts and owned presentation components.

The current implementation flow is:

```text
MDX devlog entry
-> Astro content collection
-> Astro page route
-> React client mount wrapper
-> owned Homepage / Archive wrapper
-> generated Plasmic layout
-> owned media and markdown components
-> static Cloudflare Pages deployment
```

The devlog currently supports:

```text
homepage latest-post rendering
archive listing
per-post static routes
YAML frontmatter-controlled layout fields
MDX article body content
Markdown-rendered selected frontmatter fields
hero media
article media gallery/video frame
progress cards
utility panel text
fine print
```

The static devlog exists to make project progress visible without requiring the future full interactive website, account portal, commerce surface, or player profile systems.

## Code root

Current deployable site root:

```text
web-astro/
```

Current devlog content root:

```text
web-astro/src/content/devlog/
```

Current route files:

```text
web-astro/src/pages/index.astro
web-astro/src/pages/archive/index.astro
web-astro/src/pages/devlog/[slug].astro
```

Current owned wrapper/source files:

```text
web-astro/src/components/Homepage.tsx
web-astro/src/components/HomepageClientMount.tsx
web-astro/src/components/Archive.tsx
web-astro/src/components/ArchiveClientMount.tsx
web-astro/src/components/markdown/MarkdownText.tsx
web-astro/src/content.config.ts
web-astro/src/content/homepageContent.ts
web-astro/src/content/archiveContent.ts
```

Related owned presentation components:

```text
web-astro/src/components/media/
web-astro/src/components/card/
```

Generated Plasmic output is under:

```text
web-astro/src/components/plasmic/
web-astro/src/plasmic/
web-astro/src/components/plasmic-tokens.theo.json
web-astro/plasmic.json
web-astro/plasmic.lock
```

Generated Plasmic files are not customization surfaces.

## Responsibilities

The devlog static site owns:

```text
devlog content schema
published devlog entry format
homepage latest-post selection
archive item projection
per-post static route generation
frontmatter-to-layout data mapping
MDX body insertion into the article layout
Markdown rendering for selected text fields
media path conventions for devlog entries
Cloudflare Pages static deployment settings
static-site build verification
```

The implementation should keep content and rendering responsibilities separated:

```text
Astro
-> content loading, route generation, static build output

MDX files
-> published post content and per-entry layout data

Owned wrappers
-> content normalization and Plasmic override wiring

Generated Plasmic layout
-> visual structure and named override slots

Owned components
-> media behavior, Markdown rendering, card/media presentation behavior
```

## Does not own

The devlog static site does not own:

```text
Plasmic visual design source
generated Plasmic file internals
CRT media-frame behavior internals
future full website product scope
account portal behavior
login/session behavior
commerce behavior
Steam ownership verification
leaderboards
profile pages
backend APIs
game runtime behavior
game-server state
player-data persistence
```

Those belong to their respective service, domain, planning, or systems-design docs.

## Domain roles

The devlog static site participates in the web presence domain as the current public publishing surface.

Current role split:

```text
Website/domain planning
-> decides public web presence goals and future site scope

Devlog static site
-> implements the current static public update surface

Plasmic / Astro workflow
-> defines the visual-layout and deployable-site integration boundary

CRT media frame
-> renders framed hero/article media

Cloudflare Pages
-> hosts the static build output

CanSpace
-> remains domain registrar unless intentionally changed

Cloudflare DNS
-> owns DNS after nameserver cutover
```

The devlog is a current implementation surface. Future account, commerce, profile, and launch-site features remain outside this static-site implementation until explicitly built.

## Protocols and APIs

The devlog static site does not expose a public application API.

It uses these build/runtime surfaces:

```text
Astro content collections
-> load and validate devlog entries

Astro static routes
-> generate homepage, archive, and per-post routes

React component props
-> pass normalized content into owned wrappers and generated Plasmic layouts

Static HTTP asset serving
-> serve images, UI assets, JS, CSS, and generated static pages

Cloudflare Pages build pipeline
-> install dependencies, run Astro build, publish dist/
```

The content collection schema is the main validation boundary for post files.

## Data ownership

The devlog static site owns repo-based published content data.

Owned content/data sources:

```text
web-astro/src/content/devlog/*.mdx
-> live devlog entries

web-astro/public/media/devlog/<slug>/
-> per-entry public media assets

web-astro/src/content.config.ts
-> Astro content schema

web-astro/src/content/homepageContent.ts
-> homepage content normalization contract

web-astro/src/content/archiveContent.ts
-> archive content normalization contract
```

The devlog content files are source-controlled. They are not CMS-managed at this stage.


## Content format

Current published devlog posts should use MDX:

```text
web-astro/src/content/devlog/<slug>.mdx
```

Each entry has:

```text
YAML frontmatter
MDX article body
```

The frontmatter begins and ends with `---` delimiters:

```mdx
---
title: "Devlog 0001: From Arcade to Online"
date: 2026-06-26
summary: "Archive/feed summary."
---

## Article body starts here
```

The body starts immediately after the closing frontmatter delimiter.

The current frontmatter model controls:

```text
archive/feed metadata
hero copy
hero media
article eyebrow
article title
article intro
article media frame
progress card titles and bodies
utility panel title and text
fine print
```

Current live fields include:

```yaml
title: "Devlog 0001: From Arcade to Online"
date: 2026-06-26
summary: "Short archive/feed summary."

heroLine1: "FROM 80'S ARCADES"
heroLine2: "TO AT HOME"
heroLine3: "ONLINE CHAOS"

heroMediaKind: "youtube"
heroImages: []
heroYoutubeUrl: "https://www.youtube.com/watch?v=8UXrnRvapHQ"
heroMediaAlt: "Gameplay demo"

articleLabel: "Devlog Entry - 0001"
articleTitle: "From Arcade to Online"
intro: |
  Intro paragraph one.

  Intro paragraph two.

articleMediaKind: "images"
articleImages:
  - "/media/devlog/0001-arcade-to-online/2026-06-26.torpedo_pickup.png"
  - "/media/devlog/0001-arcade-to-online/2026-06-26.one_up.png"
articleYoutubeUrl: ""
articleMediaAlt: "WIP screenshots from the game"

finishedTitle: "Playable Spine"
finishedBody: "Core loop, client, and multiplayer foundation are in place."

nowTitle: "Public Devlog"
nowBody: "The project now has a public update trail."

comingUpTitle: "Network Reinforcement"
comingUpBody: "Shoring up the networking foundations for gameplay expansion."

utilityTitle: "Status"
utilityText: >
  **Current signal:** Markdown-supported utility copy.

finePrint: "Space Rocks! is an active development project. All media, features, names, and content are work in progress and subject to change before final release."
```

Retired fields must not be required or reintroduced:

```text
whatChanged
callout
whatsNext
```

Those sections now belong in the MDX body instead of fixed frontmatter fields.

## YAML formatting rules

Frontmatter is YAML, not Markdown.

Use YAML lists with hyphens:

```yaml
articleImages:
  - "/media/devlog/example/image-one.png"
  - "/media/devlog/example/image-two.png"
```

Do not use Markdown `*` bullets for YAML arrays.

Use `|` for literal multi-paragraph content when paragraph breaks should be preserved:

```yaml
intro: |
  Paragraph one.

  Paragraph two.
```

Use `>` for folded text when line breaks may collapse into spaces:

```yaml
utilityText: >
  **Current signal:** This can wrap across lines in the source
  while rendering as one paragraph.
```

Block scalar content must be indented under the field:

```yaml
utilityText: >
  Correctly indented text.
```

Do not write:

```yaml
utilityText: >
**Incorrectly unindented text.**
```

Fields after a block scalar must return to top-level indentation:

```yaml
utilityText: >
  Utility text.

finePrint: "Top-level fine print."
```

Do not accidentally indent `finePrint` under `utilityText`.

YAML treats `*` as an alias marker. Unquoted Markdown emphasis at the beginning of an unindented YAML value can cause alias parse errors. Keep Markdown-rich text inside quoted strings or properly indented block scalars.

## MDX body rendering

The MDX body is the article body.

The body is inserted immediately after the article CRT media frame in the homepage/article layout.

Current intended order:

```text
article eyebrow
article title
intro
article CRT media frame
MDX body
```

The MDX body may contain normal Markdown:

```mdx
## What exists now

Paragraph text.

- Bullet item
- Bullet item

> Styled blockquote/callout.

## What comes next

More body content.
```

MDX allows future component usage when appropriate, but component use should remain deliberate. Content files should not become layout-code dumps.

Selected frontmatter fields are rendered through `MarkdownText` when Markdown support is needed:

```text
intro
finishedBody
nowBody
comingUpBody
utilityText
finePrint
```

Title-like fields should remain plain text:

```text
articleTitle
finishedTitle
nowTitle
comingUpTitle
utilityTitle
```

## Media asset conventions

Per-post public media should live under:

```text
web-astro/public/media/devlog/<slug>/
```

Example:

```text
web-astro/public/media/devlog/0001-arcade-to-online/
```

Public paths used in MDX/frontmatter omit `web-astro/public`:

```yaml
articleImages:
  - "/media/devlog/0001-arcade-to-online/2026-06-26.torpedo_pickup.png"
```

Hero and article media are configured independently.

Supported media kinds:

```text
""
"images"
"youtube"
```

Current usage pattern for the first live post:

```text
heroMediaKind: "youtube"
articleMediaKind: "images"
```

The hero uses a YouTube URL. The article media frame uses a list of public screenshot paths.

Media alt text should describe the media generally. For ordinary gameplay screenshots/video, a concise value such as “WIP screenshots from the game” or “Gameplay demo” is sufficient.

## Archive and latest-post behavior

The homepage selects the latest published devlog entry from the content collection.

The archive lists devlog entries and links to each generated per-post route.

The current route pattern is:

```text
/
-> latest devlog rendered through homepage layout

/archive/
-> archive listing

/devlog/<slug>/
-> individual static post route
```

Archive links must point to generated devlog routes, not raw content files.

If archive links 404, check:

```text
route path generation
slug construction
trailing slash behavior
Astro getStaticPaths output
Cloudflare build output
whether the content entry is actually included in the collection
```

If unwanted test entries appear, check for publishable `.md` or `.mdx` files under:

```text
web-astro/src/content/devlog/
```

Test entries should be removed from the active content collection path.

## Plasmic integration

Plasmic provides the visual layout source.

Generated Plasmic files are produced into:

```text
web-astro/src/components/plasmic/
web-astro/src/plasmic/
Test entries should be removed from the active content collection path.

Owned wrappers must wire content into Plasmic through overrides and owned components. Do not hand-edit generated Plasmic files for behavior.

Current relevant generated override slots include:

```text
heroLine1Media
heroLine2Media
heroLine3Media
heroLine1Desktop
heroLine2Desktop
heroLine3Desktop
heroMediaFrame
articleLabel
articleTitle
introText
articleMediaFrame
screenStack2
finishedTitle
finishedBody
nowTitle
nowBody
comingUpTitle
comingUpBody
utilityTitle
utilityText
finePrint
```

The `Homepage.tsx` wrapper owns the current custom insertion seam for the MDX body.

Important current integration rule:

```text
When inserting the article media frame in the custom MDX body stack, use the owned CrtMediaFrame directly.
Do not reinsert PlasmicHomepage.articleMediaFrame as a node component for this purpose.
```

The Plasmic node wrapper previously caused incorrect extra layout height. Owned behavior should stay in owned components/wrappers.

## Plasmic codegen workflow

Run Plasmic codegen from:

```text
web-astro/
```

Current project sync command:

```text
npx @plasmicapp/cli@0.1.365 sync -p uNJepqX5kmDcUn9dDb3UVD --yes
```

After Plasmic Studio changes:

```text
run codegen
run Astro build
inspect generated output only as output
commit generated changes if they are expected
```

Do not manually patch generated Plasmic files as a durable fix. If a generated value is wrong, change Plasmic Studio source or the owned wrapper/component contract.

## Parked Plasmic host caveat

The Plasmic Studio host currently uses the parked Next/Plasmic host under:

```text
tools/parked-plasmic-next-host/
```

The deployed site uses:

```text
web-astro/
```

The parked host has its own code-component copies. If those copies drift from `web-astro`, Plasmic Studio can render differently from the Astro site.

When code-component behavior changes, especially shared components like `CrtMediaFrame`, keep one of these true:

```text
the parked host component copy is updated to match web-astro
or Studio is pointed at a host using the current component implementation
```

Do not use Plasmic canvas mismatches as evidence that the Astro render is wrong until the host/component versions are checked.

## Cloudflare Pages deployment

The current static site deploys through Cloudflare Pages.

Use Pages, not Workers.

Correct Pages build settings:

```text
Repository: Lokee86/space-rocks
Production branch: main
Root directory: web-astro
Build command: npm run build
Build output directory: dist
```

Cloudflare UI gotcha:

```text
The Worker Git flow shows npx wrangler deploy.
That is not the current Astro static-site deployment path.
Use the Pages Git flow instead.
```

The correct product path is Cloudflare Pages. The Worker flow does not expose the Astro static-site settings needed for this project.

Before deploying, ensure the desired files are committed and pushed. Cloudflare builds from the Git commit it fetched, not from the local working tree.

If Cloudflare builds stale content, check:

```text
Cloudflare build commit SHA
local HEAD SHA
production branch
root directory
publishable files under web-astro/src/content/devlog/
whether the unwanted post is still committed
whether Cloudflare is building an older commit
```

## DNS and custom domain

The current registrar is:

```text
CanSpace
```

The intended DNS/hosting split is:

```text
CanSpace
-> remains registrar

Cloudflare DNS
-> owns DNS after nameserver change

Cloudflare Pages
-> serves the static site
```

Do not transfer the domain just to host the static site.

Domain setup flow:

```text
add domain to Cloudflare
Cloudflare assigns two nameservers
set those nameservers at CanSpace
wait for Cloudflare zone activation
add custom domain to the Pages project
```

Recommended public domain setup:

```text
apex domain
www subdomain
```

Add the custom domain from the Pages project after the `*.pages.dev` deployment works.

## Analytics

Cloudflare Web Analytics can be used for browser-side visit/page-view tracking.

Use Web Analytics for metrics closer to human visits:

```text
page views
visitors
top pages
referrers
countries
devices/browsers
```

Cloudflare zone/request analytics are not the same as human page visits. They may include bots, crawlers, asset requests, and edge traffic.

## Code map

Content and schema:

```text
web-astro/src/content.config.ts
-> Astro content collection schema for devlog entries

web-astro/src/content/devlog/
-> live MDX devlog content entries

web-astro/src/content/homepageContent.ts
-> homepage content type and normalization

web-astro/src/content/archiveContent.ts
-> archive content type and normalization
```

Routes:

```text
web-astro/src/pages/index.astro
-> loads latest devlog entry and renders homepage

web-astro/src/pages/archive/index.astro
-> loads devlog collection and renders archive

web-astro/src/pages/devlog/[slug].astro
-> generates static per-post devlog pages
```

Owned wrappers and rendering:

```text
web-astro/src/components/Homepage.tsx
-> maps devlog content into Plasmic overrides, media props, and MDX body placement

web-astro/src/components/HomepageClientMount.tsx
-> mounts Homepage client-side

web-astro/src/components/Archive.tsx
-> maps archive data into generated Archive layout

web-astro/src/components/ArchiveClientMount.tsx
-> mounts Archive client-side

web-astro/src/components/markdown/MarkdownText.tsx
-> renders Markdown-capable fields/body where used
```

Owned presentation components:

```text
web-astro/src/components/media/CrtMediaFrame.tsx
-> CRT media frame behavior and media-mode rendering

web-astro/src/components/card/CardFrame.tsx
-> reusable card frame behavior
```

Generated Plasmic output:

```text
web-astro/src/components/plasmic/space_rocks_devlog/
-> generated layout components and CSS modules

web-astro/src/plasmic/
-> generated Plasmic support files

web-astro/plasmic.json
web-astro/plasmic.lock
-> Plasmic project/codegen metadata
```

Public assets:

```text
web-astro/public/assets/ui/
-> UI art assets

web-astro/public/media/devlog/<slug>/
-> per-entry devlog media
```

Temporary parked host:

```text
tools/parked-plasmic-next-host/
-> Plasmic Studio host support, not the deployed static site
```

## Tests

No dedicated automated test suite is currently documented for the devlog static site.

Current verification is build and runtime smoke testing.

Minimum local verification from `web-astro/`:

```text
npm run build
npm run dev
```

Smoke checks:

```text
homepage renders the latest live devlog
archive renders only intended live posts
archive links resolve to generated devlog routes
devlog post route renders
hero media renders
article media renders
MDX body renders immediately after the article media frame
intro/progress/utility/finePrint fields render expected Markdown/plain text
media frame controls appear and work where applicable
no test posts appear in the published content collection
```

Cloudflare verification:

```text
Cloudflare Pages build succeeds
build uses web-astro as root directory
build command is npm run build
output directory is dist
deployed pages.dev URL works
custom domain resolves after DNS activation
```

If the deployed site differs from local output, compare:

```text
local git HEAD
Cloudflare build commit SHA
Cloudflare project root directory
Cloudflare production branch
committed content files
generated Plasmic output
```

## Related docs

* [Web](./!INDEX.md)
* [Plasmic / Astro Workflow](plasmic-astro-workflow.md)
* [CRT Media Frame](crt-media-frame.md)
* [Website and Web Presence](../../domains/web/website-and-web-presence.md)
* [Future Devlog Static Site Planning](../../planning/web/devlog-static-site.md)
* [Future Website and Web Presence Plan](../../planning/domains/web/website-and-web-presence.md)
* [Documentation Policy](../../documentation-policy.md)
* [Documentation Procedure](../../documentation-procedure.md)

## Notes

The current static devlog is intentionally smaller than the future website. It is the present public publishing path, not the final product web surface.

The MDX switch makes the article body flexible while keeping key page chrome and reusable content fields in frontmatter. This is the current contract: frontmatter controls the page structure, and the MDX body owns the article prose after the article media frame.

The parked Next/Plasmic host is a known source of confusion because it can render stale code-component behavior while the deployed Astro site uses the current `web-astro` implementation. Check host/component drift before making layout changes based only on Plasmic Studio canvas differences.

Templates and test posts should not remain in the active content collection path as publishable `.md` or `.mdx` files.
