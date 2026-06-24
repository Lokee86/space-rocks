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
      defaultValue: 5.0,
      min: 0,
      max: 20,
      step: 0.1,
    },
    screenInsetRight: {
      type: "number",
      defaultValue: 5.0,
      min: 0,
      max: 20,
      step: 0.1,
    },
    screenInsetTop: {
      type: "number",
      defaultValue: 11.8,
      min: 0,
      max: 25,
      step: 0.1,
    },
    screenInsetBottom: {
      type: "number",
      defaultValue: 12.4,
      min: 0,
      max: 25,
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
    showControls: {
      type: "boolean",
      defaultValue: false,
    },
    disabledControls: {
      type: "choice",
      options: ["previous", "rewind", "play", "pause", "fastForward", "next"],
      multiSelect: true,
      defaultValue: [],
    },
    scanlineCount: {
      type: "number",
      defaultValue: 480,
      min: 100,
      max: 1200,
      step: 10,
    },
    scanlineStrength: {
      type: "number",
      defaultValue: 0.24,
      min: 0,
      max: 1,
      step: 0.01,
    },
    scanlineHardness: {
      type: "number",
      defaultValue: 1.65,
      min: 0.1,
      max: 5,
      step: 0.05,
    },
    scanlineBreakupStrength: {
      type: "number",
      defaultValue: 0.16,
      min: 0,
      max: 1,
      step: 0.01,
    },
    scanlineBreakupSegments: {
      type: "number",
      defaultValue: 36,
      min: 1,
      max: 120,
      step: 1,
    },
    scanlineBreakupCutoff: {
      type: "number",
      defaultValue: 0.46,
      min: 0,
      max: 1,
      step: 0.01,
    },
    scanlineBreakupSoftness: {
      type: "number",
      defaultValue: 0.18,
      min: 0,
      max: 1,
      step: 0.01,
    },
    scanlineLineVarianceStrength: {
      type: "number",
      defaultValue: 0.08,
      min: 0,
      max: 1,
      step: 0.01,
    },
    waveStrength: {
      type: "number",
      defaultValue: 1,
      min: 0,
      max: 5,
      step: 0.05,
    },
    waveSpeed: {
      type: "number",
      defaultValue: 1,
      min: 0,
      max: 5,
      step: 0.05,
    },
    waveSlowScale: {
      type: "number",
      defaultValue: 18,
      min: 0,
      max: 100,
      step: 1,
    },
    waveMediumScale: {
      type: "number",
      defaultValue: 63,
      min: 0,
      max: 200,
      step: 1,
    },
    waveFineScale: {
      type: "number",
      defaultValue: 210,
      min: 0,
      max: 500,
      step: 5,
    },
    lineJitterStrength: {
      type: "number",
      defaultValue: 0.0012,
      min: 0,
      max: 0.01,
      step: 0.0001,
    },
    flickerStrength: {
      type: "number",
      defaultValue: 0.025,
      min: 0,
      max: 0.25,
      step: 0.005,
    },
    flickerSpeed: {
      type: "number",
      defaultValue: 18,
      min: 0,
      max: 60,
      step: 0.5,
    },
    flickerSpeedVariance: {
      type: "number",
      defaultValue: 0.55,
      min: 0,
      max: 3,
      step: 0.05,
    },
    flickerVarianceSpeed: {
      type: "number",
      defaultValue: 1.35,
      min: 0,
      max: 10,
      step: 0.05,
    },
    flickerSecondaryStrength: {
      type: "number",
      defaultValue: 0.35,
      min: 0,
      max: 2,
      step: 0.05,
    },
    rollStrength: {
      type: "number",
      defaultValue: 0.08,
      min: 0,
      max: 1,
      step: 0.01,
    },
    rollInterval: {
      type: "number",
      defaultValue: 5,
      min: 0.5,
      max: 20,
      step: 0.1,
    },
    rollDuration: {
      type: "number",
      defaultValue: 1.2,
      min: 0.1,
      max: 5,
      step: 0.1,
    },
    rollWidth: {
      type: "number",
      defaultValue: 0.1,
      min: 0.01,
      max: 0.5,
      step: 0.01,
    },
    rollSpeed: {
      type: "number",
      defaultValue: 1,
      min: 0,
      max: 5,
      step: 0.05,
    },
    rollHorizontalVariation: {
      type: "number",
      defaultValue: 0.15,
      min: 0,
      max: 1,
      step: 0.01,
    },
    horizontalShimmerStrength: {
      type: "number",
      defaultValue: 0.055,
      min: 0,
      max: 1,
      step: 0.01,
    },
    horizontalShimmerSpeed: {
      type: "number",
      defaultValue: 1.8,
      min: 0,
      max: 10,
      step: 0.05,
    },
    horizontalShimmerCount: {
      type: "number",
      defaultValue: 42,
      min: 0,
      max: 200,
      step: 1,
    },
    edgeGlowStrength: {
      type: "number",
      defaultValue: 0.08,
      min: 0,
      max: 1,
      step: 0.01,
    },
    edgeGlowWidth: {
      type: "number",
      defaultValue: 0.01,
      min: 0,
      max: 0.1,
      step: 0.001,
    },
    edgeCornerGlowWidth: {
      type: "number",
      defaultValue: 0.055,
      min: 0,
      max: 0.3,
      step: 0.005,
    },
    edgeCornerGlowPower: {
      type: "number",
      defaultValue: 2.2,
      min: 0.1,
      max: 8,
      step: 0.05,
    },
    edgeGlowSoftness: {
      type: "number",
      defaultValue: 0.018,
      min: 0,
      max: 0.1,
      step: 0.001,
    },
    vignetteStrength: {
      type: "number",
      defaultValue: 0.24,
      min: 0,
      max: 2,
      step: 0.01,
    },
    vignetteEdgeBypassStrength: {
      type: "number",
      defaultValue: 1,
      min: 0,
      max: 1,
      step: 0.01,
    },
    vignetteEdgeBypassWidth: {
      type: "number",
      defaultValue: 0.035,
      min: 0,
      max: 0.2,
      step: 0.005,
    },
    effectCutoff: {
      type: "number",
      defaultValue: 0.018,
      min: 0,
      max: 0.2,
      step: 0.001,
    },
    effectGain: {
      type: "number",
      defaultValue: 1.25,
      min: 0,
      max: 5,
      step: 0.05,
    },
    animationSpeed: {
      type: "number",
      defaultValue: 1,
      min: 0,
      max: 5,
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
