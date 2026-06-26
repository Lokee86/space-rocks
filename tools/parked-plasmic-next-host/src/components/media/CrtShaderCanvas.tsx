import { useEffect, useRef } from "react";

export type CrtShaderCanvasProps = {
  className?: string;
  enabled?: boolean;
  tint?: "cyan" | "yellow" | "red";
  baseColor?: string;
  glowColor?: string;
  scanlineColor?: string;
  scanlineCount?: number;
  scanlineStrength?: number;
  scanlineHardness?: number;
  scanlineBreakupStrength?: number;
  scanlineBreakupSegments?: number;
  scanlineBreakupCutoff?: number;
  scanlineBreakupSoftness?: number;
  scanlineLineVarianceStrength?: number;
  waveStrength?: number;
  waveSpeed?: number;
  waveSlowScale?: number;
  waveMediumScale?: number;
  waveFineScale?: number;
  lineJitterStrength?: number;
  flickerStrength?: number;
  flickerSpeed?: number;
  flickerSpeedVariance?: number;
  flickerVarianceSpeed?: number;
  flickerSecondaryStrength?: number;
  rollStrength?: number;
  rollInterval?: number;
  rollDuration?: number;
  rollWidth?: number;
  rollSpeed?: number;
  rollHorizontalVariation?: number;
  shimmerStrength?: number;
  horizontalShimmerStrength?: number;
  horizontalShimmerSpeed?: number;
  horizontalShimmerCount?: number;
  edgeGlowStrength?: number;
  edgeGlowWidth?: number;
  edgeCornerGlowWidth?: number;
  edgeCornerGlowPower?: number;
  edgeGlowSoftness?: number;
  vignetteStrength?: number;
  vignetteEdgeBypassStrength?: number;
  vignetteEdgeBypassWidth?: number;
  effectCutoff?: number;
  effectGain?: number;
  animationSpeed?: number;
  shaderDebug?: boolean;
  freezeShaderTime?: boolean;
  lineWarpStrength?: number;
};

const VERTEX_SHADER = `
attribute vec2 a_position;
varying vec2 v_uv;
void main() {
  v_uv = a_position * 0.5 + 0.5;
  gl_Position = vec4(a_position, 0.0, 1.0);
}
`;

