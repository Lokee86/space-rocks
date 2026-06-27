import { useEffect, useRef } from "react";
import { createRoot, type Root } from "react-dom/client";

import Homepage from "./Homepage";
import type { HomepageContent } from "../content/homepageContent";

type HomepageClientMountProps = {
  content?: Partial<HomepageContent>;
};

export default function HomepageClientMount({
  content,
}: HomepageClientMountProps) {
  const mountRef = useRef<HTMLDivElement | null>(null);
  const rootRef = useRef<Root | null>(null);

  useEffect(() => {
    const mountNode = mountRef.current;
    if (!mountNode || rootRef.current) {
      return;
    }

    const root = createRoot(mountNode);
    rootRef.current = root;
    root.render(<Homepage content={content} />);

    return () => {
      rootRef.current?.unmount();
      rootRef.current = null;
    };
  }, []);

  return <div ref={mountRef} />;
}
