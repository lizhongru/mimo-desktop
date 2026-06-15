import { useState, useEffect, useCallback } from "react";
import { AppLayout } from "./components/layout/AppLayout";
import { useAgent } from "./hooks/useAgent";
import { useChatStore } from "./stores/chatStore";
import { useSessionStore } from "./stores/sessionStore";
import { useSettingsStore } from "./stores/settingsStore";
import { useActivityStore } from "./stores/activityStore";
import { t } from "./lib/i18n";

const DEFAULT_WS = "default";

function workspaceIdFromDir(dir: string): string {
  return dir ? `ws:${dir}` : DEFAULT_WS;
}

type AdvancedSettingsConfig = {
  memory: { ccIndex: boolean; searchScoreFloor: number };
  checkpoint: {
    autoCheckpoint: boolean;
    tokenThreshold: number;
    maxCheckpoints: number;
    reconstructOnResume: boolean;
    contextBudget: number;
  };
  permission: {
    rules: Array<{ permission: string; action: string; pattern?: string }>;
  };
};

declare global {
  interface Window {
    go: {
      desktop: {
        App: {
          SendMessage: (message: string, attachmentsJSON?: string) => Promise<void>;
          CancelOperation: () => Promise<void>;
          IsBusy: () => Promise<boolean>;
          RespondToConfirm: (approved: boolean) => Promise<void>;
          RespondToConfirmAll: (approved: boolean) => Promise<void>;
          GetModelName: () => Promise<string>;
          GetVersion: () => Promise<Record<string, string>>;
          CompressContext: () => Promise<{ before: number; after: number }>;
          ExportChat: (messages: Array<{ role: string; content: string }>) => Promise<void>;
          GetWorkingDir: () => Promise<string>;
          GetTools: () => Promise<Array<{ name: string; description: string; safetyLevel: string; isMcp: boolean; serverName: string }>>;
          GetMCPServers: () => Promise<Array<{ name: string; connected: boolean; toolCount: number; tools: string[] }>>;
          // Workspace methods
          ListWorkspaces: () => Promise<Array<{ id: string; name: string; type: string; path: string }>>;
          CreateWorkspace: (path: string) => Promise<{ id: string; name: string; type: string; path: string }>;
          // Session methods
          ListSessions: (limit: number) => Promise<Array<{ id: string; workspaceId: string; modelName: string; userName: string; lastMessage: string; createdAt: string; updatedAt: string }>>;
          CreateNewSession: (workspaceId: string) => Promise<string>;
          LoadSession: (id: string) => Promise<{ id: string; workspaceId: string; modelName: string; messages: Array<{ role: string; content: string; thinking?: string; toolLines?: string[]; tokens: number; toolCalls: number; durationMs: number }> }>;
          DeleteSession: (id: string) => Promise<void>;
          RenameSession: (id: string, title: string) => Promise<void>;
          MoveSession: (sessionId: string, workspaceId: string) => Promise<void>;
          SaveSessionFromFrontend: (sessionId: string, messages: unknown[]) => Promise<void>;
          // Config methods
          GetConfig: () => Promise<{ defaultModel: string; language: string; theme: string; userName: string; models: Record<string, { provider: string; website: string; apiBase: string; apiKey: string; model: string; models: string[]; fallback: string; maxTokens: number; temperature: number; topP: number; streaming: boolean; vision: boolean; tools: boolean }>; safety: { level: string; permission: string }; agent: { maxIterations: number; planningMode: string; permission: string; reasoningLevel: string; showTokenUsage: boolean }; memory: AdvancedSettingsConfig["memory"]; checkpoint: AdvancedSettingsConfig["checkpoint"]; permission: AdvancedSettingsConfig["permission"] }>;
          SetTheme: (theme: string) => Promise<void>;
          SetLanguage: (lang: string) => Promise<void>;
          SetDefaultModel: (name: string) => Promise<void>;
          UpdateAdvancedSettings: (settings: AdvancedSettingsConfig) => Promise<void>;
          AddModel: (name: string, provider: string, website: string, apiBase: string, apiKey: string, model: string, models: string[], fallback: string, maxTokens: number, temperature: number, topP: number, streaming: boolean, vision: boolean, tools: boolean) => Promise<void>;
          UpdateModel: (name: string, provider: string, website: string, apiBase: string, apiKey: string, model: string, models: string[], fallback: string, maxTokens: number, temperature: number, topP: number, streaming: boolean, vision: boolean, tools: boolean) => Promise<void>;
          RemoveModel: (name: string) => Promise<void>;
          SetSafetyLevel: (level: string) => Promise<void>;
          SetPlanningMode: (mode: string) => Promise<void>;
          SetPermission: (perm: string) => Promise<void>;
          SetReasoningLevel: (level: string) => Promise<void>;
          WindowMinimise: () => Promise<void>;
          WindowMaximise: () => Promise<void>;
          WindowClose: () => Promise<void>;
          WindowIsMaximised: () => Promise<boolean>;
          OpenInExplorer: (path: string) => Promise<void>;
          SelectDirectory: () => Promise<string>;
          // File tree methods
          ListWorkspaceFiles: (path: string, maxDepth: number) => Promise<Array<{ name: string; path: string; isDir: boolean; children?: Array<{ name: string; path: string; isDir: boolean }> }>>;
          ListDirChildren: (dirPath: string) => Promise<Array<{ name: string; path: string; isDir: boolean }>>;
          ListRemoteModels: (modelName: string) => Promise<Array<{ id: string; owned_by: string; description?: string; context_window?: number; max_output?: number; capabilities?: string[] }>>;
          ListRemoteModelsWithConfig: (apiBase: string, apiKey: string) => Promise<Array<{ id: string; owned_by: string; description?: string; context_window?: number; max_output?: number; capabilities?: string[] }>>;
          // Checkpoint methods
          CreateCheckpoint: (summary: string) => Promise<{ success: boolean; message: string; id?: string }>;
          ListCheckpoints: () => Promise<Array<{ id: string; summary: string; token_count: number; message_offset: number; created_at: string }>>;
          RestoreCheckpoint: (checkpointId: string) => Promise<{ success: boolean; message: string; id?: string }>;
          DeleteCheckpoint: (checkpointId: string) => Promise<{ success: boolean; message: string }>;
          ExportCheckpoints: () => Promise<string>;
          // Memory methods
          MemorySearch: (query: string, scope: string, limit: number) => Promise<Array<{ path: string; snippet: string; score: number; scope: string; scope_id: string; type: string }>>;
          MemoryReconcile: () => Promise<[number, number]>;
          MemoryCount: () => Promise<number>;
          MemoryIndexFile: (path: string) => Promise<boolean>;
          WriteMemory: (path: string, content: string) => Promise<boolean>;
          ReadMemory: (path: string) => Promise<string>;
          ListMemoryFiles: () => Promise<Array<{ path: string; name: string; size: number; updatedAt: string; scope: string }>>;
          // Task methods
          TaskCreate: (summary: string, parentID: string) => Promise<{ success: boolean; message: string; task?: { id: string; session_id: string; parent_task_id?: string; status: string; summary: string; owner?: string; created_at: number; last_event_at: number; ended_at?: number } }>;
          TaskList: (status: string, includeTerminal: boolean) => Promise<Array<{ id: string; session_id: string; parent_task_id?: string; status: string; summary: string; owner?: string; created_at: number; last_event_at: number; ended_at?: number }>>;
          TaskStart: (id: string, owner: string, eventSummary: string) => Promise<{ success: boolean; message: string; task?: { id: string; session_id: string; parent_task_id?: string; status: string; summary: string; owner?: string; created_at: number; last_event_at: number; ended_at?: number } }>;
          TaskDone: (id: string, eventSummary: string) => Promise<{ success: boolean; message: string; task?: { id: string; session_id: string; parent_task_id?: string; status: string; summary: string; owner?: string; created_at: number; last_event_at: number; ended_at?: number } }>;
          TaskBlock: (id: string, eventSummary: string) => Promise<{ success: boolean; message: string; task?: { id: string; session_id: string; parent_task_id?: string; status: string; summary: string; owner?: string; created_at: number; last_event_at: number; ended_at?: number } }>;
          TaskDelete: (id: string) => Promise<{ success: boolean; message: string }>;
          TaskGetEvents: (taskID: string) => Promise<Array<{ id: number; task_id: string; at: number; kind: string; summary?: string }>>;
          // Actor methods
          ActorSpawn: (actorType: string, prompt: string, taskID: string) => Promise<{ success: boolean; message: string; actor?: { id: string; type: string; session_id: string; parent_id?: string; status: string; prompt: string; result?: string; error?: string; created_at: number; started_at?: number; completed_at?: number } }>;
          ActorList: (status: string) => Promise<Array<{ id: string; type: string; session_id: string; status: string; prompt: string; result?: string; error?: string; created_at: number; completed_at?: number }>>;
          ActorGet: (id: string) => Promise<{ id: string; type: string; session_id: string; parent_id?: string; status: string; prompt: string; result?: string; error?: string; created_at: number; started_at?: number; completed_at?: number } | null>;
          ActorCancel: (id: string) => Promise<{ success: boolean; message: string }>;
          ActorCleanup: (maxAge: number) => Promise<number>;
          // Multi-agent methods
          AgentListConfigs: () => Promise<Array<{ name: string; mode: string; color: string; description: string; prompt: string; tool_allowlist?: string[] }>>;
          AgentGetCurrent: () => Promise<{ name: string; mode: string; color: string; description: string; prompt: string; tool_allowlist?: string[] } | null>;
          AgentSwitch: (name: string) => Promise<{ success: boolean; message: string; agent?: { name: string; mode: string; color: string; description: string; prompt: string; tool_allowlist?: string[] } }>;
          AgentUpdateConfig: (name: string, config: { name: string; mode: string; color: string; description: string; prompt: string; tool_allowlist?: string[] }) => Promise<{ success: boolean; message: string }>;
          // Dream & Distill methods
          DreamRun: () => Promise<{ success: boolean; message: string; count: number }>;
          DistillRun: () => Promise<{ success: boolean; message: string; count: number }>;
          DistillListCandidates: () => Promise<Array<{ name: string; description: string; confidence: number; pattern?: string; commands?: string[] }>>;
        };
      };
    };
  }
}

