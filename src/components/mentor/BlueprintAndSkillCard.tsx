"use client";

import React from "react";
import Link from "next/link";
import { RepoMentorModel } from "@/lib/mentorData";
import { UserCheck, GitBranch, ArrowRight, Layers, ExternalLink } from "lucide-react";

interface BlueprintAndSkillCardProps {
  model: RepoMentorModel;
  onInspectFile: (filePath: string) => void;
}

export function BlueprintAndSkillCard({
  model,
  onInspectFile,
}: BlueprintAndSkillCardProps) {
  const match = model.contributorMatch;

  return (
    <div className="flex flex-col gap-3 font-mono text-xs">
      {/* 1. Target Repository Blueprint */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2">
        <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            <Layers className="w-3.5 h-3.5 text-[#fbbf24]" />
            <span className="font-bold text-[#fafafa] uppercase text-[11px]">
              REPO BLUEPRINT: {model.name}
            </span>
          </div>
          <span className="text-[10px] text-[#fbbf24]">{model.primaryLanguage}</span>
        </div>

        <p className="text-[11px] text-[#8a8a8a] leading-relaxed">
          {model.tagline}
        </p>

        <div className="border-t border-[#1a1a1a] pt-1.5 flex items-center justify-between text-[10px] text-[#525252]">
          <span>INDEXED AST: 1,489 FILES</span>
          <Link
            href="/repo"
            className="text-[#fbbf24] hover:underline flex items-center gap-1"
          >
            <span>Explore Tree</span>
            <ArrowRight className="w-2.5 h-2.5" />
          </Link>
        </div>
      </div>

      {/* 2. Contributor Alignment (GitStat Integration) */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2.5">
        <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            <UserCheck className="w-3.5 h-3.5 text-[#4ade80]" />
            <span className="font-bold text-[#fafafa] uppercase text-[11px]">
              CONTRIBUTOR ALIGNMENT
            </span>
          </div>
          <span className="text-[10px] text-[#4ade80] font-bold">
            {match.overallScore}% MATCH
          </span>
        </div>

        <p className="text-[10px] text-[#8a8a8a] leading-relaxed">
          {match.rationale}
        </p>

        {/* Matched Skills */}
        <div className="flex flex-col gap-1">
          <span className="text-[9px] text-[#525252] uppercase font-bold">
            DEMONSTRATED STRENGTHS (GITSTAT):
          </span>
          <div className="flex flex-wrap gap-1">
            {match.matchedSkills.map((skill, idx) => (
              <span
                key={idx}
                className="px-1.5 py-0.2 bg-[#0c180e] border border-[#166534] text-[#4ade80] text-[9px] font-bold"
              >
                &bull; {skill}
              </span>
            ))}
          </div>
        </div>

        {/* Growth Areas */}
        <div className="flex flex-col gap-1">
          <span className="text-[9px] text-[#525252] uppercase font-bold">
            GROWTH AREAS TO ACQUIRE:
          </span>
          <div className="flex flex-wrap gap-1">
            {match.skillsToAcquire.map((skill, idx) => (
              <span
                key={idx}
                className="px-1.5 py-0.2 bg-[#171308] border border-[#78350f] text-[#fbbf24] text-[9px]"
              >
                + {skill}
              </span>
            ))}
          </div>
        </div>
      </div>

      {/* 3. Skill-Matched Issues Recommendation */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2">
        <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between">
          <span className="font-bold text-[#fafafa] uppercase text-[11px]">
            MATCHED ISSUES ({model.matchedIssues.length})
          </span>
          <span className="text-[10px] text-[#525252]">HIGH FIT</span>
        </div>

        <div className="flex flex-col gap-2">
          {model.matchedIssues.map((issue) => (
            <div
              key={issue.issueNumber}
              className="border border-[#1f1f1f] bg-[#070707] p-2 flex flex-col gap-1.5"
            >
              <div className="flex items-center justify-between">
                <span className="font-bold text-[#fbbf24] text-[10px]">
                  #{issue.issueNumber}
                </span>
                <span className="text-[9px] text-[#4ade80] border border-[#4ade80]/30 px-1 font-bold">
                  {issue.matchScore}% MATCH
                </span>
              </div>

              <div className="text-[11px] text-[#fafafa] font-medium leading-snug line-clamp-2">
                {issue.title}
              </div>

              <p className="text-[9px] text-[#8a8a8a] leading-snug">
                {issue.matchRationale}
              </p>

              <div className="flex items-center justify-between border-t border-[#1a1a1a] pt-1 mt-0.5 text-[9px]">
                <span className="text-[#525252]">{issue.subsystem}</span>
                <Link
                  href="/issues"
                  className="text-[#fbbf24] hover:underline flex items-center gap-0.5"
                >
                  <span>Plan issue</span>
                  <ArrowRight className="w-2.5 h-2.5" />
                </Link>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
