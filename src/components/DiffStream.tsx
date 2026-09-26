"use client";

import React from "react";
import { ContributorProfile } from "@/lib/mockData";
import { GitCommit, ExternalLink, ArrowUpRight, CheckCircle2 } from "lucide-react";

interface DiffStreamProps {
  profile: ContributorProfile;
  onSelectDiff: (diffId: string) => void;
}

export function DiffStream({ profile, onSelectDiff }: DiffStreamProps) {
  return (
    <div className="border border-[#262626] bg-[#0a0a0a] p-3 text-xs flex flex-col h-full">
      {/* Header matching comp */}
      <div className="border-b border-[#262626] pb-2 mb-3 flex items-center justify-between">
        <div>
          <div className="font-bold text-[#fafafa]">LIVE stream of PR diff stats</div>
          <div className="text-[10px] text-[#525252]">REAL-TIME REPOSITORY AUDIT</div>
        </div>
        <span className="flex h-2 w-2 relative">
          <span className="animate-ping absolute inline-flex h-full w-full bg-[#4ade80] opacity-75"></span>
          <span className="relative inline-flex h-2 w-2 bg-[#4ade80]"></span>
        </span>
      </div>

      {/* Aggregate Delta Summary matching comp banner */}
      <div className="border border-[#1f1f1f] bg-[#0d0d0d] p-2.5 mb-3 font-mono">
        <div className="flex items-center justify-between text-[11px] mb-1.5">
          <span className="text-[#8a8a8a]">Total Verified Additions</span>
          <span className="text-[#4ade80] font-bold">+{profile.metrics.linesAdded.toLocaleString()} A</span>
        </div>
        <div className="flex items-center justify-between text-[11px] mb-1.5">
          <span className="text-[#8a8a8a]">Total Verified Deletions</span>
          <span className="text-[#f87171] font-bold">-{profile.metrics.linesDeleted.toLocaleString()} D</span>
        </div>
        <div className="flex items-center justify-between text-[10px] text-[#525252] border-t border-[#1a1a1a] pt-1">
          <span>Files modified: {profile.metrics.filesChanged.toLocaleString()}</span>
          <span>Net delta: +{(profile.metrics.linesAdded - profile.metrics.linesDeleted).toLocaleString()}</span>
        </div>
      </div>

      {/* Stream Items List */}
      <div className="flex flex-col gap-2 flex-1 overflow-y-auto max-h-[500px] pr-1">
        {profile.recentDiffs.map((diff) => {
          const isMerge = diff.type === "PR_MERGED";
          return (
            <div
              key={diff.id}
              onClick={() => onSelectDiff(diff.id)}
              className="border border-[#1f1f1f] hover:border-[#fbbf24] bg-[#0e0e0e] p-2 cursor-pointer transition-colors group"
            >
              <div className="flex items-center justify-between text-[10px] font-mono mb-1">
                <span className="text-[#fbbf24] font-medium group-hover:underline">
                  {diff.repo} {diff.prNumber ? `[#${diff.prNumber}]` : ""}
                </span>
                <span className="text-[#525252]">{diff.timestamp}</span>
              </div>

              <div className="text-[11px] text-[#fafafa] line-clamp-2 mb-1.5 font-mono">
                {diff.message}
              </div>

              <div className="flex items-center justify-between text-[10px] font-mono pt-1 border-t border-[#191919]">
                <span className="text-[#525252] flex items-center gap-1">
                  <GitCommit className="w-3 h-3" />
                  {diff.commitHash}
                </span>

                <div className="flex items-center gap-2">
                  {diff.added > 0 && (
                    <span className="text-[#4ade80] font-medium">+{diff.added}</span>
                  )}
                  {diff.deleted > 0 && (
                    <span className="text-[#f87171] font-medium">-{diff.deleted}</span>
                  )}
                  {diff.added === 0 && diff.deleted === 0 && (
                    <span className="text-[#8a8a8a]">Review Log</span>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Footer live status */}
      <div className="mt-3 pt-2 border-t border-[#1f1f1f] flex items-center justify-between text-[10px] text-[#525252]">
        <span>Streaming events</span>
        <span className="text-[#4ade80]">SYNCED 100%</span>
      </div>
    </div>
  );
}
