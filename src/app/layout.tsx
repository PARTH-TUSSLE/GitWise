import type { Metadata } from "next";
import "./globals.css";

import { ThemeProvider } from "@/context/ThemeContext";

export const metadata: Metadata = {
  title: "GitWise | Compiler-Grounded Intelligence for the Open Source Web",
  description: "Deterministic codebase comprehension, AST feature tracing, blast radius analysis, and semantic pull request reviews.",
};

const directionContractHtml = `<!--
THESIS: Pure developer minimalism inspired by modern terminal emulators and compiler diagnostics—high-contrast, razor-sharp typography, zero fluff, refusing rounded bento cards and vanity metrics.
OWN-WORLD: Palette #050505 ground, #fafafa phosphor white, #f43f5e rose accent, #262626 border rules. Zero-radius box-drawing borders, monospace tabular grid, CLI prompt syntax.
STORY: Contributor inputs any GitHub username or repository, inspecting verifiable open-source engineering footprint across PRs, reviews, issues, and code deltas.
FIRST VIEWPORT: Multi-column terminal workbench with top command input gitwise explore, left navigation rail, interactive AST dissection chamber, live PR diff stream.
FORM: Monochrome Terminal & Compiler Sheet.
FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance.
-->`;

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" data-theme="rose" suppressHydrationWarning>
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: `try{const s=localStorage.getItem('gitwise-theme');if(s){document.documentElement.setAttribute('data-theme',s);}else{document.documentElement.setAttribute('data-theme','rose');}}catch(e){}`,
          }}
        />
      </head>
      <body className="min-h-screen bg-[#050505] text-[#fafafa] antialiased">
        <div
          dangerouslySetInnerHTML={{ __html: directionContractHtml }}
          style={{ display: "contents" }}
        />
        <ThemeProvider>
          {children}
        </ThemeProvider>
      </body>
    </html>
  );
}
