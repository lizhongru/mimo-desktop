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
            models: Record<string, { apiBase: string; model: string; maxTokens: number; temperature: number }>;
            safety: { level: string; permission: string };
            agent: { maxIterations: number; planningMode: string; showTokenUsage: boolean };
          }>;
          SetTheme: (theme: string) => Promise<void>;
          SetLanguage: (lang: string) => Promise<void>;
          SetDefaultModel: (name: string) => Promise<void>;
          AddModel: (name: string, apiBase: string, apiKey: string, model: string, maxTokens: number, temperature: number) => Promise<void>;
          RemoveModel: (name: string) => Promise<void>;
          SetSafetyLevel: (level: string) => Promise<void>;
          SetPlanningMode: (mode: string) => Promise<void>;
          // Window controls
          WindowMinimise: () => Promise<void>;
          WindowMaximise: () => Promise<void>;
          WindowClose: () => Promise<void>;
          WindowIsMaximised: () => Promise<boolean>;
          // Explorer
          OpenInExplorer: (path: string) => Promise<void>;
        };
      };
    };
  }
}

export default function App() {
  const [modelName, setModelName] = useState("mimo");
  useAgent();

  useEffect(() => {
    window.go?.desktop?.App?.GetModelName?.().then(setModelName).catch(() => {});
    // Load sessions on mount
    window.go?.desktop?.App?.ListSessions?.(30).then((list) => {
      useSessionStore.getState().setSessions(list || []);
    }).catch(console.error);
    // Create initial session and track its ID
    window.go?.desktop?.App?.CreateNewSession?.().then((id) => {
      useSessionStore.getState().setCurrentSessionId(id);
      (window as Record<string, unknown>).__currentSessionId = id;
    }).catch(console.error);
    // Initialize settings from config
    window.go?.desktop?.App?.GetConfig?.().then((cfg) => {
      useSettingsStore.getState().initFromConfig({
        theme: cfg.theme,
        language: cfg.language,
      });
    }).catch(console.error);
  }, []);

  const handleSend = useCallback((message: string) => {
    useChatStore.getState().addUserMessage(message);
    window.go?.desktop?.App?.SendMessage?.(message).catch((err) => {
      console.error("SendMessage failed:", err);
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
      (window as Record<string, unknown>).__currentSessionId = id;
      useChatStore.getState().clearMessages();
      useActivityStore.getState().clear();
    }).catch(console.error);
  }, []);

  const handleLoadSession = useCallback((id: string) => {
    window.go?.desktop?.App?.LoadSession?.(id).then((data) => {
      useSessionStore.getState().setCurrentSessionId(id);
      (window as Record<string, unknown>).__currentSessionId = id;
      useChatStore.getState().clearMessages();
      useActivityStore.getState().clear();
      // Restore messages to chat store
      if (data?.messages) {
        for (const msg of data.messages) {
          useChatStore.getState().addRestoredMessage(msg);
        }
      }
    }).catch(console.error);
  }, []);

  const handleDeleteSession = useCallback((id: string) => {
    window.go?.desktop?.App?.DeleteSession?.(id).then(() => {
      useSessionStore.getState().removeSession(id);
      if (useSessionStore.getState().currentSessionId === id) {
        useChatStore.getState().clearMessages();
        useSessionStore.getState().setCurrentSessionId(null);
      }
    }).catch(console.error);
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
        }
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleCancel, handleNewChat]);

  return (
    <AppLayout
      modelName={modelName}
      onSend={handleSend}
      onCancel={handleCancel}
      onNewChat={handleNewChat}
      onLoadSession={handleLoadSession}
      onDeleteSession={handleDeleteSession}
      onConfirmApprove={handleConfirmApprove}
      onConfirmDeny={handleConfirmDeny}
      onConfirmApproveAll={handleConfirmApproveAll}
    />
  );
}