const FRAGMENT_SHADER = `
precision highp float;
varying vec2 v_uv;

uniform float time;
uniform float animationSpeed;
uniform float debugMode;
uniform float scanlineCount;
uniform float scanlineStrength;
uniform float scanlineHardness;
uniform float scanlineBreakupStrength;
uniform float scanlineBreakupSegments;
uniform float scanlineBreakupCutoff;
uniform float scanlineBreakupSoftness;
uniform float scanlineLineVarianceStrength;
uniform float waveStrength;
uniform float waveSpeed;
uniform float waveSlowScale;
uniform float waveMediumScale;
uniform float waveFineScale;
uniform float lineJitterStrength;
uniform float flickerStrength;
uniform float flickerSpeed;
uniform float flickerSpeedVariance;
uniform float flickerVarianceSpeed;
uniform float flickerSecondaryStrength;
uniform float rollStrength;
uniform float rollInterval;
uniform float rollDuration;
uniform float rollWidth;
uniform float rollSpeed;
uniform float rollHorizontalVariation;
uniform float horizontalShimmerStrength;
uniform float horizontalShimmerSpeed;
uniform float horizontalShimmerCount;
uniform float edgeGlowStrength;
uniform float edgeGlowWidth;
uniform float edgeCornerGlowWidth;
uniform float edgeCornerGlowPower;
uniform float edgeGlowSoftness;
uniform float vignetteStrength;
uniform float vignetteEdgeBypassStrength;
uniform float vignetteEdgeBypassWidth;
uniform float effectCutoff;
uniform float effectGain;
uniform float lineWarpStrength;
uniform vec2 resolution;
uniform vec3 baseColor;
uniform vec3 glowColor;
uniform vec3 scanlineColor;

float hash(vec2 p) {
  return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453123);
}

float lineWarp(vec2 uv, float t) {
  float y = 1.0 - uv.y;
  float lineId = floor((y * scanlineCount) / 6.2831853);
  float waveTime = t * waveSpeed;
  float slowWave = sin(y * waveSlowScale + waveTime * 0.9) * 0.0035;
  float mediumWave = sin(y * waveMediumScale + waveTime * 1.7) * 0.0018;
  float fineWave = sin(y * waveFineScale + waveTime * 3.2) * 0.0007;
  float lineJitter = (hash(vec2(lineId, 17.0)) - 0.5) * lineJitterStrength;
  return (slowWave + mediumWave + fineWave + lineJitter) * waveStrength * lineWarpStrength;
}

void main() {
  vec2 uv = v_uv;
  vec2 warpedUv = uv;
  float t = time * animationSpeed;
  float screenY = 1.0 - uv.y;
  warpedUv.x += lineWarp(uv, t);

  float scan = pow(0.5 + 0.5 * sin(screenY * scanlineCount + warpedUv.x * 4.0), max(scanlineHardness, 0.001));
  float flickerPhase =
    t * flickerSpeed +
    sin(t * flickerVarianceSpeed) * flickerSpeedVariance +
    sin(t * flickerVarianceSpeed * 2.17 + 1.9) * flickerSpeedVariance * 0.45 +
    sin(t * flickerVarianceSpeed * 0.41 + 4.2) * flickerSpeedVariance * 0.65;
  float flickerSignal =
    sin(flickerPhase) +
    sin(flickerPhase * 2.31 + 1.3) * flickerSecondaryStrength +
    sin(flickerPhase * 4.17 + 3.8) * flickerSecondaryStrength * 0.35;
  flickerSignal /= 1.0 + flickerSecondaryStrength + flickerSecondaryStrength * 0.35;
  float flicker = 1.0 + flickerSignal * flickerStrength;

  float lineId = floor((screenY * scanlineCount) / 6.2831853);
  float segmentX = warpedUv.x * scanlineBreakupSegments;
  float segmentId = floor(segmentX);
  float segmentT = fract(segmentX);
  segmentT = segmentT * segmentT * (3.0 - 2.0 * segmentT);
  float breakNoise = mix(hash(vec2(segmentId, lineId)), hash(vec2(segmentId + 1.0, lineId)), segmentT);
  float breakKeep = smoothstep(
    scanlineBreakupCutoff - scanlineBreakupSoftness,
    scanlineBreakupCutoff + scanlineBreakupSoftness,
    breakNoise
  );
  float lineBreakup = mix(1.0, breakKeep, scanlineBreakupStrength);
  float lineVariance = mix(
    1.0 - scanlineLineVarianceStrength,
    1.0 + scanlineLineVarianceStrength,
    hash(vec2(lineId, 19.37))
  );
  float scanlineDarken = scan * scanlineStrength * flicker * lineBreakup * lineVariance;

  float rollTime = mod(t * rollSpeed, max(rollInterval, 0.001));
  float rollProgress = clamp(rollTime / max(rollDuration, 0.001), 0.0, 1.0);
  float rollGate = (1.0 - step(rollDuration, rollTime)) * smoothstep(0.0, 0.12, rollProgress) * (1.0 - smoothstep(0.88, 1.0, rollProgress));
  float roll = (1.0 - smoothstep(0.0, rollWidth, abs(screenY - rollProgress))) * rollStrength * rollGate;
  roll *= 1.0 - rollHorizontalVariation + rollHorizontalVariation * (0.85 + 0.15 * sin(warpedUv.x * 22.0 + t * 2.0));

  float shimmer =
    (0.5 + 0.5 * (sin(screenY * horizontalShimmerCount + t * horizontalShimmerSpeed) + sin(screenY * 9.0 + warpedUv.x * 5.5 + t * 1.35) * 0.25)) *
    horizontalShimmerStrength;

  float edgeDist = min(min(uv.x, 1.0 - uv.x), min(uv.y, 1.0 - uv.y));
  vec2 cornerUv = abs(uv * 2.0 - 1.0);
  float cornerness = pow(cornerUv.x * cornerUv.y, max(edgeCornerGlowPower, 0.001));
  float edgeGlow = (1.0 - smoothstep(
    edgeGlowWidth + edgeCornerGlowWidth * cornerness,
    edgeGlowWidth + edgeCornerGlowWidth * cornerness + edgeGlowSoftness,
    edgeDist
  )) * edgeGlowStrength;

  vec2 centered = uv * 2.0 - 1.0;
  centered.x *= resolution.x / max(resolution.y, 1.0);
  float vignette = clamp(1.0 - dot(centered, centered) * vignetteStrength, 0.0, 1.0);
  float bypass = max(
    max(1.0 - smoothstep(0.0, vignetteEdgeBypassWidth, uv.x), 1.0 - smoothstep(0.0, vignetteEdgeBypassWidth, 1.0 - uv.x)),
    max(1.0 - smoothstep(0.0, vignetteEdgeBypassWidth, screenY), 1.0 - smoothstep(0.0, vignetteEdgeBypassWidth, 1.0 - screenY))
  );
  vignette = mix(vignette, 1.0, clamp(bypass * vignetteEdgeBypassStrength, 0.0, 1.0));

  float additive = max(roll + shimmer + edgeGlow - effectCutoff, 0.0) * effectGain;
  vec3 color = baseColor;
  color *= 1.0 - scanlineDarken;
  color += scanlineColor * (scan * 0.08 + shimmer * 0.06);
  color += glowColor * additive;
  color *= vignette;

  if (debugMode > 0.5) {
    vec3 debugColor = vec3(scanlineDarken, roll, shimmer);
    color = mix(color, debugColor, 0.65);
  }

  gl_FragColor = vec4(color, 1.0);
}
`;

