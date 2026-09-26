export interface PRFileDiff {
  path: string;
  additions: number;
  deletions: number;
  status: "modified" | "added" | "deleted";
  subsystem: string;
  behavioralSummary: string;
  diffSnippet: {
    target: string;
    before?: string;
    after: string;
  };
}

export interface ArchitecturalShift {
  publicApiContract: "unchanged" | "extended" | "breaking";
  publicApiExplanation: string;
  downstreamSubsystems: Array<{
    name: string;
    coupling: "tight" | "loose";
    impact: string;
  }>;
  executionFlowDelta: string;
  stateMutationRisk: "none" | "low" | "moderate" | "high";
  riskExplanation: string;
}

export interface ReviewFinding {
  ruleId: string;
  severity: "clean" | "advisory" | "warning";
  category: "convention" | "regression" | "performance" | "coverage";
  file: string;
  line: number;
  message: string;
  suggestion?: string;
}

export interface PullRequestModel {
  id: string;
  number: number;
  repo: string;
  title: string;
  author: string;
  status: "ready_for_review" | "in_review" | "approved" | "changes_requested";
  createdAt: string;
  stats: {
    additions: number;
    deletions: number;
    filesChanged: number;
    commitsCount: number;
  };
  summary: string;
  architecturalShift: ArchitecturalShift;
  files: PRFileDiff[];
  reviewFindings: ReviewFinding[];
  testVerification: {
    command: string;
    targetSuite: string;
    expectedResult: string;
    coverageDelta: string;
  };
}

