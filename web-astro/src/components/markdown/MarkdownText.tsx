import ReactMarkdown from "react-markdown";

import styles from "./MarkdownText.module.css";

type MarkdownTextProps = {
  value: string;
  className?: string;
};

export function MarkdownText({ value, className }: MarkdownTextProps) {
  if (!value.trim()) {
    return null;
  }

  const rootClassName = [styles.root, className].filter(Boolean).join(" ");

  return (
    <div className={rootClassName}>
      <ReactMarkdown>{value}</ReactMarkdown>
    </div>
  );
}
