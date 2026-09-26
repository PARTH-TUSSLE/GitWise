export interface AffectedFile {
  path: string;
  linesChanged: number;
  role: "entry" | "core_logic" | "type_contract" | "test_spec";
  description: string;
  snippet?: string;
}

export interface ImplementationStep {
  stepNumber: number;
  title: string;
  file: string;
  action: "modify" | "create" | "test";
  explanation: string;
  diff: {
    target: string;
    before?: string;
    after: string;
  };
}

export interface GroundedTestSuite {
  name: string;
  type: "unit" | "integration" | "e2e";
  testFile: string;
  command: string;
  expectedOutput: string;
}

export interface IssueModel {
  id: string;
  number: number;
  repo: string;
  title: string;
  status: "open" | "in_triage" | "investigating";
  subsystem: string;
  reportedDate: string;
  author: string;
  blastRadius: {
    score: number; // e.g. 2.4 (scale 1-5)
    fileCount: number;
    subsystemsAffected: string[];
    riskAssessment: string;
  };
  summary: string;
  prerequisites: string[];
  affectedFiles: AffectedFile[];
  stages: {
    stage01_triage: {
      rootCauseAnalysis: string;
      reproductionSteps: string[];
      scopeBoundary: string;
    };
    stage02_impacted_paths: {
      callChain: string[];
      criticalSymbols: string[];
      stateMutations: string;
    };
    stage03_blueprint: {
      steps: ImplementationStep[];
    };
  };
  testStrategy: {
    suites: GroundedTestSuite[];
    edgeCases: string[];
    verificationChecklist: string[];
  };
}

