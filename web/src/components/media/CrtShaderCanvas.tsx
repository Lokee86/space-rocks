import { useEffect, useRef } from "react";

import styles from "./CrtMediaFrame.module.css";

const VERTEX_SHADER_SOURCE = `
attribute vec2 a_position;

varying vec2 v_uv;

void main() {
  v_uv = a_position * 0.5 + 0.5;
  gl_Position = vec4(a_position, 0.0, 1.0);
}
`;

const FRAGMENT_SHADER_SOURCE = `
precision highp float;

varying vec2 v_uv;

uniform float time;
uniform vec2 resolution;
uniform vec3 baseColor;
uniform float scanlineCount;
uniform float scanlineStrength;
uniform float scanlineHardness;
uniform float flickerStrength;
uniform float flickerSpeed;
uniform float scanlineBreakupStrength;
uniform float scanlineBreakupSegments;
uniform float rollStrength;
uniform float rollInterval;
uniform float rollDuration;
uniform float rollWidth;
uniform float horizontalShimmerStrength;
uniform float horizontalShimmerSpeed;
uniform float horizontalShimmerCount;
uniform float edgeGlowStrength;
uniform float edgeGlowWidth;
uniform float edgeCornerGlowWidth;
uniform float edgeCornerGlowPower;
uniform float vignetteStrength;
uniform vec3 glowColor;
uniform vec3 scanlineLightColor;

float hash(vec2 p) {
  return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453123);
}

void main() {
  vec2 uv = v_uv;
  float screenY = 1.0 - uv.y;
  vec3 color = baseColor;

  float scanlineWave = sin(screenY * scanlineCount);
  float scanline = 0.5 + 0.5 * scanlineWave;
  scanline = pow(scanline, max(scanlineHardness, 0.001));

  float flickerPhase = time * flickerSpeed;
  flickerPhase += sin(time * 1.35) * 0.55;
  flickerPhase += sin(time * 2.17 + 1.9) * 0.45;
  flickerPhase += sin(time * 0.41 + 4.2) * 0.65;

  float flickerSignal = sin(flickerPhase);
  flickerSignal += sin(flickerPhase * 2.31 + 1.3) * 0.35;
  flickerSignal += sin(flickerPhase * 4.17 + 3.8) * 0.12;
  flickerSignal /= 1.47;

  float flicker = 1.0 + flickerSignal * flickerStrength;

  float tau = 6.28318530718;
  float lineId = floor((screenY * scanlineCount) / tau);

  float segmentX = uv.x * scanlineBreakupSegments;
  float segmentId = floor(segmentX);
  float segmentT = fract(segmentX);
  segmentT = segmentT * segmentT * (3.0 - 2.0 * segmentT);

  float breakA = hash(vec2(segmentId, lineId));
  float breakB = hash(vec2(segmentId + 1.0, lineId));
  float breakNoise = mix(breakA, breakB, segmentT);
  float breakKeep = smoothstep(0.28, 0.64, breakNoise);
  float lineBreakup = mix(1.0, breakKeep, scanlineBreakupStrength);

  float lineVariance = hash(vec2(lineId, 19.37));
  lineVariance = mix(0.92, 1.08, lineVariance);

  float scanlineDarken = scanline * scanlineStrength * flicker * lineBreakup * lineVariance;

  float rollTime = mod(time, max(rollInterval, 0.001));
  float rollActive = 1.0 - step(rollDuration, rollTime);
  float rollProgress = clamp(rollTime / max(rollDuration, 0.001), 0.0, 1.0);
  float rollFadeIn = smoothstep(0.0, 0.12, rollProgress);
  float rollFadeOut = 1.0 - smoothstep(0.88, 1.0, rollProgress);
  float rollGate = rollActive * rollFadeIn * rollFadeOut;
  float rollDist = abs(screenY - rollProgress);
  float roll = 1.0 - smoothstep(0.0, rollWidth, rollDist);
  roll *= rollStrength * rollGate;

  float shimmer = sin((screenY * horizontalShimmerCount) + (time * horizontalShimmerSpeed));
  shimmer = 0.5 + 0.5 * shimmer;
  shimmer *= horizontalShimmerStrength;

  float edgeDist = min(min(uv.x, 1.0 - uv.x), min(uv.y, 1.0 - uv.y));

  vec2 cornerUv = abs(uv * 2.0 - 1.0);
  float cornerness = cornerUv.x * cornerUv.y;
  cornerness = pow(cornerness, max(edgeCornerGlowPower, 0.001));

  float dynamicEdgeWidth = edgeGlowWidth + edgeCornerGlowWidth * cornerness;
  float edgeGlow = 1.0 - smoothstep(dynamicEdgeWidth, dynamicEdgeWidth + 0.02, edgeDist);
  edgeGlow *= edgeGlowStrength;

  vec2 centered = uv * 2.0 - 1.0;
  centered.x *= resolution.x / max(resolution.y, 1.0);
  float vignette = 1.0 - dot(centered, centered) * vignetteStrength;
  vignette = clamp(vignette, 0.0, 1.0);
  float bypassLeft = 1.0 - smoothstep(0.0, 0.035, uv.x);
  float bypassRight = 1.0 - smoothstep(0.0, 0.035, 1.0 - uv.x);
  float bypassTop = 1.0 - smoothstep(0.0, 0.035, screenY);
  float bypassBottom = 1.0 - smoothstep(0.0, 0.035, 1.0 - screenY);
  float vignetteBypass = max(max(bypassLeft, bypassRight), max(bypassTop, bypassBottom));
  vignette = mix(vignette, 1.0, clamp(vignetteBypass, 0.0, 1.0));

  float brighten = max((roll + shimmer + edgeGlow) - 0.018, 0.0) * 1.45;

  color *= 1.0 - scanlineDarken;
  color += scanlineLightColor * (scanline * 0.16 + shimmer * 0.14);
  color += glowColor * brighten;
  color *= vignette;

  gl_FragColor = vec4(color, 1.0);
}
`;

