import type { ReactNode } from "react";

import styles from "./CardFrame.module.css";

type CardFrameProps = {
  title?: string;
  eyebrow?: string;
  tone?: "cyan" | "yellow" | "red";
  padding?: "sm" | "md" | "lg";
  children?: ReactNode;
};

export function CardFrame({
  title,
  eyebrow,
  tone = "cyan",
  padding = "md",
  children,
}: CardFrameProps) {
  return (
    <section className={styles.root} data-tone={tone} data-padding={padding}>
      <div className={styles.background} />
      <img
        className={styles.frame}
        src="/assets/ui/card_frame.png"
        alt=""
        aria-hidden="true"
      />
      <div className={styles.body}>
        {eyebrow ? <p className={styles.eyebrow}>{eyebrow}</p> : null}
        {title ? <h2 className={styles.title}>{title}</h2> : null}
        <div className={styles.content}>{children}</div>
      </div>
    </section>
  );
}
