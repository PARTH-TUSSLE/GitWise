export interface SubsystemNode {
  id: string;
  name: string;
  fileCount: number;
  entryPoint: string;
  description: string;
  language: string;
  connections: string[];
  beginnerFriendly: boolean;
}

export interface FeatureTraceStep {
  step: number;
  title: string;
  subsystem: string;
  file: string;
  line: number;
  description: string;
  codeSnippet: string;
}

export interface FeatureTrace {
  id: string;
  name: string;
  description: string;
  steps: FeatureTraceStep[];
}

export interface RepoTreeItem {
  name: string;
  path: string;
  type: "file" | "directory";
  size?: string;
  owner?: string;
  children?: RepoTreeItem[];
}

export interface RepoModel {
  id: string;
  name: string;
  owner: string;
  branch: string;
  stars: number;
  forks: number;
  indexedFiles: number;
  primaryLanguage: string;
  description: string;
  subsystems: SubsystemNode[];
  featureTraces: FeatureTrace[];
  contributorGuide: {
    stepsToStart: string[];
    prerequisites: string[];
    beginnerFiles: Array<{ path: string; reason: string }>;
  };
  fileTree: RepoTreeItem[];
  chatHistory: Array<{
    sender: "user" | "gitwise";
    message: string;
    citations?: Array<{ file: string; line?: number; snippet?: string }>;
  }>;
}

