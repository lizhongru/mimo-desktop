import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { oneDark } from "react-syntax-highlighter/dist/esm/styles/prism";

interface Props {
  content: string;
  language: string;
  wrapLines?: boolean;
}

const customStyle: Record<string, React.CSSProperties> = {
  ...oneDark,
  'pre[class*="language-"]': {
    ...((oneDark as Record<string, React.CSSProperties>)['pre[class*="language-"]'] || {}),
    background: "transparent",
    margin: 0,
    padding: "0.625rem 0",
    fontSize: "12px",
    lineHeight: "1.7",
  },
  'code[class*="language-"]': {
    ...((oneDark as Record<string, React.CSSProperties>)['code[class*="language-"]'] || {}),
    background: "transparent",
    fontSize: "12px",
    lineHeight: "1.7",
    fontFamily: '"Cascadia Code", "Fira Code", "JetBrains Mono", Consolas, monospace',
  },
};

export function CodeBlock({ content, language, wrapLines = false }: Props) {
  const lang = language === "text" ? "plaintext" : language;
  const lines = content.split("\n");

  return (
    <div className="relative rounded-lg border border-bdr-sub bg-[#282c34] overflow-hidden">
      {/* Language badge */}
      <div className="flex items-center justify-between px-3 py-1.5 bg-[#21252b] border-b border-bdr-sub">
        <div className="flex items-center gap-2">
          <span className="text-[10px] font-semibold text-txt-g uppercase tracking-wider">
            {lang}
          </span>
          <span className="text-[10px] text-txt-m/40">
            {lines.length} {lines.length === 1 ? "line" : "lines"}
          </span>
        </div>
      </div>

      {/* Code content */}
      <div className={wrapLines ? "overflow-x-hidden" : "overflow-x-auto"}>
        <div className="flex">
          {/* Line numbers */}
          <div className="flex-shrink-0 select-none py-[0.625rem] pl-3 pr-2 text-right bg-[#21252b]/40 border-r border-bdr-sub/20">
            {lines.map((_, idx) => (
              <div
                key={idx}
                className="text-[11px] leading-[20.4px] text-txt-m/30 font-mono"
              >
                {idx + 1}
              </div>
            ))}
          </div>

          {/* Highlighted code */}
          <div className="flex-1 min-w-0 pl-2">
            <SyntaxHighlighter
              language={lang}
              style={customStyle}
              showLineNumbers={false}
              wrapLines={wrapLines}
              wrapLongLines={wrapLines}
              customStyle={{
                background: "transparent",
                margin: 0,
                padding: "0.625rem 0.75rem",
              }}
            >
              {content}
            </SyntaxHighlighter>
          </div>
        </div>
      </div>
    </div>
  );
}
