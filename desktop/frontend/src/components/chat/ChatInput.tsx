import { useState, useRef, useCallback, KeyboardEvent } from "react";
import { Send, Square } from "lucide-react";
import { useChatStore } from "../../stores/chatStore";
import { t } from "../../lib/i18n";

interface Props {
  onSend: (message: string) => void;
  onCancel: () => void;
}

export function ChatInput({ onSend, onCancel }: Props) {
  const [text, setText] = useState("");
  const isStreaming = useChatStore((s) => s.isStreaming);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleSend = useCallback(() => {
    const trimmed = text.trim();
    if (!trimmed || isStreaming) return;
    onSend(trimmed);
    setText("");
    // Reset textarea height
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
    }
  }, [text, isStreaming, onSend]);

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  // Auto-resize textarea
  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setText(e.target.value);
    const el = e.target;
    el.style.height = "auto";
    el.style.height = Math.min(el.scrollHeight, 200) + "px";
  };

  return (
    <div className="border-t border-bdr px-6 py-3">
      <div className="flex items-end gap-3 max-w-4xl mx-auto">
        <textarea
          ref={textareaRef}
          value={text}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          placeholder={t("input_placeholder")}
          rows={1}
          className="flex-1 resize-none bg-elevated/60 border border-bdr rounded-lg px-4 py-2.5 text-sm text-txt placeholder:text-txt-m focus:outline-none focus:border-accent/50 focus:ring-1 focus:ring-accent/30 transition-colors"
          readOnly={isStreaming}
        />
        {isStreaming ? (
          <button
            onClick={onCancel}
            className="w-9 h-9 flex items-center justify-center rounded-lg bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-colors flex-shrink-0"
            title={t("cancel_esc")}
          >
            <Square className="w-4 h-4" />
          </button>
        ) : (
          <button
            onClick={handleSend}
            disabled={!text.trim()}
            className="w-9 h-9 flex items-center justify-center rounded-lg bg-accent/20 text-accent hover:bg-accent/30 disabled:opacity-40 disabled:cursor-not-allowed transition-colors flex-shrink-0"
            title={t("send_enter")}
          >
            <Send className="w-4 h-4" />
          </button>
        )}
      </div>
    </div>
  );
}
