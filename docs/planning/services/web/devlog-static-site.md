# Devlog Static Site

Parent index: [Web](./!INDEX.md)

## Purpose

This document plans the V0 Space Rocks devlog static site implementation.

It defines the intended code structure, static route shape, content source model, Plasmic/Next.js responsibilities, generated outputs, basic analytics, static hosting expectations, quality gates, and boundaries for the initial public web presence.

This is a planning service document. It is not a launch website plan, CMS plan, account portal plan, commerce plan, API contract, or hosting/deployment policy.

## Overview

The V0 devlog site is the first public Space Rocks web presence.

It is a static, devlog-first website with:

```text
index page
devlog archive
individual devlog post pages
404 page
sitemap.xml
rss.xml and/or feed.xml
static assets/media
basic page-view analytics
```

The site should be built with a simple static generation path, likely using Next.js. Plasmic may own canonical visual layouts where practical. The generated site should remain compatible with static hosting, probably Netlify.

The V0 site does not need runtime account state, a CMS backend, comments, newsletter backend, API-backed publishing, leaderboards, profiles, support portal, appeal portal, commerce portal, or admin portal.

## Current status

Active planning.

The web implementation root is not established yet.

The planned V0 direction is stable:

```text
Plasmic-designed
Next.js-built
static output
devlog-first
repo-owned content
probably Netlify-hosted
```

## Expected ownership

The devlog static site owns:

```text
V0 public web implementation
static route structure
devlog post rendering
devlog archive rendering
homepage/devlog landing rendering
404 page
static metadata generation
sitemap/feed generation
static asset presentation
basic public-site analytics integration
Plasmic layout integration where practical
```

It does not own:

```text
launch website account portal
direct purchase
Steam ownership verification
download entitlement
CMS backend
comments
leaderboards
public profiles
support/appeal portal
admin portal
payment provider integration
hosting/deployment policy
```

`Website And Web Presence` owns the web-domain product plan.

`Build Release And Environment Matrix` owns hosting, deployment, and release mechanics.

Future account, purchase, ownership, profile, leaderboard, support, and community surfaces belong to the launch/intermediate website plan, not V0.

## Planned code root

Preferred implementation root:

```text
services/web/
```

Likely structure:

```text
services/web/
- package.json
- next.config.*
- tsconfig.json
- src/
- content/
- public/
- plasmic/
```

Suggested folders:

```text
services/web/src/app/
-> Next.js route files and page composition

services/web/src/components/
-> hand-coded reusable components and glue components

services/web/src/content/
-> content loaders, frontmatter parsing, post indexing

services/web/content/devlog/
-> repo-owned devlog source posts

services/web/public/
-> static assets emitted directly as public files

services/web/plasmic/
-> Plasmic project exports, generated components, or canonical layout source artifacts where practical
```

Exact file names are implementation decisions.

The root should stay small and clear enough that V0 can be deleted, replaced, or expanded without contaminating later account/platform website work.

## Responsibilities

The V0 site should provide:

```text
public homepage/devlog landing
devlog archive
static post pages
static media/assets
404 page
sitemap.xml
rss.xml and/or feed.xml
page titles and metadata
Open Graph/social preview metadata
basic analytics
static-host compatible build output
```

The V0 site should support local development, static build verification, and deployment to a static host.

The V0 site should keep public content in repo-owned source files.

The V0 site should be easy to expand into an intermediate or launch website without making V0 account-aware.

## Does not own

V0 does not own:

```text
account login
account portal
Steam account linking
ownership verification
direct-download license state
direct purchase
payment checkout
purchase receipts
downloads
support tickets
comments
CMS runtime
leaderboards
public profiles
match-result sharing
invite landing
appeals
admin tools
platform API behavior
```

The static site may link to placeholders or future pages only when the copy makes it clear the feature is not available yet.

## Domain roles

The devlog static site participates in the website domain as the V0 public surface.

Role split:

```text
V0 devlog static site
-> first public website implementation

Website And Web Presence
-> product/domain direction and launch-shape web plan

Plasmic
-> visual layout and page composition where practical

Next.js
-> routing, static generation, metadata generation, content loading, build output

Repository content files
-> devlog post source and public content source

Static host
-> serves generated files

Analytics provider
-> records basic public-site engagement
```

No backend API authority is required for V0.

## Routes and public files

Planned V0 routes:

```text
/
-> devlog-first home page

/devlog/
-> devlog archive

/devlog/<slug>/
-> individual devlog post page

/404.html
-> static not-found page
```

Generated public files:

```text
/sitemap.xml
/rss.xml
/feed.xml, if cheap to generate alongside rss.xml
```

Static asset path:

```text
/assets/
-> images, screenshots, gifs, video thumbnails, icons, and other public media
```

`robots.txt` is not required for V0. The site is intended to be publicly crawlable.

Route names may change later, but V0 should avoid route churn after posts are published. Devlog URLs should be treated as stable once public.

## Protocols and APIs

V0 has no application API.

The public surface is static HTTP file serving:

