"use client";

import React, { useState } from "react";
import { IssueModel, GroundedTestSuite } from "@/lib/issueData";
import { Terminal, Copy, Check, ExternalLink, Send, ShieldCheck, AlertCircle } from "lucide-react";

interface GroundedVerificationTerminalProps {
  issue: IssueModel;
  onSelectFile: (filePath: string) => void;
}

export function GroundedVerificationTerminal({
  issue,
  onSelectFile,
}: GroundedVerificationTerminalProps) {
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);
  const [checklist, setChecklist] = useState<Record<number, boolean>>({});
  const [query, setQuery] = useState("");
  const [messages, setMessages] = useState<Array<{ sender: "user" | "core"; text: string; citation?: string }>>([
    {
      sender: "core",
      text: `Issue #${issue.number} ingested. Identified ${issue.blastRadius.fileCount} affected files across ${issue.blastRadius.subsystemsAffected.join(", ")}. Test specs loaded.`,
      citation: issue.affectedFiles[0]?.path,
    },
  ]);

  const copyToClipboard = (text: string, index: number) => {
    navigator.clipboard.writeText(text);
    setCopiedIndex(index);
    setTimeout(() => setCopiedIndex(null), 1500);
  };

  const toggleCheck = (idx: number) => {
    setChecklist((prev) => ({ ...prev, [idx]: !prev[idx] }));
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

      if (userText.toLowerCase().includes("test") || userText.toLowerCase().includes("run")) {
        reply = `Verification runner ready. Execute '${issue.testStrategy.suites[0]?.command}' to validate patch boundaries.`;
        cit = issue.testStrategy.suites[0]?.testFile;
      } else if (userText.toLowerCase().includes("risk") || userText.toLowerCase().includes("blast")) {
        reply = `Blast radius score: ${issue.blastRadius.score.toFixed(1)} / 5.0. Primary risk: ${issue.blastRadius.riskAssessment}`;
        cit = issue.affectedFiles[0]?.path;
      } else {
        reply = `Analysis grounded in #${issue.number} triage context. Review Step 1 diff before initiating pull request.`;
        cit = issue.stages.stage03_blueprint.steps[0]?.file;
      }

      setMessages([...newMsgs, { sender: "core" as const, text: reply, citation: cit }]);
    }, 300);
  };

  return (
    <div className="flex flex-col gap-3 font-mono text-xs">
      {/* 1. Grounded Test Strategy */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-2.5">
        <div className="border-b border-[#262626] pb-2 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <ShieldCheck className="w-3.5 h-3.5 text-[#fbbf24]" />
            <span className="font-bold text-[#fafafa] uppercase text-[11px]">
              GROUNDED TEST STRATEGY
            </span>
          </div>
          <span className="text-[10px] text-[#4ade80]">
            {issue.testStrategy.suites.length} SUITES VERIFIED
          </span>
        </div>

        <div className="flex flex-col gap-2">
          {issue.testStrategy.suites.map((suite, idx) => (
            <div key={idx} className="border border-[#1f1f1f] bg-[#070707] p-2 flex flex-col gap-1.5">
              <div className="flex items-center justify-between">
                <span className="text-[10px] text-[#fbbf24] font-bold truncate">
                  [{suite.type.toUpperCase()}] {suite.name}
                </span>
                <button
                  onClick={() => onSelectFile(suite.testFile)}
                  className="text-[9px] text-[#8a8a8a] hover:text-[#fafafa] flex items-center gap-1"
                >
                  <span>{suite.testFile.split("/").pop()}</span>
                  <ExternalLink className="w-2.5 h-2.5" />
                </button>
              </div>

              {/* CLI Command with Copy Button */}
              <div className="flex items-center justify-between bg-[#040404] border border-[#171717] px-2 py-1 text-[10px] text-[#fafafa]">
                <code className="truncate pr-2">{suite.command}</code>
                <button
                  onClick={() => copyToClipboard(suite.command, idx)}
                  className="text-[#8a8a8a] hover:text-[#fbbf24] transition-colors shrink-0"
                  title="Copy command"
                >
                  {copiedIndex === idx ? (
                    <Check className="w-3 h-3 text-[#4ade80]" />
                  ) : (
                    <Copy className="w-3 h-3" />
                  )}
                </button>
              </div>

              <div className="text-[9px] text-[#525252] truncate">
                Expected: {suite.expectedOutput}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* 2. Edge Cases Matrix */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-1.5">
        <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between">
          <span className="font-bold text-[#fafafa] uppercase text-[11px]">
            EDGE CASES MATRIX
          </span>
          <span className="text-[10px] text-[#f87171]">REGRESSION AUDIT</span>
        </div>
        <ul className="flex flex-col gap-1 text-[10px] text-[#8a8a8a] mt-1">
          {issue.testStrategy.edgeCases.map((ec, idx) => (
            <li key={idx} className="flex items-start gap-1.5 leading-snug">
              <span className="text-[#f87171] select-none">&bull;</span>
              <span>{ec}</span>
            </li>
          ))}
        </ul>
      </div>

      {/* 3. Contributor Verification Checklist */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3 flex flex-col gap-1.5">
        <div className="border-b border-[#262626] pb-1.5 flex items-center justify-between">
          <span className="font-bold text-[#fafafa] uppercase text-[11px]">
            CONTRIBUTOR VERIFICATION CHECKLIST
          </span>
        </div>
        <div className="flex flex-col gap-1 mt-1">
          {issue.testStrategy.verificationChecklist.map((item, idx) => (
            <button
              key={idx}
              onClick={() => toggleCheck(idx)}
              className="flex items-start gap-2 text-left p-1 hover:bg-[#121212] transition-colors"
            >
              <div
                className={`w-3.5 h-3.5 border mt-0.5 shrink-0 flex items-center justify-center ${
                  checklist[idx] ? "border-[#4ade80] bg-[#4ade80]/20 text-[#4ade80]" : "border-[#333333]"
                }`}
              >
                {checklist[idx] && <Check className="w-2.5 h-2.5" />}
              </div>
              <span
                className={`text-[10px] leading-snug ${
                  checklist[idx] ? "text-[#525252] line-through" : "text-[#fafafa]"
                }`}
              >
                {item}
              </span>
            </button>
          ))}
        </div>
      </div>

      {/* 4. Terminal Chat / Query Bar */}
      <div className="border border-[#262626] bg-[#0c0c0c] p-3 flex flex-col gap-2">
        <div className="flex items-center justify-between text-[10px] text-[#8a8a8a] border-b border-[#1f1f1f] pb-1">
          <div className="flex items-center gap-1.5 text-[#4ade80]">
            <span className="w-1.5 h-1.5 bg-[#4ade80] inline-block animate-pulse"></span>
            <span>GROUNDED AI VERIFIER</span>
          </div>
          <span className="text-[#525252]">TTY: /dev/pts/3</span>
        </div>

        <div className="flex flex-col gap-2 max-h-[140px] overflow-y-auto pr-1">
          {messages.map((m, idx) => (
            <div key={idx} className="flex flex-col gap-0.5 text-[10px]">
              <span className={m.sender === "user" ? "text-[#fbbf24] font-bold" : "text-[#8a8a8a] font-bold"}>
                {m.sender === "user" ? "gw_user>" : "gitwise-verifier>"}
              </span>
              <p className="text-[#fafafa] leading-tight">{m.text}</p>
              {m.citation && (
                <button
                  onClick={() => onSelectFile(m.citation!)}
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
            placeholder="Ask question about issue..."
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
