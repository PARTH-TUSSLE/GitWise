"use client";

import React, { useState } from "react";
import { RepoModel, FeatureTrace, FeatureTraceStep } from "@/lib/repoData";
import { ArrowRight, Play, CheckCircle2, FileCode, ExternalLink, BookOpen, AlertCircle } from "lucide-react";

interface ArchitectureAndTracerProps {
  repo: RepoModel;
  activeViewMode: "architecture" | "feature_trace";
  onSelectFile: (filePath: string) => void;
}

export function ArchitectureAndTracer({
  repo,
  activeViewMode,
  onSelectFile,
}: ArchitectureAndTracerProps) {
  const [selectedTraceId, setSelectedTraceId] = useState(repo.featureTraces[0]?.id || "");
  const [activeStepIndex, setActiveStepIndex] = useState(0);

  const currentTrace = repo.featureTraces.find((t) => t.id === selectedTraceId) || repo.featureTraces[0];
  const activeStep = currentTrace?.steps[activeStepIndex] || currentTrace?.steps[0];

  return (
    <div className="flex flex-col gap-3 font-mono text-xs">
      {/* 1. Feature Trace Section (Prominent in Feature Trace Mode or Dual Mode) */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3">
        <div className="border-b border-[#262626] pb-2 mb-3 flex flex-wrap items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <span className="font-bold text-[#fafafa]">FEATURE TRACE:</span>
            <span className="text-[#fbbf24] font-bold">{currentTrace?.name}</span>
          </div>

          {/* Trace Selector */}
          <div className="flex items-center gap-1">
            {repo.featureTraces.map((trace) => (
              <button
                key={trace.id}
                onClick={() => {
                  setSelectedTraceId(trace.id);
                  setActiveStepIndex(0);
                }}
                className={`px-2 py-0.5 text-[10px] border transition-colors ${
                  trace.id === currentTrace?.id
                    ? "border-[#fbbf24] bg-[#1f1a09] text-[#fbbf24]"
                    : "border-[#262626] text-[#8a8a8a] hover:border-[#404040] hover:text-[#fafafa]"
                }`}
              >
                {trace.name}
              </button>
            ))}
          </div>
        </div>

        <p className="text-[11px] text-[#8a8a8a] mb-3">
          {currentTrace?.description}
        </p>

        {/* Visual Topological Pipeline Flow */}
        <div className="border border-[#1f1f1f] bg-[#070707] p-3 mb-3 overflow-x-auto">
          <div className="flex items-center gap-2 min-w-[550px]">
            {currentTrace?.steps.map((step, idx) => {
              const isSelected = idx === activeStepIndex;
              return (
                <React.Fragment key={step.step}>
                  <button
                    onClick={() => setActiveStepIndex(idx)}
                    className={`flex-1 p-2 text-left border transition-all ${
                      isSelected
                        ? "border-[#fbbf24] bg-[#141207] shadow-sm"
                        : "border-[#262626] bg-[#0c0c0c] hover:border-[#404040]"
                    }`}
                  >
                    <div className="flex items-center justify-between text-[10px] mb-1">
                      <span className={isSelected ? "text-[#fbbf24] font-bold" : "text-[#525252]"}>
                        STAGE 0{step.step}
                      </span>
                      <span className="text-[9px] text-[#525252] truncate max-w-[90px]">{step.subsystem}</span>
                    </div>
                    <div className={`font-bold text-xs truncate mb-1 ${isSelected ? "text-[#fafafa]" : "text-[#8a8a8a]"}`}>
                      {step.title}
                    </div>
                    <div className="text-[10px] text-[#525252] font-mono truncate">
                      {step.file.split("/").pop()}:{step.line}
                    </div>
                  </button>

                  {idx < currentTrace.steps.length - 1 && (
                    <div className="text-[#fbbf24] font-bold select-none px-1">
                      ───►
                    </div>
                  )}
                </React.Fragment>
              );
            })}
          </div>
        </div>

        {/* Active Step Deep-Dive Inspector */}
        {activeStep && (
          <div className="border border-[#262626] bg-[#0e0e0e] p-3">
            <div className="flex items-center justify-between border-b border-[#1f1f1f] pb-2 mb-2">
              <div className="flex items-center gap-2">
                <span className="text-[#fbbf24] font-bold">STAGE {activeStep.step}:</span>
                <span className="text-[#fafafa] font-semibold">{activeStep.title}</span>
              </div>
              <button
                onClick={() => onSelectFile(activeStep.file)}
                className="text-[10px] text-[#fbbf24] hover:underline flex items-center gap-1"
              >
                <span>Inspect in file drawer</span>
                <ExternalLink className="w-3 h-3" />
              </button>
            </div>

            <div className="text-[11px] text-[#8a8a8a] mb-2 leading-relaxed">
              {activeStep.description}
            </div>

            {/* Code Snippet */}
            <div className="bg-[#050505] border border-[#1f1f1f] p-2.5 overflow-x-auto text-[11px]">
              <div className="text-[10px] text-[#525252] mb-1">
                {"//"} {activeStep.file} (Line {activeStep.line})
              </div>
              <pre className="text-[#fafafa] font-mono leading-relaxed">
                <code>{activeStep.codeSnippet}</code>
              </pre>
            </div>
          </div>
        )}
      </div>

      {/* 2. Architectural Subsystem Matrix & Contributor Onboarding Guide */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3">
        <div className="border-b border-[#262626] pb-2 mb-3 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="font-bold text-[#fafafa]">WHERE-TO-START GUIDE:</span>
            <span className="text-[#8a8a8a]">Contributor Onboarding</span>
          </div>
          <span className="text-[10px] text-[#fbbf24]">EVIDENCE GROUNDED</span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
          {/* Steps to start */}
          <div className="border border-[#1f1f1f] bg-[#0d0d0d] p-2.5">
            <div className="text-[11px] font-bold text-[#fafafa] mb-1.5 flex items-center gap-1.5">
              <Play className="w-3 h-3 text-[#fbbf24]" />
              Setup &amp; Verification Steps
            </div>
            <ol className="list-decimal list-inside space-y-1 text-[11px] text-[#8a8a8a]">
              {repo.contributorGuide.stepsToStart.map((step, idx) => (
                <li key={idx} className="leading-relaxed">
                  <span className="text-[#fafafa]">{step}</span>
                </li>
              ))}
            </ol>
          </div>

          {/* Architectural Prerequisites */}
          <div className="border border-[#1f1f1f] bg-[#0d0d0d] p-2.5">
            <div className="text-[11px] font-bold text-[#fafafa] mb-1.5 flex items-center gap-1.5">
              <BookOpen className="w-3 h-3 text-[#fbbf24]" />
              Architectural Prerequisites
            </div>
            <ul className="list-disc list-inside space-y-1 text-[11px] text-[#8a8a8a]">
              {repo.contributorGuide.prerequisites.map((req, idx) => (
                <li key={idx} className="leading-relaxed">
                  <span>{req}</span>
                </li>
              ))}
            </ul>
          </div>
        </div>

        {/* Beginner-friendly files */}
        <div>
          <div className="text-[10px] text-[#525252] uppercase mb-1.5">
            Recommended Starting Files for New Contributors
          </div>
          <div className="flex flex-col gap-1">
            {repo.contributorGuide.beginnerFiles.map((item) => (
              <button
                key={item.path}
                onClick={() => onSelectFile(item.path)}
                className="text-left p-1.5 border border-[#1f1f1f] hover:border-[#fbbf24] bg-[#0e0e0e] transition-colors group flex items-start justify-between gap-2"
              >
                <div>
                  <div className="text-[11px] font-medium text-[#fafafa] group-hover:text-[#fbbf24]">
                    {item.path}
                  </div>
                  <div className="text-[10px] text-[#8a8a8a]">
                    {item.reason}
                  </div>
                </div>
                <FileCode className="w-3.5 h-3.5 text-[#525252] group-hover:text-[#fbbf24] flex-shrink-0 mt-0.5" />
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
