"use client";

import React from "react";
import { PullRequestModel, PRFileDiff } from "@/lib/prData";
import { GitPullRequest, FileCode, CheckCircle2, AlertCircle, ExternalLink, GitCommit } from "lucide-react";

interface PRRailProps {
  prs: PullRequestModel[];
  selectedPR: PullRequestModel;
  onSelectPR: (pr: PullRequestModel) => void;
  selectedFile: string | null;
  onSelectFile: (filePath: string) => void;
  onInspectFile: (filePath: string) => void;
}

export function PRRail({
  prs,
  selectedPR,
  onSelectPR,
  selectedFile,
  onSelectFile,
  onInspectFile,
}: PRRailProps) {
  const getStatusBadge = (status: PullRequestModel["status"]) => {
    switch (status) {
      case "approved":
        return <span className="text-[#4ade80] border border-[#4ade80]/40 px-1 py-0.2 text-[9px] font-bold">APPROVED</span>;
      case "ready_for_review":
        return <span className="text-[#fbbf24] border border-[#fbbf24]/40 px-1 py-0.2 text-[9px] font-bold">REVIEW READY</span>;
      default:
        return <span className="text-[#8a8a8a] border border-[#8a8a8a]/40 px-1 py-0.2 text-[9px] font-bold">IN REVIEW</span>;
    }
  };

  return (
    <div className="flex flex-col gap-3 font-mono text-xs">
      {/* 1. Pull Request Selector Stream */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2.5">
        <div className="border-b border-[#262626] pb-2 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <GitPullRequest className="w-3.5 h-3.5 text-[#fbbf24]" />
            <span className="font-bold text-[#fafafa] uppercase tracking-wider text-[11px]">
              PULL REQUESTS
            </span>
          </div>
          <span className="text-[10px] text-[#525252]">{prs.length} TRACKED</span>
        </div>

        <div className="flex flex-col gap-1.5 max-h-[300px] overflow-y-auto pr-1">
          {prs.map((pr) => {
            const isSelected = pr.id === selectedPR.id;
            return (
              <button
                key={pr.id}
                onClick={() => onSelectPR(pr)}
                className={`p-2.5 text-left border transition-all ${
                  isSelected
                    ? "border-[#fbbf24] bg-[#141208]"
                    : "border-[#1f1f1f] bg-[#0c0c0c] hover:border-[#383838] hover:bg-[#121212]"
                }`}
              >
                <div className="flex items-center justify-between gap-1 mb-1">
                  <span className={`font-bold ${isSelected ? "text-[#fbbf24]" : "text-[#fafafa]"}`}>
                    #{pr.number}
                  </span>
                  {getStatusBadge(pr.status)}
                </div>

                <div className="text-[11px] text-[#fafafa] font-medium leading-snug line-clamp-2 mb-2">
                  {pr.title}
                </div>

                <div className="flex items-center justify-between text-[10px] text-[#8a8a8a] border-t border-[#1f1f1f] pt-1.5">
                  <span className="text-[#8a8a8a]">@{pr.author}</span>
                  <div className="flex items-center gap-1.5">
                    <span className="text-[#4ade80]">+{pr.stats.additions}</span>
                    <span className="text-[#f87171]">-{pr.stats.deletions}</span>
                  </div>
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* 2. Modified Files & Line Deltatree */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2">
        <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            <FileCode className="w-3.5 h-3.5 text-[#fbbf24]" />
            <span className="font-bold text-[#fafafa] uppercase text-[11px]">
              MODIFIED FILES ({selectedPR.files.length})
            </span>
          </div>
          <span className="text-[10px] text-[#8a8a8a]">
            {selectedPR.stats.commitsCount} COMMITS
          </span>
        </div>

        <div className="flex flex-col gap-1 mt-1 max-h-[360px] overflow-y-auto pr-1">
          {selectedPR.files.map((file) => {
            const isFileSelected = selectedFile === file.path;
            return (
              <div
                key={file.path}
                className={`group flex items-center justify-between p-2 border transition-colors ${
                  isFileSelected
                    ? "border-[#fbbf24] bg-[#141208]"
                    : "border-[#1a1a1a] bg-[#080808] hover:border-[#383838] hover:bg-[#121212]"
                }`}
              >
                <button
                  onClick={() => onSelectFile(file.path)}
                  className="flex items-center gap-1.5 min-w-0 flex-1 text-left"
                >
                  <FileCode className="w-3 h-3 text-[#fbbf24] shrink-0" />
                  <div className="flex flex-col min-w-0">
                    <span className="text-[10px] text-[#fafafa] truncate group-hover:text-[#fbbf24]">
                      {file.path.split("/").pop()}
                    </span>
                    <span className="text-[9px] text-[#525252] truncate">
                      {file.subsystem}
                    </span>
                  </div>
                </button>

                <div className="flex items-center gap-2 shrink-0 text-[10px]">
                  <span className="text-[#4ade80]">+{file.additions}</span>
                  <span className="text-[#f87171]">-{file.deletions}</span>
                  <button
                    onClick={() => onInspectFile(file.path)}
                    title="Inspect file"
                    className="p-0.5 text-[#525252] hover:text-[#fbbf24]"
                  >
                    <ExternalLink className="w-2.5 h-2.5" />
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
