import { useEffect, useRef } from "react";
import { createRoot, type Root } from "react-dom/client";

import Homepage from "./Homepage";

export default function HomepageClientMount() {
  const mountRef = useRef<HTMLDivElement | null>(null);
  const rootRef = useRef<Root | null>(null);

  useEffect(() => {
    const mountNode = mountRef.current;
    if (!mountNode || rootRef.current) {
      return;
    }

    const root = createRoot(mountNode);
    rootRef.current = root;
    root.render(<Homepage />);

    return () => {
      rootRef.current?.unmount();
      rootRef.current = null;
    };
  }, []);

  return <div ref={mountRef} />;
}
