import { useState, useEffect, useCallback } from "react";
import { useSessionStore } from "../../stores/sessionStore";
import { useActivityStore } from "../../stores/activityStore";
import { LeftSidebar } from "./LeftSidebar";
import { RightSidebar } from "./RightSidebar";
import { MessageList } from "../chat/MessageList";
import { ChatInput } from "../chat/ChatInput";
import { StatusBar } from "../chat/StatusBar";
import { ConfirmDialog } from "../confirm/ConfirmDialog";
import { ToolsViewer } from "../common/ToolsViewer";
import { SettingsPage } from "../settings/SettingsPage";
import { useChatStore } from "../../stores/chatStore";
import {
  PanelLeft,
  PanelRight,
  Minus,
  Square,
  X,
  Copy,
  Wrench,
} from "lucide-react";
import { t } from "../../lib/i18n";

function formatTokens(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
  return String(n);
}

interface Props {
  modelName: string;
  onSend: (message: string) => void;
  onCancel: () => void;
  onNewChat: () => void;
  onLoadSession: (id: string) => void;
  onDeleteSession: (id: string) => void;
  onConfirmApprove: () => void;
  onConfirmDeny: () => void;
  onConfirmApproveAll: () => void;
}

export function AppLayout({
  modelName,
  onSend,
  onCancel,
  onNewChat,
  onLoadSession,
  onDeleteSession,
  onConfirmApprove,
  onConfirmDeny,
  onConfirmApproveAll,
}: Props) {
  const leftOpen = useSessionStore((s) => s.leftSidebarOpen);
  const rightOpen = useActivityStore((s) => s.rightSidebarOpen);
  const confirmAction = useChatStore((s) => s.confirmAction);
  const usage = useChatStore((s) => s.usage);
  const [toolsOpen, setToolsOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [isMaximised, setIsMaximised] = useState(false);

  useEffect(() => {
    window.go?.desktop?.App?.WindowIsMaximised?.()
      .then(setIsMaximised)
      .catch(() => {});
  }, []);

  const handleMinimise = useCallback(() => {
    window.go?.desktop?.App?.WindowMinimise?.().catch(console.error);
  }, []);

  const handleMaximise = useCallback(() => {
    window.go?.desktop?.App?.WindowMaximise?.().catch(console.error);
    setTimeout(() => {
      window.go?.desktop?.App?.WindowIsMaximised?.()
        .then(setIsMaximised)
        .catch(() => {});
    }, 100);
  }, []);

  const handleClose = useCallback(() => {
    window.go?.desktop?.App?.WindowClose?.().catch(console.error);
  }, []);

  return (
    <div className="h-screen flex flex-col bg-root text-txt select-none">
      {/* Modern Title Bar */}
      <div className="h-10 flex items-center border-b border-bdr-sub/60 bg-root flex-shrink-0 drag-region">
        {/* Left: sidebar toggle */}
        <div className="flex items-center gap-1 pl-3 no-drag">
          <button
            onClick={() => useSessionStore.getState().toggleLeftSidebar()}
            className={`p-1.5 rounded-md hover:bg-elevated/80 transition-colors cursor-pointer no-drag ${
              leftOpen ? "text-txt" : "text-txt-g"
            }`}
            title={t("toggle_left_sidebar")}
          >
            <PanelLeft className="w-[15px] h-[15px]" />
          </button>
        </div>

        {/* Center: App title + model + tokens */}
        <div className="flex-1 flex items-center justify-center gap-2.5 drag-region">
          <span className="text-[13px] font-semibold tracking-wide text-txt">
            MiMo
          </span>
          <span className="text-[10px] px-2 py-[2px] rounded-full bg-elevated/80 text-txt-g font-mono border border-bdr/50">
            {modelName}
          </span>
          {usage && (
            <span className="text-[10px] text-txt-m font-mono">
              {formatTokens(usage.totalTokens)} tokens
            </span>
          )}
        </div>

        {/* Right: sidebar toggle + window controls */}
        <div className="flex items-center no-drag">
          <button
            onClick={() => setToolsOpen(true)}
            className={`p-1.5 rounded-md hover:bg-elevated/80 transition-colors cursor-pointer text-txt-g hover:text-txt`}
            title={t("tools")}
          >
            <Wrench className="w-[15px] h-[15px]" />
          </button>
          <button
            onClick={() => useActivityStore.getState().toggleRightSidebar()}
            className={`p-1.5 rounded-md hover:bg-elevated/80 transition-colors cursor-pointer ${
              rightOpen ? "text-txt" : "text-txt-g"
            }`}
            title={t("toggle_right_sidebar")}
          >
            <PanelRight className="w-[15px] h-[15px]" />
          </button>

          {/* Divider */}
          <div className="w-px h-4 bg-bdr mx-1" />

          {/* Window controls */}
          <button
            onClick={handleMinimise}
            className="w-[46px] h-10 flex items-center justify-center hover:bg-elevated/80 transition-colors cursor-pointer text-txt-g hover:text-txt"
            title={t("minimize")}
          >
            <Minus className="w-[14px] h-[14px]" strokeWidth={1.5} />
          </button>
          <button
            onClick={handleMaximise}
            className="w-[46px] h-10 flex items-center justify-center hover:bg-elevated/80 transition-colors cursor-pointer text-txt-g hover:text-txt"
            title={t("maximize")}
          >
            {isMaximised ? (
              <Copy className="w-[12px] h-[12px]" strokeWidth={1.5} />
            ) : (
              <Square className="w-[12px] h-[12px]" strokeWidth={1.5} />
            )}
          </button>
          <button
            onClick={handleClose}
            className="w-[46px] h-10 flex items-center justify-center hover:bg-red-500/80 transition-colors cursor-pointer text-txt-g hover:text-white rounded-tr-md"
            title={t("close_tooltip")}
          >
            <X className="w-[15px] h-[15px]" strokeWidth={1.5} />
          </button>
        </div>
      </div>

      {/* Main area */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left Sidebar */}
        <div
          className={`border-r border-bdr bg-surface/70 transition-all duration-200 flex-shrink-0 overflow-hidden ${
            leftOpen ? "w-[260px]" : "w-0"
          }`}
        >
          {leftOpen && (
            <LeftSidebar
              onNewChat={onNewChat}
              onLoadSession={onLoadSession}
              onDeleteSession={onDeleteSession}
              onOpenSettings={() => setSettingsOpen(true)}
            />
          )}
        </div>

        {/* Center: Chat */}
        <div className="flex-1 flex flex-col min-w-0">
          <MessageList />
          <ChatInput onSend={onSend} onCancel={onCancel} />
          <StatusBar modelName={modelName} />
        </div>

        {/* Right Sidebar */}
        <div
          className={`border-l border-bdr bg-surface/70 transition-all duration-200 flex-shrink-0 overflow-hidden ${
            rightOpen ? "w-[320px]" : "w-0"
          }`}
        >
          {rightOpen && <RightSidebar />}
        </div>
      </div>

      {/* Settings Page Modal */}
      <SettingsPage
        open={settingsOpen}
        onClose={() => setSettingsOpen(false)}
        defaultModel={modelName}
      />

      {/* Tools Viewer Modal */}
      <ToolsViewer open={toolsOpen} onClose={() => setToolsOpen(false)} />

      {/* Confirm Dialog (global overlay) */}
      {confirmAction && (
        <ConfirmDialog
          action={confirmAction}
          onApprove={() => {
            onConfirmApprove();
            useChatStore.getState().setConfirmAction(null);
          }}
          onDeny={() => {
            onConfirmDeny();
            useChatStore.getState().setConfirmAction(null);
          }}
          onApproveAll={() => {
            onConfirmApproveAll();
            useChatStore.getState().setConfirmAction(null);
          }}
        />
      )}
    </div>
  );
}
