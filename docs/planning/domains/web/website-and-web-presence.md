# Website And Web Presence

Parent index: [Web](./!INDEX.md)

## Purpose

Future-facing planning for the Space Rocks website and public web presence.

See the web service docs for implementation details:

- `docs/services/web/devlog-static-site.md`
- `docs/services/web/plasmic-astro-workflow.md`
- `docs/services/web/crt-media-frame.md`

## Overview

This document plans the product and platform web surface, including the intermediate public/product site and the launch website.

The `tools/parked-plasmic-next-host/` app may be recycled or replaced when the future interactive website is established.

## Planning Scope

This document keeps future and unimplemented website planning only.

It covers:

- intermediate public/product site shape
- launch website shape
- account portal
- direct purchase
- Steam ownership verification presentation
- account-gated download access
- CMS-backed public content
- support and recovery surfaces
- analytics and conversion measurement
- public profiles and leaderboards
- SEO/shareability
- legal/policy pages
- community handoff
- security and abuse-safe flows

## Ownership Boundary

This document owns planning for public website product surfaces, website stages, roadmap/status presentation, lore and deeper public content surfaces, media/gallery/press-facing surfaces, portal presentation, support/help entry points, analytics requirements, SEO/shareability, and website quality gates.

It does not own API endpoint design, account authentication mechanics, payment-provider integration internals, purchase validation, entitlement authority, Steam verification internals, download artifact building, hosting/deployment mechanics, ranking formulas, moderation decisions, legal policy authorship, analytics provider implementation, or CMS implementation details.

The website owns user-facing presentation and interaction; backend systems own authority.

## Website Stages

### Intermediate Public/Product Site

The intermediate site may appear between the current public presence and launch.

Possible intermediate additions:

- product-first homepage
- roadmap/status page or section
- richer game overview
- screenshots/media gallery
- community page
- support/contact page
- lore/content pages
- download or wishlist placeholders

### Launch Website

The launch site is the full target for this domain.

Launch website areas:

- public content and marketing site
- account, ownership, and commerce portal
- community, engagement, and operational support

## Public Content And Marketing Site

Launch content should include:

- game title and short pitch
- product landing page
- game overview
- trailer/gameplay media/screenshots
- play/download/buy entry points
- release/current status
- roadmap/status surface
- devlog/update archive
- lore or deeper content
- community link
- support/help link
- SEO/share metadata

Valid content categories may include lore, ships, weapons, factions, enemies, modes, campaigns, events, developer posts, media gallery, and FAQ.

## Content Pipeline And CMS

For launch, a CMS is likely useful enough to scaffold and may be implemented if practical.

CMS-backed content may include devlog posts, roadmap/status entries, lore pages, FAQ/support copy, marketing/product copy, media/gallery entries, and selected developer/team posts.

CMS must not own accounts, entitlements, payments, Steam verification, download authorization, leaderboards, profiles, moderation decisions, analytics source of truth, or API contracts.

## Plasmic And Site Workflow

Plasmic may own canonical visual layouts where practical.

Planned ownership split:

- Plasmic
  - visual layout and page composition where practical
- Repo-owned wrapper/source code
  - routing, content loading, static generation, account/commerce integration, build behavior, deployment output
- API/platform services
  - authoritative account, purchase, entitlement, Steam verification, download, profile, leaderboard, moderation, and support state

## Account, Ownership, And Commerce

The launch website should plan for:

- sign in / sign out
- account identity/status presentation
- Steam account linking
- Steam ownership verification flow
- direct purchase status
- license/ownership status
- download access
- receipt/status presentation
- support/help routes

Direct purchase, Steam-linking, ownership verification, entitlement, and download access are launch-planned and should fail safely.

## Support, Recovery, And Help

The launch website should provide help routes for:

- lost account
- login recovery
- Steam linked to wrong account
- Steam ownership verification failure
- purchase email/account mismatch
- missing direct-download license
- failed download
- receipt problem
- refund/reversal question
- license dispute

## Analytics And SEO

The current static devlog can use Cloudflare Web Analytics as its baseline visit/page-view analytics path.

Launch analytics should measure website/product engagement and conversion.

Useful metrics include page views, content views, media clicks, sign-ups, sign-ins, purchase starts/completions, Steam link starts/completions, direct-download license grants, download clicks, community clicks, and support/contact starts.

SEO/shareability should include page titles, meta descriptions, Open Graph/social preview metadata, stable URLs, sitemap/feed XML, reasonable headings, and favicon/app icons.

