import type { ButtonHTMLAttributes } from "react";

import styles from "./MediaControlButton.module.css";

export type MediaControlVariant =
  | "play"
  | "pause"
  | "rewind"
  | "fastForward"
  | "previous"
  | "next";

type MediaControlButtonProps = Omit<ButtonHTMLAttributes<HTMLButtonElement>, "children"> & {
  variant: MediaControlVariant;
  label?: string;
};

const LABELS: Record<MediaControlVariant, string> = {
  play: "Play",
  pause: "Pause",
  rewind: "Rewind",
  fastForward: "Fast forward",
  previous: "Previous",
  next: "Next",
};

export function MediaControlButton({
  variant,
  label = LABELS[variant],
  className,
  type = "button",
  ...buttonProps
}: MediaControlButtonProps) {
  const classes = [styles.root, className].filter(Boolean).join(" ");

  return (
    <button
      {...buttonProps}
      aria-label={buttonProps["aria-label"] ?? label}
      className={classes}
      data-variant={variant}
      type={type}
    >
      <span className={styles.icon} aria-hidden="true" />
    </button>
  );
}
