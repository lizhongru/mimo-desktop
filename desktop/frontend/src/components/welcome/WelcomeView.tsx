import { useCallback, useEffect, useRef, useState, type KeyboardEvent } from "react";
import {
  Check,
  ChevronDown,
  FileText,
  FolderOpen,
  ImageIcon,
  Paperclip,
  Route,
  Send,
  Shield,
  X,
} from "lucide-react";
import { useChatStore } from "../../stores/chatStore";
import { useSettingsStore } from "../../stores/settingsStore";
import { useSessionStore } from "../../stores/sessionStore";
import { t } from "../../lib/i18n";
import { useAnimatedOpen } from "../../lib/useAnimatedOpen";
import { type ChatAttachment, readFilesAsAttachments } from "../../lib/attachments";
import { ModelReasoningPicker } from "../chat/ModelReasoningPicker";

interface Props {
  onSend: (message: string, attachments?: ChatAttachment[]) => void;
  onSelectWorkspace: (dir: string) => Promise<void>;
}

function WorkspacePicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (dir: string) => void | Promise<void>;
}) {
  const [rawOpen, setRawOpen] = useState(false);
  const { shouldRender, closing } = useAnimatedOpen(rawOpen, 120);
  const [recentDirs, setRecentDirs] = useState<string[]>([]);
  const workspaces = useSessionStore((s) => s.workspaces);

  useEffect(() => {
    const dirs = workspaces
      .filter((ws) => ws.type === "folder" && ws.path)
      .map((ws) => ws.path);
    setRecentDirs(dirs.slice(0, 5));
  }, [workspaces]);

  const handleSelect = async (dir: string) => {
    setRawOpen(false);
    try {
      await onChange(dir);
    } catch (err) {
      console.error("Failed to change workspace:", err);
    }
  };

  const handleBrowse = async () => {
    try {
      const dir = await window.go?.desktop?.App?.SelectDirectory?.();
      if (dir) {
        await handleSelect(dir);
      } else {
        setRawOpen(false);
      }
    } catch (err) {
      console.error("Failed to select directory:", err);
      setRawOpen(false);
    }
  };

  const displayName = value
    ? value.split(/[/\\]/).filter(Boolean).pop() || value
    : t("select_workspace");

  return (
    <div className="relative inline-block">
      <button
        onClick={() => setRawOpen(!rawOpen)}
        className="flex items-center gap-2 px-3 py-2 rounded-lg bg-elevated border border-bdr hover:border-accent/30 transition-colors cursor-pointer text-sm text-txt-2 hover:text-txt"
      >
        <FolderOpen className="w-4 h-4 text-txt-g" />
        <span className="max-w-[200px] truncate">{displayName}</span>
        <ChevronDown className="w-3.5 h-3.5 text-txt-2" />
      </button>

      {shouldRender && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setRawOpen(false)} />
          <div
            className={`absolute bottom-full mb-2 left-0 z-50 bg-surface border border-bdr rounded-lg shadow-xl px-2 py-1 min-w-[220px] ${
              closing ? "animate-pop-out" : "animate-pop-up"
            }`}
          >
            <button
              onClick={() => void handleSelect("")}
              className={`w-full flex items-center gap-2 px-2.5 py-2 text-xs rounded-md transition-colors cursor-pointer ${
                !value ? "text-accent bg-accent/10" : "text-txt-2 hover:bg-elevated"
              }`}
            >
              <span className="w-3.5 h-3.5 flex items-center justify-center text-[11px] font-bold text-txt-g">
                &#x2205;
              </span>
              <span>{t("no_project")}</span>
              {!value && <Check className="w-3 h-3 ml-auto flex-shrink-0" />}
            </button>

            <div className="h-px bg-bdr-div my-1" />
            <div className="px-2.5 py-1 text-[10px] text-txt-m uppercase tracking-wider">
              {t("recent_workspaces")}
            </div>

            {recentDirs.length > 0 ? (
              recentDirs.map((dir) => {
                const name = dir.split(/[/\\]/).filter(Boolean).pop() || dir;
                return (
                  <button
                    key={dir}
                    onClick={() => void handleSelect(dir)}
                    className={`w-full flex items-center gap-2 px-2.5 py-2 text-xs rounded-md transition-colors cursor-pointer ${
                      dir === value ? "text-accent bg-accent/10" : "text-txt-2 hover:bg-elevated"
                    }`}
                  >
                    <FolderOpen className="w-3.5 h-3.5 flex-shrink-0" />
                    <span className="truncate">{name}</span>
                    {dir === value && <Check className="w-3 h-3 ml-auto flex-shrink-0" />}
                  </button>
                );
              })
            ) : (
              <div className="px-2.5 py-2 text-xs text-txt-m">
                {t("no_recent_workspaces")}
              </div>
            )}

            <div className="h-px bg-bdr-div my-1" />
            <button
              onClick={handleBrowse}
              className="w-full flex items-center gap-2 px-2.5 py-2 text-xs text-txt-2 hover:bg-elevated rounded-md transition-colors cursor-pointer"
            >
              <FolderOpen className="w-3.5 h-3.5" />
              {t("browse_folder")}
            </button>
          </div>
        </>
      )}
    </div>
  );
}

