import { useChatStore } from "../../stores/chatStore";
import { t } from "../../lib/i18n";

interface Props {
  modelName: string;
}

export function StatusBar({ modelName }: Props) {
  const isThinking = useChatStore((s) => s.isThinking);
  const toolCalls = useChatStore((s) => s.currentToolCalls);

  const activeTool = toolCalls.find((tc) => tc.status === "running");

  let status = "";
  if (isThinking) status = t("status_thinking");
  else if (activeTool) status = `${t("status_executing_tool")} ${activeTool.name}...`;

  if (!status) {
    return (
      <div className="flex items-center justify-between px-6 py-1 border-t border-bdr-sub bg-surface/50 text-xs text-txt-g h-5 flex-shrink-0">
        <span />
        <span className="font-mono">{modelName}</span>
      </div>
    );
  }

  return (
    <div className="flex items-center px-6 py-1 border-t border-bdr-sub bg-surface/50 text-xs text-amber-400 h-5 flex-shrink-0">
      <span className="flex items-center gap-1.5">
        <span className="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse" />
        {status}
      </span>
    </div>
  );
}
