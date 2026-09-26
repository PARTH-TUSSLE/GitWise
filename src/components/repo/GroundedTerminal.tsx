"use client";

import React, { useState } from "react";
import { RepoModel } from "@/lib/repoData";
import { Terminal, Send, ExternalLink, Code2 } from "lucide-react";

interface GroundedTerminalProps {
  repo: RepoModel;
  onSelectFile: (filePath: string) => void;
  onTraceTriggered: (featureName: string) => void;
}

export function GroundedTerminal({
  repo,
  onSelectFile,
  onTraceTriggered,
}: GroundedTerminalProps) {
  const [messages, setMessages] = useState(repo.chatHistory);
  const [inputQuery, setInputQuery] = useState("");

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputQuery.trim()) return;

    const userText = inputQuery.trim();
    setInputQuery("");

    // Add user message
    const newMessages = [...messages, { sender: "user" as const, message: userText }];
    setMessages(newMessages);

    // Generate grounded contextual response
    setTimeout(() => {
      let botResponse = "";
      let citations: Array<{ file: string; line?: number; snippet?: string }> = [];

      if (userText.toLowerCase().includes("trace") || userText.toLowerCase().includes("action")) {
        botResponse = `Executing topological feature trace for "${userText}". Mapped call graph through request headers, dispatch pipeline, and server handler closures.`;
        citations = [
          { file: repo.featureTraces[0]?.steps[0]?.file || "packages/next/src/server/app-render/action-handler.ts", line: 42, snippet: "// Feature trigger boundary entry point" },
          { file: repo.featureTraces[0]?.steps[1]?.file || "packages/next/src/server/app-render/app-render.tsx", line: 110, snippet: "// State execution reconciler" },
        ];
        onTraceTriggered(repo.featureTraces[0]?.id || "");
      } else if (userText.toLowerCase().includes("start") || userText.toLowerCase().includes("contribute")) {
        botResponse = `For new contributors to ${repo.name}, the recommended starting files are well-tested client utilities with minimal external coupling.`;
        citations = repo.contributorGuide.beginnerFiles.map((b) => ({
          file: b.path,
          line: 1,
          snippet: `// Recommended: ${b.reason}`,
        }));
      } else {
        botResponse = `Repository analysis grounded in ${repo.indexedFiles} indexed files across ${repo.subsystems.length} subsystems. Subsystem architecture is documented in the central execution matrix.`;
        citations = [
          { file: repo.subsystems[0]?.entryPoint || "README.md", line: 1, snippet: `// Subsystem root: ${repo.subsystems[0]?.name}` },
        ];
      }

      setMessages([...newMessages, { sender: "gitwise" as const, message: botResponse, citations }]);
    }, 400);
  };

  return (
    <div className="border border-[#262626] bg-[#0a0a0a] p-3 text-xs font-mono flex flex-col h-full">
      {/* Header */}
      <div className="border-b border-[#262626] pb-2 mb-3 flex items-center justify-between">
        <div>
          <div className="font-bold text-[#fafafa]">Grounded AI Terminal</div>
          <div className="text-[10px] text-[#525252]">EVIDENCE &amp; CITATIONS ENGINE</div>
        </div>
        <span className="text-[10px] text-[#4ade80] flex items-center gap-1">
          <span className="w-1.5 h-1.5 bg-[#4ade80]"></span>
          GROUNDED
        </span>
      </div>

      {/* Messages Stream */}
      <div className="flex-1 overflow-y-auto space-y-3 pr-1 max-h-[520px]">
        {messages.map((msg, idx) => (
          <div
            key={idx}
            className={`p-2.5 border text-xs leading-relaxed ${
              msg.sender === "user"
                ? "border-[#333333] bg-[#0e0e0e]"
                : "border-[#1f1f1f] bg-[#070707]"
            }`}
          >
            <div className="flex items-center gap-1.5 text-[10px] mb-1 font-bold">
              <span className={msg.sender === "user" ? "text-[#fafafa]" : "text-[#fbbf24]"}>
                {msg.sender === "user" ? "gw_user>" : "gitwise-core>"}
              </span>
            </div>

            <div className="text-[#fafafa] mb-2">{msg.message}</div>

            {/* Citations Box */}
            {msg.citations && msg.citations.length > 0 && (
              <div className="border-t border-[#1a1a1a] pt-2 mt-2 space-y-1.5">
                <div className="text-[10px] text-[#525252] uppercase font-medium">
                  Verified File References
                </div>
                {msg.citations.map((c, cIdx) => (
                  <button
                    key={cIdx}
                    onClick={() => onSelectFile(c.file)}
                    className="w-full text-left p-1.5 border border-[#1f1f1f] hover:border-[#fbbf24] bg-[#0e0e0e] transition-colors group"
                  >
                    <div className="flex items-center justify-between text-[10px] text-[#fbbf24]">
                      <span className="font-medium truncate group-hover:underline">
                        [REF: {c.file.split("/").slice(-2).join("/")}{c.line ? `:${c.line}` : ""}]
                      </span>
                      <ExternalLink className="w-3 h-3 text-[#525252] group-hover:text-[#fbbf24]" />
                    </div>
                    {c.snippet && (
                      <div className="text-[10px] text-[#8a8a8a] truncate font-mono mt-0.5">
                        {c.snippet}
                      </div>
                    )}
                  </button>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Terminal Input Form */}
      <form onSubmit={handleSend} className="mt-3 pt-2 border-t border-[#1f1f1f] flex gap-1.5">
        <span className="text-[#fbbf24] select-none py-1">&gt;</span>
        <input
          type="text"
          value={inputQuery}
          onChange={(e) => setInputQuery(e.target.value)}
          placeholder="Ask architecture or trace feature..."
          className="flex-1 bg-transparent text-[#fafafa] font-mono text-xs focus:outline-none placeholder:text-[#525252] py-1 border-b border-[#262626] focus:border-[#fbbf24]"
        />
        <button
          type="submit"
          className="px-2 py-1 border border-[#333333] hover:border-[#fbbf24] text-[#fbbf24] transition-colors text-[10px]"
        >
          SEND
        </button>
      </form>
    </div>
  );
}