function MiniDropdown({
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
          <div className="fixed inset-0 z-40" onClick={() => setRawOpen(false)} />
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
                  setRawOpen(false);
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

export function WelcomeView({ onSend, onSelectWorkspace }: Props) {
  const [text, setText] = useState("");
  const [attachments, setAttachments] = useState<ChatAttachment[]>([]);
  const [workspace, setWorkspace] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const folderInputRef = useRef<HTMLInputElement>(null);
  const isStreaming = useChatStore((s) => s.isStreaming);
  const planningMode = useSettingsStore((s) => s.planningMode);
  const permission = useSettingsStore((s) => s.permission);
  const selectedWorkspace = useSessionStore((s) => s.selectedWorkspace);

  useEffect(() => {
    if (selectedWorkspace.startsWith("ws:")) {
      setWorkspace(selectedWorkspace.slice(3));
    } else {
      setWorkspace("");
    }
  }, [selectedWorkspace]);

  const handleSend = useCallback(() => {
    const trimmed = text.trim();
    if (!trimmed || isStreaming) return;
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
    }
  };

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.dataTransfer.files.length > 0) handleFiles(e.dataTransfer.files);
  }, [handleFiles]);

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setText(e.target.value);
    const el = e.target;
    el.style.height = "auto";
    el.style.height = Math.min(el.scrollHeight, 200) + "px";
  };

  const handleWorkspaceChange = async (dir: string) => {
    setWorkspace(dir);
    await onSelectWorkspace(dir);
  };

  return (
    <div className="flex-1 flex flex-col items-center justify-center px-4 pb-8">
      <div className="mb-8 text-center animate-fade-in">
        <div className="w-20 h-20 mx-auto mb-4 rounded-2xl bg-accent/15 flex items-center justify-center">
          <span className="text-4xl font-bold text-accent">M</span>
        </div>
        <h1 className="text-2xl font-semibold text-txt mb-2">MiMo Desktop</h1>
        <p className="text-sm text-txt-g">{t("welcome_subtitle")}</p>
      </div>

      <div className="w-full max-w-2xl">
        <div
          className="bg-elevated border border-bdr rounded-xl focus-within:border-accent/50 focus-within:ring-1 focus-within:ring-accent/20 transition-colors"
          onDrop={handleDrop}
          onDragOver={handleDragOver}
        >
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
            placeholder={t("welcome_input_placeholder")}
            rows={2}
            className="w-full resize-none bg-transparent px-4 pt-3 pb-1 text-sm text-txt placeholder:text-txt-2 focus:outline-none min-h-[56px]"
            readOnly={isStreaming}
            autoFocus
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
                {...({ webkitdirectory: "", directory: "" } as Record<string, string>)}
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
              <MiniDropdown
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
              <MiniDropdown
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
              <button
                onClick={handleSend}
                disabled={!text.trim()}
                className="w-7 h-7 flex items-center justify-center rounded-lg bg-accent/20 text-accent hover:bg-accent/30 disabled:opacity-30 disabled:cursor-not-allowed transition-colors cursor-pointer"
                title={t("send_enter")}
              >
                <Send className="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>
        <div className="flex justify-center mt-4">
          <WorkspacePicker value={workspace} onChange={handleWorkspaceChange} />
        </div>
        <p className="text-center text-[11px] text-txt-m mt-4">{t("welcome_hint")}</p>
      </div>
    </div>
  );
}
