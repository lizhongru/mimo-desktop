import { useEffect } from "react";
import { EventsOn, EVENTS } from "../lib/events";
import { useChatStore } from "../stores/chatStore";
import { useActivityStore } from "../stores/activityStore";
import { useSessionStore } from "../stores/sessionStore";
import { t } from "../lib/i18n";
import type { AgentUsage, ConfirmAction } from "../lib/types";

export function useAgent() {
  const store = useChatStore;
  const activity = useActivityStore;

  useEffect(() => {
    const unsubs: (() => void) = [];

    // Guard: only process events if session matches
    const isActiveSession = () => {
      const active = store.getState().activeSessionId;
      const current = useSessionStore.getState().currentSessionId;
      return !active || active === current;
    };

    unsubs.push(
      EventsOn(EVENTS.DELTA, (...args: unknown[]) => {
        if (!isActiveSession()) return;
        store.getState().appendDelta(args[0] as string);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.THINKING, (...args: unknown[]) => {
        if (!isActiveSession()) return;
        store.getState().appendThinking(args[0] as string);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.TOOL_CALL, (...args: unknown[]) => {
        if (!isActiveSession()) return;
        const name = args[0] as string;
        const toolArgs = args[1] as string;
        store.getState().addToolCall(name, toolArgs);
        // Add to activity log
        activity.getState().addEntry({
          type: "tool_call",
          name,
          detail: toolArgs,
          status: "running",
        });
      })
    );

    unsubs.push(
      EventsOn(EVENTS.TOOL_RESULT, (...args: unknown[]) => {
        if (!isActiveSession()) return;
        const name = args[0] as string;
        const result = args[1] as string;
        store.getState().updateToolResult(name, result);
        // Update activity log
        activity.getState().updateEntry(name, {
          status: result.startsWith("Error") ? "error" : "done",
          detail: result,
        });
      })
    );

    unsubs.push(
      EventsOn(EVENTS.USAGE, (...args: unknown[]) => {
        store.getState().setUsage(args[0] as AgentUsage);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.ERROR, (...args: unknown[]) => {
        const err = args[0] as string;
        console.error("Agent error:", err);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.COMPRESSING, () => {
        store.getState().setCompressing(true);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.COMPRESS_DONE, (...args: unknown[]) => {
        const data = args[0] as { before: number; after: number };
        store.getState().setCompressing(false);
        const saved = data.before - data.after;
        const pct = data.before > 0 ? Math.round((saved / data.before) * 100) : 0;
        activity.getState().addEntry({
          type: "tool_call",
          name: "compress_done",
          detail: `${t("compress_done")} ${data.before} → ${data.after} (-${saved}, -${pct}%)`,
          status: "done",
        });
      })
    );

    unsubs.push(
      EventsOn(EVENTS.PLANNING, (...args: unknown[]) => {
        const msg = args[0] as string;
        activity.getState().addEntry({
          type: "plan_step",
          name: "planning",
          detail: msg,
          status: "running",
        });
      })
    );

    unsubs.push(
      EventsOn(EVENTS.PLAN_GENERATED, (...args: unknown[]) => {
        const plan = args[0] as {
          goal: string;
          steps: Array<{
            id: number;
            description: string;
            status: string;
          }>;
          totalSteps: number;
        };
        activity.getState().setPlan({
          goal: plan.goal,
          steps: plan.steps.map((s) => ({
            id: s.id,
            description: s.description,
            status: s.status as "pending" | "in_progress" | "completed" | "failed" | "skipped",
          })),
          currentStep: 0,
          totalSteps: plan.totalSteps,
        });
      })
    );

    unsubs.push(
      EventsOn(EVENTS.PLAN_STEP_START, (...args: unknown[]) => {
        const step = args[0] as { id: number; description: string };
        activity.getState().updatePlanStep(step.id, "in_progress");
      })
    );

    unsubs.push(
      EventsOn(EVENTS.PLAN_STEP_DONE, (...args: unknown[]) => {
        const step = args[0] as { id: number; status: string };
        activity.getState().updatePlanStep(
          step.id,
          step.status === "failed" ? "failed" : "completed"
        );
      })
    );

    unsubs.push(
      EventsOn(EVENTS.CHAT_DONE, (...args: unknown[]) => {
        const data = args[0] as { response: string; duration: number };
        const activeSid = store.getState().activeSessionId;
        const currentSid = useSessionStore.getState().currentSessionId;
        if (activeSid && activeSid !== currentSid) {
          const msgs = store.getState().messages;
          const response = data.response || store.getState().currentDelta;
          const finalMsgs = [...msgs, { id: 'final-' + Date.now(), role: 'assistant' as const, content: response, thinking: store.getState().currentThinking || undefined, toolCalls: store.getState().currentToolCalls.length > 0 ? [...store.getState().currentToolCalls] : undefined, duration: data.duration, timestamp: Date.now() }];
          const dto = finalMsgs.map((m) => ({ role: m.role, content: m.content, thinking: m.thinking || '', toolLines: (m.toolCalls || []).map((tc) => tc.name + '(' + tc.args + ')'), tokens: m.tokens || 0, toolCalls: m.toolCalls?.length || 0, durationMs: m.duration || 0 }));
          window.go?.desktop?.App?.SaveSessionFromFrontend?.(activeSid, dto).catch(console.error);
          store.getState().resetStreamState();
          useSessionStore.getState().setStreamingSessionId(null);
          return;
        }
        store.getState().finalizeResponse(data.response, data.duration);
        useSessionStore.getState().setStreamingSessionId(null);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.CHAT_ERROR, (...args: unknown[]) => {
        console.error("Chat error:", args[0]);
        useSessionStore.getState().setStreamingSessionId(null);
        store.getState().finalizeResponse(`${t("error_prefix")}: ${args[0]}`, 0);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.CHAT_CANCELLED, () => {
        // Save partial content to the original session if user switched
        const activeSid = store.getState().activeSessionId;
        const currentSid = useSessionStore.getState().currentSessionId;
        if (activeSid && activeSid !== currentSid) {
          const partial = store.getState().currentDelta;
          if (partial) {
            const msgs = store.getState().messages;
            const finalMsgs = [...msgs, { id: 'cancel-' + Date.now(), role: 'assistant' as const, content: partial + ' _(cancelled)_', thinking: store.getState().currentThinking || undefined, timestamp: Date.now() }];
            const dto = finalMsgs.map((m) => ({ role: m.role, content: m.content, thinking: m.thinking || '', toolLines: [], tokens: 0, toolCalls: 0, durationMs: 0 }));
            window.go?.desktop?.App?.SaveSessionFromFrontend?.(activeSid, dto).catch(console.error);
          }
        }
        useSessionStore.getState().setStreamingSessionId(null);
        store.getState().resetStreamState();
      })
    );

    unsubs.push(
      EventsOn(EVENTS.SAFETY_CONFIRM, (...args: unknown[]) => {
        store.getState().setConfirmAction(args[0] as ConfirmAction);
      })
    );

    return () => {
      unsubs.forEach((fn) => fn());
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps
}
