import { useState, useRef, useCallback, useEffect, KeyboardEvent } from "react";
import {
  Send,
  FolderOpen,
  ChevronDown,
  Route,
  Shield,
  Brain,
  Cpu,
  Check,
  ChevronRight,
  Paperclip,
  ImageIcon,
  X,
  FileText,
} from "lucide-react";
import { useChatStore } from "../../stores/chatStore";
import { useSettingsStore } from "../../stores/settingsStore";
import { useSessionStore } from "../../stores/sessionStore";
import { t } from "../../lib/i18n";
import { useAnimatedOpen } from "../../lib/useAnimatedOpen";

interface Props {
  onSend: (message: string) => void;
  onSelectWorkspace: (dir: string) => void;
}

function WorkspacePicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (dir: string) => void;
}) {
  const [rawOpen, setRawOpen] = useState(false);
  const { shouldRender, closing } = useAnimatedOpen(rawOpen, 120);
  const [recentDirs, setRecentDirs] = useState<string[]>([]);
  const sessions = useSessionStore((s) => s.sessions);

  // Collect unique working dirs from sessions
  useEffect(() => {
    const dirs = new Set<string>();
    for (const s of sessions) {
      if (s.workingDir) dirs.add(s.workingDir);
    }
    setRecentDirs(Array.from(dirs).slice(0, 5));
  }, [sessions]);

  const handleSelect = (dir: string) => {
    onChange(dir);
    setRawOpen(false);
  };

  const handleBrowse = async () => {
    try {
      const dir = await window.go?.desktop?.App?.SelectDirectory?.();
      if (dir) {
        handleSelect(dir);
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
            <div className="px-2.5 py-1 text-[10px] text-txt-m uppercase tracking-wider">
              {t("recent_workspaces")}
            </div>

            {recentDirs.length > 0 ? (
              recentDirs.map((dir) => {
                const name = dir.split(/[/\\]/).filter(Boolean).pop() || dir;
                return (
                  <button
                    key={dir}
                    onClick={() => handleSelect(dir)}
                    className={`w-full flex items-center gap-2 px-2.5 py-2 text-xs rounded-md transition-colors cursor-pointer ${
                      dir === value
                        ? "text-accent bg-accent/10"
                        : "text-txt-2 hover:bg-elevated"
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

function ModelReasoningPicker() {
  const [rawOpen, setRawOpen] = useState(false);
  const { shouldRender: open, closing } = useAnimatedOpen(rawOpen, 150);
  const currentModel = useSettingsStore((s) => s.currentModel);
  const currentModelKey = useSettingsStore((s) => s.currentModelKey);
  const models = useSettingsStore((s) => s.models);
  const reasoningLevel = useSettingsStore((s) => s.reasoningLevel);
  useSettingsStore((s) => s.language);
  const triggerRef = useRef<HTMLButtonElement>(null);

  const reasoningOptions = [
    { key: "low", label: t("reasoning_low"), icon: "\u26a1" },
    { key: "medium", label: t("reasoning_medium"), icon: "\u2696\ufe0f" },
    { key: "high", label: t("reasoning_high"), icon: "\ud83e\udde0" },
  ];
  const reasoningIdx = reasoningOptions.findIndex((o) => o.key === reasoningLevel);
  const [showModels, setShowModels] = useState(false);
  const [panelLeft, setPanelLeft] = useState(false);

  const close = () => { setRawOpen(false); setShowModels(false); };
  const handleOpen = () => {
    if (rawOpen) { close(); return; }
    if (triggerRef.current) {
      const rect = triggerRef.current.getBoundingClientRect();
      setPanelLeft(window.innerWidth - rect.right < 460);
    }
    setRawOpen(true);
  };

  return (
    <div className="relative">
      <button ref={triggerRef} onClick={handleOpen}
        className={`flex items-center gap-1 px-2 py-1 rounded-lg text-[11px] transition-all cursor-pointer ${rawOpen ? "bg-accent/15 text-accent border border-accent/30" : "text-txt-2 hover:text-txt hover:bg-bdr/40 border border-transparent"}`}
        title={t("current_model")}>
        <Cpu className="w-3 h-3" />
        <span className="max-w-[120px] truncate font-medium">{currentModel || "..."}</span>
        <ChevronDown className={`w-2.5 h-2.5 transition-transform ${rawOpen ? "rotate-180" : ""}`} />
      </button>
      {open && <div className="fixed inset-0 z-40" onClick={close} />}
      {open && (
        <div className={`absolute bottom-full mb-2 mt-1 z-50 ${panelLeft ? "right-0" : "left-0"} ${closing ? "animate-pop-out" : "animate-pop-up"}`}>
          <div className="flex items-end gap-0.5">
            {showModels && (
              <div className={`w-[180px] bg-surface border border-bdr rounded-xl shadow-2xl px-2 py-1.5 ${panelLeft ? "order-2" : "order-1"} animate-slide-right`}>
                <div className="px-2.5 py-1 text-[10px] text-txt-m uppercase tracking-wider mb-1">{t("current_model")}</div>
                <div className="space-y-0.5 max-h-[160px] overflow-y-auto">
                  {models.map((m) => (
                    <button key={m}
                      onClick={() => { useSettingsStore.getState().setCurrentModel(m); close(); }}
                      className={`w-full flex items-center justify-between px-2.5 py-2 text-xs rounded-lg transition-all cursor-pointer ${m === currentModelKey ? "bg-accent/10 text-accent" : "text-txt-2 hover:bg-elevated"}`}>
                      <span className="truncate">{m}</span>
                      {m === currentModelKey && <Check className="w-3.5 h-3.5 text-accent flex-shrink-0" />}
                    </button>
                  ))}
                </div>
              </div>
            )}
            <div className={`${showModels ? (panelLeft ? "order-1" : "order-2") : ""} w-[260px] bg-surface border border-bdr rounded-xl shadow-2xl overflow-hidden`}>
              <div className="px-3.5 pt-3 pb-2.5">
                <div className="flex items-center gap-1.5 mb-2.5">
                  <Brain className="w-3.5 h-3.5 text-accent" />
                  <span className="text-[11px] font-medium text-txt">{t("reasoning_label")}</span>
                </div>
                <div className="relative flex bg-elevated rounded-lg p-0.5">
                  <div className="absolute top-0.5 bottom-0.5 rounded-md bg-accent shadow-sm transition-all duration-200 ease-out"
                    style={{
                      width: `calc(${100 / reasoningOptions.length}% - 2px)`,
                      left: `calc(${reasoningIdx * 100 / reasoningOptions.length}% + 2px)`,
                    }} />
                  {reasoningOptions.map((opt) => (
                    <button key={opt.key}
                      onClick={() => useSettingsStore.getState().setReasoningLevel(opt.key)}
                      className={`relative z-10 flex-1 flex items-center justify-center gap-1 px-2 py-1.5 text-[11px] rounded-md transition-colors duration-200 cursor-pointer ${opt.key === reasoningLevel ? "text-white font-medium" : "text-txt-2 hover:text-txt"}`}>
                      <span className="text-[10px]">{opt.icon}</span>
                      <span>{opt.label}</span>
                    </button>
                  ))}
                </div>
              </div>
              <div className="h-px bg-bdr-div mx-3" />
              <button onClick={() => setShowModels(!showModels)}
                className="w-full flex items-center justify-between px-3.5 py-2.5 text-xs text-txt cursor-pointer hover:bg-elevated transition-colors">
                <span className="flex items-center gap-2">
                  <Cpu className="w-3.5 h-3.5 text-accent" />
                  <span className="font-medium truncate">{currentModel || "..."}</span>
                </span>
                <ChevronRight className={`w-3 h-3 text-txt-2 transition-transform duration-200 ${showModels ? "rotate-180" : ""}`} />
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}


export function WelcomeView({ onSend, onSelectWorkspace }: Props) {
  const [text, setText] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const isStreaming = useChatStore((s) => s.isStreaming);
  const planningMode = useSettingsStore((s) => s.planningMode);
  const permission = useSettingsStore((s) => s.permission);
  const [attachments, setAttachments] = useState<{ name: string; type: string; dataUrl: string }[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [workspace, setWorkspace] = useState("");

  const handleSend = useCallback(() => {
    const trimmed = text.trim();
    if (!trimmed || isStreaming) return;
    onSend(trimmed);
    setText("");
    setAttachments([]);
    if (textareaRef.current) textareaRef.current.style.height = "auto";
  }, [text, isStreaming, onSend]);

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); handleSend(); }
  };

  const handleFiles = useCallback((files: FileList) => {
    Array.from(files).forEach((file) => {
      const reader = new FileReader();
      reader.onload = () => {
        setAttachments((prev) => [...prev, { name: file.name, type: file.type, dataUrl: reader.result as string }]);
      };
      reader.readAsDataURL(file);
    });
  }, []);

  const removeAttachment = (i: number) => setAttachments((prev) => prev.filter((_, idx) => idx !== i));
  const handleDrop = useCallback((e: React.DragEvent) => { e.preventDefault(); e.stopPropagation(); if (e.dataTransfer.files.length > 0) handleFiles(e.dataTransfer.files); }, [handleFiles]);
  const handleDragOver = (e: React.DragEvent) => { e.preventDefault(); e.stopPropagation(); };
  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => { setText(e.target.value); const el = e.target; el.style.height = "auto"; el.style.height = Math.min(el.scrollHeight, 200) + "px"; };

  const handleWorkspaceChange = (dir: string) => {
    setWorkspace(dir);
    onSelectWorkspace(dir);
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
        <div className="bg-elevated border border-bdr rounded-xl focus-within:border-accent/50 focus-within:ring-1 focus-within:ring-accent/20 transition-colors" onDrop={handleDrop} onDragOver={handleDragOver}>
          {attachments.length > 0 && (
            <div className="flex flex-wrap gap-2 px-4 pt-2">
              {attachments.map((att, i) => (
                <div key={i} className="flex items-center gap-1.5 bg-surface border border-bdr rounded-md px-2 py-1 text-xs">
                  <Paperclip className="w-3 h-3 text-txt-g" />
                  <span className="text-txt-2 max-w-[120px] truncate">{att.name}</span>
                  <button onClick={() => removeAttachment(i)} className="text-txt-g hover:text-red-400 cursor-pointer"><X className="w-3 h-3" /></button>
                </div>
              ))}
            </div>
          )}
          <textarea ref={textareaRef} value={text} onChange={handleChange} onKeyDown={handleKeyDown} placeholder={t("welcome_input_placeholder")} rows={2} className="w-full resize-none bg-transparent px-4 pt-3 pb-1 text-sm text-txt placeholder:text-txt-2 focus:outline-none min-h-[56px]" readOnly={isStreaming} autoFocus />
          <div className="flex items-center justify-between px-2 pb-2 pt-0.5">
            <div className="flex items-center gap-0.5">
              <input ref={fileInputRef} type="file" multiple accept="image/*,.pdf,.txt,.md,.json,.csv" className="hidden" onChange={(e) => { if (e.target.files) handleFiles(e.target.files); e.target.value = ""; }} />
              <button onClick={() => fileInputRef.current?.click()} className="flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] text-txt-2 hover:text-txt hover:bg-bdr/40 transition-colors cursor-pointer" title={t("attach_file")}><Paperclip className="w-3 h-3" /></button>
              <div className="w-px h-3 bg-txt-2/30 mx-0.5" />
              <MiniDropdown icon={Route} label={t("planning_mode")} value={planningMode} options={[{ key: "auto", label: t("plan_auto") }, { key: "react", label: t("plan_react") }, { key: "plan-execute", label: t("plan_execute") }]} onChange={(v) => useSettingsStore.getState().setPlanningMode(v)} />
              <div className="w-px h-3 bg-txt-2/30 mx-0.5" />
              <MiniDropdown icon={Shield} label={t("safety_level")} value={permission} options={[{ key: "readonly", label: t("perm_readonly") }, { key: "write", label: t("perm_write") }, { key: "exec", label: t("perm_exec") }]} onChange={(v) => useSettingsStore.getState().setPermission(v)} />
            </div>
            <div className="flex items-center gap-0.5">
              <ModelReasoningPicker />
              <div className="w-px h-3 bg-txt-2/30 mx-0.5" />
              <button onClick={handleSend} disabled={!text.trim()} className="w-7 h-7 flex items-center justify-center rounded-lg bg-accent/20 text-accent hover:bg-accent/30 disabled:opacity-30 disabled:cursor-not-allowed transition-colors cursor-pointer" title={t("send_enter")}><Send className="w-3.5 h-3.5" /></button>
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
