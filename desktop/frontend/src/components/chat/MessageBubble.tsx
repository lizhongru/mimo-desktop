import { useEffect, useRef, useState } from "react";
import { User, Bot, Copy, RefreshCw, Check } from "lucide-react";
import clsx from "clsx";
import type { ChatMessage } from "../../lib/types";
import { useChatStore } from "../../stores/chatStore";
import { t } from "../../lib/i18n";
import { MarkdownRenderer } from "./MarkdownRenderer";
import { ThinkingBlock } from "./ThinkingBlock";
import { ToolCallCard } from "./ToolCallCard";

interface Props {
  message: ChatMessage;
}

export function MessageBubble({ message }: Props) {
  const isUser = message.role === "user";
  const [hovered, setHovered] = useState(false);
  const [copied, setCopied] = useState(false);
  const [regenerating, setRegenerating] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (timer.current) {
        clearTimeout(timer.current);
      }
    };
  }, []);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(message.content);
      setCopied(true);
      if (timer.current) {
        clearTimeout(timer.current);
      }
      timer.current = setTimeout(() => setCopied(false), 1600);
    } catch {
      // ignore clipboard failure
    }
  };

  const hasActiveTextSelection = () => {
    const selection = window.getSelection();
    return !!selection && !selection.isCollapsed && selection.toString().trim().length > 0;
  };

  const handleMouseEnter = () => {
    if (hasActiveTextSelection()) return;
    setHovered(true);
  };

  const handleMouseLeave = () => {
    if (hasActiveTextSelection()) return;
    setHovered(false);
  };

  const handleRegenerate = () => {
    if (regenerating) {
      return;
    }

    setRegenerating(true);
    window.dispatchEvent(new CustomEvent("mimo:regenerate"));
    if (timer.current) {
      clearTimeout(timer.current);
    }
    timer.current = setTimeout(() => setRegenerating(false), 1800);
  };

  return (
    <div
      className={clsx(
        "group/msg relative flex gap-3 py-4",
        isUser && "justify-end"
      )}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      {!isUser && (
        <div className="mt-0.5 flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-[var(--sidebar-accent-soft)]">
          <Bot className="h-4 w-4 text-[var(--color-accent)]" />
        </div>
      )}

      <div
        className={clsx(
          "flex max-w-[80%] flex-col",
          isUser ? "items-end" : "items-start"
        )}
      >
        <div
          className={clsx(
            "rounded-2xl px-4 py-2.5 text-sm leading-relaxed",
            isUser
              ? "bg-[var(--color-accent)] text-white shadow-sm"
              : "border border-[var(--border-default)] bg-[var(--bg-surface)] text-[var(--text-primary)] shadow-sm"
          )}
        >
          {isUser ? (
            <p className="whitespace-pre-wrap">{message.content}</p>
          ) : (
            <div>
              {message.thinking && (
                <ThinkingBlock content={message.thinking} />
              )}

              {message.toolCalls?.map((tc) => (
                <ToolCallCard key={tc.id} toolCall={tc} />
              ))}

              <MarkdownRenderer content={message.content} />
            </div>
          )}

          {!isUser && message.duration && (
            <div className="mt-2 text-[10px] text-[var(--text-ghost)]">
              {(message.duration / 1000).toFixed(1)}s
            </div>
          )}
        </div>

        <div
          className={clsx(
            "pointer-events-none mt-1 flex items-center gap-1 transition-all duration-200",
            hovered ? "translate-y-0 opacity-100" : "translate-y-1 opacity-0",
            isUser ? "justify-end" : "justify-start"
          )}
        >
          <div className="pointer-events-auto flex items-center gap-0.5 rounded-full border border-[var(--border-default)] bg-[var(--bg-surface)]/95 px-1 py-0.5 text-[var(--text-secondary)] shadow-sm backdrop-blur">
            <button
              type="button"
              title={copied ? "Copied" : t("copy_text")}
              onClick={handleCopy}
              className={clsx(
                "inline-flex items-center gap-1 rounded-full px-2 py-1 text-[11px] transition-all duration-200",
                copied
                  ? "text-emerald-500"
                  : "hover:bg-[var(--bg-elevated)]"
              )}
            >
              {copied ? (
                <Check className="h-3.5 w-3.5" />
              ) : (
                <Copy className="h-3.5 w-3.5" />
              )}
              <span>{copied ? "Copied" : t("copy_text")}</span>
            </button>

            {!isUser && (
              <button
                type="button"
                title={regenerating ? "Regenerating..." : t("regenerate")}
                onClick={handleRegenerate}
                disabled={regenerating}
                className={clsx(
                  "inline-flex items-center gap-1 rounded-full px-2 py-1 text-[11px] transition-all duration-200",
                  regenerating
                    ? "text-[var(--color-accent)]"
                    : "hover:bg-[var(--bg-elevated)]"
                )}
              >
                <RefreshCw
                  className={clsx(
                    "h-3.5 w-3.5 transition-transform duration-500",
                    regenerating && "animate-spin"
                  )}
                />
                <span>
                  {regenerating ? "Regenerating..." : t("regenerate")}
                </span>
              </button>
            )}
          </div>
        </div>
      </div>

      {isUser && (
        <div className="mt-0.5 flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-[var(--bg-elevated)]">
          <User className="h-4 w-4 text-[var(--text-muted)]" />
        </div>
      )}
    </div>
  );
}
