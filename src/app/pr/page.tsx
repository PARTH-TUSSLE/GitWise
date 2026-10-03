"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { PR_DATABASE, PullRequestModel } from "@/lib/prData";
import { fetchPullRequests } from "@/lib/api";
import { PRCommandBar } from "@/components/pr/PRCommandBar";
import { PRRail } from "@/components/pr/PRRail";
import { ArchitecturalShiftMatrix } from "@/components/pr/ArchitecturalShiftMatrix";
import { GroundedReviewTerminal } from "@/components/pr/GroundedReviewTerminal";
import { CodeDrawerModal } from "@/components/repo/CodeDrawerModal";

export default function PRReviewerPage() {
  const router = useRouter();
  const [currentRepo, setCurrentRepo] = useState<string>("vercel/next.js");
  const [activeView, setActiveView] = useState<"architecture" | "diffs">("architecture");
  const [selectedPRId, setSelectedPRId] = useState<string>("next-pr-62145");
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [activeModalFile, setActiveModalFile] = useState<string | null>(null);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const [repoPRs, setRepoPRs] = useState<PullRequestModel[]>(
    PR_DATABASE[currentRepo] || PR_DATABASE["vercel/next.js"]
  );
  const currentPR = repoPRs.find((p) => p.id === selectedPRId) || repoPRs[0] || PR_DATABASE["vercel/next.js"][0];

  useEffect(() => {
    let isCancelled = false;
    const parts = currentRepo.split("/");
    const owner = parts[0] || "vercel";
    const repo = parts[1] || "next.js";

    fetchPullRequests(owner, repo).then((data) => {
      if (!isCancelled && data && data.length > 0) {
        setRepoPRs(data);
      }
    });

    return () => {
      isCancelled = true;
    };
  }, [currentRepo]);

  // Reset selected PR when repo changes
  useEffect(() => {
    if (repoPRs.length > 0 && !repoPRs.find((p) => p.id === selectedPRId)) {
      setSelectedPRId(repoPRs[0].id);
      setSelectedFile(null);
    }
  }, [currentRepo, repoPRs, selectedPRId]);

  // Keyboard shortcut listener
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }

      if (e.key === "Escape") {
        setActiveModalFile(null);
      } else if (e.key === "1") {
        setActiveView("architecture");
      } else if (e.key === "2") {
        setActiveView("diffs");
      } else if (e.key === "i" || e.key === "I") {
        router.push("/issues");
      } else if (e.key === "r" || e.key === "R") {
        router.push("/repo");
      } else if (e.key === "g" || e.key === "G") {
        router.push("/");
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [router]);

  const handleSelectRepo = (repo: string) => {
    const matched = Object.keys(PR_DATABASE).find(
      (k) => k.toLowerCase() === repo.toLowerCase()
    );
    if (matched) {
      setCurrentRepo(matched);
      setSelectedPRId(PR_DATABASE[matched][0].id);
    } else {
      setCurrentRepo(repo);
    }
    setIsRefreshing(true);
    setTimeout(() => setIsRefreshing(false), 250);
  };

  const handleSearchPR = (query: string) => {
    const lower = query.toLowerCase();
    const matchedRepo = Object.keys(PR_DATABASE).find((k) =>
      lower.includes(k.toLowerCase())
    );
    if (matchedRepo) {
      setCurrentRepo(matchedRepo);
      const numMatch = lower.match(/#?(\d+)/);
      if (numMatch) {
        const found = PR_DATABASE[matchedRepo].find((p) => p.number === parseInt(numMatch[1], 10));
        if (found) setSelectedPRId(found.id);
      }
      return;
    }

    const foundPR = repoPRs.find(
      (p) =>
        p.title.toLowerCase().includes(lower) ||
        p.number.toString() === lower.replace("#", "")
    );
    if (foundPR) {
      setSelectedPRId(foundPR.id);
    }
  };

  return (
    <main className="min-h-screen bg-[#050505] text-[#fafafa] font-mono p-2 sm:p-4 lg:p-6 flex flex-col gap-3 selection:bg-[#fbbf24] selection:text-[#050505]">
      {/* 1. Global PR Command Bar */}
      <PRCommandBar
        currentRepo={currentRepo}
        currentPRNumber={currentPR.number}
        activeView={activeView}
        onSelectView={setActiveView}
        onSelectRepo={handleSelectRepo}
        onSearchPR={handleSearchPR}
        stats={currentPR.stats}
      />

      {/* 2. Main 3-Column Reviewer Workbench Layout */}
      <div className={`grid grid-cols-1 lg:grid-cols-12 gap-3 flex-1 transition-opacity duration-200 ${isRefreshing ? "opacity-40" : "opacity-100"}`}>
        {/* Left Column: PR Stream & Modified File Tree (cols 1-3) */}
        <aside className="lg:col-span-3 flex flex-col gap-3">
          <PRRail
            prs={repoPRs}
            selectedPR={currentPR}
            onSelectPR={(pr) => {
              setSelectedPRId(pr.id);
              setSelectedFile(null);
            }}
            selectedFile={selectedFile}
            onSelectFile={(filePath) => {
              setSelectedFile(filePath === selectedFile ? null : filePath);
              setActiveView("diffs");
            }}
            onInspectFile={(filePath) => setActiveModalFile(filePath)}
          />
        </aside>

        {/* Center Column: Executive Architectural Shift Matrix & Diffs (cols 4-8) */}
        <section className="lg:col-span-6 flex flex-col gap-3">
          <ArchitecturalShiftMatrix
            pr={currentPR}
            activeView={activeView}
            selectedFile={selectedFile}
            onInspectFile={(filePath) => setActiveModalFile(filePath)}
          />
        </section>

        {/* Right Column: Grounded AI Code Review & Verification Terminal (cols 9-12) */}
        <aside className="lg:col-span-3 flex flex-col gap-3">
          <GroundedReviewTerminal
            pr={currentPR}
            onInspectFile={(filePath) => setActiveModalFile(filePath)}
          />
        </aside>
      </div>

      {/* 3. Terminal Status Footer Bar */}
      <footer className="border border-[#262626] bg-[#0a0a0a] px-3 py-1.5 text-[10px] text-[#525252] flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-3">
          <span className="text-[#fbbf24] font-bold">GITWISE PR REVIEWER v2.1</span>
          <span>REPO: {currentRepo}</span>
          <span className="text-[#4ade80]">PR #{currentPR.number}</span>
          <span className="hidden sm:inline">CONTRACT: {currentPR.architecturalShift.publicApiContract.toUpperCase()}</span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-[#fbbf24] border border-[#333333] px-1 py-0.2">CLI: COMING SOON</span>
          <span className="text-[#4ade80]">CI: 100% PASS</span>
          <span>VIEW: {activeView.toUpperCase()}</span>
          <span className="hidden sm:inline">[1] ARCH &middot; [2] DIFFS &middot; [I] ISSUES &middot; [R] REPO &middot; [G] GITSTAT</span>
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
