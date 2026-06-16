import { useState, type CSSProperties } from "react";
import { ChevronDown, ChevronRight, Brain } from "lucide-react";
import clsx from "clsx";
import { t } from "../../lib/i18n";

interface Props {
  content: string;
  isLive?: boolean;
  label?: string;
}

function DnaHelix() {
  return (
    <span className="dna-helix" aria-hidden="true">
      {Array.from({ length: 8 }).map((_, index) => (
        <span
          key={index}
          className="dna-rung"
          style={{ "--i": index } as CSSProperties}
        >
          <span className="dna-dot dna-dot-left" />
          <span className="dna-bridge" />
          <span className="dna-dot dna-dot-right" />
        </span>
      ))}
    </span>
  );
}

export function ThinkingBlock({ content, isLive, label }: Props) {
  const [expanded, setExpanded] = useState(false);

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

        <span className="relative flex h-5 w-16 items-center justify-center">
          {isLive ? <DnaHelix /> : <Brain className="h-3.5 w-3.5" />}
        </span>

        <span className={clsx("italic", isLive && "not-italic")}>
          {isLive ? label || t("thinking_live_analyzing") : expanded ? t("thinking_label") : preview || t("thinking_label")}
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
