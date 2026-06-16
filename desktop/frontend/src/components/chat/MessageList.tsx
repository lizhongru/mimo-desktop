import { useEffect, useLayoutEffect, useRef } from "react";
import { useChatStore } from "../../stores/chatStore";
import { MessageBubble } from "./MessageBubble";
import { ThinkingBlock } from "./ThinkingBlock";
import { ToolCallCard } from "./ToolCallCard";
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
  const actionButtonRef = useRef<HTMLButtonElement | null>(null);
  const selectionHighlightRef = useRef<HTMLDivElement | null>(null);
  const selectedTextRef = useRef("");
  const previousMessageCount = useRef(0);

  // Jump immediately for loaded history; smooth-scroll only for live streaming deltas.
  useLayoutEffect(() => {
    const el = containerRef.current;
    if (!el) return;

    const isBulkHistoryLoad = messages.length > previousMessageCount.current + 1 && !currentDelta && !currentThinking && currentToolCalls.length === 0;
    previousMessageCount.current = messages.length;

    if (isBulkHistoryLoad) {
      el.scrollTop = el.scrollHeight;
      return;
    }

    requestAnimationFrame(() => {
      el.scrollTo({ top: el.scrollHeight, behavior: "smooth" });
    });
  }, [messages.length, currentDelta, currentThinking, currentToolCalls.length]);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const button = document.createElement("button");
    button.type = "button";
    button.textContent = "添加到对话";
    button.dataset.selectionAction = "true";
    button.className = "fixed z-[300] hidden -translate-x-1/2 rounded-full border border-[var(--border-default)] bg-[var(--bg-surface)]/95 px-3 py-1.5 text-xs text-[var(--text-primary)] shadow-lg backdrop-blur transition-colors hover:bg-[var(--bg-elevated)]";
    button.style.userSelect = "none";
    button.style.webkitUserSelect = "none";
    document.body.appendChild(button);
    actionButtonRef.current = button;

    const highlightLayer = document.createElement("div");
    highlightLayer.dataset.selectionHighlight = "true";
    highlightLayer.className = "pointer-events-none fixed inset-0 z-[250]";
    document.body.appendChild(highlightLayer);
    selectionHighlightRef.current = highlightLayer;

    const clearHighlight = () => {
      highlightLayer.replaceChildren();
    };

    const drawHighlight = (range: Range) => {
      clearHighlight();
      for (const rect of Array.from(range.getClientRects())) {
        if (rect.width < 1 || rect.height < 1) continue;
        const mark = document.createElement("div");
        mark.className = "absolute rounded-[2px] bg-[var(--color-accent)]/30";
        mark.style.left = `${rect.left}px`;
        mark.style.top = `${rect.top}px`;
        mark.style.width = `${rect.width}px`;
        mark.style.height = `${rect.height}px`;
        highlightLayer.appendChild(mark);
      }
    };

    const hideButton = () => {
      selectedTextRef.current = "";
      clearHighlight();
      button.classList.add("hidden");
    };

    const showButton = () => {
      const selection = window.getSelection();
      const selectedText = selection?.toString().trim() || "";
      if (!selection || selection.rangeCount === 0 || !selectedText) {
        hideButton();
        return;
      }

      const anchorNode = selection.anchorNode;
      const focusNode = selection.focusNode;
      if (!anchorNode || !focusNode || !container.contains(anchorNode) || !container.contains(focusNode)) {
        hideButton();
        return;
      }

      const range = selection.getRangeAt(0);
      const rect = range.getBoundingClientRect();
      if (rect.width === 0 && rect.height === 0) {
        hideButton();
        return;
      }

      selectedTextRef.current = selectedText;
      drawHighlight(range);
      button.style.left = `${rect.left + rect.width / 2}px`;
      button.style.top = `${Math.max(8, rect.top - 44)}px`;
      button.classList.remove("hidden");
    };

    const handleMouseUp = () => window.setTimeout(showButton, 20);
    const handleKeyUp = () => window.setTimeout(showButton, 20);
    const handleScroll = () => hideButton();
    const handleDocumentMouseDown = (event: MouseEvent) => {
      const target = event.target as HTMLElement | null;
      if (target === button || target?.closest("[data-selection-action]")) return;
      hideButton();
    };
    const handleButtonMouseDown = (event: MouseEvent) => {
      event.preventDefault();
      event.stopPropagation();
    };
    const handleButtonClick = (event: MouseEvent) => {
      event.preventDefault();
      event.stopPropagation();
      const selectedText = selectedTextRef.current.trim();
      if (!selectedText) return;
      window.dispatchEvent(
        new CustomEvent("mimo:add-selection-to-chat", {
          detail: { text: selectedText },
        })
      );
      window.getSelection()?.removeAllRanges();
      hideButton();
    };

    container.addEventListener("mouseup", handleMouseUp);
    container.addEventListener("keyup", handleKeyUp);
    container.addEventListener("scroll", handleScroll);
    document.addEventListener("mousedown", handleDocumentMouseDown);
    button.addEventListener("mousedown", handleButtonMouseDown);
    button.addEventListener("click", handleButtonClick);

    return () => {
      container.removeEventListener("mouseup", handleMouseUp);
      container.removeEventListener("keyup", handleKeyUp);
      container.removeEventListener("scroll", handleScroll);
      document.removeEventListener("mousedown", handleDocumentMouseDown);
      button.removeEventListener("mousedown", handleButtonMouseDown);
      button.removeEventListener("click", handleButtonClick);
      button.remove();
      highlightLayer.remove();
      actionButtonRef.current = null;
      selectionHighlightRef.current = null;
    };
  }, []);

  return (
    <div
      ref={containerRef}
      className={`chat-message-scroll flex-1 overflow-y-auto px-6 py-4 select-text ${messages.length > 0 || isStreaming ? "space-y-1" : ""}`}
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

          <div className="max-w-[80%] rounded-2xl px-4 py-2.5 border border-[var(--border-default)] bg-[var(--bg-surface)] text-sm text-[var(--text-primary)] shadow-sm select-text">
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
