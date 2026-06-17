import { useState } from "react";
import { ChevronDown, ChevronRight, Wrench } from "lucide-react";
import type { ToolCallEvent } from "../../lib/types";
import { t } from "../../lib/i18n";
import { stripAnsi } from "../../lib/commandOutput";
import { useSettingsStore } from "../../stores/settingsStore";

interface Props {
  toolCall: ToolCallEvent;
  compact?: boolean;
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
  const cleanResult = stripAnsi(result);
  const lines = cleanResult.split("\n");
  if (lines.length > maxLines || cleanResult.length > maxChars) {
    const truncated = lines.slice(0, maxLines).join("\n");
    return truncated.length > maxChars
      ? truncated.slice(0, maxChars) + "..."
      : truncated + "\n...";
  }
  return cleanResult;
}

export function ToolCallCard({ toolCall, compact = false }: Props) {
  useSettingsStore((s) => s.language);
  const [expanded, setExpanded] = useState(false);
  const args = parseArgs(toolCall.args);
  const isRunning = toolCall.status === "running";
  const isMcp = toolCall.name.includes("__");

  return (
    <div className={`border border-bdr rounded-md overflow-hidden ${compact ? "my-1" : "my-1.5"}`}>
      <button
        onClick={() => setExpanded(!expanded)}
        className={`flex items-center gap-2 w-full hover:bg-elevated/50 transition-colors cursor-pointer ${compact ? "px-2 py-1.5 text-xs" : "px-3 py-2 text-sm"}`}
      >
        {expanded ? (
          <ChevronDown className="w-3.5 h-3.5 text-txt-g flex-shrink-0" />
        ) : (
          <ChevronRight className="w-3.5 h-3.5 text-txt-g flex-shrink-0" />
        )}
        <span className="relative flex h-4 w-4 flex-shrink-0 items-center justify-center">
          {isRunning && <span className="absolute inline-flex h-4 w-4 rounded-full bg-amber-400/30 animate-ping" />}
          <Wrench className="relative w-3.5 h-3.5 text-amber-500" />
        </span>
        <span className="font-mono text-amber-400 truncate">{toolCall.name}</span>
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
                  ? stripAnsi(toolCall.result)
                  : truncateResult(toolCall.result)}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

