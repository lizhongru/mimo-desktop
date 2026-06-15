import { useState } from "react";
import { ChevronDown, ChevronRight, Wrench } from "lucide-react";
import type { ToolCallEvent } from "../../lib/types";
import { t } from "../../lib/i18n";
import { useSettingsStore } from "../../stores/settingsStore";

interface Props {
  toolCall: ToolCallEvent;
}

function parseArgs(argsStr: string): Record<string, string> {
  if (!argsStr) {
    return { raw: "" };
  }

  try {
    const obj = JSON.parse(argsStr);
    if (obj && typeof obj === "object" && !Array.isArray(obj)) {
      const result: Record<string, string> = {};
      for (const [k, v] of Object.entries(obj as Record<string, unknown>)) {
        result[k] = typeof v === "string" ? v : JSON.stringify(v);
      }
      return result;
    }

    return { raw: typeof obj === "string" ? obj : JSON.stringify(obj) };
  } catch {
    return { raw: argsStr };
  }
}

function truncateResult(result: string, maxLines = 3, maxChars = 200): string {
  const lines = result.split("\n");
  if (lines.length > maxLines || result.length > maxChars) {
    const truncated = lines.slice(0, maxLines).join("\n");
    return truncated.length > maxChars
      ? truncated.slice(0, maxChars) + "..."
      : truncated + "\n...";
  }
  return result;
}

export function ToolCallCard({ toolCall }: Props) {
  useSettingsStore((s) => s.language);
  const [expanded, setExpanded] = useState(false);
  const args = parseArgs(toolCall.args);
  const isRunning = toolCall.status === "running";
  const isMcp = toolCall.name.includes("__");

  return (
    <div className="border border-bdr rounded-md my-1.5 overflow-hidden">
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex items-center gap-2 w-full px-3 py-2 text-sm hover:bg-elevated/50 transition-colors"
      >
        {expanded ? (
          <ChevronDown className="w-3.5 h-3.5 text-txt-g flex-shrink-0" />
        ) : (
          <ChevronRight className="w-3.5 h-3.5 text-txt-g flex-shrink-0" />
        )}
        <Wrench className="w-3.5 h-3.5 text-amber-500 flex-shrink-0" />
        <span className="font-mono text-amber-400">{toolCall.name}</span>
        {isMcp && (
          <span className="text-[10px] px-1.5 py-0.5 rounded bg-blue-500/20 text-blue-400">
            {t("mcp_badge")}
          </span>
        )}
        {isRunning && (
          <span className="ml-auto flex items-center gap-1.5 text-xs text-txt-g">
            <span className="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse" />
            {t("running")}
          </span>
        )}
        {!isRunning && toolCall.result && (
          <span className="ml-auto text-xs text-txt-m">{t("done")}</span>
        )}
      </button>

      {expanded && (
        <div className="px-3 pb-2 space-y-2 text-sm border-t border-bdr-sub">
          <div className="mt-2">
            <span className="text-txt-m text-xs uppercase tracking-wide">
              {t("arguments")}
            </span>
            <div className="mt-1 bg-surface rounded p-2 font-mono text-xs overflow-x-auto">
              {Object.entries(args).map(([key, value]) => (
                <div key={key} className="flex gap-2">
                  <span className="text-txt-g">{key}:</span>
                  <span className="text-txt break-all">{value}</span>
                </div>
              ))}
            </div>
          </div>

          {toolCall.result && (
            <div>
              <span className="text-txt-m text-xs uppercase tracking-wide">
                {t("result")}
              </span>
              <pre className="mt-1 bg-surface rounded p-2 text-xs overflow-x-auto whitespace-pre-wrap text-txt-2">
                {expanded
                  ? toolCall.result
                  : truncateResult(toolCall.result)}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

