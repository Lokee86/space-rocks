export type ArchiveEntry = {
  id: string;
  title: string;
  date: string;
  summary: string;
  href: string;
};

export type ArchiveContent = {
  entries: ArchiveEntry[];
};

export function normalizeArchiveContent(
  input: Partial<ArchiveContent>,
): ArchiveContent {
  return {
    entries: input.entries ?? [],
  };
}
