"use client";

import React from "react";
import { X, ExternalLink, FileCode, CheckCircle2 } from "lucide-react";

interface CodeDrawerModalProps {
  isOpen: boolean;
  onClose: () => void;
  filePath: string;
  repoName: string;
}

export function CodeDrawerModal({
  isOpen,
  onClose,
  filePath,
  repoName,
}: CodeDrawerModalProps) {
  if (!isOpen || !filePath) return null;

  // Mock file contents based on file path
  const getFileContents = (path: string) => {
    if (path.includes("app-render")) {
      return `// packages/next/src/server/app-render/app-render.tsx
import type { IncomingMessage, ServerResponse } from 'http';
import { renderToReadableStream } from 'react-dom/server.edge';
import { FlightRenderPipeline } from './flight-render-pipeline';

export async function renderToHTMLOrFlight(
  req: IncomingMessage,
  res: ServerResponse,
  pagePath: string,
  query: Record<string, any>
) {
  // Initialize stream boundaries and Flight serialization context
  const flightPipeline = new FlightRenderPipeline({
    enableServerActions: true,
    turbopackHmr: process.env.NODE_ENV === 'development',
  });

  const stream = await flightPipeline.createStream(req);
  return stream;
}`;
    }

    if (path.includes("action-handler")) {
      return `// packages/next/src/server/app-render/action-handler.ts
import { decryptActionBoundArgs } from './action-encryption';
import { revalidatePath } from '../revalidation';

export async function handleServerActionRequest(req: Request) {
  const actionId = req.headers.get('next-action');
  if (!actionId) return null;

  const boundPayload = await req.json();
  const decryptedArgs = await decryptActionBoundArgs(actionId, boundPayload);

  // Invoke action handler and return mutated state Flight response
  const actionResult = await executeAction(actionId, decryptedArgs);
  return new Response(JSON.stringify(actionResult), {
    headers: { 'Content-Type': 'application/json' },
  });
}`;
    }

    return `// ${path}
// File indexed by GitWise Architecture Engine
// Ready for inspection and contributor editing

export default function moduleDefinition() {
  console.log("Subsystem module loaded: ${path}");
  return { status: "verified", active: true };
}`;
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-none p-4 font-mono text-xs">
      <div className="w-full max-w-2xl border border-[#fbbf24] bg-[#0c0c0c] shadow-2xl flex flex-col max-h-[85vh]">
        {/* Terminal Header */}
        <div className="flex items-center justify-between border-b border-[#262626] bg-[#141414] px-3 py-2 text-[#fafafa]">
          <div className="flex items-center gap-2 truncate">
            <span className="text-[#fbbf24] font-bold">gitwise&gt;</span>
            <span className="text-[#8a8a8a]">CAT:</span>
            <span className="text-[#fafafa] font-medium truncate">{filePath}</span>
          </div>
          <button
            onClick={onClose}
            className="text-[#8a8a8a] hover:text-[#fbbf24] transition-colors p-1"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Code Content */}
        <div className="p-4 flex-1 overflow-y-auto bg-[#070707]">
          <pre className="text-[#fafafa] text-xs leading-relaxed font-mono overflow-x-auto">
            <code>{getFileContents(filePath)}</code>
          </pre>
        </div>

        {/* Footer */}
        <div className="border-t border-[#262626] bg-[#121212] px-4 py-2 flex items-center justify-between text-[11px] text-[#525252]">
          <span>Repository: {repoName}</span>
          <div className="flex items-center gap-3">
            <a
              href={`https://github.com/${repoName}/blob/main/${filePath}`}
              target="_blank"
              rel="noreferrer"
              className="text-[#fbbf24] hover:underline flex items-center gap-1"
            >
              <span>Open on GitHub</span>
              <ExternalLink className="w-3 h-3" />
            </a>
            <button
              onClick={onClose}
              className="px-2 py-0.5 border border-[#333333] hover:border-[#fbbf24] text-[#8a8a8a] hover:text-[#fafafa] transition-colors"
            >
              ESC
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
