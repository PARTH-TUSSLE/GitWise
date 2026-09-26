"use client";

import React from "react";
import { MentorshipTrack, RepoMentorModel } from "@/lib/mentorData";
import { Compass, BookOpen, CheckCircle, FileCode, ExternalLink, ArrowRight, ShieldCheck } from "lucide-react";

interface MentorshipTracksRailProps {
  model: RepoMentorModel;
  onSelectTrack: (track: MentorshipTrack) => void;
  onInspectFile: (filePath: string) => void;
}

export function MentorshipTracksRail({
  model,
  onSelectTrack,
  onInspectFile,
}: MentorshipTracksRailProps) {
  return (
    <div className="flex flex-col gap-3 font-mono text-xs">
      {/* 1. Guided Mentorship Tracks */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2.5">
        <div className="border-b border-[#262626] pb-2 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Compass className="w-3.5 h-3.5 text-[#fbbf24]" />
            <span className="font-bold text-[#fafafa] uppercase tracking-wider text-[11px]">
              MENTORSHIP TRACKS
            </span>
          </div>
          <span className="text-[10px] text-[#525252]">SOCRATIC PATHS</span>
        </div>

        <div className="flex flex-col gap-1.5">
          {model.tracks.map((track) => (
            <button
              key={track.id}
              onClick={() => onSelectTrack(track)}
              className="p-2.5 text-left border border-[#1f1f1f] bg-[#0c0c0c] hover:border-[#fbbf24] hover:bg-[#121212] transition-all flex flex-col gap-1 group"
            >
              <div className="flex items-center justify-between gap-1">
                <span className="font-bold text-[#fafafa] group-hover:text-[#fbbf24] text-[11px]">
                  {track.title}
                </span>
                <span className="text-[9px] text-[#fbbf24] border border-[#fbbf24]/30 px-1 py-0.2 font-bold shrink-0">
                  {track.badge}
                </span>
              </div>
              <p className="text-[10px] text-[#8a8a8a] leading-snug line-clamp-2">
                {track.description}
              </p>
            </button>
          ))}
        </div>
      </div>

      {/* 2. Onboarding Milestones */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2">
        <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            <ShieldCheck className="w-3.5 h-3.5 text-[#fbbf24]" />
            <span className="font-bold text-[#fafafa] uppercase text-[11px]">
              ONBOARDING MILESTONES
            </span>
          </div>
          <span className="text-[10px] text-[#4ade80]">3 / 4 VERIFIED</span>
        </div>

        <div className="flex flex-col gap-1 text-[10px] text-[#8a8a8a]">
          <div className="flex items-center gap-2 p-1 text-[#4ade80]">
            <CheckCircle className="w-3 h-3 text-[#4ade80] shrink-0" />
            <span>Monorepo workspace cloned & dependencies installed</span>
          </div>
          <div className="flex items-center gap-2 p-1 text-[#4ade80]">
            <CheckCircle className="w-3 h-3 text-[#4ade80] shrink-0" />
            <span>Rust native bindings compiled via build-native</span>
          </div>
          <div className="flex items-center gap-2 p-1 text-[#4ade80]">
            <CheckCircle className="w-3 h-3 text-[#4ade80] shrink-0" />
            <span>Unit test suite executed and passing</span>
          </div>
          <div className="flex items-center gap-2 p-1 text-[#fbbf24]">
            <span className="w-3 h-3 border border-[#fbbf24] text-[#fbbf24] flex items-center justify-center text-[9px] shrink-0">4</span>
            <span className="text-[#fafafa]">Select first skill-matched issue</span>
          </div>
        </div>
      </div>

      {/* 3. Recommended Starter Modules */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2">
        <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            <BookOpen className="w-3.5 h-3.5 text-[#fbbf24]" />
            <span className="font-bold text-[#fafafa] uppercase text-[11px]">
              STARTER MODULES
            </span>
          </div>
          <span className="text-[10px] text-[#525252]">HIGH TEST COVERAGE</span>
        </div>

        <div className="flex flex-col gap-1.5">
          <div className="border border-[#1a1a1a] bg-[#080808] p-2 flex flex-col gap-1">
            <div className="flex items-center justify-between">
              <span className="font-bold text-[#fafafa] text-[10px]">packages/next/src/client/link.tsx</span>
              <button
                onClick={() => onInspectFile("packages/next/src/client/link.tsx")}
                className="text-[#8a8a8a] hover:text-[#fbbf24]"
              >
                <ExternalLink className="w-2.5 h-2.5" />
              </button>
            </div>
            <p className="text-[9px] text-[#8a8a8a] leading-snug">
              Prefetching & client navigation logic. Zero compiler coupling.
            </p>
            <div className="flex items-center gap-2 text-[9px] text-[#4ade80]">
              <span>94% Test Coverage</span>
              <span className="text-[#525252]">|</span>
              <span className="text-[#8a8a8a]">Isolated Client Module</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
