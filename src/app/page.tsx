"use client";

import React, { useState, useEffect, useRef } from "react";
import Link from "next/link";
import {
  Terminal,
  ArrowRight,
  ExternalLink,
  ShieldCheck,
  Sparkles,
  Layers,
  GitPullRequest,
  Flame,
  UserCheck,
  Code2,
  Play,
  Pause,
  RotateCcw,
  Cpu,
  GitBranch,
  SlidersHorizontal,
  Check,
  Maximize2,
  FileCode,
  Compass,
  AlertTriangle,
  FolderGit2,
  Activity,
  Network,
  Zap,
  CheckCircle2,
  Radio,
  Search,
  BookOpen
} from "lucide-react";
import { ThemeSelector } from "@/components/ThemeSelector";
import { CompilerBackground } from "@/components/CompilerBackground";
import { useTheme } from "@/context/ThemeContext";

interface AstNodeDetail {
  id: string;
  name: string;
  type: string;
  size: string;
  subsystemPath: string;
  sha: string;
  statusText: string;
  executionSteps: {
    name: string;
    target: string;
    timing: string;
    status: string;
    description: string;
    astToken: string;
  }[];
  citations: {
    ref: string;
    file: string;
    line: string;
    description: string;
    codeSnippet: string;
    stepIndex: number;
    tokenType: string;
  }[];
}

const AST_NODES: Record<string, AstNodeDetail> = {
  "app-render.tsx": {
    id: "app-render.tsx",
    name: "app-render.tsx",
    type: "Server Pipe Root",
    size: "38KB",
    subsystemPath: "packages/next/src/server/app-render/app-render.tsx",
    sha: "sha256:7b89f0a4",
    statusText: "AST ROOT COMPILED",
    executionSteps: [
      {
        name: "dispatchServerAction(request)",
        target: "POST /_next/action",
        timing: "4.2ms",
        status: "VERIFIED",
        description: "Validates Next-Action headers & decrypts bound closure arguments via Web Crypto.",
        astToken: "CallExpression [dispatchServerAction]"
      },
      {
        name: "renderToHTMLOrFlight(context)",
        target: "Flight Stream Generator",
        timing: "11.8ms",
        status: "GROUNDED",
        description: "Generates recursive React Server Component payload chunks into memory buffers.",
        astToken: "FunctionDeclaration [renderToHTMLOrFlight]"
      },
      {
        name: "flushChunkBufferToSocket()",
        target: "TCP Socket Flush",
        timing: "2.4ms",
        status: "MONITORED",
        description: "Pipes encoded Flight chunks to HTTP/2 stream with high-watermark backpressure regulation.",
        astToken: "AwaitExpression [flushChunkBuffer]"
      },
    ],
    citations: [
      {
        ref: "CIT-01",
        file: "packages/next/src/server/app-render/app-render.tsx",
        line: "L145-L162",
        description: "renderToHTMLOrFlight stream factory binding",
        codeSnippet: "export async function renderToHTMLOrFlight(req, res, pagePath, query) {\n  return createFlightStreamFactory({ req, res, pagePath });\n}",
        stepIndex: 1,
        tokenType: "AST:ExportNamedDeclaration"
      },
      {
        ref: "CIT-02",
        file: "packages/next/src/server/lib/action-handler.ts",
        line: "L44-L58",
        description: "parseServerActionHeader integrity verification",
        codeSnippet: "const actionId = req.headers.get('next-action');\nif (!actionId) throw new ActionHeaderMissingError();",
        stepIndex: 0,
        tokenType: "AST:VariableDeclaration"
      }
    ]
  },
  "action-handler.ts": {
    id: "action-handler.ts",
    name: "action-handler.ts",
    type: "Action Dispatcher",
    size: "18KB",
    subsystemPath: "packages/next/src/server/lib/action-handler.ts",
    sha: "sha256:c381da21",
    statusText: "CRYPTO BOUND VERIFIED",
    executionSteps: [
      {
        name: "extractActionBoundArgs(token)",
        target: "Web Crypto AES-GCM",
        timing: "1.8ms",
        status: "VERIFIED",
        description: "Unseals encrypted parameters passed to server functions using host AES key.",
        astToken: "CallExpression [crypto.subtle.decrypt]"
      },
      {
        name: "validateSessionHmacToken(salt)",
        target: "Constant-time Token Check",
        timing: "0.6ms",
        status: "GROUNDED",
        description: "Guards against timing attacks using crypto.timingSafeEqual comparison.",
        astToken: "CallExpression [timingSafeEqual]"
      },
      {
        name: "invokeServerBoundMutation()",
        target: "Action Closure Execution",
        timing: "22.4ms",
        status: "ACTIVE",
        description: "Dispatches server action closure and returns serialized JSON or Flight stream response.",
        astToken: "AwaitExpression [actionClosure.apply]"
      },
    ],
    citations: [
      {
        ref: "CIT-03",
        file: "packages/next/src/server/lib/action-handler.ts",
        line: "L88-L105",
        description: "decryptActionBoundArgs HMAC payload unseal",
        codeSnippet: "const decrypted = await crypto.subtle.decrypt(\n  { name: 'AES-GCM', iv },\n  key,\n  encryptedArgs\n);",
        stepIndex: 0,
        tokenType: "AST:AwaitExpression"
      },
      {
        ref: "CIT-04",
        file: "packages/next/src/server/crypto-token.ts",
        line: "L34-L52",
        description: "Constant-time HMAC session token validation",
        codeSnippet: "function timingSafeEqual(a: Uint8Array, b: Uint8Array): boolean {\n  return crypto.timingSafeEqual(a, b);\n}",
        stepIndex: 1,
        tokenType: "AST:FunctionDeclaration"
      }
    ]
  },
  "flight-stream.ts": {
    id: "flight-stream.ts",
    name: "flight-stream.ts",
    type: "Drain Controller",
    size: "28KB",
    subsystemPath: "packages/next/src/server/flight/flight-stream.ts",
    sha: "sha256:3f88be19",
    statusText: "BACKPRESSURE REGULATED",
    executionSteps: [
      {
        name: "BackpressureDrainController.init()",
        target: "Socket Buffer Monitor",
        timing: "0.4ms",
        status: "VERIFIED",
        description: "Hooks socket 'drain' event listeners to throttle serialization during burst traffic.",
        astToken: "NewExpression [BackpressureDrainController]"
      },
      {
        name: "pipeChunkToSocketBuffer(chunk)",
        target: "Flight Chunk Serialize",
        timing: "8.1ms",
        status: "ACTIVE",
        description: "Encodes binary Flight RSC frames into outgoing TCP payload buffer.",
        astToken: "CallExpression [pipeChunkToSocket]"
      },
      {
        name: "awaitClientDrainAck()",
        target: "Flow Control Throttle",
        timing: "31.2ms",
        status: "REGULATED",
        description: "Suspends streaming iteration when socket buffer exceeds high watermark (16KB).",
        astToken: "AwaitExpression [socket.once('drain')]"
      },
    ],
    citations: [
      {
        ref: "CIT-05",
        file: "packages/next/src/server/flight/flight-stream.ts",
        line: "L182-L210",
        description: "BackpressureDrainController flow regulation loop",
        codeSnippet: "if (!socket.write(chunk)) {\n  await new Promise((resolve) => socket.once('drain', resolve));\n}",
        stepIndex: 2,
        tokenType: "AST:IfStatement"
      },
      {
        ref: "CIT-06",
        file: "packages/next/src/server/socket-adapter.ts",
        line: "L91-L104",
        description: "onSocketDrainResume serialization callback",
        codeSnippet: "drainController.resumeSerialization();\nflightStream.emit('drain_ack');",
        stepIndex: 0,
        tokenType: "AST:ExpressionStatement"
      }
    ]
  },
  "fiber-workloop.ts": {
    id: "fiber-workloop.ts",
    name: "fiber-workloop.ts",
    type: "Concurrent Scheduler",
    size: "44KB",
    subsystemPath: "packages/react-reconciler/src/ReactFiberWorkLoop.js",
    sha: "sha256:8a12e5c0",
    statusText: "LANE SCHEDULE VERIFIED",
    executionSteps: [
      {
        name: "requestUpdateLane(fiber)",
        target: "Transition Lane Priority",
        timing: "0.2ms",
        status: "VERIFIED",
        description: "Computes priority bitmask for concurrent state updates via scheduler lanes.",
        astToken: "CallExpression [requestUpdateLane]"
      },
      {
        name: "renderRootConcurrent(root, lanes)",
        target: "Fiber Work Loop Iteration",
        timing: "6.4ms",
        status: "GROUNDED",
        description: "Yields to host browser event loop while incrementally reconciling virtual DOM tree.",
        astToken: "DoWhileStatement [workLoopConcurrent]"
      },
      {
        name: "commitRootImpl(root, priority)",
        target: "DOM Mutation & Passive Effects",
        timing: "3.1ms",
        status: "ACTIVE",
        description: "Flushes reconciled mutations to host document and schedules useLayoutEffect callbacks.",
        astToken: "CallExpression [commitMutationEffects]"
      },
    ],
    citations: [
      {
        ref: "CIT-07",
        file: "packages/react-reconciler/src/ReactFiberWorkLoop.js",
        line: "L480-L512",
        description: "renderRootConcurrent yield-to-host execution loop",
        codeSnippet: "do {\n  try { workLoopConcurrent(); break; }\n  catch (thrownValue) { handleError(root, thrownValue); }\n} while (true);",
        stepIndex: 1,
        tokenType: "AST:DoWhileStatement"
      },
      {
        ref: "CIT-08",
        file: "packages/react-reconciler/src/ReactFiberLane.js",
        line: "L120-L135",
        description: "getHighestPriorityLane bitmask extraction",
        codeSnippet: "export function getHighestPriorityLane(lanes: Lanes): Lane {\n  return lanes & -lanes;\n}",
        stepIndex: 0,
        tokenType: "AST:ExportNamedDeclaration"
      }
    ]
  }
};

