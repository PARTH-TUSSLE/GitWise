export interface MentorshipTrack {
  id: string;
  title: string;
  category: "quickstart" | "skill_match" | "architecture" | "starter_code";
  description: string;
  badge: string;
  queryPrompt: string;
}

export interface SkillMatchIssue {
  issueNumber: number;
  title: string;
  matchScore: number; // e.g. 92%
  matchRationale: string;
  requiredSkills: string[];
  subsystem: string;
}

export interface MentorMessage {
  id: string;
  sender: "user" | "mentor";
  timestamp: string;
  content: string;
  citations?: Array<{
    file: string;
    line?: number;
    description: string;
  }>;
  actionCommands?: string[];
}

export interface RepoMentorModel {
  repo: string;
  name: string;
  primaryLanguage: string;
  tagline: string;
  contributorMatch: {
    username: string;
    overallScore: number;
    matchedSkills: string[];
    skillsToAcquire: string[];
    rationale: string;
  };
  tracks: MentorshipTrack[];
  matchedIssues: SkillMatchIssue[];
  defaultMessages: MentorMessage[];
  knowledgeBase: Record<string, {
    content: string;
    citations?: Array<{ file: string; line?: number; description: string }>;
    actionCommands?: string[];
  }>;
}

export const MENTOR_DATA: Record<string, RepoMentorModel> = {
  "vercel/next.js": {
    repo: "vercel/next.js",
    name: "Next.js",
    primaryLanguage: "TypeScript & Rust",
    tagline: "The React Framework for the Web with App Router and Turbopack compiler.",
    contributorMatch: {
      username: "alexR_dev",
      overallScore: 92,
      matchedSkills: ["TypeScript 5.x", "React 19 Server Components", "HTTP Chunked Streaming", "Turbopack N-API"],
      skillsToAcquire: ["Rust Proc Macros (crates/turbopack)", "SWC Custom AST Visitors"],
      rationale: "Strong alignment based on your 148 merged PRs in TypeScript/React monorepos and demonstrated work with chunked stream decoders.",
    },
    tracks: [
      {
        id: "track-start",
        title: "Where should I start?",
        category: "quickstart",
        description: "Curated beginner-friendly entry points and local workspace verification.",
        badge: "STARTER",
        queryPrompt: "Where should I start as a new contributor to Next.js?",
      },
      {
        id: "track-skills",
        title: "Match issues to my skills",
        category: "skill_match",
        description: "Cross-references your GitStat history against active open issues.",
        badge: "92% MATCH",
        queryPrompt: "Which Next.js issues best match my demonstrated skills?",
      },
      {
        id: "track-arch",
        title: "Explain Server Actions architecture",
        category: "architecture",
        description: "End-to-end walkthrough from client stub to server flight serialization.",
        badge: "DEEP-DIVE",
        queryPrompt: "Explain how Server Actions work internally in the App Router.",
      },
      {
        id: "track-test",
        title: "How do I run tests locally?",
        category: "quickstart",
        description: "Commands for unit, integration, and Turbopack native test suites.",
        badge: "VERIFY",
        queryPrompt: "How do I run and debug the Next.js test suite locally?",
      },
    ],
    matchedIssues: [
      {
        issueNumber: 54821,
        title: "Server Action payload decryption race during streaming hydration",
        matchScore: 95,
        matchRationale: "Matches your deep background with Web Crypto APIs and streaming request pipelines.",
        requiredSkills: ["TypeScript", "Web Crypto API", "App Router"],
        subsystem: "App Router Server Runtime",
      },
      {
        issueNumber: 56104,
        title: "Turbopack HMR socket drops on unhandled dynamic import error",
        matchScore: 89,
        matchRationale: "Involves dev-overlay WebSocket error handling in client packages.",
        requiredSkills: ["TypeScript", "WebSocket", "React Dev Overlay"],
        subsystem: "Turbopack Compiler",
      },
    ],
    defaultMessages: [
      {
        id: "msg-1",
        sender: "mentor",
        timestamp: "02:22",
        content: `Welcome to the Next.js Contributor Mentorship Workbench!

I am your Socratic engineering mentor for **vercel/next.js**. Rather than generating boilerplate code, my role is to help you build a deep mental model of the codebase, navigate monorepo architecture, and find high-impact contribution opportunities matched to your skills.

Your **GitStat** profile indicates extensive experience in TypeScript and React 19 concurrent internals, which gives you a strong **92% contributor match** for our App Router server runtime.

How would you like to begin your onboarding today?`,
        citations: [
          { file: "packages/next/src/server/app-render/app-render.tsx", line: 30, description: "App Router Flight render entry point" },
        ],
      },
    ],
    knowledgeBase: {
      "start": {
        content: `### Getting Started in Next.js

For new contributors, the best way to build confidence is by starting in well-isolated client utility packages where unit test coverage is high and build cycles are fast:

1. **Recommended Starter Module:**
   \`packages/next/src/client/link.tsx\`
   This module handles prefetching and client-side navigation. It has zero coupling to the Rust compiler and includes comprehensive Jest tests.

2. **Local Workspace Setup:**
   Run the following commands in the root of the repository to bootstrap the monorepo:`,
        actionCommands: [
          "pnpm install",
          "pnpm build-native",
          "pnpm test packages/next/src/client/link.test.tsx",
        ],
        citations: [
          { file: "packages/next/src/client/link.tsx", line: 1, description: "Next.js client-side Link component" },
          { file: "test/e2e/app-dir/actions-stream/actions-stream.test.ts", line: 1, description: "Server action E2E test suite" },
        ],
      },
      "skills": {
        content: `### Personalized Skill Matching

Based on your public GitHub footprint analyzed via **GITSTAT**:
- **Demonstrated Strengths:** High velocity in TypeScript 5.x, React Server Components (RSC), and HTTP streaming adapters.
- **Recommended Issue #54821:** *Server Action payload decryption race during streaming hydration* (95% match).
  - **Why:** The issue is isolated to \`packages/next/src/server/app-render/action-handler.ts\` and involves token lifespan management during chunked transfer encoding.
  - **Prerequisite:** Understand how \`renderToHTMLOrFlight\` handles socket drain events.`,
        citations: [
          { file: "packages/next/src/server/app-render/action-handler.ts", line: 42, description: "Server Action dispatch handler" },
          { file: "packages/next/src/server/app-render/action-encryption.ts", line: 18, description: "HMAC argument encryption utilities" },
        ],
      },
      "architecture": {
        content: `### Architecture: Server Actions Execution Flow

Server Actions in Next.js bridge client event dispatchers with server-side Flight serialization:

1. **Client Dispatch:** When a form submits or \`startTransition\` invokes an action, \`dispatchServerAction\` executes a POST request containing the \`next-action\` header and encrypted bound payload.
2. **Action Handler Execution:** \`handleServerActionRequest\` in \`action-handler.ts\` validates the origin, decrypts arguments using \`decryptActionBoundArgs\`, and executes the action closure.
3. **Flight Revalidation:** The handler invokes \`revalidatePath\` or \`revalidateTag\`, flushes the updated component tree via the Flight stream, and returns the serialized state.`,
        citations: [
          { file: "packages/next/src/server/app-render/action-handler.ts", line: 55, description: "Action decryption & execution boundary" },
          { file: "packages/next/src/server/app-render/app-render.tsx", line: 88, description: "Flight stream serialization pipeline" },
        ],
      },
      "test": {
        content: `### Running and Debugging Tests

Next.js uses a hybrid testing suite consisting of Jest for unit tests and a custom Playwright runner for end-to-end integration tests:

- **Unit Tests:** Fast feedback loop for client and server runtime logic.
- **E2E Tests:** Launches a real Next.js server instance in a sandbox directory and validates HTTP streaming, HMR, and browser behavior.`,
        actionCommands: [
          "pnpm jest packages/next/src/server/app-render/action-handler.test.ts",
          "pnpm test-e2e test/e2e/app-dir/actions-stream",
        ],
        citations: [
          { file: "test/e2e/app-dir/actions-stream/actions-stream.test.ts", line: 1, description: "E2E Stream test" },
        ],
      },
    },
  },
  "facebook/react": {
    repo: "facebook/react",
    name: "React",
    primaryLanguage: "JavaScript & Flow",
    tagline: "The library for web and native user interfaces.",
    contributorMatch: {
      username: "alexR_dev",
      overallScore: 88,
      matchedSkills: ["React Internals", "Concurrent Mode", "JavaScript AST"],
      skillsToAcquire: ["Meta Flow typechecker", "React Native Fabric reconciler"],
      rationale: "High match for packages/react-reconciler hooks implementation.",
    },
    tracks: [
      {
        id: "react-start",
        title: "Where should I start?",
        category: "quickstart",
        description: "Exploring React repo structure and running tests with yarn test.",
        badge: "STARTER",
        queryPrompt: "Where should I start as a new contributor to React?",
      },
      {
        id: "react-fiber",
        title: "How does the Fiber reconciler work?",
        category: "architecture",
        description: "Work loop traversal, lanes priority scheduling, and commit phase.",
        badge: "CORE",
        queryPrompt: "Explain how React Fiber work loop scheduling works.",
      },
    ],
    matchedIssues: [
      {
        issueNumber: 29381,
        title: "useActionState pending state desynchronization under high concurrency",
        matchScore: 91,
        matchRationale: "Directly touches React 19 Action hook queues.",
        requiredSkills: ["React Fiber", "Concurrent Lanes"],
        subsystem: "Fiber Reconciler",
      },
    ],
    defaultMessages: [
      {
        id: "react-msg-1",
        sender: "mentor",
        timestamp: "02:22",
        content: `Welcome to the React Contributor Mentorship Workbench!

I can help you unpack React's core reconciler, work loop scheduler, and new React 19 hooks. What would you like to explore first?`,
      },
    ],
    knowledgeBase: {
      "start": {
        content: `### Getting Started in React

React's monorepo uses Yarn v1 and custom build scripts:
- **Beginner Entry:** \`packages/react/src/ReactSharedInternals.js\` or well-tested shared helpers in \`packages/shared\`.
- **Run Tests:** \`yarn test packages/react-reconciler/src/__tests__/ReactHooks-test.js\`.`,
        actionCommands: ["yarn install", "yarn test --watch"],
        citations: [{ file: "packages/react-reconciler/src/ReactFiberHooks.js", line: 1, description: "Hooks implementation" }],
      },
    },
  },
  "kubernetes/kubernetes": {
    repo: "kubernetes/kubernetes",
    name: "Kubernetes",
    primaryLanguage: "Go",
    tagline: "Production-Grade Container Orchestration.",
    contributorMatch: {
      username: "alexR_dev",
      overallScore: 78,
      matchedSkills: ["Go (Basic)", "Container Architecture", "Distributed Systems"],
      skillsToAcquire: ["k8s.io/apimachinery", "client-go Informers", "Scheduling Plugins"],
      rationale: "Solid foundation in cloud architecture; recommended to start with plugin unit tests.",
    },
    tracks: [
      {
        id: "k8s-start",
        title: "Where should I start?",
        category: "quickstart",
        description: "SIG Scheduling onboarding and building kube-scheduler locally.",
        badge: "SIG-SCHED",
        queryPrompt: "Where should I start as a new contributor to Kubernetes?",
      },
    ],
    matchedIssues: [
      {
        issueNumber: 119420,
        title: "kube-scheduler node affinity score underflow on high pod density",
        matchScore: 84,
        matchRationale: "Isolated to plugin scoring calculation without modifying cluster state store.",
        requiredSkills: ["Go", "Kubernetes Scheduling Framework"],
        subsystem: "kube-scheduler",
      },
    ],
    defaultMessages: [
      {
        id: "k8s-msg-1",
        sender: "mentor",
        timestamp: "02:22",
        content: `Welcome to the Kubernetes Contributor Mentorship Workbench!

Kubernetes is a large distributed system organized into Special Interest Groups (SIGs). I can help you understand kube-scheduler plugins and client-go caches.`,
      },
    ],
    knowledgeBase: {
      "start": {
        content: `### Getting Started in Kubernetes (SIG-Scheduling)

The easiest subsystem to contribute to is the **Scheduling Framework Plugins** under \`pkg/scheduler/framework/plugins\`. Each plugin is self-contained with independent unit tests.`,
        actionCommands: ["make WHAT=cmd/kube-scheduler", "go test -v ./pkg/scheduler/framework/plugins/..."],
        citations: [{ file: "pkg/scheduler/framework/plugins/nodeaffinity/node_affinity.go", line: 1, description: "Node affinity plugin" }],
      },
    },
  },
  "pallets/flask": {
    repo: "pallets/flask",
    name: "Flask",
    primaryLanguage: "Python",
    tagline: "The Python micro framework for building web applications.",
    contributorMatch: {
      username: "alexR_dev",
      overallScore: 82,
      matchedSkills: ["Python 3.x", "WSGI/ASGI", "HTTP Spec"],
      skillsToAcquire: ["Werkzeug context locals", "Blinker signals"],
      rationale: "Excellent fit for async route handlers and signal context isolation.",
    },
    tracks: [
      {
        id: "flask-start",
        title: "Where should I start?",
        category: "quickstart",
        description: "Setting up virtualenv and running pytest on Flask core.",
        badge: "STARTER",
        queryPrompt: "Where should I start as a new contributor to Flask?",
      },
    ],
    matchedIssues: [
      {
        issueNumber: 5122,
        title: "Context locals leak across async generator route handlers",
        matchScore: 86,
        matchRationale: "Touches contextvars management in ctx.py.",
        requiredSkills: ["Python contextvars", "Asyncio"],
        subsystem: "Context Locals & Globals",
      },
    ],
    defaultMessages: [
      {
        id: "flask-msg-1",
        sender: "mentor",
        timestamp: "02:22",
        content: `Welcome to the Flask Contributor Mentorship Workbench!

Flask has a compact, readable codebase anchored by Werkzeug and context variables. How can I guide you today?`,
      },
    ],
    knowledgeBase: {
      "start": {
        content: `### Getting Started in Flask

Flask is one of the most accessible open-source projects in Python:
- Entry point: \`src/flask/app.py\`
- Run tests: \`pytest tests/\``,
        actionCommands: ["pip install -e .", "pytest tests/"],
        citations: [{ file: "src/flask/ctx.py", line: 1, description: "Flask context stack" }],
      },
    },
  },
  "meshery/meshery": {
    repo: "meshery/meshery",
    name: "Meshery",
    primaryLanguage: "Go & React",
    tagline: "The Cloud Native Management Plane.",
    contributorMatch: {
      username: "alexR_dev",
      overallScore: 89,
      matchedSkills: ["Go", "React", "Docker/K8s"],
      skillsToAcquire: ["Service Mesh Performance (SMP)", "OAM Patterns"],
      rationale: "Full-stack match across Go server and React cloud-native dashboard.",
    },
    tracks: [
      {
        id: "msh-start",
        title: "Where should I start?",
        category: "quickstart",
        description: "Bootstrapping Meshery Server and MeshSync discovery broker.",
        badge: "STARTER",
        queryPrompt: "Where should I start as a new contributor to Meshery?",
      },
    ],
    matchedIssues: [
      {
        issueNumber: 8491,
        title: "MeshSync pattern reconciliation drops transient service mesh CRDs",
        matchScore: 90,
        matchRationale: "Combines Go event channels with Kubernetes informers.",
        requiredSkills: ["Go", "NATS Streaming"],
        subsystem: "MeshSync Discovery Engine",
      },
    ],
    defaultMessages: [
      {
        id: "msh-msg-1",
        sender: "mentor",
        timestamp: "02:22",
        content: `Welcome to the Meshery Contributor Mentorship Workbench!

Meshery provides cloud-native management across Kubernetes clusters and service meshes. I'm ready to walk you through MeshSync discovery or pattern reconciliation.`,
      },
    ],
    knowledgeBase: {
      "start": {
        content: `### Getting Started in Meshery

Meshery consists of a Go server and a React UI:
- Server entry: \`main.go\` and \`server/meshmodel\`
- Run tests: \`make test-server\``,
        actionCommands: ["make build-server", "make test-server"],
        citations: [{ file: "server/meshmodel/meshsync/listener.go", line: 1, description: "MeshSync event listener" }],
      },
    },
  },
};
