# Website And Web Presence

Parent index: [Web](./!INDEX.md)

## Purpose

This document plans the Space Rocks website and public web presence.

It defines the staged web product surface, launch website requirements, account and ownership portal expectations, public content model, community handoff, analytics expectations, CMS direction, Plasmic/Next.js workflow, and relationship to API-backed platform systems.

This is a domain planning document. It is not a deployment plan, API contract, payment-provider implementation plan, legal policy document, CMS implementation guide, or replacement for Discord/community systems.

## Overview

The Space Rocks website starts simple but should launch as a full product and platform website.

The planned shape is:

```text
V0 static devlog-first site
-> intermediate public/product site as needed
-> launch website with public content, account portal, ownership, commerce, analytics, and support surfaces
```

The launch website has three major areas:

```text
public content and marketing site
account, ownership, and commerce portal
community, engagement, and operational support
```

V0 is intentionally smaller than the launch target. V0 exists to establish a public web presence and development log without blocking on account, commerce, CMS, or platform API work.

The launch website should be treated as a product/platform surface. It should present the game, support direct purchase, support ownership verification, expose account-owned download access, provide useful public content, measure engagement, and route community/support flows safely.

## Current Status

Active planning.

Current repository planning includes:

```text
website-and-web-presence.md as the web domain plan
service-level web stubs for a static devlog site and future interactive website
planned API Product Surface entry for the website account portal
leaderboard planning that requires website leaderboard/profile surfaces
social/community planning that reserves website profile, invite, report, appeal, and community hub surfaces
abuse/enforcement planning that expects website/email user-facing enforcement communication
build/release planning that owns deployment and hosting mechanics
```

No production website implementation is currently assumed by this document.

## Ownership Boundary

This document owns planning for:

```text
public website product surface
website stages
V0 static devlog-first site shape
launch homepage/product site shape
devlog and public update surfaces
roadmap/status website presentation
lore and deeper public content surfaces
media/gallery/press-facing surfaces
account portal presentation
purchase, ownership, license, and download presentation
Steam-link ownership verification presentation
support/help entry points
selected developer/team post comments if added
website analytics and engagement measurement requirements
SEO/shareability expectations
CMS-backed public content expectations
Plasmic/Next.js website workflow expectations
website quality gates and failure-state expectations
```

This document does not own:

```text
API endpoint design
OpenAPI schemas
account authentication mechanics
provider OAuth implementation
payment-provider integration internals
purchase validation
receipt authority
refund authority
entitlement authority
Steam ownership verification internals
download artifact building
hosting/deployment mechanics
leaderboard formulas
ranking catalog ownership
social graph implementation
Discord SDK/plugin implementation
abuse/enforcement decisions
legal policy authorship
analytics provider implementation
CMS implementation details
```

Account and Identity Systems owns account identity, login policy, linked providers, provider identity, account lifecycle, and durable account state.

Shop, Commerce, And Economy owns pricing, payment validation, receipts, refunds, entitlements, commerce policy, and future real-money transaction boundaries.

Build Release And Environment Matrix owns hosting, deployment, release packaging, environment matrix, and build artifact mechanics.

API Product Surface owns backend/account/product API surface mapping. Exact HTTP shape belongs to protocol contracts and OpenAPI when implemented.

Social And Community Systems owns Discord-first social integration, social facts, Discord/community handoff, blocks/mutes, invites, and website social requirements.

Abuse And Enforcement Admin owns moderation, report case handling, enforcement, appeals, evidence, audit history, and admin enforcement actions.

Leaderboards And Rankings owns ranking definitions, board lifecycle, ranking privacy, and leaderboard/profile data ownership.

The website owns the user-facing presentation and interaction surface for those systems.

## Website Stages

### V0 static devlog-first site

V0 is the first public web presence.

It is settled as:

```text
static devlog-first website
index page as the current public/devlog landing page
devlog archive
individual static post pages
404 page
sitemap.xml
rss.xml and/or feed.xml
basic page-view analytics
static assets/media
```

Likely V0 stack:

```text
Plasmic for canonical visual layouts where practical
Next.js or another static-site generation path
static output
probably Netlify hosting
repo-owned post/content sources
repo-stored Plasmic exports/sources where possible
```

V0 should not include:

```text
accounts
comments
newsletter backend
API-backed publishing
public profiles
leaderboards
support portal
appeal portal
admin portal
commerce portal
runtime CMS backend
```

`robots.txt` is not required for V0. The site is intended to be publicly crawlable.

Sitemap/feed XML files are useful quality-of-life outputs and should be generated if the site tooling makes them cheap.

