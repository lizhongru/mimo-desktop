import { useState, useEffect } from "react";
import {
  Wrench,
  X,
  Search,
  Shield,
  ShieldAlert,
  ShieldCheck,
  Server,
} from "lucide-react";
import { t } from "../../lib/i18n";

interface ToolInfo {
  name: string;
  description: string;
  safetyLevel: string;
  isMcp: boolean;
  serverName: string;
}

interface MCPServerInfo {
  name: string;
  connected: boolean;
  toolCount: number;
  tools: string[];
}

interface Props {
  open: boolean;
  onClose: () => void;
}

function SafetyBadge({ level }: { level: string }) {
  const colors: Record<string, string> = {
    LOW: "text-green-400 bg-green-500/10",
    MEDIUM: "text-yellow-400 bg-yellow-500/10",
    HIGH: "text-orange-400 bg-orange-500/10",
    CRITICAL: "text-red-400 bg-red-500/10",
  };
  const Icon =
    level === "CRITICAL"
      ? ShieldAlert
      : level === "HIGH"
      ? Shield
      : ShieldCheck;

  return (
    <span
      className={`inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded ${
        colors[level] || "text-txt-g bg-elevated"
      }`}
    >
      <Icon className="w-2.5 h-2.5" />
      {level}
    </span>
  );
}

export function ToolsViewer({ open, onClose }: Props) {
  const [tools, setTools] = useState<ToolInfo[]>([]);
  const [servers, setServers] = useState<MCPServerInfo[]>([]);
  const [search, setSearch] = useState("");
  const [tab, setTab] = useState<"tools" | "mcp">("tools");

  useEffect(() => {
    if (!open) return;
    window.go?.desktop?.App?.GetTools?.().then(setTools).catch(console.error);
    window.go?.desktop?.App?.GetMCPServers?.().then(setServers).catch(console.error);
  }, [open]);

  if (!open) return null;

  const filtered = tools.filter(
    (tool) =>
      tool.name.toLowerCase().includes(search.toLowerCase()) ||
      tool.description.toLowerCase().includes(search.toLowerCase())
  );

  const builtinTools = filtered.filter((tool) => !tool.isMcp);
  const mcpTools = filtered.filter((tool) => tool.isMcp);

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 backdrop-blur-sm">
      <div className="bg-surface border border-bdr rounded-xl w-[700px] max-h-[80vh] mx-4 shadow-2xl flex flex-col">
        {/* Header */}
        <div className="flex items-center gap-3 px-5 py-3 border-b border-bdr-sub">
          <Wrench className="w-5 h-5 text-accent" />
          <h2 className="text-base font-medium text-txt">
            {t("tools_mcp")}
          </h2>
          <span className="text-xs text-txt-m">
            {tools.length} {t("tools_suffix")}, {servers.length} {t("servers_suffix")}
          </span>
          <button
            onClick={onClose}
            className="ml-auto p-1.5 rounded-md hover:bg-elevated text-txt-g hover:text-txt transition-colors cursor-pointer"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Tabs */}
        <div className="flex items-center gap-1 px-5 pt-2">
          <button
            onClick={() => setTab("tools")}
            className={`px-3 py-1.5 rounded-md text-sm transition-colors cursor-pointer ${
              tab === "tools"
                ? "bg-elevated text-txt"
                : "text-txt-g hover:text-txt-2"
            }`}
          >
            <Wrench className="w-3.5 h-3.5 inline mr-1.5" />
            {t("builtin_tools")} ({builtinTools.length})
          </button>
          <button
            onClick={() => setTab("mcp")}
            className={`px-3 py-1.5 rounded-md text-sm transition-colors cursor-pointer ${
              tab === "mcp"
                ? "bg-elevated text-txt"
                : "text-txt-g hover:text-txt-2"
            }`}
          >
            <Server className="w-3.5 h-3.5 inline mr-1.5" />
            {t("mcp_tools")} ({mcpTools.length})
          </button>
        </div>

        {/* Search */}
        <div className="px-5 py-2">
          <div className="relative">
            <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-txt-m" />
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t("search_tools")}
              className="w-full bg-elevated border border-bdr rounded-md pl-9 pr-3 py-1.5 text-sm text-txt placeholder:text-txt-m focus:outline-none focus:border-accent/50"
            />
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto px-5 pb-4">
          {tab === "tools" && (
            <div className="grid grid-cols-2 gap-2">
              {builtinTools.map((tool) => (
                <div
                  key={tool.name}
                  className="border border-bdr-sub rounded-lg p-3 hover:border-bdr transition-colors"
                >
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-mono text-sm text-txt">
                      {tool.name}
                    </span>
                    <SafetyBadge level={tool.safetyLevel} />
                  </div>
                  <p className="text-xs text-txt-g line-clamp-2">
                    {tool.description}
                  </p>
                </div>
              ))}
            </div>
          )}

          {tab === "mcp" && (
            <div className="space-y-3">
              {servers.map((server) => (
                <div
                  key={server.name}
                  className="border border-bdr-sub rounded-lg overflow-hidden"
                >
                  <div className="flex items-center gap-2 px-3 py-2.5 bg-elevated/40">
                    <Server className="w-4 h-4 text-blue-400" />
                    <span className="font-medium text-sm text-txt">
                      {server.name}
                    </span>
                    <span
                      className={`text-[10px] px-1.5 py-0.5 rounded ${
                        server.connected
                          ? "bg-green-500/10 text-green-400"
                          : "bg-red-500/10 text-red-400"
                      }`}
                    >
                      {server.connected ? t("connected") : t("disconnected")}
                    </span>
                    <span className="text-xs text-txt-m ml-auto">
                      {server.toolCount} {t("tools_suffix")}
                    </span>
                  </div>
                  {server.tools.length > 0 && (
                    <div className="px-3 py-2 flex flex-wrap gap-1.5">
                      {server.tools.map((toolName) => (
                        <span
                          key={toolName}
                          className="text-xs font-mono px-2 py-0.5 bg-elevated rounded text-txt-2"
                        >
                          {toolName}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              ))}
              {mcpTools.length > 0 && (
                <div className="grid grid-cols-2 gap-2 mt-2">
                  <div className="col-span-2 text-xs text-txt-m mb-1">
                    {t("all_mcp_tools")}
                  </div>
                  {mcpTools.map((tool) => (
                    <div
                      key={tool.name}
                      className="border border-bdr-sub rounded-lg p-3 hover:border-bdr transition-colors"
                    >
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-mono text-sm text-txt truncate">
                          {tool.name}
                        </span>
                        <SafetyBadge level={tool.safetyLevel} />
                      </div>
                      <p className="text-xs text-txt-g line-clamp-2">
                        {tool.description}
                      </p>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
