import { create } from "zustand";
import type { ChatMessage, ToolCallEvent, AgentUsage, ConfirmAction } from "../lib/types";

interface ChatState {
  messages: ChatMessage[];
  isStreaming: boolean;
  isThinking: boolean;
  currentThinking: string;
  currentDelta: string;
  currentToolCalls: ToolCallEvent[];
  usage: AgentUsage | null;
  confirmAction: ConfirmAction | null;

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
  setStreaming: (streaming: boolean) => void;
  clearMessages: () => void;
  resetStreamState: () => void;
}

let nextId = 1;
function genId() {
  return `msg-${nextId++}-${Date.now()}`;
}

export const useChatStore = create<ChatState>((set, get) => ({
  messages: [],
  isStreaming: false,
  isThinking: false,
  currentThinking: "",
  currentDelta: "",
  currentToolCalls: [],
  usage: null,
  confirmAction: null,

  addUserMessage: (content) => {
    const msg: ChatMessage = {
      id: genId(),
      role: "user",
      content,
      timestamp: Date.now(),
    };
    set((state) => ({
      messages: [...state.messages, msg],
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
      toolCalls: data.toolLines?.map((line, i) => ({
        id: `restored-${i}`,
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
    const tc: ToolCallEvent = { id: `tc-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`, name, args, status: "running" };
    set((state) => ({
      currentToolCalls: [...state.currentToolCalls, tc],
    }));
  },

  updateToolResult: (name, result) => {
    set((state) => ({
      currentToolCalls: state.currentToolCalls.map((tc) =>
        tc.name === name && tc.status === "running" ? { ...tc, result, status: "done" as const } : tc
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
    set((s) => ({
      messages: [...s.messages, msg],
      isStreaming: false,
      isThinking: false,
      currentThinking: "",
      currentDelta: "",
      currentToolCalls: [],
    }));
  },

  setUsage: (usage) => set({ usage }),

  setConfirmAction: (action) => set({ confirmAction: action }),

  setStreaming: (streaming) => set({ isStreaming: streaming }),

  clearMessages: () =>
    set({
      messages: [],
      isStreaming: false,
      isThinking: false,
      currentThinking: "",
      currentDelta: "",
      currentToolCalls: [],
      usage: null,
    }),

  resetStreamState: () =>
    set({
      isStreaming: false,
      isThinking: false,
      currentThinking: "",
      currentDelta: "",
      currentToolCalls: [],
    }),
}));