export const PR_DATABASE: Record<string, PullRequestModel[]> = {
  "vercel/next.js": [
    {
      id: "next-pr-62145",
      number: 62145,
      repo: "vercel/next.js",
      title: "app-render: Reconcile Flight stream backpressure during server action revalidation",
      author: "shuding_dev",
      status: "ready_for_review",
      createdAt: "2026-09-20",
      stats: {
        additions: 182,
        deletions: 44,
        filesChanged: 4,
        commitsCount: 3,
      },
      summary: "When multiple server actions revalidate nested cache tags concurrently, unconsumed Flight stream chunks can exhaust server socket buffers. This PR introduces an asynchronous backpressure controller that throttles serialization chunks until client ACKs are received.",
      architecturalShift: {
        publicApiContract: "extended",
        publicApiExplanation: "Internal `renderToHTMLOrFlight` signature adds an optional `BackpressureDrainController` interface. External route handler exports remain 100% backward compatible.",
        downstreamSubsystems: [
          {
            name: "App Router Server Runtime",
            coupling: "tight",
            impact: "Action execution waits for stream drain before initiating nested tag invalidations.",
          },
          {
            name: "Streaming SSR Engine",
            coupling: "tight",
            impact: "React DOM server edge stream pipe delegates high-watermark events to backpressure manager.",
          },
          {
            name: "Turbopack Compiler",
            coupling: "loose",
            impact: "No compiler AST transformations modified; runtime JS chunk only.",
          },
        ],
        executionFlowDelta: "Stream Write → [NEW: Watermark Check & Drain Wait] → Flush Chunk → Invalidate Next-Cache Tag.",
        stateMutationRisk: "low",
        riskExplanation: "All buffer queues auto-release on HTTP socket abort signals, preventing memory leaks if clients disconnect.",
      },
      files: [
        {
          path: "packages/next/src/server/app-render/app-render.tsx",
          additions: 94,
          deletions: 22,
          status: "modified",
          subsystem: "App Router Server Runtime",
          behavioralSummary: "Injects BackpressureDrainController into Flight serialization loop and hooks client socket drain event.",
          diffSnippet: {
            target: "packages/next/src/server/app-render/app-render.tsx",
            before: `const stream = await flightPipeline.createStream(req);\nreturn stream;`,
            after: `const drainController = new BackpressureDrainController(req.socket);\nconst stream = await flightPipeline.createStream(req, { drainController });\nreturn stream;`,
          },
        },
        {
          path: "packages/next/src/server/app-render/action-handler.ts",
          additions: 48,
          deletions: 14,
          status: "modified",
          subsystem: "App Router Server Runtime",
          behavioralSummary: "Awaits drainController.drain() before resolving server action mutate response headers.",
          diffSnippet: {
            target: "packages/next/src/server/app-render/action-handler.ts",
            before: `await revalidatePath(targetPath);\nreturn new Response(JSON.stringify(actionResult));`,
            after: `await reqCtx.drainController?.drain();\nawait revalidatePath(targetPath);\nreturn new Response(JSON.stringify(actionResult));`,
          },
        },
        {
          path: "packages/next/src/server/lib/flight-stream.ts",
          additions: 28,
          deletions: 4,
          status: "added",
          subsystem: "Streaming SSR Engine",
          behavioralSummary: "Defines BackpressureDrainController class with watermark thresholds (default: 64KB).",
          diffSnippet: {
            target: "packages/next/src/server/lib/flight-stream.ts",
            after: `export class BackpressureDrainController {\n  private bufferSize: number = 0;\n  constructor(private socket: Socket, private highWatermark = 64 * 1024) {}\n  async drain(): Promise<void> { /* drain logic */ }\n}`,
          },
        },
        {
          path: "test/e2e/app-dir/actions-stream/actions-stream.test.ts",
          additions: 12,
          deletions: 4,
          status: "modified",
          subsystem: "Test Suite",
          behavioralSummary: "Validates high concurrency burst of 50 simultaneous server actions without buffer stalls.",
          diffSnippet: {
            target: "test/e2e/app-dir/actions-stream/actions-stream.test.ts",
            after: `it('handles 50 parallel actions with high-watermark backpressure drain', async () => {\n  const results = await runConcurrentActionSpam(50);\n  expect(results.failedCount).toBe(0);\n});`,
          },
        },
      ],
      reviewFindings: [
        {
          ruleId: "ARCH-01",
          severity: "clean",
          category: "convention",
          file: "packages/next/src/server/app-render/app-render.tsx",
          line: 42,
          message: "Follows Next.js RFC-88 for edge streaming boundaries and server action serialization.",
        },
        {
          ruleId: "REG-04",
          severity: "advisory",
          category: "regression",
          file: "packages/next/src/server/lib/flight-stream.ts",
          line: 18,
          message: "Socket abort listener must be deregistered on stream close to avoid Node EventEmitter memory leak.",
          suggestion: "Add `this.socket.once('close', () => this.cleanup())` inside constructor.",
        },
        {
          ruleId: "COV-02",
          severity: "clean",
          category: "coverage",
          file: "test/e2e/app-dir/actions-stream/actions-stream.test.ts",
          line: 1,
          message: "PR adds dedicated E2E concurrency regression tests (+12 lines).",
        },
      ],
      testVerification: {
        command: "pnpm test-e2e test/e2e/app-dir/actions-stream",
        targetSuite: "actions-stream.test.ts",
        expectedResult: "PASS (3 passed, 0 failed, 100% assertions)",
        coverageDelta: "+2.4% on app-render/action-handler.ts",
      },
    },
    {
      id: "next-pr-61980",
      number: 61980,
      repo: "vercel/next.js",
      title: "turbopack: Add persistent incremental file hash cache for monorepo roots",
      author: "sokra_watch",
      status: "in_review",
      createdAt: "2026-09-17",
      stats: {
        additions: 215,
        deletions: 68,
        filesChanged: 3,
        commitsCount: 4,
      },
      summary: "Monorepos with 10+ sub-packages experienced 1.8s startup times due to full filesystem hashing. This PR adds a persistent SHA-256 mtime cache.",
      architecturalShift: {
        publicApiContract: "unchanged",
        publicApiExplanation: "Purely internal performance optimization within Turbopack Rust native engine.",
        downstreamSubsystems: [
          {
            name: "Turbopack Compiler",
            coupling: "tight",
            impact: "Reduces initial dev-server warm up time by 68%.",
          },
        ],
        executionFlowDelta: "File Watcher → [Hash Cache Check] → Compile AST.",
        stateMutationRisk: "none",
        riskExplanation: "Mtime mismatch automatically triggers full cache invalidation.",
      },
      files: [
        {
          path: "crates/turbopack/src/lib.rs",
          additions: 140,
          deletions: 42,
          status: "modified",
          subsystem: "Turbopack Compiler",
          behavioralSummary: "Implements mtime cache lookup table in Rust.",
          diffSnippet: {
            target: "crates/turbopack/src/lib.rs",
            after: `pub struct MtimeHashCache {\n    cache: DashMap<PathBuf, (u64, Hash)>,\n}`,
          },
        },
      ],
      reviewFindings: [
        {
          ruleId: "PERF-01",
          severity: "clean",
          category: "performance",
          file: "crates/turbopack/src/lib.rs",
          line: 85,
          message: "Benchmarked 68% latency reduction on 50k file worktrees.",
        },
      ],
      testVerification: {
        command: "cargo test --package turbopack --test hash_cache",
        targetSuite: "hash_cache.rs",
        expectedResult: "test result: ok. 12 passed; 0 failed",
        coverageDelta: "+4.1% on crates/turbopack",
      },
    },
  ],
  "facebook/react": [
    {
      id: "react-pr-28904",
      number: 28904,
      repo: "facebook/react",
      title: "react-reconciler: Batch action transitions across concurrent suspense boundaries",
      author: "sophiebits_cont",
      status: "ready_for_review",
      createdAt: "2026-09-14",
      stats: {
        additions: 142,
        deletions: 38,
        filesChanged: 3,
        commitsCount: 2,
      },
      summary: "In React 19, multiple concurrent transitions triggered across sibling Suspense boundaries could execute in separate commit cycles, causing tearing. This PR unifies their lane masks.",
      architecturalShift: {
        publicApiContract: "unchanged",
        publicApiExplanation: "Internal Fiber scheduler change. Public useTransition and startTransition APIs remain identical.",
        downstreamSubsystems: [
          {
            name: "Fiber Reconciler",
            coupling: "tight",
            impact: "Entangles sibling TransitionLanes during active work loop execution.",
          },
          {
            name: "React Scheduler",
            coupling: "loose",
            impact: "Ensures post-task callback processes all sibling roots in a single batch.",
          },
        ],
        executionFlowDelta: "Dispatch Action → [Entangle Sibling Lanes] → Schedule Work → Single Commit Phase.",
        stateMutationRisk: "moderate",
        riskExplanation: "Requires rigorous verification against starvation of high-priority SyncLanes.",
      },
      files: [
        {
          path: "packages/react-reconciler/src/ReactFiberWorkLoop.js",
          additions: 88,
          deletions: 24,
          status: "modified",
          subsystem: "Fiber Reconciler",
          behavioralSummary: "Combines sibling transition lane masks during work loop traversal.",
          diffSnippet: {
            target: "packages/react-reconciler/src/ReactFiberWorkLoop.js",
            after: `const entangledLanes = entangleSiblingTransitionLanes(root, updateLane);`,
          },
        },
      ],
      reviewFindings: [
        {
          ruleId: "FIBER-03",
          severity: "clean",
          category: "convention",
          file: "packages/react-reconciler/src/ReactFiberWorkLoop.js",
          line: 110,
          message: "Preserves React lane priority invariants: SyncLane strictly outranks TransitionLane.",
        },
      ],
      testVerification: {
        command: "yarn test packages/react-reconciler/src/__tests__/ReactSuspense-test.js",
        targetSuite: "ReactSuspense-test.js",
        expectedResult: "PASS (48 passed, 0 failed)",
        coverageDelta: "+1.8% on ReactFiberWorkLoop.js",
      },
    },
  ],
  "kubernetes/kubernetes": [
    {
      id: "k8s-pr-121080",
      number: 121080,
      repo: "kubernetes/kubernetes",
      title: "scheduler: Deduplicate preemption candidate queue evaluations",
      author: "liggitt_core",
      status: "approved",
      createdAt: "2026-09-10",
      stats: {
        additions: 96,
        deletions: 51,
        filesChanged: 3,
        commitsCount: 2,
      },
      summary: "When multiple unschedulable pods trigger preemption simultaneously on large clusters, the scheduler evaluated identical victim nodes repeatedly. This PR adds a candidate memoization map.",
      architecturalShift: {
        publicApiContract: "unchanged",
        publicApiExplanation: "Purely internal optimization in default scheduler preemption plugin.",
        downstreamSubsystems: [
          {
            name: "kube-scheduler",
            coupling: "tight",
            impact: "Reduces preemption cycle latency by 45% on 5,000 node clusters.",
          },
        ],
        executionFlowDelta: "PostFilter Preempt → [Memoized Candidate Check] → Select Victim Nodes.",
        stateMutationRisk: "low",
        riskExplanation: "Memoization map is scoped strictly to a single scheduling cycle.",
      },
      files: [
        {
          path: "pkg/scheduler/framework/plugins/defaultpreemption/default_preemption.go",
          additions: 68,
          deletions: 34,
          status: "modified",
          subsystem: "kube-scheduler",
          behavioralSummary: "Adds node evaluation memoizer during preemption candidate search.",
          diffSnippet: {
            target: "pkg/scheduler/framework/plugins/defaultpreemption/default_preemption.go",
            after: `if memoized, found := cycleCache.Get(node.Name); found {\n    return memoized, nil\n}`,
          },
        },
      ],
      reviewFindings: [
        {
          ruleId: "SCHED-01",
          severity: "clean",
          category: "performance",
          file: "pkg/scheduler/framework/plugins/defaultpreemption/default_preemption.go",
          line: 45,
          message: "Benchmark validates 45% speedup in PreFilter cycle under 1,000 pending pods.",
        },
      ],
      testVerification: {
        command: "go test -v ./pkg/scheduler/framework/plugins/defaultpreemption/...",
        targetSuite: "default_preemption_test.go",
        expectedResult: "ok  k8s.io/kubernetes/pkg/scheduler/framework/plugins/defaultpreemption 2.1s",
        coverageDelta: "+3.2% on default_preemption.go",
      },
    },
  ],
  "pallets/flask": [
    {
      id: "flask-pr-5280",
      number: 5280,
      repo: "pallets/flask",
      title: "signals: Thread context isolation into async route teardown callbacks",
      author: "untitaker_flk",
      status: "in_review",
      createdAt: "2026-09-05",
      stats: {
        additions: 38,
        deletions: 12,
        filesChanged: 2,
        commitsCount: 1,
      },
      summary: "In async route handlers, teardown_request signal listeners were executing outside the task context boundary. This PR passes contextvar snapshots into teardown dispatchers.",
      architecturalShift: {
        publicApiContract: "unchanged",
        publicApiExplanation: "Signals API signature remains identical.",
        downstreamSubsystems: [
          {
            name: "WSGI Request Pipeline",
            coupling: "tight",
            impact: "Ensures teardown signal callbacks have access to current_app and request context.",
          },
        ],
        executionFlowDelta: "Request Finalize → [Copy Context Snapshot] → Dispatch Teardown Signals.",
        stateMutationRisk: "none",
        riskExplanation: "Context snapshot is read-only.",
      },
      files: [
        {
          path: "src/flask/signals.py",
          additions: 24,
          deletions: 8,
          status: "modified",
          subsystem: "WSGI Request Pipeline",
          behavioralSummary: "Contextvar snapshot wrapping around signal.send().",
          diffSnippet: {
            target: "src/flask/signals.py",
            after: `ctx = copy_current_request_context(f)\nreturn ctx(*args, **kwargs)`,
          },
        },
      ],
      reviewFindings: [
        {
          ruleId: "CTX-01",
          severity: "clean",
          category: "convention",
          file: "src/flask/signals.py",
          line: 30,
          message: "Correctly handles Python 3.10+ contextvars copy semantics.",
        },
      ],
      testVerification: {
        command: "pytest tests/test_signals.py -k test_async_teardown",
        targetSuite: "test_signals.py",
        expectedResult: "2 passed in 0.12s",
        coverageDelta: "+5.0% on signals.py",
      },
    },
  ],
  "meshery/meshery": [
    {
      id: "meshery-pr-8540",
      number: 8540,
      repo: "meshery/meshery",
      title: "meshsync: Dynamic informer resync interval scaling based on cluster event density",
      author: "leecalcote_msh",
      status: "ready_for_review",
      createdAt: "2026-09-11",
      stats: {
        additions: 115,
        deletions: 29,
        filesChanged: 4,
        commitsCount: 3,
      },
      summary: "Under high cluster churn (e.g. Istio canary rollouts), MeshSync informers overloaded the Kube API. This PR introduces adaptive resync intervals (10s to 300s).",
      architecturalShift: {
        publicApiContract: "extended",
        publicApiExplanation: "MeshSyncConfig adds `adaptive_resync_enabled` boolean flag.",
        downstreamSubsystems: [
          {
            name: "MeshSync Discovery Engine",
            coupling: "tight",
            impact: "Dynamically modulates shared informer factory resync periods.",
          },
        ],
        executionFlowDelta: "Event Burst Detected → [Scale Resync Interval] → Reconnect Informers.",
        stateMutationRisk: "low",
        riskExplanation: "Falls back to 30s default if metrics collection fails.",
      },
      files: [
        {
          path: "server/meshmodel/meshsync/informer.go",
          additions: 76,
          deletions: 18,
          status: "modified",
          subsystem: "MeshSync Discovery Engine",
          behavioralSummary: "Adaptive resync timer implementation.",
          diffSnippet: {
            target: "server/meshmodel/meshsync/informer.go",
            after: `resyncPeriod := calculateAdaptiveResync(eventRatePerSec)`,
          },
        },
      ],
      reviewFindings: [
        {
          ruleId: "SYNC-02",
          severity: "clean",
          category: "convention",
          file: "server/meshmodel/meshsync/informer.go",
          line: 52,
          message: "Implements thread-safe atomic float operations for rate measurement.",
        },
      ],
      testVerification: {
        command: "go test ./server/meshmodel/meshsync/...",
        targetSuite: "informer_test.go",
        expectedResult: "ok  github.com/meshery/meshery/server/meshmodel/meshsync 1.1s",
        coverageDelta: "+3.8% on informer.go",
      },
    },
  ],
};
