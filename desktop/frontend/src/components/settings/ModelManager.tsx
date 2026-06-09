import { useState, useCallback, useMemo } from "react";
import {
  Plus,
  Trash2,
  Cpu,
  ExternalLink,
  Edit3,
  Save,
  X,
  Globe,
  Key,
  Server,
  Zap,
  Eye,
  EyeOff,
  Wrench,
  RefreshCw,
  Loader2,
  Check,
  Search,
  ChevronDown,
  ChevronUp,
  Star,
  Copy,
  Info,
} from "lucide-react";
import { t } from "../../lib/i18n";

export interface ModelConfig {
  provider: string;
  website: string;
  apiBase: string;
  apiKey: string;
  model: string;
  models: string[];
  fallback: string;
  maxTokens: number;
  temperature: number;
  topP: number;
  streaming: boolean;
  vision: boolean;
  tools: boolean;
}

interface RemoteModel {
  id: string;
  owned_by?: string;
  description?: string;
  context_window?: number;
  max_output?: number;
  capabilities?: string[];
}

interface Props {
  models: Record<string, ModelConfig>;
  currentModel: string;
  onSetDefault: (name: string) => void;
  onRemove: (name: string) => void;
  onAdd: (name: string, config: ModelConfig) => void;
  onUpdate: (name: string, config: ModelConfig) => void;
}

const defaultModelConfig: ModelConfig = {
  provider: "",
  website: "",
  apiBase: "",
  apiKey: "",
  model: "",
  models: [],
  fallback: "",
  maxTokens: 32768,
  temperature: 0.3,
  topP: 0.95,
  streaming: true,
  vision: false,
  tools: true,
};

const providerPresets: Record<string, Partial<ModelConfig>> = {
  OpenAI: {
    provider: "OpenAI",
    website: "https://openai.com",
    apiBase: "https://api.openai.com/v1",
    streaming: true,
    vision: true,
    tools: true,
  },
  Anthropic: {
    provider: "Anthropic",
    website: "https://anthropic.com",
    apiBase: "https://api.anthropic.com/v1",
    streaming: true,
    vision: true,
    tools: true,
  },
  DeepSeek: {
    provider: "DeepSeek",
    website: "https://deepseek.com",
    apiBase: "https://api.deepseek.com/v1",
    streaming: true,
    vision: false,
    tools: true,
  },
  MiMo: {
    provider: "MiMo",
    website: "https://platform.xiaomimimo.com",
    apiBase: "https://api.xiaomimimo.com/v1",
    models: [
      "mimo-v2.5-pro",
      "mimo-v2-pro",
      "mimo-v2.5",
      "mimo-v2-omni",
      "mimo-v2-flash",
    ],
    model: "mimo-v2.5-pro",
    fallback: "mimo-v2-flash",
    maxTokens: 128000,
    streaming: true,
    vision: true,
    tools: true,
  },
  "Moonshot": {
    provider: "Moonshot",
    website: "https://moonshot.cn",
    apiBase: "https://api.moonshot.cn/v1",
    streaming: true,
    vision: false,
    tools: true,
  },
  "Zhipu": {
    provider: "Zhipu",
    website: "https://open.bigmodel.cn",
    apiBase: "https://open.bigmodel.cn/api/paas/v4",
    streaming: true,
    vision: true,
    tools: true,
  },
};

// Helper to format token count
function formatTokens(count: number): string {
  if (count >= 1000000) return `${(count / 1000000).toFixed(0)}M`;
  if (count >= 1000) return `${(count / 1000).toFixed(0)}K`;
  return String(count);
}

