import { useState, useEffect, useCallback } from "react";
import { AppLayout } from "./components/layout/AppLayout";
import { useAgent } from "./hooks/useAgent";
import { useChatStore } from "./stores/chatStore";
import { useSessionStore } from "./stores/sessionStore";
import { useSettingsStore } from "./stores/settingsStore";
import { useActivityStore } from "./stores/activityStore";
import { t } from "./lib/i18n";

// Wails Go bindings — auto-generated at build time
// Package name "desktop" matches the Go package name in desktop/
declare global {
  interface Window {
    go: {
      desktop: {
        App: {
          SendMessage: (message: string) => Promise<void>;
          CancelOperation: () => Promise<void>;
          IsBusy: () => Promise<boolean>;
          RespondToConfirm: (approved: boolean) => Promise<void>;
          RespondToConfirmAll: (approved: boolean) => Promise<void>;
          GetModelName: () => Promise<string>;
          GetVersion: () => Promise<Record<string, string>>;
          CompressContext: () => Promise<{ before: number; after: number }>;
          ExportChat: (messages: Array<{ role: string; content: string }>) => Promise<void>;
          // Project info
          GetWorkingDir: () => Promise<string>;
          // Tools
          GetTools: () => Promise<Array<{
            name: string;
            description: string;
            safetyLevel: string;
            isMcp: boolean;
            serverName: string;
          }>>;
          GetMCPServers: () => Promise<Array<{
            name: string;
            connected: boolean;
            toolCount: number;
            tools: string[];
          }>>;
          // Session methods
          ListSessions: (limit: number) => Promise<Array<{
            id: string;
            modelName: string;
            userName: string;
            lastMessage: string;
            workingDir: string;
            createdAt: string;
            updatedAt: string;
          }>>;
          CreateNewSession: () => Promise<string>;
          LoadSession: (id: string) => Promise<{
            id: string;
            modelName: string;
            messages: Array<{
              role: string;
              content: string;
              thinking?: string;
              toolLines?: string[];
              tokens: number;
              toolCalls: number;
              durationMs: number;
            }>;
          }>;
          DeleteSession: (id: string) => Promise<void>;
          RenameSession: (id: string, title: string) => Promise<void>;
          SaveSessionFromFrontend: (sessionId: string, messages: unknown[]) => Promise<void>;
          // Config methods
          GetConfig: () => Promise<{
            defaultModel: string;
            language: string;
            theme: string;
            userName: string;
            models: Record<string, { 
              provider: string; website: string; apiBase: string; apiKey: string;
              model: string; models: string[]; fallback: string;
              maxTokens: number; temperature: number; topP: number;
              streaming: boolean; vision: boolean; tools: boolean;
            }>;
            safety: { level: string; permission: string };
            agent: { maxIterations: number; planningMode: string; permission: string; reasoningLevel: string; showTokenUsage: boolean };
          }>;
          SetTheme: (theme: string) => Promise<void>;
          SetLanguage: (lang: string) => Promise<void>;
          SetDefaultModel: (name: string) => Promise<void>;
          AddModel: (name: string, provider: string, website: string, apiBase: string, apiKey: string, model: string, models: string[], fallback: string, maxTokens: number, temperature: number, topP: number, streaming: boolean, vision: boolean, tools: boolean) => Promise<void>;
          UpdateModel: (name: string, provider: string, website: string, apiBase: string, apiKey: string, model: string, models: string[], fallback: string, maxTokens: number, temperature: number, topP: number, streaming: boolean, vision: boolean, tools: boolean) => Promise<void>;
          RemoveModel: (name: string) => Promise<void>;
          SetSafetyLevel: (level: string) => Promise<void>;
          SetPlanningMode: (mode: string) => Promise<void>;
          SetPermission: (perm: string) => Promise<void>;
          SetReasoningLevel: (level: string) => Promise<void>;
          // Window controls
          WindowMinimise: () => Promise<void>;
          WindowMaximise: () => Promise<void>;
          WindowClose: () => Promise<void>;
          WindowIsMaximised: () => Promise<boolean>;
          // Explorer
          OpenInExplorer: (path: string) => Promise<void>;
          // Directory picker
          SelectDirectory: () => Promise<string>;
          // Remote model listing
          ListRemoteModels: (modelName: string) => Promise<Array<{id: string; owned_by: string; description?: string; context_window?: number; max_output?: number; capabilities?: string[]}>>;
          ListRemoteModelsWithConfig: (apiBase: string, apiKey: string) => Promise<Array<{id: string; owned_by: string; description?: string; context_window?: number; max_output?: number; capabilities?: string[]}>>;
        };
      };
    };
  }
}

