"use client";

import React, { useState } from "react";
import { ContributorProfile } from "@/lib/mockData";

interface ActivityAndMetricsProps {
  profile: ContributorProfile;
  activeFilter: string;
  onFilterChange: (filter: string) => void;
}

export function ActivityAndMetrics({
  profile,
  activeFilter,
  onFilterChange,
}: ActivityAndMetricsProps) {
  const [hoveredDay, setHoveredDay] = useState<{
    date: string;
    commits: number;
    prs: number;
    reviews: number;
  } | null>(null);

  // Character mapping for heatmap intensity
  const getIntensityChar = (level: number) => {
    switch (level) {
      case 4:
        return "█";
      case 3:
        return "▓";
      case 2:
        return "▒";
      case 1:
        return "░";
      default:
        return ".";
    }
  };

  const getIntensityColor = (level: number) => {
    switch (level) {
      case 4:
        return "text-[#fbbf24] font-bold";
      case 3:
        return "text-[#d97706]";
      case 2:
        return "text-[#b45309]";
      case 1:
        return "text-[#78350f]";
      default:
        return "text-[#333333]";
    }
  };

  const months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
  const daysOfWeek = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

  return (
    <div className="flex flex-col gap-3">
      {/* Box 1: Contribution Activity Heatmap matching comp */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 text-xs">
        <div className="border-b border-[#262626] pb-2 mb-3 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="font-bold text-[#fafafa]">Contribution Activity Heatmap</span>
            <span className="text-[#525252]">|</span>
            <span className="text-[10px] text-[#fbbf24]">2026 AUDIT MATRIX</span>
          </div>
          <div className="text-[10px] text-[#8a8a8a]">
            {hoveredDay ? (
              <span className="text-[#fbbf24]">
                {hoveredDay.date}: {hoveredDay.commits} commits, {hoveredDay.prs} PRs, {hoveredDay.reviews} reviews
              </span>
            ) : (
              <span>Hover cell for date telemetry</span>
            )}
          </div>
        </div>

        {/* Heatmap Visual Matrix */}
        <div className="overflow-x-auto pb-1">
          {/* Months header */}
          <div className="flex pl-8 mb-1 text-[10px] text-[#525252] font-mono select-none">
            {months.map((m, idx) => (
              <div key={idx} className="w-[30px] sm:w-[38px] text-left">
                {m}
              </div>
            ))}
          </div>

          {/* Grid rows */}
          <div className="flex flex-col gap-1 font-mono text-[11px] select-none">
            {[1, 3, 5].map((dayIdx) => (
              <div key={dayIdx} className="flex items-center gap-1">
                <span className="w-7 text-[10px] text-[#525252]">{daysOfWeek[dayIdx]}</span>
                <span className="text-[#262626]">|</span>
                <div className="flex items-center gap-[2px]">
                  {profile.activityWeeks.slice(0, 48).map((week, wIdx) => {
                    const day = week.days[dayIdx];
                    const char = getIntensityChar(day?.level || 0);
                    const color = getIntensityColor(day?.level || 0);
                    return (
                      <span
                        key={wIdx}
                        onMouseEnter={() =>
                          setHoveredDay(
                            day
                              ? {
                                  date: day.date,
                                  commits: day.commits,
                                  prs: day.prs,
                                  reviews: day.reviews,
                                }
                              : null
                          )
                        }
                        onMouseLeave={() => setHoveredDay(null)}
                        className={`cursor-pointer hover:bg-[#262626] px-[1px] ${color} transition-colors`}
                      >
                        {char}
                      </span>
                    );
                  })}
                </div>
                <span className="text-[#262626]">||</span>
              </div>
            ))}
          </div>
        </div>

        {/* Heatmap Footer Legend matching comp */}
        <div className="mt-3 pt-2 border-t border-[#1f1f1f] flex items-center justify-between text-[10px] text-[#525252]">
          <span>&lt;- Working activity -&gt;</span>
          <div className="flex items-center gap-1.5 font-mono">
            <span>Low</span>
            <span className="text-[#333333]">.</span>
            <span className="text-[#78350f]">░</span>
            <span className="text-[#b45309]">▒</span>
            <span className="text-[#d97706]">▓</span>
            <span className="text-[#fbbf24] font-bold">█</span>
            <span>High</span>
          </div>
        </div>
      </div>

      {/* Box 2: Factual KPI Table matching comp */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 text-xs">
        <div className="border-b border-[#262626] pb-2 mb-3 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="font-bold text-[#fafafa]">Factual KPI</span>
            <span className="text-[#525252]">|</span>
            <span className="text-[10px] text-[#8a8a8a]">VERIFIED GITHUB TELEMETRY</span>
          </div>
          <span className="text-[10px] text-[#4ade80] font-mono">NO ESTIMATES</span>
        </div>

        {/* Tabular Matrix with Box-Drawing Styling */}
        <div className="overflow-x-auto">
          <table className="w-full text-left font-mono border-collapse border border-[#262626]">
            <thead>
              <tr className="bg-[#121212] text-[#8a8a8a] text-[11px] border-b border-[#262626]">
                <th className="py-1.5 px-3 border-r border-[#262626] font-medium">METRIC IDENTIFIER</th>
                <th className="py-1.5 px-3 border-r border-[#262626] font-medium text-right">COUNT</th>
                <th className="py-1.5 px-3 border-r border-[#262626] font-medium">METHODOLOGY / EVIDENCE</th>
                <th className="py-1.5 px-3 font-medium text-right">TREND</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1f1f1f] text-xs">
              <tr className="hover:bg-[#141414] transition-colors">
                <td className="py-2 px-3 border-r border-[#262626] font-medium text-[#fafafa]">
                  Merged PRs
                </td>
                <td className="py-2 px-3 border-r border-[#262626] text-right font-bold text-[#fbbf24]">
                  {profile.metrics.mergedPRs}
                </td>
                <td className="py-2 px-3 border-r border-[#262626] text-[#8a8a8a] text-[11px]">
                  Pull requests successfully merged into default branches
                </td>
                <td className="py-2 px-3 text-right text-[#4ade80] text-[11px]">
                  +{profile.metrics.mergeSuccessRatePct}% rate
                </td>
              </tr>

              <tr className="hover:bg-[#141414] transition-colors">
                <td className="py-2 px-3 border-r border-[#262626] font-medium text-[#fafafa]">
                  Code Reviews Given
                </td>
                <td className="py-2 px-3 border-r border-[#262626] text-right font-bold text-[#fbbf24]">
                  {profile.metrics.codeReviewsGiven}
                </td>
                <td className="py-2 px-3 border-r border-[#262626] text-[#8a8a8a] text-[11px]">
                  Submitted reviews ({profile.metrics.reviewCommentVolume} review comments across PRs)
                </td>
                <td className="py-2 px-3 text-right text-[#fafafa] text-[11px]">
                  {profile.metrics.reviewTurnaroundHours}h avg response
                </td>
              </tr>

              <tr className="hover:bg-[#141414] transition-colors bg-[#0e0e0e]">
                <td className="py-2 px-3 border-r border-[#262626] font-medium text-[#fafafa]">
                  Issues Opened
                </td>
                <td className="py-2 px-3 border-r border-[#262626] text-right font-bold text-[#fafafa]">
                  {profile.metrics.issuesOpened}
                </td>
                <td className="py-2 px-3 border-r border-[#262626] text-[#8a8a8a] text-[11px]">
                  Factual issue creation events authored by user
                </td>
                <td className="py-2 px-3 text-right text-[#525252] text-[11px]">
                  Documented
                </td>
              </tr>

              <tr className="hover:bg-[#141414] transition-colors bg-[#0e0e0e]">
                <td className="py-2 px-3 border-r border-[#262626] font-medium text-[#fafafa]">
                  Issues Participated In
                </td>
                <td className="py-2 px-3 border-r border-[#262626] text-right font-bold text-[#fafafa]">
                  {profile.metrics.issuesParticipatedIn}
                </td>
                <td className="py-2 px-3 border-r border-[#262626] text-[#8a8a8a] text-[11px]">
                  Discussions, technical triage, and reproduction reports
                </td>
                <td className="py-2 px-3 text-right text-[#525252] text-[11px]">
                  Verified
                </td>
              </tr>

              <tr className="hover:bg-[#141414] transition-colors bg-[#0e0e0e]">
                <td className="py-2 px-3 border-r border-[#262626] font-medium text-[#fafafa]">
                  Issues Linked to Merged PRs
                </td>
                <td className="py-2 px-3 border-r border-[#262626] text-right font-bold text-[#4ade80]">
                  {profile.metrics.issuesLinkedToMergedPRs}
                </td>
                <td className="py-2 px-3 border-r border-[#262626] text-[#8a8a8a] text-[11px]">
                  Issues resolved via direct commit/PR linkages (`fixes #id`, `closes #id`)
                </td>
                <td className="py-2 px-3 text-right text-[#4ade80] text-[11px]">
                  Direct closure
                </td>
              </tr>

              <tr className="hover:bg-[#141414] transition-colors">
                <td className="py-2 px-3 border-r border-[#262626] font-medium text-[#fafafa]">
                  Active Repositories
                </td>
                <td className="py-2 px-3 border-r border-[#262626] text-right font-bold text-[#fafafa]">
                  {profile.metrics.activeRepositories}
                </td>
                <td className="py-2 px-3 border-r border-[#262626] text-[#8a8a8a] text-[11px]">
                  Distinct codebases with verified contributions in past 12 months
                </td>
                <td className="py-2 px-3 text-right text-[#8a8a8a] text-[11px]">
                  Cross-org
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        {/* Factual Integrity Note */}
        <div className="mt-3 pt-2 border-t border-[#1f1f1f] text-[10px] text-[#525252] flex items-center justify-between">
          <span>* Lines of code are reported strictly as delta statistics, not velocity or developer score.</span>
          <span className="text-[#fbbf24]">GITWISE ENGINE V1.2</span>
        </div>
      </div>
    </div>
  );
}
