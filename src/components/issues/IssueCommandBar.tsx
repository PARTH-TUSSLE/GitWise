"use client";

import React, { useState } from "react";
import Link from "next/link";
import { Terminal, RefreshCw, GitPullRequest, ArrowRight, HelpCircle } from "lucide-react";
import { ThemeSelector } from "@/components/ThemeSelector";

interface IssueCommandBarProps {
  currentRepo: string;
  currentIssueNumber: number;
  activeStage: number; // 1, 2, or 3
  onSelectStage: (stage: number) => void;
  onSelectRepo: (repo: string) => void;
  onSearchIssue: (query: string) => void;
}

export function IssueCommandBar({
  currentRepo,
  currentIssueNumber,
  activeStage,
  onSelectStage,
  onSelectRepo,
  onSearchIssue,
}: IssueCommandBarProps) {
  const [inputValue, setInputValue] = useState(`${currentRepo}#${currentIssueNumber}`);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (inputValue.trim()) {
      onSearchIssue(inputValue.trim());
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
          <span className="text-[#8a8a8a] select-none">issue planner:</span>
          <div className="relative flex items-center flex-1">
            <input
              type="text"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              placeholder="owner/repo#number or keyword"
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

        {/* Presets and Stage Controls */}
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

          {/* 3-Stage Quick Pickers */}
          <div className="flex items-center border border-[#262626] p-0.5 bg-[#080808]">
            <button
              onClick={() => onSelectStage(1)}
              className={`px-2 py-0.5 text-[10px] transition-colors ${
                activeStage === 1
                  ? "bg-[#1f1a09] text-[#fbbf24] font-bold"
                  : "text-[#8a8a8a] hover:text-[#fafafa]"
              }`}
            >
              01 TRIAGE
            </button>
            <button
              onClick={() => onSelectStage(2)}
              className={`px-2 py-0.5 text-[10px] transition-colors ${
                activeStage === 2
                  ? "bg-[#1f1a09] text-[#fbbf24] font-bold"
                  : "text-[#8a8a8a] hover:text-[#fafafa]"
              }`}
            >
              02 BLAST RADIUS
            </button>
            <button
              onClick={() => onSelectStage(3)}
              className={`px-2 py-0.5 text-[10px] transition-colors ${
                activeStage === 3
                  ? "bg-[#1f1a09] text-[#fbbf24] font-bold"
                  : "text-[#8a8a8a] hover:text-[#fafafa]"
              }`}
            >
              03 BLUEPRINT
            </button>
          </div>

          <div className="hidden xl:flex items-center gap-1 text-[10px] text-[#fbbf24] border border-[#333333] px-1.5 py-0.5 bg-[#121008]">
            <span>CLI: SOON</span>
          </div>

          <ThemeSelector />

          {/* Cross-Surface Links */}
          <div className="flex items-center gap-1.5 border-l border-[#262626] pl-2">
            <Link
              href="/pr"
              className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#8a8a8a] hover:text-[#fbbf24] text-[10px] transition-colors"
            >
              PR REVIEWER →
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
