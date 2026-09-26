"use client";

import React, { useState } from "react";
import { RepoModel, RepoTreeItem, SubsystemNode } from "@/lib/repoData";
import { Folder, FolderOpen, FileCode, CheckCircle, ChevronRight, ChevronDown } from "lucide-react";

interface RepoTreeRailProps {
  repo: RepoModel;
  selectedSubsystem: SubsystemNode | null;
  onSelectSubsystem: (subsystem: SubsystemNode) => void;
  onSelectFile: (filePath: string) => void;
}

export function RepoTreeRail({
  repo,
  selectedSubsystem,
  onSelectSubsystem,
  onSelectFile,
}: RepoTreeRailProps) {
  const [openDirs, setOpenDirs] = useState<Record<string, boolean>>({
    packages: true,
    "packages/next": true,
    crates: true,
    src: true,
    cmd: true,
    pkg: true,
    server: true,
  });

  const toggleDir = (dirPath: string) => {
    setOpenDirs((prev) => ({ ...prev, [dirPath]: !prev[dirPath] }));
  };

  const renderTree = (items: RepoTreeItem[], depth = 0) => {
    return (
      <div className="flex flex-col gap-0.5">
        {items.map((item) => {
          if (item.type === "directory") {
            const isOpen = openDirs[item.path] ?? true;
            return (
              <div key={item.path} className="flex flex-col">
                <button
                  onClick={() => toggleDir(item.path)}
                  style={{ paddingLeft: `${depth * 12 + 4}px` }}
                  className="flex items-center gap-1.5 py-0.5 text-left text-[#8a8a8a] hover:text-[#fafafa] hover:bg-[#121212] transition-colors"
                >
                  {isOpen ? <ChevronDown className="w-3 h-3 text-[#525252]" /> : <ChevronRight className="w-3 h-3 text-[#525252]" />}
                  {isOpen ? <FolderOpen className="w-3.5 h-3.5 text-[#fbbf24]" /> : <Folder className="w-3.5 h-3.5 text-[#fbbf24]" />}
                  <span className="font-medium text-[#fafafa]">{item.name}</span>
                </button>
                {isOpen && item.children && renderTree(item.children, depth + 1)}
              </div>
            );
          }

          return (
            <button
              key={item.path}
              onClick={() => onSelectFile(item.path)}
              style={{ paddingLeft: `${depth * 12 + 16}px` }}
              className="flex items-center justify-between py-0.5 text-left text-[#8a8a8a] hover:text-[#fbbf24] hover:bg-[#141414] transition-colors group"
            >
              <div className="flex items-center gap-1.5 truncate">
                <FileCode className="w-3 h-3 text-[#525252] group-hover:text-[#fbbf24]" />
                <span className="truncate">{item.name}</span>
              </div>
              {item.size && (
                <span className="text-[10px] text-[#525252] mr-2">{item.size}</span>
              )}
            </button>
          );
        })}
      </div>
    );
  };

  return (
    <div className="flex flex-col gap-3 font-mono text-xs">
      {/* Box 1: Repository Tree Header & Navigation */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3">
        <div className="border-b border-[#262626] pb-2 mb-2 flex items-center justify-between text-[11px]">
          <span className="font-bold text-[#fafafa]">Repository Tree &amp; Modules</span>
          <span className="text-[#525252]">{repo.branch}</span>
        </div>

        {/* Tree Container */}
        <div className="max-h-[300px] overflow-y-auto pr-1">
          {renderTree(repo.fileTree)}
        </div>
      </div>

      {/* Box 2: Architectural Subsystems Quick List */}
      <div className="border border-[#262626] bg-[#0a0a0a] p-3">
        <div className="border-b border-[#262626] pb-2 mb-2 flex items-center justify-between text-[11px]">
          <span className="font-bold text-[#fafafa]">Subsystems ({repo.subsystems.length})</span>
          <span className="text-[10px] text-[#4ade80]">INDEXED</span>
        </div>

        <div className="flex flex-col gap-1">
          {repo.subsystems.map((sub) => {
            const isSelected = selectedSubsystem?.id === sub.id;
            return (
              <button
                key={sub.id}
                onClick={() => onSelectSubsystem(sub)}
                className={`p-1.5 text-left border transition-colors ${
                  isSelected
                    ? "border-[#fbbf24] bg-[#1a1708] text-[#fbbf24]"
                    : "border-[#1f1f1f] hover:border-[#404040] bg-[#0c0c0c] text-[#8a8a8a] hover:text-[#fafafa]"
                }`}
              >
                <div className="flex items-center justify-between text-[11px] font-medium">
                  <span className="truncate">{sub.name}</span>
                  <span className="text-[10px] text-[#525252]">{sub.fileCount} files</span>
                </div>
                <div className="text-[10px] text-[#525252] truncate mt-0.5">
                  {sub.entryPoint}
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* Box 3: Selected Module Telemetry */}
      {selectedSubsystem && (
        <div className="border border-[#262626] bg-[#0a0a0a] p-3 text-[11px]">
          <div className="text-[10px] text-[#525252] uppercase mb-1">Module Telemetry</div>
          <div className="font-bold text-[#fbbf24] mb-1">{selectedSubsystem.name}</div>
          <div className="text-[#8a8a8a] text-[10px] leading-relaxed mb-2">
            {selectedSubsystem.description}
          </div>
          <div className="border-t border-[#1f1f1f] pt-1.5 flex justify-between text-[10px] text-[#525252]">
            <span>Lang: <strong className="text-[#fafafa]">{selectedSubsystem.language}</strong></span>
            <span>Beginner: <strong className={selectedSubsystem.beginnerFriendly ? "text-[#4ade80]" : "text-[#8a8a8a]"}>{selectedSubsystem.beginnerFriendly ? "Yes" : "Advanced"}</strong></span>
          </div>
        </div>
      )}
    </div>
  );
}
