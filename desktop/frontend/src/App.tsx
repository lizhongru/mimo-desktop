import { useState, useEffect, useCallback } from "react";
import { AppLayout } from "./components/layout/AppLayout";
import { useAgent } from "./hooks/useAgent";
import { useChatStore } from "./stores/chatStore";
import { useSessionStore } from "./stores/sessionStore";
import { useSettingsStore } from "./stores/settingsStore";
import { useActivityStore } from "./stores/activityStore";
import { t } from "./lib/i18n";
import type { ChatMessage } from "./lib/types";

const DEFAULT_WS = "default";

function workspaceIdFromDir(dir: string): string {
  return dir ? `ws:${dir}` : DEFAULT_WS;
}

export default function App() {
  const currentModel = useSettingsStore((s) => s.currentModel);
  const [toast, setToast] = useState<string | null>(null);
  useAgent();

  // Chat store state
  const messages = useChatStore((s) => s.messages);
  const addUserMessage = useChatStore((s) => s.addUserMessage);
  const addRestoredMessage = useChatStore((s) => s.addRestoredMessage);
  const setStreaming = useChatStore((s) => s.setStreaming);
  const setCompressing = useChatStore((s) => s.setCompressing);
  const clearMessages = useChatStore((s) => s.clearMessages);
  const resetStreamState = useChatStore((s) => s.resetStreamState);
  const stashCurrentStreamToBackground = useChatStore((s) => s.stashCurrentStreamToBackground);
  const restoreBackgroundStream = useChatStore((s) => s.restoreBackgroundStream);
  const getSessionSnapshot = useChatStore((s) => s.getSessionSnapshot);
  const replaceMessages = useChatStore((s) => s.replaceMessages);

  // Check if current session has messages (for workspace switching)
  const hasMessages = messages.length > 0;

  // Session store
  const sessStore = useSessionStore();
  const { currentSessionId } = sessStore;

  // Load session list on mount
  useEffect(() => {
    window.go?.desktop?.App?.ListSessions?.(100)
      .then((list) => {
        if (list) sessStore.setSessions(list);
      })
      .catch(console.error);
  }, []);


  const handleSend = useCallback(
    async (message: string, attachments?: string) => {
      if (!message.trim() && !attachments) return;

      // If no session exists yet, create one lazily
      let sessionId = useSessionStore.getState().currentSessionId;
      if (!sessionId) {
        try {
          const activeWs = useSessionStore.getState().selectedWorkspace || DEFAULT_WS;
          sessionId = await window.go?.desktop?.App?.CreateNewSession?.(activeWs);
          if (sessionId) {
            useSessionStore.getState().setCurrentSessionId(sessionId);
            useSessionStore.getState().addSession({
              id: sessionId,
              workspaceId: activeWs,
              modelName: currentModel,
              userName: "",
              firstMessage: message.slice(0, 80),
              lastMessage: message.slice(0, 80),
              createdAt: new Date().toISOString(),
              updatedAt: new Date().toISOString(),
            });
          }
        } catch (e) {
          console.error("Failed to create session:", e);
          return;
        }
      }

      // Reset compressing state on new message
      if (useChatStore.getState().isCompressing) {
        setCompressing(false);
      }

      addUserMessage(message);
      useSessionStore.getState().setStreamingSessionId(useSessionStore.getState().currentSessionId);
      try {
        await window.go?.desktop?.App?.SendMessage?.(message, attachments || "");
      } catch (e) {
        const err = e instanceof Error ? e.message : String(e);
        if (!err.includes("cancelled")) {
          addRestoredMessage({ role: "assistant", content: `Error: ${err}` });
        }
        setStreaming(false);
      }
    },
    [addUserMessage, addRestoredMessage, setStreaming, setCompressing]
  );

  const handleCancel = useCallback(() => {
    window.go?.desktop?.App?.CancelOperation?.().catch(console.error);
    setStreaming(false);
  }, [setStreaming]);

  const handleNewChat = useCallback(() => {
    // Don't create session yet; just clear UI state and go to welcome view
    clearMessages();
    resetStreamState();
    sessStore.setCurrentSessionId(null);
  }, [clearMessages, resetStreamState, sessStore]);

  const handleLoadSession = useCallback(
    async (id: string) => {
      try {
        const data = await window.go?.desktop?.App?.LoadSession?.(id);
        if (!data) return;

        const chatState = useChatStore.getState();
        const switchingAwayFromStreaming = chatState.isStreaming && chatState.activeSessionId !== id;

        if (switchingAwayFromStreaming) {
          stashCurrentStreamToBackground();
          clearMessages();
        } else {
          clearMessages();
          resetStreamState();
        }

        sessStore.setCurrentSessionId(id);

        const restoredBackground = restoreBackgroundStream(id);
        const snapshot = getSessionSnapshot(id);
        if (!restoredBackground && snapshot) {
          replaceMessages(snapshot);
        }
        if (!restoredBackground && !snapshot) {
          const restoredMessages: ChatMessage[] = data.messages.map((m, index) => ({
            id: `restored-${id}-${index}`,
            role: m.role as "user" | "assistant",
            content: m.content,
            thinking: m.thinking || undefined,
            toolCalls: m.toolLines?.map((line, toolIndex) => ({
              id: `restored-${id}-${index}-${toolIndex}`,
              name: line.split("(")[0]?.trim() || "tool",
              args: "",
              result: line,
              status: "done" as const,
            })),
            tokens: m.tokens || undefined,
            duration: m.durationMs || undefined,
            timestamp: Date.now() + index,
          }));
          replaceMessages(restoredMessages);
        }
      } catch (e) {
        console.error("Failed to load session:", e);
      }
    },
    [addRestoredMessage, sessStore, clearMessages, resetStreamState, stashCurrentStreamToBackground, restoreBackgroundStream, getSessionSnapshot, replaceMessages]
  );

  const handleDeleteSession = useCallback(
    async (id: string) => {
      try {
        await window.go?.desktop?.App?.DeleteSession?.(id);
        // If deleted session was current, clear it
        if (sessStore.currentSessionId === id) {
          sessStore.setCurrentSessionId(null);
          clearMessages();
          resetStreamState();
        }
        // Refresh the session list
        const sessions = await window.go?.desktop?.App?.ListSessions?.(100);
        sessStore.setSessions(sessions || []);
      } catch (e) {
        console.error("Failed to delete session:", e);
      }
    },
    [sessStore, clearMessages, resetStreamState]
  );

  const handleSelectWorkspace = useCallback(async (dir: string) => {
    const workspaceId = workspaceIdFromDir(dir);

    // Switch workspace
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
  }, [hasMessages, sessStore]);

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


