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

export default function App() {
  const currentModel = useSettingsStore((s) => s.currentModel);
  const [toast, setToast] = useState<string | null>(null);
  useAgent();

  // Chat store state
  const messages = useChatStore((s) => s.messages);
  const addUserMessage = useChatStore((s) => s.addUserMessage);
  const addRestoredMessage = useChatStore((s) => s.addRestoredMessage);
  const appendDelta = useChatStore((s) => s.appendDelta);
  const appendThinking = useChatStore((s) => s.appendThinking);
  const addToolCall = useChatStore((s) => s.addToolCall);
  const updateToolResult = useChatStore((s) => s.updateToolResult);
  const finalizeResponse = useChatStore((s) => s.finalizeResponse);
  const setStreaming = useChatStore((s) => s.setStreaming);
  const setConfirmAction = useChatStore((s) => s.setConfirmAction);
  const setCompressing = useChatStore((s) => s.setCompressing);
  const setUsage = useChatStore((s) => s.setUsage);
  const clearMessages = useChatStore((s) => s.clearMessages);
  const resetStreamState = useChatStore((s) => s.resetStreamState);

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

  useEffect(() => {
    const handleDelta = (_: unknown, text: string) => {
      appendDelta(text);
    };
    const handleThinking = (_: unknown, delta: string) => {
      appendThinking(delta);
    };
    const handleToolCall = (_: unknown, name: string, args: string) => {
      addToolCall(name, args);
    };
    const handleToolResult = (_: unknown, name: string, result: string) => {
      updateToolResult(name, result);
    };
    const handleError = (_: unknown, err: string) => {
      addRestoredMessage({ role: "assistant", content: `Error: ${err}` });
      setStreaming(false);
    };
    const handleDone = () => {
      // finalizeResponse is handled by useAgent.ts; here we only save + update sidebar
      const sid = useSessionStore.getState().currentSessionId;
      if (sid) {
        // Wait a tick for finalizeResponse in useAgent to finish
        setTimeout(() => {
          const msgs = useChatStore.getState().messages.map((m) => ({
            role: m.role, content: m.content, thinking: m.thinking || "",
            toolLines: (m.toolCalls || []).map((tc) => tc.name + "(" + tc.args + ")"),
            tokens: m.tokens || 0, toolCalls: m.toolCalls?.length || 0, durationMs: m.duration || 0,
          }));
          window.go?.desktop?.App?.SaveSessionFromFrontend?.(sid, msgs).then(() => {
            const lastUser = [...useChatStore.getState().messages].reverse().find((m) => m.role === "user");
            useSessionStore.getState().updateSession(sid, lastUser?.content || "");
          }).catch(console.error);
        }, 50);
      }
    };
    const handleUsage = (_: unknown, usage: { promptTokens: number; completionTokens: number; totalTokens: number }) => {
      setUsage(usage);
    };
    const handleCompressing = () => {
      setCompressing(true);
    };
    const handleCompressed = (_: unknown, result: { before: number; after: number }) => {
      setCompressing(false);
      if (result) {
        addRestoredMessage({
          role: "assistant",
          content: `Context compressed: ${result.before} → ${result.after} tokens`,
        });
      }
    };
    const handleSafetyConfirm = (_: unknown, action: { level: string; description: string; tool: string; params: Record<string, unknown> }) => {
      setConfirmAction(action);
    };
    const handlePlanning = (_: unknown, message: string) => {
      useActivityStore.getState().addEntry({
        type: "plan_step",
        name: "Planning",
        status: "running",
        detail: message,
      });
    };
    const handlePlanGenerated = (_: unknown, plan: { goal: string; steps: Array<{ id: string; description: string; status: string }> }) => {
      useActivityStore.getState().setPlan({
        goal: plan.goal,
        currentStep: 0,
        totalSteps: plan.steps.length,
        steps: plan.steps.map((s) => ({
          id: Number(s.id) || 0,
          description: s.description,
          status: (s.status as "pending" | "in_progress" | "completed" | "failed" | "skipped") || "pending",
        })),
      });
    };
    const handlePlanStepStart = (_: unknown, step: { id: string; description: string; status: string }) => {
      useActivityStore.getState().updatePlanStep(Number(step.id) || 0, "in_progress");
    };
    const handlePlanStepDone = (_: unknown, step: { id: string; description: string; status: string }) => {
      useActivityStore.getState().updatePlanStep(
        Number(step.id) || 0,
        (step.status as "completed" | "failed") || "completed"
      );
    };

    // Subscribe to events
    window.runtime.EventsOn("chat:delta", handleDelta);
    window.runtime.EventsOn("chat:thinking", handleThinking);
    window.runtime.EventsOn("chat:tool_call", handleToolCall);
    window.runtime.EventsOn("chat:tool_result", handleToolResult);
    window.runtime.EventsOn("chat:error", handleError);
    window.runtime.EventsOn("chat:done", handleDone);
    window.runtime.EventsOn("chat:usage", handleUsage);
    window.runtime.EventsOn("chat:compressing", handleCompressing);
    window.runtime.EventsOn("chat:compressed", handleCompressed);
    window.runtime.EventsOn("safety:confirm", handleSafetyConfirm);
    window.runtime.EventsOn("agent:planning", handlePlanning);
    window.runtime.EventsOn("agent:plan_generated", handlePlanGenerated);
    window.runtime.EventsOn("agent:plan_step_start", handlePlanStepStart);
    window.runtime.EventsOn("agent:plan_step_done", handlePlanStepDone);

    return () => {
      window.runtime.EventsOff(
        "chat:delta",
        "chat:thinking",
        "chat:tool_call",
        "chat:tool_result",
        "chat:error",
        "chat:done",
        "chat:usage",
        "chat:compressing",
        "chat:compressed",
        "safety:confirm",
        "agent:planning",
        "agent:plan_generated",
        "agent:plan_step_start",
        "agent:plan_step_done"
      );
    };
  }, [addRestoredMessage, appendDelta, appendThinking, addToolCall, updateToolResult, setStreaming, setConfirmAction, setCompressing, setUsage, finalizeResponse]);

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

        // Clear existing messages before loading new session
        clearMessages();
        resetStreamState();

        sessStore.setCurrentSessionId(id);

        // Convert loaded messages to ChatMessage format
        for (const m of data.messages) {
          addRestoredMessage({
            role: m.role as "user" | "assistant",
            content: m.content,
            thinking: m.thinking,
            toolLines: m.toolLines,
            tokens: m.tokens,
            toolCalls: m.toolCalls,
            durationMs: m.durationMs,
          });
        }
        // Scroll to bottom after messages are rendered
        requestAnimationFrame(() => {
          const el = document.querySelector(".flex-1.overflow-y-auto");
          if (el) el.scrollTop = el.scrollHeight;
        });
      } catch (e) {
        console.error("Failed to load session:", e);
      }
    },
    [addRestoredMessage, sessStore, clearMessages, resetStreamState]
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