export const REPOSITORIES: Record<string, RepoModel> = {
  "vercel/next.js": {
    id: "nextjs",
    name: "vercel/next.js",
    owner: "vercel",
    branch: "canary",
    stars: 124000,
    forks: 26100,
    indexedFiles: 2840,
    primaryLanguage: "TypeScript",
    description: "The React Framework for the Web: App Router, Turbopack, and hybrid SSR.",
    subsystems: [
      {
        id: "app-router",
        name: "App Router Server Runtime",
        fileCount: 312,
        entryPoint: "packages/next/src/server/app-render/app-render.tsx",
        description: "Renders server components, manages action dispatch, and coordinates React Flight stream.",
        language: "TypeScript",
        connections: ["turbopack", "ssr-engine"],
        beginnerFriendly: false,
      },
      {
        id: "turbopack",
        name: "Turbopack Compiler",
        fileCount: 489,
        entryPoint: "crates/turbopack/src/lib.rs",
        description: "Incremental Rust-based bundler and module graph evaluator for ultra-fast HMR.",
        language: "Rust",
        connections: ["ssr-engine"],
        beginnerFriendly: false,
      },
      {
        id: "ssr-engine",
        name: "Streaming SSR Engine",
        fileCount: 198,
        entryPoint: "packages/next/src/server/web/spec-extension/adapters/next-request.ts",
        description: "Web standard request/response adapters and streaming HTML transformer.",
        language: "TypeScript",
        connections: ["client-runtime"],
        beginnerFriendly: true,
      },
      {
        id: "client-runtime",
        name: "Client Islands & Hydration",
        fileCount: 145,
        entryPoint: "packages/next/src/client/app-index.tsx",
        description: "Browser entrypoint reconciling Flight payload and hydrating interactive components.",
        language: "TypeScript",
        connections: [],
        beginnerFriendly: true,
      },
    ],
    featureTraces: [
      {
        id: "server-actions",
        name: "Server Actions Execution",
        description: "End-to-end trace from client form submission to server mutation and revalidation.",
        steps: [
          {
            step: 1,
            title: "Client Action Dispatch",
            subsystem: "Client Islands & Hydration",
            file: "packages/next/src/client/components/react-dev-overlay/client-action.ts",
            line: 42,
            description: "Form submit or startTransition invokes generated server action bound stub.",
            codeSnippet: `export function dispatchServerAction(actionId: string, boundArgs: any[]) {\n  return fetch(window.location.href, {\n    method: 'POST',\n    headers: { 'Next-Action': actionId },\n    body: serializeActionArgs(boundArgs)\n  });\n}`,
          },
          {
            step: 2,
            title: "Header Injection & Middleware Router",
            subsystem: "Streaming SSR Engine",
            file: "packages/next/src/server/app-render/action-handler.ts",
            line: 88,
            description: "Next.js server validates action header, decrypts closure args, and executes boundary checks.",
            codeSnippet: `const actionId = req.headers.get('Next-Action');\nif (actionId) {\n  const decryptedClosure = await decryptActionBoundArgs(actionId, req);\n  return executeServerAction(decryptedClosure, req);\n}`,
          },
          {
            step: 3,
            title: "Mutation & Revalidation Execution",
            subsystem: "App Router Server Runtime",
            file: "packages/next/src/server/app-render/action-encryption-utils.ts",
            line: 142,
            description: "Calls user-defined async function, handles throw/redirect, and streams back revalidated Flight tree.",
            codeSnippet: `const result = await actionFunction.apply(null, args);\nconst revalidatedTree = await renderFlightTree(req.url);\nreturn createFlightResponse(result, revalidatedTree);`,
          },
        ],
      },
      {
        id: "streaming-ssr",
        name: "Streaming SSR Hydration",
        description: "Tracing React Server Component generation and progressive browser rendering.",
        steps: [
          {
            step: 1,
            title: "Flight Payload Stream Initialization",
            subsystem: "App Router Server Runtime",
            file: "packages/next/src/server/app-render/app-render.tsx",
            line: 215,
            description: "Initializes React Server Components streaming renderer and establishes boundary slots.",
            codeSnippet: `const stream = renderToReadableStream(\n  <AppRouterComponent url={req.url} />,\n  { onError: handleRenderError }\n);`,
          },
          {
            step: 2,
            title: "Progressive Browser Island Hydration",
            subsystem: "Client Islands & Hydration",
            file: "packages/next/src/client/app-index.tsx",
            line: 64,
            description: "Reads incoming HTML stream, mounts client component islands progressively.",
            codeSnippet: `const root = createRoot(document.getElementById('__next')!);\nroot.render(<RootFlightStreamConsumer stream={window.__FLIGHT_STREAM__} />);`,
          },
        ],
      },
    ],
    contributorGuide: {
      stepsToStart: [
        "Clone repo and run 'pnpm install' at root",
        "Build Turbopack Rust native addons: 'pnpm build-native'",
        "Run test suite for a specific package: 'pnpm --filter next test'",
        "Inspect packages/next/src/client for beginner-friendly contributions",
      ],
      prerequisites: [
        "Understanding of React Server Components (RSC)",
        "Node.js 20+ and Rust toolchain (cargo, rustc)",
        "TypeScript strict typing conventions",
      ],
      beginnerFiles: [
        { path: "packages/next/src/client/link.tsx", reason: "Well-isolated client navigation component with high test coverage" },
        { path: "packages/next/src/shared/lib/constants.ts", reason: "Header definitions, environment flags, and global config" },
        { path: "packages/next/src/server/lib/format-server-error.ts", reason: "Error message formatting and CLI diagnostic reporting" },
      ],
    },
    fileTree: [
      {
        name: "packages",
        path: "packages",
        type: "directory",
        children: [
          {
            name: "next",
            path: "packages/next",
            type: "directory",
            children: [
              { name: "src/server/app-render/app-render.tsx", path: "packages/next/src/server/app-render/app-render.tsx", type: "file", size: "38KB", owner: "gaearon" },
              { name: "src/server/app-render/action-handler.ts", path: "packages/next/src/server/app-render/action-handler.ts", type: "file", size: "14KB", owner: "alexR_dev" },
              { name: "src/client/app-index.tsx", path: "packages/next/src/client/app-index.tsx", type: "file", size: "9KB", owner: "sindresorhus" },
              { name: "package.json", path: "packages/next/package.json", type: "file", size: "4KB", owner: "vercel" },
            ],
          },
        ],
      },
      {
        name: "crates",
        path: "crates",
        type: "directory",
        children: [
          { name: "turbopack/src/lib.rs", path: "crates/turbopack/src/lib.rs", type: "file", size: "84KB", owner: "alexR_dev" },
          { name: "Cargo.toml", path: "crates/Cargo.toml", type: "file", size: "2KB", owner: "vercel" },
        ],
      },
      { name: "README.md", path: "README.md", type: "file", size: "12KB", owner: "vercel" },
    ],
    chatHistory: [
      {
        sender: "user",
        message: "How does the App Router connect to the Turbopack engine during local development?",
      },
      {
        sender: "gitwise",
        message: "Next.js invokes Turbopack via native N-API bindings compiled from `crates/turbopack/src/lib.rs`. During local development (`next dev --turbo`), `app-render.tsx` connects to Turbopack's HMR WebSocket handler to invalidate and rebuild individual route segments on file edits.",
        citations: [
          { file: "packages/next/src/server/app-render/app-render.tsx", line: 45, snippet: "import { getTurbopackDevServer } from '../dev/turbopack-bindings';" },
          { file: "crates/turbopack/src/lib.rs", line: 112, snippet: "pub fn watch_entrypoint(path: &Path) -> Result<HmrStream> { ... }" },
        ],
      },
    ],
  },
  "facebook/react": {
    id: "react",
    name: "facebook/react",
    owner: "facebook",
    branch: "main",
    stars: 228000,
    forks: 46200,
    indexedFiles: 1420,
    primaryLanguage: "JavaScript",
    description: "The library for web and native user interfaces.",
    subsystems: [
      {
        id: "reconciler",
        name: "Fiber Reconciler",
        fileCount: 184,
        entryPoint: "packages/react-reconciler/src/ReactFiberWorkLoop.js",
        description: "Concurrent scheduler, priority lane assignment, and double-buffering DOM diffing.",
        language: "JavaScript",
        connections: ["hooks-runtime", "dom-renderer"],
        beginnerFriendly: false,
      },
      {
        id: "hooks-runtime",
        name: "Hooks State Runtime",
        fileCount: 62,
        entryPoint: "packages/react-reconciler/src/ReactFiberHooks.js",
        description: "Manages state hook linked lists, effect dispatch queues, and memoized values.",
        language: "JavaScript",
        connections: ["reconciler"],
        beginnerFriendly: true,
      },
      {
        id: "dom-renderer",
        name: "ReactDOM Host Environment",
        fileCount: 94,
        entryPoint: "packages/react-dom/src/client/ReactDOM.js",
        description: "Event plugin delegator and host mutations on actual DOM elements.",
        language: "JavaScript",
        connections: [],
        beginnerFriendly: true,
      },
    ],
    featureTraces: [
      {
        id: "state-update",
        name: "State Update & Fiber Reconciliation",
        description: "Trace from setState() call to priority lane scheduling and DOM commit.",
        steps: [
          {
            step: 1,
            title: "dispatchAction Invocation",
            subsystem: "Hooks State Runtime",
            file: "packages/react-reconciler/src/ReactFiberHooks.js",
            line: 2314,
            description: "User calls setState(x); React enqueues update in fiber hook's queue.",
            codeSnippet: `function dispatchSetState(fiber, queue, action) {\n  const lane = requestUpdateLane(fiber);\n  const update = { lane, action, next: null };\n  enqueueUpdate(fiber, queue, update, lane);\n  scheduleUpdateOnFiber(fiber, lane);\n}`,
          },
          {
            step: 2,
            title: "Work Loop Scheduling",
            subsystem: "Fiber Reconciler",
            file: "packages/react-reconciler/src/ReactFiberWorkLoop.js",
            line: 840,
            description: "Scheduler prioritizes lane, iterates fiber tree, computes next element diff.",
            codeSnippet: `function workLoopConcurrent() {\n  while (workInProgress !== null && !shouldYield()) {\n    performUnitOfWork(workInProgress);\n  }\n}`,
          },
          {
            step: 3,
            title: "Commit Phase DOM Mutation",
            subsystem: "ReactDOM Host Environment",
            file: "packages/react-dom/src/client/ReactDOMComponent.js",
            line: 412,
            description: "Applies calculated style and attribute deltas to actual browser DOM nodes.",
            codeSnippet: `updateDOMProperties(domElement, updatePayload, wasCustomComponentTag);`,
          },
        ],
      },
    ],
    contributorGuide: {
      stepsToStart: [
        "Clone repo and run 'yarn' at root",
        "Run compiler build: 'yarn build'",
        "Run Jest tests: 'yarn test'",
        "Read 'fixtures/packaging' to understand build artifacts",
      ],
      prerequisites: [
        "Solid grasp of JavaScript prototype chain and closures",
        "Understanding of cooperative concurrency and scheduling",
      ],
      beginnerFiles: [
        { path: "packages/shared/ReactSymbols.js", reason: "Symbol constants for React element types ($$typeof)" },
        { path: "packages/react/src/ReactSharedInternals.js", reason: "Bridge between React core package and reconcilers" },
      ],
    },
    fileTree: [
      {
        name: "packages",
        path: "packages",
        type: "directory",
        children: [
          {
            name: "react-reconciler",
            path: "packages/react-reconciler",
            type: "directory",
            children: [
              { name: "src/ReactFiberWorkLoop.js", path: "packages/react-reconciler/src/ReactFiberWorkLoop.js", type: "file", size: "72KB", owner: "gaearon" },
              { name: "src/ReactFiberHooks.js", path: "packages/react-reconciler/src/ReactFiberHooks.js", type: "file", size: "89KB", owner: "gaearon" },
            ],
          },
          {
            name: "react-dom",
            path: "packages/react-dom",
            type: "directory",
            children: [
              { name: "src/client/ReactDOM.js", path: "packages/react-dom/src/client/ReactDOM.js", type: "file", size: "18KB", owner: "facebook" },
            ],
          },
        ],
      },
      { name: "package.json", path: "package.json", type: "file", size: "5KB", owner: "facebook" },
    ],
    chatHistory: [
      {
        sender: "user",
        message: "Where does React keep track of component hook states between renders?",
      },
      {
        sender: "gitwise",
        message: "React tracks hook state as a singly-linked list on `fiber.memoizedState`. During rendering, `ReactFiberHooks.js` maintains `workInProgressHook` pointing to the currently active hook item in the list.",
        citations: [
          { file: "packages/react-reconciler/src/ReactFiberHooks.js", line: 412, snippet: "let currentlyRenderingFiber: Fiber = null;\nlet workInProgressHook: Hook | null = null;" },
        ],
      },
    ],
  },
  "kubernetes/kubernetes": {
    id: "kubernetes",
    name: "kubernetes/kubernetes",
    owner: "kubernetes",
    branch: "master",
    stars: 109000,
    forks: 39100,
    indexedFiles: 8400,
    primaryLanguage: "Go",
    description: "Production-Grade Container Scheduling and Management.",
    subsystems: [
      {
        id: "apiserver",
        name: "Kube API Server",
        fileCount: 840,
        entryPoint: "cmd/kube-apiserver/app/server.go",
        description: "RESTful HTTP server validating and configuring data for pods, services, and replication controllers.",
        language: "Go",
        connections: ["scheduler", "controller-manager"],
        beginnerFriendly: false,
      },
      {
        id: "scheduler",
        name: "Kube Scheduler",
        fileCount: 240,
        entryPoint: "cmd/kube-scheduler/app/server.go",
        description: "Selects optimal worker nodes for newly created pods based on resource constraints and affinity rules.",
        language: "Go",
        connections: ["kubelet"],
        beginnerFriendly: true,
      },
      {
        id: "kubelet",
        name: "Kubelet Node Daemon",
        fileCount: 680,
        entryPoint: "cmd/kubelet/app/server.go",
        description: "Agent running on each node ensuring containers described in PodSpecs are running and healthy.",
        language: "Go",
        connections: [],
        beginnerFriendly: false,
      },
    ],
    featureTraces: [
      {
        id: "pod-lifecycle",
        name: "Pod Scheduling & Reconciliation Loop",
        description: "Trace from kubectl apply -f pod.yaml to Kubelet container creation via CRI.",
        steps: [
          {
            step: 1,
            title: "API Admission & etcd Storage",
            subsystem: "Kube API Server",
            file: "pkg/api/pod/strategy.go",
            line: 120,
            description: "API server validates pod spec, assigns UID, and writes to etcd storage.",
            codeSnippet: `func (podStrategy) Canonicalize(obj runtime.Object) {\n  pod := obj.(*api.Pod)\n  // normalize annotations and spec defaults\n}`,
          },
          {
            step: 2,
            title: "Node Filtering & Scoring",
            subsystem: "Kube Scheduler",
            file: "pkg/scheduler/schedule_one.go",
            line: 74,
            description: "Scheduler watches unassigned pods, filters candidate nodes, scores affinity.",
            codeSnippet: `scheduleResult, err := sched.SchedulePod(ctx, fwk, state, pod)\nerr = sched.bind(ctx, fwk, pod, scheduleResult.SuggestedHost, state)`,
          },
          {
            step: 3,
            title: "Kubelet CRI Container Run",
            subsystem: "Kubelet Node Daemon",
            file: "pkg/kubelet/kuberuntime/kuberuntime_container.go",
            line: 184,
            description: "Kubelet on assigned node detects binding, pulls container image, calls CRI runtime.",
            codeSnippet: `msg, err := m.runtimeService.CreateContainer(podSandboxID, containerConfig, sandboxConfig)`,
          },
        ],
      },
    ],
    contributorGuide: {
      stepsToStart: [
        "Fork and clone repository (requires ~10GB disk space)",
        "Set up Go 1.22+ and run 'make'",
        "Run unit tests: 'make test'",
        "Run local cluster: 'hack/local-up-cluster.sh'",
      ],
      prerequisites: [
        "Go interfaces and concurrency patterns (goroutines, channels)",
        "Understanding of Kubernetes primitives (Pods, Deployments, CRDs)",
      ],
      beginnerFiles: [
        { path: "pkg/apis/core/types.go", reason: "Core struct definitions for PodSpec, Container, and Node" },
        { path: "pkg/kubelet/status/status_manager.go", reason: "Clean worker loop syncing pod health back to API server" },
      ],
    },
    fileTree: [
      {
        name: "cmd",
        path: "cmd",
        type: "directory",
        children: [
          { name: "kube-apiserver/app/server.go", path: "cmd/kube-apiserver/app/server.go", type: "file", size: "48KB", owner: "kubernetes" },
          { name: "kube-scheduler/app/server.go", path: "cmd/kube-scheduler/app/server.go", type: "file", size: "32KB", owner: "kubernetes" },
        ],
      },
      {
        name: "pkg",
        path: "pkg",
        type: "directory",
        children: [
          { name: "scheduler/schedule_one.go", path: "pkg/scheduler/schedule_one.go", type: "file", size: "54KB", owner: "kubernetes" },
          { name: "kubelet/kuberuntime/kuberuntime_container.go", path: "pkg/kubelet/kuberuntime/kuberuntime_container.go", type: "file", size: "42KB", owner: "kubernetes" },
        ],
      },
    ],
    chatHistory: [
      {
        sender: "user",
        message: "Where does the Kubelet communicate with container runtimes like containerd?",
      },
      {
        sender: "gitwise",
        message: "The Kubelet interacts with container runtimes exclusively via the Container Runtime Interface (CRI) gRPC protocol in `pkg/kubelet/cri/remote/remote_runtime.go`.",
        citations: [
          { file: "pkg/kubelet/cri/remote/remote_runtime.go", line: 55, snippet: "type RemoteRuntimeService struct { runtimeClient runtimeapi.RuntimeServiceClient }" },
        ],
      },
    ],
  },
  "pallets/flask": {
    id: "flask",
    name: "pallets/flask",
    owner: "pallets",
    branch: "main",
    stars: 67400,
    forks: 16100,
    indexedFiles: 280,
    primaryLanguage: "Python",
    description: "The Python micro framework for building web applications.",
    subsystems: [
      {
        id: "wsgi-app",
        name: "WSGI Core Dispatcher",
        fileCount: 42,
        entryPoint: "src/flask/app.py",
        description: "WSGI application callable managing request lifecycles and response encoding.",
        language: "Python",
        connections: ["routing-engine", "ctx-manager"],
        beginnerFriendly: true,
      },
      {
        id: "routing-engine",
        name: "Werkzeug Route RuleMap",
        fileCount: 28,
        entryPoint: "src/flask/sansio/scaffold.py",
        description: "Registers view decorators, extracts URL parameters, and matches HTTP verbs.",
        language: "Python",
        connections: [],
        beginnerFriendly: true,
      },
      {
        id: "ctx-manager",
        name: "Request / App Context Manager",
        fileCount: 34,
        entryPoint: "src/flask/ctx.py",
        description: "Thread-local and greenlet context stacks for `g`, `request`, and `current_app`.",
        language: "Python",
        connections: ["wsgi-app"],
        beginnerFriendly: true,
      },
    ],
    featureTraces: [
      {
        id: "request-lifecycle",
        name: "WSGI Request Lifecycle & Teardown",
        description: "Tracing HTTP WSGI environment to request context push, route dispatch, and teardown.",
        steps: [
          {
            step: 1,
            title: "WSGI Callable Invocation",
            subsystem: "WSGI Core Dispatcher",
            file: "src/flask/app.py",
            line: 1475,
            description: "WSGI server calls Flask instance with environ and start_response callable.",
            codeSnippet: `def wsgi_app(self, environ, start_response):\n    ctx = self.request_context(environ)\n    ctx.push()\n    try:\n        response = self.full_dispatch_request()\n    finally:\n        ctx.pop()`,
          },
          {
            step: 2,
            title: "View Function Resolution",
            subsystem: "Werkzeug Route RuleMap",
            file: "src/flask/app.py",
            line: 1820,
            description: "Matches path in URL map, executes before_request hooks, calls endpoint view.",
            codeSnippet: `rule, args = self.url_map.bind_to_environ(environ).match()\nreturn self.view_functions[rule.endpoint](**args)`,
          },
        ],
      },
    ],
    contributorGuide: {
      stepsToStart: [
        "Set up Python 3.10+ virtualenv: 'python -m venv .venv'",
        "Install editable dependencies: 'pip install -e .[dev]'",
        "Run pytest: 'pytest tests/'",
      ],
      prerequisites: [
        "WSGI PEP 3333 standard specification",
        "Python decorators and context managers (with statement)",
      ],
      beginnerFiles: [
        { path: "src/flask/helpers.py", reason: "Utility functions like flash(), send_file(), and url_for()" },
        { path: "src/flask/signals.py", reason: "Blinker signal dispatchers for request_started and request_finished" },
      ],
    },
    fileTree: [
      {
        name: "src",
        path: "src",
        type: "directory",
        children: [
          {
            name: "flask",
            path: "src/flask",
            type: "directory",
            children: [
              { name: "app.py", path: "src/flask/app.py", type: "file", size: "48KB", owner: "pallets" },
              { name: "ctx.py", path: "src/flask/ctx.py", type: "file", size: "18KB", owner: "pallets" },
              { name: "helpers.py", path: "src/flask/helpers.py", type: "file", size: "22KB", owner: "pallets" },
            ],
          },
        ],
      },
      { name: "pyproject.toml", path: "pyproject.toml", type: "file", size: "4KB", owner: "pallets" },
    ],
    chatHistory: [
      {
        sender: "user",
        message: "How does Flask isolate request objects across concurrent threads?",
      },
      {
        sender: "gitwise",
        message: "Flask uses context-local stacks implemented via Werkzeug's `LocalStack` (backed by Python's `contextvars.ContextVar`), guaranteeing thread and async-task safety.",
        citations: [
          { file: "src/flask/ctx.py", line: 72, snippet: "_cv_request: ContextVar[RequestContext] = ContextVar('flask.request_ctx')" },
        ],
      },
    ],
  },
  "meshery/meshery": {
    id: "meshery",
    name: "meshery/meshery",
    owner: "meshery",
    branch: "master",
    stars: 6200,
    forks: 1800,
    indexedFiles: 1850,
    primaryLanguage: "Go",
    description: "Cloud Native Management Plane and Extensible Service Mesh Manager.",
    subsystems: [
      {
        id: "mesh-orchestrator",
        name: "Mesh Orchestrator Server",
        fileCount: 420,
        entryPoint: "server/main.go",
        description: "Core REST and GraphQL management daemon coordinating multi-mesh topologies.",
        language: "Go",
        connections: ["adapter-broker", "meshsync"],
        beginnerFriendly: false,
      },
      {
        id: "adapter-broker",
        name: "Mesh Adapter Broker",
        fileCount: 160,
        entryPoint: "server/handlers/adapter_handler.go",
        description: "gRPC client communicating with external adapters (Istio, Linkerd, Consul).",
        language: "Go",
        connections: [],
        beginnerFriendly: true,
      },
      {
        id: "meshsync",
        name: "MeshSync Discovery Engine",
        fileCount: 210,
        entryPoint: "server/meshsync/discovery.go",
        description: "Watches Kubernetes custom resources in realtime and streams live cluster state.",
        language: "Go",
        connections: ["mesh-orchestrator"],
        beginnerFriendly: true,
      },
    ],
    featureTraces: [
      {
        id: "pattern-deployment",
        name: "Cloud Native Pattern Deployment",
        description: "Tracing a declarative service mesh pattern deployment down to adapter gRPC calls.",
        steps: [
          {
            step: 1,
            title: "Declarative Pattern Submission",
            subsystem: "Mesh Orchestrator Server",
            file: "server/handlers/pattern_handler.go",
            line: 94,
            description: "User submits service mesh design pattern via UI or CLI.",
            codeSnippet: `func (h *Handler) PatternFileHandler(w http.ResponseWriter, req *http.Request) {\n  parsedPattern, err := parseDesignPattern(req.Body)\n  go h.broker.DeployPattern(parsedPattern)\n}`,
          },
          {
            step: 2,
            title: "Adapter gRPC Dispatch",
            subsystem: "Mesh Adapter Broker",
            file: "server/models/meshery_adapter.go",
            line: 142,
            description: "Translates pattern into vendor-specific custom resources (e.g. Istio VirtualService).",
            codeSnippet: `client.ApplyOperation(ctx, &meshes.ApplyRule{ Rule: patternRule })`,
          },
        ],
      },
    ],
    contributorGuide: {
      stepsToStart: [
        "Fork and clone repository",
        "Start backend server: 'make server'",
        "Start UI workbench: 'cd ui && npm run dev'",
        "Check Meshery discussion forum and open issues tagged 'help-wanted'",
      ],
      prerequisites: [
        "Basic Go and Docker/Kubernetes familiarity",
        "Understanding of service mesh concepts (mTLS, traffic shifting)",
      ],
      beginnerFiles: [
        { path: "server/models/handlers.go", reason: "Standard HTTP handler definitions and response helpers" },
        { path: "server/helpers/utils.go", reason: "String manipulation, formatting, and file loaders" },
      ],
    },
    fileTree: [
      {
        name: "server",
        path: "server",
        type: "directory",
        children: [
          { name: "main.go", path: "server/main.go", type: "file", size: "34KB", owner: "meshery" },
          { name: "handlers/pattern_handler.go", path: "server/handlers/pattern_handler.go", type: "file", size: "28KB", owner: "meshery" },
          { name: "models/meshery_adapter.go", path: "server/models/meshery_adapter.go", type: "file", size: "21KB", owner: "meshery" },
        ],
      },
      { name: "Makefile", path: "Makefile", type: "file", size: "6KB", owner: "meshery" },
    ],
    chatHistory: [
      {
        sender: "user",
        message: "How does Meshery communicate with service mesh adapters?",
      },
      {
        sender: "gitwise",
        message: "Meshery core communicates with mesh adapters (such as `meshery-istio` or `meshery-linkerd`) over bidirectional gRPC streams defined in the Meshery Adapter proto interface.",
        citations: [
          { file: "server/models/meshery_adapter.go", line: 68, snippet: "type AdapterClient interface { ApplyOperation(ctx context.Context, req *ApplyRule) }" },
        ],
      },
    ],
  },
};
