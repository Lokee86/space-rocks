import { initPlasmicLoader } from "@plasmicapp/loader-nextjs";

import { CardFrame } from "@/src/components/card/CardFrame";
import { CrtMediaFrame } from "@/src/components/media/CrtMediaFrame";

const projectId = process.env.NEXT_PUBLIC_PLASMIC_PROJECT_ID;
const projectToken = process.env.NEXT_PUBLIC_PLASMIC_PROJECT_TOKEN;

if (!projectId) {
  throw new Error("Missing required env var NEXT_PUBLIC_PLASMIC_PROJECT_ID");
}

if (!projectToken) {
  throw new Error("Missing required env var NEXT_PUBLIC_PLASMIC_PROJECT_TOKEN");
}

export const PLASMIC = initPlasmicLoader({
  projects: [
    {
      id: projectId,
      token: projectToken,
    },
  ],
  preview: process.env.NODE_ENV !== "production",
});

PLASMIC.registerComponent(CrtMediaFrame, {
  name: "CrtMediaFrame",
  displayName: "CRT Media Frame",
  importPath: "@/src/components/media/CrtMediaFrame",
  props: {
    shaderEnabled: {
      type: "boolean",
      defaultValue: true,
    },
    shaderDebug: {
      type: "boolean",
      defaultValue: false,
    },
    freezeShaderTime: {
      type: "boolean",
      defaultValue: false,
    },
    screenInsetLeft: {
      type: "number",
      defaultValue: 0,
      min: -9999,
      max: 9999,
      step: 0.1,
    },
    screenInsetRight: {
      type: "number",
      defaultValue: 0,
      min: -9999,
      max: 9999,
      step: 0.1,
    },
    screenInsetTop: {
      type: "number",
      defaultValue: 0,
      min: -9999,
      max: 9999,
      step: 0.1,
    },
    screenInsetBottom: {
      type: "number",
      defaultValue: 0,
      min: -9999,
      max: 9999,
      step: 0.1,
    },
    baseColor: {
      type: "string",
      defaultValue: "#020617",
    },
    glowColor: {
      type: "string",
      defaultValue: "#00e5ff",
    },
    scanlineColor: {
      type: "string",
      defaultValue: "#7dd3fc",
    },
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
    mediaMode: {
      type: "choice",
      options: ["imageList", "video"],
    },
    imageItems: {
      type: "string",
      defaultValue: "",
    },
    videoSrc: {
      type: "string",
      defaultValue: "",
    },
    autoAdvanceMs: {
      type: "number",
      defaultValue: 5000,
      min: -9999,
      max: 9999,
      step: 100,
    },
    seekSeconds: {
      type: "number",
      defaultValue: 10,
      min: -9999,
      max: 9999,
      step: 1,
    },
    initialIndex: {
      type: "number",
      defaultValue: 0,
      min: -9999,
      max: 9999,
      step: 1,
    },
    showControls: {
      type: "boolean",
      defaultValue: false,
    },
    disabledControls: {
      type: "string",
      defaultValue: "",
    },
    scanlineCount: {
      type: "number",
      defaultValue: 480,
      min: -9999,
      max: 9999,
      step: 10,
    },
    scanlineStrength: {
      type: "number",
      defaultValue: 0.24,
      min: -9999,
      max: 9999,
      step: 0.01,
    },
    scanlineHardness: {
      type: "number",
      defaultValue: 1.65,
      min: -9999,
      max: 9999,
      step: 0.05,
    },
    scanlineBreakupStrength: {
      type: "number",
      defaultValue: 0.16,
      min: -9999,
      max: 9999,
      step: 0.01,
    },
    scanlineBreakupSegments: {
      type: "number",
      defaultValue: 36,
      min: -9999,
      max: 9999,
      step: 1,
    },
    scanlineBreakupCutoff: {
      type: "number",
      defaultValue: 0.46,
      min: -9999,
      max: 9999,
      step: 0.01,
    },
    scanlineBreakupSoftness: {
      type: "number",
      defaultValue: 0.18,
      min: -9999,
      max: 9999,
      step: 0.01,
    },
    scanlineLineVarianceStrength: {
      type: "number",
      defaultValue: 0.08,
      min: -9999,
      max: 9999,
      step: 0.01,
    },
    waveStrength: {
      type: "number",
      defaultValue: 1,
      min: -9999,
      max: 9999,
      step: 0.05,
    },
    waveSpeed: {
      type: "number",
      defaultValue: 1,
      min: -9999,
      max: 9999,
      step: 0.05,
    },
    waveSlowScale: {
      type: "number",
      defaultValue: 18,
      min: -9999,
      max: 9999,
      step: 1,
    },
    waveMediumScale: {
      type: "number",
      defaultValue: 63,
      min: -9999,
      max: 9999,
      step: 1,
    },
    waveFineScale: {
      type: "number",
      defaultValue: 210,
      min: -9999,
      max: 9999,
      step: 5,
    },
    lineJitterStrength: {
      type: "number",
      defaultValue: 0.0012,
      min: -9999,
      max: 9999,
      step: 0.0001,
    },
    flickerStrength: {
      type: "number",
      defaultValue: 0.025,
      min: -9999,
      max: 9999,
      step: 0.005,
    },
    flickerSpeed: {
      type: "number",
      defaultValue: 18,
      min: -9999,
      max: 9999,
      step: 0.5,
    },
    flickerSpeedVariance: {
      type: "number",
      defaultValue: 0.55,
      min: -9999,
      max: 9999,
      step: 0.05,
    },
    flickerVarianceSpeed: {
      type: "number",
      defaultValue: 1.35,
      min: -9999,
      max: 9999,
      step: 0.05,
    },
    flickerSecondaryStrength: {
      type: "number",
      defaultValue: 0.35,
      min: -9999,
      max: 9999,
      step: 0.05,
    },
    rollStrength: {
      type: "number",
      defaultValue: 0.08,
      min: -9999,
      max: 9999,
      step: 0.01,
    },
    rollInterval: {
      type: "number",
      defaultValue: 5,
      min: -9999,
      max: 9999,
      step: 0.1,
    },
    rollDuration: {
      type: "number",
      defaultValue: 1.2,
      min: -9999,
      max: 9999,
      step: 0.1,
    },
    rollWidth: {
      type: "number",
      defaultValue: 0.1,
      min: -9999,
      max: 9999,
      step: 0.01,
    },
    rollSpeed: {
      type: "number",
      defaultValue: 1,
      min: -9999,
      max: 9999,
      step: 0.05,
    },
    rollHorizontalVariation: {
      type: "number",
      defaultValue: 0.15,
      min: -9999,
      max: 9999,
      step: 0.01,
    },
    horizontalShimmerStrength: {
      type: "number",
      defaultValue: 0.055,
      min: -9999,
      max: 9999,
      step: 0.01,
    },
    horizontalShimmerSpeed: {
      type: "number",
      defaultValue: 1.8,
      min: -9999,
      max: 9999,
      step: 0.05,
    },
    horizontalShimmerCount: {
      type: "number",
      defaultValue: 42,
      min: -9999,
      max: 9999,
      step: 1,
    },
    edgeGlowStrength: {
      type: "number",
      defaultValue: 0.08,
      min: -9999,
      max: 9999,
      step: 0.01,
    },
    edgeGlowWidth: {
      type: "number",
      defaultValue: 0.01,
      min: -9999,
      max: 9999,
      step: 0.001,
    },
    edgeCornerGlowWidth: {
      type: "number",
      defaultValue: 0.055,
      min: -9999,
      max: 9999,
      step: 0.005,
    },
    edgeCornerGlowPower: {
      type: "number",
      defaultValue: 2.2,
      min: -9999,
      max: 9999,
      step: 0.05,
    },
    edgeGlowSoftness: {
      type: "number",
      defaultValue: 0.018,
      min: -9999,
      max: 9999,
      step: 0.001,
    },
    vignetteStrength: {
      type: "number",
      defaultValue: 0.24,
      min: -9999,
      max: 9999,
      step: 0.01,
    },
    vignetteEdgeBypassStrength: {
      type: "number",
      defaultValue: 1,
      min: -9999,
      max: 9999,
      step: 0.01,
    },
    vignetteEdgeBypassWidth: {
      type: "number",
      defaultValue: 0.035,
      min: -9999,
      max: 9999,
      step: 0.005,
    },
    effectCutoff: {
      type: "number",
      defaultValue: 0.018,
      min: -9999,
      max: 9999,
      step: 0.001,
    },
    effectGain: {
      type: "number",
      defaultValue: 1.25,
      min: -9999,
      max: 9999,
      step: 0.05,
    },
    animationSpeed: {
      type: "number",
      defaultValue: 1,
      min: -9999,
      max: 9999,
      step: 0.05,
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
