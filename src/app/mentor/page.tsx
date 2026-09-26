"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { MENTOR_DATA, RepoMentorModel, MentorMessage, MentorshipTrack } from "@/lib/mentorData";
import { MentorCommandBar } from "@/components/mentor/MentorCommandBar";
import { MentorshipTracksRail } from "@/components/mentor/MentorshipTracksRail";
import { MentorTerminalStream } from "@/components/mentor/MentorTerminalStream";
import { BlueprintAndSkillCard } from "@/components/mentor/BlueprintAndSkillCard";
import { CodeDrawerModal } from "@/components/repo/CodeDrawerModal";

export default function MentorPage() {
  const router = useRouter();
  const [currentRepo, setCurrentRepo] = useState<string>("vercel/next.js");
  const [messages, setMessages] = useState<MentorMessage[]>([]);
  const [isThinking, setIsThinking] = useState(false);
  const [activeModalFile, setActiveModalFile] = useState<string | null>(null);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const model: RepoMentorModel =
    MENTOR_DATA[currentRepo] ||
    MENTOR_DATA["vercel/next.js"];

  // Initialize messages when repo changes
  useEffect(() => {
    if (model) {
      setMessages(model.defaultMessages);
    }
  }, [currentRepo, model]);

  // Keyboard shortcut listener
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }

      if (e.key === "Escape") {
        setActiveModalFile(null);
      } else if (e.key === "p" || e.key === "P") {
        router.push("/pr");
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
    const matched = Object.keys(MENTOR_DATA).find(
      (k) => k.toLowerCase() === repo.toLowerCase()
    );
    if (matched) {
      setCurrentRepo(matched);
    } else {
      setCurrentRepo(repo);
    }
    setIsRefreshing(true);
    setTimeout(() => setIsRefreshing(false), 250);
  };

  const handleSendMessage = (text: string) => {
    const userMsg: MentorMessage = {
      id: `msg-${Date.now()}`,
      sender: "user",
      timestamp: new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
      content: text,
    };

    setMessages((prev) => [...prev, userMsg]);
    setIsThinking(true);

    setTimeout(() => {
      const lower = text.toLowerCase();
      let responseContent = "";
      let responseCitations: MentorMessage["citations"] = undefined;
      let responseCommands: string[] | undefined = undefined;

      if (lower.includes("start") || lower.includes("begin") || lower.includes("where")) {
        const entry = model.knowledgeBase["start"] || model.knowledgeBase["react-start"];
        responseContent = entry?.content || "For new contributors, begin by reading the contribution guidelines and executing local test suites.";
        responseCitations = entry?.citations;
        responseCommands = entry?.actionCommands;
      } else if (lower.includes("skill") || lower.includes("match") || lower.includes("issue")) {
        const entry = model.knowledgeBase["skills"];
        responseContent = entry?.content || `Matched ${model.matchedIssues.length} issues to your GitStat profile with high confidence.`;
        responseCitations = entry?.citations;
      } else if (lower.includes("arch") || lower.includes("intern") || lower.includes("work")) {
        const entry = model.knowledgeBase["architecture"];
        responseContent = entry?.content || `The architecture of ${model.name} is split across multiple core modules. Explore the topological call graph in the Repo Explorer surface.`;
        responseCitations = entry?.citations;
      } else if (lower.includes("test") || lower.includes("run") || lower.includes("debug")) {
        const entry = model.knowledgeBase["test"];
        responseContent = entry?.content || "Execute local unit test suites before pushing changes.";
        responseCommands = entry?.actionCommands;
      } else {
        responseContent = `I have indexed the architecture of **${model.name}** across ${model.primaryLanguage}. You can ask about:
- Where to start as a beginner.
- How to match open issues to your demonstrated GitStat skills.
- The internal execution flow of core features.
- Testing and verification workflows.`;
      }

      const mentorMsg: MentorMessage = {
        id: `msg-${Date.now() + 1}`,
        sender: "mentor",
        timestamp: new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
        content: responseContent,
        citations: responseCitations,
        actionCommands: responseCommands,
      };

      setMessages((prev) => [...prev, mentorMsg]);
      setIsThinking(false);
    }, 400);
  };

  const handleSelectTrack = (track: MentorshipTrack) => {
    handleSendMessage(track.queryPrompt);
  };

  return (
    <main className="min-h-screen bg-[#050505] text-[#fafafa] font-mono p-2 sm:p-4 lg:p-6 flex flex-col gap-3 selection:bg-[#fbbf24] selection:text-[#050505]">
      {/* 1. Global Mentor Command Bar */}
      <MentorCommandBar
        currentRepo={currentRepo}
        onSelectRepo={handleSelectRepo}
        onResetChat={() => setMessages(model.defaultMessages)}
        overallScore={model.contributorMatch.overallScore}
      />

      {/* 2. Main 3-Column Mentor Workbench Layout */}
      <div className={`grid grid-cols-1 lg:grid-cols-12 gap-3 flex-1 transition-opacity duration-200 ${isRefreshing ? "opacity-40" : "opacity-100"}`}>
        {/* Left Column: Mentorship Tracks & Milestones (cols 1-3) */}
        <aside className="lg:col-span-3 flex flex-col gap-3">
          <MentorshipTracksRail
            model={model}
            onSelectTrack={handleSelectTrack}
            onInspectFile={(filePath) => setActiveModalFile(filePath)}
          />
        </aside>

        {/* Center Column: Highly Readable Socratic Mentor Terminal Stream (cols 4-8) */}
        <section className="lg:col-span-6 flex flex-col gap-3">
          <MentorTerminalStream
            messages={messages}
            onSendMessage={handleSendMessage}
            onInspectFile={(filePath) => setActiveModalFile(filePath)}
            isThinking={isThinking}
          />
        </section>

        {/* Right Column: Repository Blueprint & Contributor Match Card (cols 9-12) */}
        <aside className="lg:col-span-3 flex flex-col gap-3">
          <BlueprintAndSkillCard
            model={model}
            onInspectFile={(filePath) => setActiveModalFile(filePath)}
          />
        </aside>
      </div>

      {/* 3. Terminal Status Footer Bar */}
      <footer className="border border-[#262626] bg-[#0a0a0a] px-3 py-1.5 text-[10px] text-[#525252] flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-3">
          <span className="text-[#fbbf24] font-bold">GITWISE MENTOR v3.0</span>
          <span>TARGET: {model.name}</span>
          <span className="text-[#4ade80]">CONTRIBUTOR: {model.contributorMatch.username}</span>
          <span className="hidden sm:inline">LANG: {model.primaryLanguage}</span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-[#fbbf24] border border-[#333333] px-1 py-0.2">CLI: COMING SOON</span>
          <span className="text-[#4ade80]">MODE: SOCRATIC GUIDANCE</span>
          <span>TTY: /dev/pts/5</span>
          <span className="hidden sm:inline">[P] PR &middot; [I] ISSUES &middot; [R] REPO &middot; [G] GITSTAT</span>
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
