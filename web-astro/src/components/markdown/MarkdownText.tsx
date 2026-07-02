import * as React from "react";
import ReactMarkdown from "react-markdown";

import styles from "./MarkdownText.module.css";

type MarkdownTextProps = {
  value: string;
  className?: string;
};

function slugifyHeading(value: string) {
  return value
    .toLowerCase()
    .trim()
    .replace(/['’]/g, "")
    .replace(/&/g, "and")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

function extractPlainText(children: React.ReactNode): string {
  if (typeof children === "string" || typeof children === "number") {
    return String(children);
  }

  if (Array.isArray(children)) {
    return children.map(extractPlainText).join("");
  }

  if (React.isValidElement(children)) {
    return extractPlainText(children.props.children);
  }

  return "";
}

export function MarkdownText({ value, className }: MarkdownTextProps) {
  if (!value.trim()) {
    return null;
  }

  const rootClassName = [styles.root, className].filter(Boolean).join(" ");

  return (
    <div className={rootClassName}>
      <ReactMarkdown
        components={{
          h2: ({ children, ...props }) => (
            <h2 id={slugifyHeading(extractPlainText(children))} {...props}>
              {children}
            </h2>
          ),
          h3: ({ children, ...props }) => (
            <h3 id={slugifyHeading(extractPlainText(children))} {...props}>
              {children}
            </h3>
          ),
          h4: ({ children, ...props }) => (
            <h4 id={slugifyHeading(extractPlainText(children))} {...props}>
              {children}
            </h4>
          ),
        }}
      >
        {value}
      </ReactMarkdown>
    </div>
  );
}