const PR_SPECIMENS = [
  {
    id: "nextjs-68421",
    label: "PR #68421 (Next.js 15)",
    repo: "vercel/next.js",
    title: "feat(server): add backpressure drain controller for Flight socket streams",
    linesChanged: 428,
    auditTime: "45 mins manual audit",
    rawDiffLines: [
      { type: "del", text: "- export async function handleAction(req, res, chunk) {" },
      { type: "add", text: "+ export async function handleAction(req: Request, res?: Socket, drain?: Drain) {", mapsTo: "contract" },
      { type: "del", text: "-   const payload = await req.json();" },
      { type: "add", text: "+   const payload = await parseActionPayloadStream(req);" },
      { type: "add", text: "+   if (drain) await drain.waitHighWatermark();", mapsTo: "drain" },
      { type: "ctx", text: "    const token = req.headers.get('next-action');" },
      { type: "del", text: "-   return execute(payload);" },
      { type: "add", text: "+   return executeWithSessionSalt(payload, token);", mapsTo: "crypto" },
      { type: "meta", text: "@@ -210,18 +210,34 @@" },
      { type: "del", text: "-   flightBuffer.drain(socket);" },
      { type: "add", text: "+   drainController.hookSocket(socket);", mapsTo: "drain" },
      { type: "meta", text: "... 418 additional unannotated lines ..." }
    ],
    semanticImpact: {
      behavior: "Prevents server socket buffer exhaustion during high concurrency bursts by throttling Flight chunk serialization until client ACKs are received.",
      contractStatus: "NON-BREAKING (EXTENDED)",
      contractColor: "text-[#4ade80]",
      coupling: "App Router Runtime [Contained]",
      blastRadiusScore: "2.4 / 10 (LOW RISK)",
      blastRadiusPct: 24,
      subsystems: ["packages/next/src/server/flight", "packages/next/src/server/app-render"],
      testCoverage: "96.4% test coverage preserved"
    }
  },
  {
    id: "react-29840",
    label: "PR #29840 (React 19)",
    repo: "facebook/react",
    title: "scheduler: decouple transition lane priority from sync input events",
    linesChanged: 312,
    auditTime: "60 mins manual audit",
    rawDiffLines: [
      { type: "del", text: "- export function requestUpdateLane(fiber: Fiber): Lane {" },
      { type: "add", text: "+ export function requestUpdateLane(fiber: Fiber, priority: PriorityLevel): Lane {", mapsTo: "contract" },
      { type: "del", text: "-   const mode = fiber.mode;" },
      { type: "add", text: "+   const mode = fiber.mode | ConcurrentMode;" },
      { type: "del", text: "-   if ((mode & ConcurrentMode) === NoMode) return SyncLane;" },
      { type: "add", text: "+   if (isTransitionLane(priority)) return claimNextTransitionLane();", mapsTo: "drain" },
      { type: "ctx", text: "    return getCurrentUpdatePriority();" },
      { type: "meta", text: "@@ -142,8 +142,22 @@" },
      { type: "add", text: "+   entangleLanes(root, lanes);", mapsTo: "crypto" },
      { type: "meta", text: "... 304 additional unannotated lines ..." }
    ],
    semanticImpact: {
      behavior: "Prevents UI thread lockup during intensive offscreen re-renders by isolating startTransition dispatches to non-interfering scheduler lane bitmasks.",
      contractStatus: "INTERNAL RECONCILER SHIFT",
      contractColor: "text-accent",
      coupling: "React DOM Server & Client [Synchronized]",
      blastRadiusScore: "3.8 / 10 (MODERATE RISK)",
      blastRadiusPct: 38,
      subsystems: ["packages/react-reconciler", "packages/scheduler"],
      testCoverage: "98.1% test coverage verified"
    }
  },
  {
    id: "k8s-124018",
    label: "PR #124018 (Kubernetes)",
    repo: "kubernetes/kubernetes",
    title: "client-go/tools/cache: convert shared informer lock to shared RWMutex read lock",
    linesChanged: 86,
    auditTime: "25 mins manual audit",
    rawDiffLines: [
      { type: "del", text: "- func (c *cache) Get(key string) (interface{}, bool, error) {" },
      { type: "add", text: "+ func (c *cache) Get(key string) (interface{}, bool, error) {", mapsTo: "contract" },
      { type: "del", text: "-   c.lock.Lock()" },
      { type: "add", text: "+   c.lock.RLock()", mapsTo: "drain" },
      { type: "del", text: "-   defer c.lock.Unlock()" },
      { type: "add", text: "+   defer c.lock.RUnlock()", mapsTo: "drain" },
      { type: "ctx", text: "    item, exists := c.items[key]" },
      { type: "meta", text: "@@ -88,14 +88,26 @@" },
      { type: "add", text: "+   atomic.AddInt64(&c.readHits, 1)", mapsTo: "crypto" },
      { type: "meta", text: "... 78 additional unannotated lines ..." }
    ],
    semanticImpact: {
      behavior: "Eliminates read lock contention in kube-controller-manager by replacing exclusive mutexes with concurrent read-safe sync.RWMutex implementations.",
      contractStatus: "ZERO SIGNATURE DRIFT (CLEAN)",
      contractColor: "text-[#4ade80]",
      coupling: "Shared Informer Factory [Contained]",
      blastRadiusScore: "1.2 / 10 (MINIMAL RISK)",
      blastRadiusPct: 12,
      subsystems: ["staging/src/k8s.io/client-go/tools/cache"],
      testCoverage: "100% unit tests passing"
    }
  }
];

const CLI_SNIPPETS = [
  {
    tab: "explore",
    command: "gitwise explore vercel/next.js",
    output: [
      "→ Indexing 312 files with tree-sitter-typescript... [done in 140ms]",
      "→ AST Graph compiled: 1,420 functions, 89 routes, 42 server actions",
      "✓ Entry point detected: packages/next/src/server/app-render/app-render.tsx",
      "✓ Launching interactive terminal AST visualizer..."
    ]
  },
  {
    tab: "trace",
    command: "gitwise trace 'dispatchServerAction'",
    output: [
      "→ Tracing call graph for symbol: dispatchServerAction",
      "  ├── [01] POST /_next/action (HTTP/2 entry)",
      "  ├── [02] parseServerActionHeader() [lib/action-handler.ts:L44]",
      "  ├── [03] decryptActionBoundArgs() [Web Crypto AES-GCM]",
      "  └── [04] renderToHTMLOrFlight() [app-render.tsx:L145]",
      "✓ Call graph resolved without cyclic deadlocks (Depth: 4)"
    ]
  },
  {
    tab: "pr-review",
    command: "gitwise pr review 68421",
    output: [
      "→ Fetching diff for PR #68421: 'feat(server): add backpressure drain controller'",
      "→ AST blast radius: 3 files affected, 428 lines modified",
      "✓ API Contract: NON-BREAKING (EXTENDED)",
      "✓ Subsystem coupling: App Router Runtime (Contained)",
      "✓ Grounded evidence generated: 2 citation links"
    ]
  },
  {
    tab: "blueprint",
    command: "gitwise issue blueprint 1402",
    output: [
      "→ Analyzing issue #1402: 'Memory spike during concurrent RSC streaming'",
      "→ Root cause identified: Socket buffer unthrottled during large payloads",
      "→ Recommended patch sequence: 3 files",
      "  1. packages/next/src/server/flight/flight-stream.ts (+14 lines)",
      "  2. packages/next/src/server/socket-adapter.ts (+8 lines)",
      "  3. test/e2e/app-dir/flight-backpressure.test.ts (+42 lines)",
      "✓ Blueprint ready. Run with --apply to scaffold changes."
    ]
  }
];

const LIVE_COMMITS = [
  { repo: "vercel/next.js", branch: "canary", sha: "7b89f0a", file: "app-render.tsx", risk: "2.4/10 Low", author: "shuding" },
  { repo: "facebook/react", branch: "main", sha: "8a12e5c", file: "ReactFiberWorkLoop.js", risk: "3.8/10 Med", author: "acdlite" },
  { repo: "kubernetes/kubernetes", branch: "master", sha: "c381da2", file: "cache.go", risk: "1.2/10 Min", author: "thockin" },
  { repo: "pallets/flask", branch: "main", sha: "3f88be1", file: "app.py", risk: "1.0/10 Clean", author: "davidism" },
  { repo: "tailwindlabs/tailwindcss", branch: "next", sha: "e4e4e71", file: "parser.rs", risk: "2.1/10 Low", author: "adamwathan" }
];

