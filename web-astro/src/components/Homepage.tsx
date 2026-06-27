import * as React from "react";

import PlasmicHomepage from "./plasmic/space_rocks_devlog/PlasmicHomepage";
import { CrtMediaFrame } from "./media/CrtMediaFrame";
import { MarkdownText } from "./markdown/MarkdownText";
import markdownStyles from "./markdown/MarkdownText.module.css";
import type { HomepageContent } from "../content/homepageContent";
import { normalizeHomepageContent } from "../content/homepageContent";

type MediaFrameOverrides = {
  alt: string;
  fit: "contain";
  mediaMode?: "imageList";
  imageItems: string[];
  youtubeUrl: string;
};

export interface HomepageProps {
  content?: Partial<HomepageContent>;
}

function getMediaFrameOverrides(
  mediaKind: HomepageContent["heroMediaKind"],
  imageItems: string[],
  youtubeUrl: string,
  alt: string,
): MediaFrameOverrides {
  if (mediaKind === "images") {
    return {
      mediaMode: "imageList",
      imageItems,
      youtubeUrl: "",
      alt,
      fit: "contain",
    };
  }

  if (mediaKind === "youtube") {
    return {
      imageItems: [],
      youtubeUrl,
      alt,
      fit: "contain",
    };
  }

  return {
    imageItems: [],
    youtubeUrl: "",
    alt,
    fit: "contain",
  };
}

function Homepage_(props: HomepageProps, _ref: React.ForwardedRef<unknown>) {
  const content = normalizeHomepageContent(props.content ?? {});
  const heroMediaFrameProps = getMediaFrameOverrides(
    content.heroMediaKind,
    content.heroImages,
    content.heroYoutubeUrl,
    content.heroMediaAlt,
  );
  const articleMediaFrameProps = getMediaFrameOverrides(
    content.articleMediaKind,
    content.articleImages,
    content.articleYoutubeUrl,
    content.articleMediaAlt,
  );
  const hasPublishedDevlog =
    content.heroLine1.trim() !== "" ||
    content.articleTitle.trim() !== "" ||
    content.intro.trim() !== "";

  if (!hasPublishedDevlog) {
    return (
      <main aria-label="Devlog home">
        <h1>Space Rocks Devlog</h1>
        <p>No canonical devlog posts are published yet.</p>
      </main>
    );
  }

  return (
    <PlasmicHomepage
      overrides={{
        heroLine1Media: { children: content.heroLine1 },
        heroLine2Media: { children: content.heroLine2 },
        heroLine3Media: { children: content.heroLine3 },
        heroLine1Desktop: { children: content.heroLine1 },
        heroLine2Desktop: { children: content.heroLine2 },
        heroLine3Desktop: { children: content.heroLine3 },
        heroMediaFrame: heroMediaFrameProps,
        articleLabel: { children: content.articleLabel },
        articleTitle: { children: content.articleTitle },
        introText: {
          children: (
            <MarkdownText
              value={content.intro}
              className={markdownStyles.fullWidth}
            />
          ),
        },
        articleMediaFrame: articleMediaFrameProps,
        screenStack2: {
          children: (
            <>
              <CrtMediaFrame
                {...articleMediaFrameProps}
                aspectRatio="16 / 9"
                autoAdvanceMs={5000}
                showControls={true}
              />
              <MarkdownText
                value={content.body}
                className={`devlogBody ${markdownStyles.fullWidth}`}
              />
            </>
          ),
        },
        finishedTitle: { children: content.finishedTitle },
        finishedBody: {
          children: <MarkdownText value={content.finishedBody} />,
        },
        nowTitle: { children: content.nowTitle },
        nowBody: { children: <MarkdownText value={content.nowBody} /> },
        comingUpTitle: { children: content.comingUpTitle },
        comingUpBody: {
          children: <MarkdownText value={content.comingUpBody} />,
        },
        utilityTitle: { children: content.utilityTitle },
        utilityText: {
          children: <MarkdownText value={content.utilityText} />,
        },
        finePrint: {
          children: <MarkdownText value={content.finePrint} />,
        },
      }}
    />
  );
}

const Homepage = React.forwardRef(Homepage_);
export default Homepage;
