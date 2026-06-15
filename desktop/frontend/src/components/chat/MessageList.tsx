import { useEffect, useRef } from "react";
import { useChatStore } from "../../stores/chatStore";
import { MessageBubble } from "./MessageBubble";
import { ThinkingBlock } from "./ThinkingBlock";
import { ToolCallCard } from "./ToolCallCard";
import { Loader2 } from "lucide-react";
import { t } from "../../lib/i18n";

export function MessageList() {
  const messages = useChatStore((s) => s.messages);
  const isStreaming = useChatStore((s) => s.isStreaming);
  const isThinking = useChatStore((s) => s.isThinking);
  const currentDelta = useChatStore((s) => s.currentDelta);
  const currentThinking = useChatStore((s) => s.currentThinking);
  const currentToolCalls = useChatStore((s) => s.currentToolCalls);

  const bottomRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom on new content
  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const behavior = "smooth";
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        el.scrollTo({ top: el.scrollHeight, behavior });
      });
    });
  }, [messages.length, currentDelta, currentThinking, currentToolCalls.length]);

  return (
    <div
      ref={containerRef}
      className={`flex-1 overflow-y-auto px-6 py-4 ${messages.length > 0 || isStreaming ? "space-y-1" : ""}`}
    >
      {messages.length === 0 && !isStreaming && (
        <div className="flex flex-col items-center justify-center h-full text-[var(--text-ghost)] -mt-4">
          <div className="text-4xl font-bold text-[var(--color-accent)]/60 mb-3">MiMo</div>
          <p className="text-sm">{t("app_subtitle")}</p>
          <p className="text-xs mt-1 text-[var(--text-muted)]">
            {t("empty_hint")}
          </p>
        </div>
      )}

      {messages.map((msg) => (
        <MessageBubble key={msg.id} message={msg} />
      ))}

      {isStreaming && (
        <div className="flex gap-3 py-3">
          <div className="w-7 h-7 rounded-full bg-[var(--sidebar-accent-soft)] flex items-center justify-center flex-shrink-0 mt-0.5">
            <span className="w-4 h-4 text-[var(--color-accent)] text-center text-xs flex items-center justify-center font-semibold">
              M
            </span>
          </div>

          <div className="max-w-[80%] rounded-2xl px-4 py-2.5 border border-[var(--border-default)] bg-[var(--bg-surface)] text-sm text-[var(--text-primary)] shadow-sm">
            {(currentThinking || isThinking) && <ThinkingBlock content={currentThinking} isLive />}

            {currentToolCalls.map((tc) => (
              <ToolCallCard key={tc.id} toolCall={tc} />
            ))}

            {currentDelta && (
              <div className="whitespace-pre-wrap">{currentDelta}</div>
            )}

            {!currentDelta && !currentThinking && currentToolCalls.length === 0 && (
              <ThinkingBlock content="" isLive />
            )}
          </div>
        </div>
      )}

      {(messages.length > 0 || isStreaming) && <div ref={bottomRef} />}
    </div>
  );
}
