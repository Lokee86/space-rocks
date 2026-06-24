import Head from "next/head";

export default function HomePage() {
  return (
    <>
      <Head>
        <title>Space Rocks Website Host</title>
      </Head>
      <main
        style={{
          minHeight: "100vh",
          display: "grid",
          placeItems: "center",
          background: "#050814",
          color: "#e5eefc",
          fontFamily:
            'Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
          padding: "2rem",
        }}
      >
        <section
          style={{
            maxWidth: "42rem",
            textAlign: "center",
            padding: "3rem 2rem",
            borderRadius: "1.5rem",
            border: "1px solid rgba(148, 163, 184, 0.2)",
            background: "linear-gradient(180deg, rgba(15, 23, 42, 0.96), rgba(2, 6, 23, 0.98))",
            boxShadow: "0 24px 80px rgba(0, 0, 0, 0.45)",
          }}
        >
          <p
            style={{
              margin: 0,
              letterSpacing: "0.18em",
              textTransform: "uppercase",
              color: "#7dd3fc",
              fontSize: "0.8rem",
            }}
          >
            Space Rocks
          </p>
          <h1
            style={{
              margin: "1rem 0 0.75rem",
              fontSize: "clamp(2.5rem, 7vw, 4.5rem)",
              lineHeight: 1.05,
            }}
          >
            Space Rocks Website Host
          </h1>
          <p
            style={{
              margin: 0,
              fontSize: "1.1rem",
              color: "#cbd5e1",
            }}
          >
            React/Next app host is running.
          </p>
        </section>
      </main>
    </>
  );
}
