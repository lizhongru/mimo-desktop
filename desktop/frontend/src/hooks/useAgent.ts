import { useEffect } from "react";
import { EventsOn, EVENTS } from "../lib/events";
import { useChatStore } from "../stores/chatStore";
import { useActivityStore } from "../stores/activityStore";
import { t } from "../lib/i18n";
import type { AgentUsage, ConfirmAction } from "../lib/types";

export function useAgent() {
  const store = useChatStore;
  const activity = useActivityStore;

  useEffect(() => {
    const unsubs: (() => void)[] = [];

    unsubs.push(
      EventsOn(EVENTS.DELTA, (...args: unknown[]) => {
        store.getState().appendDelta(args[0] as string);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.THINKING, (...args: unknown[]) => {
        store.getState().appendThinking(args[0] as string);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.TOOL_CALL, (...args: unknown[]) => {
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
        activity.getState().addEntry({
          type: "tool_call",
          name: "compress_context",
          detail: t("compressing_context"),
          status: "running",
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
        store.getState().finalizeResponse(data.response, data.duration);
        // Auto-save session
        try {
          const msgs = store.getState().messages;
          const sid = (window as Record<string, unknown>).__currentSessionId as string | undefined;
          if (sid && msgs.length > 0) {
            window.go?.desktop?.App?.SaveSessionFromFrontend?.(sid, msgs).catch((err) => {
              console.warn("Auto-save session failed:", err);
            });
          }
        } catch (e) {
          console.warn("Auto-save error:", e);
        }
      })
    );

    unsubs.push(
      EventsOn(EVENTS.CHAT_ERROR, (...args: unknown[]) => {
        console.error("Chat error:", args[0]);
        store.getState().finalizeResponse(`${t("error_prefix")}: ${args[0]}`, 0);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.CHAT_CANCELLED, () => {
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
