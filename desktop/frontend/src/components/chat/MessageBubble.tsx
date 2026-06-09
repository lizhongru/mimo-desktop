import { useState, useRef, useEffect } from "react";
import { User, Bot, Copy, RefreshCw, Trash2 } from "lucide-react";
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
  const [menuPos, setMenuPos] = useState<{ x: number; y: number } | null>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  // Close menu on outside click / scroll
  useEffect(() => {
    if (!menuPos) return;
    const close = () => setMenuPos(null);
    const handleScroll = () => close();
    document.addEventListener("mousedown", close);
    document.addEventListener("scroll", handleScroll, true);
    return () => {
      document.removeEventListener("mousedown", close);
      document.removeEventListener("scroll", handleScroll, true);
    };
  }, [menuPos]);

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setMenuPos({ x: e.clientX, y: e.clientY });
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(message.content).catch(console.error);
    setMenuPos(null);
  };

  const handleRegenerate = () => {
    // Dispatch custom event for App.tsx to handle
    window.dispatchEvent(new CustomEvent("mimo:regenerate"));
    setMenuPos(null);
  };

  const handleDelete = () => {
    useChatStore.getState().deleteMessage(message.id);
    setMenuPos(null);
  };

  // Check if this is the last assistant message (for regenerate option)
  const messages = useChatStore.getState().messages;
  const lastAssistantIdx = [...messages].reverse().findIndex((m) => m.role === "assistant");
  const isLastAssistant = !isUser && lastAssistantIdx === 0;

  return (
    <div
      className={`flex gap-3 py-3 ${isUser ? "justify-end" : ""} relative`}
      onContextMenu={handleContextMenu}
    >
      {!isUser && (
        <div className="w-7 h-7 rounded-full bg-accent/20 flex items-center justify-center flex-shrink-0 mt-0.5">
          <Bot className="w-4 h-4 text-accent" />
        </div>
      )}

      <div
        className={`max-w-[80%] rounded-lg px-4 py-2.5 text-sm ${
          isUser
            ? "bg-accent text-white"
            : "bg-elevated/60 text-txt"
        }`}
      >
        {isUser ? (
          <p className="whitespace-pre-wrap">{message.content}</p>
        ) : (
          <div>
            {/* Completed thinking */}
            {message.thinking && (
              <ThinkingBlock content={message.thinking} />
            )}

            {/* Completed tool calls */}
            {message.toolCalls?.map((tc) => (
              <ToolCallCard key={tc.id} toolCall={tc} />
            ))}

            {/* Completed content */}
            <MarkdownRenderer content={message.content} />
          </div>
        )}

        {/* Duration for assistant messages */}
        {!isUser && message.duration && (
          <div className="mt-1.5 text-[10px] text-txt-m">
            {(message.duration / 1000).toFixed(1)}s
          </div>
        )}
      </div>

      {isUser && (
        <div className="w-7 h-7 rounded-full bg-elevated flex items-center justify-center flex-shrink-0 mt-0.5">
          <User className="w-4 h-4 text-txt-2" />
        </div>
      )}

      {/* Context Menu */}
      {menuPos && (
        <ContextMenu
          ref={menuRef}
          x={menuPos.x}
          y={menuPos.y}
          isLastAssistant={isLastAssistant}
          onCopy={handleCopy}
          onRegenerate={handleRegenerate}
          onDelete={handleDelete}
        />
      )}
    </div>
  );
}

const ContextMenu = ({
  ref,
  x,
  y,
  isLastAssistant,
  onCopy,
  onRegenerate,
  onDelete,
}: {
  ref: React.RefObject<HTMLDivElement | null>;
  x: number;
  y: number;
  isLastAssistant: boolean;
  onCopy: () => void;
  onRegenerate: () => void;
  onDelete: () => void;
}) => {
  const items = [
    { icon: Copy, label: t("copy_text"), onClick: onCopy },
    ...(isLastAssistant
      ? [{ icon: RefreshCw, label: t("regenerate"), onClick: onRegenerate }]
      : []),
    { icon: Trash2, label: t("delete_message"), onClick: onDelete, danger: true },
  ];

  return (
    <div
      ref={ref}
      className="fixed z-[100] bg-surface border border-bdr rounded-lg shadow-xl px-2 py-1 min-w-[160px] animate-pop-up"
      style={{ left: x, top: y }}
      onMouseDown={(e) => e.stopPropagation()}
    >
      {items.map((item, i) => (
        <button
          key={i}
          onClick={item.onClick}
          className={`w-full flex items-center gap-2.5 px-3 py-2 text-sm rounded-md transition-colors cursor-pointer ${
            item.danger
              ? "text-red-400 hover:bg-red-500/10"
              : "text-txt-2 hover:bg-elevated"
          }`}
        >
          <item.icon className="w-3.5 h-3.5" />
          {item.label}
        </button>
      ))}
    </div>
  );
};
