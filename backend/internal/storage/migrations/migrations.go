package migrations

import _ "embed"

// InitSchemaUp contains the Phase 1 schema creation SQL.
//
//go:embed 000001_init_schema.up.sql
var InitSchemaUp string

// InitSchemaDown contains the Phase 1 schema rollback SQL.
//
//go:embed 000001_init_schema.down.sql
var InitSchemaDown string

// GithubGitstatUp contains the Phase 2 schema creation SQL.
//
//go:embed 000002_github_gitstat.up.sql
var GithubGitstatUp string

// GithubGitstatDown contains the Phase 2 schema rollback SQL.
//
//go:embed 000002_github_gitstat.down.sql
var GithubGitstatDown string

// RepositoryFilesUp contains the Phase 3 repository_files schema creation SQL.
//
//go:embed 000003_repository_files.up.sql
var RepositoryFilesUp string

// RepositoryFilesDown contains the Phase 3 repository_files schema rollback SQL.
//
//go:embed 000003_repository_files.down.sql
var RepositoryFilesDown string

// ActiveJobsIndexUp contains the Phase 3 active jobs partial unique index creation SQL.
//
//go:embed 000004_active_jobs_partial_index.up.sql
var ActiveJobsIndexUp string

// ActiveJobsIndexDown contains the Phase 3 active jobs partial unique index rollback SQL.
//
//go:embed 000004_active_jobs_partial_index.down.sql
var ActiveJobsIndexDown string

// CodeSymbolsUp contains the Phase 4 code_symbols schema creation SQL.
//
//go:embed 000005_code_symbols.up.sql
var CodeSymbolsUp string

// CodeSymbolsDown contains the Phase 4 code_symbols schema rollback SQL.
//
//go:embed 000005_code_symbols.down.sql
var CodeSymbolsDown string

// DependencyEdgesUp contains the Phase 5 dependency_edges schema creation SQL.
//
//go:embed 000006_dependency_edges.up.sql
var DependencyEdgesUp string

// DependencyEdgesDown contains the Phase 5 dependency_edges schema rollback SQL.
//
//go:embed 000006_dependency_edges.down.sql
var DependencyEdgesDown string
