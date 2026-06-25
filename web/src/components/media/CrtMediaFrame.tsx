import { useEffect, useRef, useState, type CSSProperties, type ReactNode } from "react";

import { CrtShaderCanvas } from "./CrtShaderCanvas";
import { MediaControlButton, type MediaControlVariant } from "./MediaControlButton";
import styles from "./CrtMediaFrame.module.css";

const CONTROL_VARIANTS = ["previous", "play", "next"] as const satisfies readonly MediaControlVariant[];
const ALL_CONTROL_VARIANTS: MediaControlVariant[] = ["previous", "rewind", "play", "pause", "fastForward", "next"];

// Source-image coordinates for media_frame.png.
const BOTTOM_STRIP_SOURCE = { x: 46, y: 440, width: 592, height: 84 };
// playCenterX/Y are aligned to the red anchor pixel in media_frame.png.
const CONTROL_SOURCE = { playCenterX: 334, playCenterY: 494, buttonPitch: 68, buttonWidth: 68, buttonHeight: 46 };
const CONTROL_ROW_SOURCE = {
  x: CONTROL_SOURCE.playCenterX - CONTROL_SOURCE.buttonPitch - CONTROL_SOURCE.buttonWidth / 2,
  y: CONTROL_SOURCE.playCenterY - CONTROL_SOURCE.buttonHeight / 2,
  width: CONTROL_SOURCE.buttonWidth + CONTROL_SOURCE.buttonPitch * 2,
  height: CONTROL_SOURCE.buttonHeight,
};
type MediaMode = "imageList" | "video";
type ControlSlot = (typeof CONTROL_VARIANTS)[number];
type ControlConfig = { slot: ControlSlot; variant: MediaControlVariant; disabled: boolean; label: string; onClick?: () => void };

function percent(value: number) {
  return `${value * 100}%`;
}

function isMediaControlVariant(value: string): value is MediaControlVariant {
  return (ALL_CONTROL_VARIANTS as string[]).includes(value);
}

function parseDisabledControls(disabledControls: string | string[]) {
  const values = Array.isArray(disabledControls) ? disabledControls : disabledControls.split(",");
  return new Set(values.map((value) => value.trim()).filter(isMediaControlVariant));
}

function parseImageItems(imageItems: string | string[] | undefined) {
  const values = Array.isArray(imageItems)
    ? imageItems
    : imageItems
      ? imageItems.split(/[\n,]+/)
      : [];

  return values.map((value) => value.trim()).filter(Boolean);
}

function clampIndex(index: number, length: number) {
  if (length <= 0) return 0;
  return Math.min(Math.max(Math.trunc(index), 0), length - 1);
}

function buttonRowStyle(): CSSProperties {
  const x = (value: number) => percent(value / BOTTOM_STRIP_SOURCE.width);
  const y = (value: number) => percent(value / BOTTOM_STRIP_SOURCE.height);
  const rowLeft = CONTROL_ROW_SOURCE.x - BOTTOM_STRIP_SOURCE.x;
  const rowTop = CONTROL_ROW_SOURCE.y - BOTTOM_STRIP_SOURCE.y;

  return {
    left: x(rowLeft),
    top: y(rowTop),
    width: x(CONTROL_ROW_SOURCE.width),
    height: y(CONTROL_ROW_SOURCE.height),
  };
}

type CrtMediaFrameProps = {
  src?: string;
  alt?: string;
  caption?: string;
  aspectRatio?: "16 / 9" | "4 / 3" | "1 / 1";
  fit?: "cover" | "contain";
  tint?: "cyan" | "yellow" | "red";
  mediaMode?: MediaMode;
  imageItems?: string | string[];
  videoSrc?: string;
  autoAdvanceMs?: number;
  seekSeconds?: number;
  initialIndex?: number;
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
  disabledControls?: string | string[];
  children?: ReactNode;
};

