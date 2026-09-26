"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { REPOSITORIES, RepoModel, SubsystemNode } from "@/lib/repoData";
import { fetchRepository } from "@/lib/api";
import { RepoCommandBar } from "@/components/repo/RepoCommandBar";
import { RepoTreeRail } from "@/components/repo/RepoTreeRail";
import { ArchitectureAndTracer } from "@/components/repo/ArchitectureAndTracer";
import { GroundedTerminal } from "@/components/repo/GroundedTerminal";
import { CodeDrawerModal } from "@/components/repo/CodeDrawerModal";

export default function RepoExplorerPage() {
  const router = useRouter();
  const [currentRepoKey, setCurrentRepoKey] = useState<string>("vercel/next.js");
  const [activeViewMode, setActiveViewMode] = useState<"architecture" | "feature_trace">("feature_trace");
  const [selectedSubsystem, setSelectedSubsystem] = useState<SubsystemNode | null>(null);
  const [activeModalFile, setActiveModalFile] = useState<string | null>(null);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const [currentRepo, setCurrentRepo] = useState<RepoModel>(
    REPOSITORIES[currentRepoKey] ||
    REPOSITORIES["vercel/next.js"]
  );

  useEffect(() => {
    let isCancelled = false;
    const parts = currentRepoKey.split("/");
    const owner = parts[0] || "vercel";
    const repo = parts[1] || "next.js";

    fetchRepository(owner, repo).then((data) => {
      if (!isCancelled && data) {
        setCurrentRepo(data);
      }
    });

    return () => {
      isCancelled = true;
    };
  }, [currentRepoKey]);

  // Reset subsystem when switching repos
  useEffect(() => {
    if (currentRepo && currentRepo.subsystems.length > 0) {
      setSelectedSubsystem(currentRepo.subsystems[0]);
    }
  }, [currentRepoKey, currentRepo]);

  // Keyboard shortcut listener
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }

      if (e.key === "Escape") {
        setActiveModalFile(null);
      } else if (e.key === "1") {
        setActiveViewMode("feature_trace");
      } else if (e.key === "2") {
        setActiveViewMode("architecture");
      } else if (e.key === "g" || e.key === "G") {
        router.push("/");
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [router]);

  const handleSelectRepo = (repoKey: string) => {
    const key = repoKey.trim().toLowerCase();
    const matched = Object.keys(REPOSITORIES).find((k) => k.toLowerCase() === key);
    if (matched) {
      setCurrentRepoKey(matched);
    } else {
      setCurrentRepoKey(repoKey);
    }
    setIsRefreshing(true);
    setTimeout(() => setIsRefreshing(false), 300);
  };

  return (
    <main className="min-h-screen bg-[#050505] text-[#fafafa] font-mono p-2 sm:p-4 lg:p-6 flex flex-col gap-3 selection:bg-[#fbbf24] selection:text-[#050505]">
      {/* 1. Global Repo Command Bar */}
      <RepoCommandBar
        currentRepoKey={currentRepoKey}
        onSelectRepo={handleSelectRepo}
        activeViewMode={activeViewMode}
        onSelectViewMode={setActiveViewMode}
        onSwitchToGitStat={() => router.push("/")}
        onHelp={() => setActiveModalFile("README.md")}
      />

      {/* 2. Main 3-Column Compiler Workbench Layout */}
      <div className={`grid grid-cols-1 lg:grid-cols-12 gap-3 flex-1 transition-opacity duration-200 ${isRefreshing ? "opacity-40" : "opacity-100"}`}>
        {/* Left Column: Repository Tree & Subsystem Directory Rail (cols 1-3) */}
        <aside className="lg:col-span-3 flex flex-col gap-3">
          <RepoTreeRail
            repo={currentRepo}
            selectedSubsystem={selectedSubsystem}
            onSelectSubsystem={setSelectedSubsystem}
            onSelectFile={(filePath) => setActiveModalFile(filePath)}
          />
        </aside>

        {/* Center Column: Topological Execution Pipeline & Feature Trace (cols 4-8) */}
        <section className="lg:col-span-6 flex flex-col gap-3">
          <ArchitectureAndTracer
            repo={currentRepo}
            activeViewMode={activeViewMode}
            onSelectFile={(filePath) => setActiveModalFile(filePath)}
          />
        </section>

        {/* Right Column: Grounded AI Analysis Terminal with Verifiable Citations (cols 9-12) */}
        <aside className="lg:col-span-3 flex flex-col gap-3">
          <GroundedTerminal
            repo={currentRepo}
            onSelectFile={(filePath) => setActiveModalFile(filePath)}
            onTraceTriggered={(traceId) => {
              setActiveViewMode("feature_trace");
            }}
          />
        </aside>
      </div>

      {/* 3. Terminal Status Footer Bar */}
      <footer className="border border-[#262626] bg-[#0a0a0a] px-3 py-1.5 text-[10px] text-[#525252] flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-3">
          <span className="text-[#fbbf24] font-bold">GITWISE REPO EXPLORER v2.4</span>
          <span>TARGET: {currentRepo.name}</span>
          <span className="hidden sm:inline">INDEXED: {currentRepo.indexedFiles} files</span>
          <span className="hidden sm:inline">LANG: {currentRepo.primaryLanguage}</span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-[#fbbf24] border border-[#333333] px-1 py-0.2">CLI: COMING SOON</span>
          <span className="text-[#4ade80]">AST COMPILED</span>
          <span>TOPOLOGY: ACTIVE</span>
          <span className="hidden sm:inline">[1] TRACER &middot; [2] ARCH &middot; [G] GITSTAT</span>
        </div>
      </footer>

      {/* 4. Verifiable Code Drawer / Inspection Modal */}
      <CodeDrawerModal
        isOpen={Boolean(activeModalFile)}
        onClose={() => setActiveModalFile(null)}
        filePath={activeModalFile || ""}
        repoName={currentRepo.name}
      />
    </main>
  );
}