// Capability display config
const capabilityConfig: Record<string, { label: string; color: string; icon: string }> = {
  text_generation: { label: "Text", color: "bg-blue-500/10 text-blue-400", icon: "T" },
  deep_thinking: { label: "Think", color: "bg-purple-500/10 text-purple-400", icon: "💡" },
  streaming: { label: "Stream", color: "bg-green-500/10 text-green-400", icon: "⚡" },
  function_calling: { label: "Tools", color: "bg-orange-500/10 text-orange-400", icon: "🔧" },
  structured_output: { label: "JSON", color: "bg-cyan-500/10 text-cyan-400", icon: "{ }" },
  web_search: { label: "Search", color: "bg-yellow-500/10 text-yellow-400", icon: "🔍" },
  multimodal: { label: "Multi", color: "bg-pink-500/10 text-pink-400", icon: "🖼" },
  speech_recognition: { label: "ASR", color: "bg-indigo-500/10 text-indigo-400", icon: "🎤" },
  speech_synthesis: { label: "TTS", color: "bg-teal-500/10 text-teal-400", icon: "🔊" },
  voice_clone: { label: "Clone", color: "bg-rose-500/10 text-rose-400", icon: "👤" },
  voice_design: { label: "Design", color: "bg-violet-500/10 text-violet-400", icon: "🎨" },
};

// Tooltip component with hover
function Tip({ text }: { text: string }) {
  return (
    <span className="relative inline-flex group/tip ml-1">
      <Info className="w-3 h-3 text-txt-m cursor-help" />
      <span className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-2.5 py-1.5 text-[10px] text-txt bg-elevated border border-bdr rounded-md shadow-lg opacity-0 group-hover/tip:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-50">
        {text}
        <span className="absolute top-full left-1/2 -translate-x-1/2 -mt-px border-4 border-transparent border-t-bdr" />
      </span>
    </span>
  );
}

