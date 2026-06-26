import { defineCollection, z } from "astro:content";
import { glob } from "astro/loaders";

const devlog = defineCollection({
  loader: glob({
    pattern: "**/*.md",
    base: "./src/content/devlog",
  }),
  schema: z.object({
    title: z.string(),
    date: z.date(),
    summary: z.string(),
    heroLine1: z.string(),
    heroLine2: z.string(),
    heroLine3: z.string(),
    articleLabel: z.string(),
    articleTitle: z.string(),
    intro: z.string(),
    whatChanged: z.string(),
    callout: z.string(),
    whatsNext: z.string(),
    finishedBody: z.string(),
    nowBody: z.string(),
    comingUpBody: z.string(),
    utilityText: z.string(),
    finePrint: z.string(),
  }),
});

export const collections = {
  devlog,
};
