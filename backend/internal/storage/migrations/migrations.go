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