export function ModelManager({
  models,
  currentModel,
  onSetDefault,
  onRemove,
  onAdd,
  onUpdate,
}: Props) {
  const safeModels = models || {};
  const [editingModel, setEditingModel] = useState<string | null>(null);
  const [showAddForm, setShowAddForm] = useState(false);
  const [newModelName, setNewModelName] = useState("");
  const [config, setConfig] = useState<ModelConfig>({ ...defaultModelConfig });
  const [showApiKey, setShowApiKey] = useState(false);
  const [remoteModels, setRemoteModels] = useState<RemoteModel[]>([]);
  const [loadingModels, setLoadingModels] = useState(false);
  const [loadingModelError, setLoadingModelError] = useState<string | null>(null);
  const [modelSearch, setModelSearch] = useState("");
  const [showAdvanced, setShowAdvanced] = useState(false);

  // Filtered remote models
  const filteredRemoteModels = useMemo(() => {
    if (!modelSearch) return remoteModels;
    const search = modelSearch.toLowerCase();
    return remoteModels.filter(
      (m) =>
        m.id.toLowerCase().includes(search) ||
        (m.owned_by && m.owned_by.toLowerCase().includes(search))
    );
  }, [remoteModels, modelSearch]);

  // Group remote models by owner
  const groupedRemoteModels = useMemo(() => {
    const groups: Record<string, RemoteModel[]> = {};
    filteredRemoteModels.forEach((m) => {
      const group = m.owned_by || "Other";
      if (!groups[group]) groups[group] = [];
      groups[group].push(m);
    });
    return groups;
  }, [filteredRemoteModels]);

  const resetForm = () => {
    setConfig({ ...defaultModelConfig });
    setNewModelName("");
    setShowAddForm(false);
    setEditingModel(null);
    setShowApiKey(false);
    setRemoteModels([]);
    setLoadingModelError(null);
    setModelSearch("");
    setShowAdvanced(false);
  };

  const handleEdit = (name: string) => {
    const model = safeModels[name];
    if (model) {
      setEditingModel(name);
      setConfig({ ...model });
      setShowAddForm(false);
      setShowApiKey(false);
    }
  };

  const handleSave = () => {
    if (!newModelName && !editingModel) return;
    if (editingModel) {
      onUpdate(editingModel, config);
    } else {
      onAdd(newModelName, config);
    }
    resetForm();
  };

  const handlePresetSelect = (presetName: string) => {
    const preset = providerPresets[presetName];
    if (preset) {
      setConfig((prev) => ({
        ...prev,
        ...preset,
        provider: presetName,
      }));
      setNewModelName(presetName);
    }
  };

  const handleFetchModels = useCallback(async () => {
    if (!config.apiBase || !config.apiKey) {
      setLoadingModelError(t("api_required_for_models"));
      return;
    }

    setLoadingModels(true);
    setLoadingModelError(null);
    setRemoteModels([]);

    try {
      const result = await window.go?.desktop?.App?.ListRemoteModelsWithConfig?.(
        config.apiBase,
        config.apiKey
      );
      if (result && Array.isArray(result)) {
        setRemoteModels(result);
        if (!config.model && result.length > 0) {
          setConfig((prev) => ({ ...prev, model: result[0].id }));
        }
      } else {
        setLoadingModelError(t("fetch_models_failed"));
      }
    } catch (err: any) {
      console.error("Failed to fetch models:", err);
      const errMsg = err?.message || err?.toString() || "";
      setLoadingModelError(errMsg ? `${t("fetch_models_failed")}: ${errMsg}` : t("fetch_models_failed"));
    } finally {
      setLoadingModels(false);
    }
  }, [config.apiBase, config.apiKey, config.model]);

  const handleAddModel = (modelId: string) => {
    if (!config.models.includes(modelId)) {
      setConfig((prev) => ({
        ...prev,
        models: [...prev.models, modelId],
        model: prev.model || modelId,
      }));
    }
  };

  const handleRemoveModel = (modelId: string) => {
    setConfig((prev) => ({
      ...prev,
      models: prev.models.filter((m) => m !== modelId),
      model: prev.model === modelId ? (prev.models.find((m) => m !== modelId) || "") : prev.model,
    }));
  };

  const handleSetDefault = (modelId: string) => {
    setConfig((prev) => ({ ...prev, model: modelId }));
  };

  const isEditing = editingModel !== null;
  const formTitle = isEditing ? `${t("edit_model")}: ${editingModel}` : t("add_model");

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-sm font-medium text-txt">{t("model_management")}</h3>
          <p className="text-xs text-txt-g mt-0.5">
            {Object.keys(safeModels).length} {t("models_configured")}
          </p>
        </div>
        <button
          onClick={() => {
            resetForm();
            setShowAddForm(true);
          }}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm bg-accent text-white hover:bg-accent/90 transition-colors cursor-pointer"
        >
          <Plus className="w-3.5 h-3.5" />
          {t("add_model")}
        </button>
      </div>

      {/* Add/Edit Form */}
      {(showAddForm || editingModel) && (
        <div className="bg-surface border border-bdr rounded-lg overflow-hidden">
          {/* Form Header */}
          <div className="flex items-center justify-between px-4 py-3 border-b border-bdr-div bg-elevated/50">
            <h3 className="text-sm font-medium text-txt">{formTitle}</h3>
            <button
              onClick={resetForm}
              className="p-1 text-txt-g hover:text-txt cursor-pointer rounded hover:bg-elevated"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          <div className="p-4 space-y-4">
            {/* Quick Setup Presets */}
            {!isEditing && (
              <div>
                <label className="text-xs font-medium text-txt-2 mb-2 block">{t("quick_setup")}</label>
                <div className="grid grid-cols-3 gap-2">
                  {Object.entries(providerPresets).map(([name, preset]) => (
                    <button
                      key={name}
                      onClick={() => handlePresetSelect(name)}
                      className="flex items-center gap-2 px-3 py-2 text-xs rounded-md border border-bdr hover:border-accent/50 hover:bg-accent/5 transition-colors cursor-pointer text-left"
                    >
                      <div className="w-6 h-6 rounded bg-accent/10 flex items-center justify-center flex-shrink-0">
                        <Cpu className="w-3.5 h-3.5 text-accent" />
                      </div>
                      <div>
                        <div className="font-medium text-txt">{name}</div>
                        <div className="text-[10px] text-txt-g truncate">{preset.apiBase?.replace("https://", "").split("/")[0]}</div>
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Basic Info */}
            <div className="space-y-3">
              <div className="flex items-center gap-2 text-xs font-medium text-txt-2">
                <Server className="w-3.5 h-3.5" />
                {t("api_settings")}
              </div>
              
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs text-txt-g mb-1 block">{t("provider_name")}</label>
                  <input
                    value={newModelName || config.provider}
                    onChange={(e) => {
                      setNewModelName(e.target.value);
                      setConfig((prev) => ({ ...prev, provider: e.target.value }));
                    }}
                    placeholder="My Provider"
                    disabled={isEditing}
                    className="w-full bg-elevated border border-bdr rounded-md px-3 py-2 text-sm text-txt focus:outline-none focus:border-accent/50 disabled:opacity-50"
                  />
                </div>
                <div>
                  <label className="text-xs text-txt-g mb-1 flex items-center gap-1">
                    <Globe className="w-3 h-3" />
                    {t("website")}
                  </label>
                  <input
                    value={config.website}
                    onChange={(e) => setConfig({ ...config, website: e.target.value })}
                    placeholder="https://example.com"
                    className="w-full bg-elevated border border-bdr rounded-md px-3 py-2 text-sm text-txt focus:outline-none focus:border-accent/50"
                  />
                </div>
              </div>

              <div>
                <label className="text-xs text-txt-g mb-1 block">{t("api_base_url")} *</label>
                <input
                  value={config.apiBase}
                  onChange={(e) => setConfig({ ...config, apiBase: e.target.value })}
                  placeholder="https://api.openai.com/v1"
                  className="w-full bg-elevated border border-bdr rounded-md px-3 py-2 text-sm text-txt font-mono focus:outline-none focus:border-accent/50"
                />
              </div>

              <div>
                <label className="text-xs text-txt-g mb-1 flex items-center gap-1">
                  <Key className="w-3 h-3" />
                  {t("api_key")} *
                </label>
                <div className="relative">
                  <input
                    type={showApiKey ? "text" : "password"}
                    value={config.apiKey}
                    onChange={(e) => setConfig({ ...config, apiKey: e.target.value })}
                    placeholder="sk-..."
                    className="w-full bg-elevated border border-bdr rounded-md px-3 py-2 pr-10 text-sm text-txt font-mono focus:outline-none focus:border-accent/50"
                  />
                  <button
                    type="button"
                    onClick={() => setShowApiKey(!showApiKey)}
                    className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-txt-g hover:text-txt cursor-pointer"
                    title={showApiKey ? t("hide_api_key") : t("show_api_key")}
                  >
                    {showApiKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
              </div>
            </div>

            {/* Fetch Models */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label className="text-xs font-medium text-txt-2">{t("available_models")}</label>
                <button
                  onClick={handleFetchModels}
                  disabled={loadingModels || !config.apiBase || !config.apiKey}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-accent bg-accent/10 hover:bg-accent/20 rounded-md transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {loadingModels ? (
                    <>
                      <Loader2 className="w-3 h-3 animate-spin" />
                      {t("fetching")}
                    </>
                  ) : (
                    <>
                      <RefreshCw className="w-3 h-3" />
                      {t("fetch_models")}
                    </>
                  )}
                </button>
              </div>

              {loadingModelError && (
                <div className="text-xs text-red-400 bg-red-500/10 border border-red-500/20 px-3 py-2 rounded-md">
                  {loadingModelError}
                </div>
              )}

              {remoteModels.length > 0 && (
                <div className="space-y-2">
                  {/* Search */}
                  <div className="relative">
                    <Search className="w-3.5 h-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-txt-m" />
                    <input
                      value={modelSearch}
                      onChange={(e) => setModelSearch(e.target.value)}
                      placeholder={`${t("search_models")} (${remoteModels.length})`}
                      className="w-full bg-elevated border border-bdr rounded-md pl-8 pr-3 py-1.5 text-xs text-txt focus:outline-none focus:border-accent/50"
                    />
                  </div>

                  {/* Model Groups */}
                  <div className="max-h-48 overflow-y-auto border border-bdr rounded-md divide-y divide-bdr-div">
                    {Object.entries(groupedRemoteModels).map(([group, models]) => (
                      <div key={group}>
                        <div className="px-2.5 py-1.5 text-[10px] font-medium text-txt-g uppercase tracking-wider bg-elevated/50">
                          {group}
                        </div>
                        <div className="divide-y divide-bdr-div">
                          {models.map((m) => {
                            const isSelected = config.models.includes(m.id);
                            const isDefault = config.model === m.id;
                            return (
                              <div
                                key={m.id}
                                className={`px-2.5 py-2 cursor-pointer hover:bg-elevated/50 transition-colors ${
                                  isSelected ? "bg-accent/5" : ""
                                }`}
                                onClick={() => isSelected ? handleRemoveModel(m.id) : handleAddModel(m.id)}
                              >
                                <div className="flex items-center gap-2">
                                  <div className={`w-4 h-4 rounded border flex items-center justify-center flex-shrink-0 ${
                                    isSelected ? "bg-accent border-accent text-white" : "border-bdr"
                                  }`}>
                                    {isSelected && <Check className="w-3 h-3" />}
                                  </div>
                                  <span className={`flex-1 font-mono text-xs ${isSelected ? "text-txt" : "text-txt-2"}`}>
                                    {m.id}
                                  </span>
                                  {isDefault && (
                                    <Star className="w-3 h-3 text-yellow-500 fill-yellow-500 flex-shrink-0" />
                                  )}
                                </div>
                                
                                {/* Model details */}
                                {(m.description || m.context_window || m.capabilities?.length) && (
                                  <div className="ml-6 mt-1.5 space-y-1">
                                    {m.description && (
                                      <p className="text-[10px] text-txt-g">{m.description}</p>
                                    )}
                                    
                                    {/* Context and output limits */}
                                    {(m.context_window || m.max_output) && (
                                      <div className="flex items-center gap-2 text-[10px] text-txt-m">
                                        {m.context_window && (
                                          <span>上下文: {formatTokens(m.context_window)}</span>
                                        )}
                                        {m.max_output && (
                                          <span>输出: {formatTokens(m.max_output)}</span>
                                        )}
                                      </div>
                                    )}
                                    
                                    {/* Capability badges */}
                                    {m.capabilities && m.capabilities.length > 0 && (
                                      <div className="flex flex-wrap gap-1">
                                        {m.capabilities.map((cap) => {
                                          const config = capabilityConfig[cap];
                                          if (!config) return null;
                                          return (
                                            <span
                                              key={cap}
                                              className={`inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[9px] ${config.color}`}
                                            >
                                              {config.label}
                                            </span>
                                          );
                                        })}
                                      </div>
                                    )}
                                  </div>
                                )}
                              </div>
                            );
                          })}
                        </div>
                      </div>
                    ))}
                  </div>

                  <div className="text-[10px] text-txt-g">
                    {t("selected")}: {config.models.length} / {remoteModels.length}
                  </div>
                </div>
              )}
            </div>

            {/* Selected Models */}
            {config.models.length > 0 && (
              <div className="space-y-2">
                <label className="text-xs font-medium text-txt-2">
                  {t("selected_models")} ({config.models.length})
                </label>
                <div className="flex flex-wrap gap-1.5">
                  {config.models.map((m) => (
                    <span
                      key={m}
                      className={`inline-flex items-center gap-1 px-2 py-1 text-xs rounded-md cursor-pointer transition-colors ${
                        config.model === m
                          ? "bg-accent text-white"
                          : "bg-elevated text-txt-2 hover:bg-elevated/80"
                      }`}
                      onClick={() => handleSetDefault(m)}
                    >
                      {config.model === m && <Star className="w-3 h-3" />}
                      {m}
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleRemoveModel(m);
                        }}
                        className="ml-0.5 hover:text-red-400"
                      >
                        <X className="w-3 h-3" />
                      </button>
                    </span>
                  ))}
                </div>
                <p className="text-[10px] text-txt-g">{t("click_to_set_default")}</p>
              </div>
            )}

            {/* Advanced Settings */}
            <div>
              <button
                onClick={() => setShowAdvanced(!showAdvanced)}
                className="flex items-center gap-1.5 text-xs font-medium text-txt-2 hover:text-txt cursor-pointer"
              >
                {showAdvanced ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
                {t("advanced_settings")}
              </button>

              {showAdvanced && (
                <div className="mt-3 space-y-4 pl-5 border-l-2 border-bdr-div">
                  {/* Fallback Model */}
                  <div>
                    <label className="text-xs text-txt-g mb-1 block">{t("fallback_model")}<Tip text={t("fallback_tip")} /></label>
                    <input
                      value={config.fallback}
                      onChange={(e) => setConfig({ ...config, fallback: e.target.value })}
                      placeholder={t("optional")}
                      className="w-full bg-elevated border border-bdr rounded-md px-3 py-1.5 text-sm text-txt focus:outline-none focus:border-accent/50"
                    />
                  </div>

                  {/* Generation Params */}
                  <div className="space-y-3">
                    <div className="flex items-center gap-2 text-xs font-medium text-txt-2">
                      <Zap className="w-3.5 h-3.5" />
                      {t("generation_params")}
                    </div>
                    <div className="grid grid-cols-3 gap-3">
                      <div>
                        <label className="text-xs text-txt-g mb-1 block">{t("max_tokens")}<Tip text={t("max_tokens_tip")} /></label>
                        <input
                          type="number"
                          value={config.maxTokens}
                          onChange={(e) => setConfig({ ...config, maxTokens: Number(e.target.value) })}
                          className="w-full bg-elevated border border-bdr rounded-md px-3 py-1.5 text-sm text-txt focus:outline-none focus:border-accent/50"
                        />
                      </div>
                      <div>
                        <label className="text-xs text-txt-g mb-1 block">Temperature<Tip text={t("temperature_tip")} /></label>
                        <input
                          type="number"
                          step="0.1"
                          min="0"
                          max="2"
                          value={config.temperature}
                          onChange={(e) => setConfig({ ...config, temperature: Number(e.target.value) })}
                          className="w-full bg-elevated border border-bdr rounded-md px-3 py-1.5 text-sm text-txt focus:outline-none focus:border-accent/50"
                        />
                      </div>
                      <div>
                        <label className="text-xs text-txt-g mb-1 block">Top P<Tip text={t("top_p_tip")} /></label>
                        <input
                          type="number"
                          step="0.05"
                          min="0"
                          max="1"
                          value={config.topP}
                          onChange={(e) => setConfig({ ...config, topP: Number(e.target.value) })}
                          className="w-full bg-elevated border border-bdr rounded-md px-3 py-1.5 text-sm text-txt focus:outline-none focus:border-accent/50"
                        />
                      </div>
                    </div>
                  </div>

                  {/* Features */}
                  <div className="space-y-2">
                    <div className="flex items-center gap-2 text-xs font-medium text-txt-2">
                      <Wrench className="w-3.5 h-3.5" />
                      {t("supported_features")}
                    </div>
                    <div className="flex gap-4">
                      {[
                        { key: "streaming", label: t("streaming") },
                        { key: "vision", label: t("vision") },
                        { key: "tools", label: t("function_calling") },
                      ].map(({ key, label }) => (
                        <label key={key} className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            checked={config[key as keyof ModelConfig] as boolean}
                            onChange={(e) =>
                              setConfig({ ...config, [key]: e.target.checked })
                            }
                            className="accent-accent rounded"
                          />
                          <span className="text-xs text-txt-2">{label}</span>
                        </label>
                      ))}
                    </div>
                  </div>
                </div>
              )}
            </div>

            {/* Actions */}
            <div className="flex gap-2 justify-end pt-3 border-t border-bdr-div">
              <button
                onClick={resetForm}
                className="px-4 py-2 rounded-md text-sm text-txt-2 hover:text-txt hover:bg-elevated transition-colors cursor-pointer"
              >
                {t("cancel")}
              </button>
              <button
                onClick={handleSave}
                disabled={!config.apiBase || config.models.length === 0}
                className="flex items-center gap-1.5 px-4 py-2 rounded-md text-sm bg-accent text-white hover:bg-accent/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors cursor-pointer"
              >
                <Save className="w-3.5 h-3.5" />
                {isEditing ? t("update") : t("save")}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Model List */}
      <div className="space-y-2">
        {Object.keys(safeModels).length === 0 ? (
          <div className="text-center py-8 text-txt-g">
            <Cpu className="w-8 h-8 mx-auto mb-2 opacity-50" />
            <p className="text-sm">{t("no_models_configured")}</p>
            <p className="text-xs mt-1">{t("click_add_model_to_start")}</p>
          </div>
        ) : (
          Object.entries(safeModels).map(([name, cfg]) => {
            const isActive = currentModel === name;
            return (
              <div
                key={name}
                className={`group flex items-center gap-3 px-4 py-3 rounded-lg border transition-all ${
                  isActive
                    ? "border-accent bg-accent/5 shadow-sm"
                    : "border-bdr hover:border-accent/30 bg-surface"
                }`}
              >
                {/* Icon */}
                <div
                  className={`w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 ${
                    isActive ? "bg-accent/20" : "bg-elevated"
                  }`}
                >
                  <Cpu className={`w-5 h-5 ${isActive ? "text-accent" : "text-txt-g"}`} />
                </div>

                {/* Info */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className={`font-medium text-sm ${isActive ? "text-accent" : "text-txt"}`}>
                      {name}
                    </span>
                    {cfg.provider && (
                      <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-elevated text-txt-m">
                        {cfg.provider}
                      </span>
                    )}
                    {isActive && (
                      <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-accent text-white">
                        {t("active")}
                      </span>
                    )}
                  </div>
                  <div className="text-xs text-txt-g mt-0.5 flex items-center gap-2">
                    <span className="font-mono truncate">{cfg.model}</span>
                    {cfg.models.length > 1 && (
                      <span className="text-txt-m">+{cfg.models.length - 1}</span>
                    )}
                  </div>
                  <div className="flex items-center gap-2 mt-1">
                    {cfg.streaming && (
                      <span className="text-[10px] px-1 py-0.5 rounded bg-green-500/10 text-green-400">
                        Stream
                      </span>
                    )}
                    {cfg.tools && (
                      <span className="text-[10px] px-1 py-0.5 rounded bg-blue-500/10 text-blue-400">
                        Tools
                      </span>
                    )}
                    {cfg.vision && (
                      <span className="text-[10px] px-1 py-0.5 rounded bg-purple-500/10 text-purple-400">
                        Vision
                      </span>
                    )}
                  </div>
                </div>

                {/* Actions */}
                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  {cfg.website && (
                    <a
                      href={cfg.website}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="p-1.5 text-txt-g hover:text-accent hover:bg-elevated rounded cursor-pointer"
                      title={t("visit_website")}
                    >
                      <ExternalLink className="w-3.5 h-3.5" />
                    </a>
                  )}
                  <button
                    onClick={() => handleEdit(name)}
                    className="p-1.5 text-txt-g hover:text-accent hover:bg-elevated rounded cursor-pointer"
                    title={t("edit")}
                  >
                    <Edit3 className="w-3.5 h-3.5" />
                  </button>
                  {!isActive && (
                    <button
                      onClick={() => onSetDefault(name)}
                      className="p-1.5 text-txt-g hover:text-yellow-500 hover:bg-elevated rounded cursor-pointer"
                      title={t("set_default")}
                    >
                      <Star className="w-3.5 h-3.5" />
                    </button>
                  )}
                  {!isActive && name !== "mimo" && (
                    <button
                      onClick={() => onRemove(name)}
                      className="p-1.5 text-txt-g hover:text-red-400 hover:bg-red-500/10 rounded cursor-pointer"
                      title={t("remove")}
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  )}
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
