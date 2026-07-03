export const siteName = "Laughing Skull Entertainment";
export const defaultShareImagePath = "/assets/ui/thumbnailv2.png";
export const defaultThemeColor = "#00e5ff";

export type PageMetadata = {
  title: string;
  description: string;
  canonicalUrl: string;
  type: "website" | "article";
  imageUrl: string;
  imageAlt: string;
  imageWidth: number;
  imageHeight: number;
  twitterCard: "summary" | "summary_large_image";
  themeColor: string;
  publishedTime?: string;
};

type BuildPageMetadataInput = {
  site: URL | string;
  title: string;
  description: string;
  path: string;
  type?: "website" | "article";
  imagePath?: string;
  imageAlt?: string;
  imageWidth?: number;
  imageHeight?: number;
  publishedTime?: string;
};

export function buildPageMetadata({
  site,
  title,
  description,
  path,
  type = "website",
  imagePath = defaultShareImagePath,
  imageAlt = "",
  imageWidth = 1200,
  imageHeight = 630,
  publishedTime,
}: BuildPageMetadataInput): PageMetadata {
  const resolvedImagePath = imagePath.trim() || defaultShareImagePath;

  return {
    title,
    description,
    canonicalUrl: new URL(path, site).toString(),
    type,
    imageUrl: new URL(resolvedImagePath, site).toString(),
    imageAlt,
    imageWidth,
    imageHeight,
    twitterCard: "summary_large_image",
    themeColor: defaultThemeColor,
    publishedTime,
  };
}
