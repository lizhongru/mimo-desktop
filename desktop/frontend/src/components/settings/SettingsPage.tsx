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
import { ModelManager, type ModelConfig } from "./ModelManager";
import { AdvancedSettings, type AdvancedSettingsConfig } from "./AdvancedSettings";

type Tab = "general" | "advanced" | "tools" | "models" | "help";

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

const defaultAdvancedSettings: AdvancedSettingsConfig = {
  checkpoint: {
    autoCheckpoint: true,
    tokenThreshold: 0.75,
    maxCheckpoints: 10,
    reconstructOnResume: true,
    contextBudget: 128000,
  },
  memory: {
    ccIndex: true,
    searchScoreFloor: 0.15,
  },
  permission: {
    read: "allow",
    write: "ask",
    edit: "ask",
    bash: "ask",
  },
};

function permissionRulesToForm(
  rules?: Array<{ permission: string; action: string }>
): AdvancedSettingsConfig["permission"] {
  const result = { ...defaultAdvancedSettings.permission };
  for (const rule of rules || []) {
    if (
      rule.permission === "read" ||
      rule.permission === "write" ||
      rule.permission === "edit" ||
      rule.permission === "bash"
    ) {
      result[rule.permission] = rule.action;
    }
  }
  return result;
}

function permissionFormToRules(permission: AdvancedSettingsConfig["permission"]) {
  return Object.entries(permission).map(([key, action]) => ({
    permission: key,
    action,
  }));
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
  const [appVersion, setAppVersion] = useState("dev");

  // Models state
  const [models, setModels] = useState<Record<string, ModelConfig>>({});
  const [currentModel, setCurrentModel] = useState(defaultModel);
  const [advancedSettings, setAdvancedSettings] =
    useState<AdvancedSettingsConfig>(defaultAdvancedSettings);
  const [advancedSaving, setAdvancedSaving] = useState(false);


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
        setAdvancedSettings({
          checkpoint: cfg.checkpoint || defaultAdvancedSettings.checkpoint,
          memory: cfg.memory || defaultAdvancedSettings.memory,
          permission: permissionRulesToForm(cfg.permission?.rules),
        });
      })
      .catch(console.error);
    window.go?.desktop?.App?.GetVersion?.()
      .then((v) => setAppVersion(v.version || "dev"))
      .catch(console.error);
  }, [open]);



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



  const handleSetDefault = (name: string) => {
    window.go?.desktop?.App?.SetDefaultModel?.(name)
      .then(() => { setCurrentModel(name); useSettingsStore.getState().refreshModels(); })
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
      }).then(() => { useSettingsStore.getState().refreshModels(); }).catch(console.error);
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
    { id: "advanced", icon: SettingsIcon, label: "高级设置" },
    { id: "tools", icon: Wrench, label: t("tools_mcp") },
    { id: "models", icon: Cpu, label: t("models") },
    { id: "help", icon: HelpCircle, label: t("help_feedback") },
  ];

  return (
    <div className={`modal-overlay ${open ? "is-open" : ""}`} onClick={onClose}>
      <div className="modal-dialog !p-0 w-[780px] h-[600px] mx-4 flex flex-col overflow-hidden" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="flex items-center gap-3 px-5 py-3 border-b border-bdr-div flex-shrink-0">
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
                        onClick={(e) => setTheme(th, e.clientX, e.clientY)}
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
                    <p>MiMo Desktop v{appVersion}</p>
                    <p>{t("app_description")}</p>
                  </div>
                </div>
              </div>
            )}

            {/* Advanced Settings */}
            {tab === "advanced" && (
              <AdvancedSettings
                value={advancedSettings}
                saving={advancedSaving}
                onSave={(config) => {
                  setAdvancedSaving(true);
                  window.go?.desktop?.App?.UpdateAdvancedSettings?.({
                    checkpoint: config.checkpoint,
                    memory: config.memory,
                    permission: { rules: permissionFormToRules(config.permission) },
                  })
                    .then(() => setAdvancedSettings(config))
                    .catch(console.error)
                    .finally(() => setAdvancedSaving(false));
                }}
              />
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
              <ModelManager
                models={models}
                currentModel={currentModel}
                onSetDefault={handleSetDefault}
                onRemove={handleRemoveModel}
                onAdd={(name, cfg) => {
                  window.go?.desktop?.App?.AddModel?.(
                    name, cfg.provider, cfg.website, cfg.apiBase, cfg.apiKey,
                    cfg.model, cfg.models, cfg.fallback, cfg.maxTokens,
                    cfg.temperature, cfg.topP, cfg.streaming, cfg.vision, cfg.tools
                  ).then(() => {
                    setModels((prev) => ({ ...prev, [name]: cfg }));
                    useSettingsStore.getState().refreshModels();
                  }).catch(console.error);
                }}
                onUpdate={(name, cfg) => {
                  window.go?.desktop?.App?.UpdateModel?.(
                    name, cfg.provider, cfg.website, cfg.apiBase, cfg.apiKey,
                    cfg.model, cfg.models, cfg.fallback, cfg.maxTokens,
                    cfg.temperature, cfg.topP, cfg.streaming, cfg.vision, cfg.tools
                  ).then(() => {
                    setModels((prev) => ({ ...prev, [name]: cfg }));
                    useSettingsStore.getState().refreshModels();
                  }).catch(console.error);
                }}
              />
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
                    <p>MiMo Desktop v{appVersion}</p>
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