## Public Identity, Profiles, And Leaderboards

The website should preserve Space Rocks-owned public identity.

Planned surfaces include public profiles, public leaderboards, season/campaign/event pages, archived boards, shareable board/profile URLs, invite landing pages, shareable match-result pages, and report/appeal/account-status surfaces where relevant.

## Security, Policy, And Failure States

Launch web flows must treat account, purchase, Steam claim, and download paths as sensitive.

The website should plan for safe redirects, rate limiting, non-guessable download authorization, no internal account_id exposure, and safe failure messages.

Public policy pages should include privacy policy, terms of use, refund policy, purchase/license terms, community guidelines, and analytics/cookie disclosure if relevant.

Failure states should include payment provider unavailable, purchase failed or pending, Steam linking unavailable, Steam ownership cannot be verified, already-claimed license, wrong Steam account linked, download unavailable, account unavailable, expired/invalid session, CMS/content unavailable, analytics unavailable, and support/contact unavailable.

## Accessibility And Quality Gates

The public website should support mobile, tablet, and desktop layouts, keyboard navigation, reasonable contrast, alt text for important images, readable content pages, clear error messages, working 404 behavior, working sitemap/feed output, valid internal links, and Open Graph previews.

## Static Versus API-Backed Surfaces

Static or CMS-backed surfaces:

- home/product pages
- devlog/update pages
- roadmap/status content
- lore/deep content
- media/gallery
- FAQ/support copy
- policy pages
- community pages
- press/media pages if added

API-backed surfaces:

- account portal
- direct purchase
- Steam linking
- ownership verification
- license/entitlement status
- download access
- purchase/receipt status
- support/account help
- comments if added
- public profiles
- leaderboards
- invite landing
- match result sharing
- report/appeal/account-status flows

## Implementation Sequence

1. Keep current implementation details in the web service docs.
2. Add intermediate public/product site pages or sections where useful.
3. Define launch content model and CMS scaffold.
4. Add launch homepage/product site shape.
5. Add roadmap/status and deeper public content surfaces.
6. Add full account portal surface.
7. Add direct purchase surface and payment-provider handoff.
8. Add Steam linking and ownership verification presentation.
9. Add perpetual direct-download license grant presentation.
10. Add account-gated download access surface.
11. Add purchase, claim, download, recovery, and support states.
12. Add legal/policy/disclosure pages.
13. Add launch analytics and conversion measurement.
14. Add SEO/shareability metadata and verification.
15. Add comments only if selected-post comments are worth the account/moderation/support cost.
16. Add leaderboard, profile, season, invite, match-result, report, and appeal surfaces when their platform systems are ready.
17. Keep backend authority for account, commerce, entitlements, verification, moderation, and rankings outside the website domain.

## Open Decisions

Gametime decisions:

- exact CMS provider or implementation path
- whether CMS is fully implemented at launch or only scaffolded
- analytics provider
- exact launch content areas
- whether selected-post comments make launch
- how complete the account portal gets beyond ownership/download/purchase needs
- exact Steam verification implementation
- exact refund/support workflow
- whether press kit is worth adding
- exact SEO tuning
- exact download artifact/version-history policy

## Related Docs

- [Web](./!INDEX.md)
- [Planning](../../!INDEX.md)
- [Planned API Product Surface](../../protocol/api-product-surface.md)
- [Current API Product Surface](../../../protocol/api-product-surface.md)
- [Build Release And Environment Matrix](../technical/build-release-and-environment-matrix.md)
- [Devlog Static Site](../../../services/web/devlog-static-site.md)
- [Plasmic / Astro Workflow](../../../services/web/plasmic-astro-workflow.md)
- [CRT Media Frame](../../../services/web/crt-media-frame.md)
- [Account And Identity Systems](../platform/account-and-identity-systems.md)
- [Social And Community Systems](../platform/social-and-community-systems.md)
- [Leaderboards And Rankings](../platform/leaderboards-and-rankings.md)
- [Abuse And Enforcement Admin](../platform/security-and-admin/abuse-and-enforcement-admin.md)
- [Shop Commerce And Economy](../gameplay/shop-commerce-and-economy.md)
- [Season Format And Progression](../platform/season-format-and-progression.md)

## Notes

The launch website is a full product and platform surface.

The phrase “Steam key claim” should be avoided for this plan. The intended launch behavior is Steam ownership verification followed by a Space Rocks account-owned perpetual direct-download license, not issuing a second Steam key.