### Intermediate public/product site

An intermediate site may appear between V0 and launch.

It can expand public content without requiring the full launch account/commerce surface.

Possible intermediate additions:

```text
product-first homepage
roadmap/status page or section
richer game overview
screenshots/media gallery
community page
support/contact page
early lore/content pages
download/wishlist placeholders
```

The intermediate site should not imply unsupported account, purchase, leaderboard, profile, or platform features already exist.

### Launch website

The launch site is the full target for this domain.

The launch homepage should not be devlog-first. It should be a product landing page that quickly explains the game and routes users to play, buy, download, follow, read, or join.

Launch website areas:

```text
public content and marketing site
account, ownership, and commerce portal
community, engagement, and operational support
```

The launch website should be full-featured and API-backed where useful. It is not limited to static content.

## Public Content And Marketing Site

The public content and marketing site owns the public product-facing experience.

Launch content should include:

```text
game title and short pitch
product landing page
game overview
trailer/gameplay media/screenshots
play/download/buy entry points
release/current status
roadmap/status surface
devlog/update archive
lore or other deeper content
community/Discord link
support/help link
SEO/share metadata
```

The launch website should give interested users a reason to spend more than a few minutes browsing. Exact content areas are TBD, but valid content categories include:

```text
lore / setting
ships
weapons
factions or groups
enemies and bosses
modes
campaigns / seasons / events
roadmap / development status
developer posts
media gallery
FAQ
```

Launch content requires the site to be operational. Exact launch content scope is a gametime product decision.

A press/media kit is not required by default. It should be added if attention, press interest, or release strategy makes it useful.

## Devlog, Roadmap, And Developer Posts

Devlog/update content remains part of the site, but it is not the launch homepage’s primary identity.

Devlog expectations:

```text
stable post URLs
archive page
newest-first listing
individual post pages
date/title/summary metadata
RSS/feed output if practical
```

Roadmap/status is likely useful before launch and may appear as early as V0 or an intermediate site.

Selected developer/team posts may eventually support comments, but comments are not launch-required.

## Lore And Deep Content

Lore and deeper public content are valid launch website surfaces.

Their purpose is to give the site durable browseable material beyond purchase/download flows and update posts.

Possible lore/content surfaces include:

```text
setting overview
ship pages
weapon pages
faction/group pages
enemy/boss previews
campaign/event pages
mode explanations
```

Exact content categories should be decided when launch content strategy is clearer.

## Content Pipeline And CMS

V0 can remain file/static-content driven.

For launch, a CMS is likely useful enough to scaffold and may be implemented if practical.

CMS-backed content may include:

```text
devlog posts
roadmap/status entries
lore pages
FAQ/support copy
marketing/product copy
media/gallery entries
selected developer/team posts
```

CMS must not own:

```text
accounts
entitlements
payments
Steam verification
download authorization
leaderboards
profiles
moderation decisions
analytics source of truth
API contracts
```

Markdown is the default content-source format.

MDX is optional later if interactive React-style content inside posts/pages becomes useful.

Content governance should exist once CMS or CMS-like publishing exists:

```text
draft vs published state
author/team attribution
published dates
preview flow
edit/correction behavior
rollback path
media ownership
content review before publish where needed
```

## Plasmic, Next.js, And Site Workflow

Plasmic may own canonical visual layouts where practical.

Plasmic sources or exports should be stored in the repository if possible.

Next.js owns routing, build behavior, static generation, interactive pages, API-backed page integration, and deployable site output.

Repo-owned files remain the operational/deployable source of truth.

Recommended ownership split:

```text
Plasmic
-> visual layout and page composition where practical

Next.js / repo code
-> routing, content loading, static generation, account/commerce integration, build behavior, deployment output

API/platform services
-> authoritative account, purchase, entitlement, Steam verification, download, profile, leaderboard, moderation, and support state
```

## Account, Ownership, And Commerce Portal

A full account portal is a launch requirement.

Launch account portal should support:

```text
sign in / sign out
account identity/status presentation
linked provider status
Steam account linking
Steam ownership verification flow
direct purchase status
license/ownership status
download access
purchase receipt/status presentation
license claim status
basic account settings
support/help routes
```

Build as much account portal as possible without severely delaying release, but ownership, purchase, Steam-link, entitlement, and download access are core launch requirements if direct purchase and Steam ownership verification ship.

The website owns portal presentation and interaction.

Account/API/commerce systems own the authoritative state.

## Direct Purchase

Direct purchase is launch-planned.

Payment provider is a gametime decision:

