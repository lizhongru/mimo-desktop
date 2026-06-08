import { create } from "zustand";

export interface ActivityEntry {
  id: string;
  type: "tool_call" | "file_change" | "plan_step" | "error";
  name: string;
  detail: string;
  status: "running" | "done" | "error";
  timestamp: number;
}

export interface FileDiff {
  path: string;
  additions: number;
  deletions: number;
}

export interface PlanStep {
  id: number;
  description: string;
  status: "pending" | "in_progress" | "completed" | "failed" | "skipped";
}

export interface PlanInfo {
  goal: string;
  steps: PlanStep[];
  currentStep: number;
  totalSteps: number;
}

interface ActivityState {
  entries: ActivityEntry[];
  fileDiffs: FileDiff[];
  plan: PlanInfo | null;
  rightSidebarOpen: boolean;

  addEntry: (entry: Omit<ActivityEntry, "id" | "timestamp">) => void;
  updateEntry: (name: string, updates: Partial<ActivityEntry>) => void;
  addFileDiff: (diff: FileDiff) => void;
  setPlan: (plan: PlanInfo | null) => void;
  updatePlanStep: (stepId: number, status: PlanStep["status"]) => void;
  toggleRightSidebar: () => void;
  setRightSidebarOpen: (open: boolean) => void;
  clear: () => void;
}

let nextId = 1;

export const useActivityStore = create<ActivityState>((set) => ({
  entries: [],
  fileDiffs: [],
  plan: null,
  rightSidebarOpen: false,

  addEntry: (entry) =>
    set((s) => ({
      entries: [
        { ...entry, id: `act-${nextId++}`, timestamp: Date.now() },
        ...s.entries,
      ].slice(0, 100), // keep last 100
      rightSidebarOpen: true,
    })),

  updateEntry: (name, updates) =>
    set((s) => ({
      entries: s.entries.map((e) =>
        e.name === name ? { ...e, ...updates } : e
      ),
    })),

  addFileDiff: (diff) =>
    set((s) => {
      const exists = s.fileDiffs.findIndex((f) => f.path === diff.path);
      if (exists >= 0) {
        const updated = [...s.fileDiffs];
        updated[exists] = diff;
        return { fileDiffs: updated };
      }
      return { fileDiffs: [...s.fileDiffs, diff] };
    }),

  setPlan: (plan) => set({ plan, rightSidebarOpen: plan !== null }),

  updatePlanStep: (stepId, status) =>
    set((s) => {
      if (!s.plan) return {};
      return {
        plan: {
          ...s.plan,
          steps: s.plan.steps.map((step) =>
            step.id === stepId ? { ...step, status } : step
          ),
        },
      };
    }),

  toggleRightSidebar: () =>
    set((s) => ({ rightSidebarOpen: !s.rightSidebarOpen })),

  setRightSidebarOpen: (open) => set({ rightSidebarOpen: open }),

  clear: () => set({ entries: [], fileDiffs: [], plan: null }),
}));
