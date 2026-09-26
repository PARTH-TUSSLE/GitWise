"use client";

import React from "react";
import { ContributorProfile } from "@/lib/mockData";
import { X, ExternalLink, GitCommit, FileCode, CheckCircle, Info } from "lucide-react";

interface EvidenceModalProps {
  isOpen: boolean;
  onClose: () => void;
  selectedItem: {
    type: "repo" | "diff" | "help" | "export";
    id: string;
  } | null;
  profile: ContributorProfile;
}

export function EvidenceModal({
  isOpen,
  onClose,
  selectedItem,
  profile,
}: EvidenceModalProps) {
  if (!isOpen || !selectedItem) return null;

  const diff = selectedItem.type === "diff"
    ? profile.recentDiffs.find((d) => d.id === selectedItem.id)
    : null;

  const repo = selectedItem.type === "repo"
    ? profile.repositories.find((r) => r.name === selectedItem.id)
    : null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-none p-4">
      <div className="w-full max-w-xl border border-[#fbbf24] bg-[#0c0c0c] text-xs font-mono shadow-2xl">
        {/* Modal Terminal Header */}
        <div className="flex items-center justify-between border-b border-[#262626] bg-[#141414] px-3 py-2 text-[#fafafa]">
          <div className="flex items-center gap-2">
            <span className="text-[#fbbf24] font-bold">gitstat&gt;</span>
            <span className="text-[#8a8a8a]">
              {selectedItem.type === "diff" && "INSPECT_DIFF_EVIDENCE"}
              {selectedItem.type === "repo" && "INSPECT_REPOSITORY_EVIDENCE"}
              {selectedItem.type === "help" && "KEYBOARD_SHORTCUTS"}
              {selectedItem.type === "export" && "EXPORT_PORTFOLIO_SUMMARY"}
            </span>
          </div>
          <button
            onClick={onClose}
            className="text-[#8a8a8a] hover:text-[#fbbf24] transition-colors p-1"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Modal Body */}
        <div className="p-4 max-h-[80vh] overflow-y-auto">
          {diff && (
            <div className="flex flex-col gap-3">
              <div className="border border-[#1f1f1f] bg-[#080808] p-3">
                <div className="text-[#fbbf24] font-bold text-sm mb-1">{diff.repo}</div>
                <div className="text-xs text-[#fafafa] mb-2">{diff.message}</div>
                <div className="flex flex-wrap gap-3 text-[11px] text-[#8a8a8a] border-t border-[#1f1f1f] pt-2">
                  <span>Commit: <code className="text-[#fafafa]">{diff.commitHash}</code></span>
                  <span>Type: <code className="text-[#fbbf24]">{diff.type}</code></span>
                  <span>Timestamp: {diff.timestamp}</span>
                </div>
              </div>

              <div className="border border-[#262626] p-3 bg-[#0a0a0a]">
                <div className="font-bold text-[#fafafa] mb-2 text-[11px]">VERIFIED CODE DELTA</div>
                <div className="flex gap-4 mb-2 text-xs">
                  <span className="text-[#4ade80] font-bold">+{diff.added} lines added</span>
                  <span className="text-[#f87171] font-bold">-{diff.deleted} lines deleted</span>
                </div>
                <div className="text-[11px] text-[#8a8a8a] leading-relaxed">
                  Evidence verified through Git tree delta inspection against the upstream target branch.
                </div>
              </div>

              <div className="flex justify-end pt-2">
                <a
                  href={`https://github.com/${diff.repo}/commit/${diff.commitHash}`}
                  target="_blank"
                  rel="noreferrer"
                  className="px-3 py-1.5 border border-[#fbbf24] text-[#fbbf24] hover:bg-[#fbbf24] hover:text-[#050505] transition-colors flex items-center gap-1.5"
                >
                  <span>View on GitHub</span>
                  <ExternalLink className="w-3.5 h-3.5" />
                </a>
              </div>
            </div>
          )}

          {repo && (
            <div className="flex flex-col gap-3">
              <div className="border border-[#1f1f1f] bg-[#080808] p-3">
                <div className="text-[#fbbf24] font-bold text-sm mb-1">{repo.name}</div>
                <div className="text-xs text-[#fafafa] mb-2">{repo.description}</div>
                <div className="flex flex-wrap gap-4 text-[11px] text-[#8a8a8a] border-t border-[#1f1f1f] pt-2">
                  <span>Language: <span className="text-[#fafafa]">{repo.language}</span></span>
                  <span>Stars: ★{repo.stars}</span>
                  <span>Forks: {repo.forks}</span>
                  <span>Role: <span className="text-[#fbbf24]">{repo.role}</span></span>
                </div>
              </div>

              <div className="border border-[#262626] p-3 bg-[#0a0a0a]">
                <div className="font-bold text-[#fafafa] mb-1.5 text-[11px]">CONTRIBUTION BREAKDOWN</div>
                <div className="flex gap-4 text-xs mb-2">
                  <span>Commits: <strong className="text-[#fafafa]">{repo.commits}</strong></span>
                  <span>Merged PRs: <strong className="text-[#fbbf24]">{repo.prs}</strong></span>
                </div>
                <p className="text-[11px] text-[#8a8a8a] leading-relaxed">
                  All metrics are sourced directly from repository log events and pull request audits.
                </p>
              </div>

              <div className="flex justify-end pt-2">
                <a
                  href={repo.evidenceUrl}
                  target="_blank"
                  rel="noreferrer"
                  className="px-3 py-1.5 border border-[#fbbf24] text-[#fbbf24] hover:bg-[#fbbf24] hover:text-[#050505] transition-colors flex items-center gap-1.5"
                >
                  <span>Open Repository</span>
                  <ExternalLink className="w-3.5 h-3.5" />
                </a>
              </div>
            </div>
          )}

          {selectedItem.type === "help" && (
            <div className="flex flex-col gap-3">
              <div className="border border-[#1f1f1f] bg-[#080808] p-3 text-[11px]">
                <div className="font-bold text-[#fafafa] mb-2">KEYBOARD COMMANDS</div>
                <div className="grid grid-cols-2 gap-2 text-[#8a8a8a]">
                  <div><kbd className="text-[#fbbf24] font-bold">[ / ]</kbd> Focus username input</div>
                  <div><kbd className="text-[#fbbf24] font-bold">[ r ]</kbd> Refresh telemetry</div>
                  <div><kbd className="text-[#fbbf24] font-bold">[ e ]</kbd> Export portfolio</div>
                  <div><kbd className="text-[#fbbf24] font-bold">[ Esc ]</kbd> Close inspector</div>
                  <div><kbd className="text-[#fbbf24] font-bold">[ 1-4 ]</kbd> Switch views</div>
                  <div><kbd className="text-[#fbbf24] font-bold">[ ? ]</kbd> Toggle help</div>
                </div>
              </div>
              <p className="text-[10px] text-[#525252]">
                GITSTAT is built following strict developer minimalism: zero fluff, pure verifiable telemetry.
              </p>
            </div>
          )}

          {selectedItem.type === "export" && (
            <div className="flex flex-col gap-3">
              <div className="border border-[#1f1f1f] bg-[#080808] p-3">
                <div className="font-bold text-[#fafafa] mb-1.5">MARKDOWN BADGE SNIPPET</div>
                <div className="bg-[#121212] p-2 text-[11px] text-[#4ade80] border border-[#262626] select-all overflow-x-auto">
                  {`[![GitStat: ${profile.username}](https://img.shields.io/badge/GitStat-${profile.metrics.mergedPRs}%20Merged%20PRs-fbbf24?style=flat-square&logo=github)](https://gitwise.dev/stat/${profile.username})`}
                </div>
              </div>

              <div className="border border-[#1f1f1f] bg-[#080808] p-3">
                <div className="font-bold text-[#fafafa] mb-1.5">ASCII RESUME SUMMARY</div>
                <pre className="bg-[#121212] p-2 text-[10px] text-[#8a8a8a] border border-[#262626] overflow-x-auto select-all leading-tight">
{`---------------------------------------------------------
GITSTAT VERIFIED DEVELOPER SUMMARY: @${profile.username}
---------------------------------------------------------
Name: ${profile.name} (${profile.title})
Merged PRs: ${profile.metrics.mergedPRs} | Open: ${profile.metrics.openPRs}
Reviews Given: ${profile.metrics.codeReviewsGiven} (Avg Turnaround: ${profile.metrics.reviewTurnaroundHours}h)
Issues Opened: ${profile.metrics.issuesOpened} | Issues Linked to PRs: ${profile.metrics.issuesLinkedToMergedPRs}
Code Delta: +${profile.metrics.linesAdded.toLocaleString()} / -${profile.metrics.linesDeleted.toLocaleString()} lines
Active Repos: ${profile.metrics.activeRepositories} | Total Commits: ${profile.metrics.totalCommits.toLocaleString()}
---------------------------------------------------------`}
                </pre>
              </div>
            </div>
          )}
        </div>

        {/* Modal Footer */}
        <div className="border-t border-[#262626] bg-[#121212] px-4 py-2 flex justify-between items-center text-[10px] text-[#525252]">
          <span>Press [Esc] to exit</span>
          <button
            onClick={onClose}
            className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#8a8a8a] hover:text-[#fafafa] transition-colors"
          >
            CLOSE
          </button>
        </div>
      </div>
    </div>
  );
}
