import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { oneDark } from "react-syntax-highlighter/dist/esm/styles/prism";
import { oneLight } from "react-syntax-highlighter/dist/esm/styles/prism";
import { useSettingsStore } from "../../stores/settingsStore";

interface Props {
  content: string;
}

export function MarkdownRenderer({ content }: Props) {
  const theme = useSettingsStore((s) => s.theme);

  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      className={`prose prose-sm max-w-none ${theme === "dark" ? "prose-invert" : ""}`}
      components={{
        code({ className, children, ...props }) {
          const match = /language-(\w+)/.exec(className || "");
          const isBlock = String(children).includes("\n");

          if (match && isBlock) {
            return (
              <SyntaxHighlighter
                style={(theme === "dark" ? oneDark : oneLight) as Record<string, React.CSSProperties>}
                language={match[1]}
                PreTag="div"
                className="rounded-md text-sm my-2"
              >
                {String(children).replace(/\n$/, "")}
              </SyntaxHighlighter>
            );
          }

          return (
            <code
              className="bg-elevated text-amber-400 px-1.5 py-0.5 rounded text-sm"
              {...props}
            >
              {children}
            </code>
          );
        },
        p({ children }) {
          return <p className="mb-2 last:mb-0 leading-relaxed">{children}</p>;
        },
        ul({ children }) {
          return <ul className="list-disc pl-4 mb-2 space-y-1">{children}</ul>;
        },
        ol({ children }) {
          return <ol className="list-decimal pl-4 mb-2 space-y-1">{children}</ol>;
        },
        a({ href, children }) {
          return (
            <a
              href={href}
              target="_blank"
              rel="noopener noreferrer"
              className="text-accent hover:text-accent-light underline"
            >
              {children}
            </a>
          );
        },
        table({ children }) {
          return (
            <div className="overflow-x-auto my-2">
              <table className="border-collapse border border-bdr text-sm">
                {children}
              </table>
            </div>
          );
        },
        th({ children }) {
          return (
            <th className="border border-bdr px-3 py-1.5 bg-elevated text-left font-medium">
              {children}
            </th>
          );
        },
        td({ children }) {
          return (
            <td className="border border-bdr px-3 py-1.5">{children}</td>
          );
        },
      }}
    >
      {content}
    </ReactMarkdown>
  );
}
