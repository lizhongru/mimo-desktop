import { useState, useEffect, useCallback, useRef } from "react";
import { useSessionStore } from "../../stores/sessionStore";
import { useSettingsStore } from "../../stores/settingsStore";
import { useActivityStore } from "../../stores/activityStore";
import { LeftSidebar } from "./LeftSidebar";
import { RightSidebar } from "./RightSidebar";
import { SessionTabs } from "../session/SessionTabs";
import { MessageList } from "../chat/MessageList";
import { ChatInput } from "../chat/ChatInput";
import { StatusBar } from "../chat/StatusBar";
import { ConfirmDialog } from "../confirm/ConfirmDialog";
import { ToolsViewer } from "../common/ToolsViewer";
import { SettingsPage } from "../settings/SettingsPage";
import { MemoryPanelModal } from "../common/MemoryPanelModal";
import { CheckpointPanelModal } from "../common/CheckpointPanelModal";
import { TaskPanelModal } from "../common/TaskPanelModal";
import { ActorPanelModal } from "../common/ActorPanelModal";
import { WelcomeView } from "../welcome/WelcomeView";
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

interface Props {
  modelName: string;
  onSend: (message: string, attachments?: { name: string; type: string; dataUrl: string }[]) => void;
  onCancel: () => void;
  onNewChat: () => void;
  onLoadSession: (id: string) => void;
  onDeleteSession: (id: string) => Promise<void>;
  onExportSession: (id: string) => Promise<void>;
  onConfirmApprove: () => void;
  onConfirmDeny: () => void;
  onConfirmApproveAll: () => void;
  onSelectWorkspace: (dir: string) => Promise<void>;
}

export function AppLayout({
  modelName,
  onSend,
  onCancel,
  onNewChat,
  onLoadSession,
  onDeleteSession,
  onExportSession,
  onConfirmApprove,
  onConfirmDeny,
  onConfirmApproveAll,
  onSelectWorkspace,
}: Props) {
  useSettingsStore((s) => s.language);
  const messages = useChatStore((s) => s.messages);
  const leftOpen = useSessionStore((s) => s.leftSidebarOpen);
  const rightOpen = useActivityStore((s) => s.rightSidebarOpen);
  const confirmAction = useChatStore((s) => s.confirmAction);
  const [toolsOpen, setToolsOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [memoryOpen, setMemoryOpen] = useState(false);
  const [checkpointOpen, setCheckpointOpen] = useState(false);
  const [taskOpen, setTaskOpen] = useState(false);
  const [actorOpen, setActorOpen] = useState(false);
  const [rightWidth, setRightWidth] = useState(320);
  const [isDragging, setIsDragging] = useState(false);

  const [isMaximised, setIsMaximised] = useState(false);
  const [dragActive, setDragActive] = useState(false);
  const dragCounter = useRef(0);

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
    <div
      className={`h-screen flex flex-col bg-root text-txt select-none relative ${dragActive ? "drag-active" : ""}`} onDragEnter={(e) => { e.preventDefault(); dragCounter.current += 1; if (dragCounter.current === 1) setDragActive(true); }} onDragLeave={(e) => { e.preventDefault(); dragCounter.current -= 1; if (dragCounter.current <= 0) { dragCounter.current = 0; setDragActive(false); } }} onDragOver={(e) => e.preventDefault()} onDrop={(e) => { e.preventDefault(); dragCounter.current = 0; setDragActive(false); }}>
      {/* Modern Title Bar */}
      <div className="relative h-10 flex items-center border-b border-bdr-div bg-root flex-shrink-0 drag-region">
        <div className="flex items-center gap-1 pl-3 no-drag z-10">
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

        <div className="absolute inset-0 flex items-center justify-center pointer-events-none drag-region">
          <span className="text-xs font-medium text-txt-g tracking-wide">
            {t("app_name")}
          </span>
        </div>

        <div className="ml-auto flex items-center gap-0.5 pr-1 no-drag z-10">
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

          <div className="w-px h-4 bg-bdr-div mx-1" />

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
          className={`border-r border-bdr bg-sidebar transition-all duration-200 flex-shrink-0 overflow-hidden relative z-10 ${
            leftOpen ? "w-[284px]" : "w-0"
          }`}
        >
          {leftOpen && (
            <LeftSidebar
              onNewChat={onNewChat}
              onLoadSession={onLoadSession}
              onDeleteSession={onDeleteSession}
              onExportSession={onExportSession}
              onOpenSettings={() => setSettingsOpen(true)}
              onOpenMemory={() => setMemoryOpen(true)}
              onOpenCheckpoint={() => setCheckpointOpen(true)}
              onOpenTask={() => setTaskOpen(true)}
              onOpenActor={() => setActorOpen(true)}
            />
          )}
        </div>

        {/* Center: Chat */}
        <div className="flex-1 flex flex-col min-w-0">
          {messages.length === 0 ? (
            <WelcomeView onSend={onSend} onSelectWorkspace={onSelectWorkspace} />
          ) : (
            <>
              <SessionTabs
                onNewChat={onNewChat}
                onLoadSession={onLoadSession}
                onCloseSession={onDeleteSession}
              />
              <MessageList />
              <ChatInput onSend={onSend} onCancel={onCancel} />
              <StatusBar modelName={modelName} />
            </>
          )}
        </div>

        {/* Right Sidebar — dynamic width */}
        <div
          className="border-l border-bdr bg-surface flex-shrink-0 overflow-hidden"
          style={{
            width: rightOpen ? rightWidth : 0,
            transition: isDragging ? "none" : "width 200ms ease",
          }}
        >
          {rightOpen && (
            <RightSidebar
              width={rightWidth}
              onWidthChange={setRightWidth}
              onDragStart={() => setIsDragging(true)}
              onDragEnd={() => setIsDragging(false)}
            />
          )}
        </div>
      </div>

      {/* Modals */}
      <SettingsPage open={settingsOpen} onClose={() => setSettingsOpen(false)} defaultModel={modelName} />
      <ToolsViewer open={toolsOpen} onClose={() => setToolsOpen(false)} />
      <MemoryPanelModal open={memoryOpen} onClose={() => setMemoryOpen(false)} />
      <CheckpointPanelModal open={checkpointOpen} onClose={() => setCheckpointOpen(false)} />
      <TaskPanelModal open={taskOpen} onClose={() => setTaskOpen(false)} />
      <ActorPanelModal open={actorOpen} onClose={() => setActorOpen(false)} />

      <ConfirmDialog
        action={confirmAction}
        onApprove={() => { onConfirmApprove(); useChatStore.getState().setConfirmAction(null); }}
        onDeny={() => { onConfirmDeny(); useChatStore.getState().setConfirmAction(null); }}
        onApproveAll={() => { onConfirmApproveAll(); useChatStore.getState().setConfirmAction(null); }}
      />
    </div>
  );
}
