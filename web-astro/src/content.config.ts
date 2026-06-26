import { defineCollection, z } from "astro:content";
import { glob } from "astro/loaders";

const mediaKind = z.enum(["", "images", "youtube"]).default("");

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
    heroMediaKind: mediaKind,
    heroImages: z.array(z.string()).default([]),
    heroYoutubeUrl: z.string().default(""),
    heroMediaAlt: z.string().default(""),
    articleLabel: z.string(),
    articleTitle: z.string(),
    intro: z.string(),
    articleMediaKind: mediaKind,
    articleImages: z.array(z.string()).default([]),
    articleYoutubeUrl: z.string().default(""),
    articleMediaAlt: z.string().default(""),
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