export default function GitWiseLandingPage() {
  const { cycleTheme, currentThemeMeta } = useTheme();

  // Ingestion Simulator State
  const [selectedPreset, setSelectedPreset] = useState("vercel/next.js");
  const [cliInput, setCliInput] = useState("gitwise explore vercel/next.js");
  const [isSimulatingIngestion, setIsSimulatingIngestion] = useState(false);
  const [ingestionProgress, setIngestionProgress] = useState(0);
  const [ingestionStageText, setIngestionStageText] = useState("");
  const [ingestionComplete, setIngestionComplete] = useState(false);

  // Dissection Chamber State
  const [activeAstNodeKey, setActiveAstNodeKey] = useState("app-render.tsx");
  const [activeStepIndex, setActiveStepIndex] = useState(0);
  const [isPlayingTrace, setIsPlayingTrace] = useState(false);

  // Comparative Diff State
  const [selectedPrSpecimenIndex, setSelectedPrSpecimenIndex] = useState(0);
  const [sliderPosition, setSliderPosition] = useState(50);
  const [diffViewMode, setDiffViewMode] = useState<"slider" | "split" | "synthesized">("slider");
  const [hoveredDiffTag, setHoveredDiffTag] = useState<string | null>(null);

  // Unified Engines Playground State
  const [activeEngineTab, setActiveEngineTab] = useState(0);

  // CLI State
  const [activeCliTab, setActiveCliTab] = useState(0);

  // Live Telemetry Feed
  const [liveCommitIndex, setLiveCommitIndex] = useState(0);

  // Presets list
  const presetRepos = [
    { key: "vercel/next.js", label: "@next.js", desc: "App Router + Server Actions" },
    { key: "facebook/react", label: "@react", desc: "Fiber Scheduler + Reconciler" },
    { key: "kubernetes/kubernetes", label: "@kubernetes", desc: "Informer + Controller Runtime" },
    { key: "pallets/flask", label: "@flask", desc: "WSGI Dispatcher + Context Locals" },
    { key: "meshery/meshery", label: "@meshery", desc: "Service Mesh Adapter" },
  ];

  const handleSelectPreset = (repoKey: string) => {
    setSelectedPreset(repoKey);
    setCliInput(`gitwise explore ${repoKey}`);
    if (repoKey === "facebook/react") {
      setActiveAstNodeKey("fiber-workloop.ts");
    } else {
      setActiveAstNodeKey("app-render.tsx");
    }
    setActiveStepIndex(0);
    setIngestionComplete(false);
  };

  // Run Live Ingestion Simulation
  const runIngestionSimulation = (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (isSimulatingIngestion) return;

    setIsSimulatingIngestion(true);
    setIngestionProgress(0);
    setIngestionComplete(false);

    setIngestionStageText("Connecting to GitHub Tree-Sitter grammar...");

    setTimeout(() => {
      setIngestionProgress(32);
      setIngestionStageText("Parsing 312 source modules with Rust Tree-Sitter...");
    }, 200);

    setTimeout(() => {
      setIngestionProgress(68);
      setIngestionStageText("Resolving function call DAG and symbol bindings...");
    }, 450);

    setTimeout(() => {
      setIngestionProgress(100);
      setIngestionStageText("AST compiled: 1,420 functions indexed & 100% grounded.");
      setIsSimulatingIngestion(false);
      setIngestionComplete(true);
    }, 750);
  };

  // Auto-play Trace Loop in Dissection Chamber
  useEffect(() => {
    let interval: NodeJS.Timeout;
    if (isPlayingTrace) {
      interval = setInterval(() => {
        const currentNode = AST_NODES[activeAstNodeKey] || AST_NODES["app-render.tsx"];
        setActiveStepIndex((prev) => (prev + 1) % currentNode.executionSteps.length);
      }, 1600);
    }
    return () => clearInterval(interval);
  }, [isPlayingTrace, activeAstNodeKey]);

  // Live Telemetry Commit Ticker
  useEffect(() => {
    const ticker = setInterval(() => {
      setLiveCommitIndex((prev) => (prev + 1) % LIVE_COMMITS.length);
    }, 3800);
    return () => clearInterval(ticker);
  }, []);

  // Global Keyboard Shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't trigger if user is typing in an input
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }

      if (e.key === "t" || e.key === "T") {
        cycleTheme();
      } else if (e.key === " ") {
        e.preventDefault();
        setIsPlayingTrace((prev) => !prev);
      } else if (["1", "2", "3", "4", "5"].includes(e.key)) {
        const stageId = `stage-0${e.key}`;
        const el = document.getElementById(stageId);
        if (el) {
          el.scrollIntoView({ behavior: "smooth" });
        }
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [cycleTheme]);

  const activeNodeData = AST_NODES[activeAstNodeKey] || AST_NODES["app-render.tsx"];
  const currentPr = PR_SPECIMENS[selectedPrSpecimenIndex];
  const currentLiveCommit = LIVE_COMMITS[liveCommitIndex];

  return (
    <div className="relative min-h-screen bg-transparent text-[#fafafa] font-mono selection:bg-accent selection:text-[#050505] flex flex-col">
      {/* 0. Global Theme-Reactive Background System (Canvas Topology) */}
      <CompilerBackground />

      <div className="relative z-10 flex flex-col min-h-screen">
        {/* 1. Global Terminal Header (Sticky & Lightweight) */}
        <header className="border-b border-[#262626] bg-[#0c0c0c]/90 backdrop-blur-md sticky top-0 z-50 px-4 sm:px-6 py-2.5 flex items-center justify-between text-xs">
          <div className="flex items-center gap-3">
            <Link
              href="/"
              className="flex items-center gap-2 font-bold text-sm tracking-tight hover:opacity-80 transition-opacity focus-visible:outline-accent"
            >
              <span className="text-accent select-none font-black">&gt;</span>
              <span className="text-[#fafafa] tracking-wider font-bold">GitWise</span>
              <span className="text-[10px] text-[#a3a3a3] border border-[#262626] px-1 py-0.2 select-none">
                v3.4
              </span>
            </Link>

            <nav className="hidden lg:flex items-center gap-1 text-[11px] text-[#b3b3b3] ml-4">
              <Link
                href="/repo"
                className="px-2 py-0.5 hover:text-accent hover:bg-[#141414] transition-colors border border-transparent hover:border-[#262626]"
              >
                [ ARCHITECTURE ]
              </Link>
              <Link
                href="/issues"
                className="px-2 py-0.5 hover:text-accent hover:bg-[#141414] transition-colors border border-transparent hover:border-[#262626]"
              >
                [ ISSUES ]
              </Link>
              <Link
                href="/pr"
                className="px-2 py-0.5 hover:text-accent hover:bg-[#141414] transition-colors border border-transparent hover:border-[#262626]"
              >
                [ PR REVIEWER ]
              </Link>
              <Link
                href="/mentor"
                className="px-2 py-0.5 hover:text-accent hover:bg-[#141414] transition-colors border border-transparent hover:border-[#262626]"
              >
                [ AI MENTOR ]
              </Link>
              <Link
                href="/gitstat"
                className="px-2 py-0.5 hover:text-accent hover:bg-[#141414] transition-colors border border-transparent hover:border-[#262626]"
              >
                [ GITSTAT ]
              </Link>
            </nav>
          </div>

          <div className="flex items-center gap-2.5 text-[11px]">
            {/* Quick Keyboard Shortcut Tooltip */}
            <div className="hidden xl:flex items-center gap-1.5 text-[9px] text-[#737373] border border-[#262626] px-2 py-0.5">
              <span>KEYS:</span>
              <kbd className="text-[#cccccc] bg-[#141414] px-1">[T] PALETTE</kbd>
              <kbd className="text-[#cccccc] bg-[#141414] px-1">[SPACE] TRACE</kbd>
              <kbd className="text-[#cccccc] bg-[#141414] px-1">[1-5] STAGES</kbd>
            </div>

            {/* Dynamic Theme Preset Picker */}
            <ThemeSelector />

            <div className="hidden sm:flex items-center gap-1.5 text-accent text-[10px] border border-accent-border px-2 py-0.5 bg-accent-soft">
              <span className="w-1.5 h-1.5 bg-accent"></span>
              <span>CLI: PREVIEW</span>
            </div>

            <div className="hidden md:flex items-center gap-1.5 text-accent-secondary text-[10px] border border-accent-secondary-border px-2 py-0.5 bg-accent-secondary-soft">
              <span className="w-1.5 h-1.5 bg-accent-secondary animate-pulse"></span>
              <span>WASM READY</span>
            </div>

            <Link
              href="/repo"
              className="px-2.5 py-1 border border-accent bg-accent-soft text-accent hover:bg-accent hover:text-[#050505] font-bold text-[10px] transition-all focus-visible:outline-accent active:translate-y-px flex items-center gap-1"
            >
              <span>LAUNCH WORKBENCH</span>
              <span className="text-[9px] opacity-75">[↵]</span>
            </Link>
          </div>
        </header>

        {/* PIPELINE REGISTRATION GUIDE: STAGE 01 */}
        <div id="stage-01" className="border-b border-[#262626] bg-[#070707]/75 backdrop-blur-sm px-4 sm:px-6 py-1.5 text-[10px] text-[#b3b3b3] flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-accent font-bold tracking-wider">
              + --- PIPELINE // 01 · TARGET INGESTION &amp; COMPILER GROUNDING
            </span>
            <span className="hidden md:inline text-[#737373]">
              -------------------------------------------
            </span>
          </div>
          <div className="flex items-center gap-3">
            <span>PARSER: RUST TREE-SITTER &amp; SWC</span>
            <span className="text-accent-secondary font-semibold">&bull; 100% GROUNDED</span>
          </div>
        </div>

        {/* 2. Elevated Hero Section (Authoritative Typography, Crisp Hierarchy) */}
        <section className="relative px-4 sm:px-6 lg:px-12 pt-14 pb-12 flex flex-col items-center text-center border-b border-[#262626] bg-transparent">
          {/* Compiler Status Stamp */}
          <div className="inline-flex items-center gap-2 border border-[#262626] bg-[#0c0c0c]/95 px-3 py-1 mb-6 text-[10px] text-[#cccccc] shadow-lg">
            <span className="w-1.5 h-1.5 bg-accent animate-pulse"></span>
            <span className="text-accent font-bold">GITWISE v3.4 IR ENGINE</span>
            <span className="text-[#737373]">|</span>
            <span className="text-accent-secondary font-medium">ZERO-HALLUCINATION AST</span>
            <span className="text-[#737373]">|</span>
            <span className="text-[#4ade80] font-medium">DETERMINISTIC &amp; GROUNDED</span>
          </div>

          {/* Commanding Headline */}
          <h1 className="text-2xl sm:text-4xl lg:text-5xl font-bold tracking-tight text-[#fafafa] max-w-4xl leading-tight mb-4 uppercase">
            Compiler-Grounded Intelligence<br className="hidden sm:inline" /> For The Open Source Web
          </h1>

          {/* Supporting Copy with high contrast and readable line-height */}
          <p className="text-xs sm:text-sm text-[#b8b8b8] max-w-2xl leading-relaxed mb-8">
            GitWise indexes, maps, and reasons through complex open-source codebases to accelerate developer comprehension—from discovery and feature tracing to issue implementation blueprints and semantic pull request reviews.
          </p>

          {/* Interactive Ingestion Terminal Box (Immediate Product Entry Point) */}
          <div className="w-full max-w-3xl border border-[#262626] hover:border-accent-border bg-[#0a0a0a]/95 p-3.5 text-left mb-3 shadow-2xl relative transition-all">
            <div className="flex items-center justify-between border-b border-[#1f1f1f] pb-2 mb-2.5 text-[10px] text-[#b3b3b3]">
              <div className="flex items-center gap-2">
                <span className="text-accent font-bold">gitwise&gt;</span>
                <span className="text-[#cccccc]">interactive target ingestion</span>
                <span className="text-accent-secondary text-[9px] border border-accent-secondary-border bg-accent-secondary-soft px-1.5 font-bold">
                  [PARSER: ZERO-CLONE READY]
                </span>
              </div>
              <div className="flex items-center gap-1.5">
                <span className="text-[#737373]">benchmark presets:</span>
                {presetRepos.map((preset) => (
                  <button
                    key={preset.key}
                    onClick={() => handleSelectPreset(preset.key)}
                    className={`px-1.5 py-0.5 border text-[9px] transition-colors focus-visible:outline-accent ${
                      selectedPreset === preset.key
                        ? "border-accent text-accent bg-accent-soft font-bold"
                        : "border-[#262626] text-[#a3a3a3] hover:border-[#525252] hover:text-[#fafafa]"
                    }`}
                    title={preset.desc}
                  >
                    {preset.label}
                  </button>
                ))}
              </div>
            </div>

            {/* Ingestion Command Form */}
            <form onSubmit={runIngestionSimulation} className="flex items-center gap-2 bg-[#040404] border border-[#1f1f1f] focus-within:border-accent px-3 py-2 text-xs transition-colors">
              <span className="text-accent select-none font-bold">$</span>
              <input
                type="text"
                value={cliInput}
                onChange={(e) => {
                  setCliInput(e.target.value);
                  setIngestionComplete(false);
                }}
                className="flex-1 bg-transparent text-[#fafafa] font-mono focus:outline-none placeholder-[#737373]"
                placeholder="gitwise explore owner/repository"
              />
              <button
                type="submit"
                disabled={isSimulatingIngestion}
                className="px-3 py-1 bg-accent-soft border border-accent text-accent text-[10px] font-bold hover:bg-accent hover:text-[#050505] transition-colors flex items-center gap-1.5 flex-shrink-0 active:translate-y-px disabled:opacity-50"
              >
                {isSimulatingIngestion ? (
                  <>
                    <span className="w-2 h-2 border-2 border-current border-t-transparent animate-spin rounded-full"></span>
                    <span>INDEXING...</span>
                  </>
                ) : (
                  <>
                    <span>ANALYZE REPOSITORY</span>
                    <span className="text-[9px] opacity-75">[↵]</span>
                  </>
                )}
              </button>
            </form>

            {/* Ingestion Progress Bar (Active when compiling) */}
            {isSimulatingIngestion && (
              <div className="mt-2.5 pt-2 border-t border-[#1a1a1a]">
                <div className="flex items-center justify-between text-[10px] text-[#cccccc] mb-1">
                  <span className="text-accent flex items-center gap-1.5">
                    <span className="w-1.5 h-1.5 bg-accent animate-pulse"></span>
                    {ingestionStageText}
                  </span>
                  <span className="font-bold">{ingestionProgress}%</span>
                </div>
                <div className="w-full bg-[#141414] h-1 border border-[#262626]">
                  <div
                    className="bg-accent h-full transition-all duration-300"
                    style={{ width: `${ingestionProgress}%` }}
                  ></div>
                </div>
              </div>
            )}

            {/* Ingestion Result Banner (Appears when simulated analysis is complete) */}
            {ingestionComplete && !isSimulatingIngestion && (
              <div className="mt-2.5 pt-2 border-t border-[#1f3a24] bg-[#07120a] p-2 flex flex-wrap items-center justify-between gap-2 text-[10px]">
                <div className="flex items-center gap-2 text-[#4ade80]">
                  <CheckCircle2 className="w-3.5 h-3.5 text-[#4ade80]" />
                  <span>
                    <strong>{selectedPreset}</strong> AST Indexing Complete &middot; 312 files &middot; 1,420 functions &middot; 0 cyclic faults
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => {
                      const el = document.getElementById("stage-02");
                      if (el) el.scrollIntoView({ behavior: "smooth" });
                    }}
                    className="px-2 py-0.5 border border-[#14532d] text-[#4ade80] hover:bg-[#14532d] hover:text-[#fafafa] transition-colors"
                  >
                    INSPECT DISSECTION CHAMBER ↓
                  </button>
                  <Link
                    href="/repo"
                    className="px-2 py-0.5 border border-accent bg-accent-soft text-accent hover:bg-accent hover:text-[#050505] font-bold transition-colors"
                  >
                    OPEN WORKBENCH [↵]
                  </Link>
                </div>
              </div>
            )}

            {/* Under-Terminal Micro-Telemetry */}
            <div className="flex flex-wrap items-center justify-between text-[10px] text-[#a3a3a3] pt-2.5 mt-2 border-t border-[#1a1a1a]">
              <div className="flex items-center gap-3">
                <span>
                  TARGET: <strong className="text-[#fafafa]">{selectedPreset}</strong>
                </span>
                <span className="text-[#737373]">|</span>
                <span>
                  DIALECT: <strong className="text-accent">TypeScript 5.5 + JSX</strong>
                </span>
                <span className="hidden sm:inline text-[#737373]">|</span>
                <span className="hidden sm:inline">
                  MODE: <strong className="text-[#fafafa]">Zero-Clone In-Browser WASM</strong>
                </span>
              </div>
              <div className="flex items-center gap-2 text-[#4ade80] font-medium">
                <span>100% DETERMINISTIC</span>
                <span>&bull;</span>
                <span>NO HALLUCINATION</span>
              </div>
            </div>
          </div>

          {/* Clear Callout Line */}
          <p className="text-[11px] text-[#a3a3a3] mb-1">
            Paste any public GitHub repository to extract AST topology, feature execution paths, and verifiable implementation blueprints.
          </p>
        </section>

        {/* PIPELINE REGISTRATION GUIDE: STAGE 02 */}
        <div id="stage-02" className="border-b border-[#262626] bg-[#070707]/75 backdrop-blur-sm px-4 sm:px-6 py-1.5 text-[10px] text-[#b3b3b3] flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-accent font-bold tracking-wider">
              + --- PIPELINE // 02 · COMPILER DISSECTION CHAMBER [LIVE SPECIMEN]
            </span>
            <span className="hidden md:inline text-[#737373]">
              -------------------------------------------
            </span>
          </div>
          <div className="flex items-center gap-3">
            <span>SOURCE (AST) &rarr; REASONING (GRAPH) &rarr; EVIDENCE (CITATIONS)</span>
            <span className="text-[#4ade80] font-semibold">&bull; INTERACTIVE PROTOTYPE</span>
          </div>
        </div>

        {/* 3. Interactive Live Compiler Dissection Chamber */}
        <section className="px-4 sm:px-6 lg:px-12 py-12 border-b border-[#262626] bg-[#050505]/45 backdrop-blur-[2px]">
          <div className="max-w-5xl mx-auto flex flex-col gap-4">
            <div className="border border-[#262626] bg-[#080808] p-3.5 text-left shadow-2xl">
              {/* Specimen Header & Trace Controller */}
              <div className="border-b border-[#262626] pb-2.5 mb-3 flex flex-wrap items-center justify-between gap-2 text-[11px]">
                <div className="flex items-center gap-2">
                  <Layers className="w-3.5 h-3.5 text-accent" />
                  <span className="font-bold text-[#fafafa] uppercase tracking-wide">
                    COMPILER DISSECTION CHAMBER: {selectedPreset}
                  </span>
                  <span className="text-[10px] text-accent border border-accent-border px-1.5 py-0.2 bg-accent-soft font-semibold">
                    NODE: {activeNodeData.name}
                  </span>
                  <span className="text-[10px] text-accent-secondary border border-accent-secondary-border px-1.5 py-0.2 bg-accent-secondary-soft font-semibold hidden md:inline">
                    SUBSYSTEM: {activeNodeData.type}
                  </span>
                </div>

                <div className="flex items-center gap-2 text-[10px]">
                  {/* Trace Simulator Button */}
                  <button
                    onClick={() => setIsPlayingTrace(!isPlayingTrace)}
                    className={`flex items-center gap-1.5 px-2 py-0.5 border transition-all ${
                      isPlayingTrace
                        ? "border-accent-secondary bg-accent-secondary-soft text-accent-secondary font-bold animate-pulse"
                        : "border-[#333333] hover:border-accent text-[#cccccc] hover:text-accent bg-[#0c0c0c]"
                    }`}
                    title="Simulate step-by-step runtime execution flow"
                  >
                    {isPlayingTrace ? (
                      <>
                        <Pause className="w-3 h-3 text-accent-secondary" />
                        <span>PAUSE TRACE</span>
                      </>
                    ) : (
                      <>
                        <Play className="w-3 h-3 text-accent" />
                        <span>SIMULATE TRACE [SPACE]</span>
                      </>
                    )}
                  </button>

                  <span className="text-[#4ade80] font-semibold hidden sm:inline">&bull; {activeNodeData.statusText}</span>
                  <span className="text-[#737373] hidden sm:inline">|</span>
                  <span className="text-accent-secondary font-mono text-[9px] hidden sm:inline">HASH: {activeNodeData.sha}</span>
                </div>
              </div>

              {/* Three-Panel Chamber Layout */}
              <div className="grid grid-cols-1 md:grid-cols-12 gap-3.5 text-xs">
                {/* Panel A: AST Node Hierarchy */}
                <div className="md:col-span-4 border border-[#1f1f1f] bg-[#040404] p-3 flex flex-col gap-2">
                  <div className="flex items-center justify-between border-b border-[#1a1a1a] pb-1.5 text-[10px]">
                    <span className="text-[#cccccc] font-bold uppercase tracking-wider flex items-center gap-1">
                      <FileCode className="w-3 h-3 text-accent" />
                      AST NODE HIERARCHY
                    </span>
                    <span className="text-accent-secondary font-semibold">312 FILES INDEXED</span>
                  </div>
                  <div className="flex flex-col gap-1 text-[10px]">
                    {Object.values(AST_NODES).map((node) => {
                      const isSelected = activeAstNodeKey === node.id;
                      return (
                        <button
                          key={node.id}
                          onClick={() => {
                            setActiveAstNodeKey(node.id);
                            setActiveStepIndex(0);
                            setIsPlayingTrace(false);
                          }}
                          className={`p-2 text-left border flex items-center justify-between transition-all focus-visible:outline-accent ${
                            isSelected
                              ? "border-accent bg-accent-soft text-accent font-semibold"
                              : "border-[#141414] hover:border-[#383838] text-[#cccccc] hover:text-[#fafafa] bg-[#070707]"
                          }`}
                        >
                          <span className="truncate">
                            {isSelected ? "> " : "  "}
                            {node.name} <span className="text-[#a3a3a3]">({node.type})</span>
                          </span>
                          <span className={`text-[9px] font-mono ${isSelected ? "text-accent-secondary font-bold" : "text-[#737373]"}`}>
                            {node.size}
                          </span>
                        </button>
                      );
                    })}
                  </div>
                  <div className="text-[10px] text-[#a3a3a3] pt-1.5 border-t border-[#1a1a1a] flex items-center justify-between">
                    <span>Click node to switch source specimen</span>
                    <span className="text-accent">&bull;</span>
                  </div>
                </div>

                {/* Panel B: Function Execution Graph (Clickable Steps) */}
                <div className="md:col-span-5 border border-[#1f1f1f] bg-[#040404] p-3 flex flex-col gap-2">
                  <div className="flex items-center justify-between border-b border-[#1a1a1a] pb-1.5 text-[10px]">
                    <span className="text-[#cccccc] font-bold uppercase tracking-wider flex items-center gap-1">
                      <Network className="w-3 h-3 text-accent" />
                      FUNCTION EXECUTION GRAPH
                    </span>
                    <span className="text-accent-secondary font-semibold">TOPOLOGICAL FLOW</span>
                  </div>
                  <div className="flex flex-col gap-2 p-0.5 text-[10px]">
                    {activeNodeData.executionSteps.map((step, idx) => {
                      const isStepActive = activeStepIndex === idx;
                      return (
                        <button
                          key={idx}
                          onClick={() => {
                            setActiveStepIndex(idx);
                            setIsPlayingTrace(false);
                          }}
                          className={`p-2 border text-left flex flex-col gap-1 transition-all focus-visible:outline-accent ${
                            isStepActive
                              ? "border-accent bg-accent-soft text-accent shadow-md translate-x-0.5 border-l-2 border-l-accent"
                              : "border-[#1a1a1a] bg-[#080808] hover:border-[#383838] text-[#cccccc]"
                          }`}
                        >
                          <div className="flex items-center justify-between w-full">
                            <div className="flex items-center gap-1.5 truncate">
                              <span className={`text-[9px] font-bold ${isStepActive ? "text-accent" : "text-[#737373]"}`}>
                                0{idx + 1}.
                              </span>
                              <span className={isStepActive ? "text-accent font-bold truncate" : "text-[#fafafa] truncate font-medium"}>
                                {step.name}
                              </span>
                            </div>
                            <div className="flex items-center gap-2 flex-shrink-0 text-[9px]">
                              <span className="text-accent-secondary border border-accent-secondary-border bg-accent-secondary-soft px-1 font-semibold">{step.timing}</span>
                              <span className="text-[#4ade80] border border-[#14532d] bg-[#08170c] px-1 font-semibold">
                                {step.status}
                              </span>
                            </div>
                          </div>
                          <p className={`text-[9px] leading-tight ${isStepActive ? "text-[#e5e5e5]" : "text-[#8a8a8a]"}`}>
                            {step.description}
                          </p>
                          <div className="flex items-center justify-between text-[8px] text-[#737373] pt-0.5">
                            <span>{step.target}</span>
                            <span className="text-accent-secondary font-mono">{step.astToken}</span>
                          </div>
                        </button>
                      );
                    })}
                  </div>
                  <div className="flex items-center justify-between text-[10px] text-[#a3a3a3] pt-1.5 border-t border-[#1a1a1a]">
                    <span>STEP FOCUS: <strong className="text-accent">0{activeStepIndex + 1}</strong> OF {activeNodeData.executionSteps.length}</span>
                    <span className="text-accent-secondary font-semibold">DAG DEPTH: LINEAR</span>
                  </div>
                </div>

                {/* Panel C: Verifiable Citations (Grounded Source Evidence) */}
                <div className="md:col-span-3 border border-[#1f1f1f] bg-[#040404] p-3 flex flex-col gap-2">
                  <div className="flex items-center justify-between border-b border-[#1a1a1a] pb-1.5 text-[10px]">
                    <span className="text-[#cccccc] font-bold uppercase tracking-wider flex items-center gap-1">
                      <ShieldCheck className="w-3 h-3 text-[#4ade80]" />
                      VERIFIABLE CITATIONS
                    </span>
                    <span className="text-[#4ade80] font-semibold">GROUNDED</span>
                  </div>
                  <div className="flex flex-col gap-2 text-[10px] text-[#cccccc]">
                    {activeNodeData.citations.map((cit) => {
                      const isRelatedStep = cit.stepIndex === activeStepIndex;
                      return (
                        <div
                          key={cit.ref}
                          className={`border p-2 flex flex-col gap-1 transition-all ${
                            isRelatedStep
                              ? "border-accent bg-accent-soft shadow-sm"
                              : "border-[#141414] bg-[#080808]"
                          }`}
                        >
                          <div className="flex items-center justify-between text-[9px]">
                            <span className={`font-bold ${isRelatedStep ? "text-accent" : "text-[#cccccc]"}`}>
                              [{cit.ref} : {cit.line}]
                            </span>
                            <span className="text-accent-secondary truncate max-w-[100px] font-mono">
                              {cit.file.split("/").pop()}
                            </span>
                          </div>
                          <p className="text-[#fafafa] leading-tight text-[10px]">
                            {cit.description}
                          </p>
                          <pre className="text-[9px] text-[#d4d4d4] bg-[#040404] p-1.5 border border-[#1a1a1a] overflow-x-auto select-all leading-tight font-mono">
                            {cit.codeSnippet}
                          </pre>
                          <span className="text-[8px] text-accent-secondary font-mono mt-0.5">
                            {cit.tokenType}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                  <Link
                    href="/repo"
                    className="mt-auto px-2.5 py-1.5 border border-[#333333] hover:border-accent text-[#b3b3b3] hover:text-accent text-[10px] transition-colors text-center block font-semibold active:translate-y-px"
                  >
                    OPEN ARCHITECTURE WORKBENCH &rarr;
                  </Link>
                </div>
              </div>

              {/* Bottom 3-Stage Compiler Flow Strip */}
              <div className="mt-3 pt-2.5 border-t border-[#1a1a1a] flex flex-wrap items-center justify-between gap-2 text-[10px]">
                <div className="flex items-center gap-2">
                  <span className="text-accent font-bold">[01 AST: {activeNodeData.name}]</span>
                  <span className="text-[#737373]">&rarr;</span>
                  <span className="text-accent-secondary font-bold">[02 GRAPH: STEP 0{activeStepIndex + 1}]</span>
                  <span className="text-[#737373]">&rarr;</span>
                  <span className="text-[#4ade80] font-bold">[03 EVIDENCE: 100% GROUNDED]</span>
                </div>
                <div className="text-[9px] text-[#8a8a8a]">
                  STATUS: <span className="text-accent-secondary font-mono">DETERMINISTIC COMPILER REASONING ACTIVE</span>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* PIPELINE REGISTRATION GUIDE: STAGE 03 */}
        <div id="stage-03" className="border-b border-[#262626] bg-[#070707]/75 backdrop-blur-sm px-4 sm:px-6 py-1.5 text-[10px] text-[#b3b3b3] flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-accent font-bold tracking-wider">
              + --- PIPELINE // 03 · COMPARATIVE SYNTHESIS [NOISE VS REASONED IMPACT]
            </span>
            <span className="hidden md:inline text-[#737373]">
              -------------------------------------------
            </span>
          </div>
          <div className="flex items-center gap-3">
            <span>NOISE VS REASONED IMPACT</span>
            <span className="text-[#4ade80] font-semibold">&bull; 45 MINS &rarr; 18 SECONDS</span>
          </div>
        </div>

        {/* 4. Interactive Comparative Semantic Slider (Raw Diff vs. GitWise Impact) */}
        <section className="px-4 sm:px-6 lg:px-12 py-12 border-b border-[#262626] bg-[#050505]/50 backdrop-blur-[2px]">
          <div className="max-w-5xl mx-auto flex flex-col gap-4">
            <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[#262626] pb-3">
              <div>
                <span className="text-accent text-[10px] font-bold uppercase tracking-wider block mb-0.5">
                  THE ARCHITECTURAL DIFFERENCE
                </span>
                <h2 className="text-lg sm:text-xl font-bold text-[#fafafa]">
                  RAW 10,000-LINE DIFF VS. GITWISE SEMANTIC IMPACT
                </h2>
                <p className="text-xs text-[#b8b8b8] mt-1 leading-relaxed">
                  Mechanical line deltas conceal architectural consequences. GitWise parses downstream blast radius and contract shifts.
                </p>
              </div>

              {/* View Mode Controls & Multi-PR Specimen Tabs */}
              <div className="flex flex-wrap items-center gap-2 text-[10px]">
                {/* PR Specimen Picker */}
                <div className="flex items-center border border-[#262626] p-0.5 bg-[#0a0a0a]">
                  {PR_SPECIMENS.map((specimen, idx) => (
                    <button
                      key={specimen.id}
                      onClick={() => setSelectedPrSpecimenIndex(idx)}
                      className={`px-2 py-0.5 text-[9px] transition-colors ${
                        selectedPrSpecimenIndex === idx
                          ? "bg-accent-soft text-accent font-bold"
                          : "text-[#8a8a8a] hover:text-[#fafafa]"
                      }`}
                    >
                      {specimen.label}
                    </button>
                  ))}
                </div>

                {/* View Mode Tabs */}
                <div className="flex items-center border border-[#262626] p-0.5 bg-[#0a0a0a]">
                  <button
                    onClick={() => {
                      setDiffViewMode("slider");
                      setSliderPosition(50);
                    }}
                    className={`px-2 py-0.5 transition-colors ${
                      diffViewMode === "slider"
                        ? "bg-accent-soft text-accent font-bold"
                        : "text-[#a3a3a3] hover:text-[#fafafa]"
                    }`}
                  >
                    SLIDER
                  </button>
                  <button
                    onClick={() => setDiffViewMode("split")}
                    className={`px-2 py-0.5 transition-colors ${
                      diffViewMode === "split"
                        ? "bg-accent-soft text-accent font-bold"
                        : "text-[#a3a3a3] hover:text-[#fafafa]"
                    }`}
                  >
                    SPLIT
                  </button>
                  <button
                    onClick={() => setDiffViewMode("synthesized")}
                    className={`px-2 py-0.5 transition-colors ${
                      diffViewMode === "synthesized"
                        ? "bg-accent-soft text-accent font-bold"
                        : "text-[#a3a3a3] hover:text-[#fafafa]"
                    }`}
                  >
                    SYNTHESIS
                  </button>
                </div>

                {diffViewMode === "slider" && (
                  <div className="flex items-center gap-2 ml-1">
                    <span className="text-[#f87171] text-[9px] font-semibold">RAW</span>
                    <input
                      type="range"
                      min="0"
                      max="100"
                      value={sliderPosition}
                      onChange={(e) => setSliderPosition(Number(e.target.value))}
                      className="w-20 accent-current text-accent cursor-pointer"
                      aria-label="Comparison slider"
                    />
                    <span className="text-[#4ade80] text-[9px] font-semibold">SEMANTIC</span>
                  </div>
                )}
              </div>
            </div>

            {/* Target Specimen Subtitle */}
            <div className="text-[11px] text-[#a3a3a3] flex items-center justify-between border border-[#1f1f1f] bg-[#070707] px-3 py-1.5">
              <span>TARGET SPECIMEN: <strong className="text-[#fafafa]">{currentPr.title}</strong></span>
              <span className="text-accent font-mono">REPOSITORY: {currentPr.repo}</span>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Left: Raw Diff Mechanical Noise */}
              {(diffViewMode === "split" || diffViewMode === "slider") && (
                <div
                  className={`border border-[#262626] bg-[#040404] p-3.5 flex flex-col gap-2.5 transition-all ${
                    diffViewMode === "slider" && sliderPosition > 70 ? "opacity-35" : "opacity-100"
                  }`}
                >
                  <div className="flex items-center justify-between border-b border-[#1f1f1f] pb-1.5 text-[10px]">
                    <span className="text-[#f87171] font-bold uppercase tracking-wider">
                      &times; RAW GITHUB DIFF (UNREASONED DELTA)
                    </span>
                    <span className="text-[#a3a3a3]">{currentPr.linesChanged} LINES CHANGED</span>
                  </div>
                  <div className="font-mono text-[10px] leading-relaxed text-[#b3b3b3] flex flex-col gap-1 overflow-hidden max-h-[230px]">
                    {currentPr.rawDiffLines.map((line, idx) => {
                      const isHovered = hoveredDiffTag && line.mapsTo === hoveredDiffTag;
                      return (
                        <div
                          key={idx}
                          onMouseEnter={() => line.mapsTo && setHoveredDiffTag(line.mapsTo)}
                          onMouseLeave={() => setHoveredDiffTag(null)}
                          className={`px-1 py-0.5 transition-colors cursor-pointer ${
                            line.type === "del"
                              ? "text-[#f87171] hover:bg-[#f87171]/20"
                              : line.type === "add"
                              ? isHovered
                                ? "bg-[#4ade80]/20 text-[#fafafa] font-bold"
                                : "text-[#4ade80] hover:bg-[#4ade80]/15"
                              : line.type === "meta"
                              ? "text-[#737373]"
                              : "text-[#a3a3a3]"
                          }`}
                        >
                          {line.text}
                        </div>
                      );
                    })}
                  </div>
                  <div className="text-[10px] text-[#737373] border-t border-[#1a1a1a] pt-2 flex justify-between">
                    <span>{currentPr.auditTime}</span>
                    <span className="text-[#f87171] font-semibold">NO DOWNSTREAM INSIGHT</span>
                  </div>
                </div>
              )}

              {/* Right: GitWise Semantic Impact (Structured Synthesis) */}
              {(diffViewMode === "split" || diffViewMode === "slider" || diffViewMode === "synthesized") && (
                <div
                  className={`border border-accent bg-[#0d0c07] p-3.5 flex flex-col gap-2.5 transition-all ${
                    diffViewMode === "synthesized" ? "md:col-span-2" : ""
                  } ${diffViewMode === "slider" && sliderPosition < 30 ? "opacity-35" : "opacity-100"}`}
                >
                  <div className="flex items-center justify-between border-b border-[#262626] pb-1.5 text-[10px]">
                    <span className="text-[#4ade80] font-bold uppercase tracking-wider">
                      ✓ GITWISE REASONED IMPACT (COMPILER VALIDATED)
                    </span>
                    <span className="text-accent font-bold">18 SECONDS SYNTHESIS</span>
                  </div>
                  <div className="text-[11px] leading-relaxed text-[#fafafa] flex flex-col gap-2.5">
                    <div>
                      <span className="text-accent font-bold block text-[10px] uppercase tracking-wider mb-0.5">
                        BEHAVIORAL SHIFT:
                      </span>
                      <p className="text-[#cccccc]">
                        {currentPr.semanticImpact.behavior}
                      </p>
                    </div>

                    <div className="grid grid-cols-2 gap-2 text-[10px] pt-1">
                      <div className={`border p-2 ${
                        hoveredDiffTag === "contract" ? "border-accent bg-accent-soft" : "border-[#14532d] bg-[#0c180e]"
                      } transition-colors`}>
                        <span className="text-[#4ade80] font-bold block mb-0.5">PUBLIC API CONTRACT:</span>
                        <span className={`${currentPr.semanticImpact.contractColor} font-medium`}>
                          {currentPr.semanticImpact.contractStatus}
                        </span>
                      </div>
                      <div className={`border p-2 ${
                        hoveredDiffTag === "drain" ? "border-accent-secondary bg-accent-secondary-soft shadow-md" : "border-accent-secondary-border bg-accent-secondary-soft"
                      } transition-colors`}>
                        <span className="text-accent-secondary font-bold block mb-0.5">COUPLING IMPACT:</span>
                        <span className="text-[#fafafa] font-medium">
                          {currentPr.semanticImpact.coupling}
                        </span>
                      </div>
                    </div>

                    {/* Blast Radius Visual Bar */}
                    <div className="border border-[#1a1a1a] p-2 bg-[#080808]">
                      <div className="flex items-center justify-between text-[10px] text-[#cccccc] mb-1">
                        <span>BLAST RADIUS SCORE: <strong className="text-accent-secondary">{currentPr.semanticImpact.blastRadiusScore}</strong></span>
                        <span className="text-[#a3a3a3]">{currentPr.semanticImpact.testCoverage}</span>
                      </div>
                      <div className="w-full bg-[#141414] h-1.5 border border-[#262626]">
                        <div
                          className="h-full transition-all duration-300"
                          style={{
                            width: `${currentPr.semanticImpact.blastRadiusPct}%`,
                            background: "linear-gradient(90deg, var(--accent) 0%, var(--accent-secondary) 100%)"
                          }}
                        ></div>
                      </div>
                    </div>

                    <div className="flex items-center justify-between border-t border-[#262626] pt-2 text-[10px]">
                      <span className="text-[#737373]">
                        CONTAINED IN: {currentPr.semanticImpact.subsystems[0]}
                      </span>
                      <Link
                        href="/pr"
                        className="text-accent hover:text-accent-secondary flex items-center gap-1 font-bold transition-colors"
                      >
                        <span>AUDIT IN PR REVIEWER</span>
                        <ArrowRight className="w-3.5 h-3.5" />
                      </Link>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </section>

        {/* PIPELINE REGISTRATION GUIDE: STAGE 04 */}
        <div id="stage-04" className="border-b border-[#262626] bg-[#070707]/75 backdrop-blur-sm px-4 sm:px-6 py-1.5 text-[10px] text-[#b3b3b3] flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-accent font-bold tracking-wider">
              + --- PIPELINE // 04 · UNIFIED REASONING ENGINES [5 WORKBENCHES]
            </span>
            <span className="hidden md:inline text-[#737373]">
              -------------------------------------------
            </span>
          </div>
          <div className="flex items-center gap-3">
            <span>ONE COMPILER INTELLIGENCE LAYER</span>
            <span className="text-accent-secondary font-semibold">&bull; 5 ENGINES &middot; 1 COMPILER IR</span>
          </div>
        </div>

        {/* 5. The 5 Intelligence Pillars (Connective Topology Architecture & Interactive Drawer) */}
        <section className="px-4 sm:px-6 lg:px-12 py-12 border-b border-[#262626] bg-transparent">
          <div className="max-w-5xl mx-auto flex flex-col gap-6">
            <div className="border-b border-[#262626] pb-3">
              <span className="text-accent text-[10px] font-bold uppercase tracking-wider block mb-0.5">
                UNIFIED PLATFORM PIPELINE
              </span>
              <h2 className="text-xl sm:text-2xl font-bold text-[#fafafa]">
                THE 5 CORE REASONING ENGINES
              </h2>
              <p className="text-xs text-[#b8b8b8] mt-1 leading-relaxed">
                Every engine connects to the same underlying Rust AST indexer and dependency graph. One unified intelligence layer powering discovery, implementation, code review, and developer reputation.
              </p>
            </div>

            {/* Connective Architecture Circuit Bus Diagram */}
            <div className="hidden sm:block border border-[#1f1f1f] bg-[#050505] p-3 text-[10px] font-mono text-[#a3a3a3] select-none">
              <div className="flex items-center justify-between text-[9px] text-[#737373] border-b border-[#141414] pb-1.5 mb-2">
                <span className="text-accent font-bold">{"// UNIFIED COMPILER INTERMEDIATE REPRESENTATION (IR) BUS"}</span>
                <span className="text-accent-secondary font-semibold">SHARED MEMORY BUS: ZERO-COPY AST DAG</span>
              </div>
              <div className="text-center leading-tight">
                <div className="inline-block text-accent font-bold border border-accent-border px-3 py-1 bg-accent-soft">
                  GITWISE COMPILER CORE &middot; RUST AST IR
                </div>
                <div className="text-[#525252] text-xs">│</div>
                <div className="text-[#525252] text-xs">┌───────────────────────┬───────────────────────┼───────────────────────┬───────────────────────┐</div>
                <div className="grid grid-cols-5 text-center text-[9px] gap-2 pt-1">
                  <button
                    onClick={() => setActiveEngineTab(0)}
                    className={`p-1.5 border transition-all ${
                      activeEngineTab === 0
                        ? "border-accent text-accent bg-accent-soft font-bold shadow-sm"
                        : "border-[#1a1a1a] text-[#737373] hover:border-[#383838] hover:text-[#cccccc]"
                    }`}
                  >
                    01 REPO ENGINE
                  </button>
                  <button
                    onClick={() => setActiveEngineTab(1)}
                    className={`p-1.5 border transition-all ${
                      activeEngineTab === 1
                        ? "border-accent-secondary text-accent-secondary bg-accent-secondary-soft font-bold shadow-sm"
                        : "border-[#1a1a1a] text-[#737373] hover:border-[#383838] hover:text-[#cccccc]"
                    }`}
                  >
                    02 ISSUE ENGINE
                  </button>
                  <button
                    onClick={() => setActiveEngineTab(2)}
                    className={`p-1.5 border transition-all ${
                      activeEngineTab === 2
                        ? "border-accent text-accent bg-accent-soft font-bold shadow-sm"
                        : "border-[#1a1a1a] text-[#737373] hover:border-[#383838] hover:text-[#cccccc]"
                    }`}
                  >
                    03 PR REVIEWER
                  </button>
                  <button
                    onClick={() => setActiveEngineTab(3)}
                    className={`p-1.5 border transition-all ${
                      activeEngineTab === 3
                        ? "border-accent-secondary text-accent-secondary bg-accent-secondary-soft font-bold shadow-sm"
                        : "border-[#1a1a1a] text-[#737373] hover:border-[#383838] hover:text-[#cccccc]"
                    }`}
                  >
                    04 AI MENTOR
                  </button>
                  <button
                    onClick={() => setActiveEngineTab(4)}
                    className={`p-1.5 border transition-all ${
                      activeEngineTab === 4
                        ? "border-accent text-accent bg-accent-soft font-bold shadow-sm"
                        : "border-[#1a1a1a] text-[#737373] hover:border-[#383838] hover:text-[#cccccc]"
                    }`}
                  >
                    05 GITSTAT
                  </button>
                </div>
              </div>
            </div>

            {/* Interactive Connective Topology Bus Switcher */}
            <div className="border border-[#262626] bg-[#0a0a0a] p-3 text-[10px] flex flex-col gap-3 shadow-xl">
              <div className="flex items-center justify-between border-b border-[#1a1a1a] pb-1.5">
                <span className="text-accent font-bold tracking-wider flex items-center gap-2">
                  <span className="w-2 h-2 bg-accent inline-block"></span>
                  [ BUS: GITWISE DETERMINISTIC AST DAG ]
                </span>
                <span className="text-[#cccccc]">
                  CLICK ENGINE TO INSPECT LIVE ARCHITECTURAL PREVIEW
                </span>
                <span className="text-accent-secondary font-semibold">&bull; ACTIVE TRACE BUS</span>
              </div>

              {/* Interactive Circuit Switcher Tabs */}
              <div className="grid grid-cols-2 sm:grid-cols-5 gap-2 text-[10px]">
                {[
                  { label: "01 REPO ENGINE", sub: "Topological Tracer", isSecondary: false },
                  { label: "02 ISSUE ENGINE", sub: "Diff Blueprints", isSecondary: true },
                  { label: "03 PR REVIEWER", sub: "Contract Shifts", isSecondary: false },
                  { label: "04 AI MENTOR", sub: "Socratic Guide", isSecondary: true },
                  { label: "05 GITSTAT", sub: "Footprint Telemetry", isSecondary: false }
                ].map((item, idx) => (
                  <button
                    key={idx}
                    onClick={() => setActiveEngineTab(idx)}
                    className={`p-2 border text-left flex flex-col gap-0.5 transition-all focus-visible:outline-accent ${
                      activeEngineTab === idx
                        ? item.isSecondary
                          ? "border-accent-secondary bg-accent-secondary-soft text-accent-secondary font-bold shadow-md"
                          : "border-accent bg-accent-soft text-accent font-bold shadow-md"
                        : "border-[#1f1f1f] bg-[#060606] text-[#8a8a8a] hover:border-[#383838] hover:text-[#cccccc]"
                    }`}
                  >
                    <span className="font-mono">{item.label}</span>
                    <span className="text-[9px] text-[#737373]">{item.sub}</span>
                  </button>
                ))}
              </div>

              {/* Interactive Engine Live Preview Drawer */}
              <div className="border border-[#1f1f1f] bg-[#040404] p-3 text-[11px]">
                {activeEngineTab === 0 && (
                  <div className="flex flex-col gap-2">
                    <div className="flex items-center justify-between border-b border-[#141414] pb-1 text-[10px]">
                      <span className="text-accent font-bold">01 · REPO ENGINE: LIVE CALL GRAPH TRACER</span>
                      <Link href="/repo" className="text-accent hover:underline">[ENTER WORKBENCH &rarr;]</Link>
                    </div>
                    <div className="font-mono text-[10px] text-[#cccccc] flex flex-col gap-1 p-1 bg-[#070707] border border-[#141414]">
                      <div>[ROOT] POST /_next/action (HTTP/2 Stream Handler)</div>
                      <div className="pl-4 text-accent">└── app-render.tsx : dispatchServerAction() [4.2ms]</div>
                      <div className="pl-8 text-accent-secondary">├── action-handler.ts : parseServerActionHeader() [0.4ms]</div>
                      <div className="pl-8 text-accent-secondary">├── action-handler.ts : decryptActionBoundArgs(token) [1.8ms]</div>
                      <div className="pl-8 text-accent">└── flight-stream.ts : pipeChunkToSocketBuffer() [8.1ms]</div>
                    </div>
                  </div>
                )}

                {activeEngineTab === 1 && (
                  <div className="flex flex-col gap-2">
                    <div className="flex items-center justify-between border-b border-[#141414] pb-1 text-[10px]">
                      <span className="text-accent-secondary font-bold">02 · ISSUE INTELLIGENCE: 3-PHASE SEQUENCED BLUEPRINT</span>
                      <Link href="/issues" className="text-accent-secondary hover:underline">[ENTER WORKBENCH &rarr;]</Link>
                    </div>
                    <div className="flex flex-col gap-1.5 text-[10px]">
                      <div className="border border-accent-secondary-border bg-accent-secondary-soft p-1.5 flex items-center justify-between">
                        <span className="text-accent-secondary font-semibold">PHASE 01: Hook socket buffer drain monitor in flight-stream.ts</span>
                        <span className="text-[9px] text-[#4ade80]">TEST COVERAGE: 98%</span>
                      </div>
                      <div className="border border-[#1f1f1f] bg-[#080808] p-1.5 flex items-center justify-between text-[#cccccc]">
                        <span>PHASE 02: Add high-watermark suspension loop in socket-adapter.ts</span>
                        <span className="text-[9px] text-accent font-mono">EST. DIFF: +14 lines</span>
                      </div>
                      <div className="border border-[#1f1f1f] bg-[#080808] p-1.5 flex items-center justify-between text-[#cccccc]">
                        <span>PHASE 03: Run concurrent stream stress test suite</span>
                        <span className="text-[9px] text-[#4ade80]">NO REGRESSIONS</span>
                      </div>
                    </div>
                  </div>
                )}

                {activeEngineTab === 2 && (
                  <div className="flex flex-col gap-2">
                    <div className="flex items-center justify-between border-b border-[#141414] pb-1 text-[10px]">
                      <span className="text-accent font-bold">03 · PR REVIEWER: ARCHITECTURAL CONTRACT MATRIX</span>
                      <Link href="/pr" className="text-accent hover:underline">[ENTER WORKBENCH &rarr;]</Link>
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-3 gap-2 text-[10px]">
                      <div className="border border-[#14532d] bg-[#0c180e] p-2">
                        <span className="text-[#4ade80] font-bold block mb-0.5">BREAKING API CHANGES:</span>
                        <span className="text-[#fafafa]">0 Detected (Safe)</span>
                      </div>
                      <div className="border border-accent-secondary-border bg-accent-secondary-soft p-2">
                        <span className="text-accent-secondary font-bold block mb-0.5">BLAST RADIUS:</span>
                        <span className="text-[#fafafa]">2.4 / 10 Low Impact</span>
                      </div>
                      <div className="border border-[#1f1f1f] bg-[#080808] p-2">
                        <span className="text-[#cccccc] font-bold block mb-0.5">EVIDENCE CITATIONS:</span>
                        <span className="text-[#4ade80]">100% Grounded</span>
                      </div>
                    </div>
                  </div>
                )}

                {activeEngineTab === 3 && (
                  <div className="flex flex-col gap-2">
                    <div className="flex items-center justify-between border-b border-[#141414] pb-1 text-[10px]">
                      <span className="text-accent-secondary font-bold">04 · AI MENTOR: SOCRATIC ARCHITECTURAL ONBOARDING</span>
                      <Link href="/mentor" className="text-accent-secondary hover:underline">[ENTER WORKBENCH &rarr;]</Link>
                    </div>
                    <div className="font-mono text-[10px] bg-[#070707] p-2 border border-[#141414] flex flex-col gap-1 text-[#cccccc]">
                      <div className="text-accent">&gt; Contributor: &quot;Where should I start implementing Flight stream retry?&quot;</div>
                      <div className="text-accent-secondary">&gt; GitWise Mentor: &quot;Look at packages/next/src/server/flight/flight-stream.ts:L182. Note how BackpressureDrainController manages socket drain. Why would client-side chunk replay risk out-of-order execution in concurrent mode?&quot;</div>
                    </div>
                  </div>
                )}

                {activeEngineTab === 4 && (
                  <div className="flex flex-col gap-2">
                    <div className="flex items-center justify-between border-b border-[#141414] pb-1 text-[10px]">
                      <span className="text-accent font-bold">05 · GITSTAT: VERIFIED ENGINEERING FOOTPRINT</span>
                      <Link href="/gitstat" className="text-accent hover:underline">[ENTER WORKBENCH &rarr;]</Link>
                    </div>
                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 text-[10px]">
                      <div className="border border-[#1f1f1f] bg-[#080808] p-1.5">
                        <span className="text-[#737373] block text-[9px]">MERGED PRS</span>
                        <span className="text-accent font-bold text-sm">48 PRs</span>
                      </div>
                      <div className="border border-[#1f1f1f] bg-[#080808] p-1.5">
                        <span className="text-[#737373] block text-[9px]">REVIEW DEPTH</span>
                        <span className="text-accent-secondary font-bold text-sm">142 reviews</span>
                      </div>
                      <div className="border border-[#1f1f1f] bg-[#080808] p-1.5">
                        <span className="text-[#737373] block text-[9px]">GROUNDED CITATIONS</span>
                        <span className="text-[#4ade80] font-bold text-sm">100% Verified</span>
                      </div>
                      <div className="border border-[#1f1f1f] bg-[#080808] p-1.5">
                        <span className="text-[#737373] block text-[9px]">DIFF STREAM</span>
                        <span className="text-accent font-bold text-sm">Active Telemetry</span>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* The 5 Engine Full Grid Cards */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3.5">
              {/* Card 1: Repository Architecture & Feature Tracer */}
              <Link
                href="/repo"
                className="group border border-[#262626] bg-[#0a0a0a] hover:border-accent hover:bg-[#121212] p-4 flex flex-col justify-between transition-all hover:-translate-y-0.5 focus-visible:outline-accent"
              >
                <div>
                  <div className="flex items-center justify-between mb-3 text-[10px]">
                    <span className="text-accent font-bold tracking-wider">01 &middot; REPOSITORY ENGINE</span>
                    <ArrowRight className="w-3.5 h-3.5 text-[#737373] group-hover:text-accent transition-colors" />
                  </div>
                  <h3 className="text-sm font-bold text-[#fafafa] mb-1.5">Topological Feature Tracer</h3>
                  <p className="text-[11px] text-[#b3b3b3] leading-relaxed">
                    Map execution flows end-to-end from client dispatch to server handlers and database closures instead of manually grepping thousands of files.
                  </p>
                </div>
                <div className="border-t border-[#1a1a1a] pt-2.5 mt-4 text-[10px] text-accent-secondary flex justify-between items-center font-medium">
                  <span>Indexed AST &middot; Call Graph Analysis</span>
                  <span className="text-[#8a8a8a] group-hover:text-accent transition-colors text-[9px] font-bold">
                    [EXPLORE &rarr;]
                  </span>
                </div>
              </Link>

              {/* Card 2: Issue Intelligence & Blast Radius */}
              <Link
                href="/issues"
                className="group border border-[#262626] bg-[#0a0a0a] hover:border-accent-secondary hover:bg-[#121212] p-4 flex flex-col justify-between transition-all hover:-translate-y-0.5 focus-visible:outline-accent"
              >
                <div>
                  <div className="flex items-center justify-between mb-3 text-[10px]">
                    <span className="text-accent-secondary font-bold tracking-wider">02 &middot; ISSUE INTELLIGENCE</span>
                    <ArrowRight className="w-3.5 h-3.5 text-[#737373] group-hover:text-accent-secondary transition-colors" />
                  </div>
                  <h3 className="text-sm font-bold text-[#fafafa] mb-1.5">Grounded Blast Radius &amp; Blueprints</h3>
                  <p className="text-[11px] text-[#b3b3b3] leading-relaxed">
                    Synthesize exact impacted files, prerequisite knowledge, edge cases, and sequenced code diff blueprints without arbitrary difficulty badges.
                  </p>
                </div>
                <div className="border-t border-[#1a1a1a] pt-2.5 mt-4 text-[10px] text-[#4ade80] flex justify-between items-center font-medium">
                  <span>Automated Test Specs &middot; Sequenced Diff</span>
                  <span className="text-[#8a8a8a] group-hover:text-accent-secondary transition-colors text-[9px] font-bold">
                    [PLAN &rarr;]
                  </span>
                </div>
              </Link>

              {/* Card 3: Pull Request Semantic Reviewer */}
              <Link
                href="/pr"
                className="group border border-[#262626] bg-[#0a0a0a] hover:border-accent hover:bg-[#121212] p-4 flex flex-col justify-between transition-all hover:-translate-y-0.5 focus-visible:outline-accent"
              >
                <div>
                  <div className="flex items-center justify-between mb-3 text-[10px]">
                    <span className="text-accent font-bold tracking-wider">03 &middot; PR INTELLIGENCE</span>
                    <ArrowRight className="w-3.5 h-3.5 text-[#737373] group-hover:text-accent transition-colors" />
                  </div>
                  <h3 className="text-sm font-bold text-[#fafafa] mb-1.5">Architectural Impact Matrix</h3>
                  <p className="text-[11px] text-[#b3b3b3] leading-relaxed">
                    Understand whether a pull request breaks public API contracts, which downstream subsystems are coupled, and how runtime execution order mutates.
                  </p>
                </div>
                <div className="border-t border-[#1a1a1a] pt-2.5 mt-4 text-[10px] text-accent-secondary flex justify-between items-center font-medium">
                  <span>Contract Shift Detection &middot; AI Audit</span>
                  <span className="text-[#8a8a8a] group-hover:text-accent transition-colors text-[9px] font-bold">
                    [REVIEW &rarr;]
                  </span>
                </div>
              </Link>

              {/* Card 4: AI Open Source Mentor */}
              <Link
                href="/mentor"
                className="group border border-[#262626] bg-[#0a0a0a] hover:border-accent-secondary hover:bg-[#121212] p-4 flex flex-col justify-between transition-all hover:-translate-y-0.5 focus-visible:outline-accent"
              >
                <div>
                  <div className="flex items-center justify-between mb-3 text-[10px]">
                    <span className="text-accent-secondary font-bold tracking-wider">04 &middot; AI MENTORSHIP</span>
                    <ArrowRight className="w-3.5 h-3.5 text-[#737373] group-hover:text-accent-secondary transition-colors" />
                  </div>
                  <h3 className="text-sm font-bold text-[#fafafa] mb-1.5">Socratic Contributor Onboarding</h3>
                  <p className="text-[11px] text-[#b3b3b3] leading-relaxed">
                    High-readability architectural guidance paired with demonstrated skill matching, pointing you to beginner-friendly modules with 94%+ test coverage.
                  </p>
                </div>
                <div className="border-t border-[#1a1a1a] pt-2.5 mt-4 text-[10px] text-[#4ade80] flex justify-between items-center font-medium">
                  <span>Socratic Guidance &middot; Skill-Matched Issues</span>
                  <span className="text-[#8a8a8a] group-hover:text-accent-secondary transition-colors text-[9px] font-bold">
                    [CONSULT &rarr;]
                  </span>
                </div>
              </Link>

              {/* Card 5: GITSTAT Developer Footprint */}
              <Link
                href="/gitstat"
                className="group border border-[#262626] bg-[#0a0a0a] hover:border-accent hover:bg-[#121212] p-4 flex flex-col justify-between transition-all hover:-translate-y-0.5 focus-visible:outline-accent md:col-span-2 lg:col-span-2"
              >
                <div>
                  <div className="flex items-center justify-between mb-3 text-[10px]">
                    <span className="text-accent font-bold tracking-wider">05 &middot; GITSTAT PORTFOLIO</span>
                    <ArrowRight className="w-3.5 h-3.5 text-[#737373] group-hover:text-accent transition-colors" />
                  </div>
                  <h3 className="text-sm font-bold text-[#fafafa] mb-1.5">Factual Engineering Telemetry &amp; Portfolio</h3>
                  <p className="text-[11px] text-[#b3b3b3] leading-relaxed">
                    Turn raw GitHub activity into a verified engineering footprint: merged PRs, active review depth, factual issue tracking, and live diff stream telemetry. Zero gamification badges, pure evidence.
                  </p>
                </div>
                <div className="border-t border-[#1a1a1a] pt-2.5 mt-4 text-[10px] text-accent-secondary flex justify-between items-center font-medium">
                  <span>Auditable Git Footprint &middot; Real-time PR Diff Stream</span>
                  <span className="text-[#8a8a8a] group-hover:text-accent transition-colors text-[9px] font-bold">
                    [VIEW STATS &rarr;]
                  </span>
                </div>
              </Link>
            </div>
          </div>
        </section>

        {/* PIPELINE REGISTRATION GUIDE: STAGE 05 */}
        <div id="stage-05" className="border-b border-[#262626] bg-[#070707]/75 backdrop-blur-sm px-4 sm:px-6 py-1.5 text-[10px] text-[#b3b3b3] flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-accent font-bold tracking-wider">
              + --- PIPELINE // 05 · INSTRUMENTATION &amp; LIVE RADAR STREAM
            </span>
            <span className="hidden md:inline text-[#737373]">
              -------------------------------------------
            </span>
          </div>
          <div className="flex items-center gap-3">
            <span>MEASURABLE ARCHITECTURAL CAPABILITY</span>
            <span className="text-[#4ade80] font-semibold">&bull; FACTUAL STATUS</span>
          </div>
        </div>

        {/* 6. Authentic Grounded Telemetry & Live Diff Stream Radar */}
        <section className="border-b border-[#262626] bg-[#080808]/75 backdrop-blur-[2px] px-4 py-4 text-[11px]">
          <div className="max-w-5xl mx-auto flex flex-col gap-4">
            {/* Live Commit Radar Stream Ticker */}
            <div className="border border-[#1f1f1f] bg-[#040404] p-2.5 flex flex-wrap items-center justify-between gap-2 text-[10px]">
              <div className="flex items-center gap-2">
                <Radio className="w-3.5 h-3.5 text-accent animate-pulse" />
                <span className="text-accent font-bold">GITWISE GLOBAL REASONING RADAR:</span>
                <span className="text-[#fafafa] font-semibold">{currentLiveCommit.repo}@{currentLiveCommit.branch}</span>
                <span className="text-accent-secondary font-mono">({currentLiveCommit.sha})</span>
              </div>
              <div className="flex items-center gap-3">
                <span className="text-[#cccccc]">MODIFIED: <strong className="text-accent">{currentLiveCommit.file}</strong></span>
                <span className="text-[#737373]">|</span>
                <span className="text-accent-secondary font-bold border border-accent-secondary-border bg-accent-secondary-soft px-1.5 py-0.2">BLAST RADIUS: {currentLiveCommit.risk}</span>
                <span className="text-[#737373]">|</span>
                <span className="text-[#4ade80] font-semibold">&bull; 100% GROUNDED</span>
              </div>
            </div>

            {/* Core Capability Specifications Grid */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 text-[#a3a3a3]">
              <div className="border border-[#1f1f1f] p-2.5 bg-[#060606]">
                <span className="text-accent font-bold block text-[10px] mb-1">AST PARSER GRAMMAR</span>
                <span className="text-[#fafafa] text-xs font-semibold block">Rust Tree-Sitter &amp; SWC</span>
                <span className="text-[#737373] text-[9px] mt-0.5 block">Full TypeScript 5.5, JSX &amp; Rust ASTs</span>
              </div>
              <div className="border border-[#1f1f1f] p-2.5 bg-[#060606]">
                <span className="text-accent-secondary font-bold block text-[10px] mb-1">CITATION GUARANTEE</span>
                <span className="text-[#4ade80] text-xs font-semibold block">100% Verifiable File Anchors</span>
                <span className="text-[#737373] text-[9px] mt-0.5 block">Every deduction links to line citations</span>
              </div>
              <div className="border border-[#1f1f1f] p-2.5 bg-[#060606]">
                <span className="text-accent font-bold block text-[10px] mb-1">IN-BROWSER ENGINE</span>
                <span className="text-[#fafafa] text-xs font-semibold block">Zero-Clone WebAssembly</span>
                <span className="text-[#737373] text-[9px] mt-0.5 block">No local clone required for exploration</span>
              </div>
              <div className="border border-[#1f1f1f] p-2.5 bg-[#060606]">
                <span className="text-accent-secondary font-bold block text-[10px] mb-1">BENCHMARK CODEBASES</span>
                <span className="text-[#fafafa] text-xs font-semibold block">Next.js, React, K8s, Flask</span>
                <span className="text-[#737373] text-[9px] mt-0.5 block">Production-scale repository graphs</span>
              </div>
            </div>
          </div>
        </section>

        {/* PIPELINE REGISTRATION GUIDE: STAGE 06 */}
        <div id="stage-06" className="border-b border-[#262626] bg-[#070707]/75 backdrop-blur-sm px-4 sm:px-6 py-1.5 text-[10px] text-[#b3b3b3] flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-accent font-bold tracking-wider">
              + --- PIPELINE // 06 · DEVELOPER WORKFLOW ECOSYSTEM &amp; STANDALONE CLI
            </span>
            <span className="hidden md:inline text-[#737373]">
              -------------------------------------------
            </span>
          </div>
          <div className="flex items-center gap-3">
            <span>WEB WORKBENCH: AVAILABLE NOW</span>
            <span className="text-accent-secondary font-semibold">&bull; CLI IN ACTIVE DEV</span>
          </div>
        </div>

        {/* 7. Developer Quickstart & Standalone CLI Interactive Preview */}
        <section className="px-4 sm:px-6 lg:px-12 py-14 border-b border-[#262626] bg-[#050505]/50 backdrop-blur-[2px] text-center">
          <div className="max-w-3xl mx-auto flex flex-col items-center gap-4">
            <div className="inline-flex items-center gap-2 border border-accent-secondary-border bg-accent-secondary-soft px-3 py-1 text-[10px] text-accent-secondary font-bold tracking-wider">
              <span className="w-1.5 h-1.5 bg-accent-secondary animate-pulse"></span>
              <span>DEVELOPER WORKFLOW ECOSYSTEM</span>
            </div>

            <h2 className="text-xl sm:text-2xl font-bold text-[#fafafa] uppercase">
              Web Workbench Ready Now &middot; Terminal CLI In Development
            </h2>

            <p className="text-xs text-[#b8b8b8] leading-relaxed max-w-xl">
              Explore and reason through complex codebases right now in the browser compiler workbench. The standalone native CLI service (<code className="text-accent font-bold">@gitwise/cli</code>) is currently under active development for local terminal workflows.
            </p>

            {/* Interactive CLI Terminal Preview */}
            <div className="w-full border border-[#262626] bg-[#040404] p-3 text-left font-mono text-xs mt-2 shadow-2xl">
              {/* Terminal Tab Bar */}
              <div className="flex items-center justify-between border-b border-[#1f1f1f] pb-2 mb-2.5 text-[10px]">
                <div className="flex items-center gap-1.5">
                  {CLI_SNIPPETS.map((snip, i) => (
                    <button
                      key={snip.tab}
                      onClick={() => setActiveCliTab(i)}
                      className={`px-2 py-0.5 border text-[9px] transition-colors focus-visible:outline-accent ${
                        activeCliTab === i
                          ? "border-accent text-accent bg-accent-soft font-bold"
                          : "border-[#262626] text-[#8a8a8a] hover:text-[#fafafa]"
                      }`}
                    >
                      ${snip.tab}
                    </button>
                  ))}
                </div>
                <span className="text-accent-secondary border border-accent-secondary-border bg-accent-secondary-soft px-1.5 py-0.2 text-[9px] font-bold">
                  CLI: PREVIEW TRACK
                </span>
              </div>

              {/* Terminal Command Line */}
              <div className="flex flex-col gap-1.5 text-[11px] text-[#b3b3b3]">
                <div className="flex items-center gap-2">
                  <span className="text-accent select-none font-bold">$</span>
                  <span className="text-[#fafafa] font-bold">
                    {CLI_SNIPPETS[activeCliTab].command}
                  </span>
                  <span className="text-[9px] text-accent-secondary ml-auto border border-accent-secondary-border px-1">
                    # local binary
                  </span>
                </div>
                <div className="flex flex-col gap-1 pl-3 text-[10px] text-[#cccccc] font-mono border-l border-[#1f1f1f] mt-1">
                  {CLI_SNIPPETS[activeCliTab].output.map((line, idx) => (
                    <div
                      key={idx}
                      className={
                        line.startsWith("✓")
                          ? "text-[#4ade80]"
                          : line.startsWith("→")
                          ? "text-accent"
                          : line.includes("├──") || line.includes("└──")
                          ? "text-accent-secondary"
                          : "text-[#a3a3a3]"
                      }
                    >
                      {line}
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div className="flex flex-wrap items-center justify-center gap-3 mt-3">
              <Link
                href="/repo"
                className="px-4 py-2 border border-accent bg-accent-soft text-accent hover:bg-accent hover:text-[#050505] font-bold text-xs transition-all flex items-center gap-1.5 focus-visible:outline-accent active:translate-y-px"
              >
                <span>LAUNCH WEB WORKBENCH NOW</span>
                <ArrowRight className="w-3.5 h-3.5" />
              </Link>
              <Link
                href="/gitstat"
                className="px-4 py-2 border border-[#333333] hover:border-accent text-[#b3b3b3] hover:text-[#fafafa] text-xs transition-colors focus-visible:outline-accent"
              >
                VIEW GITSTAT PROFILE
              </Link>
            </div>
          </div>
        </section>

        {/* 8. Terminal Status Footer */}
        <footer className="px-4 sm:px-6 py-4 bg-[#050505]/85 backdrop-blur-sm text-[10px] text-[#a3a3a3] flex flex-wrap items-center justify-between gap-3 border-t border-[#1f1f1f]">
          <div className="flex items-center gap-3">
            <span className="text-accent font-bold">GITWISE PLATFORM v3.4</span>
            <span>THE OPEN-SOURCE INTELLIGENCE LAYER</span>
            <span className="hidden sm:inline text-[#737373]">&copy; 2026 GITWISE INC.</span>
          </div>
          <div className="flex items-center gap-4">
            <Link href="/repo" className="hover:text-accent transition-colors">
              ARCHITECTURE
            </Link>
            <Link href="/issues" className="hover:text-accent transition-colors">
              ISSUES
            </Link>
            <Link href="/pr" className="hover:text-accent transition-colors">
              PR REVIEW
            </Link>
            <Link href="/mentor" className="hover:text-accent transition-colors">
              MENTOR
            </Link>
            <Link href="/gitstat" className="hover:text-accent transition-colors">
              GITSTAT
            </Link>
            <span className="text-accent-secondary border border-accent-secondary-border px-1.5 py-0.5 bg-accent-secondary-soft font-semibold">
              CLI: IN DEV
            </span>
            <span className="text-[#4ade80] hidden md:inline font-semibold">
              SYSTEM: ONLINE
            </span>
          </div>
        </footer>
      </div>
    </div>
  );
}
