import type { CSSProperties, ReactNode } from "react";

import { CrtShaderCanvas } from "./CrtShaderCanvas";
import { MediaControlButton, type MediaControlVariant } from "./MediaControlButton";
import styles from "./CrtMediaFrame.module.css";

const CONTROL_VARIANTS: MediaControlVariant[] = ["previous", "play", "next"];

type CrtMediaFrameProps = {
  src?: string;
  alt?: string;
  caption?: string;
  aspectRatio?: "16 / 9" | "4 / 3" | "1 / 1";
  fit?: "cover" | "contain";
  tint?: "cyan" | "yellow" | "red";
  enabled?: boolean;
  scanlineStrength?: number;
  rollStrength?: number;
  shimmerStrength?: number;
  vignetteStrength?: number;
  edgeGlowStrength?: number;
  lineWarpStrength?: number;
  screenInsetLeft?: number;
  screenInsetRight?: number;
  screenInsetTop?: number;
  screenInsetBottom?: number;
  showControls?: boolean;
  disabledControls?: string[];
  children?: ReactNode;
};

export function CrtMediaFrame({
  src,
  alt = "",
  caption,
  aspectRatio = "16 / 9",
  fit = "cover",
  tint = "cyan",
  enabled = true,
  scanlineStrength = 0.22,
  rollStrength = 0.12,
  shimmerStrength = 0.1,
  vignetteStrength = 0.32,
  edgeGlowStrength = 0.12,
  lineWarpStrength = 1.0,
  screenInsetLeft = 5.0,
  screenInsetRight = 5.0,
  screenInsetTop = 11.8,
  screenInsetBottom = 12.4,
  showControls = false,
  disabledControls = [],
  children,
}: CrtMediaFrameProps) {
  const mediaAlt = src ? alt : "";
  const style: CSSProperties = {
    ["--crt-aspect-ratio" as string]: aspectRatio,
    ["--crt-screen-inset-left" as string]: `${screenInsetLeft}%`,
    ["--crt-screen-inset-right" as string]: `${screenInsetRight}%`,
    ["--crt-screen-inset-top" as string]: `${screenInsetTop}%`,
    ["--crt-screen-inset-bottom" as string]: `${screenInsetBottom}%`,
  };

  return (
    <figure className={styles.root} data-tint={tint} style={style}>
      <div className={styles.shell}>
        <div className={styles.viewport} data-fit={fit}>
          <div className={styles.mediaLayer}>
            {src ? (
              <img className={styles.media} src={src} alt={mediaAlt} />
            ) : children ? (
              <div className={styles.children}>{children}</div>
            ) : null}
          </div>
          <CrtShaderCanvas
            className={styles.shaderCanvas}
            enabled={enabled}
            tint={tint}
            scanlineStrength={scanlineStrength}
            rollStrength={rollStrength}
            shimmerStrength={shimmerStrength}
            vignetteStrength={vignetteStrength}
            edgeGlowStrength={edgeGlowStrength}
            lineWarpStrength={lineWarpStrength}
          />
        </div>
        <img
          className={styles.frame}
          src="/assets/ui/media_frame.png"
          alt=""
          aria-hidden="true"
        />
        {showControls ? (
          <div className={styles.controls} role="group" aria-label="Media controls">
            {CONTROL_VARIANTS.map((variant) => (
              <MediaControlButton
                key={variant}
                className={styles.controlButton}
                variant={variant}
                disabled={disabledControls.includes(variant)}
              />
            ))}
          </div>
        ) : null}
      </div>
      {caption ? <figcaption className={styles.caption}>{caption}</figcaption> : null}
    </figure>
  );
}
