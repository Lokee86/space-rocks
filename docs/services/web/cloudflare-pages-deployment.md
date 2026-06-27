# Cloudflare Pages Deployment

Parent index: [Web](./!INDEX.md)

## Purpose

This document is the canonical Cloudflare Pages deployment reference for the Space Rocks devlog static site.

It records the product choice, repository/build settings, DNS ownership model, CanSpace registrar flow, custom-domain setup, analytics expectations, and common deployment failure modes.

## Overview

The current public Space Rocks web surface is the Astro static devlog site under:

```text
web-astro/
```

The site deploys to Cloudflare Pages.

Cloudflare Pages owns static hosting for the generated site. Cloudflare DNS owns DNS after the domain nameservers are pointed at Cloudflare. CanSpace remains the domain registrar.

The intended ownership split is:

```text
CanSpace
-> domain registration

Cloudflare DNS
-> authoritative DNS after nameserver cutover

Cloudflare Pages
-> static site hosting and deploys

web-astro/
-> deployable Astro site root
```

Use Cloudflare Pages, not Cloudflare Workers, for this site.

## Responsibilities

Cloudflare Pages deployment owns:

```text
GitHub-connected static deploys
Astro build execution
static asset publication
pages.dev preview/production URL
custom-domain attachment
production branch deployment
deployment history
basic deployment logs
```

This deployment path also records:

```text
required build settings
DNS/custom domain sequence
Cloudflare UI gotchas
analytics setup expectations
stale-build troubleshooting
```

## Does not own

Cloudflare Pages deployment does not own:

```text
devlog content schema
MDX post format
Plasmic Studio layout source
generated Plasmic files
CRT media-frame behavior
account portal behavior
game services
player-data services
commerce
authentication
runtime multiplayer services
```

Those are documented elsewhere.

## Current deployment target

Repository:

```text
Lokee86/space-rocks
```

Production branch:

```text
main
```

Deployable root:

```text
web-astro
```

Build command:

```text
npm run build
```

Build output directory:

```text
dist
```

Cloudflare Pages should build the Astro site from `web-astro/`, not from the repository root.

## Cloudflare product choice

Use:

```text
Cloudflare Pages
```

Do not use:

```text
Cloudflare Workers
```

The wrong Cloudflare flow shows Worker-specific deploy settings such as:

```text
npx wrangler deploy
```

That is not the deploy path for the current Astro static site.

If the Cloudflare dashboard shows a Worker deploy command, Worker bindings, Worker routes, or Wrangler-only configuration, the project is in the wrong setup flow.

The correct project should expose static-site build settings such as:

```text
framework preset
build command
build output directory
root directory
production branch
```

## Cloudflare UI path

Cloudflare’s dashboard can make Pages hard to find.

The correct flow is:

```text
Cloudflare dashboard
-> Compute
-> Workers & Pages
-> Create
-> Pages
-> Connect to Git / Import existing Git repository
-> select Lokee86/space-rocks
-> configure build settings
```

If the UI starts from a “Create a Worker” screen, use the Pages-specific link or prompt on that screen, usually phrased like:

```text
Looking to deploy Pages? Get started
```

Do not continue through the main Worker GitHub flow for this site.

## Build settings

Use these settings in Cloudflare Pages:

```text
Project name: space-rocks-devlog
Production branch: main
Root directory: web-astro
Build command: npm run build
Build output directory: dist
```

If Cloudflare offers a framework preset, use:

```text
Astro
```

If Astro is not available, use no preset/manual settings and set the build command/output directory directly.

## Pre-deploy checklist

Before triggering or relying on a Cloudflare deploy, ensure the local repo state is committed and pushed.

From `web-astro/`, run Plasmic codegen when Plasmic Studio changed:

```text
npx @plasmicapp/cli@0.1.365 sync -p uNJepqX5kmDcUn9dDb3UVD --yes
```

Then run the Astro build locally:

```text
npm run build
```

From repo root, confirm and push:

```text
git status
git add web-astro docs
git commit -m "<appropriate deployment/docs message>"
git push
```

Cloudflare builds from the Git commit it fetches. It does not build the local working tree unless those changes are committed and pushed.

## First deploy flow

Use this order:

```text
1. Commit and push the deployable web-astro state.
2. Create the Cloudflare Pages project.
3. Configure root/build/output settings.
4. Deploy to the generated *.pages.dev URL.
5. Verify the pages.dev deployment.
6. Attach custom domain.
7. Change DNS/nameservers if needed.
8. Enable analytics if desired.
```

Do not begin DNS/domain cutover until the `*.pages.dev` deployment works.

## Smoke checks after Pages deploy

After the first successful deploy, check:

```text
/
```

Expected:

```text
latest live devlog renders
hero media renders
article media renders
progress cards render
utility panel renders
fine print renders
```

Check archive:

```text
/archive/
```

Expected:

```text
archive lists only intended live posts
archive links resolve
no test posts are visible
```

Check post route:

```text
/devlog/<slug>/
```

Expected:

```text
individual devlog page renders
MDX body appears after article media frame
media paths resolve
no stale test content appears
```

## DNS and registrar model

The current registrar is:

```text
CanSpace
```

Do not transfer the domain just to use Cloudflare Pages.

The intended model is:

```text
CanSpace remains registrar.
Cloudflare becomes authoritative DNS.
Cloudflare Pages serves the static site.
```

Cloudflare assigns two nameservers when the domain is added to Cloudflare DNS. Those nameservers must be set at CanSpace.

## Domain setup flow

In Cloudflare:

```text
Account home
-> Domains
-> Add/onboard domain
-> enter apex domain
-> choose plan
-> let Cloudflare scan existing records
-> continue until Cloudflare provides two nameservers
```

In CanSpace:

