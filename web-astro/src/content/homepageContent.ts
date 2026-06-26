export type HomepageContent = {
  heroLine1: string;
  heroLine2: string;
  heroLine3: string;
  articleLabel: string;
  articleTitle: string;
  intro: string;
  whatChanged: string;
  callout: string;
  whatsNext: string;
  finishedBody: string;
  nowBody: string;
  comingUpBody: string;
  utilityText: string;
  finePrint: string;
};

const homepageContentFields: (keyof HomepageContent)[] = [
  "heroLine1",
  "heroLine2",
  "heroLine3",
  "articleLabel",
  "articleTitle",
  "intro",
  "whatChanged",
  "callout",
  "whatsNext",
  "finishedBody",
  "nowBody",
  "comingUpBody",
  "utilityText",
  "finePrint",
];

export function normalizeHomepageContent(
  input: Partial<HomepageContent>,
): HomepageContent {
  const normalized = {} as HomepageContent;

  for (const field of homepageContentFields) {
    normalized[field] = input[field] ?? "";
  }

  return normalized;
}
