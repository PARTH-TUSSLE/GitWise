"use client";

import React, { createContext, useContext, useEffect, useState } from "react";

export type ThemeId =
  | "crimson-cyan"
  | "amber-emerald"
  | "cobalt-orange"
  | "amethyst-gold"
  | "matrix-crimson"
  | "fuchsia-cyan"
  | "titanium-gold"
  | "rose"
  | "amber"
  | "emerald"
  | "cyan"
  | "violet"
  | "zinc";

export interface ThemeMeta {
  id: ThemeId;
  name: string;
  label: string;
  category: "contrast" | "monochrome";
  accentHex: string;
  accentSecondaryHex: string;
  dimHex: string;
  description: string;
}

export const THEMES: ThemeMeta[] = [
  // 1. Dual-Tone High-Contrast Presets
  {
    id: "crimson-cyan",
    name: "Crimson & Cyan",
    label: "Rust × Electric Cobalt",
    category: "contrast",
    accentHex: "#f43f5e",
    accentSecondaryHex: "#06b6d4",
    dimHex: "#e11d48",
    description: "High-contrast compiler diagnostic: crimson AST with electric cyan telemetry",
  },
  {
    id: "amber-emerald",
    name: "Amber & Emerald",
    label: "Phosphor × Matrix Green",
    category: "contrast",
    accentHex: "#fbbf24",
    accentSecondaryHex: "#10b981",
    dimHex: "#d97706",
    description: "Dual phosphor terminal: vintage amber CRT with green execution flow signals",
  },
  {
    id: "cobalt-orange",
    name: "Cobalt & Solar",
    label: "Aerospace Blue × Solar Flare",
    category: "contrast",
    accentHex: "#38bdf8",
    accentSecondaryHex: "#fb923c",
    dimHex: "#0284c7",
    description: "Cloud telemetry: electric blue architecture with solar risk indicators",
  },
  {
    id: "amethyst-gold",
    name: "Amethyst & Gold",
    label: "Neon Violet × Laser Gold",
    category: "contrast",
    accentHex: "#a855f7",
    accentSecondaryHex: "#f59e0b",
    dimHex: "#9333ea",
    description: "Cyber-industrial compiler: deep violet engine with gold contract pins",
  },
  {
    id: "matrix-crimson",
    name: "Matrix & Crimson",
    label: "Terminal Green × Fault Red",
    category: "contrast",
    accentHex: "#10b981",
    accentSecondaryHex: "#f43f5e",
    dimHex: "#059669",
    description: "Linux kernel tracer: terminal green tree with crimson blast radius markers",
  },
  {
    id: "fuchsia-cyan",
    name: "Fuchsia & Cyan",
    label: "Tokyo Neon × Arctic Ice",
    category: "contrast",
    accentHex: "#ec4899",
    accentSecondaryHex: "#22d3ee",
    dimHex: "#db2777",
    description: "Synthwave telemetry: hyper fuchsia parser with arctic ice citations",
  },
  {
    id: "titanium-gold",
    name: "Titanium & Gold",
    label: "Monolith × Hazard Gold",
    category: "contrast",
    accentHex: "#e4e4e7",
    accentSecondaryHex: "#eab308",
    dimHex: "#a1a1aa",
    description: "Monochrome precision: industrial silver layout with high-vis gold warnings",
  },

  // 2. Classic Monochrome Presets
  {
    id: "rose",
    name: "Pure Crimson",
    label: "Crimson Rust",
    category: "monochrome",
    accentHex: "#f43f5e",
    accentSecondaryHex: "#f43f5e",
    dimHex: "#e11d48",
    description: "Radical high-visibility single-tone compiler red",
  },
  {
    id: "amber",
    name: "Pure Amber",
    label: "Phosphor Terminal",
    category: "monochrome",
    accentHex: "#fbbf24",
    accentSecondaryHex: "#fbbf24",
    dimHex: "#d97706",
    description: "Classic single-tone compiler diagnostic amber",
  },
  {
    id: "emerald",
    name: "Pure Emerald",
    label: "Matrix Console",
    category: "monochrome",
    accentHex: "#10b981",
    accentSecondaryHex: "#10b981",
    dimHex: "#059669",
    description: "Phosphor green single-tone terminal hacker view",
  },
  {
    id: "cyan",
    name: "Pure Cyan",
    label: "Electric Cobalt",
    category: "monochrome",
    accentHex: "#06b6d4",
    accentSecondaryHex: "#06b6d4",
    dimHex: "#0891b2",
    description: "High-contrast single-tone telemetry cyber blue",
  },
  {
    id: "violet",
    name: "Pure Violet",
    label: "Cyber Amethyst",
    category: "monochrome",
    accentHex: "#a855f7",
    accentSecondaryHex: "#a855f7",
    dimHex: "#9333ea",
    description: "Deep neon single-tone compiler purple",
  },
  {
    id: "zinc",
    name: "Pure Zinc",
    label: "Pure Monolith",
    category: "monochrome",
    accentHex: "#e4e4e7",
    accentSecondaryHex: "#e4e4e7",
    dimHex: "#a1a1aa",
    description: "Industrial silver single-tone monochrome",
  },
];

interface ThemeContextType {
  theme: ThemeId;
  setTheme: (theme: ThemeId) => void;
  cycleTheme: () => void;
  currentThemeMeta: ThemeMeta;
  themes: ThemeMeta[];
}

const ThemeContext = createContext<ThemeContextType>({
  theme: "crimson-cyan",
  setTheme: () => {},
  cycleTheme: () => {},
  currentThemeMeta: THEMES[0],
  themes: THEMES,
});

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<ThemeId>("crimson-cyan");
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    try {
      const saved = localStorage.getItem("gitwise-theme") as ThemeId;
      if (saved && THEMES.some((t) => t.id === saved)) {
        setThemeState(saved);
        document.documentElement.setAttribute("data-theme", saved);
      } else {
        document.documentElement.setAttribute("data-theme", "crimson-cyan");
      }
    } catch {
      document.documentElement.setAttribute("data-theme", "crimson-cyan");
    }
    setMounted(true);
  }, []);

  const setTheme = (newTheme: ThemeId) => {
    setThemeState(newTheme);
    try {
      localStorage.setItem("gitwise-theme", newTheme);
      document.documentElement.setAttribute("data-theme", newTheme);
      if (typeof window !== "undefined") {
        window.dispatchEvent(
          new CustomEvent("gitwise-theme-change", { detail: { theme: newTheme } })
        );
      }
    } catch {
      // ignore storage errors
    }
  };

  const cycleTheme = () => {
    const currentIndex = THEMES.findIndex((t) => t.id === theme);
    const nextIndex = (currentIndex + 1) % THEMES.length;
    setTheme(THEMES[nextIndex].id);
  };

  const currentThemeMeta = THEMES.find((t) => t.id === theme) || THEMES[0];

  return (
    <ThemeContext.Provider
      value={{
        theme,
        setTheme,
        cycleTheme,
        currentThemeMeta,
        themes: THEMES,
      }}
    >
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error("useTheme must be used within a ThemeProvider");
  }
  return context;
}
