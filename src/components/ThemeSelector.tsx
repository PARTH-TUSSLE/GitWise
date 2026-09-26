"use client";

import React, { useState, useRef, useEffect } from "react";
import { useTheme, THEMES, ThemeId } from "@/context/ThemeContext";
import { Palette, Check, Sparkles } from "lucide-react";

interface ThemeSelectorProps {
  compact?: boolean;
}

export function ThemeSelector({ compact = false }: ThemeSelectorProps) {
  const { theme, setTheme, currentThemeMeta, cycleTheme } = useTheme();
  const [isOpen, setIsOpen] = useState(false);
  const [categoryFilter, setCategoryFilter] = useState<"all" | "contrast" | "monochrome">("all");
  const containerRef = useRef<HTMLDivElement>(null);

  // Close on outside click
  useEffect(() => {
    const handleOutsideClick = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    if (isOpen) {
      document.addEventListener("mousedown", handleOutsideClick);
    }
    return () => document.removeEventListener("mousedown", handleOutsideClick);
  }, [isOpen]);

  // Close on Escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isOpen) {
        setIsOpen(false);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen]);

  const filteredThemes = THEMES.filter((t) => {
    if (categoryFilter === "all") return true;
    return t.category === categoryFilter;
  });

  return (
    <div className="relative inline-block text-left font-mono" ref={containerRef}>
      {/* Trigger Button with Dual-Color Swatch */}
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        title="Switch color palette preset (Press 'T' to cycle)"
        className={`flex items-center gap-2 px-2.5 py-1 border border-[#262626] bg-[#0c0c0c] hover:border-accent hover:bg-[#141414] text-[10px] text-[#fafafa] transition-colors focus-visible:outline-accent ${
          isOpen ? "border-accent bg-[#141208]" : ""
        }`}
      >
        {/* Dual-color Swatch Badge */}
        <div className="flex items-center border border-[#333333] overflow-hidden flex-shrink-0">
          <span
            className="w-2 h-2.5 inline-block"
            style={{ backgroundColor: currentThemeMeta.accentHex }}
          />
          <span
            className="w-2 h-2.5 inline-block"
            style={{ backgroundColor: currentThemeMeta.accentSecondaryHex }}
          />
        </div>

        <span className="hidden sm:inline text-[#8a8a8a] select-none text-[9px]">THEME:</span>
        <span className="font-bold text-[#fafafa] uppercase tracking-wide truncate max-w-[110px]">
          {currentThemeMeta.name}
        </span>
        <span className="text-[9px] text-[#525252] select-none">▾</span>
      </button>

      {/* Popover Menu */}
      {isOpen && (
        <div className="absolute right-0 mt-1 w-80 max-h-[85vh] overflow-y-auto border border-[#262626] bg-[#0c0c0c] shadow-2xl z-50 p-2.5 text-xs">
          {/* Header */}
          <div className="border-b border-[#262626] pb-2 mb-2 flex items-center justify-between text-[10px] text-[#8a8a8a]">
            <div className="flex items-center gap-1.5 font-bold text-[#fafafa]">
              <Palette className="w-3.5 h-3.5 text-accent" />
              <span>COMPILER PALETTE PRESETS</span>
            </div>
            <span className="text-[9px] text-accent font-mono border border-accent-border px-1">
              [T] SHORTCUT
            </span>
          </div>

          {/* Filter Pills */}
          <div className="flex items-center gap-1 mb-2 border-b border-[#1a1a1a] pb-2 text-[9px]">
            <button
              onClick={() => setCategoryFilter("all")}
              className={`px-2 py-0.5 border transition-colors ${
                categoryFilter === "all"
                  ? "border-accent text-accent bg-accent-soft font-bold"
                  : "border-[#262626] text-[#8a8a8a] hover:text-[#fafafa]"
              }`}
            >
              ALL ({THEMES.length})
            </button>
            <button
              onClick={() => setCategoryFilter("contrast")}
              className={`px-2 py-0.5 border transition-colors ${
                categoryFilter === "contrast"
                  ? "border-accent text-accent bg-accent-soft font-bold"
                  : "border-[#262626] text-[#8a8a8a] hover:text-[#fafafa]"
              }`}
            >
              DUAL CONTRAST (7)
            </button>
            <button
              onClick={() => setCategoryFilter("monochrome")}
              className={`px-2 py-0.5 border transition-colors ${
                categoryFilter === "monochrome"
                  ? "border-accent text-accent bg-accent-soft font-bold"
                  : "border-[#262626] text-[#8a8a8a] hover:text-[#fafafa]"
              }`}
            >
              MONOCHROME (6)
            </button>
          </div>

          {/* Theme list */}
          <div className="flex flex-col gap-1.5 max-h-72 overflow-y-auto pr-1">
            {filteredThemes.map((t) => {
              const isActive = theme === t.id;
              return (
                <button
                  key={t.id}
                  onClick={() => {
                    setTheme(t.id);
                    setIsOpen(false);
                  }}
                  className={`flex items-start justify-between p-2 border text-left transition-all ${
                    isActive
                      ? "border-accent bg-accent-soft text-[#fafafa] shadow-md"
                      : "border-[#191919] hover:border-[#333333] hover:bg-[#141414] text-[#8a8a8a] hover:text-[#fafafa] bg-[#070707]"
                  }`}
                >
                  <div className="flex items-start gap-2.5">
                    {/* Dual-color Swatch */}
                    <div className="flex flex-col border border-[#262626] overflow-hidden mt-0.5 flex-shrink-0">
                      <span
                        className="w-3.5 h-2.5 inline-block"
                        style={{ backgroundColor: t.accentHex }}
                      />
                      <span
                        className="w-3.5 h-2.5 inline-block"
                        style={{ backgroundColor: t.accentSecondaryHex }}
                      />
                    </div>

                    <div className="flex flex-col gap-0.5">
                      <div className="flex items-center gap-2 text-[11px] font-medium leading-none">
                        <span className={isActive ? "text-accent font-bold" : "text-[#fafafa]"}>
                          {t.name}
                        </span>
                        {t.category === "contrast" && (
                          <span className="text-[8px] border border-[#262626] px-1 text-accent-secondary font-mono">
                            DUAL
                          </span>
                        )}
                      </div>
                      <span className="text-[9px] text-[#737373] font-mono">
                        {t.label}
                      </span>
                      <p className="text-[9px] text-[#a3a3a3] leading-tight mt-0.5">
                        {t.description}
                      </p>
                    </div>
                  </div>

                  {isActive && (
                    <span className="text-[9px] text-[#4ade80] font-bold flex items-center gap-0.5 flex-shrink-0 ml-1">
                      <Check className="w-3.5 h-3.5 text-[#4ade80]" />
                    </span>
                  )}
                </button>
              );
            })}
          </div>

          {/* Footer explanation */}
          <div className="border-t border-[#1f1f1f] pt-2 mt-2 text-[9px] text-[#737373] flex items-center justify-between">
            <span>Dual-Tone Contrast Engine</span>
            <span className="text-[#a3a3a3]">Dynamic CSS Variables</span>
          </div>
        </div>
      )}
    </div>
  );
}