type CrtShaderCanvasProps = {
  className?: string;
  tint?: "cyan" | "yellow" | "red";
};

function createShader(gl: WebGLRenderingContext, type: number, source: string) {
  const shader = gl.createShader(type);

  if (!shader) {
    throw new Error("Unable to create shader.");
  }

  gl.shaderSource(shader, source);
  gl.compileShader(shader);

  if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
    const log = gl.getShaderInfoLog(shader) ?? "Unknown shader compile error.";
    gl.deleteShader(shader);
    throw new Error(log);
  }

  return shader;
}

function createProgram(gl: WebGLRenderingContext) {
  const vertexShader = createShader(gl, gl.VERTEX_SHADER, VERTEX_SHADER_SOURCE);
  const fragmentShader = createShader(gl, gl.FRAGMENT_SHADER, FRAGMENT_SHADER_SOURCE);
  const program = gl.createProgram();

  if (!program) {
    throw new Error("Unable to create program.");
  }

  gl.attachShader(program, vertexShader);
  gl.attachShader(program, fragmentShader);
  gl.linkProgram(program);

  gl.deleteShader(vertexShader);
  gl.deleteShader(fragmentShader);

  if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
    const log = gl.getProgramInfoLog(program) ?? "Unknown program link error.";
    gl.deleteProgram(program);
    throw new Error(log);
  }

  return program;
}

function getUniformLocation(
  gl: WebGLRenderingContext,
  program: WebGLProgram,
  name: string,
) {
  const location = gl.getUniformLocation(program, name);

  if (!location) {
    throw new Error(`Missing uniform: ${name}`);
  }

  return location;
}

function getPalette(tint: "cyan" | "yellow" | "red" = "cyan") {
  if (tint === "yellow") {
    return {
      base: [0.062, 0.051, 0.018] as [number, number, number],
      glow: [1.0, 0.845, 0.25] as [number, number, number],
      scanlineLight: [0.3, 0.26, 0.11] as [number, number, number],
    };
  }

  if (tint === "red") {
    return {
      base: [0.07, 0.022, 0.03] as [number, number, number],
      glow: [1.0, 0.2, 0.26] as [number, number, number],
      scanlineLight: [0.3, 0.09, 0.12] as [number, number, number],
    };
  }

  return {
    base: [0.008, 0.033, 0.051] as [number, number, number],
    glow: [0.0, 0.898, 1.0] as [number, number, number],
    scanlineLight: [0.065, 0.26, 0.3] as [number, number, number],
  };
}

function drawFailureFallbackFrame(canvas: HTMLCanvasElement) {
  const context = canvas.getContext("2d");

  if (!context) {
    return;
  }

  const rect = canvas.getBoundingClientRect();
  const dpr = window.devicePixelRatio || 1;
  const width = Math.max(1, Math.round(rect.width * dpr));
  const height = Math.max(1, Math.round(rect.height * dpr));

  canvas.width = width;
  canvas.height = height;

  context.setTransform(1, 0, 0, 1, 0, 0);
  context.clearRect(0, 0, width, height);
  context.fillStyle = "#020617";
  context.fillRect(0, 0, width, height);

  for (let y = 0; y < height; y += 4) {
    context.fillStyle = y % 8 === 0 ? "rgba(0, 255, 255, 0.55)" : "rgba(255, 0, 255, 0.42)";
    context.fillRect(0, y, width, 1);
  }

  context.fillStyle = "rgba(0, 255, 255, 0.14)";
  context.fillRect(0, Math.floor(height * 0.46), width, Math.max(2, Math.floor(height * 0.06)));

  const gradient = context.createRadialGradient(
    width * 0.5,
    height * 0.5,
    Math.min(width, height) * 0.08,
    width * 0.5,
    height * 0.5,
    Math.max(width, height) * 0.65,
  );

  gradient.addColorStop(0, "rgba(0, 255, 255, 0.16)");
  gradient.addColorStop(0.6, "rgba(0, 0, 0, 0.08)");
  gradient.addColorStop(1, "rgba(0, 0, 0, 0.84)");
  context.fillStyle = gradient;
  context.fillRect(0, 0, width, height);
}

