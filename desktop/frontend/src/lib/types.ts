export interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  thinking?: string;
  toolCalls?: ToolCallEvent[];
  tokens?: number;
  duration?: number;
  selectedSkills?: string[];
  timestamp: number;
}

export interface ToolCallEvent {
  id: string;
  name: string;
  args: string;
  result?: string;
  status: "running" | "done" | "error";
}

export interface AgentUsage {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
}

export interface ConfirmAction {
  level: string;
  description: string;
  tool: string;
  params: Record<string, unknown>;
}