export default function App() {
  const currentModel = useSettingsStore((s) => s.currentModel);
  useAgent();

  // Load sessions, workspaces and config on mount
  useEffect(() => {
    Promise.all([
      window.go?.desktop?.App?.ListSessions?.(30),
      window.go?.desktop?.App?.ListWorkspaces?.(),
    ]).then(([sessions, workspaces]) => {
      useSessionStore.getState().setSessions(sessions || []);
      useSessionStore.getState().setWorkspaces(workspaces || []);
    }).catch(console.error);

    window.go?.desktop?.App?.GetConfig?.().then((cfg) => {
      useSettingsStore.getState().initFromConfig({
        theme: cfg.theme, language: cfg.language, defaultModel: cfg.defaultModel,
        models: cfg.models, planningMode: cfg.agent?.planningMode,
        safetyLevel: cfg.safety?.level, permission: cfg.agent?.permission,
        reasoningLevel: cfg.agent?.reasoningLevel,
      });
    }).catch(console.error);
  }, []);

  // Send a message — lazy-creates session if needed
  const handleSend = useCallback(async (message: string, attachments?: { name: string; type: string; dataUrl: string }[]) => {
    const sessStore = useSessionStore.getState();
    let sid = sessStore.currentSessionId;
    const ws = sessStore.selectedWorkspace || DEFAULT_WS;

    if (!sid) {
      console.log("[handleSend] selectedWorkspace:", sessStore.selectedWorkspace, "=> ws:", ws);
      try {
        sid = await window.go?.desktop?.App?.CreateNewSession?.(ws);
        sessStore.setCurrentSessionId(sid);
      } catch (e) {
        console.error("CreateNewSession failed:", e);
        return;
      }
    }

    if (sid && useChatStore.getState().messages.length === 0) {
      await window.go?.desktop?.App?.MoveSession?.(sid, ws);
      sessStore.updateSessionWorkspace(sid, ws);
    }

    if (!sessStore.sessions.find((s) => s.id === sid)) {
      sessStore.addSession({
        id: sid!,
        workspaceId: ws,
        modelName: "",
        userName: "",
        lastMessage: message,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      });
    }

    sessStore.setStreamingSessionId(sid);
    useChatStore.getState().addUserMessage(message);
    const attachmentsJSON = attachments && attachments.length > 0 ? JSON.stringify(attachments) : "";
    window.go?.desktop?.App?.SendMessage?.(message, attachmentsJSON).catch((err) => {
      console.error("SendMessage failed:", err);
      useSessionStore.getState().setStreamingSessionId(null);
      useChatStore.getState().finalizeResponse(`${t("error_prefix")}: ${err}`, 0);
    });
  }, []);

  const handleCancel = useCallback(() => {
    window.go?.desktop?.App?.CancelOperation?.().catch(console.error);
  }, []);

  // New chat - just reset to welcome view; actual session created on first send
  const handleNewChat = useCallback(() => {
    const sessStore = useSessionStore.getState();
    const msgs = useChatStore.getState().messages;
    if (msgs.length === 0 && !sessStore.currentSessionId) {
      return;
    }
    sessStore.setCurrentSessionId(null as unknown as string);
    useChatStore.getState().clearMessages();
    useActivityStore.getState().clear();
  }, []);

  // Load existing session — restore workspace from backend
  const handleLoadSession = useCallback((id: string) => {
    window.go?.desktop?.App?.LoadSession?.(id).then((data) => {
      useSessionStore.getState().setCurrentSessionId(id);
      useChatStore.getState().clearMessages();
      useActivityStore.getState().clear();
      if (data?.messages) {
        for (const msg of data.messages) {
          useChatStore.getState().addRestoredMessage(msg as { role: "user" | "assistant"; content: string; thinking?: string; toolLines?: string[]; tokens?: number; toolCalls?: number; durationMs?: number });
        }
      }
      // Restore workspace selection from backend
      if (data?.workspaceId) {
        useSessionStore.getState().setSelectedWorkspace(data.workspaceId);
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

  // Select workspace — update store; session binding happens at creation time
  const handleSelectWorkspace = useCallback(async (dir: string) => {
    const workspaceId = workspaceIdFromDir(dir);
    const sessStore = useSessionStore.getState();
    const hasMessages = useChatStore.getState().messages.length > 0;

    sessStore.setSelectedWorkspace(workspaceId);
    if (sessStore.currentSessionId && !hasMessages) {
      await window.go?.desktop?.App?.MoveSession?.(sessStore.currentSessionId, workspaceId);
      sessStore.updateSessionWorkspace(sessStore.currentSessionId, workspaceId);
    }

    // dir is a filesystem path; convert to workspace ID
    if (dir) {
      try {
        const ws = await window.go?.desktop?.App?.CreateWorkspace?.(dir);
        const nextStore = useSessionStore.getState();
        nextStore.setSelectedWorkspace(ws.id);
        if (nextStore.currentSessionId && !hasMessages) {
          nextStore.updateSessionWorkspace(nextStore.currentSessionId, ws.id);
        }
        // Refresh workspaces list
        const list = await window.go?.desktop?.App?.ListWorkspaces?.();
        useSessionStore.getState().setWorkspaces(list || []);
      } catch (e) {
        console.error("CreateWorkspace failed:", e);
      }
    } else {
      useSessionStore.getState().setSelectedWorkspace(DEFAULT_WS);
    }
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
