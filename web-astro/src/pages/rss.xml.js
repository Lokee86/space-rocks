import rss from "@astrojs/rss";
import { getCollection } from "astro:content";

function stripSummaryMarkdown(summary) {
  return summary
    .replace(/\*\*Status:\*\*\s*/gi, "")
    .replace(/\*\*/g, "")
    .replace(/__(.*?)__/g, "$1")
    .replace(/`(.*?)`/g, "$1")
    .replace(/\[(.*?)\]\((.*?)\)/g, "$1")
    .trim();
}

export async function GET(context) {
  const entries = await getCollection("devlog");
  const sortedEntries = [...entries].sort(
    (a, b) => b.data.date.getTime() - a.data.date.getTime(),
  );

  return rss({
    title: "Space Rocks! Devlog",
    description:
      "Development updates, transmissions, and progress reports for Space Rocks!",
    site: context.site,
    customData: "<language>en-us</language>",
    items: sortedEntries.map((entry) => ({
      title: entry.data.title,
      description: stripSummaryMarkdown(entry.data.summary),
      pubDate: entry.data.date,
      link: `/devlog/${entry.id}/`,
    })),
  });
}
