"use client";

import React, { useState } from "react";
import { MentorMessage } from "@/lib/mentorData";
import { Send, Copy, Check, ExternalLink, Sparkles, Terminal } from "lucide-react";

interface MentorTerminalStreamProps {
  messages: MentorMessage[];
  onSendMessage: (text: string) => void;
  onInspectFile: (filePath: string) => void;
  isThinking?: boolean;
}

export function MentorTerminalStream({
  messages,
  onSendMessage,
  onInspectFile,
  isThinking = false,
}: MentorTerminalStreamProps) {
  const [inputText, setInputText] = useState("");
  const [copiedCmd, setCopiedCmd] = useState<string | null>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (inputText.trim()) {
      onSendMessage(inputText.trim());
      setInputText("");
    }
  };

  const copyCommand = (cmd: string) => {
    navigator.clipboard.writeText(cmd);
    setCopiedCmd(cmd);
    setTimeout(() => setCopiedCmd(null), 1500);
  };

  const quickPrompts = [
    "Where should I start?",
    "Match issues to my skills",
    "Explain Server Actions architecture",
    "How do I run tests locally?",
  ];

  return (
    <div className="flex flex-col h-full border border-[#262626] bg-[#0a0a0a] font-mono">
      {/* 1. Header Banner */}
      <div className="border-b border-[#262626] p-3 bg-[#0c0c0c] flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Sparkles className="w-3.5 h-3.5 text-[#fbbf24]" />
          <span className="font-bold text-[#fafafa] uppercase text-[11px] tracking-wider">
            SOCRATIC MENTORSHIP STREAM
          </span>
        </div>
        <div className="flex items-center gap-2 text-[10px]">
          <span className="w-1.5 h-1.5 bg-[#4ade80] inline-block animate-pulse"></span>
          <span className="text-[#4ade80]">MENTOR ONLINE</span>
          <span className="text-[#525252]">|</span>
          <span className="text-[#8a8a8a]">GROUNDED ARCHITECTURE REASONING</span>
        </div>
      </div>

      {/* 2. Message History Stream */}
      <div className="flex-1 p-4 overflow-y-auto flex flex-col gap-4 max-h-[640px]">
        {messages.map((msg) => (
          <div
            key={msg.id}
            className={`flex flex-col gap-1.5 ${
              msg.sender === "user"
                ? "items-end"
                : "items-start border-l-2 border-[#fbbf24] pl-3 py-1"
            }`}
          >
            {/* Sender and Timestamp Label */}
            <div className="flex items-center gap-2 text-[10px] text-[#8a8a8a]">
              <span className={`font-bold ${msg.sender === "user" ? "text-[#fbbf24]" : "text-[#fafafa]"}`}>
                {msg.sender === "user" ? "alexR_dev (contributor)" : "gitwise mentor (socratic guide)"}
              </span>
              <span>{msg.timestamp}</span>
            </div>

            {/* Content Body - Crafted for Maximum Readability */}
            <div
              className={`max-w-[95%] text-xs leading-relaxed whitespace-pre-line ${
                msg.sender === "user"
                  ? "bg-[#141208] border border-[#333333] text-[#fafafa] p-2.5"
                  : "text-[#ededed]"
              }`}
            >
              {msg.content}
            </div>

            {/* Actionable Terminal Command Blocks */}
            {msg.actionCommands && msg.actionCommands.length > 0 && (
              <div className="w-full flex flex-col gap-1 mt-1">
                <span className="text-[10px] text-[#8a8a8a] uppercase font-bold">
                  VERIFICATION COMMANDS:
                </span>
                {msg.actionCommands.map((cmd, idx) => (
                  <div
                    key={idx}
                    className="flex items-center justify-between bg-[#040404] border border-[#1f1f1f] px-2.5 py-1.5 text-[11px] text-[#fafafa]"
                  >
                    <code className="text-[#fbbf24]">{cmd}</code>
                    <button
                      onClick={() => copyCommand(cmd)}
                      className="text-[#8a8a8a] hover:text-[#fbbf24] transition-colors p-1"
                      title="Copy command"
                    >
                      {copiedCmd === cmd ? (
                        <Check className="w-3 h-3 text-[#4ade80]" />
                      ) : (
                        <Copy className="w-3 h-3" />
                      )}
                    </button>
                  </div>
                ))}
              </div>
            )}

            {/* Clickable Citations */}
            {msg.citations && msg.citations.length > 0 && (
              <div className="flex flex-wrap items-center gap-1.5 mt-1">
                <span className="text-[9px] text-[#525252] uppercase select-none">
                  VERIFIED CITATIONS:
                </span>
                {msg.citations.map((cit, idx) => (
                  <button
                    key={idx}
                    onClick={() => onInspectFile(cit.file)}
                    className="inline-flex items-center gap-1 px-1.5 py-0.5 border border-[#262626] bg-[#0d0d0d] hover:border-[#fbbf24] text-[10px] text-[#fbbf24] transition-colors"
                  >
                    <span>[REF: {cit.file.split("/").pop()}{cit.line ? `:${cit.line}` : ""}]</span>
                    <ExternalLink className="w-2.5 h-2.5" />
                  </button>
                ))}
              </div>
            )}
          </div>
        ))}

        {isThinking && (
          <div className="flex items-center gap-2 text-[11px] text-[#8a8a8a] border-l-2 border-[#fbbf24] pl-3 py-2">
            <span className="w-2 h-2 bg-[#fbbf24] animate-ping inline-block"></span>
            <span>Synthesizing architectural reasoning from indexed AST...</span>
          </div>
        )}
      </div>

      {/* 3. Quick Inquiry Chips & Input Form */}
      <div className="border-t border-[#262626] p-3 bg-[#080808] flex flex-col gap-2">
        {/* Quick prompt chips */}
        <div className="flex flex-wrap items-center gap-1">
          <span className="text-[9px] text-[#525252] select-none">Ask:</span>
          {quickPrompts.map((prompt, idx) => (
            <button
              key={idx}
              onClick={() => onSendMessage(prompt)}
              className="px-2 py-0.5 border border-[#1f1f1f] bg-[#0c0c0c] hover:border-[#fbbf24] hover:text-[#fbbf24] text-[#8a8a8a] text-[10px] transition-colors"
            >
              {prompt}
            </button>
          ))}
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="flex items-center gap-2">
          <input
            type="text"
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            placeholder="Ask a question about the codebase, architecture, or where to start..."
            className="flex-1 bg-transparent text-[#fafafa] font-mono text-xs focus:outline-none placeholder:text-[#525252] border border-[#262626] focus:border-[#fbbf24] px-3 py-2"
          />
          <button
            type="submit"
            className="px-3 py-2 bg-[#141208] border border-[#fbbf24] text-[#fbbf24] hover:bg-[#fbbf24] hover:text-[#050505] font-mono text-xs font-bold transition-colors flex items-center gap-1.5"
          >
            <span>SEND</span>
            <Send className="w-3 h-3" />
          </button>
        </form>
      </div>
    </div>
  );
}