export function CrtShaderCanvas({ className, tint = "cyan" }: CrtShaderCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const resolvedClassName = className ?? styles.shaderCanvas;
  const palette = getPalette(tint);

  useEffect(() => {
    const canvas = canvasRef.current;

    if (!canvas) {
      return;
    }

    let rafId = 0;
    let resizeObserver: ResizeObserver | null = null;
    let windowResizeHandler: (() => void) | null = null;
    let gl: WebGLRenderingContext | null = null;

    try {
      gl = canvas.getContext("webgl", {
        alpha: false,
        antialias: false,
        depth: false,
        stencil: false,
        premultipliedAlpha: false,
        preserveDrawingBuffer: false,
      });
    } catch {
      gl = null;
    }

    if (!gl) {
      console.warn("CrtShaderCanvas: WebGL unavailable, using failure fallback.");
      drawFailureFallbackFrame(canvas);

      if (typeof ResizeObserver !== "undefined") {
        resizeObserver = new ResizeObserver(() => drawFailureFallbackFrame(canvas));
        resizeObserver.observe(canvas);
      } else {
        windowResizeHandler = () => drawFailureFallbackFrame(canvas);
        window.addEventListener("resize", windowResizeHandler);
        return () => {
          if (windowResizeHandler) {
            window.removeEventListener("resize", windowResizeHandler);
          }
        };
      }

      return () => {
        resizeObserver?.disconnect();
      };
    }

    try {
      const program = createProgram(gl);
      gl.useProgram(program);
      gl.disable(gl.DEPTH_TEST);
      gl.disable(gl.CULL_FACE);
      gl.disable(gl.BLEND);
      let attributeLocation = -1;

      const uniforms = {
        time: getUniformLocation(gl, program, "time"),
        resolution: getUniformLocation(gl, program, "resolution"),
        baseColor: getUniformLocation(gl, program, "baseColor"),
        scanlineCount: getUniformLocation(gl, program, "scanlineCount"),
        scanlineStrength: getUniformLocation(gl, program, "scanlineStrength"),
        scanlineHardness: getUniformLocation(gl, program, "scanlineHardness"),
        flickerStrength: getUniformLocation(gl, program, "flickerStrength"),
        flickerSpeed: getUniformLocation(gl, program, "flickerSpeed"),
        scanlineBreakupStrength: getUniformLocation(
          gl,
          program,
          "scanlineBreakupStrength",
        ),
        scanlineBreakupSegments: getUniformLocation(
          gl,
          program,
          "scanlineBreakupSegments",
        ),
        rollStrength: getUniformLocation(gl, program, "rollStrength"),
        rollInterval: getUniformLocation(gl, program, "rollInterval"),
        rollDuration: getUniformLocation(gl, program, "rollDuration"),
        rollWidth: getUniformLocation(gl, program, "rollWidth"),
        horizontalShimmerStrength: getUniformLocation(
          gl,
          program,
          "horizontalShimmerStrength",
        ),
        horizontalShimmerSpeed: getUniformLocation(
          gl,
          program,
          "horizontalShimmerSpeed",
        ),
        horizontalShimmerCount: getUniformLocation(
          gl,
          program,
          "horizontalShimmerCount",
        ),
        edgeGlowStrength: getUniformLocation(gl, program, "edgeGlowStrength"),
        edgeGlowWidth: getUniformLocation(gl, program, "edgeGlowWidth"),
        edgeCornerGlowWidth: getUniformLocation(
          gl,
          program,
          "edgeCornerGlowWidth",
        ),
        edgeCornerGlowPower: getUniformLocation(
          gl,
          program,
          "edgeCornerGlowPower",
        ),
        vignetteStrength: getUniformLocation(gl, program, "vignetteStrength"),
        glowColor: getUniformLocation(gl, program, "glowColor"),
        scanlineLightColor: getUniformLocation(
          gl,
          program,
          "scanlineLightColor",
        ),
      };

      const positionBuffer = gl.createBuffer();

      if (!positionBuffer) {
        throw new Error("Unable to create buffer.");
      }

      gl.bindBuffer(gl.ARRAY_BUFFER, positionBuffer);
      gl.bufferData(
        gl.ARRAY_BUFFER,
        new Float32Array([
          -1, -1,
          1, -1,
          -1, 1,
          -1, 1,
          1, -1,
          1, 1,
        ]),
        gl.STATIC_DRAW,
      );

      attributeLocation = gl.getAttribLocation(program, "a_position");
      if (attributeLocation < 0) {
        throw new Error("Missing attribute: a_position");
      }
      gl.enableVertexAttribArray(attributeLocation);
      gl.vertexAttribPointer(attributeLocation, 2, gl.FLOAT, false, 0, 0);

      const syncSize = () => {
        const rect = canvas.getBoundingClientRect();
        const dpr = window.devicePixelRatio || 1;
        const width = Math.max(1, Math.round(rect.width * dpr));
        const height = Math.max(1, Math.round(rect.height * dpr));

        if (canvas.width !== width || canvas.height !== height) {
          canvas.width = width;
          canvas.height = height;
        }

        gl.viewport(0, 0, width, height);
      };

      const draw = (now: number) => {
        syncSize();

        gl.useProgram(program);
        gl.bindBuffer(gl.ARRAY_BUFFER, positionBuffer);
        gl.vertexAttribPointer(attributeLocation, 2, gl.FLOAT, false, 0, 0);

        gl.uniform1f(uniforms.time, now * 0.001);
        gl.uniform2f(uniforms.resolution, canvas.width, canvas.height);
        gl.uniform3f(uniforms.baseColor, palette.base[0], palette.base[1], palette.base[2]);
        gl.uniform1f(uniforms.scanlineCount, 480.0);
        gl.uniform1f(uniforms.scanlineStrength, 0.22);
        gl.uniform1f(uniforms.scanlineHardness, 1.65);
        gl.uniform1f(uniforms.flickerStrength, 0.045);
        gl.uniform1f(uniforms.flickerSpeed, 18.0);
        gl.uniform1f(uniforms.scanlineBreakupStrength, 0.16);
        gl.uniform1f(uniforms.scanlineBreakupSegments, 36.0);
        gl.uniform1f(uniforms.rollStrength, 0.12);
        gl.uniform1f(uniforms.rollInterval, 5.0);
        gl.uniform1f(uniforms.rollDuration, 1.2);
        gl.uniform1f(uniforms.rollWidth, 0.1);
        gl.uniform1f(uniforms.horizontalShimmerStrength, 0.1);
        gl.uniform1f(uniforms.horizontalShimmerSpeed, 1.8);
        gl.uniform1f(uniforms.horizontalShimmerCount, 42.0);
        gl.uniform1f(uniforms.edgeGlowStrength, 0.12);
        gl.uniform1f(uniforms.edgeGlowWidth, 0.018);
        gl.uniform1f(uniforms.edgeCornerGlowWidth, 0.1);
        gl.uniform1f(uniforms.edgeCornerGlowPower, 2.2);
        gl.uniform1f(uniforms.vignetteStrength, 0.32);
        gl.uniform3f(uniforms.glowColor, palette.glow[0], palette.glow[1], palette.glow[2]);
        gl.uniform3f(
          uniforms.scanlineLightColor,
          palette.scanlineLight[0],
          palette.scanlineLight[1],
          palette.scanlineLight[2],
        );

        gl.drawArrays(gl.TRIANGLES, 0, 6);
        rafId = window.requestAnimationFrame(draw);
      };

      syncSize();
      rafId = window.requestAnimationFrame(draw);

      if (typeof ResizeObserver !== "undefined") {
        resizeObserver = new ResizeObserver(syncSize);
        resizeObserver.observe(canvas);
      } else {
        windowResizeHandler = syncSize;
        window.addEventListener("resize", windowResizeHandler);
      }
    } catch (error) {
      console.warn("CrtShaderCanvas: WebGL setup failed, using failure fallback.", error);
      drawFailureFallbackFrame(canvas);

      if (typeof ResizeObserver !== "undefined") {
        resizeObserver = new ResizeObserver(() => drawFailureFallbackFrame(canvas));
        resizeObserver.observe(canvas);
      } else {
        windowResizeHandler = () => drawFailureFallbackFrame(canvas);
        window.addEventListener("resize", windowResizeHandler);
      }
    }

    return () => {
      window.cancelAnimationFrame(rafId);
      resizeObserver?.disconnect();
      if (windowResizeHandler) {
        window.removeEventListener("resize", windowResizeHandler);
      }
    };
  }, []);

  return <canvas ref={canvasRef} className={resolvedClassName} aria-hidden="true" />;
}
