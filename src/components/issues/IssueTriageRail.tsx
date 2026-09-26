"use client";

import React from "react";
import { IssueModel, AffectedFile } from "@/lib/issueData";
import { AlertCircle, FileCode, CheckCircle, Flame, Layers, ExternalLink } from "lucide-react";

interface IssueTriageRailProps {
  issues: IssueModel[];
  selectedIssue: IssueModel;
  onSelectIssue: (issue: IssueModel) => void;
  onSelectFile: (filePath: string) => void;
}

export function IssueTriageRail({
  issues,
  selectedIssue,
  onSelectIssue,
  onSelectFile,
}: IssueTriageRailProps) {
  const getBlastRadiusBadge = (score: number) => {
    if (score >= 3.0) {
      return <span className="text-[#f87171] border border-[#f87171]/40 px-1 py-0.2 text-[9px] font-bold">CRITICAL {score.toFixed(1)}</span>;
    }
    if (score >= 2.0) {
      return <span className="text-[#fbbf24] border border-[#fbbf24]/40 px-1 py-0.2 text-[9px] font-bold">MODERATE {score.toFixed(1)}</span>;
    }
    return <span className="text-[#4ade80] border border-[#4ade80]/40 px-1 py-0.2 text-[9px] font-bold">ISOLATED {score.toFixed(1)}</span>;
  };

  return (
    <div className="flex flex-col gap-3 font-mono text-xs">
      {/* 1. Issue Triage Stream */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2.5">
        <div className="border-b border-[#262626] pb-2 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Flame className="w-3.5 h-3.5 text-[#fbbf24]" />
            <span className="font-bold text-[#fafafa] uppercase tracking-wider text-[11px]">
              ISSUE TRIAGE STREAM
            </span>
          </div>
          <span className="text-[10px] text-[#525252]">{issues.length} TRACKED</span>
        </div>

        <div className="flex flex-col gap-1.5 max-h-[360px] overflow-y-auto pr-1">
          {issues.map((issue) => {
            const isSelected = issue.id === selectedIssue.id;
            return (
              <button
                key={issue.id}
                onClick={() => onSelectIssue(issue)}
                className={`p-2.5 text-left border transition-all ${
                  isSelected
                    ? "border-[#fbbf24] bg-[#141208]"
                    : "border-[#1f1f1f] bg-[#0c0c0c] hover:border-[#383838] hover:bg-[#121212]"
                }`}
              >
                <div className="flex items-center justify-between gap-1 mb-1">
                  <span className={`font-bold ${isSelected ? "text-[#fbbf24]" : "text-[#fafafa]"}`}>
                    #{issue.number}
                  </span>
                  {getBlastRadiusBadge(issue.blastRadius.score)}
                </div>

                <div className="text-[11px] text-[#fafafa] font-medium leading-snug line-clamp-2 mb-2">
                  {issue.title}
                </div>

                <div className="flex flex-wrap items-center justify-between text-[10px] text-[#8a8a8a] border-t border-[#1f1f1f] pt-1.5">
                  <span className="truncate max-w-[140px] text-[#8a8a8a]">
                    {issue.subsystem}
                  </span>
                  <span className="text-[#fbbf24]">
                    {issue.blastRadius.fileCount} files
                  </span>
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* 2. Impacted Files & Blast Radius Summary */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2">
        <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            <Layers className="w-3.5 h-3.5 text-[#fbbf24]" />
            <span className="font-bold text-[#fafafa] uppercase text-[11px]">
              IMPACTED REPO FILES
            </span>
          </div>
          <span className="text-[10px] text-[#4ade80]">
            {selectedIssue.affectedFiles.length} FILES IDENTIFIED
          </span>
        </div>

        <p className="text-[10px] text-[#8a8a8a] leading-relaxed">
          {selectedIssue.blastRadius.riskAssessment}
        </p>

        <div className="flex flex-col gap-1 mt-1">
          {selectedIssue.affectedFiles.map((file) => (
            <button
              key={file.path}
              onClick={() => onSelectFile(file.path)}
              className="group flex items-center justify-between p-1.5 text-left border border-[#1a1a1a] bg-[#080808] hover:border-[#fbbf24] hover:bg-[#121212] transition-colors"
            >
              <div className="flex items-center gap-1.5 min-w-0">
                <FileCode className="w-3 h-3 text-[#fbbf24] shrink-0" />
                <span className="text-[10px] text-[#fafafa] truncate group-hover:text-[#fbbf24]">
                  {file.path.split("/").pop()}
                </span>
                <span className="text-[9px] text-[#525252] truncate hidden sm:inline">
                  ({file.role})
                </span>
              </div>
              <div className="flex items-center gap-2 shrink-0 text-[10px]">
                <span className="text-[#4ade80]">+{file.linesChanged}</span>
                <ExternalLink className="w-2.5 h-2.5 text-[#525252] group-hover:text-[#fbbf24]" />
              </div>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
