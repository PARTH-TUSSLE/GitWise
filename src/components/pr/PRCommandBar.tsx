"use client";

import React, { useState } from "react";
import Link from "next/link";
import { GitPullRequest, ArrowRight } from "lucide-react";
import { ThemeSelector } from "@/components/ThemeSelector";

interface PRCommandBarProps {
  currentRepo: string;
  currentPRNumber: number;
  activeView: "architecture" | "diffs";
  onSelectView: (view: "architecture" | "diffs") => void;
  onSelectRepo: (repo: string) => void;
  onSearchPR: (query: string) => void;
  stats: {
    additions: number;
    deletions: number;
    filesChanged: number;
  };
}

export function PRCommandBar({
  currentRepo,
  currentPRNumber,
  activeView,
  onSelectView,
  onSelectRepo,
  onSearchPR,
  stats,
}: PRCommandBarProps) {
  const [inputValue, setInputValue] = useState(`${currentRepo}#${currentPRNumber}`);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (inputValue.trim()) {
      onSearchPR(inputValue.trim());
    }
  };

  const presetRepos = [
    { repo: "vercel/next.js", label: "@next.js" },
    { repo: "facebook/react", label: "@react" },
    { repo: "kubernetes/kubernetes", label: "@kubernetes" },
    { repo: "pallets/flask", label: "@flask" },
    { repo: "meshery/meshery", label: "@meshery" },
  ];

  return (
    <div className="border border-[#262626] bg-[#0c0c0c] p-3 text-xs font-mono">
      <div className="flex flex-wrap items-center justify-between gap-3">
        {/* CLI Prompt Form */}
        <form onSubmit={handleSubmit} className="flex items-center gap-2 flex-1 min-w-[280px]">
          <span className="text-[#fbbf24] font-bold select-none">gitwise&gt;</span>
          <span className="text-[#8a8a8a] select-none">pr reviewer:</span>
          <div className="relative flex items-center flex-1">
            <input
              type="text"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              placeholder="owner/repo#pr or keyword"
              className="w-full bg-transparent text-[#fafafa] font-mono focus:outline-none placeholder:text-[#525252] border-b border-dashed border-[#404040] focus:border-[#fbbf24] py-0.5"
            />
          </div>
          <button
            type="submit"
            className="px-2 py-0.5 border border-[#404040] hover:border-[#fbbf24] hover:bg-[#1f1f1f] text-[#fbbf24] font-mono text-[11px] transition-colors"
          >
            DISSECT [↵]
          </button>
        </form>

        {/* Presets and Controls */}
        <div className="flex items-center gap-3 text-[11px]">
          {/* Presets */}
          <div className="flex items-center gap-1">
            <span className="hidden sm:inline select-none text-[#525252]">presets:</span>
            {presetRepos.map((preset) => (
              <button
                key={preset.repo}
                onClick={() => {
                  setInputValue(`${preset.repo}#recent`);
                  onSelectRepo(preset.repo);
                }}
                className={`px-1.5 py-0.5 border text-[10px] transition-colors ${
                  currentRepo === preset.repo
                    ? "border-[#fbbf24] text-[#fbbf24] bg-[#1a1708]"
                    : "border-[#262626] text-[#8a8a8a] hover:border-[#525252] hover:text-[#fafafa]"
                }`}
              >
                {preset.label}
              </button>
            ))}
          </div>

          {/* Diff Stats Badge */}
          <div className="hidden md:flex items-center gap-1.5 border border-[#1f1f1f] bg-[#070707] px-2 py-0.5 text-[10px]">
            <span className="text-[#4ade80]">+{stats.additions}</span>
            <span className="text-[#f87171]">-{stats.deletions}</span>
            <span className="text-[#525252]">|</span>
            <span className="text-[#8a8a8a]">{stats.filesChanged} files</span>
          </div>

          {/* View Mode Toggle: ARCH MATRIX vs FILE DIFFS */}
          <div className="flex items-center border border-[#262626] p-0.5 bg-[#080808]">
            <button
              onClick={() => onSelectView("architecture")}
              className={`px-2 py-0.5 text-[10px] transition-colors ${
                activeView === "architecture"
                  ? "bg-[#1f1a09] text-[#fbbf24] font-bold"
                  : "text-[#8a8a8a] hover:text-[#fafafa]"
              }`}
            >
              ARCH MATRIX
            </button>
            <button
              onClick={() => onSelectView("diffs")}
              className={`px-2 py-0.5 text-[10px] transition-colors ${
                activeView === "diffs"
                  ? "bg-[#1f1a09] text-[#fbbf24] font-bold"
                  : "text-[#8a8a8a] hover:text-[#fafafa]"
              }`}
            >
              FILE DIFFS
            </button>
          </div>

          <div className="hidden xl:flex items-center gap-1 text-[10px] text-[#fbbf24] border border-[#333333] px-1.5 py-0.5 bg-[#121008]">
            <span>CLI: SOON</span>
          </div>

          <ThemeSelector />

          {/* Cross-Surface Links */}
          <div className="flex items-center gap-1.5 border-l border-[#262626] pl-2">
            <Link
              href="/mentor"
              className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#8a8a8a] hover:text-[#fbbf24] text-[10px] transition-colors"
            >
              MENTOR →
            </Link>
            <Link
              href="/issues"
              className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#8a8a8a] hover:text-[#fbbf24] text-[10px] transition-colors"
            >
              ISSUES
            </Link>
            <Link
              href="/repo"
              className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#8a8a8a] hover:text-[#fbbf24] text-[10px] transition-colors"
            >
              REPO
            </Link>
            <Link
              href="/"
              className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#8a8a8a] hover:text-[#fbbf24] text-[10px] transition-colors"
            >
              GITSTAT
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
