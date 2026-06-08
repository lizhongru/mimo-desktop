import { useState, useEffect } from "react";
import {
  X,
  Settings as SettingsIcon,
  Wrench,
  Globe,
  Type,
  Palette,
  Cpu,
  HelpCircle,
  Plus,
  Trash2,
  Server,
  Search,
  Shield,
  ShieldAlert,
  ShieldCheck,
  Info,
  Keyboard,
  ExternalLink,
} from "lucide-react";
import { t, td } from "../../lib/i18n";
import { useSettingsStore } from "../../stores/settingsStore";

type Tab = "general" | "tools" | "models" | "help";

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

interface ModelDTO {
  apiBase: string;
  model: string;
  maxTokens: number;
  temperature: number;
}

interface Props {
  open: boolean;
  onClose: () => void;
  defaultModel: string;
}

export function SettingsPage({ open, onClose, defaultModel }: Props) {
  const [tab, setTab] = useState<Tab>("general");
  const { theme, language, fontSize, setTheme, setLanguage, setFontSize } =
    useSettingsStore();

  // Tools state
  const [tools, setTools] = useState<ToolInfo[]>([]);
  const [servers, setServers] = useState<MCPServerInfo[]>([]);
  const [toolSearch, setToolSearch] = useState("");
  const [toolTab, setToolTab] = useState<"builtin" | "mcp">("builtin");

  // Models state
  const [models, setModels] = useState<Record<string, ModelDTO>>({});
  const [currentModel, setCurrentModel] = useState(defaultModel);
  const [showAddModel, setShowAddModel] = useState(false);
  const [newModel, setNewModel] = useState({
    name: "",
    apiBase: "",
    apiKey: "",
    model: "",
    maxTokens: 32768,
    temperature: 0.3,
  });

  useEffect(() => {
    if (!open) return;
    window.go?.desktop?.App?.GetTools?.().then(setTools).catch(console.error);
    window.go?.desktop?.App?.GetMCPServers?.()
      .then(setServers)
      .catch(console.error);
    window.go?.desktop?.App?.GetConfig?.()
      .then((cfg) => {
        setModels(cfg.models || {});
        setCurrentModel(cfg.defaultModel);
      })
      .catch(console.error);
  }, [open]);

  if (!open) return null;

  const builtinTools = tools.filter((t) => !t.isMcp);
  const mcpTools = tools.filter((t) => t.isMcp);
  const filteredBuiltin = builtinTools.filter(
    (t) =>
      t.name.toLowerCase().includes(toolSearch.toLowerCase()) ||
      t.description.toLowerCase().includes(toolSearch.toLowerCase())
  );
  const filteredMcp = mcpTools.filter(
    (t) =>
      t.name.toLowerCase().includes(toolSearch.toLowerCase()) ||
      t.description.toLowerCase().includes(toolSearch.toLowerCase())
  );

  const handleAddModel = () => {
    if (!newModel.name || !newModel.apiBase) return;
    window.go?.desktop?.App?.AddModel?.(
      newModel.name,
      newModel.apiBase,
      newModel.apiKey,
      newModel.model,
      newModel.maxTokens,
      newModel.temperature
    )
      .then(() => {
        setModels((prev) => ({
          ...prev,
          [newModel.name]: {
            apiBase: newModel.apiBase,
            model: newModel.model,
            maxTokens: newModel.maxTokens,
            temperature: newModel.temperature,
          },
        }));
        setShowAddModel(false);
        setNewModel({
          name: "",
          apiBase: "",
          apiKey: "",
          model: "",
          maxTokens: 32768,
          temperature: 0.3,
        });
      })
      .catch(console.error);
  };

  const handleSetDefault = (name: string) => {
    window.go?.desktop?.App?.SetDefaultModel?.(name)
      .then(() => setCurrentModel(name))
      .catch(console.error);
  };

  const handleRemoveModel = (name: string) => {
    window.go?.desktop?.App?.RemoveModel?.(name)
      .then(() => {
        setModels((prev) => {
          const next = { ...prev };
          delete next[name];
          return next;
        });
      })
      .catch(console.error);
  };

  const SafetyBadge = ({ level }: { level: string }) => {
    const colors: Record<string, string> = {
      LOW: "text-green-400 bg-green-500/10",
      MEDIUM: "text-yellow-400 bg-yellow-500/10",
      HIGH: "text-orange-400 bg-orange-500/10",
      CRITICAL: "text-red-400 bg-red-500/10",
    };
    const Icon =
      level === "CRITICAL" ? ShieldAlert : level === "HIGH" ? Shield : ShieldCheck;
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
  };

  const tabs: { id: Tab; icon: typeof SettingsIcon; label: string }[] = [
    { id: "general", icon: SettingsIcon, label: t("general") },
    { id: "tools", icon: Wrench, label: t("tools_mcp") },
    { id: "models", icon: Cpu, label: t("models") },
    { id: "help", icon: HelpCircle, label: t("help_feedback") },
  ];

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 backdrop-blur-sm">
      <div className="bg-surface border border-bdr rounded-xl w-[780px] h-[600px] mx-4 shadow-2xl flex flex-col overflow-hidden">
        {/* Header */}
        <div className="flex items-center gap-3 px-5 py-3 border-b border-bdr-sub flex-shrink-0">
          <SettingsIcon className="w-5 h-5 text-accent" />
          <h2 className="text-base font-medium text-txt">
            {t("settings")}
          </h2>
          <button
            onClick={onClose}
            className="ml-auto p-1.5 rounded-md hover:bg-elevated text-txt-g hover:text-txt transition-colors cursor-pointer"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="flex flex-1 overflow-hidden">
          {/* Tab sidebar */}
          <div className="w-[160px] border-r border-bdr-sub p-2 flex-shrink-0">
            {tabs.map(({ id, icon: Icon, label }) => (
              <button
                key={id}
                onClick={() => setTab(id)}
                className={`w-full flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors cursor-pointer mb-0.5 ${
                  tab === id
                    ? "bg-elevated text-txt"
                    : "text-txt-g hover:text-txt-2 hover:bg-elevated/50"
                }`}
              >
                <Icon className="w-4 h-4" />
                {label}
              </button>
            ))}
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto p-5">
            {/* General */}
            {tab === "general" && (
              <div className="space-y-6">
                {/* Theme */}
                <div>
                  <div className="flex items-center gap-2 mb-3">
                    <Palette className="w-4 h-4 text-txt-g" />
                    <span className="text-sm font-medium text-txt-2">
                      {t("theme")}
                    </span>
                  </div>
                  <div className="flex gap-2">
                    {(["dark", "light"] as const).map((th) => (
                      <button
                        key={th}
                        onClick={() => setTheme(th)}
                        className={`px-4 py-2 rounded-md text-sm transition-colors cursor-pointer ${
                          theme === th
                            ? "bg-accent/20 text-accent border border-accent/30"
                            : "bg-elevated text-txt-2 border border-bdr hover:border-bdr"
                        }`}
                      >
                        {th === "dark" ? t("dark") : t("light")}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Language */}
                <div>
                  <div className="flex items-center gap-2 mb-3">
                    <Globe className="w-4 h-4 text-txt-g" />
                    <span className="text-sm font-medium text-txt-2">
                      {t("language")}
                    </span>
                  </div>
                  <div className="flex gap-2">
                    {(["zh", "en"] as const).map((lang) => (
                      <button
                        key={lang}
                        onClick={() => setLanguage(lang)}
                        className={`px-4 py-2 rounded-md text-sm transition-colors cursor-pointer ${
                          language === lang
                            ? "bg-accent/20 text-accent border border-accent/30"
                            : "bg-elevated text-txt-2 border border-bdr hover:border-bdr"
                        }`}
                      >
                        {lang === "zh" ? t("chinese") : t("english")}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Font Size */}
                <div>
                  <div className="flex items-center gap-2 mb-3">
                    <Type className="w-4 h-4 text-txt-g" />
                    <span className="text-sm font-medium text-txt-2">
                      {t("font_size")}
                    </span>
                    <span className="text-xs text-txt-m ml-1">
                      {fontSize}px
                    </span>
                  </div>
                  <div className="flex items-center gap-3">
                    <button
                      onClick={() => setFontSize(Math.max(12, fontSize - 1))}
                      className="w-8 h-8 rounded-md bg-elevated text-txt-2 hover:bg-elevated/80 transition-colors cursor-pointer text-lg font-mono"
                    >
                      -
                    </button>
                    <input
                      type="range"
                      min={12}
                      max={20}
                      value={fontSize}
                      onChange={(e) => setFontSize(Number(e.target.value))}
                      className="flex-1 accent-accent"
                    />
                    <button
                      onClick={() => setFontSize(Math.min(20, fontSize + 1))}
                      className="w-8 h-8 rounded-md bg-elevated text-txt-2 hover:bg-elevated/80 transition-colors cursor-pointer text-lg font-mono"
                    >
                      +
                    </button>
                  </div>
                </div>

                {/* About */}
                <div className="border-t border-bdr-sub pt-4">
                  <div className="flex items-center gap-2 mb-3">
                    <Info className="w-4 h-4 text-txt-g" />
                    <span className="text-sm font-medium text-txt-2">
                      {t("about")}
                    </span>
                  </div>
                  <div className="text-xs text-txt-g space-y-1">
                    <p>MiMo Desktop v0.1.0</p>
                    <p>{t("app_description")}</p>
                  </div>
                </div>
              </div>
            )}

            {/* Tools & MCP */}
            {tab === "tools" && (
              <div className="space-y-4">
                {/* Tool tabs */}
                <div className="flex items-center gap-1 mb-3">
                  <button
                    onClick={() => setToolTab("builtin")}
                    className={`px-3 py-1.5 rounded-md text-sm transition-colors cursor-pointer ${
                      toolTab === "builtin"
                        ? "bg-elevated text-txt"
                        : "text-txt-g hover:text-txt-2"
                    }`}
                  >
                    <Wrench className="w-3.5 h-3.5 inline mr-1.5" />
                    {t("builtin_tools")} ({builtinTools.length})
                  </button>
                  <button
                    onClick={() => setToolTab("mcp")}
                    className={`px-3 py-1.5 rounded-md text-sm transition-colors cursor-pointer ${
                      toolTab === "mcp"
                        ? "bg-elevated text-txt"
                        : "text-txt-g hover:text-txt-2"
                    }`}
                  >
                    <Server className="w-3.5 h-3.5 inline mr-1.5" />
                    {t("mcp_tools")} ({mcpTools.length})
                  </button>
                </div>

                {/* Search */}
                <div className="relative mb-3">
                  <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-txt-m" />
                  <input
                    value={toolSearch}
                    onChange={(e) => setToolSearch(e.target.value)}
                    placeholder={t("search_tools")}
                    className="w-full bg-elevated border border-bdr rounded-md pl-9 pr-3 py-1.5 text-sm text-txt placeholder:text-txt-m focus:outline-none focus:border-accent/50"
                  />
                </div>

                {/* MCP Servers */}
                {toolTab === "mcp" && servers.length > 0 && (
                  <div className="space-y-2 mb-4">
                    {servers.map((s) => (
                      <div
                        key={s.name}
                        className="flex items-center gap-2 px-3 py-2 rounded-md bg-elevated/50 border border-bdr"
                      >
                        <Server className="w-4 h-4 text-blue-400" />
                        <span className="text-sm text-txt-2">{s.name}</span>
                        <span
                          className={`text-[10px] px-1.5 py-0.5 rounded ${
                            s.connected
                              ? "bg-green-500/10 text-green-400"
                              : "bg-red-500/10 text-red-400"
                          }`}
                        >
                          {s.connected ? t("connected") : t("disconnected")}
                        </span>
                        <span className="text-xs text-txt-m ml-auto">
                          {s.toolCount} {t("tools_suffix")}
                        </span>
                      </div>
                    ))}
                  </div>
                )}

                {/* Tool grid */}
                <div className="grid grid-cols-2 gap-2">
                  {(toolTab === "builtin" ? filteredBuiltin : filteredMcp).map(
                    (tool) => (
                      <div
                        key={tool.name}
                        className="border border-bdr-sub rounded-lg p-3 hover:border-bdr transition-colors"
                      >
                        <div className="flex items-center gap-2 mb-1">
                          <span className="font-mono text-sm text-txt-2 truncate">
                            {tool.name}
                          </span>
                          <span className="ml-auto flex-shrink-0">
                            <SafetyBadge level={tool.safetyLevel} />
                          </span>
                        </div>
                        <p className="text-xs text-txt-g line-clamp-2">
                          {td(tool.name)}
                        </p>
                      </div>
                    )
                  )}
                </div>
              </div>
            )}

            {/* Models */}
            {tab === "models" && (
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-txt-2">
                    {t("current_model")}:{" "}
                    <span className="text-accent font-mono">{currentModel}</span>
                  </span>
                  <button
                    onClick={() => setShowAddModel(!showAddModel)}
                    className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm bg-accent/15 text-accent hover:bg-accent/25 transition-colors cursor-pointer"
                  >
                    <Plus className="w-3.5 h-3.5" />
                    {t("add_model")}
                  </button>
                </div>

                {/* Add model form */}
                {showAddModel && (
                  <div className="border border-bdr rounded-lg p-4 space-y-3">
                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <label className="text-xs text-txt-g mb-1 block">
                          {t("model_name")} *
                        </label>
                        <input
                          value={newModel.name}
                          onChange={(e) =>
                            setNewModel({ ...newModel, name: e.target.value })
                          }
                          placeholder="my-model"
                          className="w-full bg-elevated border border-bdr rounded-md px-3 py-1.5 text-sm text-txt focus:outline-none focus:border-accent/50"
                        />
                      </div>
                      <div>
                        <label className="text-xs text-txt-g mb-1 block">
                          {t("model_id")}
                        </label>
                        <input
                          value={newModel.model}
                          onChange={(e) =>
                            setNewModel({ ...newModel, model: e.target.value })
                          }
                          placeholder="gpt-4o"
                          className="w-full bg-elevated border border-bdr rounded-md px-3 py-1.5 text-sm text-txt focus:outline-none focus:border-accent/50"
                        />
                      </div>
                    </div>
                    <div>
                      <label className="text-xs text-txt-g mb-1 block">
                        {t("api_base")} *
                      </label>
                      <input
                        value={newModel.apiBase}
                        onChange={(e) =>
                          setNewModel({ ...newModel, apiBase: e.target.value })
                        }
                        placeholder="https://api.openai.com/v1"
                        className="w-full bg-elevated border border-bdr rounded-md px-3 py-1.5 text-sm text-txt focus:outline-none focus:border-accent/50"
                      />
                    </div>
                    <div>
                      <label className="text-xs text-txt-g mb-1 block">
                        {t("api_key")}
                      </label>
                      <input
                        type="password"
                        value={newModel.apiKey}
                        onChange={(e) =>
                          setNewModel({ ...newModel, apiKey: e.target.value })
                        }
                        placeholder="sk-..."
                        className="w-full bg-elevated border border-bdr rounded-md px-3 py-1.5 text-sm text-txt focus:outline-none focus:border-accent/50"
                      />
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <label className="text-xs text-txt-g mb-1 block">
                          {t("max_tokens")}
                        </label>
                        <input
                          type="number"
                          value={newModel.maxTokens}
                          onChange={(e) =>
                            setNewModel({
                              ...newModel,
                              maxTokens: Number(e.target.value),
                            })
                          }
                          className="w-full bg-elevated border border-bdr rounded-md px-3 py-1.5 text-sm text-txt focus:outline-none focus:border-accent/50"
                        />
                      </div>
                      <div>
                        <label className="text-xs text-txt-g mb-1 block">
                          {t("temperature")}
                        </label>
                        <input
                          type="number"
                          step="0.1"
                          min="0"
                          max="2"
                          value={newModel.temperature}
                          onChange={(e) =>
                            setNewModel({
                              ...newModel,
                              temperature: Number(e.target.value),
                            })
                          }
                          className="w-full bg-elevated border border-bdr rounded-md px-3 py-1.5 text-sm text-txt focus:outline-none focus:border-accent/50"
                        />
                      </div>
                    </div>
                    <div className="flex gap-2 justify-end">
                      <button
                        onClick={() => setShowAddModel(false)}
                        className="px-3 py-1.5 rounded-md text-sm bg-elevated text-txt-2 hover:bg-elevated/80 cursor-pointer"
                      >
                        {t("cancel")}
                      </button>
                      <button
                        onClick={handleAddModel}
                        className="px-3 py-1.5 rounded-md text-sm bg-accent/20 text-accent hover:bg-accent/30 cursor-pointer"
                      >
                        {t("save")}
                      </button>
                    </div>
                  </div>
                )}

                {/* Model list */}
                <div className="space-y-2">
                  {Object.entries(models).map(([name, cfg]) => (
                    <div
                      key={name}
                      className={`flex items-center gap-3 px-3 py-2.5 rounded-lg border transition-colors ${
                        currentModel === name
                          ? "border-accent/30 bg-accent/5"
                          : "border-bdr-sub bg-elevated/30"
                      }`}
                    >
                      <Cpu className="w-4 h-4 text-txt-g flex-shrink-0" />
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="font-mono text-sm text-txt-2">
                            {name}
                          </span>
                          {currentModel === name && (
                            <span className="text-[10px] px-1.5 py-0.5 rounded bg-accent/20 text-accent">
                              {t("default_badge")}
                            </span>
                          )}
                        </div>
                        <div className="text-xs text-txt-m truncate">
                          {cfg.model} / {cfg.apiBase}
                        </div>
                      </div>
                      {currentModel !== name && (
                        <button
                          onClick={() => handleSetDefault(name)}
                          className="text-xs text-txt-g hover:text-accent transition-colors cursor-pointer px-2 py-1 rounded hover:bg-elevated"
                        >
                          {t("set_default")}
                        </button>
                      )}
                      {currentModel !== name && name !== "mimo" && (
                        <button
                          onClick={() => handleRemoveModel(name)}
                          className="p-1 text-txt-m hover:text-red-400 transition-colors cursor-pointer"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Help & Feedback */}
            {tab === "help" && (
              <div className="space-y-6">
                {/* Shortcuts */}
                <div>
                  <div className="flex items-center gap-2 mb-3">
                    <Keyboard className="w-4 h-4 text-txt-g" />
                    <span className="text-sm font-medium text-txt-2">
                      {t("help_shortcuts")}
                    </span>
                  </div>
                  <div className="space-y-1.5">
                    {[
                      ["Enter", t("send_message")],
                      ["Shift+Enter", t("new_line")],
                      ["Escape", t("cancel_close")],
                      ["Ctrl+N", t("new_chat")],
                      ["Ctrl+B", t("toggle_left")],
                      ["Ctrl+I", t("toggle_right")],
                    ].map(([key, desc]) => (
                      <div
                        key={key}
                        className="flex items-center gap-3 text-sm"
                      >
                        <kbd className="px-2 py-0.5 rounded bg-elevated border border-bdr text-txt-2 font-mono text-xs min-w-[90px] text-center">
                          {key}
                        </kbd>
                        <span className="text-txt-g">{desc}</span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Links */}
                <div>
                  <div className="flex items-center gap-2 mb-3">
                    <ExternalLink className="w-4 h-4 text-txt-g" />
                    <span className="text-sm font-medium text-txt-2">
                      {t("links")}
                    </span>
                  </div>
                  <div className="space-y-2">
                    <a
                      href="https://github.com/mimo-cli/mimo-cli"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-2 text-sm text-txt-g hover:text-accent transition-colors"
                    >
                      <ExternalLink className="w-3.5 h-3.5" />
                      {t("help_docs")}
                    </a>
                    <a
                      href="https://github.com/mimo-cli/mimo-cli/issues"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-2 text-sm text-txt-g hover:text-accent transition-colors"
                    >
                      <ExternalLink className="w-3.5 h-3.5" />
                      {t("help_report")}
                    </a>
                  </div>
                </div>

                {/* About */}
                <div className="border-t border-bdr-sub pt-4">
                  <div className="flex items-center gap-2 mb-2">
                    <Info className="w-4 h-4 text-txt-g" />
                    <span className="text-sm font-medium text-txt-2">
                      {t("about")}
                    </span>
                  </div>
                  <div className="text-xs text-txt-g space-y-1">
                    <p>MiMo Desktop v0.1.0</p>
                    <p>
                      {t("app_description")}
                    </p>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
