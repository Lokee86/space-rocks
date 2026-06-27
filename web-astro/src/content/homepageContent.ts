export type HomepageMediaKind = "" | "images" | "youtube";

export type HomepageContent = {
  heroLine1: string;
  heroLine2: string;
  heroLine3: string;
  heroMediaKind: HomepageMediaKind;
  heroImages: string[];
  heroYoutubeUrl: string;
  heroMediaAlt: string;
  articleLabel: string;
  articleTitle: string;
  intro: string;
  articleMediaKind: HomepageMediaKind;
  articleImages: string[];
  articleYoutubeUrl: string;
  articleMediaAlt: string;
  finishedTitle: string;
  nowTitle: string;
  comingUpTitle: string;
  finishedBody: string;
  nowBody: string;
  comingUpBody: string;
  body: string;
  utilityTitle: string;
  utilityText: string;
  finePrint: string;
};

const textContentFields: (keyof HomepageContent)[] = [
  "heroLine1",
  "heroLine2",
  "heroLine3",
  "heroYoutubeUrl",
  "heroMediaAlt",
  "articleLabel",
  "articleTitle",
  "intro",
  "articleYoutubeUrl",
  "articleMediaAlt",
  "finishedTitle",
  "nowTitle",
  "comingUpTitle",
  "finishedBody",
  "nowBody",
  "comingUpBody",
  "body",
  "utilityTitle",
  "utilityText",
  "finePrint",
];

const mediaKindFields: Array<"heroMediaKind" | "articleMediaKind"> = [
  "heroMediaKind",
  "articleMediaKind",
];

const imageArrayFields: Array<"heroImages" | "articleImages"> = [
  "heroImages",
  "articleImages",
];

export function normalizeHomepageContent(
  input: Partial<HomepageContent>,
): HomepageContent {
  const normalized = {} as HomepageContent;

  for (const field of textContentFields) {
    normalized[field] = input[field] ?? "";
  }

  for (const field of mediaKindFields) {
    normalized[field] = input[field] ?? "";
  }

  for (const field of imageArrayFields) {
    normalized[field] = input[field] ?? [];
  }

  return normalized;
}
