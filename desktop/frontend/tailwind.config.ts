/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        accent: {
          DEFAULT: "#be8367",
          light: "#d4a088",
          dark: "#a06b50",
        },
        root: "var(--bg-root)",
        surface: "var(--bg-surface)",
        elevated: "var(--bg-elevated)",
        bdr: "var(--border-default)",
        "bdr-sub": "var(--border-subtle)",
        txt: "var(--text-primary)",
        "txt-2": "var(--text-secondary)",
        "txt-m": "var(--text-muted)",
        "txt-g": "var(--text-ghost)",
      },
    },
  },
  plugins: [require("@tailwindcss/typography")],
};
