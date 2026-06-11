import { useState, useCallback, useEffect } from "react";
import { Users, Check, Settings } from "lucide-react";

interface AgentConfig {
  name: string;
  mode: string;
  color: string;
  description: string;
  prompt: string;
  tool_allowlist?: string[];
}

interface AgentSwitchResult {
  success: boolean;
  message: string;
  agent?: AgentConfig;
}

export function AgentSwitcher() {
  const [agents, setAgents] = useState<AgentConfig[]>([]);
  const [currentAgent, setCurrentAgent] = useState<AgentConfig | null>(null);
  const [isExpanded, setIsExpanded] = useState(false);

  const loadAgents = useCallback(async () => {
    try {
      const configs = await window.go?.desktop?.App?.AgentListConfigs?.();
      setAgents(configs || []);

      const current = await window.go?.desktop?.App?.AgentGetCurrent?.();
      setCurrentAgent(current);
    } catch (error) {
      console.error("Failed to load agents:", error);
    }
  }, []);

  useEffect(() => {
    loadAgents();
  }, [loadAgents]);

  const handleSwitch = useCallback(
    async (name: string) => {
      try {
        const result: AgentSwitchResult = await window.go?.desktop?.App?.AgentSwitch?.(name);
        if (result?.success) {
          setCurrentAgent(result.agent || null);
          setIsExpanded(false);
        }
      } catch (error) {
        console.error("Failed to switch agent:", error);
      }
    },
    []
  );

  return (
    <div className="relative">
      {/* Current agent button */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="flex items-center gap-2 px-3 py-1.5 text-sm rounded-md hover:bg-elevated transition-colors"
        style={{
          borderLeft: `3px solid ${currentAgent?.color || "#6b7280"}`,
        }}
      >
        <Users className="w-4 h-4" />
        <span>{currentAgent?.name || "Build"}</span>
      </button>

      {/* Dropdown */}
      {isExpanded && (
        <>
          <div
            className="fixed inset-0 z-40"
            onClick={() => setIsExpanded(false)}
          />
          <div className="absolute left-0 top-full mt-1 w-64 bg-surface border border-bdr rounded-lg shadow-xl z-50 animate-pop-up">
            <div className="px-3 py-2 border-b border-bdr">
              <span className="text-xs font-medium text-txt-m">
                选择智能体
              </span>
            </div>
            <div className="py-1">
              {agents.map((agent) => (
                <button
                  key={agent.name}
                  onClick={() => handleSwitch(agent.name)}
                  className="w-full flex items-center gap-3 px-3 py-2 text-sm hover:bg-elevated transition-colors"
                >
                  <div
                    className="w-3 h-3 rounded-full"
                    style={{ backgroundColor: agent.color }}
                  />
                  <div className="flex-1 text-left">
                    <div className="font-medium">{agent.name}</div>
                    <div className="text-xs text-txt-m truncate">
                      {agent.description}
                    </div>
                  </div>
                  {currentAgent?.name === agent.name && (
                    <Check className="w-4 h-4 text-accent" />
                  )}
                </button>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
