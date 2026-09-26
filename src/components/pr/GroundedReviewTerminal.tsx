"use client";

import React, { useState } from "react";
import { PullRequestModel, ReviewFinding } from "@/lib/prData";
import { Terminal, ShieldCheck, Copy, Check, ExternalLink, Send, AlertTriangle, CheckCircle2 } from "lucide-react";

interface GroundedReviewTerminalProps {
  pr: PullRequestModel;
  onInspectFile: (filePath: string) => void;
}

export function GroundedReviewTerminal({
  pr,
  onInspectFile,
}: GroundedReviewTerminalProps) {
  const [copied, setCopied] = useState(false);
  const [query, setQuery] = useState("");
  const [messages, setMessages] = useState<Array<{ sender: "user" | "reviewer"; text: string; citation?: string }>>([
    {
      sender: "reviewer",
      text: `PR #${pr.number} semantic review completed. Verified ${pr.reviewFindings.length} repository rules across ${pr.stats.filesChanged} files.`,
      citation: pr.files[0]?.path,
    },
  ]);

  const copyCommand = (cmd: string) => {
    navigator.clipboard.writeText(cmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    const userText = query.trim();
    setQuery("");

    const newMsgs = [...messages, { sender: "user" as const, text: userText }];
    setMessages(newMsgs);

    setTimeout(() => {
      let reply = "";
      let cit: string | undefined = undefined;

      if (userText.toLowerCase().includes("break") || userText.toLowerCase().includes("api")) {
        reply = `Public API contract is ${pr.architecturalShift.publicApiContract.toUpperCase()}. ${pr.architecturalShift.publicApiExplanation}`;
        cit = pr.files[0]?.path;
      } else if (userText.toLowerCase().includes("test") || userText.toLowerCase().includes("run")) {
        reply = `Run '${pr.testVerification.command}' to verify all touched paths. Expected result: ${pr.testVerification.expectedResult}.`;
        cit = pr.files.find((f) => f.path.includes("test"))?.path || pr.files[0]?.path;
      } else {
        reply = `Reviewed against ${pr.repo} conventions. State mutation risk assessed as ${pr.architecturalShift.stateMutationRisk.toUpperCase()}.`;
        cit = pr.files[0]?.path;
      }

      setMessages([...newMsgs, { sender: "reviewer" as const, text: reply, citation: cit }]);
    }, 300);
  };

  return (
    <div className="flex flex-col gap-3 font-mono text-xs">
      {/* 1. Grounded AI Code Review Findings */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2.5">
        <div className="border-b border-[#262626] pb-2 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <ShieldCheck className="w-3.5 h-3.5 text-[#fbbf24]" />
            <span className="font-bold text-[#fafafa] uppercase text-[11px]">
              GROUNDED AI CODE REVIEW
            </span>
          </div>
          <span className="text-[10px] text-[#4ade80]">
            {pr.reviewFindings.length} CHECKS PASSED
          </span>
        </div>

        <div className="flex flex-col gap-2">
          {pr.reviewFindings.map((finding) => (
            <div
              key={finding.ruleId}
              className="border border-[#1f1f1f] bg-[#070707] p-2.5 flex flex-col gap-1.5"
            >
              <div className="flex items-center justify-between">
                <span className="text-[10px] font-bold text-[#fbbf24]">
                  [{finding.ruleId}] {finding.category.toUpperCase()}
                </span>
                <button
                  onClick={() => onInspectFile(finding.file)}
                  className="text-[9px] text-[#8a8a8a] hover:text-[#fbbf24] flex items-center gap-1"
                >
                  <span>{finding.file.split("/").pop()}</span>
                  <ExternalLink className="w-2.5 h-2.5" />
                </button>
              </div>

              <p className="text-[10px] text-[#fafafa] leading-snug">
                {finding.message}
              </p>

              {finding.suggestion && (
                <div className="border-t border-[#1a1a1a] pt-1 text-[9px] text-[#fbbf24]">
                  Suggestion: {finding.suggestion}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* 2. Test Verification & Coverage Delta */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2">
        <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between">
          <span className="font-bold text-[#fafafa] uppercase text-[11px]">
            TEST VERIFICATION
          </span>
          <span className="text-[10px] text-[#4ade80]">
            {pr.testVerification.coverageDelta}
          </span>
        </div>

        <div className="border border-[#1f1f1f] bg-[#070707] p-2 flex flex-col gap-1.5">
          <div className="text-[10px] text-[#8a8a8a] truncate">
            Target: {pr.testVerification.targetSuite}
          </div>

          <div className="flex items-center justify-between bg-[#040404] border border-[#171717] px-2 py-1 text-[10px] text-[#fafafa]">
            <code className="truncate pr-2">{pr.testVerification.command}</code>
            <button
              onClick={() => copyCommand(pr.testVerification.command)}
              className="text-[#8a8a8a] hover:text-[#fbbf24] transition-colors shrink-0"
              title="Copy command"
            >
              {copied ? (
                <Check className="w-3 h-3 text-[#4ade80]" />
              ) : (
                <Copy className="w-3 h-3" />
              )}
            </button>
          </div>

          <div className="text-[9px] text-[#525252]">
            Expected: {pr.testVerification.expectedResult}
          </div>
        </div>
      </div>

      {/* 3. Interactive Review Query Bar */}
      <div className="border border-[#262626] bg-[#0c0c0c] p-3 flex flex-col gap-2">
        <div className="flex items-center justify-between text-[10px] text-[#8a8a8a] border-b border-[#1f1f1f] pb-1">
          <div className="flex items-center gap-1.5 text-[#4ade80]">
            <span className="w-1.5 h-1.5 bg-[#4ade80] inline-block animate-pulse"></span>
            <span>SEMANTIC PR COPILOT</span>
          </div>
          <span className="text-[#525252]">TTY: /dev/pts/4</span>
        </div>

        <div className="flex flex-col gap-2 max-h-[140px] overflow-y-auto pr-1">
          {messages.map((m, idx) => (
            <div key={idx} className="flex flex-col gap-0.5 text-[10px]">
              <span className={m.sender === "user" ? "text-[#fbbf24] font-bold" : "text-[#8a8a8a] font-bold"}>
                {m.sender === "user" ? "gw_reviewer>" : "gitwise-audit>"}
              </span>
              <p className="text-[#fafafa] leading-tight">{m.text}</p>
              {m.citation && (
                <button
                  onClick={() => onInspectFile(m.citation!)}
                  className="text-[9px] text-[#fbbf24] hover:underline self-start"
                >
                  [REF: {m.citation.split("/").pop()}]
                </button>
              )}
            </div>
          ))}
        </div>

        <form onSubmit={handleSend} className="flex items-center gap-1.5 border-t border-[#1f1f1f] pt-1.5">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Ask question about PR changes..."
            className="flex-1 bg-transparent text-[#fafafa] font-mono text-[10px] focus:outline-none placeholder:text-[#404040]"
          />
          <button
            type="submit"
            className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#fbbf24] text-[9px]"
          >
            SEND
          </button>
        </form>
      </div>
    </div>
  );
}
