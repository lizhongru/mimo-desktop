import { useState } from "react";
import { ChevronDown, ChevronRight, Brain } from "lucide-react";
import { t } from "../../lib/i18n";

interface Props {
  content: string;
}

export function ThinkingBlock({ content }: Props) {
  const [expanded, setExpanded] = useState(false);

  if (!content) return null;

  const preview =
    content.length > 120 ? content.slice(0, 120) + "..." : content;

  return (
    <div className="my-1.5">
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex items-center gap-2 text-xs text-txt-g hover:text-txt-m transition-colors"
      >
        {expanded ? (
          <ChevronDown className="w-3 h-3" />
        ) : (
          <ChevronRight className="w-3 h-3" />
        )}
        <Brain className="w-3 h-3" />
        <span className="italic">
          {expanded ? t("thinking_label") : preview}
        </span>
      </button>

      {expanded && (
        <div className="mt-1 ml-5 pl-3 border-l-2 border-bdr text-sm text-txt-g italic whitespace-pre-wrap">
          {content}
        </div>
      )}
    </div>
  );
}
