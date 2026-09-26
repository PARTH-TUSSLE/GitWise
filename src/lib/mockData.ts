export interface ContributorProfile {
  username: string;
  name: string;
  avatarUrl: string;
  title: string;
  bio: string;
  joined: string;
  status: string;
  primaryLanguages: string[];
  metrics: {
    mergedPRs: number;
    openPRs: number;
    codeReviewsGiven: number;
    reviewCommentVolume: number;
    issuesOpened: number;
    issuesParticipatedIn: number;
    issuesLinkedToMergedPRs: number;
    activeRepositories: number;
    totalCommits: number;
    linesAdded: number;
    linesDeleted: number;
    filesChanged: number;
    reviewTurnaroundHours: number;
    mergeSuccessRatePct: number;
  };
  repositories: Array<{
    name: string;
    description: string;
    stars: number;
    forks: number;
    language: string;
    commits: number;
    prs: number;
    role: "Maintainer" | "Core Contributor" | "External Contributor";
    evidenceUrl: string;
  }>;
  recentDiffs: Array<{
    id: string;
    repo: string;
    prNumber?: number;
    commitHash: string;
    message: string;
    added: number;
    deleted: number;
    timestamp: string;
    type: "PR_MERGED" | "COMMIT" | "REVIEW_COMMENT" | "ISSUE_CLOSED";
  }>;
  activityWeeks: Array<{
    week: string;
    days: Array<{
      date: string;
      level: 0 | 1 | 2 | 3 | 4;
      commits: number;
      prs: number;
      reviews: number;
    }>;
  }>;
}

