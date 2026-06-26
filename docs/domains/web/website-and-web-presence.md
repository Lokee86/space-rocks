# Website And Web Presence

Parent index: [Web](./!INDEX.md)

## Purpose

This document describes the current Space Rocks website and public web-presence domain.

It defines the implemented V0 public website surface, participating systems, authority boundaries, public-content flow, inputs and outputs, and current out-of-scope areas.

## Overview

The current Space Rocks website is the Astro static devlog site. It is the implemented V0 public web presence for the project.

The current site provides:

```text
public homepage
devlog archive
individual devlog post pages
static public media
project/repository navigation
Plasmic-based visual layout
owned wrapper/source behavior
```

The site is content-driven and repo-owned. Devlog Markdown entries provide the current public content. Astro loads those entries and generates public routes. React wrappers map content into Plasmic-generated layouts. Owned components provide custom behavior such as media presentation.

The current domain shape is:

```text
repo-owned devlog content
-> Astro static site
-> owned React wrappers
-> Plasmic-generated visual layout
-> public static website
```

The current website is not the launch website target. Account, commerce, Steam ownership, download access, CMS, comments, leaderboards, profiles, support, and admin surfaces remain outside the current implemented domain.

## Participating systems

Current participating systems:

```text
Astro static devlog site
-> current deployable public website

Repo-owned devlog content
-> source of public devlog pages and homepage content

Owned React wrappers
-> map content into layouts and attach custom behavior

Plasmic visual/layout source
-> visual page structure and responsive layout source

Generated Plasmic output
-> generated layout code consumed by owned wrappers

CRT media frame
-> framed image and video presentation

Static public assets
-> UI art, media, icons, and post-specific media

Temporary parked Next/Plasmic host
-> Plasmic Studio workflow support only

Static host
-> serves generated site output
```

The current web domain also links out to external project/community surfaces where configured, such as the project repository.

## Authority boundaries

The website domain owns public presentation and navigation for the current V0 web surface.

Authority split:

```text
Repo-owned devlog content
-> authoritative source for current public devlog entries

Astro static site
-> authoritative route/build surface for the current deployable site

Owned wrapper/source files
-> authoritative home for custom web behavior and content wiring

Plasmic visual/layout source
-> authoritative visual/layout source where used

Generated Plasmic files
-> layout output, not customization authority

CRT media frame
-> media presentation behavior inside the web site

Static host
-> file serving only; not content authority

Temporary parked Next/Plasmic host
-> Studio workflow aid only; not the deployed website
```

Custom behavior must go in owned wrapper/source files. Generated Plasmic files are not customization surfaces.

The current website does not own backend authority. It does not decide account state, purchase state, entitlement state, player identity, leaderboard ranking, moderation state, support status, or gameplay state.

## Flow summary

The current public web flow is static and content-driven.

Homepage flow:

```text
devlog content collection
-> newest entry selected by date
-> entry data passed into the homepage wrapper
-> wrapper maps content into Plasmic layout overrides
-> Astro/React renders the public homepage
```

Archive flow:

```text
devlog content collection
-> entries sorted newest-first
-> archive entry summaries generated
-> archive wrapper maps entries into Plasmic layout
-> public archive page renders
```

Devlog post flow:

```text
devlog Markdown entry
-> static route generated from entry id
-> entry data passed into homepage-style devlog layout
-> public devlog post page renders
```

Media flow:

```text
devlog frontmatter media fields
-> owned wrapper maps media fields into CrtMediaFrame props
-> CrtMediaFrame presents image galleries or YouTube media
-> static assets or embedded YouTube media render inside the frame
```

The public reader receives generated pages and public assets. No account session or platform API is required for the current V0 site.

## Inputs and outputs

Inputs:

```text
repo-owned devlog Markdown entries
devlog frontmatter fields
post-specific public media assets
shared public UI assets
Plasmic visual/layout source
generated Plasmic layout output
owned wrapper/source files
Astro route and content configuration
```

Public outputs:

```text
homepage
devlog archive page
individual devlog post pages
static HTML
CSS and JavaScript bundles
public UI assets
post-specific media assets
YouTube iframe embeds where configured
repository/project links where configured
```

Internal implementation outputs:

```text
Astro static build output
generated Plasmic React components
generated Plasmic CSS
normalized homepage content objects
normalized archive content objects
```

The current site does not output account-specific, purchase-specific, entitlement-specific, or player-specific pages.

## Current public surfaces

Current implemented public surfaces:

```text
/
-> homepage using the newest devlog entry

/archive/
-> devlog archive list

/devlog/<slug>/
-> individual static devlog post page
```

Current content surfaces:

```text
latest devlog entry presentation
devlog archive summaries
individual devlog article content
hero media
article media
side-card status copy
utility/fine-print copy
```

Current media surfaces:

```text
post-specific image galleries
YouTube video embeds
CRT-styled media frame
shared UI frame assets
```

## Out of scope

The current website domain does not include:

```text
account portal
sign in / sign out
Steam account linking
Steam ownership verification
direct purchase
payment checkout
purchase receipts
download entitlement
account-gated downloads
CMS backend
comments
newsletter backend
leaderboards
public player profiles
support portal
appeal portal
admin portal
moderation tools
user-generated content
platform API behavior
gameplay runtime behavior
```

These may belong to future website planning or other platform/domain/service docs when implemented. They should not be described as current behavior in this document.

## Related docs

* [Web Services](../../services/web/!INDEX.md)
* [Devlog Static Site](../../services/web/devlog-static-site.md)
* [Plasmic / Astro Workflow](../../services/web/plasmic-astro-workflow.md)
* [CRT Media Frame](../../services/web/crt-media-frame.md)
* [Future Website and Web Presence Plan](../../planning/domains/web/website-and-web-presence.md)
* [Future Interactive Website](../../planning/web/stubs/interactive-website.md)
* [Documentation Policy](../../documentation-policy.md)
* [Documentation Procedure](../../documentation-procedure.md)

## Notes

This is the current domain document for the implemented V0 website surface. It should stay focused on current cross-system website behavior.

Implementation details belong in web service docs. Future launch website plans belong in planning docs.

The current V0 website and the future launch website are separate targets. The current site is static and devlog-first; the future launch site may become a larger product/platform surface when those systems are implemented.
