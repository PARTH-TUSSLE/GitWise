import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        term: {
          bg: "#050505",
          panel: "#0c0c0c",
          border: "#262626",
          "border-bright": "#404040",
          text: "#fafafa",
          muted: "#8a8a8a",
          faint: "#525252",
          amber: "#fbbf24",
          "amber-dim": "#d97706",
          green: "#4ade80",
          red: "#f87171",
        },
      },
      fontFamily: {
        mono: [
          "ui-monospace",
          "SFMono-Regular",
          "Menlo",
          "Monaco",
          "Consolas",
          '"Liberation Mono"',
          '"Courier New"',
          "monospace",
        ],
      },
    },
  },
  plugins: [],
};
export default config;
