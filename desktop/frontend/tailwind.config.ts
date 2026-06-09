/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  darkMode: "class",
  theme: {
    extend: {
      keyframes: {
        "fade-in": {
          "0%": { opacity: "0", transform: "translateY(8px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        "pop-up": {
          "0%": { opacity: "0", transform: "translateY(6px) scale(0.97)" },
          "100%": { opacity: "1", transform: "translateY(0) scale(1)" },
        },
        "slide-right": {
          "0%": { opacity: "0", transform: "translateX(-6px)" },
          "100%": { opacity: "1", transform: "translateX(0)" },
        },
        "modal-in": {
          "0%": { opacity: "0", transform: "scale(0.97) translateY(12px)" },
          "100%": { opacity: "1", transform: "scale(1) translateY(0)" },
        },
        "pop-out": {
          "0%": { opacity: "1", transform: "translateY(0) scale(1)" },
          "100%": { opacity: "0", transform: "translateY(4px) scale(0.97)" },
        },
        "slide-left": {
          "0%": { opacity: "1", transform: "translateX(0)" },
          "100%": { opacity: "0", transform: "translateX(-6px)" },
        },
        "backdrop-in": {
          "0%": { opacity: "0" },
          "100%": { opacity: "1" },
        },
        "backdrop-out": {
          "0%": { opacity: "1" },
          "100%": { opacity: "0" },
        },
        "modal-out": {
          "0%": { opacity: "1", transform: "scale(1) translateY(0)" },
          "100%": { opacity: "0", transform: "scale(0.97) translateY(8px)" },
        },
      },
      animation: {
        "fade-in": "fade-in 0.2s ease-out",
        "pop-up": "pop-up 0.15s ease-out",
        "slide-right": "slide-right 0.15s ease-out",
        "modal-in": "modal-in 0.25s cubic-bezier(0.16, 1, 0.3, 1)",
        "pop-out": "pop-out 0.12s ease-in forwards",
        "slide-left": "slide-left 0.12s ease-in forwards",
        "modal-out": "modal-out 0.2s ease-in forwards",
        "backdrop-in": "backdrop-in 0.2s ease-out",
        "backdrop-out": "backdrop-out 0.2s ease-in forwards",
      },
      colors: {
        accent: {
          DEFAULT: "#c48870",
          light: "#daa890",
          dark: "#a06b50",
        },
        root: "var(--bg-root)",
        surface: "var(--bg-surface)",
        sidebar: "var(--bg-sidebar)",
        elevated: "var(--bg-elevated)",
        bdr: "var(--border-default)",
        "bdr-sub": "var(--border-subtle)",
        "bdr-div": "var(--border-divider)",
        txt: "var(--text-primary)",
        "txt-2": "var(--text-secondary)",
        "txt-m": "var(--text-muted)",
        "txt-g": "var(--text-ghost)",
      },
    },
  },
  plugins: [require("@tailwindcss/typography")],
};