```text
Domains / My Domains
-> select domain
-> Nameservers / Change Nameservers
-> replace current nameservers with the two Cloudflare nameservers
-> save
```

Back in Cloudflare:

```text
wait for domain activation
```

Then attach the domain to Pages:

```text
Compute
-> Workers & Pages
-> Pages project
-> Custom domains
-> Set up a domain
```

Recommended domain setup:

```text
apex domain
www subdomain
```

Example:

```text
spacerocks.ca
www.spacerocks.ca
```

Use the actual production domain when known.

## DNS notes

When moving DNS to Cloudflare, preserve any existing records that still matter.

Before changing nameservers, check for records such as:

```text
A
AAAA
CNAME
MX
TXT
SPF
DKIM
DMARC
verification records
old hosting records
```

If the domain has email, do not delete mail-related records without confirming they are obsolete.

Cloudflare must have the records needed for any domain services that should continue working after the nameserver cutover.

## Custom domain notes

Attach the custom domain from the Cloudflare Pages project after the project has a working `*.pages.dev` deployment.

If attaching the apex domain fails, check:

```text
domain is active in Cloudflare DNS
nameservers at CanSpace match Cloudflare-assigned nameservers
Pages project exists and has a successful deployment
no conflicting DNS records are blocking the Pages custom domain
```

If using both apex and `www`, configure both through the Pages custom-domain flow.

## Analytics

Cloudflare Web Analytics can be enabled for browser-side site metrics.

Use Web Analytics for visit-like metrics such as:

```text
page views
visitors
top pages
referrers
countries
devices
browsers
```

Cloudflare zone/request analytics are not the same as human visits. Broader traffic analytics can include:

```text
bots
crawlers
asset requests
edge requests
non-page traffic
```

Use Web Analytics when the question is “how many people viewed the site?”

## Common failure modes

## Wrong Cloudflare product

Symptom:

```text
Cloudflare asks for npx wrangler deploy
```

Cause:

```text
Worker project flow was selected instead of Pages
```

Fix:

```text
go back and create a Pages project from Git
```

## Missing build settings

Symptom:

```text
no root directory / build output / Astro build settings visible
```

Cause:

```text
wrong Cloudflare flow, usually Worker Git setup
```

Fix:

```text
use Pages -> Connect to Git
```

## Cloudflare builds stale content

Symptom:

```text
Cloudflare build includes deleted or test devlog entries
```

Likely causes:

```text
changes were not committed
changes were not pushed
Cloudflare is building a different branch
Cloudflare is building an older commit
test files remain as .md or .mdx under web-astro/src/content/devlog/
root directory is wrong
```

Check:

```text
Cloudflare build commit SHA
local git rev-parse HEAD
Cloudflare production branch
git status
web-astro/src/content/devlog/
```

## Cloudflare builds from wrong root

Symptom:

```text
build fails to find Astro project files
wrong package.json used
wrong output produced
```

Cause:

```text
Root directory was not set to web-astro
```

Fix:

```text
set Root directory / Path to web-astro
```

## Build succeeds but page is wrong

Check:

```text
generated Plasmic output committed
Cloudflare build commit SHA
content files committed
media files committed
Cloudflare cache/deployment version
wrong branch deployed
```

## Images fail in deployed site

Check:

```text
file exists under web-astro/public/
frontmatter path starts at public root
case-sensitive filename match
spaces/special characters in filename
articleImages list uses YAML hyphens
Cloudflare deployment includes the file
```

Correct path pattern:

```text
web-astro/public/media/devlog/<slug>/<file>.png
```

Correct frontmatter path pattern:

```text
/media/devlog/<slug>/<file>.png
```

## MDX/content build fails

Check:

```text
frontmatter opens with ---
frontmatter closes with ---
YAML block scalar content is indented
YAML arrays use -
Markdown * bullets are not used as YAML list markers
finePrint is top-level, not nested under utilityText
retired fields are not required by schema
```

## Current build verification

Minimum local verification:

```text
cd web-astro
npm run build
```

Minimum Cloudflare verification:

```text
Cloudflare Pages build succeeds
pages.dev deployment renders homepage
archive renders
post route renders
custom domain resolves after DNS activation
```

Compare local and Cloudflare commits when output differs:

```text
git rev-parse HEAD
Cloudflare build commit SHA
```

The SHAs should match the intended deployed commit.

## Code map

Deployment-relevant files:

```text
web-astro/package.json
-> build scripts and dependencies

web-astro/package-lock.json
-> locked dependency graph

web-astro/astro.config.mjs
-> Astro static-site configuration

web-astro/src/content.config.ts
-> content collection schema

web-astro/src/pages/
-> generated static route source

web-astro/src/content/devlog/
-> live devlog entries

web-astro/public/
-> public static assets

web-astro/plasmic.json
web-astro/plasmic.lock
web-astro/src/components/plasmic/
-> generated Plasmic output that must be committed after codegen
```

Cloudflare should not need files outside `web-astro/` to build the static site, except for repository context and Git metadata.

## Related docs

* [Web](./!INDEX.md)
* [Devlog Static Site](devlog-static-site.md)
* [Plasmic / Astro Workflow](plasmic-astro-workflow.md)
* [CRT Media Frame](crt-media-frame.md)
* [Website and Web Presence](../../domains/web/website-and-web-presence.md)
* [Future Devlog Static Site Planning](../../planning/web/devlog-static-site.md)

## Notes

The Cloudflare deployment is intentionally static. It does not require a game server, player-data service, Rails API, OAuth service, or database to render the current devlog.

The Pages deployment should stay boring: build the Astro site from `web-astro`, publish `dist`, attach the domain after the preview deployment works, and keep CanSpace as registrar unless there is a separate reason to transfer the domain.
