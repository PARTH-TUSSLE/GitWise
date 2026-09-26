"use client";

import React from "react";
import { IssueModel, ImplementationStep } from "@/lib/issueData";
import { ArrowRight, CheckCircle2, AlertTriangle, Code2, ExternalLink, GitCommit, ShieldAlert } from "lucide-react";

interface ContributionPipelineCanvasProps {
  issue: IssueModel;
  activeStage: number;
  onSelectStage: (stage: number) => void;
  onSelectFile: (filePath: string) => void;
}

export function ContributionPipelineCanvas({
  issue,
  activeStage,
  onSelectStage,
  onSelectFile,
}: ContributionPipelineCanvasProps) {
  return (
    <div className="flex flex-col gap-3 font-mono text-xs">
      {/* 1. Topological Contribution Pipeline Header */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3">
        <div className="border-b border-[#262626] pb-2 mb-3 flex flex-wrap items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <span className="font-bold text-[#fafafa]">CONTRIBUTION PIPELINE:</span>
            <span className="text-[#fbbf24] font-bold">#{issue.number}</span>
          </div>
          <span className="text-[10px] text-[#8a8a8a]">
            REPORTED: {issue.reportedDate} BY @{issue.author}
          </span>
        </div>

        <p className="text-[11px] text-[#8a8a8a] mb-3 leading-relaxed">
          {issue.summary}
        </p>

        {/* 3-Stage Topological Flow */}
        <div className="border border-[#1f1f1f] bg-[#070707] p-3 mb-1 overflow-x-auto">
          <div className="flex items-center gap-2 min-w-[540px]">
            {/* Stage 1 */}
            <button
              onClick={() => onSelectStage(1)}
              className={`flex-1 p-2 text-left border transition-all ${
                activeStage === 1
                  ? "border-[#fbbf24] bg-[#1a1708]"
                  : "border-[#262626] bg-[#0d0d0d] hover:border-[#404040]"
              }`}
            >
              <div className="text-[9px] text-[#fbbf24] font-bold uppercase mb-0.5">STAGE 01</div>
              <div className="font-bold text-[#fafafa] text-[11px] truncate">Issue Triage</div>
              <div className="text-[9px] text-[#8a8a8a] truncate">Root cause & repro</div>
            </button>

            <ArrowRight className="w-3.5 h-3.5 text-[#525252] shrink-0" />

            {/* Stage 2 */}
            <button
              onClick={() => onSelectStage(2)}
              className={`flex-1 p-2 text-left border transition-all ${
                activeStage === 2
                  ? "border-[#fbbf24] bg-[#1a1708]"
                  : "border-[#262626] bg-[#0d0d0d] hover:border-[#404040]"
              }`}
            >
              <div className="text-[9px] text-[#fbbf24] font-bold uppercase mb-0.5">STAGE 02</div>
              <div className="font-bold text-[#fafafa] text-[11px] truncate">Impacted Code Paths</div>
              <div className="text-[9px] text-[#8a8a8a] truncate">Call chains & symbols</div>
            </button>

            <ArrowRight className="w-3.5 h-3.5 text-[#525252] shrink-0" />

            {/* Stage 3 */}
            <button
              onClick={() => onSelectStage(3)}
              className={`flex-1 p-2 text-left border transition-all ${
                activeStage === 3
                  ? "border-[#fbbf24] bg-[#1a1708]"
                  : "border-[#262626] bg-[#0d0d0d] hover:border-[#404040]"
              }`}
            >
              <div className="text-[9px] text-[#fbbf24] font-bold uppercase mb-0.5">STAGE 03</div>
              <div className="font-bold text-[#fafafa] text-[11px] truncate">Implementation Blueprint</div>
              <div className="text-[9px] text-[#8a8a8a] truncate">{issue.stages.stage03_blueprint.steps.length} sequenced steps</div>
            </button>
          </div>
        </div>
      </div>

      {/* 2. Active Stage Detail Display */}
      {activeStage === 1 && (
        <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-3">
          <div className="border-b border-[#262626] pb-2 flex items-center justify-between">
            <span className="font-bold text-[#fbbf24] text-[11px] uppercase">
              STAGE 01: ISSUE TRIAGE & ROOT CAUSE ANALYSIS
            </span>
            <span className="text-[10px] text-[#525252]">DIAGNOSTIC EVIDENCE</span>
          </div>

          <div className="border border-[#1f1f1f] bg-[#070707] p-2.5">
            <span className="text-[#8a8a8a] text-[10px] uppercase font-bold block mb-1">
              ROOT CAUSE EXPLANATION:
            </span>
            <p className="text-[11px] text-[#fafafa] leading-relaxed">
              {issue.stages.stage01_triage.rootCauseAnalysis}
            </p>
          </div>

          <div className="border border-[#1f1f1f] bg-[#070707] p-2.5">
            <span className="text-[#8a8a8a] text-[10px] uppercase font-bold block mb-1.5">
              REPRODUCTION SEQUENCE:
            </span>
            <ol className="list-decimal list-inside flex flex-col gap-1 text-[11px] text-[#fafafa]">
              {issue.stages.stage01_triage.reproductionSteps.map((step, idx) => (
                <li key={idx} className="leading-snug">{step}</li>
              ))}
            </ol>
          </div>

          <div className="border-l-2 border-[#fbbf24] bg-[#0e0e0e] p-2 text-[10px] text-[#8a8a8a]">
            <span className="text-[#fbbf24] font-bold">SCOPE BOUNDARY: </span>
            {issue.stages.stage01_triage.scopeBoundary}
          </div>
        </div>
      )}

      {activeStage === 2 && (
        <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-3">
          <div className="border-b border-[#262626] pb-2 flex items-center justify-between">
            <span className="font-bold text-[#fbbf24] text-[11px] uppercase">
              STAGE 02: IMPACTED CODE PATHS & CALL CHAIN
            </span>
            <span className="text-[10px] text-[#4ade80]">TOPOLOGICAL FLOW</span>
          </div>

          <div className="border border-[#1f1f1f] bg-[#070707] p-2.5">
            <span className="text-[#8a8a8a] text-[10px] uppercase font-bold block mb-2">
              EXECUTION CALL CHAIN:
            </span>
            <div className="flex flex-col gap-1.5">
              {issue.stages.stage02_impacted_paths.callChain.map((call, idx) => {
                const [filePath, func] = call.split(":");
                return (
                  <div key={idx} className="flex items-center gap-2 text-[11px]">
                    <span className="text-[#525252] text-[10px] w-4">{idx + 1}.</span>
                    <button
                      onClick={() => onSelectFile(filePath)}
                      className="text-[#fbbf24] hover:underline"
                    >
                      {filePath}
                    </button>
                    <span className="text-[#525252]">→</span>
                    <code className="text-[#fafafa] bg-[#141414] px-1 py-0.2">{func}</code>
                  </div>
                );
              })}
            </div>
          </div>

          <div className="border border-[#1f1f1f] bg-[#070707] p-2.5">
            <span className="text-[#8a8a8a] text-[10px] uppercase font-bold block mb-1">
              CRITICAL SYMBOLS & INTERFACES:
            </span>
            <div className="flex flex-wrap gap-1 mt-1">
              {issue.stages.stage02_impacted_paths.criticalSymbols.map((sym, idx) => (
                <span
                  key={idx}
                  className="border border-[#333333] bg-[#121212] px-1.5 py-0.5 text-[10px] text-[#fafafa]"
                >
                  {sym}
                </span>
              ))}
            </div>
          </div>

          <div className="border-l-2 border-[#4ade80] bg-[#0e0e0e] p-2 text-[10px] text-[#8a8a8a]">
            <span className="text-[#4ade80] font-bold">STATE MUTATION GUARANTEE: </span>
            {issue.stages.stage02_impacted_paths.stateMutations}
          </div>
        </div>
      )}

      {activeStage === 3 && (
        <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-3">
          <div className="border-b border-[#262626] pb-2 flex items-center justify-between">
            <span className="font-bold text-[#fbbf24] text-[11px] uppercase">
              STAGE 03: STEP-BY-STEP IMPLEMENTATION BLUEPRINT
            </span>
            <span className="text-[10px] text-[#8a8a8a]">GROUNDED DIFF SPECIFICATION</span>
          </div>

          <div className="flex flex-col gap-3">
            {issue.stages.stage03_blueprint.steps.map((step) => (
              <div key={step.stepNumber} className="border border-[#1f1f1f] bg-[#070707] p-2.5">
                <div className="flex items-center justify-between gap-2 border-b border-[#1c1c1c] pb-1.5 mb-2">
                  <div className="flex items-center gap-2">
                    <span className="bg-[#fbbf24] text-[#050505] font-bold px-1 text-[10px]">
                      STEP {step.stepNumber}
                    </span>
                    <span className="font-bold text-[#fafafa] text-[11px]">{step.title}</span>
                  </div>
                  <button
                    onClick={() => onSelectFile(step.file)}
                    className="flex items-center gap-1 text-[10px] text-[#8a8a8a] hover:text-[#fbbf24] transition-colors"
                  >
                    <span>{step.file.split("/").pop()}</span>
                    <ExternalLink className="w-2.5 h-2.5" />
                  </button>
                </div>

                <p className="text-[10px] text-[#8a8a8a] mb-2 leading-relaxed">
                  {step.explanation}
                </p>

                {/* Diff Viewer */}
                <div className="bg-[#030303] border border-[#1a1a1a] p-2 overflow-x-auto text-[10px] font-mono leading-tight">
                  {step.diff.before && (
                    <div className="text-[#f87171] mb-1 line-through opacity-70">
                      <span className="select-none mr-2">-</span>
                      {step.diff.before}
                    </div>
                  )}
                  <div className="text-[#4ade80]">
                    <span className="select-none mr-2">+</span>
                    {step.diff.after}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 3. Architectural Prerequisites & Environment Setup */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3">
        <div className="border-b border-[#262626] pb-1.5 mb-2 flex items-center justify-between">
          <span className="font-bold text-[#fafafa] text-[11px] uppercase">
            ARCHITECTURAL PREREQUISITES
          </span>
          <span className="text-[10px] text-[#525252]">READ BEFORE MODIFYING</span>
        </div>
        <ul className="flex flex-col gap-1 text-[11px] text-[#8a8a8a]">
          {issue.prerequisites.map((prereq, idx) => (
            <li key={idx} className="flex items-start gap-1.5">
              <span className="text-[#fbbf24] select-none">&bull;</span>
              <span>{prereq}</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