export const CONTRIBUTORS: Record<string, ContributorProfile> = {
  alexR_dev: {
    username: "alexR_dev",
    name: "Alex Rivera",
    avatarUrl: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=256&q=80",
    title: "Senior Infrastructure & Systems Engineer",
    bio: "Building developer tooling, distributed consensus, and platform infrastructure.",
    joined: "2018",
    status: "Active Contributor",
    primaryLanguages: ["Rust", "TypeScript", "Go", "C++"],
    metrics: {
      mergedPRs: 148,
      openPRs: 6,
      codeReviewsGiven: 291,
      reviewCommentVolume: 842,
      issuesOpened: 63,
      issuesParticipatedIn: 117,
      issuesLinkedToMergedPRs: 102,
      activeRepositories: 18,
      totalCommits: 4112,
      linesAdded: 54200,
      linesDeleted: 31700,
      filesChanged: 1847,
      reviewTurnaroundHours: 4.2,
      mergeSuccessRatePct: 89.5,
    },
    repositories: [
      {
        name: "vercel/next.js",
        description: "The React Framework for the Web with App Router and Turbopack compiler",
        stars: 122000,
        forks: 26400,
        language: "TypeScript",
        commits: 542,
        prs: 38,
        role: "Core Contributor",
        evidenceUrl: "https://github.com/vercel/next.js",
      },
      {
        name: "kubernetes/kubernetes",
        description: "Production-Grade Container Scheduling and Automated Workload Orchestration",
        stars: 108000,
        forks: 39100,
        language: "Go",
        commits: 418,
        prs: 44,
        role: "External Contributor",
        evidenceUrl: "https://github.com/kubernetes/kubernetes",
      },
      {
        name: "tokio-rs/tokio",
        description: "A runtime for writing reliable, asynchronous, and slim applications with Rust",
        stars: 25400,
        forks: 2400,
        language: "Rust",
        commits: 34,
        prs: 12,
        role: "External Contributor",
        evidenceUrl: "https://github.com/tokio-rs/tokio",
      },
      {
        name: "facebook/react",
        description: "The library for web and native user interfaces",
        stars: 228000,
        forks: 46200,
        language: "TypeScript",
        commits: 18,
        prs: 6,
        role: "External Contributor",
        evidenceUrl: "https://github.com/facebook/react",
      },
    ],
    recentDiffs: [
      {
        id: "d1",
        repo: "vercel/next.js",
        prNumber: 62145,
        commitHash: "7b89f0a",
        message: "Merge PR #62145: Add Flight stream backpressure drain controller for server actions",
        added: 1420,
        deleted: 380,
        timestamp: "14:38:05",
        type: "PR_MERGED",
      },
      {
        id: "d2",
        repo: "kubernetes/kubernetes",
        prNumber: 124580,
        commitHash: "c381da2",
        message: "Refactor kube-scheduler pre-filter plugin node score caching to prevent contention",
        added: 840,
        deleted: 612,
        timestamp: "13:12:44",
        type: "COMMIT",
      },
      {
        id: "d3",
        repo: "tokio-rs/tokio",
        prNumber: 5891,
        commitHash: "9a21ef4",
        message: "Review comment: Validate poll_ready backpressure handling in mpsc channel",
        added: 0,
        deleted: 0,
        timestamp: "11:05:19",
        type: "REVIEW_COMMENT",
      },
      {
        id: "d4",
        repo: "vercel/next.js",
        prNumber: 54821,
        commitHash: "3f88be1",
        message: "Merge PR #54821: Fix concurrent server action revalidation race in chunked streaming",
        added: 620,
        deleted: 140,
        timestamp: "09:41:02",
        type: "PR_MERGED",
      },
      {
        id: "d5",
        repo: "kubernetes/kubernetes",
        commitHash: "e102f9c",
        message: "Fix boundary check in chunked kubelet pod status watcher; closes issue #801",
        added: 45,
        deleted: 12,
        timestamp: "Yesterday",
        type: "ISSUE_CLOSED",
      },
    ],
    activityWeeks: Array.from({ length: 52 }, (_, w) => ({
      week: `W${w + 1}`,
      days: Array.from({ length: 7 }, (_, d) => {
        // deterministic distribution creating realistic patterns
        const seed = (w * 7 + d * 13) % 100;
        let level: 0 | 1 | 2 | 3 | 4 = 0;
        if (d === 0 || d === 6) {
          level = seed > 80 ? 1 : 0;
        } else if (seed > 85) {
          level = 4;
        } else if (seed > 65) {
          level = 3;
        } else if (seed > 40) {
          level = 2;
        } else if (seed > 20) {
          level = 1;
        }
        return {
          date: `2026-${String(Math.floor(w / 4.4) + 1).padStart(2, "0")}-${String((d * 4 + 1) % 28 + 1).padStart(2, "0")}`,
          level,
          commits: level * 3,
          prs: level > 2 ? 1 : 0,
          reviews: level > 1 ? 2 : 0,
        };
      }),
    })),
  },
  torvalds: {
    username: "torvalds",
    name: "Linus Torvalds",
    avatarUrl: "https://avatars.githubusercontent.com/u/1024025?v=4",
    title: "Creator of Linux and Git",
    bio: "Software developer. Creator of Linux and Git. Linux kernel maintainer.",
    joined: "2011",
    status: "Kernel Maintainer",
    primaryLanguages: ["C", "Shell", "Makefile"],
    metrics: {
      mergedPRs: 2840,
      openPRs: 14,
      codeReviewsGiven: 14200,
      reviewCommentVolume: 38900,
      issuesOpened: 18,
      issuesParticipatedIn: 4890,
      issuesLinkedToMergedPRs: 2100,
      activeRepositories: 4,
      totalCommits: 31200,
      linesAdded: 890400,
      linesDeleted: 640200,
      filesChanged: 28400,
      reviewTurnaroundHours: 1.8,
      mergeSuccessRatePct: 98.2,
    },
    repositories: [
      {
        name: "torvalds/linux",
        description: "Linux kernel source tree",
        stars: 182000,
        forks: 54000,
        language: "C",
        commits: 29800,
        prs: 2400,
        role: "Maintainer",
        evidenceUrl: "https://github.com/torvalds/linux",
      },
      {
        name: "torvalds/subsurface-for-dirk",
        description: "Subsurface dive log program",
        stars: 1200,
        forks: 230,
        language: "C",
        commits: 1400,
        prs: 440,
        role: "Maintainer",
        evidenceUrl: "https://github.com/torvalds/subsurface-for-dirk",
      },
    ],
    recentDiffs: [
      {
        id: "t1",
        repo: "torvalds/linux",
        commitHash: "9a01f42",
        message: "Linux 6.12-rc7 release tag and merge window synchronization",
        added: 8420,
        deleted: 5120,
        timestamp: "3h ago",
        type: "PR_MERGED",
      },
      {
        id: "t2",
        repo: "torvalds/linux",
        commitHash: "7b411d9",
        message: "Merge branch 'x86/urgent' of git://git.kernel.org/pub/scm/linux/kernel/git/tip/tip",
        added: 210,
        deleted: 85,
        timestamp: "8h ago",
        type: "PR_MERGED",
      },
    ],
    activityWeeks: Array.from({ length: 52 }, (_, w) => ({
      week: `W${w + 1}`,
      days: Array.from({ length: 7 }, (_, d) => ({
        date: `2026-W${w}-${d}`,
        level: (d === 0 || d === 6 ? 2 : 4) as 0 | 1 | 2 | 3 | 4,
        commits: 12,
        prs: 4,
        reviews: 18,
      })),
    })),
  },
  gaearon: {
    username: "gaearon",
    name: "Dan Abramov",
    avatarUrl: "https://avatars.githubusercontent.com/u/810438?v=4",
    title: "Software Engineer & Co-author of Redux",
    bio: "Working on React, Redux, and open-source JavaScript architecture.",
    joined: "2011",
    status: "Core Maintainer",
    primaryLanguages: ["JavaScript", "TypeScript", "CSS"],
    metrics: {
      mergedPRs: 1840,
      openPRs: 22,
      codeReviewsGiven: 3910,
      reviewCommentVolume: 12400,
      issuesOpened: 340,
      issuesParticipatedIn: 2180,
      issuesLinkedToMergedPRs: 1420,
      activeRepositories: 42,
      totalCommits: 8490,
      linesAdded: 210400,
      linesDeleted: 148900,
      filesChanged: 7200,
      reviewTurnaroundHours: 3.4,
      mergeSuccessRatePct: 94.1,
    },
    repositories: [
      {
        name: "facebook/react",
        description: "The library for web and native user interfaces",
        stars: 228000,
        forks: 46200,
        language: "JavaScript",
        commits: 2410,
        prs: 840,
        role: "Core Contributor",
        evidenceUrl: "https://github.com/facebook/react",
      },
      {
        name: "reduxjs/redux",
        description: "Predictable state container for JavaScript apps",
        stars: 60400,
        forks: 15300,
        language: "TypeScript",
        commits: 1100,
        prs: 420,
        role: "Maintainer",
        evidenceUrl: "https://github.com/reduxjs/redux",
      },
    ],
    recentDiffs: [
      {
        id: "g1",
        repo: "facebook/react",
        commitHash: "e441da0",
        message: "Refactor Suspense error boundary reconciliation priority",
        added: 412,
        deleted: 198,
        timestamp: "5h ago",
        type: "PR_MERGED",
      },
    ],
    activityWeeks: Array.from({ length: 52 }, (_, w) => ({
      week: `W${w + 1}`,
      days: Array.from({ length: 7 }, (_, d) => ({
        date: `2026-W${w}-${d}`,
        level: (d === 0 || d === 6 ? 1 : 3) as 0 | 1 | 2 | 3 | 4,
        commits: 6,
        prs: 2,
        reviews: 8,
      })),
    })),
  },
};
