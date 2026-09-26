"use client";

import React, { useState } from "react";
import Link from "next/link";
import { Search, Terminal, RefreshCw, Share2, HelpCircle } from "lucide-react";
import { ThemeSelector } from "@/components/ThemeSelector";

interface CommandPromptProps {
  currentUsername: string;
  onSelectUser: (username: string) => void;
  onRefresh: () => void;
  onExport: () => void;
  onHelp: () => void;
}

export function CommandPrompt({
  currentUsername,
  onSelectUser,
  onRefresh,
  onExport,
  onHelp,
}: CommandPromptProps) {
  const [inputValue, setInputValue] = useState(currentUsername);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (inputValue.trim()) {
      onSelectUser(inputValue.trim().replace(/^@/, ""));
    }
  };

  const sampleUsers = ["alexR_dev", "torvalds", "gaearon"];

  return (
    <div className="border border-[#262626] bg-[#0c0c0c] p-3 text-xs">
      <div className="flex flex-wrap items-center justify-between gap-3">
        {/* CLI Prompt Form */}
        <form onSubmit={handleSubmit} className="flex items-center gap-2 flex-1 min-w-[280px]">
          <span className="text-[#fbbf24] font-bold select-none">gitstat&gt;</span>
          <span className="text-[#8a8a8a] select-none">user:</span>
          <div className="relative flex items-center flex-1">
            <span className="text-[#fbbf24] mr-1 select-none">@</span>
            <input
              type="text"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              placeholder="github_username"
              className="w-full bg-transparent text-[#fafafa] font-mono focus:outline-none placeholder:text-[#525252] border-b border-dashed border-[#404040] focus:border-[#fbbf24] py-0.5"
            />
          </div>
          <button
            type="submit"
            className="px-2 py-0.5 border border-[#404040] hover:border-[#fbbf24] hover:bg-[#1f1f1f] text-[#fbbf24] font-mono text-[11px] transition-colors"
          >
            EXEC [↵]
          </button>
        </form>

        <div className="flex items-center gap-4 text-[11px]">
          <div className="hidden sm:flex items-center gap-1.5 text-[#4ade80]">
            <span className="inline-block w-2 h-2 rounded-none bg-[#4ade80] animate-pulse"></span>
            <span className="font-bold">STATUS: 200 OK</span>
            <span className="text-[#525252]">|</span>
            <span className="text-[#8a8a8a]">WEB LIVE</span>
          </div>
          <div className="hidden lg:flex items-center gap-1 text-[10px] text-[#fbbf24] border border-[#333333] px-1.5 py-0.5 bg-[#121008]">
            <span>CLI: COMING SOON</span>
          </div>

          {/* Sample quick picks */}
          <div className="flex items-center gap-1 text-[#8a8a8a]">
            <span className="hidden md:inline select-none text-[#525252]">samples:</span>
            {sampleUsers.map((u) => (
              <button
                key={u}
                onClick={() => {
                  setInputValue(u);
                  onSelectUser(u);
                }}
                className={`px-1.5 py-0.5 border text-[10px] transition-colors ${
                  currentUsername.toLowerCase() === u.toLowerCase()
                    ? "border-[#fbbf24] text-[#fbbf24] bg-[#1a1708]"
                    : "border-[#262626] text-[#8a8a8a] hover:border-[#525252] hover:text-[#fafafa]"
                }`}
              >
                @{u}
              </button>
            ))}
          </div>

          {/* Quick Actions */}
          <div className="flex items-center gap-2 border-l border-[#262626] pl-3">
            <ThemeSelector />
            <Link
              href="/repo"
              className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#8a8a8a] hover:text-[#fbbf24] text-[10px] transition-colors"
            >
              REPO EXPLORER →
            </Link>
            <Link
              href="/mentor"
              className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#8a8a8a] hover:text-[#fbbf24] text-[10px] transition-colors"
            >
              MENTOR →
            </Link>
            <Link
              href="/pr"
              className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#8a8a8a] hover:text-[#fbbf24] text-[10px] transition-colors"
            >
              PR REVIEWER
            </Link>
            <Link
              href="/issues"
              className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#8a8a8a] hover:text-[#fbbf24] text-[10px] transition-colors"
            >
              ISSUES
            </Link>
            <button
              onClick={onRefresh}
              title="Refresh telemetry [r]"
              className="p-1 text-[#8a8a8a] hover:text-[#fbbf24] hover:bg-[#1a1a1a] transition-colors"
            >
              <RefreshCw className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={onExport}
              title="Export portfolio summary [e]"
              className="p-1 text-[#8a8a8a] hover:text-[#fbbf24] hover:bg-[#1a1a1a] transition-colors"
            >
              <Share2 className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={onHelp}
              title="Keyboard shortcuts & help [?]"
              className="p-1 text-[#8a8a8a] hover:text-[#fbbf24] hover:bg-[#1a1a1a] transition-colors"
            >
              <HelpCircle className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
