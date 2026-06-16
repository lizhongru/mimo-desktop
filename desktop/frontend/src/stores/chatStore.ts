import { create } from "zustand";
import { useSessionStore } from "./sessionStore";
import type { ChatMessage, ToolCallEvent, AgentUsage, ConfirmAction } from "../lib/types";

interface ChatState {
  messages: ChatMessage[];
  activeSessionId: string | null;
  backgroundSessionId: string | null;
  backgroundMessages: ChatMessage[];
  backgroundIsThinking: boolean;
  backgroundThinking: string;
  backgroundDelta: string;
  backgroundToolCalls: ToolCallEvent[];
  sessionSnapshots: Record<string, ChatMessage[]>;
  isStreaming: boolean;
  isThinking: boolean;
  currentThinking: string;
  currentDelta: string;
  currentToolCalls: ToolCallEvent[];
  usage: AgentUsage | null;
  confirmAction: ConfirmAction | null;
  isCompressing: boolean;

  // Actions
  addUserMessage: (content: string) => void;
  addRestoredMessage: (msg: {
    role: "user" | "assistant";
    content: string;
    thinking?: string;
    toolLines?: string[];
    tokens?: number;
    toolCalls?: number;
    durationMs?: number;
  }) => void;
  appendDelta: (delta: string) => void;
  appendThinking: (delta: string) => void;
  addToolCall: (name: string, args: string) => void;
  updateToolResult: (name: string, result: string) => void;
  finalizeResponse: (response: string, duration: number) => void;
  setUsage: (usage: AgentUsage) => void;
  setConfirmAction: (action: ConfirmAction | null) => void;
  setCompressing: (compressing: boolean) => void;
  setStreaming: (streaming: boolean) => void;
  clearMessages: () => void;
  resetStreamState: () => void;
  setActiveSessionId: (id: string | null) => void;
  stashCurrentStreamToBackground: () => void;
  restoreBackgroundStream: (sessionId: string) => boolean;
  appendBackgroundDelta: (delta: string) => void;
  appendBackgroundThinking: (delta: string) => void;
  addBackgroundToolCall: (name: string, args: string) => void;
  updateBackgroundToolResult: (name: string, result: string) => void;
  finalizeBackgroundResponse: (response: string, duration: number) => ChatMessage[];
  cancelBackgroundResponse: () => ChatMessage[];
  clearBackgroundStream: () => void;
  setSessionSnapshot: (sessionId: string, messages: ChatMessage[]) => void;
  getSessionSnapshot: (sessionId: string) => ChatMessage[] | undefined;
  clearSessionSnapshot: (sessionId: string) => void;
  replaceMessages: (messages: ChatMessage[]) => void;
  deleteMessage: (id: string) => void;
}

let nextId = 1;
function genId() {
  return `msg-${nextId++}-${Date.now()}`;
}

