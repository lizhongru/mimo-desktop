import { useCallback, useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { oneDark } from "react-syntax-highlighter/dist/esm/styles/prism";
import { oneLight } from "react-syntax-highlighter/dist/esm/styles/prism";
import { Check, Copy } from "lucide-react";
import { useSettingsStore } from "../../stores/settingsStore";

interface Props {
  content: string;
}

function CodeBlock({ code, language, theme }: { code: string; language: string; theme: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    await navigator.clipboard.writeText(code).catch(() => {});
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1400);
  }, [code]);

  return (
    <div className="group/code relative my-2">
      <button
        type="button"
        title={copied ? "Copied" : "Copy code"}
        onClick={handleCopy}
        className="mimo-code-copy-button absolute right-2 top-2 z-10 inline-flex h-7 w-7 items-center justify-center rounded-md border border-[var(--code-block-border)] text-[var(--text-muted)] opacity-0 shadow-sm backdrop-blur transition-all hover:text-[var(--text-primary)] group-hover/code:opacity-100"
      >
        {copied ? <Check className="h-3.5 w-3.5 text-emerald-500" /> : <Copy className="h-3.5 w-3.5" />}
      </button>
      <SyntaxHighlighter
        style={(theme === "dark" ? oneDark : oneLight) as Record<string, React.CSSProperties>}
        language={language}
        PreTag="div"
        className="mimo-code-block rounded-lg text-sm"
        customStyle={{
          margin: 0,
          padding: "1rem",
          background: "var(--code-block-surface)",
        }}
        codeTagProps={{
          style: {
            background: "transparent",
            fontFamily: '"Cascadia Code", "Fira Code", "JetBrains Mono", monospace',
          },
        }}
      >
        {code}
      </SyntaxHighlighter>
    </div>
  );
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
              <CodeBlock
                code={String(children).replace(/\n$/, "")}
                language={match[1]}
                theme={theme}
              />
            );
          }

          return (
            <code
              className="mimo-inline-code px-1.5 py-0.5 rounded-md text-sm"
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
