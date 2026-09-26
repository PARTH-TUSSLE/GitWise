package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gitwise/backend/internal/api/sse"
	"github.com/gitwise/backend/internal/domain"
	"github.com/gitwise/backend/internal/jobs"
	"github.com/gitwise/backend/internal/service/repo"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RepoHandler struct {
	repositories map[string]domain.RepoModel
	repoSvc      *repo.Service
	jobManager   *jobs.JobManager
	sseBroker    *sse.Broker
	logger       *slog.Logger
}

func NewRepoHandler(repoSvc *repo.Service, jobManager *jobs.JobManager, sseBroker *sse.Broker, logger *slog.Logger) *RepoHandler {
	if logger == nil {
		logger = slog.Default()
	}
	h := &RepoHandler{
		repositories: make(map[string]domain.RepoModel),
		repoSvc:      repoSvc,
		jobManager:   jobManager,
		sseBroker:    sseBroker,
		logger:       logger,
	}
	h.initMockRepos()
	return h
}

type IngestRequest struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	Ref   string `json:"ref"`
}

type IngestResponse struct {
	JobID      uuid.UUID        `json:"jobId"`
	SnapshotID uuid.UUID        `json:"snapshotId"`
	Status     domain.JobStatus `json:"status"`
	Stage      domain.JobStage  `json:"stage"`
	CommitSHA  string           `json:"commitSha"`
}

// IngestRepository handles POST /api/v1/repositories/ingest
func (h *RepoHandler) IngestRepository(w http.ResponseWriter, r *http.Request) {
	if h.repoSvc == nil {
		http.Error(w, `{"error":"Repository service is not configured"}`, http.StatusServiceUnavailable)
		return
	}

	var req IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	req.Owner = strings.TrimSpace(req.Owner)
	req.Repo = strings.TrimSpace(req.Repo)
	req.Ref = strings.TrimSpace(req.Ref)

	if req.Owner == "" || req.Repo == "" {
		http.Error(w, `{"error":"owner and repo fields are required"}`, http.StatusBadRequest)
		return
	}
	if strings.Contains(req.Owner, "/") || strings.Contains(req.Repo, "/") {
		http.Error(w, `{"error":"invalid repository identifier format"}`, http.StatusBadRequest)
		return
	}

	job, snapshot, err := h.repoSvc.Ingest(r.Context(), req.Owner, req.Repo, req.Ref)
	if err != nil {
		h.logger.Error("Repository ingestion failed to queue",
			slog.String("owner", req.Owner),
			slog.String("repo", req.Repo),
			slog.String("ref", req.Ref),
			slog.String("error", err.Error()),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to initiate repository ingestion"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(IngestResponse{
		JobID:      job.ID,
		SnapshotID: snapshot.ID,
		Status:     job.Status,
		Stage:      job.Stage,
		CommitSHA:  snapshot.CommitSHA,
	})
}

// GetJob handles GET /api/v1/jobs/{id}
func (h *RepoHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	if h.jobManager == nil {
		http.Error(w, `{"error":"Job manager is not configured"}`, http.StatusServiceUnavailable)
		return
	}

	jobIDStr := chi.URLParam(r, "id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid job uuid"}`, http.StatusBadRequest)
		return
	}

	job, err := h.jobManager.GetJob(r.Context(), jobID)
	if err != nil {
		h.logger.Warn("Job lookup failed", slog.String("job_id", jobIDStr), slog.String("error", err.Error()))
		http.Error(w, `{"error":"job not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(job)
}

// GetJobStream handles GET /api/v1/jobs/{id}/stream via Server-Sent Events
func (h *RepoHandler) GetJobStream(w http.ResponseWriter, r *http.Request) {
	if h.sseBroker == nil {
		http.Error(w, `{"error":"SSE streaming broker is not configured"}`, http.StatusServiceUnavailable)
		return
	}

	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		http.Error(w, `{"error":"job id is required"}`, http.StatusBadRequest)
		return
	}

	h.sseBroker.ServeHTTP(w, r, jobID)
}

