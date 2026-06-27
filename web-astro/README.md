# Space Rocks Web

Current Astro site for the Space Rocks devlog.

## Quickstart

Run these from `web-astro/`.

- `npm install` if dependencies are missing
- `npm run dev` for local Astro development
- `npm run build` for build verification
- Plasmic codegen when Studio changes require it:

```text
npx @plasmicapp/cli@0.1.365 sync -p uNJepqX5kmDcUn9dDb3UVD --yes
```

## Content and assets

- Live devlog content: `src/content/devlog/`
- Devlog media assets: `public/media/devlog/<slug>/`
- Deployment root: `web-astro/`
- Deployment output: `dist`

## More detail

- [Devlog Static Site](../docs/services/web/devlog-static-site.md)
- [Cloudflare Pages Deployment](../docs/services/web/cloudflare-pages-deployment.md)
- [Plasmic / Astro Workflow](../docs/services/web/plasmic-astro-workflow.md)
