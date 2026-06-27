# Devlog Static Site

Parent index: [Web](./!INDEX.md)

## Purpose

Future-only planning for remaining devlog static-site improvements.

See the current web service docs for implementation details:

- `docs/services/web/devlog-static-site.md`
- `docs/services/web/plasmic-astro-workflow.md`
- `docs/services/web/crt-media-frame.md`
- `docs/services/web/cloudflare-pages-deployment.md`

## Current status

The public devlog V0 exists. Implementation facts now live in the service docs.

Cloudflare Web Analytics is the current baseline analytics path for the static devlog site.

## Remaining scope

- 404 behavior
- sitemap/feed generation
- SEO/shareability hygiene
- validation, link checking, and static build checks
- static-site quality gates

## Out of scope

- homepage route details
- archive route details
- devlog post route details
- `web-astro/` implementation details
- content schema details
- media asset conventions
- Plasmic/Astro workflow mechanics
- metadata/Open Graph implementation details
- analytics wiring beyond the current Cloudflare Web Analytics baseline

## Open decisions

- Which of the remaining static-site improvements still need implementation work
- Whether this note should remain once the remaining improvements are closed out

## Related docs

- [Devlog Static Site](../../services/web/devlog-static-site.md)
- [Plasmic / Astro Workflow](../../services/web/plasmic-astro-workflow.md)
- [CRT Media Frame](../../services/web/crt-media-frame.md)
- [Cloudflare Pages Deployment](../../services/web/cloudflare-pages-deployment.md)
- [Website and web presence](../domains/web/website-and-web-presence.md)

## Notes

Keep implementation detail in the service docs and use this note only for unresolved planning items.
