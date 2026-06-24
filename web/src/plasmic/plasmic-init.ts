import { initPlasmicLoader } from "@plasmicapp/loader-nextjs";

import { CardFrame } from "@/src/components/card/CardFrame";
import { CrtMediaFrame } from "@/src/components/media/CrtMediaFrame";

const projectId =
  process.env.NEXT_PUBLIC_PLASMIC_PROJECT_ID ?? "plasmic-placeholder-project-id";
const projectToken =
  process.env.NEXT_PUBLIC_PLASMIC_PROJECT_TOKEN ?? "plasmic-placeholder-project-token";

export const PLASMIC = initPlasmicLoader({
  projects: [
    {
      id: projectId,
      token: projectToken,
    },
  ],
  preview: true,
});

PLASMIC.registerComponent(CrtMediaFrame, {
  name: "CrtMediaFrame",
  displayName: "CRT Media Frame",
  importPath: "@/src/components/media/CrtMediaFrame",
  props: {
    src: "string",
    alt: {
      type: "string",
      defaultValue: "",
    },
    caption: "string",
    aspectRatio: {
      type: "choice",
      options: ["16 / 9", "4 / 3", "1 / 1"],
      defaultValue: "16 / 9",
    },
    fit: {
      type: "choice",
      options: ["cover", "contain"],
      defaultValue: "cover",
    },
    tint: {
      type: "choice",
      options: ["cyan", "yellow", "red"],
      defaultValue: "cyan",
    },
  },
  defaultStyles: {
    width: "100%",
  },
});

PLASMIC.registerComponent(CardFrame, {
  name: "CardFrame",
  displayName: "Card Frame",
  importPath: "@/src/components/card/CardFrame",
  props: {
    title: "string",
    eyebrow: "string",
    tone: {
      type: "choice",
      options: ["cyan", "yellow", "red"],
    },
    padding: {
      type: "choice",
      options: ["sm", "md", "lg"],
    },
    children: "slot",
  },
  defaultStyles: {
    width: "100%",
  },
});