```text
likely Square
Stripe remains possible
```

Website-owned purchase surfaces include:

```text
buy page
checkout entry point
purchase result page
purchase failure/retry presentation
receipt/account-facing purchase status
download/play route after successful ownership
support path for purchase problems
```

Commerce-owned behavior includes:

```text
payment provider integration
purchase validation
receipt authority
refund/reversal authority
entitlement creation
fraud/abuse handling
```

The website must never be authoritative for payment success, receipt state, entitlement state, or ownership.

## Steam Linking And Perpetual Direct-Download License

Steam purchasers should be eligible to claim a perpetual direct-download license on the website.

This is not a second Steam key claim.

The intended model is:

```text
user creates/signs into Space Rocks account
user links Steam
Steam ownership is verified
verified Steam ownership grants a perpetual direct-download entitlement/license to the Space Rocks account
account can access direct downloads through the website
```

An account is required.

Core policy:

```text
Buying is owning.
Steam ownership should unlock permanent direct-download access where technically and legally supported.
```

Website-owned Steam claim surfaces include:

```text
Steam link entry point
verification status
already-claimed state
wrong-account/error state
successful license grant presentation
download access after entitlement exists
support path for claim problems
```

Account/API/commerce/platform systems own actual Steam verification, linked-provider records, entitlement grants, and abuse/fraud handling.

## Download Access And Build Availability

A direct-download license requires a real download access surface.

Website-owned download surfaces should include:

```text
owned-download page
current build access
platform/OS labels
download unavailable state
license/ownership missing state
support route for download problems
```

Build/release owns artifacts, build packaging, versioning mechanics, checksums if used, and deployment/storage mechanics.

The website owns access presentation and account-gated routing.

Open launch decisions include:

```text
exact artifact hosting path
version history policy
checksum/file verification policy
download access after refund/reversal
download access when Steam verification later changes or fails
```

## Support, Recovery, And Account Help

Purchases, Steam claims, accounts, and downloads create support obligations.

The launch website should provide help routes for:

```text
lost account
login recovery
Steam linked to wrong account
Steam ownership verification failure
purchase email/account mismatch
missing direct-download license
failed download
receipt problem
refund/reversal question
license dispute
```

A full support-ticket system is not necessarily required for launch, but the user-facing route must exist.

Email flows may be needed for:

```text
account verification
password reset / login recovery
purchase receipt
Steam-link confirmation
license-claim confirmation
download/license support
enforcement/support notices
```

Email infrastructure is outside this document, but the website should plan the linked pages and user-facing states.

## Support And Admin Requirements

Launch account/ownership flows need support visibility.

Minimum support/admin needs:

```text
look up account
view ownership/entitlement status
view purchase/receipt status
view Steam-link/claim status
resend receipt or claim confirmation
manual correction/escalation path
fraud/abuse flag visibility
```

Internal admin/support implementation belongs outside this document, but the website plan should acknowledge that launch flows are not self-supporting without visibility and correction paths.

Developer-tool-like internal tools may come before polished website/admin workflows.

## Community And Discord Relationship

Discord remains the primary communication and community surface.

The website supports community presence but does not replace Discord.

Website community surfaces may include:

```text
official Discord link/hub
community rules
code of conduct
event/update posts
support/report links
selected developer/team post comments if added
```

The website should not add:

```text
custom forums
custom DMs
custom social feed
Discord profile mirrors
replacement Discord community
```

Website public/community surfaces must respect Space Rocks moderation rules.

## Comments

Comments are optional and not launch-required.

If added, comments should be limited to selected developer/team posts.

Comments require:

```text
authenticated accounts
per-post enable/disable
moderation
reporting
admin/developer removal tools
abuse/enforcement handoff
lock/disable behavior
```

General site-wide comments, forums, DMs, and social feeds are out of scope unless a later plan explicitly changes direction.

## Analytics And Engagement Measurement

V0 analytics can be limited to page views and broad engagement.

Launch analytics should measure website/product engagement and conversion.

Useful launch metrics include:

```text
page views
post/content views
roadmap views
media/trailer clicks
follows
sign-ups
sign-ins
direct purchase starts
direct purchase completions
purchase failures
Steam link starts
Steam ownership verification completions
direct-download license grants
download/play clicks
Discord/community clicks
comment engagement if comments exist
support/contact starts
```

Analytics provider is gametime.

Selection criteria:

```text
easy
cheap
effective enough
```

Analytics must not become:

```text
authoritative purchase state
authoritative ownership state
gameplay telemetry
account tracking beyond what is required for product flow measurement
behavioral profiling
required dependency for rendering the site
```