export const ISSUES_DATABASE: Record<string, IssueModel[]> = {
  "vercel/next.js": [
    {
      id: "nextjs-54821",
      number: 54821,
      repo: "vercel/next.js",
      title: "Server Action payload decryption race during streaming hydration",
      status: "open",
      subsystem: "App Router Server Runtime",
      reportedDate: "2026-09-12",
      author: "leerob_contrib",
      blastRadius: {
        score: 2.4,
        fileCount: 4,
        subsystemsAffected: ["App Router Server Runtime", "Streaming SSR Engine"],
        riskAssessment: "Medium blast radius: touches server action handler headers and encryption argument binding, isolated from Turbopack compiler.",
      },
      summary: "When multiple form actions trigger concurrently while the initial HTML stream is still being piped to the browser, the action argument encryption key token is prematurely invalidated, causing HTTP 500 'Failed to decrypt action bound args'.",
      prerequisites: [
        "Familiarity with React Server Actions and Flight protocol",
        "Understanding of Web Crypto API and HMAC key derivation",
        "Node.js HTTP streaming pipeline (TransformStream / ReadableStream)",
      ],
      affectedFiles: [
        {
          path: "packages/next/src/server/app-render/action-handler.ts",
          linesChanged: 34,
          role: "core_logic",
          description: "Contains server action request boundary, parses 'next-action' header and decrypts arguments.",
          snippet: `export async function handleServerActionRequest(req: Request) {\n  const actionId = req.headers.get('next-action');\n  // Need to ensure action token cache survives pending stream chunks\n}`,
        },
        {
          path: "packages/next/src/server/app-render/action-encryption.ts",
          linesChanged: 18,
          role: "core_logic",
          description: "Derives temporary action argument encryption keys using session salt.",
          snippet: `export async function decryptActionBoundArgs(actionId: string, boundPayload: string) {\n  // Validate ciphertext timestamp against active stream window\n}`,
        },
        {
          path: "packages/next/src/client/components/react-dev-overlay/client-action.ts",
          linesChanged: 12,
          role: "entry",
          description: "Client dispatcher stub that packages formData and invokes window.fetch with next-action header.",
          snippet: `export function dispatchServerAction(actionId: string, boundArgs: any[]) {\n  return fetch(window.location.href, { method: 'POST' });\n}`,
        },
        {
          path: "test/e2e/app-dir/actions-encryption/actions-encryption.test.ts",
          linesChanged: 42,
          role: "test_spec",
          description: "Integration test suite validating concurrent action submissions during chunked transfer encoding.",
        },
      ],
      stages: {
        stage01_triage: {
          rootCauseAnalysis: "Action handler checks token expiration against req.headers['date'] instead of the server session stream timestamp. In concurrent bursts, the cipher payload verification window expires before the stream body flushes.",
          reproductionSteps: [
            "Render an App Router page with multiple <form action={myServerAction}> inputs.",
            "Throttle network downlink to 2G simulating slow stream flush.",
            "Submit Form A immediately followed by Form B within 80ms.",
            "Form B throws 500 Internal Server Error in decryption layer.",
          ],
          scopeBoundary: "Confined to server action dispatch handler. Does not alter React DOM Flight serialization or Turbopack chunking.",
        },
        stage02_impacted_paths: {
          callChain: [
            "client-action.ts:dispatchServerAction()",
            "action-handler.ts:handleServerActionRequest()",
            "action-encryption.ts:decryptActionBoundArgs()",
            "app-render.tsx:renderToHTMLOrFlight()",
          ],
          criticalSymbols: [
            "decryptActionBoundArgs(actionId, payload)",
            "handleServerActionRequest(req)",
            "ActionBoundArgsCache",
          ],
          stateMutations: "Expands the active token cache window by associating decrypt keys with the parent request lifecycle context instead of a single-use token.",
        },
        stage03_blueprint: {
          steps: [
            {
              stepNumber: 1,
              title: "Persist action decryption context across streaming lifetime",
              file: "packages/next/src/server/app-render/action-encryption.ts",
              action: "modify",
              explanation: "Update `decryptActionBoundArgs` to accept the request stream lifecycle token, allowing retry within the active HTTP connection window.",
              diff: {
                target: "packages/next/src/server/app-render/action-encryption.ts",
                before: `export async function decryptActionBoundArgs(actionId: string, payload: string) {\n  if (isTokenExpired(payload)) throw new Error('Token expired');\n  return unsealPayload(payload);\n}`,
                after: `export async function decryptActionBoundArgs(actionId: string, payload: string, reqCtx?: RequestContext) {\n  const isValid = validateTokenWindow(payload, reqCtx?.streamCreatedAt);\n  if (!isValid) throw new DecryptionError('Token verification expired', { actionId });\n  return unsealPayload(payload);\n}`,
              },
            },
            {
              stepNumber: 2,
              title: "Pass request context into action handler execution boundary",
              file: "packages/next/src/server/app-render/action-handler.ts",
              action: "modify",
              explanation: "Extract stream initiation timestamp from request headers and thread it into `decryptActionBoundArgs`.",
              diff: {
                target: "packages/next/src/server/app-render/action-handler.ts",
                before: `const decryptedArgs = await decryptActionBoundArgs(actionId, boundPayload);`,
                after: `const reqCtx = { streamCreatedAt: req.headers.get('x-next-stream-time') || Date.now() };\nconst decryptedArgs = await decryptActionBoundArgs(actionId, boundPayload, reqCtx);`,
              },
            },
            {
              stepNumber: 3,
              title: "Add concurrency regression test case",
              file: "test/e2e/app-dir/actions-encryption/actions-encryption.test.ts",
              action: "test",
              explanation: "Simulate concurrent streaming action calls with throttled chunk transfer.",
              diff: {
                target: "test/e2e/app-dir/actions-encryption/actions-encryption.test.ts",
                after: `it('handles concurrent server actions under slow chunked streaming without decryption failure', async () => {\n  const browser = await next.browser('/actions-stream-test');\n  await Promise.all([\n    browser.elementById('form-a-submit').click(),\n    browser.elementById('form-b-submit').click(),\n  ]);\n  expect(await browser.elementById('status').text()).toBe('BOTH_RESOLVED');\n});`,
              },
            },
          ],
        },
      },
      testStrategy: {
        suites: [
          {
            name: "Server Action Encryption E2E Suite",
            type: "e2e",
            testFile: "test/e2e/app-dir/actions-encryption/actions-encryption.test.ts",
            command: "pnpm test-e2e test/e2e/app-dir/actions-encryption",
            expectedOutput: "PASS test/e2e/app-dir/actions-encryption/actions-encryption.test.ts (2 passed, 0 failed)",
          },
          {
            name: "App Router Server Runtime Unit Tests",
            type: "unit",
            testFile: "packages/next/src/server/app-render/action-handler.test.ts",
            command: "pnpm jest packages/next/src/server/app-render/action-handler.test.ts",
            expectedOutput: "PASS packages/next/src/server/app-render/action-handler.test.ts (14 passed, 0 failed)",
          },
        ],
        edgeCases: [
          "Client re-submits action after server stream has completed and closed.",
          "Tampered payload containing invalid base64 characters.",
          "Action invocation with zero bound arguments (null payload).",
          "Cross-origin POST request missing next-action header.",
        ],
        verificationChecklist: [
          "Run 'pnpm build-native' to ensure Turbopack Rust bindings compile cleanly.",
          "Execute action handler unit tests with 100 iterations under Jest worker pool.",
          "Verify that error response carries HTTP 403 / 400 instead of unhandled 500.",
        ],
      },
    },
    {
      id: "nextjs-56104",
      number: 56104,
      repo: "vercel/next.js",
      title: "Turbopack HMR socket drops on unhandled dynamic import error",
      status: "in_triage",
      subsystem: "Turbopack Compiler",
      reportedDate: "2026-09-18",
      author: "sokra_watch",
      blastRadius: {
        score: 1.8,
        fileCount: 3,
        subsystemsAffected: ["Turbopack Compiler"],
        riskAssessment: "Low blast radius: localized to dev overlay WebSocket retry loop and crate HMR channel.",
      },
      summary: "When a dynamic import fails due to a missing file in development, Turbopack triggers an unhandled rejection in the client HMR receiver, causing the WebSocket to close permanently without reconnecting.",
      prerequisites: [
        "Turbopack architecture & Rust N-API bindings",
        "WebSocket lifecycle & reconnect exponential backoff",
      ],
      affectedFiles: [
        {
          path: "crates/turbopack/src/lib.rs",
          linesChanged: 14,
          role: "core_logic",
          description: "WebSocket connection listener for local dev server.",
        },
        {
          path: "packages/next/src/client/components/react-dev-overlay/client.ts",
          linesChanged: 28,
          role: "entry",
          description: "Handles incoming HMR patches from Turbopack.",
        },
      ],
      stages: {
        stage01_triage: {
          rootCauseAnalysis: "Client-side HMR event dispatcher lacks an error boundary around import() resolution.",
          reproductionSteps: [
            "Start dev server with next dev --turbo.",
            "Create a dynamic import pointing to non-existent './module'.",
            "Save file. HMR fails and dev console reports socket closed.",
          ],
          scopeBoundary: "Client HMR listener and error event emission.",
        },
        stage02_impacted_paths: {
          callChain: [
            "crates/turbopack/src/lib.rs:watch_entrypoint()",
            "client.ts:onHmrMessage()",
            "client.ts:reconnectSocket()",
          ],
          criticalSymbols: ["onHmrMessage", "reconnectSocket", "TurbopackDevServer"],
          stateMutations: "Adds persistent exponential backoff state to WebSocket client.",
        },
        stage03_blueprint: {
          steps: [
            {
              stepNumber: 1,
              title: "Wrap dynamic import resolution in try/catch in HMR client",
              file: "packages/next/src/client/components/react-dev-overlay/client.ts",
              action: "modify",
              explanation: "Prevent unhandled error from bubble-terminating the socket session.",
              diff: {
                target: "packages/next/src/client/components/react-dev-overlay/client.ts",
                after: `try {\n  await applyHmrUpdate(payload);\n} catch (err) {\n  reportDevOverlayError(err);\n  scheduleSocketKeepalive();\n}`,
              },
            },
          ],
        },
      },
      testStrategy: {
        suites: [
          {
            name: "Turbopack HMR Client Suite",
            type: "unit",
            testFile: "packages/next/src/client/components/react-dev-overlay/client.test.ts",
            command: "pnpm test packages/next/src/client/components/react-dev-overlay",
            expectedOutput: "PASS packages/next/src/client/components/react-dev-overlay (8 passed)",
          },
        ],
        edgeCases: ["Network drops during socket reconnect attempt."],
        verificationChecklist: ["Verify WebSocket retries after syntax error fix."],
      },
    },
  ],
  "facebook/react": [
    {
      id: "react-29381",
      number: 29381,
      repo: "facebook/react",
      title: "useActionState pending state desynchronization under high concurrency",
      status: "open",
      subsystem: "Fiber Reconciler",
      reportedDate: "2026-09-08",
      author: "acdlite_tracker",
      blastRadius: {
        score: 3.2,
        fileCount: 5,
        subsystemsAffected: ["Fiber Reconciler", "React DOM Server"],
        riskAssessment: "High complexity: touches React Fiber work loop scheduling and transition queue priority lanes.",
      },
      summary: "In React 19, dispatching an action via useActionState while a parent transition is yielding causes the 'isPending' boolean flag to resolve to false before the transition commit phase concludes.",
      prerequisites: [
        "React Fiber architecture and priority lanes (SyncLane, TransitionLane)",
        "Understanding of React 19 Action queue semantics",
      ],
      affectedFiles: [
        {
          path: "packages/react-reconciler/src/ReactFiberHooks.js",
          linesChanged: 48,
          role: "core_logic",
          description: "Implementation of useActionState and hook update queues.",
        },
        {
          path: "packages/react-reconciler/src/ReactFiberWorkLoop.js",
          linesChanged: 22,
          role: "core_logic",
          description: "Fiber tree reconciliation work loop and commit scheduler.",
        },
      ],
      stages: {
        stage01_triage: {
          rootCauseAnalysis: "Action hook updates mark the pending flag complete upon dispatch resolution rather than after the Fiber lane commit phase.",
          reproductionSteps: [
            "Create component with const [state, dispatch, isPending] = useActionState(fn, null);",
            "Wrap dispatch in startTransition(() => dispatch());",
            "Inspect isPending during concurrent yield.",
          ],
          scopeBoundary: "ReactFiberHooks.js updateActionStateQueue.",
        },
        stage02_impacted_paths: {
          callChain: [
            "ReactFiberHooks.js:dispatchAction()",
            "ReactFiberWorkLoop.js:ensureRootIsScheduled()",
            "ReactFiberCommitWork.js:commitHookEffectListMount()",
          ],
          criticalSymbols: ["updateActionState", "mountActionState", "TransitionLane"],
          stateMutations: "Ties pending flag lifecycle to EntangledActionLane resolution.",
        },
        stage03_blueprint: {
          steps: [
            {
              stepNumber: 1,
              title: "Entangle action pending lane with active transition root",
              file: "packages/react-reconciler/src/ReactFiberHooks.js",
              action: "modify",
              explanation: "Ensure `isPending` state only flips after the root commit phase marks the transition lane clean.",
              diff: {
                target: "packages/react-reconciler/src/ReactFiberHooks.js",
                after: `hook.memoizedState[2] = isLaneActive(root.pendingLanes, actionLane);`,
              },
            },
          ],
        },
      },
      testStrategy: {
        suites: [
          {
            name: "React Concurrent Hooks Spec",
            type: "unit",
            testFile: "packages/react-reconciler/src/__tests__/ReactHooks-test.js",
            command: "yarn test packages/react-reconciler/src/__tests__/ReactHooks-test.js -t 'useActionState'",
            expectedOutput: "PASS packages/react-reconciler/src/__tests__/ReactHooks-test.js (32 passed)",
          },
        ],
        edgeCases: ["Nested startTransition calls.", "Action throwing unhandled exception."],
        verificationChecklist: ["Verify yarn test-dom passes completely."],
      },
    },
  ],
  "kubernetes/kubernetes": [
    {
      id: "k8s-119420",
      number: 119420,
      repo: "kubernetes/kubernetes",
      title: "kube-scheduler node affinity score underflow on high pod density",
      status: "open",
      subsystem: "kube-scheduler",
      reportedDate: "2026-08-28",
      author: "k8s_sig_scheduling",
      blastRadius: {
        score: 2.1,
        fileCount: 4,
        subsystemsAffected: ["kube-scheduler", "pkg/scheduler/framework"],
        riskAssessment: "Medium blast radius: isolated to scheduling plugin scoring algorithm without modifying cluster state store.",
      },
      summary: "When evaluating PreferredDuringSchedulingIgnoredDuringExecution node affinity on nodes with >500 co-located pods, integer multiplication overflows int64 score range, resulting in negative affinity weights.",
      prerequisites: [
        "Go programming language (Go 1.22+)",
        "Kubernetes Scheduling Framework plugins (PreScore, Score, NormalizeScore)",
      ],
      affectedFiles: [
        {
          path: "pkg/scheduler/framework/plugins/nodeaffinity/node_affinity.go",
          linesChanged: 24,
          role: "core_logic",
          description: "Scoring logic for PreferredDuringSchedulingIgnoredDuringExecution.",
        },
        {
          path: "pkg/scheduler/framework/plugins/nodeaffinity/node_affinity_test.go",
          linesChanged: 36,
          role: "test_spec",
          description: "Unit tests simulating high-density node affinity calculations.",
        },
      ],
      stages: {
        stage01_triage: {
          rootCauseAnalysis: "Weight multiplication accumulator uses int32 cast before multiplying matching term counts.",
          reproductionSteps: [
            "Deploy node with 600 running pods.",
            "Schedule pod with 10 preferred affinity terms.",
            "Scheduler logs node score as negative integer.",
          ],
          scopeBoundary: "pkg/scheduler/framework/plugins/nodeaffinity.",
        },
        stage02_impacted_paths: {
          callChain: [
            "plugin.go:Score()",
            "node_affinity.go:calculateAffinityWeight()",
            "framework.go:NormalizeScore()",
          ],
          criticalSymbols: ["Score", "NormalizeScore", "NodeAffinity"],
          stateMutations: "Clamps intermediate score values within [framework.MinNodeScore, framework.MaxNodeScore].",
        },
        stage03_blueprint: {
          steps: [
            {
              stepNumber: 1,
              title: "Promote score accumulator to int64 with bounds saturation",
              file: "pkg/scheduler/framework/plugins/nodeaffinity/node_affinity.go",
              action: "modify",
              explanation: "Use math/bits or explicit saturation checks to prevent score underflow.",
              diff: {
                target: "pkg/scheduler/framework/plugins/nodeaffinity/node_affinity.go",
                after: `var totalScore int64\nif totalScore > framework.MaxNodeScore {\n    totalScore = framework.MaxNodeScore\n}`,
              },
            },
          ],
        },
      },
      testStrategy: {
        suites: [
          {
            name: "Node Affinity Plugin Suite",
            type: "unit",
            testFile: "pkg/scheduler/framework/plugins/nodeaffinity/node_affinity_test.go",
            command: "go test -v ./pkg/scheduler/framework/plugins/nodeaffinity/...",
            expectedOutput: "ok  k8s.io/kubernetes/pkg/scheduler/framework/plugins/nodeaffinity 1.24s",
          },
        ],
        edgeCases: ["Zero matching terms.", "Node score saturation at MaxNodeScore (100)."],
        verificationChecklist: ["Run make test-integration to confirm scheduler end-to-end."],
      },
    },
  ],
  "pallets/flask": [
    {
      id: "flask-5122",
      number: 5122,
      repo: "pallets/flask",
      title: "Context locals leak across async generator route handlers",
      status: "open",
      subsystem: "Context Locals & Globals",
      reportedDate: "2026-09-02",
      author: "pallets_triage",
      blastRadius: {
        score: 1.5,
        fileCount: 2,
        subsystemsAffected: ["Context Locals & Globals", "WSGI Request Pipeline"],
        riskAssessment: "Low blast radius: localized to Flask appcontext and request context copy utilities.",
      },
      summary: "In async views returning StreamingResponse with generator yield, task context variable token is not detached upon client disconnect, leaking `g` state to subsequent requests.",
      prerequisites: [
        "Python 3.10+ contextvars module",
        "WSGI vs ASGI request context lifecycle in Werkzeug",
      ],
      affectedFiles: [
        {
          path: "src/flask/ctx.py",
          linesChanged: 16,
          role: "core_logic",
          description: "RequestContext and AppContext stack manager.",
        },
      ],
      stages: {
        stage01_triage: {
          rootCauseAnalysis: "RequestContext.pop() does not invoke token.reset() when generator raises GeneratorExit.",
          reproductionSteps: ["Define async generator route.", "Abort client request mid-stream.", "Inspect g in next request."],
          scopeBoundary: "src/flask/ctx.py.",
        },
        stage02_impacted_paths: {
          callChain: ["ctx.py:push()", "ctx.py:pop()", "app.py:handle_exception()"],
          criticalSymbols: ["_cv_request", "_cv_app", "RequestContext"],
          stateMutations: "Resets context variable tokens reliably in a finally block.",
        },
        stage03_blueprint: {
          steps: [
            {
              stepNumber: 1,
              title: "Ensure contextvar reset token executed in generator cleanup",
              file: "src/flask/ctx.py",
              action: "modify",
              explanation: "Guarantee token reset on GeneratorExit exceptions.",
              diff: {
                target: "src/flask/ctx.py",
                after: `finally:\n    _cv_request.reset(self._token)`,
              },
            },
          ],
        },
      },
      testStrategy: {
        suites: [
          {
            name: "Flask Context Test Suite",
            type: "unit",
            testFile: "tests/test_async.py",
            command: "pytest tests/test_async.py -k test_context_leak",
            expectedOutput: "1 passed in 0.18s",
          },
        ],
        edgeCases: ["Nested request context push."],
        verificationChecklist: ["Run tox -e py312."],
      },
    },
  ],
  "meshery/meshery": [
    {
      id: "meshery-8491",
      number: 8491,
      repo: "meshery/meshery",
      title: "MeshSync pattern reconciliation drops transient service mesh CRDs",
      status: "open",
      subsystem: "MeshSync Discovery Engine",
      reportedDate: "2026-09-04",
      author: "layer5_contrib",
      blastRadius: {
        score: 2.8,
        fileCount: 5,
        subsystemsAffected: ["MeshSync Discovery Engine", "Pattern Engine & OAM"],
        riskAssessment: "Medium blast radius: affects MeshSync broker and NATS event dispatcher.",
      },
      summary: "When Istio or Linkerd operators update custom resource definitions rapidly during canary rollout, MeshSync event broker drops schema updates due to locked channel queue.",
      prerequisites: [
        "Kubernetes Dynamic Client & Informer caches",
        "NATS message streaming and queue subscriptions in Go",
      ],
      affectedFiles: [
        {
          path: "server/meshmodel/meshsync/listener.go",
          linesChanged: 32,
          role: "core_logic",
          description: "Listens for cluster resource discovery events.",
        },
      ],
      stages: {
        stage01_triage: {
          rootCauseAnalysis: "Event channel has unbuffered capacity of 100 items; bursts exceed buffer and drop messages without retry.",
          reproductionSteps: ["Apply 150 CRDs simultaneously.", "Observe dropped events in meshsync logs."],
          scopeBoundary: "server/meshmodel/meshsync.",
        },
        stage02_impacted_paths: {
          callChain: ["listener.go:ProcessEvent()", "broker.go:Publish()", "pattern.go:Reconcile()"],
          criticalSymbols: ["ProcessEvent", "MeshSyncBroker", "Reconcile"],
          stateMutations: "Implements buffered ring queue with backpressure logging.",
        },
        stage03_blueprint: {
          steps: [
            {
              stepNumber: 1,
              title: "Increase channel buffer and add exponential retry backpressure",
              file: "server/meshmodel/meshsync/listener.go",
              action: "modify",
              explanation: "Prevent channel blocking during heavy cluster event flushes.",
              diff: {
                target: "server/meshmodel/meshsync/listener.go",
                after: `select {\ncase l.eventChan <- event:\ncase <-time.After(200 * time.Millisecond):\n    l.logger.Warn("MeshSync channel full, buffering")\n}`,
              },
            },
          ],
        },
      },
      testStrategy: {
        suites: [
          {
            name: "MeshSync Broker Tests",
            type: "unit",
            testFile: "server/meshmodel/meshsync/listener_test.go",
            command: "go test ./server/meshmodel/meshsync/...",
            expectedOutput: "ok  github.com/meshery/meshery/server/meshmodel/meshsync 0.89s",
          },
        ],
        edgeCases: ["Cluster API connection disconnect."],
        verificationChecklist: ["Run make test-server."],
      },
    },
  ],
};