function makeToolCall(name: string, args: string): ToolCallEvent {
  return {
    id: `tc-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    name,
    args,
    status: "running",
  };
}

export const useChatStore = create<ChatState>((set, get) => ({
  messages: [],
  activeSessionId: null,
  backgroundSessionId: null,
  backgroundMessages: [],
  backgroundIsThinking: false,
  backgroundThinking: "",
  backgroundDelta: "",
  backgroundToolCalls: [],
  sessionSnapshots: {},
  isStreaming: false,
  isThinking: false,
  currentThinking: "",
  currentDelta: "",
  currentToolCalls: [],
  usage: null,
  confirmAction: null,
  isCompressing: false,

  addUserMessage: (content) => {
    const msg: ChatMessage = {
      id: genId(),
      role: "user",
      content,
      timestamp: Date.now(),
    };
    set((state) => ({
      messages: [...state.messages, msg],
      activeSessionId: useSessionStore.getState().currentSessionId,
      isStreaming: true,
      isThinking: false,
      currentThinking: "",
      currentDelta: "",
      currentToolCalls: [],
    }));
  },

  addRestoredMessage: (data) => {
    const msg: ChatMessage = {
      id: genId(),
      role: data.role,
      content: data.content,
      thinking: data.thinking || undefined,
      toolCalls: data.toolLines?.map((line, index) => ({
        id: `restored-${index}`,
        name: line.split("(")[0]?.trim() || "tool",
        args: "",
        result: line,
        status: "done" as const,
      })),
      tokens: data.tokens || undefined,
      duration: data.durationMs || undefined,
      timestamp: Date.now(),
    };
    set((state) => ({
      messages: [...state.messages, msg],
    }));
  },

  appendDelta: (delta) => {
    set((state) => ({
      currentDelta: state.currentDelta + delta,
      isThinking: false,
    }));
  },

  appendThinking: (delta) => {
    set((state) => ({
      currentThinking: state.currentThinking + delta,
      isThinking: true,
    }));
  },

  addToolCall: (name, args) => {
    set((state) => ({
      currentToolCalls: [...state.currentToolCalls, makeToolCall(name, args)],
    }));
  },

  updateToolResult: (name, result) => {
    set((state) => ({
      currentToolCalls: state.currentToolCalls.map((toolCall) =>
        toolCall.name === name && toolCall.status === "running"
          ? { ...toolCall, result, status: "done" as const }
          : toolCall
      ),
    }));
  },

  finalizeResponse: (response, duration) => {
    const state = get();
    const msg: ChatMessage = {
      id: genId(),
      role: "assistant",
      content: response || state.currentDelta,
      thinking: state.currentThinking || undefined,
      toolCalls: state.currentToolCalls.length > 0 ? [...state.currentToolCalls] : undefined,
      duration,
      timestamp: Date.now(),
    };
    set((current) => ({
      messages: [...current.messages, msg],
      activeSessionId: null,
      isStreaming: false,
      isThinking: false,
      currentThinking: "",
      currentDelta: "",
      currentToolCalls: [],
    }));
  },

  setUsage: (usage) => set({ usage }),
  setConfirmAction: (action) => set({ confirmAction: action }),
  setCompressing: (compressing) => set({ isCompressing: compressing }),
  setStreaming: (streaming) => set({ isStreaming: streaming }),

  clearMessages: () =>
    set({
      messages: [],
      activeSessionId: null,
      isStreaming: false,
      isThinking: false,
      currentThinking: "",
      currentDelta: "",
      currentToolCalls: [],
      usage: null,
    }),

  resetStreamState: () =>
    set({
      activeSessionId: null,
      isStreaming: false,
      isThinking: false,
      currentThinking: "",
      currentDelta: "",
      currentToolCalls: [],
    }),

  setActiveSessionId: (id) => set({ activeSessionId: id }),

  stashCurrentStreamToBackground: () => {
    const state = get();
    if (!state.activeSessionId || !state.isStreaming) {
      return;
    }

    set({
      backgroundSessionId: state.activeSessionId,
      backgroundMessages: [...state.messages],
      backgroundIsThinking: state.isThinking,
      backgroundThinking: state.currentThinking,
      backgroundDelta: state.currentDelta,
      backgroundToolCalls: [...state.currentToolCalls],
      activeSessionId: null,
      isStreaming: false,
      isThinking: false,
      currentThinking: "",
      currentDelta: "",
      currentToolCalls: [],
    });
  },

  restoreBackgroundStream: (sessionId) => {
    const state = get();
    if (state.backgroundSessionId !== sessionId) {
      return false;
    }

    set({
      messages: [...state.backgroundMessages],
      activeSessionId: state.backgroundSessionId,
      isStreaming: true,
      isThinking: state.backgroundIsThinking,
      currentThinking: state.backgroundThinking,
      currentDelta: state.backgroundDelta,
      currentToolCalls: [...state.backgroundToolCalls],
      backgroundSessionId: null,
      backgroundMessages: [],
      backgroundIsThinking: false,
      backgroundThinking: "",
      backgroundDelta: "",
      backgroundToolCalls: [],
    });

    return true;
  },

  appendBackgroundDelta: (delta) =>
    set((state) => ({
      backgroundDelta: state.backgroundDelta + delta,
      backgroundIsThinking: false,
    })),

  appendBackgroundThinking: (delta) =>
    set((state) => ({
      backgroundThinking: state.backgroundThinking + delta,
      backgroundIsThinking: true,
    })),

  addBackgroundToolCall: (name, args) =>
    set((state) => ({
      backgroundToolCalls: [...state.backgroundToolCalls, makeToolCall(name, args)],
    })),

  updateBackgroundToolResult: (name, result) =>
    set((state) => ({
      backgroundToolCalls: state.backgroundToolCalls.map((toolCall) =>
        toolCall.name === name && toolCall.status === "running"
          ? { ...toolCall, result, status: "done" as const }
          : toolCall
      ),
    })),

  finalizeBackgroundResponse: (response, duration) => {
    const state = get();
    const msg: ChatMessage = {
      id: genId(),
      role: "assistant",
      content: response || state.backgroundDelta,
      thinking: state.backgroundThinking || undefined,
      toolCalls: state.backgroundToolCalls.length > 0 ? [...state.backgroundToolCalls] : undefined,
      duration,
      timestamp: Date.now(),
    };
    const finalMessages = [...state.backgroundMessages, msg];
    set({
      backgroundSessionId: null,
      backgroundMessages: [],
      backgroundIsThinking: false,
      backgroundThinking: "",
      backgroundDelta: "",
      backgroundToolCalls: [],
    });
    return finalMessages;
  },

  cancelBackgroundResponse: () => {
    const state = get();
    const partial = state.backgroundDelta;
    const finalMessages = partial
      ? [
          ...state.backgroundMessages,
          {
            id: genId(),
            role: "assistant" as const,
            content: `${partial} _(cancelled)_`,
            thinking: state.backgroundThinking || undefined,
            toolCalls: state.backgroundToolCalls.length > 0 ? [...state.backgroundToolCalls] : undefined,
            timestamp: Date.now(),
          },
        ]
      : [...state.backgroundMessages];

    set({
      backgroundSessionId: null,
      backgroundMessages: [],
      backgroundIsThinking: false,
      backgroundThinking: "",
      backgroundDelta: "",
      backgroundToolCalls: [],
    });

    return finalMessages;
  },

  clearBackgroundStream: () =>
    set({
      backgroundSessionId: null,
      backgroundMessages: [],
      backgroundIsThinking: false,
      backgroundThinking: "",
      backgroundDelta: "",
      backgroundToolCalls: [],
    }),

  setSessionSnapshot: (sessionId, messages) =>
    set((state) => ({
      sessionSnapshots: {
        ...state.sessionSnapshots,
        [sessionId]: [...messages],
      },
    })),

  getSessionSnapshot: (sessionId) => get().sessionSnapshots[sessionId],

  clearSessionSnapshot: (sessionId) =>
    set((state) => {
      const { [sessionId]: _removed, ...nextSnapshots } = state.sessionSnapshots;
      return { sessionSnapshots: nextSnapshots };
    }),

  replaceMessages: (messages) => set({ messages: [...messages] }),

  deleteMessage: (id) =>
    set((state) => ({
      messages: state.messages.filter((message) => message.id !== id),
    })),
}));
