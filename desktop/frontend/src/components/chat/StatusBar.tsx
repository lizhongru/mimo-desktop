import { useChatStore } from "../../stores/chatStore";
import { t } from "../../lib/i18n";
import { useSettingsStore } from "../../stores/settingsStore";

interface Props {
  modelName: string;
}

const TOKEN_WARN_THRESHOLD = 96000; // 75% of 128k context window

export function StatusBar({ modelName }: Props) {
  useSettingsStore((s) => s.language);
  const isThinking = useChatStore((s) => s.isThinking);
  const isCompressing = useChatStore((s) => s.isCompressing);
  const usage = useChatStore((s) => s.usage);

  const totalTokens = usage?.totalTokens ?? 0;
  const isOverBudget = totalTokens >= TOKEN_WARN_THRESHOLD;

  let status = "";
  if (isCompressing) status = t("compressing_context");
  else if (isThinking) status = t("status_thinking");

  if (!status) {
    return (
      <div className={`flex items-center justify-between px-6 py-1 border-t border-bdr-sub bg-surface/50 text-xs h-5 flex-shrink-0 ${isOverBudget ? "text-red-400" : "text-txt-g"}`}>
        <span>{isOverBudget ? `\u26A0 ${totalTokens.toLocaleString()} tokens` : ""}</span>
      </div>
    );
  }

  return (
    <div className="flex items-center px-6 py-1 border-t border-bdr-sub bg-surface/50 text-xs text-amber-400 h-5 flex-shrink-0">
      <span className="flex items-center gap-1.5">
        <span className="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse" />
        {status}
      </span>
      {isOverBudget && (
        <span className="ml-auto text-red-400 font-mono">\u26A0 {totalTokens.toLocaleString()} tokens</span>
      )}
    </div>
  );
}
