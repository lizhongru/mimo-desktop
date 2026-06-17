import { create } from "zustand";

export interface ActivityEntry {
  id: string;
  type: "tool_call" | "file_change" | "plan_step" | "error";
  name: string;
  detail: string;
  status: "running" | "done" | "error";
  timestamp: number;
  lastUpdated: number;
  count: number;
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
  pruneOldEntries: () => void;
  toggleRightSidebar: () => void;
  setRightSidebarOpen: (open: boolean) => void;
  clear: () => void;
}

let nextId = 1;

const MAX_ACTIVITY_ENTRIES = 12;
const DONE_ENTRY_TTL_MS = 45_000;
const ERROR_ENTRY_TTL_MS = 120_000;

function compactEntries(entries: ActivityEntry[], now = Date.now()): ActivityEntry[] {
  return entries
    .filter((entry) => {
      if (entry.status === "running") return true;
      const ttl = entry.status === "error" ? ERROR_ENTRY_TTL_MS : DONE_ENTRY_TTL_MS;
      return now - entry.lastUpdated <= ttl;
    })
    .slice(0, MAX_ACTIVITY_ENTRIES);
}

export const useActivityStore = create<ActivityState>((set) => ({
  entries: [],
  fileDiffs: [],
  plan: null,
  rightSidebarOpen: false,

  addEntry: (entry) =>
    set((s) => {
      const now = Date.now();
      const existingIndex = s.entries.findIndex(
        (existing) => existing.type === entry.type && existing.name === entry.name
      );
      const entries = [...s.entries];
      if (existingIndex >= 0) {
        const existing = entries[existingIndex];
        entries.splice(existingIndex, 1);
        entries.unshift({
          ...existing,
          ...entry,
          detail: entry.detail || existing.detail,
          count: existing.count + 1,
          lastUpdated: now,
        });
      } else {
        entries.unshift({
          ...entry,
          id: `act-${nextId++}`,
          timestamp: now,
          lastUpdated: now,
          count: 1,
        });
      }
      return {
        entries: compactEntries(entries, now),
        rightSidebarOpen: true,
      };
    }),

  updateEntry: (name, updates) =>
    set((s) => {
      const now = Date.now();
      const entries = s.entries.map((entry) =>
        entry.name === name ? { ...entry, ...updates, lastUpdated: now } : entry
      );
      return { entries: compactEntries(entries, now) };
    }),

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

  pruneOldEntries: () =>
    set((s) => ({ entries: compactEntries(s.entries) })),

  toggleRightSidebar: () =>
    set((s) => ({ rightSidebarOpen: !s.rightSidebarOpen })),

  setRightSidebarOpen: (open) => set({ rightSidebarOpen: open }),

  clear: () => set({ entries: [], fileDiffs: [], plan: null }),
}));
