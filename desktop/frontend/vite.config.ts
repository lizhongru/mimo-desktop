import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: "dist",
    rollupOptions: {
      output: {
        manualChunks: {
          // Syntax highlighter is ~300KB, split into its own chunk
          "syntax-highlighter": [
            "react-syntax-highlighter",
            "react-syntax-highlighter/dist/esm/styles/prism",
          ],
          // React + React DOM
          "vendor-react": ["react", "react-dom", "react/jsx-runtime"],
          // Zustand stores
          "vendor-state": ["zustand"],
          // Lucide icons (tree-shaken but still large)
          "vendor-icons": ["lucide-react"],
        },
      },
    },
  },
});
