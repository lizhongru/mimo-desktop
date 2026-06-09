import { create } from "zustand";

export interface SessionItem {
  id: string;
  modelName: string;
  userName: string;
  lastMessage: string;
  workingDir: string;
  createdAt: string;
  updatedAt: string;
}

interface SessionState {
  sessions: SessionItem[];
  currentSessionId: string | null;
  streamingSessionId: string | null;
  exportingSessionId: string | null;
  leftSidebarOpen: boolean;

  setSessions: (sessions: SessionItem[]) => void;
  setCurrentSessionId: (id: string | null) => void;
  setStreamingSessionId: (id: string | null) => void;
  setExportingSessionId: (id: string | null) => void;
  addSession: (session: SessionItem) => void;
  removeSession: (id: string) => void;
  updateSession: (id: string, lastMessage: string) => void;
  toggleLeftSidebar: () => void;
  setLeftSidebarOpen: (open: boolean) => void;
}

export const useSessionStore = create<SessionState>((set) => ({
  sessions: [],
  currentSessionId: null,
  streamingSessionId: null,
  exportingSessionId: null,
  leftSidebarOpen: true,

  setSessions: (sessions) => set({ sessions }),
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
        sess.id === id ? { ...sess, lastMessage } : sess
      ),
    })),
  toggleLeftSidebar: () =>
    set((s) => ({ leftSidebarOpen: !s.leftSidebarOpen })),
  setLeftSidebarOpen: (open) => set({ leftSidebarOpen: open }),
}));