function compileShader(gl: WebGLRenderingContext, type: number, source: string) {
  const shader = gl.createShader(type);
  if (!shader) throw new Error("Unable to create shader.");
  gl.shaderSource(shader, source);
  gl.compileShader(shader);
  if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
    const log = gl.getShaderInfoLog(shader) ?? "Unknown shader compile error.";
    gl.deleteShader(shader);
    throw new Error(log);
  }
  return shader;
}

function linkProgram(gl: WebGLRenderingContext) {
  const program = gl.createProgram();
  if (!program) throw new Error("Unable to create program.");
  const vertex = compileShader(gl, gl.VERTEX_SHADER, VERTEX_SHADER);
  const fragment = compileShader(gl, gl.FRAGMENT_SHADER, FRAGMENT_SHADER);
  gl.attachShader(program, vertex);
  gl.attachShader(program, fragment);
  gl.linkProgram(program);
  gl.deleteShader(vertex);
  gl.deleteShader(fragment);
  if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
    const log = gl.getProgramInfoLog(program) ?? "Unknown program link error.";
    gl.deleteProgram(program);
    throw new Error(log);
  }
  return program;
}

function resizeCanvas(canvas: HTMLCanvasElement, gl?: WebGLRenderingContext) {
  const rect = canvas.getBoundingClientRect();
  const dpr = window.devicePixelRatio || 1;
  const width = Math.max(1, Math.round(rect.width * dpr));
  const height = Math.max(1, Math.round(rect.height * dpr));
  if (canvas.width !== width || canvas.height !== height) {
    canvas.width = width;
    canvas.height = height;
  }
  if (gl) gl.viewport(0, 0, width, height);
}

