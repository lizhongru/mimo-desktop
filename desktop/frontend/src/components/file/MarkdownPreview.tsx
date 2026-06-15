import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { oneDark } from "react-syntax-highlighter/dist/esm/styles/prism";
import { Copy, Check } from "lucide-react";
import { useState, useCallback, type ReactNode } from "react";

const codeBlockStyle: Record<string, React.CSSProperties> = {
  ...oneDark,
  'pre[class*="language-"]': {
    ...((oneDark as Record<string, React.CSSProperties>)['pre[class*="language-"]'] || {}),
    background: "#282c34",
    margin: 0,
    padding: "0.75rem 1rem",
    fontSize: "11px",
    lineHeight: "1.55",
    borderRadius: "0 0 6px 6px",
  },
  'code[class*="language-"]': {
    ...((oneDark as Record<string, React.CSSProperties>)['code[class*="language-"]'] || {}),
    background: "transparent",
    fontSize: "11px",
    lineHeight: "1.55",
    fontFamily: '"Cascadia Code", "Fira Code", "JetBrains Mono", Consolas, monospace',
  },
};

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  const handleCopy = useCallback(async () => {
    await navigator.clipboard.writeText(text).catch(() => {});
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }, [text]);
  return (
    <button
      onClick={handleCopy}
      className="flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] text-gray-400 hover:text-gray-200 hover:bg-white/5 transition-colors cursor-pointer"
      title="Copy"
    >
      {copied ? <Check className="w-3 h-3 text-green-400" /> : <Copy className="w-3 h-3" />}
    </button>
  );
}

interface CodeComponentProps {
  className?: string;
  children?: ReactNode;
  [key: string]: unknown;
}

function CodeComponent({ className, children, ...props }: CodeComponentProps) {
  const match = /language-(\w+)/.exec(className || "");
  const codeStr = String(children).replace(/\n$/, "");

  if (match) {
    return (
      <div className="group relative rounded-lg border border-bdr-sub overflow-hidden my-2">
        <div className="flex items-center justify-between px-3 py-1.5 bg-[#21252b] border-b border-bdr-sub">
          <span className="text-[10px] font-medium text-gray-400 uppercase tracking-wider">
            {match[1]}
          </span>
          <CopyButton text={codeStr} />
        </div>
        <SyntaxHighlighter
          language={match[1]}
          style={codeBlockStyle}
          showLineNumbers={false}
          wrapLongLines={false}
          customStyle={{
            background: "#282c34",
            margin: 0,
            padding: "0.75rem 1rem",
          }}
        >
          {codeStr}
        </SyntaxHighlighter>
      </div>
    );
  }

  return (
    <code
      className="text-[11px] bg-elevated/80 px-1.5 py-[2px] rounded font-mono text-rose-300 border border-bdr-sub/50"
      {...props}
    >
      {children}
    </code>
  );
}

interface Props {
  content: string;
}

export function MarkdownPreview({ content }: Props) {
  return (
    <div className="text-txt-2 text-[12px] leading-relaxed">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          code: CodeComponent as any,
          h1: ({ children }) => (
            <h1 className="text-[18px] font-bold text-txt mt-5 mb-3 pb-1.5 border-b border-bdr-sub">
              {children}
            </h1>
          ),
          h2: ({ children }) => (
            <h2 className="text-[15px] font-semibold text-txt mt-4 mb-2 pb-1 border-b border-bdr-sub/50">
              {children}
            </h2>
          ),
          h3: ({ children }) => (
            <h3 className="text-[13px] font-semibold text-txt mt-3 mb-1.5">
              {children}
            </h3>
          ),
          h4: ({ children }) => (
            <h4 className="text-[12px] font-semibold text-txt-2 mt-2 mb-1">
              {children}
            </h4>
          ),
          p: ({ children }) => (
            <p className="text-[12px] leading-relaxed mb-2.5 text-txt-2">
              {children}
            </p>
          ),
          ul: ({ children }) => (
            <ul className="list-disc pl-5 mb-2.5 space-y-0.5 text-[12px] marker:text-accent/60">
              {children}
            </ul>
          ),
          ol: ({ children }) => (
            <ol className="list-decimal pl-5 mb-2.5 space-y-0.5 text-[12px] marker:text-accent/60">
              {children}
            </ol>
          ),
          li: ({ children }) => (
            <li className="text-txt-2">{children}</li>
          ),
          a: ({ href, children }) => (
            <a
              href={href}
              className="text-accent hover:text-accent-light underline underline-offset-2 decoration-accent/40 hover:decoration-accent transition-colors"
              target="_blank"
              rel="noopener noreferrer"
            >
              {children}
            </a>
          ),
          blockquote: ({ children }) => (
            <blockquote className="border-l-[3px] border-accent/60 bg-accent/5 pl-3.5 pr-3 py-2 my-2.5 rounded-r-md text-txt-m italic text-[12px]">
              {children}
            </blockquote>
          ),
          table: ({ children }) => (
            <div className="overflow-x-auto my-2.5 rounded-lg border border-bdr-sub">
              <table className="text-[11px] border-collapse w-full">
                {children}
              </table>
            </div>
          ),
          thead: ({ children }) => (
            <thead className="bg-elevated/80">{children}</thead>
          ),
          th: ({ children }) => (
            <th className="border-b border-bdr px-3 py-1.5 text-left font-semibold text-txt text-[11px]">
              {children}
            </th>
          ),
          td: ({ children }) => (
            <td className="border-b border-bdr-sub px-3 py-1.5 text-txt-2 text-[11px]">
              {children}
            </td>
          ),
          hr: () => <hr className="border-bdr my-4" />,
          img: ({ src, alt }) => (
            <img
              src={src}
              alt={alt}
              className="max-w-full rounded-lg border border-bdr-sub my-2"
            />
          ),
          strong: ({ children }) => (
            <strong className="font-semibold text-txt">{children}</strong>
          ),
          em: ({ children }) => (
            <em className="italic text-txt-m">{children}</em>
          ),
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