// GetSnapshotFiles handles GET /api/v1/repositories/{owner}/{repo}/snapshots/{commitSha}/files
func (h *RepoHandler) GetSnapshotFiles(w http.ResponseWriter, r *http.Request) {
	if h.repoSvc == nil {
		http.Error(w, `{"error":"Repository service is not configured"}`, http.StatusServiceUnavailable)
		return
	}

	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")
	commitSha := chi.URLParam(r, "commitSha")

	if owner == "" || repo == "" || commitSha == "" {
		http.Error(w, `{"error":"owner, repo, and commitSha are required"}`, http.StatusBadRequest)
		return
	}

	files, err := h.repoSvc.GetSnapshotFiles(r.Context(), owner, repo, commitSha)
	if err != nil {
		h.logger.Error("Failed to fetch snapshot files",
			slog.String("owner", owner),
			slog.String("repo", repo),
			slog.String("commit_sha", commitSha),
			slog.String("error", err.Error()),
		)
		http.Error(w, `{"error":"Failed to retrieve snapshot files"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(files)
}

func (h *RepoHandler) GetRepository(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")
	if owner == "" || repo == "" {
		http.Error(w, `{"error":"owner and repo are required"}`, http.StatusBadRequest)
		return
	}

	fullName := fmt.Sprintf("%s/%s", owner, repo)
	normName := strings.ToLower(fullName)

	model, exists := h.repositories[normName]
	if !exists {
		// Dynamic fallback repository model
		model = domain.RepoModel{
			ID:              repo,
			Name:            fullName,
			Owner:           owner,
			Branch:          "main",
			Stars:           120,
			Forks:           25,
			IndexedFiles:    18,
			PrimaryLanguage: "Go",
			Description:     fmt.Sprintf("Repository intelligence snapshot for %s.", fullName),
			Subsystems: []domain.SubsystemNode{
				{
					ID:               "core-cmd",
					Name:             "CLI / Entrypoint",
					FileCount:        2,
					EntryPoint:       "cmd/server/main.go",
					Description:      "Application bootstrap, signals, and routing loop.",
					Language:         "Go",
					Connections:      []string{"internal-api"},
					BeginnerFriendly: true,
				},
				{
					ID:               "internal-api",
					Name:             "API Handlers & Middleware",
					FileCount:        8,
					EntryPoint:       "internal/api/router.go",
					Description:      "HTTP routing and domain payload serialization.",
					Language:         "Go",
					Connections:      []string{},
					BeginnerFriendly: true,
				},
			},
			FeatureTraces: []domain.FeatureTrace{
				{
					ID:          "trace-init",
					Name:        "Request Lifecycle",
					Description: "HTTP request flow from router middleware to response.",
					Steps: []domain.FeatureTraceStep{
						{
							Step:        1,
							Title:       "HTTP Request Intake",
							Subsystem:   "CLI / Entrypoint",
							File:        "cmd/server/main.go",
							Line:        25,
							Description: "Chi router receives inbound request and applies middleware.",
							CodeSnippet: "r.Use(middleware.RequestID)\nr.Use(middleware.StructuredLogger(logger))",
						},
					},
				},
			},
			ContributorGuide: domain.ContributorGuide{
				StepsToStart: []string{
					"Clone the repository locally: git clone https://github.com/" + fullName + ".git",
					"Run test suite: go test ./...",
					"Submit a PR referencing an existing issue",
				},
				Prerequisites: []string{
					"Go 1.25 or higher",
					"Git CLI",
				},
				BeginnerFiles: []domain.BeginnerFile{
					{
						Path:   "README.md",
						Reason: "Project orientation and contribution setup instructions.",
					},
				},
			},
			FileTree: []domain.RepoTreeItem{
				{
					Name: "cmd",
					Path: "cmd",
					Type: "directory",
					Children: []domain.RepoTreeItem{
						{Name: "main.go", Path: "cmd/main.go", Type: "file", Size: "2KB"},
					},
				},
				{
					Name: "README.md",
					Path: "README.md",
					Type: "file",
					Size: "1.2KB",
				},
			},
			ChatHistory: []domain.ChatMessage{
				{
					Sender:  "user",
					Message: fmt.Sprintf("What is the architecture of %s?", fullName),
				},
				{
					Sender:  "gitwise",
					Message: fmt.Sprintf("%s is structured as a Go service with CLI entrypoints in cmd/ and internal domain handlers.", fullName),
					Citations: []domain.ChatCitation{
						{File: "cmd/server/main.go", Line: intPtr(1), Snippet: strPtr("package main")},
					},
				},
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model)
}

func (h *RepoHandler) initMockRepos() {
	// 1. vercel/next.js
	h.repositories["vercel/next.js"] = domain.RepoModel{
		ID:              "nextjs",
		Name:            "vercel/next.js",
		Owner:           "vercel",
		Branch:          "canary",
		Stars:           124000,
		Forks:           26100,
		IndexedFiles:    2840,
		PrimaryLanguage: "TypeScript",
		Description:     "The React Framework for the Web: App Router, Turbopack, and hybrid SSR.",
		Subsystems: []domain.SubsystemNode{
			{
				ID:               "app-router",
				Name:             "App Router Server Runtime",
				FileCount:        312,
				EntryPoint:       "packages/next/src/server/app-render/app-render.tsx",
				Description:      "Renders server components, manages action dispatch, and coordinates React Flight stream.",
				Language:         "TypeScript",
				Connections:      []string{"turbopack", "ssr-engine"},
				BeginnerFriendly: false,
			},
			{
				ID:               "turbopack",
				Name:             "Turbopack Compiler",
				FileCount:        489,
				EntryPoint:       "crates/turbopack/src/lib.rs",
				Description:      "Incremental Rust-based bundler and module graph evaluator for ultra-fast HMR.",
				Language:         "Rust",
				Connections:      []string{"ssr-engine"},
				BeginnerFriendly: false,
			},
			{
				ID:               "ssr-engine",
				Name:             "Streaming SSR Engine",
				FileCount:        198,
				EntryPoint:       "packages/next/src/server/web/spec-extension/adapters/next-request.ts",
				Description:      "Web standard request/response adapters and streaming HTML transformer.",
				Language:         "TypeScript",
				Connections:      []string{"client-runtime"},
				BeginnerFriendly: true,
			},
			{
				ID:               "client-runtime",
				Name:             "Client Islands & Hydration",
				FileCount:        145,
				EntryPoint:       "packages/next/src/client/app-index.tsx",
				Description:      "Browser entrypoint reconciling Flight payload and hydrating interactive components.",
				Language:         "TypeScript",
				Connections:      []string{},
				BeginnerFriendly: true,
			},
		},
		FeatureTraces: []domain.FeatureTrace{
			{
				ID:          "server-actions",
				Name:        "Server Actions Execution",
				Description: "End-to-end trace from client form submission to server mutation and revalidation.",
				Steps: []domain.FeatureTraceStep{
					{
						Step:        1,
						Title:       "Client Form Action Trigger",
						Subsystem:   "Client Islands & Hydration",
						File:        "packages/next/src/client/components/form.tsx",
						Line:        48,
						Description: "User submits form; client interceptor wraps formData into multipart Flight payload with actionId header.",
						CodeSnippet: "const res = await fetch(actionUrl, {\n  method: 'POST',\n  headers: { 'Next-Action': actionId },\n  body: formData\n})",
					},
					{
						Step:        2,
						Title:       "HTTP Request Routing",
						Subsystem:   "Streaming SSR Engine",
						File:        "packages/next/src/server/web/spec-extension/adapters/next-request.ts",
						Line:        112,
						Description: "NextRequest parses 'Next-Action' header and forwards to server action handler queue.",
						CodeSnippet: "if (req.headers.has('next-action')) {\n  return handleServerAction(req, actionId);\n}",
					},
					{
						Step:        3,
						Title:       "Server Component Mutation & Execution",
						Subsystem:   "App Router Server Runtime",
						File:        "packages/next/src/server/app-render/action-handler.ts",
						Line:        89,
						Description: "Server action runs within AsyncLocalStorage context, mutates state, and marks route segment paths for revalidation.",
						CodeSnippet: "const result = await actionFn.apply(null, boundArgs);\nrevalidateSegmentTree(segmentPath);",
					},
					{
						Step:        4,
						Title:       "Flight Payload Streaming Response",
						Subsystem:   "App Router Server Runtime",
						File:        "packages/next/src/server/app-render/app-render.tsx",
						Line:        340,
						Description: "Updated React Server Component tree streamed back as text/x-component Flight payload for browser reconciliation.",
						CodeSnippet: "const stream = renderToReadableStream(rscElement, {\n  onError: handleFlightError\n});",
					},
				},
			},
		},
		ContributorGuide: domain.ContributorGuide{
			StepsToStart: []string{
				"Fork and clone vercel/next.js repository",
				"Ensure Node.js 18.17+ and pnpm v9 are installed",
				"Run `pnpm install` at the repository root",
				"Build packages: `pnpm run build`",
				"Run unit tests: `pnpm test-unit`",
				"Pick an issue tagged `good first issue` in the Next.js tracker",
			},
			Prerequisites: []string{
				"TypeScript 5.x",
				"React 19 Server Components paradigm",
				"Turbopack / Rust toolchain (optional for TS-only PRs)",
			},
			BeginnerFiles: []domain.BeginnerFile{
				{
					Path:   "packages/next/src/client/components/react-dev-overlay",
					Reason: "Self-contained UI components with rich test suite; great for learning React error handling",
				},
				{
					Path:   "packages/next/src/lib/constants.ts",
					Reason: "Central location of framework constants, route conventions, and headers",
				},
			},
		},
		FileTree: []domain.RepoTreeItem{
			{
				Name: "packages",
				Path: "packages",
				Type: "directory",
				Children: []domain.RepoTreeItem{
					{
						Name: "next",
						Path: "packages/next",
						Type: "directory",
						Children: []domain.RepoTreeItem{
							{
								Name: "src",
								Path: "packages/next/src",
								Type: "directory",
								Children: []domain.RepoTreeItem{
									{
										Name: "server",
										Path: "packages/next/src/server",
										Type: "directory",
										Children: []domain.RepoTreeItem{
											{
												Name: "app-render",
												Path: "packages/next/src/server/app-render",
												Type: "directory",
												Children: []domain.RepoTreeItem{
													{Name: "app-render.tsx", Path: "packages/next/src/server/app-render/app-render.tsx", Type: "file", Size: "48KB", Owner: "next.js"},
													{Name: "action-handler.ts", Path: "packages/next/src/server/app-render/action-handler.ts", Type: "file", Size: "22KB", Owner: "next.js"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		ChatHistory: []domain.ChatMessage{
			{
				Sender:  "user",
				Message: "Where is the React Server Component Flight stream rendered on the server?",
			},
			{
				Sender:  "gitwise",
				Message: "The Flight stream serialization is coordinated in `packages/next/src/server/app-render/app-render.tsx` inside `createFlightStream()`.",
				Citations: []domain.ChatCitation{
					{File: "packages/next/src/server/app-render/app-render.tsx", Line: intPtr(340), Snippet: strPtr("const stream = renderToReadableStream(rscElement, ...)")},
				},
			},
		},
	}

	// 2. facebook/react
	h.repositories["facebook/react"] = domain.RepoModel{
		ID:              "react",
		Name:            "facebook/react",
		Owner:           "facebook",
		Branch:          "main",
		Stars:           228000,
		Forks:           46200,
		IndexedFiles:    1420,
		PrimaryLanguage: "JavaScript",
		Description:     "The library for web and native user interfaces.",
		Subsystems: []domain.SubsystemNode{
			{
				ID:               "reconciler",
				Name:             "Fiber Reconciler",
				FileCount:        184,
				EntryPoint:       "packages/react-reconciler/src/ReactFiberWorkLoop.js",
				Description:      "Concurrent scheduler, priority lane assignment, and double-buffering DOM diffing.",
				Language:         "JavaScript",
				Connections:      []string{"hooks-runtime", "dom-renderer"},
				BeginnerFriendly: false,
			},
			{
				ID:               "hooks-runtime",
				Name:             "Hooks State Runtime",
				FileCount:        62,
				EntryPoint:       "packages/react-reconciler/src/ReactFiberHooks.js",
				Description:      "Manages state hook linked lists, effect dispatch queues, and memoized values.",
				Language:         "JavaScript",
				Connections:      []string{"reconciler"},
				BeginnerFriendly: true,
			},
			{
				ID:               "dom-renderer",
				Name:             "ReactDOM Host Environment",
				FileCount:        94,
				EntryPoint:       "packages/react-dom/src/client/ReactDOM.js",
				Description:      "Event plugin delegator and host mutations on actual DOM elements.",
				Language:         "JavaScript",
				Connections:      []string{},
				BeginnerFriendly: true,
			},
		},
		FeatureTraces: []domain.FeatureTrace{
			{
				ID:          "state-update",
				Name:        "State Update & Fiber Reconciliation",
				Description: "Trace from setState() call to priority lane scheduling and DOM commit.",
				Steps: []domain.FeatureTraceStep{
					{
						Step:        1,
						Title:       "dispatchAction Invocation",
						Subsystem:   "Hooks State Runtime",
						File:        "packages/react-reconciler/src/ReactFiberHooks.js",
						Line:        2314,
						Description: "User calls setState(x); React enqueues update in fiber hook's queue.",
						CodeSnippet: "function dispatchSetState(fiber, queue, action) {\n  const lane = requestUpdateLane(fiber);\n  const update = { lane, action, next: null };\n  enqueueUpdate(fiber, queue, update, lane);\n  scheduleUpdateOnFiber(fiber, lane);\n}",
					},
				},
			},
		},
		ContributorGuide: domain.ContributorGuide{
			StepsToStart: []string{
				"Fork and clone facebook/react",
				"Install yarn: yarn install",
				"Run compiler and unit tests: yarn test",
			},
			Prerequisites: []string{
				"JavaScript ESNext / Flow / TypeScript",
				"Understanding of Fiber reconciliation tree",
			},
			BeginnerFiles: []domain.BeginnerFile{
				{
					Path:   "packages/react-dom/src/shared/void-elements.js",
					Reason: "Small dictionary maps of HTML void tags with simple unit tests",
				},
			},
		},
		FileTree: []domain.RepoTreeItem{
			{
				Name: "packages",
				Path: "packages",
				Type: "directory",
				Children: []domain.RepoTreeItem{
					{
						Name: "react-reconciler",
						Path: "packages/react-reconciler",
						Type: "directory",
					},
				},
			},
		},
		ChatHistory: []domain.ChatMessage{
			{
				Sender:  "user",
				Message: "Where does React keep track of component hook states between renders?",
			},
			{
				Sender:  "gitwise",
				Message: "React tracks hook state as a singly-linked list on `fiber.memoizedState`.",
				Citations: []domain.ChatCitation{
					{File: "packages/react-reconciler/src/ReactFiberHooks.js", Line: intPtr(412), Snippet: strPtr("let workInProgressHook: Hook | null = null;")},
				},
			},
		},
	}

	// 3. kubernetes/kubernetes
	h.repositories["kubernetes/kubernetes"] = domain.RepoModel{
		ID:              "kubernetes",
		Name:            "kubernetes/kubernetes",
		Owner:           "kubernetes",
		Branch:          "master",
		Stars:           109000,
		Forks:           39100,
		IndexedFiles:    6400,
		PrimaryLanguage: "Go",
		Description:     "Production-Grade Container Scheduling and Automated Workload Orchestration.",
		Subsystems: []domain.SubsystemNode{
			{
				ID:               "kube-scheduler",
				Name:             "kube-scheduler",
				FileCount:        210,
				EntryPoint:       "pkg/scheduler/scheduler.go",
				Description:      "Assigns unassigned Pods to Nodes using filter and score plugin pipelines.",
				Language:         "Go",
				Connections:      []string{"kube-apiserver"},
				BeginnerFriendly: false,
			},
			{
				ID:               "kubelet",
				Name:             "kubelet node agent",
				FileCount:        480,
				EntryPoint:       "pkg/kubelet/kubelet.go",
				Description:      "Primary node agent registering Pod specs with container runtime over CRI.",
				Language:         "Go",
				Connections:      []string{"kube-apiserver"},
				BeginnerFriendly: false,
			},
		},
		FeatureTraces: []domain.FeatureTrace{
			{
				ID:          "pod-scheduling",
				Name:        "Pod Scheduling Pipeline",
				Description: "Filter and score evaluation for placing an unassigned Pod.",
				Steps: []domain.FeatureTraceStep{
					{
						Step:        1,
						Title:       "ScheduleOne Loop",
						Subsystem:   "kube-scheduler",
						File:        "pkg/scheduler/schedule_one.go",
						Line:        82,
						Description: "Scheduler pops Pod from priority queue and executes filter plugins.",
						CodeSnippet: "pod := sched.NextPod()\nsuggestedHost, err := sched.schedulePod(ctx, fwk, state, pod)",
					},
				},
			},
		},
		ContributorGuide: domain.ContributorGuide{
			StepsToStart: []string{
				"Fork and clone kubernetes/kubernetes",
				"Ensure Go 1.25 is installed with make",
				"Run verify check: make verify",
				"Run package unit tests: go test ./pkg/scheduler/...",
			},
			Prerequisites: []string{
				"Go concurrency patterns and interfaces",
				"Kubernetes API primitives and Informers",
			},
			BeginnerFiles: []domain.BeginnerFile{
				{
					Path:   "pkg/scheduler/framework/plugins/names/names.go",
					Reason: "Constant declarations and registration mappings for scheduler plugins",
				},
			},
		},
		FileTree: []domain.RepoTreeItem{
			{
				Name: "pkg",
				Path: "pkg",
				Type: "directory",
				Children: []domain.RepoTreeItem{
					{Name: "scheduler/schedule_one.go", Path: "pkg/scheduler/schedule_one.go", Type: "file", Size: "54KB", Owner: "kubernetes"},
				},
			},
		},
		ChatHistory: []domain.ChatMessage{
			{
				Sender:  "user",
				Message: "Where does the Kubelet communicate with container runtimes like containerd?",
			},
			{
				Sender:  "gitwise",
				Message: "The Kubelet interacts with container runtimes exclusively via CRI gRPC in `pkg/kubelet/cri/remote/remote_runtime.go`.",
				Citations: []domain.ChatCitation{
					{File: "pkg/kubelet/cri/remote/remote_runtime.go", Line: intPtr(55), Snippet: strPtr("type RemoteRuntimeService struct { runtimeClient runtimeapi.RuntimeServiceClient }")},
				},
			},
		},
	}
}

func strPtr(s string) *string {
	return &s
}
