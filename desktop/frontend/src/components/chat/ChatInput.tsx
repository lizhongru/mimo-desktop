import { useCallback, useRef, useState, type KeyboardEvent } from "react";
import {
  ChevronDown,
  FileText,
  FolderOpen,
  ImageIcon,
  Paperclip,
  Route,
  Send,
  Shield,
  Square,
  X,
} from "lucide-react";
import { useChatStore } from "../../stores/chatStore";
import { useSettingsStore } from "../../stores/settingsStore";
import { t } from "../../lib/i18n";
import { useAnimatedOpen } from "../../lib/useAnimatedOpen";
import { type ChatAttachment, readFilesAsAttachments } from "../../lib/attachments";
import { ModelReasoningPicker } from "./ModelReasoningPicker";

interface Props {
  onSend: (message: string, attachments?: ChatAttachment[]) => void;
  onCancel: () => void;
}

function Dropdown({
  icon: Icon,
  label,
  value,
  options,
  onChange,
}: {
  icon: typeof ChevronDown;
  label: string;
  value: string;
  options: { key: string; label: string }[];
  onChange: (v: string) => void;
}) {
  const [rawOpen, setRawOpen] = useState(false);
  const { shouldRender, closing } = useAnimatedOpen(rawOpen, 120);
  const close = () => setRawOpen(false);

  return (
    <div className="relative">
      <button
        onClick={() => setRawOpen(!rawOpen)}
        className="flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] text-txt-2 hover:text-txt hover:bg-bdr/40 transition-colors cursor-pointer"
        title={label}
      >
        <Icon className="w-3 h-3" />
        <span className="max-w-[72px] truncate">
          {options.find((o) => o.key === value)?.label || value}
        </span>
        <ChevronDown className="w-2.5 h-2.5 text-txt-2" />
      </button>
      {shouldRender && (
        <>
          <div className="fixed inset-0 z-40" onClick={close} />
          <div
            className={`absolute bottom-full mb-1.5 left-0 z-50 bg-surface border border-bdr rounded-lg shadow-xl px-2 py-1 min-w-[120px] ${
              closing ? "animate-pop-out" : "animate-pop-up"
            }`}
          >
            {options.map((opt) => (
              <button
                key={opt.key}
                onClick={() => {
                  onChange(opt.key);
                  close();
                }}
                className={`w-full text-left px-3 py-2 text-xs rounded-md transition-colors cursor-pointer ${
                  opt.key === value
                    ? "text-accent bg-accent/10"
                    : "text-txt-2 hover:bg-elevated"
                }`}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}

function AttachmentIcon({ attachment }: { attachment: ChatAttachment }) {
  if (attachment.type.startsWith("image/")) {
    return <ImageIcon className="w-3 h-3 text-accent" />;
  }
  return <FileText className="w-3 h-3 text-txt-g" />;
}

export function ChatInput({ onSend, onCancel }: Props) {
  const [text, setText] = useState("");
  const [attachments, setAttachments] = useState<ChatAttachment[]>([]);
  const [dragHighlight, setDragHighlight] = useState(false);
  const isStreaming = useChatStore((s) => s.isStreaming);
  const planningMode = useSettingsStore((s) => s.planningMode);
  const permission = useSettingsStore((s) => s.permission);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const folderInputRef = useRef<HTMLInputElement>(null);
  const dragCounter = useRef(0);
  const inputHistory = useRef<string[]>([]);
  const historyIndex = useRef(-1);
  const savedInput = useRef("");

  const handleSend = useCallback(() => {
    const trimmed = text.trim();
    if (!trimmed || isStreaming) return;
    // Push to input history (deduplicate consecutive)
    const hist = inputHistory.current;
    if (hist.length === 0 || hist[hist.length - 1] !== trimmed) {
      hist.push(trimmed);
    }
    historyIndex.current = -1;
    savedInput.current = "";
    onSend(trimmed, attachments.length > 0 ? attachments : undefined);
    setText("");
    setAttachments([]);
    if (textareaRef.current) textareaRef.current.style.height = "auto";
  }, [text, isStreaming, onSend, attachments]);

  const handleFiles = useCallback((files: FileList) => {
    void readFilesAsAttachments(files)
      .then((items) => setAttachments((prev) => [...prev, ...items]))
      .catch((err) => console.error("Failed to read attachments:", err));
  }, []);

  const removeAttachment = (index: number) => {
    setAttachments((prev) => prev.filter((_, i) => i !== index));
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
      return;
    }
    const hist = inputHistory.current;
    if (hist.length === 0) return;
    const ta = textareaRef.current;
    if (!ta) return;
    // Only navigate history when cursor is on first line (Up) or last line (Down)
    // and the selection is collapsed
    if (e.key === "ArrowUp" && !e.shiftKey) {
      const cursorPos = ta.selectionStart;
      const beforeCursor = text.slice(0, cursorPos);
      if (beforeCursor.includes("\n")) return; // not on first line
      e.preventDefault();
      if (historyIndex.current === -1) {
        savedInput.current = text;
        historyIndex.current = hist.length - 1;
      } else if (historyIndex.current > 0) {
        historyIndex.current -= 1;
      }
      setText(hist[historyIndex.current]);
    } else if (e.key === "ArrowDown" && !e.shiftKey) {
      const cursorPos = ta.selectionStart;
      const afterCursor = text.slice(cursorPos);
      if (afterCursor.includes("\n")) return; // not on last line
      e.preventDefault();
      if (historyIndex.current === -1) return;
      if (historyIndex.current < hist.length - 1) {
        historyIndex.current += 1;
        setText(hist[historyIndex.current]);
      } else {
        historyIndex.current = -1;
        setText(savedInput.current);
      }
    }
  };

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    dragCounter.current = 0;
    setDragHighlight(false);
    if (e.dataTransfer.files.length > 0) handleFiles(e.dataTransfer.files);
  }, [handleFiles]);

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = "copy";
  };

  const handleDragEnter = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    dragCounter.current += 1;
    if (dragCounter.current === 1) setDragHighlight(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    dragCounter.current -= 1;
    if (dragCounter.current <= 0) {
      dragCounter.current = 0;
      setDragHighlight(false);
    }
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setText(e.target.value);
    const el = e.target;
    el.style.height = "auto";
    el.style.height = Math.min(el.scrollHeight, 200) + "px";
  };

  return (
    <div className="px-4 py-3">
      <div className="max-w-4xl mx-auto">
        <div
          className={`bg-elevated border rounded-xl focus-within:border-accent/50 focus-within:ring-1 focus-within:ring-accent/20 transition-colors ${
            dragHighlight ? "border-accent ring-2 ring-accent/30 bg-accent/5 drag-drop-zone" : "border-bdr"
          }`}
          onDrop={handleDrop}
          onDragOver={handleDragOver}
          onDragEnter={handleDragEnter}
          onDragLeave={handleDragLeave}
        >
          {dragHighlight && attachments.length === 0 && (
            <div className="flex items-center justify-center gap-2 px-4 py-3 text-accent/80">
              <Paperclip className="w-4 h-4" />
              <span className="text-xs font-medium">{t("drop_to_add")}</span>
            </div>
          )}

          {attachments.length > 0 && (
            <div className="flex flex-wrap gap-2 px-4 pt-2">
              {attachments.map((att, i) => (
                <div key={`${att.name}-${i}`} className="flex items-center gap-1.5 bg-surface border border-bdr rounded-md px-2 py-1 text-xs">
                  <AttachmentIcon attachment={att} />
                  <span className="text-txt-2 max-w-[160px] truncate">{att.name}</span>
                  <button onClick={() => removeAttachment(i)} className="text-txt-g hover:text-red-400 cursor-pointer">
                    <X className="w-3 h-3" />
                  </button>
                </div>
              ))}
            </div>
          )}

          <textarea
            ref={textareaRef}
            value={text}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            placeholder={t("input_placeholder")}
            rows={2}
            className="w-full resize-none bg-transparent px-4 pt-3 pb-1 text-sm text-txt placeholder:text-txt-2 focus:outline-none min-h-[56px]"
            readOnly={isStreaming}
          />

          <div className="flex items-center justify-between px-2 pb-2 pt-0.5">
            <div className="flex items-center gap-0.5">
              <input
                ref={fileInputRef}
                type="file"
                multiple
                className="hidden"
                onChange={(e) => {
                  if (e.target.files) handleFiles(e.target.files);
                  e.target.value = "";
                }}
              />
              <input
                ref={folderInputRef}
                type="file"
                multiple
                className="hidden"
                {...({ webkitdirectory: "", directory: "" } as React.InputHTMLAttributes<HTMLInputElement> & { webkitdirectory: string; directory: string })}
                onChange={(e) => {
                  if (e.target.files) handleFiles(e.target.files);
                  e.target.value = "";
                }}
              />
              <button
                onClick={() => fileInputRef.current?.click()}
                className="flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] text-txt-2 hover:text-txt hover:bg-bdr/40 transition-colors cursor-pointer"
                title={t("attach_file")}
                aria-label={t("attach_file")}
              >
                <Paperclip className="w-3 h-3" />
              </button>
              <button
                onClick={() => folderInputRef.current?.click()}
                className="flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] text-txt-2 hover:text-txt hover:bg-bdr/40 transition-colors cursor-pointer"
                title={t("attach_folder")}
                aria-label={t("attach_folder")}
              >
                <FolderOpen className="w-3 h-3" />
              </button>
              <div className="w-px h-3 bg-txt-2/30 mx-0.5" />
              <Dropdown
                icon={Route}
                label={t("planning_mode")}
                value={planningMode}
                options={[
                  { key: "auto", label: t("plan_auto") },
                  { key: "react", label: t("plan_react") },
                  { key: "plan-execute", label: t("plan_execute") },
                ]}
                onChange={(v) => useSettingsStore.getState().setPlanningMode(v)}
              />
              <div className="w-px h-3 bg-txt-2/30 mx-0.5" />
              <Dropdown
                icon={Shield}
                label={t("safety_level")}
                value={permission}
                options={[
                  { key: "readonly", label: t("perm_readonly") },
                  { key: "write", label: t("perm_write") },
                  { key: "exec", label: t("perm_exec") },
                ]}
                onChange={(v) => useSettingsStore.getState().setPermission(v)}
              />
            </div>
            <div className="flex items-center gap-0.5">
              <ModelReasoningPicker />
              <div className="w-px h-3 bg-txt-2/30 mx-0.5" />
              {isStreaming ? (
                <button
                  onClick={onCancel}
                  className="w-7 h-7 flex items-center justify-center rounded-lg bg-red-500/15 text-red-400 hover:bg-red-500/25 transition-colors cursor-pointer"
                  title={t("cancel_esc")}
                >
                  <Square className="w-3.5 h-3.5" />
                </button>
              ) : (
                <button
                  onClick={handleSend}
                  disabled={!text.trim()}
                  className="w-7 h-7 flex items-center justify-center rounded-lg bg-accent/20 text-accent hover:bg-accent/30 disabled:opacity-30 disabled:cursor-not-allowed transition-colors cursor-pointer"
                  title={t("send_enter")}
                >
                  <Send className="w-3.5 h-3.5" />
                </button>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
