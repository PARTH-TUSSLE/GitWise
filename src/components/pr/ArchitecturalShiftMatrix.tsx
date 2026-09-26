"use client";

import React from "react";
import { PullRequestModel } from "@/lib/prData";
import { ShieldCheck, AlertTriangle, ArrowRight, ExternalLink, Layers, GitCompare, FileCode, CheckCircle2 } from "lucide-react";

interface ArchitecturalShiftMatrixProps {
  pr: PullRequestModel;
  activeView: "architecture" | "diffs";
  selectedFile: string | null;
  onInspectFile: (filePath: string) => void;
}

export function ArchitecturalShiftMatrix({
  pr,
  activeView,
  selectedFile,
  onInspectFile,
}: ArchitecturalShiftMatrixProps) {
  const getContractBadge = (contract: "unchanged" | "extended" | "breaking") => {
    switch (contract) {
      case "breaking":
        return <span className="text-[#f87171] border border-[#f87171] bg-[#7f1d1d]/20 px-1.5 py-0.5 text-[10px] font-bold">BREAKING CONTRACT</span>;
      case "extended":
        return <span className="text-[#fbbf24] border border-[#fbbf24] bg-[#78350f]/20 px-1.5 py-0.5 text-[10px] font-bold">CONTRACT EXTENDED</span>;
      default:
        return <span className="text-[#4ade80] border border-[#4ade80] bg-[#14532d]/20 px-1.5 py-0.5 text-[10px] font-bold">CONTRACT UNCHANGED</span>;
    }
  };

  const displayedFiles = selectedFile
    ? pr.files.filter((f) => f.path === selectedFile)
    : pr.files;

  return (
    <div className="flex flex-col gap-3 font-mono text-xs">
      {/* 1. PR Executive Overview Banner */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3">
        <div className="border-b border-[#262626] pb-2 mb-2 flex flex-wrap items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <span className="font-bold text-[#fafafa]">PULL REQUEST:</span>
            <span className="text-[#fbbf24] font-bold">#{pr.number}</span>
          </div>
          <span className="text-[10px] text-[#8a8a8a]">
            CREATED: {pr.createdAt} BY @{pr.author}
          </span>
        </div>

        <h2 className="text-xs text-[#fafafa] font-bold mb-2 leading-snug">
          {pr.title}
        </h2>

        <p className="text-[11px] text-[#8a8a8a] leading-relaxed">
          {pr.summary}
        </p>
      </div>

      {/* 2. Architectural Impact Matrix View */}
      {activeView === "architecture" && (
        <div className="flex flex-col gap-3">
          {/* Public API Contract Shift Box */}
          <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2">
            <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between">
              <span className="font-bold text-[#fafafa] uppercase text-[11px]">
                PUBLIC API CONTRACT SHIFT
              </span>
              {getContractBadge(pr.architecturalShift.publicApiContract)}
            </div>
            <p className="text-[11px] text-[#fafafa] leading-relaxed">
              {pr.architecturalShift.publicApiExplanation}
            </p>
          </div>

          {/* Downstream Subsystems Impact Matrix */}
          <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2.5">
            <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between">
              <span className="font-bold text-[#fafafa] uppercase text-[11px]">
                DOWNSTREAM SUBSYSTEMS COUPLING ({pr.architecturalShift.downstreamSubsystems.length})
              </span>
              <span className="text-[10px] text-[#525252]">AST DEPENDENCY GRAPH</span>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
              {pr.architecturalShift.downstreamSubsystems.map((sub, idx) => (
                <div key={idx} className="border border-[#1f1f1f] bg-[#070707] p-2.5 flex flex-col gap-1.5">
                  <div className="flex items-center justify-between">
                    <span className="font-bold text-[#fafafa] text-[11px]">{sub.name}</span>
                    <span className={`text-[9px] px-1 py-0.2 font-bold border ${
                      sub.coupling === "tight"
                        ? "border-[#fbbf24] text-[#fbbf24] bg-[#1a1708]"
                        : "border-[#404040] text-[#8a8a8a]"
                    }`}>
                      {sub.coupling.toUpperCase()} COUPLING
                    </span>
                  </div>
                  <p className="text-[10px] text-[#8a8a8a] leading-snug">
                    {sub.impact}
                  </p>
                </div>
              ))}
            </div>
          </div>

          {/* Execution Flow Shift Delta */}
          <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2">
            <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between">
              <span className="font-bold text-[#fafafa] uppercase text-[11px]">
                EXECUTION FLOW SHIFT DELTA
              </span>
              <span className="text-[10px] text-[#4ade80]">RUNTIME VERIFIED</span>
            </div>
            <div className="border border-[#1a1a1a] bg-[#040404] p-2 text-[11px] text-[#fbbf24] font-mono leading-relaxed">
              {pr.architecturalShift.executionFlowDelta}
            </div>
          </div>

          {/* State Mutation Risk Guard */}
          <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-1.5">
            <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between">
              <span className="font-bold text-[#fafafa] uppercase text-[11px]">
                STATE MUTATION RISK GUARD
              </span>
              <span className="text-[10px] text-[#4ade80] uppercase">
                RISK: {pr.architecturalShift.stateMutationRisk}
              </span>
            </div>
            <p className="text-[10px] text-[#8a8a8a] leading-relaxed">
              {pr.architecturalShift.riskExplanation}
            </p>
          </div>
        </div>
      )}

      {/* 3. Behavioral File Diffs View */}
      {activeView === "diffs" && (
        <div className="flex flex-col gap-3">
          {displayedFiles.map((file) => (
            <div key={file.path} className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2">
              <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between gap-2">
                <div className="flex items-center gap-2 min-w-0">
                  <FileCode className="w-3.5 h-3.5 text-[#fbbf24] shrink-0" />
                  <span className="font-bold text-[#fafafa] text-[11px] truncate">
                    {file.path}
                  </span>
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  <span className="text-[#4ade80] text-[10px]">+{file.additions}</span>
                  <span className="text-[#f87171] text-[10px]">-{file.deletions}</span>
                  <button
                    onClick={() => onInspectFile(file.path)}
                    className="flex items-center gap-1 text-[10px] text-[#8a8a8a] hover:text-[#fbbf24] transition-colors"
                  >
                    <span>Inspect</span>
                    <ExternalLink className="w-2.5 h-2.5" />
                  </button>
                </div>
              </div>

              {/* Behavioral Change Summary */}
              <div className="border-l-2 border-[#fbbf24] bg-[#0f0e08] p-2 text-[10px] text-[#8a8a8a] leading-snug">
                <span className="text-[#fbbf24] font-bold uppercase">BEHAVIORAL SHIFT: </span>
                {file.behavioralSummary}
              </div>

              {/* Syntax-Highlighted Diff Viewer */}
              <div className="bg-[#030303] border border-[#1a1a1a] p-2.5 overflow-x-auto text-[10px] font-mono leading-snug flex flex-col gap-1">
                {file.diffSnippet.before && (
                  <div className="text-[#f87171] opacity-75">
                    <span className="select-none mr-2">-</span>
                    {file.diffSnippet.before}
                  </div>
                )}
                <div className="text-[#4ade80]">
                  <span className="select-none mr-2">+</span>
                  {file.diffSnippet.after}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