Avoid sending sensitive account, payment, or ownership data into analytics events.

## SEO And Shareability

SEO should be treated as public-site hygiene.

For Space Rocks, SEO means:

```text
search engines can understand the site
shared links render correctly in Discord and other apps
devlog posts have stable titles/descriptions/images
public pages are discoverable
```

Launch requirements:

```text
page titles
meta descriptions
Open Graph/social preview metadata
stable URLs
sitemap.xml
rss/feed XML
reasonable page headings
important content not hidden behind client-only rendering
favicon/app icons
```

This is not about gaming search results. It is about making public pages machine-readable, searchable, and shareable.

## Legal, Policy, And Public Disclosures

Launch account, analytics, purchase, Steam-linking, download, and comment surfaces require public policy pages.

Website presentation should include:

```text
privacy policy
terms of service / terms of use
refund policy
purchase/license terms
community guidelines / code of conduct
analytics/cookie disclosure if relevant
```

The website owns presenting and routing to these policies.

Legal/commercial policy authorship belongs outside this document.

## Public Identity And Profile Rules

Public Space Rocks identity must remain Space Rocks-owned.

Stable rules:

```text
website public profiles use public_profile_id
website must not expose internal account_id
Discord profile identity does not replace Space Rocks profile identity
display names are moderated Space Rocks presentation identity
public website-visible text uses Space Rocks moderation rules
```

Discord identity may appear only as optional linked-provider or community context where allowed by product and privacy rules.

## Leaderboard, Season, And Platform Surfaces

The website eventually needs leaderboard and public profile surfaces because platform planning requires them.

Planned website platform surfaces include:

```text
public leaderboard browser
public player profiles
season/campaign/event pages
archived boards
shareable board/profile URLs
invite landing pages
shareable match-result pages
report/appeal/account-status surfaces where relevant
```

The website and in-game client must not maintain separate hardcoded leaderboard catalogs.

Leaderboards, ranking formulas, board lifecycle, privacy behavior, and authoritative ranking data belong to Leaderboards And Rankings and API/platform systems.

The website owns presentation and navigation.

## Security And Abuse-Safe Website Flows

The launch site must treat account, purchase, Steam claim, and download paths as sensitive.

Security expectations to route into implementation planning:

```text
session and CSRF protection
rate limiting on login/claim/purchase paths
safe redirect handling
non-guessable download authorization
no internal account_id exposure
audit logs for ownership/claim/purchase changes
safe failure messages that do not expose private account or social facts
```

Abuse/enforcement owns moderation and enforcement decisions.

The website owns safe presentation and user-facing states.

## Failure And Degraded States

The launch website should have planned error states for major flows.

Required failure states include:

```text
payment provider unavailable
purchase failed or pending
Steam linking unavailable
Steam ownership cannot be verified
already-claimed license
wrong Steam account linked
download service unavailable
account service unavailable
expired/invalid session
CMS/content unavailable
analytics unavailable
support/contact unavailable
```

Analytics failure must not break the site.

API-backed feature failures should fail safely, especially around ownership, payment, download, privacy, and account state.

## Status And Outage Communication

The website should support basic public outage communication.

Early options:

```text
website notice/banner
support page notice
Discord announcement fallback
```

A dedicated public status page can come later if production operations justify it.

## Accessibility, Responsive Layout, And Quality Gates

The public website should work across common screen sizes and input methods.

Launch quality expectations:

```text
mobile/tablet/desktop layouts
keyboard navigation
reasonable contrast
alt text for important images
readable content pages
no critical actions available only by hover/animation
clear error messages
working 404 page
working sitemap/feed output
all internal links valid
Open Graph previews work
account/purchase/download flows tested in success and failure states
```

## Static Versus API-Backed Surfaces

Static or CMS-backed content surfaces:

```text
home/product pages
devlog/update pages
roadmap/status content
lore/deep content
media/gallery
FAQ/support copy
policy pages
community pages
press/media pages if added
```

API-backed surfaces:

```text
account portal
direct purchase
Steam linking
ownership verification
license/entitlement status
download access
purchase/receipt status
support/account help
comments if added
public profiles
leaderboards
invite landing
match result sharing
report/appeal/account-status flows
```

V0 should avoid API-backed surfaces.

Launch may include API-backed surfaces where required by account, commerce, ownership, download, or engagement needs.

## Implementation Sequence

