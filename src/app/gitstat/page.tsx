"use client";

import React, { useState, useEffect } from "react";
import { CONTRIBUTORS, ContributorProfile } from "@/lib/mockData";
import { fetchContributorProfile } from "@/lib/api";
import { CommandPrompt } from "@/components/CommandPrompt";
import { CommandRail } from "@/components/CommandRail";
import { ActivityAndMetrics } from "@/components/ActivityAndMetrics";
import { DiffStream } from "@/components/DiffStream";
import { EvidenceModal } from "@/components/EvidenceModal";

export default function GitStatPage() {
  const [currentUsername, setCurrentUsername] = useState("alexR_dev");
  const [activeTab, setActiveTab] = useState("overview");
  const [activeFilter, setActiveFilter] = useState("all");
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [modalState, setModalState] = useState<{
    isOpen: boolean;
    type: "repo" | "diff" | "help" | "export";
    id: string;
  }>({
    isOpen: false,
    type: "diff",
    id: "",
  });

  const [profile, setProfile] = useState<ContributorProfile>(
    CONTRIBUTORS[currentUsername.toLowerCase()] ||
    CONTRIBUTORS["alexr_dev"] ||
    CONTRIBUTORS["alexR_dev"]
  );

  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    let isCancelled = false;
    setLoadError(null);
    fetchContributorProfile(currentUsername)
      .then((data) => {
        if (!isCancelled && data) {
          setProfile(data);
        }
      })
      .catch((err) => {
        if (!isCancelled) {
          setLoadError(err instanceof Error ? err.message : "Failed to load telemetry");
        }
      });
    return () => {
      isCancelled = true;
    };
  }, [currentUsername]);

  const handleRefresh = React.useCallback(() => {
    setIsRefreshing(true);
    setLoadError(null);
    fetchContributorProfile(currentUsername)
      .then((data) => {
        if (data) setProfile(data);
      })
      .catch((err) => {
        setLoadError(err instanceof Error ? err.message : "Failed to refresh telemetry");
      })
      .finally(() => {
        setTimeout(() => setIsRefreshing(false), 400);
      });
  }, [currentUsername]);

  // Keyboard shortcut listener
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }

      if (e.key === "Escape") {
        setModalState((prev) => ({ ...prev, isOpen: false }));
      } else if (e.key === "r" || e.key === "R") {
        handleRefresh();
      } else if (e.key === "e" || e.key === "E") {
        setModalState({ isOpen: true, type: "export", id: "export" });
      } else if (e.key === "?") {
        setModalState({ isOpen: true, type: "help", id: "help" });
      } else if (e.key === "1") {
        setActiveTab("overview");
      } else if (e.key === "2") {
        setActiveTab("merged_prs");
      } else if (e.key === "3") {
        setActiveTab("reviews");
      } else if (e.key === "4") {
        setActiveTab("issues");
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleRefresh]);

  const handleSelectUser = (username: string) => {
    const matched = Object.keys(CONTRIBUTORS).find(
      (k) => k.toLowerCase() === username.toLowerCase()
    );
    if (matched) {
      setCurrentUsername(matched);
    } else {
      setCurrentUsername(username);
    }
  };

  return (
    <main className="min-h-screen bg-[#050505] text-[#fafafa] font-mono p-2 sm:p-4 lg:p-6 flex flex-col gap-3 selection:bg-[#fbbf24] selection:text-[#050505]">
      {/* 1. Global Terminal Command Prompt */}
      <CommandPrompt
        currentUsername={currentUsername}
        onSelectUser={handleSelectUser}
        onRefresh={handleRefresh}
        onExport={() => setModalState({ isOpen: true, type: "export", id: "export" })}
        onHelp={() => setModalState({ isOpen: true, type: "help", id: "help" })}
      />

      {/* Live API Error Notice */}
      {loadError && (
        <div className="border border-[#ef4444] bg-[#ef4444]/10 text-[#fca5a5] px-3 py-2 text-xs flex items-center justify-between">
          <span>[LIVE API ERROR] {loadError}</span>
          <button onClick={() => setLoadError(null)} className="text-[#a3a3a3] hover:text-[#fafafa]">
            ✕
          </button>
        </div>
      )}

      {/* 2. Main Multi-Column Command Workbench Layout (matching Comp 2) */}
      <div className={`grid grid-cols-1 lg:grid-cols-12 gap-3 flex-1 transition-opacity duration-200 ${isRefreshing ? "opacity-40" : "opacity-100"}`}>
        {/* Left Column: Command & Navigation Rail (cols 1-3) */}
        <aside className="lg:col-span-3 flex flex-col gap-3">
          <CommandRail
            profile={profile}
            activeTab={activeTab}
            onSelectTab={setActiveTab}
            onInspectRepo={(repoName) =>
              setModalState({ isOpen: true, type: "repo", id: repoName })
            }
          />
        </aside>

        {/* Center Column: Contribution Activity Heatmap & Factual KPI Grid (cols 4-8) */}
        <section className="lg:col-span-6 flex flex-col gap-3">
          <ActivityAndMetrics
            profile={profile}
            activeFilter={activeFilter}
            onFilterChange={setActiveFilter}
          />
        </section>

        {/* Right Column: Live Stream of PR Diff Stats (cols 9-12) */}
        <aside className="lg:col-span-3 flex flex-col gap-3">
          <DiffStream
            profile={profile}
            onSelectDiff={(diffId) =>
              setModalState({ isOpen: true, type: "diff", id: diffId })
            }
          />
        </aside>
      </div>

      {/* 3. Terminal Status Footer Bar */}
      <footer className="border border-[#262626] bg-[#0a0a0a] px-3 py-1.5 text-[10px] text-[#525252] flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-3">
          <span className="text-[#fbbf24] font-bold">GITWISE GITSTAT v1.2</span>
          <span>TTY: /dev/pts/1</span>
          <span className="hidden sm:inline">BUFFER: 4112 commits</span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-[#fbbf24] border border-[#333333] px-1 py-0.2">CLI: COMING SOON</span>
          <span className="text-[#4ade80]">DIAGNOSTICS: GROUNDED</span>
          <span>UTF-8</span>
          <span className="hidden sm:inline">PRESS [?] FOR COMMANDS</span>
        </div>
      </footer>

      {/* 4. Inspection / Evidence Modal */}
      <EvidenceModal
        isOpen={modalState.isOpen}
        onClose={() => setModalState((prev) => ({ ...prev, isOpen: false }))}
        selectedItem={modalState.isOpen ? modalState : null}
        profile={profile}
      />
    </main>
  );
}