function hexToRgb(hex: string, fallback: [number, number, number]) {
  const value = hex.replace(/^#/, "");
  if (value.length !== 6) return fallback;
  const r = Number.parseInt(value.slice(0, 2), 16);
  const g = Number.parseInt(value.slice(2, 4), 16);
  const b = Number.parseInt(value.slice(4, 6), 16);
  if ([r, g, b].some(Number.isNaN)) return fallback;
  return [r / 255, g / 255, b / 255] as [number, number, number];
}

function paletteForTint(tint: "cyan" | "yellow" | "red") {
  if (tint === "yellow") {
    return {
      base: [0.04, 0.03, 0.012] as [number, number, number],
      glow: [1, 0.845, 0.25] as [number, number, number],
      scanline: [0.18, 0.14, 0.05] as [number, number, number],
    };
  }
  if (tint === "red") {
    return {
      base: [0.045, 0.015, 0.022] as [number, number, number],
      glow: [1, 0.18, 0.26] as [number, number, number],
      scanline: [0.18, 0.06, 0.08] as [number, number, number],
    };
  }
  return {
    base: [0.004, 0.01, 0.028] as [number, number, number],
    glow: [0, 0.62, 0.72] as [number, number, number],
    scanline: [0.03, 0.15, 0.17] as [number, number, number],
  };
}

function drawFallback(canvas: HTMLCanvasElement, baseColor: [number, number, number]) {
  const ctx = canvas.getContext("2d");
  if (!ctx) return;
  resizeCanvas(canvas);
  ctx.fillStyle = `rgb(${Math.round(baseColor[0] * 255)} ${Math.round(baseColor[1] * 255)} ${Math.round(baseColor[2] * 255)})`;
  ctx.fillRect(0, 0, canvas.width, canvas.height);
  for (let y = 0; y < canvas.height; y += 4) {
    ctx.fillStyle = y % 8 === 0 ? "rgba(255,255,255,0.03)" : "rgba(0,0,0,0.06)";
    ctx.fillRect(0, y, canvas.width, 1);
  }
}

export function CrtShaderCanvas({
  className,
  enabled = true,
  tint = "cyan",
  baseColor,
  glowColor,
  scanlineColor,
  scanlineCount = 480,
  scanlineStrength = 0.22,
  scanlineHardness = 1.65,
  scanlineBreakupStrength = 0.16,
  scanlineBreakupSegments = 36,
  scanlineBreakupCutoff = 0.46,
  scanlineBreakupSoftness = 0.18,
  scanlineLineVarianceStrength = 0.08,
  waveStrength = 1,
  waveSpeed = 1,
  waveSlowScale = 18,
  waveMediumScale = 63,
  waveFineScale = 210,
  lineJitterStrength = 0.0012,
  flickerStrength = 0.025,
  flickerSpeed = 18,
  flickerSpeedVariance = 0.55,
  flickerVarianceSpeed = 1.35,
  flickerSecondaryStrength = 0.35,
  rollStrength = 0.12,
  rollInterval = 5,
  rollDuration = 1.2,
  rollWidth = 0.1,
  rollSpeed = 1,
  rollHorizontalVariation = 0.15,
  shimmerStrength,
  horizontalShimmerStrength,
  horizontalShimmerSpeed = 1.8,
  horizontalShimmerCount = 42,
  edgeGlowStrength = 0.12,
  edgeGlowWidth = 0.01,
  edgeCornerGlowWidth = 0.055,
  edgeCornerGlowPower = 2.2,
  edgeGlowSoftness = 0.018,
  vignetteStrength = 0.32,
  vignetteEdgeBypassStrength = 1,
  vignetteEdgeBypassWidth = 0.035,
  effectCutoff = 0.018,
  effectGain = 1.25,
  animationSpeed = 1,
  shaderDebug = false,
  freezeShaderTime = false,
  lineWarpStrength = 1,
}: CrtShaderCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const colors = paletteForTint(tint);
  const baseColorValue = baseColor ?? "";
  const glowColorValue = glowColor ?? "";
  const scanlineColorValue = scanlineColor ?? "";
  const resolvedHorizontalShimmerStrength = horizontalShimmerStrength ?? shimmerStrength ?? 0.055;
  const frozenTimeRef = useRef<number | null>(null);

  useEffect(() => {
    if (!enabled) return;
    const canvas = canvasRef.current;
    if (!canvas) return;

    const gl = canvas.getContext("webgl", {
      alpha: false,
      antialias: false,
      depth: false,
      stencil: false,
      premultipliedAlpha: false,
      preserveDrawingBuffer: false,
    });

    if (!gl) {
      const baseRgb = hexToRgb(baseColorValue, colors.base);
      drawFallback(canvas, baseRgb);
      const onResize = () => drawFallback(canvas, hexToRgb(baseColorValue, colors.base));
      if (typeof ResizeObserver !== "undefined") {
        const observer = new ResizeObserver(onResize);
        observer.observe(canvas);
        return () => observer.disconnect();
      }
      window.addEventListener("resize", onResize);
      return () => window.removeEventListener("resize", onResize);
    }

    let program: WebGLProgram;
    try {
      program = linkProgram(gl);
    } catch (error) {
      console.error("CrtShaderCanvas shader setup failed:", error);
      return;
    }

    gl.useProgram(program);
    const buffer = gl.createBuffer();
    if (!buffer) return;
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 1, -1, -1, 1, -1, 1, 1, -1, 1, 1]), gl.STATIC_DRAW);

    const aPosition = gl.getAttribLocation(program, "a_position");
    if (aPosition < 0) return;
    gl.enableVertexAttribArray(aPosition);
    gl.vertexAttribPointer(aPosition, 2, gl.FLOAT, false, 0, 0);
    gl.disable(gl.DEPTH_TEST);
    gl.disable(gl.CULL_FACE);
    gl.disable(gl.BLEND);

    const uniforms = {
      time: gl.getUniformLocation(program, "time"),
      animationSpeed: gl.getUniformLocation(program, "animationSpeed"),
      debugMode: gl.getUniformLocation(program, "debugMode"),
      resolution: gl.getUniformLocation(program, "resolution"),
      baseColor: gl.getUniformLocation(program, "baseColor"),
      glowColor: gl.getUniformLocation(program, "glowColor"),
      scanlineColor: gl.getUniformLocation(program, "scanlineColor"),
      scanlineCount: gl.getUniformLocation(program, "scanlineCount"),
      scanlineStrength: gl.getUniformLocation(program, "scanlineStrength"),
      scanlineHardness: gl.getUniformLocation(program, "scanlineHardness"),
      scanlineBreakupStrength: gl.getUniformLocation(program, "scanlineBreakupStrength"),
      scanlineBreakupSegments: gl.getUniformLocation(program, "scanlineBreakupSegments"),
      scanlineBreakupCutoff: gl.getUniformLocation(program, "scanlineBreakupCutoff"),
      scanlineBreakupSoftness: gl.getUniformLocation(program, "scanlineBreakupSoftness"),
      scanlineLineVarianceStrength: gl.getUniformLocation(program, "scanlineLineVarianceStrength"),
      waveStrength: gl.getUniformLocation(program, "waveStrength"),
      waveSpeed: gl.getUniformLocation(program, "waveSpeed"),
      waveSlowScale: gl.getUniformLocation(program, "waveSlowScale"),
      waveMediumScale: gl.getUniformLocation(program, "waveMediumScale"),
      waveFineScale: gl.getUniformLocation(program, "waveFineScale"),
      lineJitterStrength: gl.getUniformLocation(program, "lineJitterStrength"),
      flickerStrength: gl.getUniformLocation(program, "flickerStrength"),
      flickerSpeed: gl.getUniformLocation(program, "flickerSpeed"),
      flickerSpeedVariance: gl.getUniformLocation(program, "flickerSpeedVariance"),
      flickerVarianceSpeed: gl.getUniformLocation(program, "flickerVarianceSpeed"),
      flickerSecondaryStrength: gl.getUniformLocation(program, "flickerSecondaryStrength"),
      rollStrength: gl.getUniformLocation(program, "rollStrength"),
      rollInterval: gl.getUniformLocation(program, "rollInterval"),
      rollDuration: gl.getUniformLocation(program, "rollDuration"),
      rollWidth: gl.getUniformLocation(program, "rollWidth"),
      rollSpeed: gl.getUniformLocation(program, "rollSpeed"),
      rollHorizontalVariation: gl.getUniformLocation(program, "rollHorizontalVariation"),
      horizontalShimmerStrength: gl.getUniformLocation(program, "horizontalShimmerStrength"),
      horizontalShimmerSpeed: gl.getUniformLocation(program, "horizontalShimmerSpeed"),
      horizontalShimmerCount: gl.getUniformLocation(program, "horizontalShimmerCount"),
      edgeGlowStrength: gl.getUniformLocation(program, "edgeGlowStrength"),
      edgeGlowWidth: gl.getUniformLocation(program, "edgeGlowWidth"),
      edgeCornerGlowWidth: gl.getUniformLocation(program, "edgeCornerGlowWidth"),
      edgeCornerGlowPower: gl.getUniformLocation(program, "edgeCornerGlowPower"),
      edgeGlowSoftness: gl.getUniformLocation(program, "edgeGlowSoftness"),
      vignetteStrength: gl.getUniformLocation(program, "vignetteStrength"),
      vignetteEdgeBypassStrength: gl.getUniformLocation(program, "vignetteEdgeBypassStrength"),
      vignetteEdgeBypassWidth: gl.getUniformLocation(program, "vignetteEdgeBypassWidth"),
      effectCutoff: gl.getUniformLocation(program, "effectCutoff"),
      effectGain: gl.getUniformLocation(program, "effectGain"),
      lineWarpStrength: gl.getUniformLocation(program, "lineWarpStrength"),
    };

    const set1 = (loc: WebGLUniformLocation | null, value: number) => {
      if (loc) gl.uniform1f(loc, value);
    };

    const set3 = (loc: WebGLUniformLocation | null, rgb: [number, number, number]) => {
      if (loc) gl.uniform3f(loc, rgb[0], rgb[1], rgb[2]);
    };

    let rafId = 0;
    let resizeObserver: ResizeObserver | null = null;
    let onResize: (() => void) | null = null;

    const render = (now: number) => {
      resizeCanvas(canvas, gl);
      const shaderTime = freezeShaderTime
        ? (frozenTimeRef.current ??= now * 0.001)
        : now * 0.001;
      if (uniforms.time) gl.uniform1f(uniforms.time, shaderTime);
      set1(uniforms.animationSpeed, animationSpeed);
      set1(uniforms.debugMode, shaderDebug ? 1 : 0);
      if (uniforms.resolution) gl.uniform2f(uniforms.resolution, canvas.width, canvas.height);
      set3(uniforms.baseColor, hexToRgb(baseColorValue, colors.base));
      set3(uniforms.glowColor, hexToRgb(glowColorValue, colors.glow));
      set3(uniforms.scanlineColor, hexToRgb(scanlineColorValue, colors.scanline));
      set1(uniforms.scanlineCount, scanlineCount);
      set1(uniforms.scanlineStrength, scanlineStrength);
      set1(uniforms.scanlineHardness, scanlineHardness);
      set1(uniforms.scanlineBreakupStrength, scanlineBreakupStrength);
      set1(uniforms.scanlineBreakupSegments, scanlineBreakupSegments);
      set1(uniforms.scanlineBreakupCutoff, scanlineBreakupCutoff);
      set1(uniforms.scanlineBreakupSoftness, scanlineBreakupSoftness);
      set1(uniforms.scanlineLineVarianceStrength, scanlineLineVarianceStrength);
      set1(uniforms.waveStrength, waveStrength);
      set1(uniforms.waveSpeed, waveSpeed);
      set1(uniforms.waveSlowScale, waveSlowScale);
      set1(uniforms.waveMediumScale, waveMediumScale);
      set1(uniforms.waveFineScale, waveFineScale);
      set1(uniforms.lineJitterStrength, lineJitterStrength);
      set1(uniforms.flickerStrength, flickerStrength);
      set1(uniforms.flickerSpeed, flickerSpeed);
      set1(uniforms.flickerSpeedVariance, flickerSpeedVariance);
      set1(uniforms.flickerVarianceSpeed, flickerVarianceSpeed);
      set1(uniforms.flickerSecondaryStrength, flickerSecondaryStrength);
      set1(uniforms.rollStrength, rollStrength);
      set1(uniforms.rollInterval, rollInterval);
      set1(uniforms.rollDuration, rollDuration);
      set1(uniforms.rollWidth, rollWidth);
      set1(uniforms.rollSpeed, rollSpeed);
      set1(uniforms.rollHorizontalVariation, rollHorizontalVariation);
      set1(uniforms.horizontalShimmerStrength, resolvedHorizontalShimmerStrength);
      set1(uniforms.horizontalShimmerSpeed, horizontalShimmerSpeed);
      set1(uniforms.horizontalShimmerCount, horizontalShimmerCount);
      set1(uniforms.edgeGlowStrength, edgeGlowStrength);
      set1(uniforms.edgeGlowWidth, edgeGlowWidth);
      set1(uniforms.edgeCornerGlowWidth, edgeCornerGlowWidth);
      set1(uniforms.edgeCornerGlowPower, edgeCornerGlowPower);
      set1(uniforms.edgeGlowSoftness, edgeGlowSoftness);
      set1(uniforms.vignetteStrength, vignetteStrength);
      set1(uniforms.vignetteEdgeBypassStrength, vignetteEdgeBypassStrength);
      set1(uniforms.vignetteEdgeBypassWidth, vignetteEdgeBypassWidth);
      set1(uniforms.effectCutoff, effectCutoff);
      set1(uniforms.effectGain, effectGain);
      set1(uniforms.lineWarpStrength, lineWarpStrength);
      gl.drawArrays(gl.TRIANGLES, 0, 6);
      rafId = window.requestAnimationFrame(render);
    };

    const onCanvasResize = () => resizeCanvas(canvas, gl);
    resizeCanvas(canvas, gl);
    rafId = window.requestAnimationFrame(render);
    if (typeof ResizeObserver !== "undefined") {
      resizeObserver = new ResizeObserver(onCanvasResize);
      resizeObserver.observe(canvas);
    } else {
      onResize = onCanvasResize;
      window.addEventListener("resize", onResize);
    }

    return () => {
      window.cancelAnimationFrame(rafId);
      resizeObserver?.disconnect();
      if (onResize) window.removeEventListener("resize", onResize);
    };
  }, [
    animationSpeed,
    enabled,
    effectCutoff,
    effectGain,
    edgeCornerGlowPower,
    edgeCornerGlowWidth,
    edgeGlowSoftness,
    edgeGlowStrength,
    edgeGlowWidth,
    flickerSecondaryStrength,
    flickerSpeed,
    flickerSpeedVariance,
    flickerStrength,
    flickerVarianceSpeed,
    freezeShaderTime,
    horizontalShimmerCount,
    horizontalShimmerSpeed,
    lineJitterStrength,
    lineWarpStrength,
    scanlineBreakupCutoff,
    scanlineBreakupSegments,
    scanlineBreakupSoftness,
    scanlineBreakupStrength,
    scanlineCount,
    scanlineHardness,
    scanlineLineVarianceStrength,
    scanlineStrength,
    shaderDebug,
    baseColorValue,
    glowColorValue,
    scanlineColorValue,
    tint,
    rollDuration,
    rollHorizontalVariation,
    rollInterval,
    rollSpeed,
    rollStrength,
    rollWidth,
    vignetteEdgeBypassStrength,
    vignetteEdgeBypassWidth,
    vignetteStrength,
    resolvedHorizontalShimmerStrength,
    waveFineScale,
    waveMediumScale,
    waveSlowScale,
    waveSpeed,
    waveStrength,
  ]);

  if (!enabled) return null;
  return <canvas ref={canvasRef} className={className} aria-hidden="true" />;
}
