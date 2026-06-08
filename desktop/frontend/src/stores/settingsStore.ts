import { create } from "zustand";

interface SettingsState {
  theme: "dark" | "light";
  language: "zh" | "en";
  fontSize: number; // px, default 14

  setTheme: (theme: "dark" | "light") => void;
  setLanguage: (lang: "zh" | "en") => void;
  setFontSize: (size: number) => void;
  initFromConfig: (cfg: {
    theme?: string;
    language?: string;
  }) => void;
}

export const useSettingsStore = create<SettingsState>((set) => ({
  theme: "dark",
  language: "zh",
  fontSize: 14,

  setTheme: (theme) => {
    set({ theme });
    document.documentElement.classList.toggle("dark", theme === "dark");
    document.documentElement.style.colorScheme = theme;
    window.go?.desktop?.App?.SetTheme?.(theme).catch(console.error);
  },

  setLanguage: (language) => {
    set({ language });
    window.go?.desktop?.App?.SetLanguage?.(language).catch(console.error);
  },

  setFontSize: (fontSize) => {
    set({ fontSize });
    document.documentElement.style.fontSize = `${fontSize}px`;
  },

  initFromConfig: (cfg) => {
    const theme = (cfg.theme === "light" ? "light" : "dark") as "dark" | "light";
    const language = (cfg.language?.startsWith("zh") ? "zh" : "en") as "zh" | "en";
    set({ theme, language });
    document.documentElement.classList.toggle("dark", theme === "dark");
    document.documentElement.style.colorScheme = theme;
  },
}));
