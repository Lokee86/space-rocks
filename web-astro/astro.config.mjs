// @ts-check
import { defineConfig } from 'astro/config';
import { fileURLToPath } from 'node:url';

import react from '@astrojs/react';

import mdx from '@astrojs/mdx';

const projectRoot = fileURLToPath(new URL('./', import.meta.url));

// https://astro.build/config
export default defineConfig({
  site: 'https://space-rocks.pages.dev',
  integrations: [react(), mdx()],
  vite: {
    resolve: {
      alias: [
        {
          find: '@plasmicapp/react-web/lib/host',
          replacement: `${projectRoot}node_modules/@plasmicapp/react-web/lib/host/index.js`
        },
        {
          find: '@plasmicapp/react-web/lib/query',
          replacement: `${projectRoot}node_modules/@plasmicapp/react-web/lib/query/index.js`
        },
        {
          find: '@plasmicapp/react-web/lib/plasmic.css',
          replacement: `${projectRoot}node_modules/@plasmicapp/react-web/lib/plasmic.css`
        },
        {
          find: 'classnames',
          replacement: `${projectRoot}src/compat/classnames.ts`
        },
        {
          find: 'classnames/index.js',
          replacement: `${projectRoot}src/compat/classnames.ts`
        },
        {
          find: 'clone',
          replacement: `${projectRoot}src/compat/clone.ts`
        },
        {
          find: 'clone/clone.js',
          replacement: `${projectRoot}src/compat/clone.ts`
        },
        {
          find: 'dlv',
          replacement: `${projectRoot}node_modules/dlv/dist/dlv.es.js`
        },
        {
          find: 'dlv/dist/dlv.umd.js',
          replacement: `${projectRoot}node_modules/dlv/dist/dlv.es.js`
        },
        {
          find: 'use-sync-external-store/shim/index.js',
          replacement: `${projectRoot}src/compat/use-sync-external-store-shim.ts`
        },
        {
          find: '@plasmicapp/react-web',
          replacement: `${projectRoot}src/compat/plasmic-react-web.ts`
        },
        {
          find: '@',
          replacement: projectRoot
        }
      ]
    }
  }
});
