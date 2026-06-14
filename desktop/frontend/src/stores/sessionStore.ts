import { create } from "zustand";

export interface SessionItem {
  id: string;
  workspaceId: string;
  modelName: string;
  userName: string;
  lastMessage: string;
  createdAt: string;
  updatedAt: string;
}

export interface WorkspaceItem {
  id: string;
  name: string;
  type: string; // "chat" | "folder"
  path: string;
}

const WS_STORAGE_KEY = "mimo_selectedWorkspace";

interface SessionState {
  sessions: SessionItem[];
  workspaces: WorkspaceItem[];
  currentSessionId: string | null;
  streamingSessionId: string | null;
  exportingSessionId: string | null;
  leftSidebarOpen: boolean;
  selectedWorkspace: string; // workspace ID

  setSessions: (sessions: SessionItem[]) => void;
  setWorkspaces: (workspaces: WorkspaceItem[]) => void;
  setCurrentSessionId: (id: string | null) => void;
  setStreamingSessionId: (id: string | null) => void;
  setExportingSessionId: (id: string | null) => void;
  addSession: (session: SessionItem) => void;
  removeSession: (id: string) => void;
  updateSession: (id: string, lastMessage: string) => void;
  updateSessionWorkspace: (id: string, workspaceId: string) => void;
  toggleLeftSidebar: () => void;
  setLeftSidebarOpen: (open: boolean) => void;
  setSelectedWorkspace: (id: string) => void;
}

function loadSelectedWorkspace(): string {
  try {
    return localStorage.getItem(WS_STORAGE_KEY) || "";
  } catch {
    return "";
  }
}

export const useSessionStore = create<SessionState>((set) => ({
  sessions: [],
  workspaces: [],
  currentSessionId: null,
  streamingSessionId: null,
  exportingSessionId: null,
  leftSidebarOpen: true,
  selectedWorkspace: loadSelectedWorkspace(),

  setSessions: (sessions) => set({ sessions }),
  setWorkspaces: (workspaces) => set({ workspaces }),
  setCurrentSessionId: (id) => set({ currentSessionId: id }),
  setStreamingSessionId: (id) => set({ streamingSessionId: id }),
  setExportingSessionId: (id) => set({ exportingSessionId: id }),
  addSession: (session) =>
    set((s) => ({ sessions: [session, ...s.sessions] })),
  removeSession: (id) =>
    set((s) => ({
      sessions: s.sessions.filter((sess) => sess.id !== id),
      currentSessionId: s.currentSessionId === id ? null : s.currentSessionId,
    })),
  updateSession: (id, lastMessage) =>
    set((s) => ({
      sessions: s.sessions.map((sess) =>
        sess.id === id
          ? { ...sess, lastMessage, updatedAt: new Date().toISOString() }
          : sess
      ),
    })),
  updateSessionWorkspace: (id, workspaceId) =>
    set((s) => ({
      sessions: s.sessions.map((sess) =>
        sess.id === id
          ? { ...sess, workspaceId, updatedAt: new Date().toISOString() }
          : sess
      ),
    })),
  toggleLeftSidebar: () =>
    set((s) => ({ leftSidebarOpen: !s.leftSidebarOpen })),
  setLeftSidebarOpen: (open) => set({ leftSidebarOpen: open }),
  setSelectedWorkspace: (id) => {
    try {
      if (id) {
        localStorage.setItem(WS_STORAGE_KEY, id);
      } else {
        localStorage.removeItem(WS_STORAGE_KEY);
      }
    } catch { /* ignore */ }
    set({ selectedWorkspace: id });
  },
}));