export default function App() {
  const currentModel = useSettingsStore((s) => s.currentModel);
  useAgent();

  useEffect(() => {
    // Load sessions on mount
    window.go?.desktop?.App?.ListSessions?.(30).then((list) => {
      useSessionStore.getState().setSessions(list || []);
    }).catch(console.error);
    // Create initial session and track its ID
    window.go?.desktop?.App?.CreateNewSession?.().then((id) => {
      useSessionStore.getState().setCurrentSessionId(id);
    }).catch(console.error);
    // Initialize settings from config
    window.go?.desktop?.App?.GetConfig?.().then((cfg) => {
      useSettingsStore.getState().initFromConfig({
        theme: cfg.theme,
        language: cfg.language,
        defaultModel: cfg.defaultModel,
        models: cfg.models,
        planningMode: cfg.agent?.planningMode,
        safetyLevel: cfg.safety?.level,
        permission: cfg.agent?.permission,
        reasoningLevel: cfg.agent?.reasoningLevel,
      });
    }).catch(console.error);
  }, []);

  const handleSend = useCallback((message: string) => {
    // If current session is not in the sidebar list, add it immediately
    const sessStore = useSessionStore.getState();
    const sid = sessStore.currentSessionId;
    if (sid && !sessStore.sessions.find((s) => s.id === sid)) {
      sessStore.addSession({
        id: sid,
        modelName: "",
        userName: "",
        lastMessage: message,
        workingDir: "",
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      });
    }
    sessStore.setStreamingSessionId(sid);
    useChatStore.getState().addUserMessage(message);
    window.go?.desktop?.App?.SendMessage?.(message).catch((err) => {
      console.error("SendMessage failed:", err);
      useSessionStore.getState().setStreamingSessionId(null);
      useChatStore.getState().finalizeResponse(`${t("error_prefix")}: ${err}`, 0);
    });
  }, []);

  const handleCancel = useCallback(() => {
    window.go?.desktop?.App?.CancelOperation?.().catch(console.error);
  }, []);

  const handleNewChat = useCallback(() => {
    // Create new session (auto-save already happened in chat:done)
    window.go?.desktop?.App?.CreateNewSession?.().then((id) => {
      useSessionStore.getState().setCurrentSessionId(id);
      useChatStore.getState().clearMessages();
      useActivityStore.getState().clear();
    }).catch(console.error);
  }, []);

  const handleLoadSession = useCallback((id: string) => {
    window.go?.desktop?.App?.LoadSession?.(id).then((data) => {
      useSessionStore.getState().setCurrentSessionId(id);
      useChatStore.getState().clearMessages();
      useActivityStore.getState().clear();
      // Restore messages to chat store
      if (data?.messages) {
        for (const msg of data.messages) {
          useChatStore.getState().addRestoredMessage(msg as { role: "user" | "assistant"; content: string; thinking?: string; toolLines?: string[]; tokens?: number; toolCalls?: number; durationMs?: number });
        }
      }
    }).catch(console.error);
  }, []);

  const handleDeleteSession = useCallback(async (id: string) => {
    await window.go?.desktop?.App?.DeleteSession?.(id);
    useSessionStore.getState().removeSession(id);
    if (useSessionStore.getState().currentSessionId === id) {
      useChatStore.getState().clearMessages();
      useSessionStore.getState().setCurrentSessionId(null);
    }
  }, []);

  const [toast, setToast] = useState<string | null>(null);
  const [selectedWorkspace, setSelectedWorkspace] = useState("");

  const handleSelectWorkspace = useCallback((dir: string) => {
    setSelectedWorkspace(dir);
  }, []);

  const handleExportSession = useCallback(async (id: string) => {
    useSessionStore.getState().setExportingSessionId(id);
    try {
      const data = await window.go?.desktop?.App?.LoadSession?.(id);
      if (!data?.messages?.length) {
        setToast(t("export_empty"));
        return;
      }
      const exportMsgs = data.messages.map((m) => ({ role: m.role, content: m.content }));
      await window.go?.desktop?.App?.ExportChat?.(exportMsgs);
      setToast(t("export_success"));
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      if (!msg.includes("cancelled")) {
        setToast(t("export_failed"));
      }
    } finally {
      useSessionStore.getState().setExportingSessionId(null);
      setTimeout(() => setToast(null), 2500);
    }
  }, []);

  const handleConfirmApprove = useCallback(() => {
    window.go?.desktop?.App?.RespondToConfirm?.(true).catch(console.error);
  }, []);

  const handleConfirmDeny = useCallback(() => {
    window.go?.desktop?.App?.RespondToConfirm?.(false).catch(console.error);
  }, []);

  const handleConfirmApproveAll = useCallback(() => {
    window.go?.desktop?.App?.RespondToConfirmAll?.(true).catch(console.error);
  }, []);

  // Keyboard shortcuts
  useEffect(() => {
    // Regenerate: re-send the last user message
    const handleRegenerate = () => {
      const msgs = useChatStore.getState().messages;
      const lastUser = [...msgs].reverse().find((m) => m.role === "user");
      if (lastUser) {
        const lastUserIdx = msgs.lastIndexOf(lastUser);
        useChatStore.setState({ messages: msgs.slice(0, lastUserIdx) });
        handleSend(lastUser.content);
      }
    };

    const handleKeyDown = (e: KeyboardEvent) => {
      // Escape: cancel or close dialog
      if (e.key === "Escape") {
        const state = useChatStore.getState();
        if (state.confirmAction) {
          state.setConfirmAction(null);
          window.go?.desktop?.App?.RespondToConfirm?.(false).catch(console.error);
        } else if (state.isStreaming) {
          handleCancel();
        }
        return;
      }

      // Ctrl/Cmd shortcuts
      const mod = e.ctrlKey || e.metaKey;
      if (mod) {
        switch (e.key.toLowerCase()) {
          case "b":
            e.preventDefault();
            useSessionStore.getState().toggleLeftSidebar();
            break;
          case "i":
            e.preventDefault();
            useActivityStore.getState().toggleRightSidebar();
            break;
          case "n":
            e.preventDefault();
            handleNewChat();
            break;
          case "k":
            e.preventDefault();
            if (!useChatStore.getState().isCompressing && !useChatStore.getState().isStreaming) {
              window.go?.desktop?.App?.CompressContext?.().catch(console.error);
            }
            break;
        }
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    window.addEventListener("mimo:regenerate", handleRegenerate);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("mimo:regenerate", handleRegenerate);
    };
  }, [handleCancel, handleNewChat, handleSend]);

  return (
    <>
      <AppLayout
        modelName={currentModel}
        onSend={handleSend}
        onCancel={handleCancel}
        onNewChat={handleNewChat}
        onLoadSession={handleLoadSession}
        onDeleteSession={handleDeleteSession}
        onExportSession={handleExportSession}
        onConfirmApprove={handleConfirmApprove}
        onConfirmDeny={handleConfirmDeny}
        onConfirmApproveAll={handleConfirmApproveAll}
        onSelectWorkspace={handleSelectWorkspace}
      />
      {toast && (
        <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-[200] bg-elevated border border-bdr text-txt text-sm px-4 py-2 rounded-lg shadow-lg animate-fade-in">
          {toast}
        </div>
      )}
    </>
  );
}