export function CrtMediaFrame({
  src,
  alt = "",
  caption,
  aspectRatio = "16 / 9",
  fit = "cover",
  tint = "cyan",
  mediaMode,
  imageItems,
  videoSrc,
  autoAdvanceMs = 5000,
  seekSeconds = 10,
  initialIndex = 0,
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
  screenInsetBottom = 15.8,
  showControls = false,
  disabledControls = "",
  children,
}: CrtMediaFrameProps) {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const [currentIndex, setCurrentIndex] = useState(() => clampIndex(initialIndex, 1));
  const [isSlideshowPlaying, setIsSlideshowPlaying] = useState(false);
  const [isVideoPlaying, setIsVideoPlaying] = useState(false);
  const parsedImageItems = parseImageItems(imageItems);
  const effectiveMediaMode =
    mediaMode ?? (videoSrc ? "video" : parsedImageItems.length > 0 ? "imageList" : undefined);
  const imageSources =
    effectiveMediaMode === "imageList"
      ? parsedImageItems.length > 0
        ? parsedImageItems
        : src
          ? [src]
          : []
      : [];
  const imageItemsKey = imageSources.join("\n");
  const mediaAlt = src ? alt : "";
  const disabledControlSet = parseDisabledControls(disabledControls);
  const style: CSSProperties = {
    ["--crt-aspect-ratio" as string]: aspectRatio,
    ["--crt-screen-inset-left" as string]: `${screenInsetLeft}%`,
    ["--crt-screen-inset-right" as string]: `${screenInsetRight}%`,
    ["--crt-screen-inset-top" as string]: `${screenInsetTop}%`,
    ["--crt-screen-inset-bottom" as string]: `${screenInsetBottom}%`,
  };
  const currentImage = imageSources[clampIndex(currentIndex, imageSources.length)];
  const hasMultipleImages = imageSources.length > 1;
  const isImageMode = effectiveMediaMode === "imageList";
  const isVideoMode = effectiveMediaMode === "video" && Boolean(videoSrc);

  useEffect(() => {
    setCurrentIndex(clampIndex(initialIndex, imageSources.length));
    setIsSlideshowPlaying(false);
  }, [imageItemsKey, initialIndex, imageSources.length]);

  useEffect(() => {
    if (!isImageMode || !isSlideshowPlaying || imageSources.length <= 1) return;

    const intervalId = window.setInterval(() => {
      setCurrentIndex((index) => (index + 1) % imageSources.length);
    }, Math.max(1, autoAdvanceMs));

    return () => window.clearInterval(intervalId);
  }, [autoAdvanceMs, imageSources.length, isImageMode, isSlideshowPlaying]);

  const seekVideo = (delta: number) => {
    const video = videoRef.current;
    if (!video) return;
    const maxTime = Number.isFinite(video.duration) ? video.duration : Number.POSITIVE_INFINITY;
    video.currentTime = Math.min(Math.max(video.currentTime + delta, 0), maxTime);
  };

  const toggleVideoPlayback = () => {
    const video = videoRef.current;
    if (!video) return;

    if (video.paused || video.ended) {
      const playResult = video.play();
      if (playResult) {
        playResult.catch(() => setIsVideoPlaying(false));
      }
      return;
    }

    video.pause();
  };

  const isControlDisabled = (variant: MediaControlVariant, automaticDisabled: boolean, slot?: ControlSlot) =>
    automaticDisabled ||
    disabledControlSet.has(variant) ||
    (slot ? disabledControlSet.has(slot) : false) ||
    (variant === "play" && disabledControlSet.has("pause")) ||
    (variant === "pause" && disabledControlSet.has("play"));

  const controls: ControlConfig[] = isImageMode
    ? [
        { slot: "previous", variant: "previous", disabled: isControlDisabled("previous", !hasMultipleImages, "previous"), label: "Previous image", onClick: () => setCurrentIndex((index) => (index - 1 + imageSources.length) % imageSources.length) },
        {
          slot: "play",
          variant: isSlideshowPlaying ? "pause" : "play",
          disabled: isControlDisabled(isSlideshowPlaying ? "pause" : "play", !hasMultipleImages, "play"),
          label: isSlideshowPlaying ? "Pause slideshow" : "Play slideshow",
          onClick: () => setIsSlideshowPlaying((playing) => !playing),
        },
        { slot: "next", variant: "next", disabled: isControlDisabled("next", !hasMultipleImages, "next"), label: "Next image", onClick: () => setCurrentIndex((index) => (index + 1) % imageSources.length) },
      ]
    : isVideoMode
      ? [
          { slot: "previous", variant: "rewind", disabled: isControlDisabled("rewind", false, "previous"), label: `Rewind ${seekSeconds} seconds`, onClick: () => seekVideo(-seekSeconds) },
          {
            slot: "play",
            variant: isVideoPlaying ? "pause" : "play",
            disabled: isControlDisabled(isVideoPlaying ? "pause" : "play", false, "play"),
            label: isVideoPlaying ? "Pause video" : "Play video",
            onClick: toggleVideoPlayback,
          },
          { slot: "next", variant: "fastForward", disabled: isControlDisabled("fastForward", false, "next"), label: `Fast forward ${seekSeconds} seconds`, onClick: () => seekVideo(seekSeconds) },
        ]
      : [
          { slot: "previous", variant: "previous", disabled: true, label: "Previous" },
          { slot: "play", variant: "play", disabled: true, label: "Play" },
          { slot: "next", variant: "next", disabled: true, label: "Next" },
        ];

  return (
    <figure className={styles.root} data-tint={tint} style={style}>
      <div className={styles.shell}>
        <div className={styles.viewport} data-fit={fit}>
          <div className={styles.mediaLayer}>
            {isVideoMode ? (
              <video
                ref={videoRef}
                className={styles.media}
                src={videoSrc}
                playsInline
                preload="metadata"
                onPlay={() => setIsVideoPlaying(true)}
                onPause={() => setIsVideoPlaying(false)}
                onEnded={() => setIsVideoPlaying(false)}
              />
            ) : isImageMode && currentImage ? (
              <img className={styles.media} src={currentImage} alt={mediaAlt} />
            ) : src ? (
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
        <div className={styles.frame} aria-hidden="true" />
        <div className={styles.bottomTraySlot}>
          {showControls ? (
            <div
              className={styles.controls}
              role="group"
              aria-label="Media controls"
            >
              <div className={styles.buttonRow} style={buttonRowStyle()}>
                {controls.map((control) => (
                  <MediaControlButton
                    key={control.slot}
                    className={styles.controlButton}
                    variant={control.variant}
                    label={control.label}
                    disabled={control.disabled}
                    onClick={control.onClick}
                  />
                ))}
              </div>
            </div>
          ) : null}
        </div>
      </div>
      {caption ? <figcaption className={styles.caption}>{caption}</figcaption> : null}
    </figure>
  );
}
