import type { AppProps } from "next/app";

import "../components/plasmic/space_rocks_devlog/plasmic.css"; // plasmic-import: uNJepqX5kmDcUn9dDb3UVD/projectcss

export default function App({ Component, pageProps }: AppProps) {
  return <Component {...pageProps} />;
}
