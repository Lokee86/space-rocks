import { useEffect, useRef } from "react";
import { createRoot, type Root } from "react-dom/client";

import Archive from "./Archive";
import type { ArchiveContent } from "../content/archiveContent";

type ArchiveClientMountProps = {
  content: Partial<ArchiveContent>;
};

export default function ArchiveClientMount({
  content,
}: ArchiveClientMountProps) {
  const mountRef = useRef<HTMLDivElement | null>(null);
  const rootRef = useRef<Root | null>(null);

  useEffect(() => {
    const mountNode = mountRef.current;
    if (!mountNode || rootRef.current) {
      return;
    }

    const root = createRoot(mountNode);
    rootRef.current = root;
    root.render(<Archive content={content} />);

    return () => {
      rootRef.current?.unmount();
      rootRef.current = null;
    };
  }, []);

  return <div ref={mountRef} />;
}
