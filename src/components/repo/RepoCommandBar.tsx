"use client";

import React, { useState } from "react";
import Link from "next/link";
import { REPOSITORIES } from "@/lib/repoData";
import { Terminal, RefreshCw, Layers, Compass, HelpCircle } from "lucide-react";
import { ThemeSelector } from "@/components/ThemeSelector";

interface RepoCommandBarProps {
  currentRepoKey: string;
  onSelectRepo: (repoKey: string) => void;
  activeViewMode: "architecture" | "feature_trace";
  onSelectViewMode: (mode: "architecture" | "feature_trace") => void;
  onSwitchToGitStat: () => void;
  onHelp: () => void;
}

export function RepoCommandBar({
  currentRepoKey,
  onSelectRepo,
  activeViewMode,
  onSelectViewMode,
  onSwitchToGitStat,
  onHelp,
}: RepoCommandBarProps) {
  const [inputValue, setInputValue] = useState(currentRepoKey);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (inputValue.trim()) {
      onSelectRepo(inputValue.trim());
    }
  };

  const presetRepos = [
    { key: "vercel/next.js", label: "@next.js" },
    { key: "facebook/react", label: "@react" },
    { key: "kubernetes/kubernetes", label: "@kubernetes" },
    { key: "pallets/flask", label: "@flask" },
    { key: "meshery/meshery", label: "@meshery" },
  ];

  return (
    <div className="border border-[#262626] bg-[#0c0c0c] p-3 text-xs font-mono">
      <div className="flex flex-wrap items-center justify-between gap-3">
        {/* CLI Prompt Form */}
        <form onSubmit={handleSubmit} className="flex items-center gap-2 flex-1 min-w-[280px]">
          <span className="text-[#fbbf24] font-bold select-none">gitwise&gt;</span>
          <span className="text-[#8a8a8a] select-none">repo explorer:</span>
          <div className="relative flex items-center flex-1">
            <input
              type="text"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              placeholder="owner/repository"
              className="w-full bg-transparent text-[#fafafa] font-mono focus:outline-none placeholder:text-[#525252] border-b border-dashed border-[#404040] focus:border-[#fbbf24] py-0.5"
            />
          </div>
          <button
            type="submit"
            className="px-2 py-0.5 border border-[#404040] hover:border-[#fbbf24] hover:bg-[#1f1f1f] text-[#fbbf24] font-mono text-[11px] transition-colors"
          >
            INGEST [↵]
          </button>
        </form>

        {/* Presets and Mode Controls */}
        <div className="flex items-center gap-3 text-[11px]">
          {/* Presets */}
          <div className="flex items-center gap-1">
            <span className="hidden sm:inline select-none text-[#525252]">presets:</span>
            {presetRepos.map((preset) => (
              <button
                key={preset.key}
                onClick={() => {
                  setInputValue(preset.key);
                  onSelectRepo(preset.key);
                }}
                className={`px-1.5 py-0.5 border text-[10px] transition-colors ${
                  currentRepoKey === preset.key
                    ? "border-[#fbbf24] text-[#fbbf24] bg-[#1a1708]"
                    : "border-[#262626] text-[#8a8a8a] hover:border-[#525252] hover:text-[#fafafa]"
                }`}
              >
                {preset.label}
              </button>
            ))}
          </div>

          {/* View Mode Toggle: Architecture vs Feature Trace */}
          <div className="flex items-center border border-[#262626] p-0.5 bg-[#080808]">
            <button
              onClick={() => onSelectViewMode("feature_trace")}
              className={`px-2 py-0.5 text-[10px] transition-colors ${
                activeViewMode === "feature_trace"
                  ? "bg-[#1f1a09] text-[#fbbf24] font-bold"
                  : "text-[#8a8a8a] hover:text-[#fafafa]"
              }`}
            >
              FEATURE TRACER
            </button>
            <button
              onClick={() => onSelectViewMode("architecture")}
              className={`px-2 py-0.5 text-[10px] transition-colors ${
                activeViewMode === "architecture"
                  ? "bg-[#1f1a09] text-[#fbbf24] font-bold"
                  : "text-[#8a8a8a] hover:text-[#fafafa]"
              }`}
            >
              ARCH MATRIX
            </button>
          </div>

          <div className="hidden xl:flex items-center gap-1 text-[10px] text-[#fbbf24] border border-[#333333] px-1.5 py-0.5 bg-[#121008]">
            <span>CLI: SOON</span>
          </div>

          <ThemeSelector />

          {/* Workbench Switcher */}
          <Link
            href="/pr"
            className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#8a8a8a] hover:text-[#fbbf24] text-[10px] transition-colors"
          >
            PR REVIEWER →
          </Link>
          <Link
            href="/issues"
            className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#8a8a8a] hover:text-[#fbbf24] text-[10px] transition-colors"
          >
            ISSUES
          </Link>
          <button
            onClick={onSwitchToGitStat}
            className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#8a8a8a] hover:text-[#fbbf24] text-[10px] transition-colors"
          >
            ← GITSTAT
          </button>
        </div>
      </div>
    </div>
  );
}
