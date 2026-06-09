import { animateThemeSwitch } from "../lib/theme-transition";
import { create } from "zustand";

interface SettingsState {
  theme: "dark" | "light";
  language: "zh" | "en";
  fontSize: number;
  currentModel: string;
  currentModelKey: string;
  models: string[];
  planningMode: string;
  safetyLevel: string;
  permission: string;
  reasoningLevel: string;

  setTheme: (theme: "dark" | "light", originX?: number, originY?: number) => void;
  setLanguage: (lang: "zh" | "en") => void;
  setFontSize: (size: number) => void;
  setCurrentModel: (key: string) => void;
  setPlanningMode: (mode: string) => void;
  setSafetyLevel: (level: string) => void;
  setPermission: (perm: string) => void;
  setReasoningLevel: (level: string) => void;
  initFromConfig: (cfg: {
    theme?: string;
    language?: string;
    defaultModel?: string;
    models?: Record<string, unknown>;
    planningMode?: string;
    safetyLevel?: string;
    permission?: string;
    reasoningLevel?: string;
  }) => void;
}

export const useSettingsStore = create<SettingsState>((set) => ({
  theme: "dark",
  language: "zh",
  fontSize: 14,
  currentModel: "mimo",
  currentModelKey: "mimo",
  models: [],
  planningMode: "auto",
  safetyLevel: "confirm",
  permission: "exec",
  reasoningLevel: "medium",

  setTheme: (theme, originX?, originY?) => {
    set({ theme });
    animateThemeSwitch(theme, originX, originY);
    window.go?.desktop?.App?.SetTheme?.(theme).catch(console.error);
  },

  setLanguage: (language) => {
    set({ language });
    window.go?.desktop?.App?.SetLanguage?.(language).catch(console.error);
  },

  setFontSize: (fontSize) => {
    set({ fontSize });
    document.documentElement.style.fontSize = fontSize + "px";
  },

  setCurrentModel: (key: string) => {
    const modelsMap = (useSettingsStore.getState() as any)._modelsMap as Record<string, { model?: string }> | undefined;
    const displayModel = modelsMap?.[key]?.model || key;
    set({ currentModel: displayModel, currentModelKey: key });
    window.go?.desktop?.App?.SetDefaultModel?.(key).catch(console.error);
  },

  setPlanningMode: (mode) => {
    set({ planningMode: mode });
    window.go?.desktop?.App?.SetPlanningMode?.(mode).catch(console.error);
  },

  setSafetyLevel: (level) => {
    set({ safetyLevel: level });
    window.go?.desktop?.App?.SetSafetyLevel?.(level).catch(console.error);
  },

  setPermission: (perm) => {
    set({ permission: perm });
    window.go?.desktop?.App?.SetPermission?.(perm).catch(console.error);
  },

  setReasoningLevel: (level) => {
    set({ reasoningLevel: level });
    window.go?.desktop?.App?.SetReasoningLevel?.(level).catch(console.error);
  },

  initFromConfig: (cfg) => {
    const theme = (cfg.theme === "light" ? "light" : "dark") as "dark" | "light";
    const language = (cfg.language?.startsWith("zh") ? "zh" : "en") as "zh" | "en";
    const modelKey = cfg.defaultModel || "";
    const modelsMap = (cfg.models || {}) as Record<string, { model?: string }>;
    let displayModel = modelKey;
    if (modelKey && modelsMap[modelKey]?.model) {
      displayModel = modelsMap[modelKey].model!;
    }
    set({
      theme, language,
      currentModel: displayModel,
      currentModelKey: modelKey,
      models: cfg.models ? Object.keys(cfg.models) : [],
      _modelsMap: modelsMap,
      planningMode: cfg.planningMode || "auto",
      safetyLevel: cfg.safetyLevel || "confirm",
      permission: cfg.permission || "exec",
      reasoningLevel: cfg.reasoningLevel || "medium",
    } as any);
    document.documentElement.classList.toggle("dark", theme === "dark");
    document.documentElement.style.colorScheme = theme;
    try { localStorage.setItem("mimo-theme", theme); } catch(e) {}
  },
}));
