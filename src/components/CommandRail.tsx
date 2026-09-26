"use client";

import React from "react";
import { ContributorProfile } from "@/lib/mockData";
import { GitPullRequest, GitMerge, MessageSquare, AlertCircle, FolderGit2, Activity, ShieldCheck, Terminal } from "lucide-react";

interface CommandRailProps {
  profile: ContributorProfile;
  activeTab: string;
  onSelectTab: (tabId: string) => void;
  onInspectRepo: (repoName: string) => void;
}

export function CommandRail({
  profile,
  activeTab,
  onSelectTab,
  onInspectRepo,
}: CommandRailProps) {
  const navItems = [
    { id: "overview", label: "_overview_", count: null },
    { id: "merged_prs", label: "_merged_prs_", count: profile.metrics.mergedPRs },
    { id: "reviews", label: "_code_reviews_", count: profile.metrics.codeReviewsGiven },
    { id: "issues", label: "_issues_", count: `${profile.metrics.issuesOpened}/${profile.metrics.issuesParticipatedIn}` },
    { id: "repositories", label: "_repositories_", count: profile.metrics.activeRepositories },
    { id: "diff_delta", label: "_diff_delta_", count: "+54.2k" },
  ];

  return (
    <div className="flex flex-col gap-3">
      {/* Box 1: Navigation Rail Header */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3">
        <div className="border-b border-[#262626] pb-2 mb-3 flex items-center justify-between text-[11px] text-[#8a8a8a]">
          <span className="font-bold text-[#fafafa]">Navigation / rail look</span>
          <div className="flex items-center gap-1.5">
            <span className="text-[#525252]">WEB WORKBENCH</span>
            <span className="text-[#fbbf24] text-[9px] border border-[#333333] px-1 bg-[#141208]">CLI: SOON</span>
          </div>
        </div>

        {/* Menu Items matching comp */}
        <div className="flex flex-col gap-1 text-xs">
          <div className="text-[10px] text-[#525252] select-none mb-1">&gt; _gatsta_</div>
          {navItems.map((item) => {
            const isActive = activeTab === item.id;
            return (
              <button
                key={item.id}
                onClick={() => onSelectTab(item.id)}
                className={`flex items-center justify-between px-2 py-1 text-left transition-colors font-mono ${
                  isActive
                    ? "bg-[#1f1a09] text-[#fbbf24] border-l-2 border-[#fbbf24] font-semibold"
                    : "text-[#8a8a8a] hover:text-[#fafafa] hover:bg-[#141414]"
                }`}
              >
                <span>
                  {isActive ? "> " : "  "}
                  {item.label}
                </span>
                {item.count !== null && (
                  <span className={`text-[10px] ${isActive ? "text-[#fbbf24]" : "text-[#525252]"}`}>
                    [{item.count}]
                  </span>
                )}
              </button>
            );
          })}
        </div>
      </div>

      {/* Box 2: Contributor Identity Details */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 text-xs">
        <div className="border-b border-[#262626] pb-2 mb-3 flex items-center justify-between text-[11px]">
          <span className="text-[#8a8a8a]">CONTRIBUTOR_ID</span>
          <span className="text-[#4ade80] flex items-center gap-1">
            <span className="w-1.5 h-1.5 bg-[#4ade80]"></span>
            VERIFIED
          </span>
        </div>

        {/* Profile Card */}
        <div className="flex items-start gap-3 mb-3">
          <div className="w-12 h-12 border border-[#404040] bg-[#141414] overflow-hidden flex-shrink-0">
            <img
              src={profile.avatarUrl}
              alt={profile.name}
              className="w-full h-full object-cover grayscale contrast-125 hover:grayscale-0 transition-all"
            />
          </div>
          <div className="min-w-0 flex-1">
            <div className="font-bold text-[#fafafa] truncate">{profile.name}</div>
            <div className="text-[#fbbf24] text-[11px] truncate">@{profile.username}</div>
            <div className="text-[10px] text-[#525252] mt-0.5">Joined {profile.joined}</div>
          </div>
        </div>

        <p className="text-[11px] text-[#8a8a8a] leading-relaxed mb-3 border-l border-[#262626] pl-2">
          {profile.bio}
        </p>

        {/* Primary Tech Stack */}
        <div className="mb-3">
          <div className="text-[10px] text-[#525252] mb-1.5 uppercase tracking-wider">Primary Languages</div>
          <div className="flex flex-wrap gap-1">
            {profile.primaryLanguages.map((lang) => (
              <span
                key={lang}
                className="px-1.5 py-0.5 text-[10px] border border-[#262626] bg-[#121212] text-[#fafafa]"
              >
                {lang}
              </span>
            ))}
          </div>
        </div>

        {/* Top Repositories */}
        <div>
          <div className="text-[10px] text-[#525252] mb-1.5 uppercase tracking-wider">Key Repositories</div>
          <div className="flex flex-col gap-1">
            {profile.repositories.slice(0, 3).map((repo) => (
              <button
                key={repo.name}
                onClick={() => onInspectRepo(repo.name)}
                className="text-left p-1.5 border border-[#1f1f1f] hover:border-[#fbbf24] hover:bg-[#141414] transition-colors group"
              >
                <div className="flex items-center justify-between text-[11px]">
                  <span className="text-[#fafafa] group-hover:text-[#fbbf24] truncate font-medium">
                    {repo.name}
                  </span>
                  <span className="text-[9px] text-[#525252]">★{repo.stars}</span>
                </div>
                <div className="flex items-center justify-between text-[10px] text-[#525252] mt-0.5">
                  <span>{repo.role}</span>
                  <span>{repo.prs} PRs</span>
                </div>
              </button>
            ))}
          </div>
        </div>

        {/* Footer shortcuts */}
        <div className="mt-3 pt-2 border-t border-[#1f1f1f] text-[10px] text-[#525252] flex justify-between">
          <span>[q] Quit</span>
          <span>[r] Reload</span>
          <span>[?] Help</span>
        </div>
      </div>
    </div>
  );
}
