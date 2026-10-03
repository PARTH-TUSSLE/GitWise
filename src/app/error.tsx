"use client";

import React, { useEffect } from "react";
import { AlertTriangle, RotateCcw, Home } from "lucide-react";

interface GlobalErrorProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function GlobalError({ error, reset }: GlobalErrorProps) {
  useEffect(() => {
    console.error("[GitWise Global Error Boundary]:", error);
  }, [error]);

  return (
    <div className="min-h-screen bg-[#050505] text-[#fafafa] font-mono flex items-center justify-center p-4">
      <div className="max-w-md w-full border border-[#262626] bg-[#0d0d0d] p-6 shadow-2xl flex flex-col gap-4">
        <div className="flex items-center gap-3 text-[#ef4444] border-b border-[#262626] pb-3">
          <AlertTriangle className="w-5 h-5 flex-shrink-0 animate-pulse" />
          <h1 className="text-sm font-bold uppercase tracking-wider">
            SYSTEM EXCEPTION INTERCEPTED
          </h1>
        </div>

        <p className="text-xs text-[#a3a3a3] leading-relaxed">
          An unhandled error occurred while rendering the current interface view.
          The backend connection may have been interrupted or the requested state was corrupted.
        </p>

        {error.message && (
          <div className="bg-[#171717] border border-[#262626] p-3 text-[11px] text-[#ef4444] font-mono break-words">
            {error.message}
            {error.digest && (
              <div className="text-[10px] text-[#737373] mt-1">
                Digest: {error.digest}
              </div>
            )}
          </div>
        )}

        <div className="flex items-center gap-2 pt-2">
          <button
            onClick={() => reset()}
            className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 bg-[#fbbf24] text-[#050505] text-xs font-bold hover:bg-[#f59e0b] transition-colors"
          >
            <RotateCcw className="w-3.5 h-3.5" />
            RETRY VIEW
          </button>
          <button
            onClick={() => (window.location.href = "/")}
            className="flex items-center justify-center gap-1.5 px-3 py-2 border border-[#333333] text-[#fafafa] text-xs font-medium hover:bg-[#1a1a1a] transition-colors"
          >
            <Home className="w-3.5 h-3.5" />
            RETURN HOME
          </button>
        </div>
      </div>
    </div>
  );
}
