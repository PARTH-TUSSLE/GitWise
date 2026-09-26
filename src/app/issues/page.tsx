"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { ISSUES_DATABASE, IssueModel } from "@/lib/issueData";
import { IssueCommandBar } from "@/components/issues/IssueCommandBar";
import { IssueTriageRail } from "@/components/issues/IssueTriageRail";
import { ContributionPipelineCanvas } from "@/components/issues/ContributionPipelineCanvas";
import { GroundedVerificationTerminal } from "@/components/issues/GroundedVerificationTerminal";
import { CodeDrawerModal } from "@/components/repo/CodeDrawerModal";

export default function IssuePlannerPage() {
  const router = useRouter();
  const [currentRepo, setCurrentRepo] = useState<string>("vercel/next.js");
  const [activeStage, setActiveStage] = useState<number>(1);
  const [selectedIssueId, setSelectedIssueId] = useState<string>("nextjs-54821");
  const [activeModalFile, setActiveModalFile] = useState<string | null>(null);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const repoIssues = ISSUES_DATABASE[currentRepo] || ISSUES_DATABASE["vercel/next.js"];
  const currentIssue = repoIssues.find((i) => i.id === selectedIssueId) || repoIssues[0];

  // Auto-select first issue when switching repos
  useEffect(() => {
    if (repoIssues.length > 0 && !repoIssues.find((i) => i.id === selectedIssueId)) {
      setSelectedIssueId(repoIssues[0].id);
    }
  }, [currentRepo, repoIssues, selectedIssueId]);

  // Keyboard shortcut listener
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }

      if (e.key === "Escape") {
        setActiveModalFile(null);
      } else if (e.key === "1") {
        setActiveStage(1);
      } else if (e.key === "2") {
        setActiveStage(2);
      } else if (e.key === "3") {
        setActiveStage(3);
      } else if (e.key === "g" || e.key === "G") {
        router.push("/");
      } else if (e.key === "r" || e.key === "R") {
        router.push("/repo");
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [router]);

  const handleSelectRepo = (repo: string) => {
    const matched = Object.keys(ISSUES_DATABASE).find(
      (k) => k.toLowerCase() === repo.toLowerCase()
    );
    if (matched) {
      setCurrentRepo(matched);
      setSelectedIssueId(ISSUES_DATABASE[matched][0].id);
    } else {
      setCurrentRepo(repo);
    }
    setIsRefreshing(true);
    setTimeout(() => setIsRefreshing(false), 250);
  };

  const handleSearchIssue = (query: string) => {
    const lower = query.toLowerCase();
    // Check if repo match
    const matchedRepo = Object.keys(ISSUES_DATABASE).find((k) =>
      lower.includes(k.toLowerCase())
    );
    if (matchedRepo) {
      setCurrentRepo(matchedRepo);
      // Check if issue number specified
      const numMatch = lower.match(/#?(\d+)/);
      if (numMatch) {
        const found = ISSUES_DATABASE[matchedRepo].find((i) => i.number === parseInt(numMatch[1], 10));
        if (found) setSelectedIssueId(found.id);
      }
      return;
    }

    // Keyword search in current repo
    const foundIssue = repoIssues.find(
      (i) =>
        i.title.toLowerCase().includes(lower) ||
        i.number.toString() === lower.replace("#", "")
    );
    if (foundIssue) {
      setSelectedIssueId(foundIssue.id);
    }
  };

  return (
    <main className="min-h-screen bg-[#050505] text-[#fafafa] font-mono p-2 sm:p-4 lg:p-6 flex flex-col gap-3 selection:bg-[#fbbf24] selection:text-[#050505]">
      {/* 1. Global Issue Command Bar */}
      <IssueCommandBar
        currentRepo={currentRepo}
        currentIssueNumber={currentIssue.number}
        activeStage={activeStage}
        onSelectStage={setActiveStage}
        onSelectRepo={handleSelectRepo}
        onSearchIssue={handleSearchIssue}
      />

      {/* 2. Main 3-Column Compiler Workbench Layout */}
      <div className={`grid grid-cols-1 lg:grid-cols-12 gap-3 flex-1 transition-opacity duration-200 ${isRefreshing ? "opacity-40" : "opacity-100"}`}>
        {/* Left Column: Issue Triage Stream & Blast Radius (cols 1-3) */}
        <aside className="lg:col-span-3 flex flex-col gap-3">
          <IssueTriageRail
            issues={repoIssues}
            selectedIssue={currentIssue}
            onSelectIssue={(issue) => {
              setSelectedIssueId(issue.id);
            }}
            onSelectFile={(filePath) => setActiveModalFile(filePath)}
          />
        </aside>

        {/* Center Column: Contribution Pipeline Canvas (cols 4-8) */}
        <section className="lg:col-span-6 flex flex-col gap-3">
          <ContributionPipelineCanvas
            issue={currentIssue}
            activeStage={activeStage}
            onSelectStage={setActiveStage}
            onSelectFile={(filePath) => setActiveModalFile(filePath)}
          />
        </section>

        {/* Right Column: Grounded Test Strategy & Verification Terminal (cols 9-12) */}
        <aside className="lg:col-span-3 flex flex-col gap-3">
          <GroundedVerificationTerminal
            issue={currentIssue}
            onSelectFile={(filePath) => setActiveModalFile(filePath)}
          />
        </aside>
      </div>

      {/* 3. Terminal Status Footer Bar */}
      <footer className="border border-[#262626] bg-[#0a0a0a] px-3 py-1.5 text-[10px] text-[#525252] flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-3">
          <span className="text-[#fbbf24] font-bold">GITWISE ISSUE PLANNER v1.4</span>
          <span>REPO: {currentRepo}</span>
          <span className="text-[#4ade80]">TARGET: #{currentIssue.number}</span>
          <span className="hidden sm:inline">BLAST RADIUS: {currentIssue.blastRadius.score.toFixed(1)}</span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-[#fbbf24] border border-[#333333] px-1 py-0.2">CLI: COMING SOON</span>
          <span className="text-[#4ade80]">TESTS: {currentIssue.testStrategy.suites.length} SUITES</span>
          <span>STAGE: 0{activeStage} ACTIVE</span>
          <span className="hidden sm:inline">[1-3] STAGES &middot; [R] REPO &middot; [G] GITSTAT</span>
        </div>
      </footer>

      {/* 4. Verifiable Code Drawer / Inspection Modal */}
      <CodeDrawerModal
        isOpen={Boolean(activeModalFile)}
        onClose={() => setActiveModalFile(null)}
        filePath={activeModalFile || ""}
        repoName={currentRepo}
      />
    </main>
  );
}
