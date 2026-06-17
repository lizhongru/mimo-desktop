import { useEffect, useRef, useState } from "react";
import { User, Bot, Copy, RefreshCw, Check, ChevronDown, ChevronRight, Wrench } from "lucide-react";
import clsx from "clsx";
import type { ChatMessage, ToolCallEvent } from "../../lib/types";
import { useChatStore } from "../../stores/chatStore";
import { t } from "../../lib/i18n";
import { MarkdownRenderer } from "./MarkdownRenderer";
import { ThinkingBlock } from "./ThinkingBlock";
import { ToolCallCard } from "./ToolCallCard";
import { getSkillCommandLabel } from "../../hooks/useSkillCommands";


function ToolCallsBlock({ toolCalls }: { toolCalls?: ToolCallEvent[] }) {
  const [expanded, setExpanded] = useState(false);
  if (!toolCalls || toolCalls.length === 0) return null;
  if (toolCalls.length === 1) return <ToolCallCard toolCall={toolCalls[0]} />;

  const runningCount = toolCalls.filter((toolCall) => toolCall.status === "running").length;
  const counts = toolCalls.reduce<Record<string, number>>((acc, toolCall) => {
    acc[toolCall.name] = (acc[toolCall.name] || 0) + 1;
    return acc;
  }, {});
  const summary = Object.entries(counts)
    .slice(0, 4)
    .map(([name, count]) => `${name} x${count}`)
    .join(" / ");
  const hiddenKinds = Math.max(0, Object.keys(counts).length - 4);

  return (
    <div className="my-1.5 overflow-hidden rounded-lg border border-bdr bg-surface/40">
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="flex w-full cursor-pointer items-center gap-2 px-3 py-2 text-left text-xs transition-colors hover:bg-elevated/50"
        title={expanded ? t("tool_calls_collapse") : t("tool_calls_expand")}
      >
        {expanded ? (
          <ChevronDown className="h-3.5 w-3.5 flex-shrink-0 text-txt-g" />
        ) : (
          <ChevronRight className="h-3.5 w-3.5 flex-shrink-0 text-txt-g" />
        )}
        <span className="relative flex h-4 w-4 flex-shrink-0 items-center justify-center">
          {runningCount > 0 && <span className="absolute inline-flex h-4 w-4 rounded-full bg-amber-400/30 animate-ping" />}
          <Wrench className="relative h-3.5 w-3.5 text-amber-500" />
        </span>
        <span className="font-medium text-txt-2">{t("tool_calls_summary")}</span>
        <span className="rounded-full bg-elevated px-1.5 py-0.5 font-mono text-[10px] text-txt-m">{toolCalls.length}</span>
        <span className="min-w-0 flex-1 truncate font-mono text-[11px] text-txt-g">
          {summary}{hiddenKinds > 0 ? ` / +${hiddenKinds}` : ""}
        </span>
        <span className={runningCount > 0 ? "flex-shrink-0 text-amber-500" : "flex-shrink-0 text-txt-m"}>
          {runningCount > 0 ? `${runningCount} ${t("running")}` : t("done")}
        </span>
      </button>

      {expanded && (
        <div className="border-t border-bdr-sub px-2 py-1.5">
          {toolCalls.map((toolCall) => (
            <ToolCallCard key={toolCall.id} toolCall={toolCall} compact />
          ))}
        </div>
      )}
    </div>
  );
}

interface Props {
  message: ChatMessage;
  skillCommands?: Record<string, string>;
}

export function MessageBubble({ message, skillCommands = {} }: Props) {
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

              <ToolCallsBlock toolCalls={message.toolCalls} />

              <MarkdownRenderer content={message.content} />
            </div>
          )}

          {!isUser && message.duration && (
            <div className="mt-2 text-[10px] text-[var(--text-ghost)]">
              {(message.duration / 1000).toFixed(1)}s
            </div>
          )}
        </div>

        {isUser && message.selectedSkills && message.selectedSkills.length > 0 && (
          <div className="mt-1 flex max-w-full flex-wrap justify-end gap-1 text-[10px] text-[var(--text-muted)]">
            <span className="rounded-full border border-[var(--border-default)] bg-[var(--bg-surface)]/80 px-2 py-0.5">
              {t("message_selected_skills")}
            </span>
            {message.selectedSkills.map((skill) => (
              <span
                key={skill}
                className="max-w-[160px] truncate rounded-full border border-[var(--color-accent)]/25 bg-[var(--color-accent)]/10 px-2 py-0.5 text-[var(--color-accent)]"
                title={skill}
              >
                {getSkillCommandLabel(skill, skillCommands)}
              </span>
            ))}
          </div>
        )}

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
              title={copied ? t("copied") : t("copy_text")}
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
              <span>{copied ? t("copied") : t("copy_text")}</span>
            </button>

            {!isUser && (
              <button
                type="button"
                title={regenerating ? t("regenerating") : t("regenerate")}
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
                  {regenerating ? t("regenerating") : t("regenerate")}
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