```text
browser
-> static host
-> generated HTML/CSS/JS/assets/XML
```

The site is for public read-only content. Anyone can consume the generated pages, assets, sitemap, and feed.

Authority behind the site is repo-owned source content and the static build process.

Data crossing the boundary is public website content only:

```text
HTML pages
CSS/JS bundles
images/media
sitemap XML
RSS/feed XML
analytics page-view events
```

The V0 site explicitly does not own account, purchase, entitlement, profile, leaderboard, support, or moderation APIs.

Analytics may call a provider endpoint, but analytics is not authoritative for site rendering, content state, ownership, or product state.

## Content model

Devlog posts should be repo-owned source files.

Preferred source format:

```text
Markdown
```

MDX is not required for V0. It may be introduced later only if posts need embedded interactive React components.

Minimum post metadata:

```text
title
slug
date
summary
```

Optional metadata:

```text
tags
hero_image
author
draft
updated_at
social_image
```

Recommended post source shape:

```text
services/web/content/devlog/<date-or-slug>.md
```

Posts should generate stable URLs:

```text
/devlog/<slug>/
```

Archive behavior:

```text
newest-first listing
post title
post date
summary
link to full post
```

Draft posts should not appear in production builds.

## Plasmic layout model

Plasmic may own canonical visual layouts where practical.

Expected use:

```text
Plasmic
-> page layout
-> visual composition
-> responsive layout reference
-> reusable visual sections where useful
```

Repo/Next.js should own:

```text
routing
content loading
devlog post indexing
static generation
metadata generation
sitemap/feed generation
analytics wiring
build/deploy behavior
```

Plasmic sources or exports should be stored under the web service root where practical, likely under:

```text
services/web/plasmic/
```

Generated Plasmic code should not be edited casually by hand unless the chosen Plasmic workflow expects that.

Hand-coded glue should remain outside generated Plasmic output.

## Static generation

V0 should build without a runtime server requirement.

Build output should be compatible with static hosting.

The build should generate:

```text
homepage
devlog archive
all published devlog post pages
404 page
sitemap.xml
rss.xml and/or feed.xml
static assets
```

The static build should fail when required post metadata is missing or invalid.

The build should not require:

```text
database
CMS backend
account API
commerce API
runtime server
private environment secrets for public content generation
```

Analytics configuration may use an environment value if the chosen provider needs one, but missing analytics config should not block local rendering unless explicitly required by the build profile.

## Data ownership

V0 owns public content source and static site configuration.

Owned data:

```text
devlog Markdown source files
post metadata/frontmatter
static page copy
static assets/media
site metadata
RSS/feed metadata
sitemap generation config
analytics site identifier/config where needed
Plasmic layout exports/sources where practical
```

Not owned data:

```text
account data
purchase data
entitlement data
Steam ownership data
download authorization data
leaderboard data
profile data
support case data
moderation data
CMS runtime data
```

## Metadata and shareability

V0 should include basic public-site metadata.

Required:

```text
page title
page description
Open Graph title
Open Graph description
Open Graph image where available
favicon/app icon
canonical/stable URL behavior where practical
```

The goal is simple discoverability and usable link previews in Discord and other apps.

SEO for V0 means making public pages machine-readable and shareable, not trying to manipulate search ranking.

## Analytics

V0 should include basic engagement analytics.

Minimum desired metrics:

```text
page views
post views
referrers if available
basic traffic trends
```

Analytics provider is not decided.

Likely options:

```text
Netlify analytics
simple third-party analytics
later provider chosen by cost/ease/effectiveness
```

Analytics must not:

```text
block rendering
be required for static build success
store sensitive account/payment data
act as product telemetry
act as gameplay telemetry
become authoritative state
```

V0 has no accounts, purchases, or user identity, so analytics should remain public-site engagement only.

## XML outputs

V0 should generate XML outputs if tooling makes them cheap.

Required or preferred:

```text
sitemap.xml
rss.xml
feed.xml, if easy to generate alongside rss.xml
```

Purpose:

```text
sitemap.xml
-> helps crawlers discover public pages

rss.xml / feed.xml
-> lets users and feed readers subscribe to devlog posts
```

These are static generated files. They should not require a backend.

## 404 page

V0 should include a static 404 page.

The 404 page should:

```text
match the site visual style
link back to /
link to /devlog/
avoid implying account/platform features exist
```

## Assets and media

Static assets should live under the web service root.

Likely path:

```text
services/web/public/assets/
```

Assets may include:

```text
screenshots
gifs
video thumbnails
logos
icons
Open Graph images
post hero images
```

Large video hosting strategy is not part of V0 unless needed. V0 can link to external video hosting if that is simpler.

Asset filenames should be stable once referenced by public posts.

## Styling and responsive layout

The V0 site should support common desktop and mobile layouts.

Minimum layout expectations:

```text
readable mobile layout
readable desktop layout
usable tablet/narrow layout
no critical navigation hidden behind broken hover behavior
reasonable contrast
readable devlog post typography
```

