import * as React from "react";

import PlasmicArchive from "./plasmic/space_rocks_devlog/PlasmicArchive";
import styles from "./Archive.module.css";
import type { ArchiveContent } from "../content/archiveContent";
import { normalizeArchiveContent } from "../content/archiveContent";

export interface ArchiveProps {
  content?: Partial<ArchiveContent>;
}

function Archive_(props: ArchiveProps, _ref: React.ForwardedRef<unknown>) {
  const content = normalizeArchiveContent(props.content ?? {});

  return (
    <PlasmicArchive
      overrides={{
        h1: { children: "All Posts" },
        p: { children: "All development posts for Space Rocks." },
        archiveList: {
          children:
            content.entries.length === 0 ? (
              <p>No devlog entries yet.</p>
            ) : (
              <ul className={styles.archiveList}>
                {content.entries.map((entry) => (
                  <li key={entry.id} className={styles.archiveItem}>
                    <a href={entry.href} className={styles.archiveLink}>
                      {entry.title}
                    </a>
                    <p className={styles.archiveDate}>{entry.date}</p>
                    <p className={styles.archiveSummary}>{entry.summary}</p>
                  </li>
                ))}
              </ul>
            ),
        },
      }}
    />
  );
}

const Archive = React.forwardRef(Archive_);
export default Archive;