1. Rewrite this document as the canonical website-and-web-presence domain plan.
2. Preserve V0 as a settled static devlog-first stage.
3. Establish the V0 Plasmic/Next.js/static-site path.
4. Add V0 index, devlog archive, post pages, 404, sitemap/feed XML, static assets, and basic analytics.
5. Add intermediate product-site pages or sections where useful.
6. Define launch content model and CMS scaffold.
7. Add launch homepage/product site shape.
8. Add roadmap/status and deeper public content surfaces.
9. Add full account portal surface.
10. Add direct purchase surface and payment-provider handoff.
11. Add Steam linking and ownership verification presentation.
12. Add perpetual direct-download license grant presentation.
13. Add account-gated download access surface.
14. Add purchase, claim, download, recovery, and support states.
15. Add legal/policy/disclosure pages.
16. Add launch analytics and conversion measurement.
17. Add SEO/shareability metadata and verification.
18. Add comments only if selected-post comments are worth the account/moderation/support cost.
19. Add leaderboard, profile, season, invite, match-result, report, and appeal surfaces when their platform systems are ready.
20. Keep backend authority for account, commerce, entitlements, verification, moderation, and rankings outside the website domain.

## Open Decisions

Gametime decisions:

```text
Square vs Stripe
exact CMS provider or implementation path
whether CMS is fully implemented at launch or only scaffolded
analytics provider
exact launch content areas
whether selected-post comments make launch
how complete the account portal gets beyond ownership/download/purchase needs
exact Steam verification implementation
exact refund/support workflow
whether press kit is worth adding
exact Plasmic export/source path under services/web
exact route names
exact SEO tuning
exact download artifact/version-history policy
```

These are not open direction questions:

```text
V0 is static and devlog-first.
Launch homepage is not devlog-first.
Launch website is a full product/platform surface.
Full account portal is launch-required.
Direct purchase is launch-planned.
Steam purchasers can link Steam to verify ownership and receive perpetual direct-download access.
An account is required for Steam ownership claim.
Buying is owning.
CMS scaffolding is justified for launch.
Discord remains the main community communication surface.
General forums, DMs, and social feeds are out of scope.
Comments are optional and limited to selected developer/team posts if added.
Website public identity uses Space Rocks public identity, not Discord profile identity.
```

## Core Invariants

```text
The website owns public/user-facing web presentation and interaction surfaces.

The website is not authoritative for account, payment, entitlement, Steam ownership, download authorization, moderation, leaderboard, or platform data.

V0 is a static devlog-first site and not the launch target.

The launch website is a full product/platform surface.

Launch content should provide enough depth for meaningful browsing beyond the first product impression.

Full account portal is launch-required.

Direct purchase is launch-planned.

Steam purchasers should be able to link Steam, verify ownership, and receive a perpetual direct-download entitlement attached to their Space Rocks account.

Buying is owning.

An account is required for ownership claims and account-gated download access.

CMS may own public content, but not platform state.

Plasmic may own canonical layouts where practical; repo-owned site code remains the deployable source.

Next.js owns routing, static generation, build behavior, and API-backed page integration.

Discord remains the primary community communication surface.

The website should not become a custom forum, DM system, social feed, or Discord profile mirror.

Analytics measure public-site engagement and conversion, not authoritative ownership or gameplay state.

SEO/shareability is required as public-site hygiene.

Website public profiles use public_profile_id and never expose internal account_id.

Public website-visible text must follow Space Rocks moderation rules.

Sensitive account, purchase, claim, and download flows must fail safely.
```

## Related Docs

* [Web](./!INDEX.md)
* [Planning](../../!INDEX.md)
* [Planned API Product Surface](../../protocol/api-product-surface.md)
* [Current API Product Surface](../../../protocol/api-product-surface.md)
* [Build Release And Environment Matrix](../technical/build-release-and-environment-matrix.md)
* [Account And Identity Systems](../platform/account-and-identity-systems.md)
* [Social And Community Systems](../platform/social-and-community-systems.md)
* [Leaderboards And Rankings](../platform/leaderboards-and-rankings.md)
* [Abuse And Enforcement Admin](../platform/security-and-admin/abuse-and-enforcement-admin.md)
* [Shop Commerce And Economy](../gameplay/shop-commerce-and-economy.md)
* [Season Format And Progression](../platform/season-format-and-progression.md)

## Notes

The V0 site and launch site should not be described as the same target.

V0 is a small static/devlog-first presence. Launch is a full product and platform website with account, ownership, commerce, content, analytics, and support expectations.

The phrase “Steam key claim” should be avoided for this plan. The intended launch behavior is Steam ownership verification followed by a Space Rocks account-owned perpetual direct-download license, not issuing a second Steam key.
