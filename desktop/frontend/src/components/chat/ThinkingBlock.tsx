import { useState, useEffect } from "react";
import { ChevronDown, ChevronRight, Brain, Sparkles } from "lucide-react";
import clsx from "clsx";
import { t } from "../../lib/i18n";

interface Props {
  content: string;
  isLive?: boolean;
}

export function ThinkingBlock({ content, isLive }: Props) {
  const [expanded, setExpanded] = useState(false);
  const [dots, setDots] = useState("");

  useEffect(() => {
    if (!isLive) {
      setDots("");
      return;
    }
    const timer = setInterval(() => {
      setDots((prev) => (prev.length >= 3 ? "" : prev + "."));
    }, 500);
    return () => clearInterval(timer);
  }, [isLive]);

  if (!content && !isLive) return null;

  const preview = content
    ? content.length > 120
      ? content.slice(0, 120) + "..."
      : content
    : "";

  return (
    <div className="my-1.5">
      <button
        onClick={() => setExpanded(!expanded)}
        className={clsx(
          "group/think flex items-center gap-2 text-xs transition-colors",
          isLive
            ? "text-[var(--color-accent)]"
            : "text-txt-g hover:text-txt-m"
        )}
      >
        {expanded ? (
          <ChevronDown className="w-3 h-3" />
        ) : (
          <ChevronRight className="w-3 h-3" />
        )}

        <span className="relative flex h-4 w-4 items-center justify-center">
          {isLive ? (
            <>
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-[var(--color-accent)] opacity-30" />
              <Sparkles className="h-3.5 w-3.5 animate-pulse text-[var(--color-accent)]" />
            </>
          ) : (
            <Brain className="h-3.5 w-3.5" />
          )}
        </span>

        <span className={clsx("italic", isLive && "font-medium")}>
          {isLive && !content
            ? `${t("thinking_label")}${dots}`
            : expanded
              ? t("thinking_label")
              : preview || t("thinking_label")}
        </span>
      </button>

      {expanded && content && (
        <div className="mt-1 ml-5 pl-3 border-l-2 border-[var(--color-accent)]/30 text-sm text-txt-g italic whitespace-pre-wrap">
          {content}
        </div>
      )}
    </div>
  );
}