Exact visual design belongs to Plasmic/design work.

## Accessibility

V0 should include basic accessibility hygiene:

```text
semantic headings where practical
keyboard-usable links/navigation
alt text for important images
readable contrast
clear focus behavior where applicable
no critical content available only through animation
```

This should stay lightweight but should not be ignored.

## Failure modes

V0 failure modes are mostly build-time or static-hosting issues.

Planned failure handling:

```text
invalid post metadata
-> build should fail or clearly report the bad post

missing post asset
-> build should fail or visibly report the missing asset

draft post in production
-> draft should be excluded

missing analytics config
-> site should still render unless the selected provider requires explicit production config

unknown route
-> 404 page

missing Plasmic export/source
-> build should fail clearly if the page depends on it
```

The site should not have runtime degraded account/API states because V0 has no account/API dependency.

## Code map

Planned paths:

```text
services/web/
-> web service root

services/web/src/app/
-> Next.js routes and pages

services/web/src/components/
-> hand-coded components

services/web/src/content/
-> content loading and post indexing

services/web/content/devlog/
-> devlog source posts

services/web/public/
-> static assets and public files

services/web/public/assets/
-> site/post/media assets

services/web/plasmic/
-> Plasmic exports or source artifacts where practical
```

Planned key route files may include:

```text
services/web/src/app/page.*
-> homepage/devlog landing

services/web/src/app/devlog/page.*
-> devlog archive

services/web/src/app/devlog/[slug]/page.*
-> individual devlog post

services/web/src/app/not-found.*
-> 404 behavior if using Next.js app router
```

Planned content/build helpers may include:

```text
post loader
frontmatter parser
archive index builder
sitemap generator
RSS/feed generator
metadata helper
```

Exact filenames depend on the chosen Next.js structure and tooling.

## Tests and verification

V0 should have lightweight verification rather than a heavy test suite.

Minimum verification:

```text
install/build succeeds
static export succeeds
homepage renders
devlog archive renders
all published posts render
404 page exists
sitemap.xml exists
rss.xml and/or feed.xml exists
internal links resolve
required metadata exists
draft posts are excluded from production
missing required post fields fail clearly
basic responsive smoke check
Open Graph metadata exists for public pages
analytics script/config is present in production build when configured
```

Possible commands, exact names TBD:

```text
npm run lint
npm run build
npm run check
npm run test
```

If link checking is added, it should verify internal links and generated devlog URLs.

## Implementation sequence

1. Establish `services/web/` as the web service root.
2. Create the initial Next.js/static-generation project.
3. Add basic site metadata and app shell.
4. Add Plasmic integration or placeholder layout boundary.
5. Add homepage/devlog landing route.
6. Add repo-owned devlog Markdown content folder.
7. Add frontmatter parsing and post index loading.
8. Add individual devlog post route generation.
9. Add devlog archive route.
10. Add static 404 page.
11. Add static assets folder and first image/media handling.
12. Add Open Graph/favicon/basic metadata handling.
13. Add sitemap generation.
14. Add RSS/feed generation.
15. Add basic analytics integration.
16. Add build/lint/check scripts.
17. Add internal link and metadata verification where practical.
18. Add responsive/accessibility smoke checks where practical.
19. Confirm static output can be hosted without a runtime server.
20. Update current service docs if the implementation becomes current.

## Open decisions

Implementation details still open:

```text
exact Next.js version and router structure
exact static export mode
exact Plasmic integration mode
exact Plasmic export/source path
exact Markdown parser/frontmatter library
whether feed.xml is generated in addition to rss.xml
exact analytics provider
exact package manager
exact CSS/styling approach
exact local development command names
exact static host config
exact asset naming conventions
exact first post metadata shape beyond required fields
```

These are not open direction questions:

```text
V0 is static.
V0 is devlog-first.
V0 has an index page.
V0 has a devlog archive.
V0 has individual post pages.
V0 has a 404 page.
V0 should generate sitemap/feed XML.
V0 should include basic analytics.
V0 has no accounts, CMS backend, comments, commerce, profiles, leaderboards, support portal, or platform API dependency.
Plasmic may own canonical layouts where practical.
Next.js/static generation is the expected implementation direction.
Netlify is the likely hosting target, but deployment mechanics belong elsewhere.
```

## Related docs

* [Web Planning](../../domains/web/website-and-web-presence.md)
* [Web Services](./!INDEX.md)
* [Build Release And Environment Matrix](../../domains/technical/build-release-and-environment-matrix.md)
* [API Product Surface](../../protocol/api-product-surface.md)
* [Current API Product Surface](../../../protocol/api-product-surface.md)

## Notes

This document plans the V0 devlog static site implementation only.

The launch website is larger and includes account, ownership, commerce, CMS, support, analytics, and platform/API-backed surfaces. Those belong to the broader Website And Web Presence plan and future web service planning, not this V0 static-site plan.

If V0 implementation becomes current, current facts should move from `docs/planning/services/web/` into `docs/services/web/` using the normal documentation procedure.
