import type { ReactNode } from "react";

import { CrtShaderCanvas } from "./CrtShaderCanvas";
import styles from "./CrtMediaFrame.module.css";

type CrtMediaFrameProps = {
  src?: string;
  alt?: string;
  caption?: string;
  aspectRatio?: "16 / 9" | "4 / 3" | "1 / 1";
  fit?: "cover" | "contain";
  tint?: "cyan" | "yellow" | "red";
  children?: ReactNode;
};

export function CrtMediaFrame({
  src,
  alt = "",
  caption,
  aspectRatio = "16 / 9",
  fit = "cover",
  tint = "cyan",
  children,
}: CrtMediaFrameProps) {
  const mediaAlt = src ? alt : "";

  return (
    <figure
      className={styles.root}
      data-tint={tint}
      style={{ ["--crt-aspect-ratio" as string]: aspectRatio }}
    >
      <div className={styles.shell}>
        <div className={styles.viewport} data-fit={fit}>
          <div className={styles.mediaLayer}>
            {src ? (
              <img className={styles.media} src={src} alt={mediaAlt} />
            ) : children ? (
              <div className={styles.children}>{children}</div>
            ) : null}
          </div>
          <CrtShaderCanvas className={styles.shaderCanvas} tint={tint} />
        </div>
        <img
          className={styles.frame}
          src="/assets/ui/media_frame.png"
          alt=""
          aria-hidden="true"
        />
      </div>
      {caption ? <figcaption className={styles.caption}>{caption}</figcaption> : null}
    </figure>
  );
}
