package analysis

import (
	"path"
	"sort"
	"strings"

	"github.com/gitwise/backend/internal/domain"
)

// SubsystemDefinition specifies a recognizable path pattern for architectural classification.
type SubsystemDefinition struct {
	ID          string
	Name        string
	PathPrefix  string
	Description string
	Connections []string
}

// DefaultSubsystems provides path-based heuristics for common repository architectures.
var DefaultSubsystems = []SubsystemDefinition{
	{
		ID:          "cli",
		Name:        "Command Line & Entrypoints",
		PathPrefix:  "cmd/",
		Description: "Application binary entrypoints and command line interfaces.",
		Connections: []string{"core-service", "api-server"},
	},
	{
		ID:          "api-server",
		Name:        "API & Routing Layer",
		PathPrefix:  "api/",
		Description: "HTTP routes, middleware, and request/response handlers.",
		Connections: []string{"core-service", "data-layer"},
	},
	{
		ID:          "core-service",
		Name:        "Core Domain Services",
		PathPrefix:  "internal/",
		Description: "Business logic, domain models, and core system abstractions.",
		Connections: []string{"data-layer"},
	},
	{
		ID:          "public-pkg",
		Name:        "Public Packages & Libraries",
		PathPrefix:  "pkg/",
		Description: "Reusable public utility packages and shared libraries.",
		Connections: []string{"core-service"},
	},
	{
		ID:          "app-routing",
		Name:        "App Router & Pages",
		PathPrefix:  "src/app/",
		Description: "Next.js App Router, layout definitions, and page entrypoints.",
		Connections: []string{"components", "client-lib"},
	},
	{
		ID:          "components",
		Name:        "UI Components",
		PathPrefix:  "src/components/",
		Description: "Reusable visual components, interactive terminals, and UI widgets.",
		Connections: []string{"client-lib"},
	},
	{
		ID:          "client-lib",
		Name:        "Client Utilities & API",
		PathPrefix:  "src/lib/",
		Description: "Client-side state, API wrappers, and data structures.",
		Connections: []string{},
	},
	{
		ID:          "monorepo-pkgs",
		Name:        "Monorepo Packages",
		PathPrefix:  "packages/",
		Description: "Independently versioned workspaces within a monorepo.",
		Connections: []string{},
	},
	{
		ID:          "rust-crates",
		Name:        "Native Rust Crates",
		PathPrefix:  "crates/",
		Description: "Native Rust modules, compilers, or performance-critical extensions.",
		Connections: []string{},
	},
}

// ClassifySubsystems groups files and symbols into explainable architectural subsystems based on path prefixes.
func ClassifySubsystems(files []domain.RepositoryFile, symbols []domain.CodeSymbol) []domain.SubsystemNode {
	subsystemMap := make(map[string]*domain.SubsystemNode)
	subsystemFiles := make(map[string][]domain.RepositoryFile)

	// Map symbols to files
	fileSymbolCount := make(map[string]int)
	for _, sym := range symbols {
		fileSymbolCount[sym.FilePath]++
	}

	for _, file := range files {
		cleanPath := path.Clean(strings.ReplaceAll(file.Path, "\\", "/"))
		matchedDef := findMatchingSubsystem(cleanPath)

		node, exists := subsystemMap[matchedDef.ID]
		if !exists {
			node = &domain.SubsystemNode{
				ID:               matchedDef.ID,
				Name:             matchedDef.Name,
				FileCount:        0,
				SymbolCount:      0,
				Description:      matchedDef.Description,
				Language:         file.Language,
				Connections:      matchedDef.Connections,
				BeginnerFriendly: isBeginnerFriendly(matchedDef.ID),
				PathPrefix:       matchedDef.PathPrefix,
			}
			subsystemMap[matchedDef.ID] = node
		}

		node.FileCount++
		node.SymbolCount += fileSymbolCount[cleanPath]
		subsystemFiles[matchedDef.ID] = append(subsystemFiles[matchedDef.ID], file)
	}

	// Determine entry point and dominant language for each subsystem
	result := make([]domain.SubsystemNode, 0, len(subsystemMap))
	for id, node := range subsystemMap {
		fList := subsystemFiles[id]
		node.EntryPoint = detectEntryPoint(fList)
		node.Language = detectDominantLanguage(fList)
		result = append(result, *node)
	}

	// Deterministic sort: by file count descending, then ID ascending
	sort.Slice(result, func(i, j int) bool {
		if result[i].FileCount != result[j].FileCount {
			return result[i].FileCount > result[j].FileCount
		}
		return result[i].ID < result[j].ID
	})

	return result
}

func findMatchingSubsystem(filePath string) SubsystemDefinition {
	for _, def := range DefaultSubsystems {
		if strings.HasPrefix(filePath, def.PathPrefix) {
			return def
		}
	}

	// Check top-level directory if not in default prefixes
	parts := strings.Split(filePath, "/")
	if len(parts) > 1 {
		firstDir := parts[0] + "/"
		return SubsystemDefinition{
			ID:          "subsystem-" + parts[0],
			Name:        strings.Title(parts[0]) + " Subsystem",
			PathPrefix:  firstDir,
			Description: "Modules and components located in " + firstDir,
			Connections: []string{},
		}
	}

	// Root files
	return SubsystemDefinition{
		ID:          "root",
		Name:        "Root & Configuration",
		PathPrefix:  "",
		Description: "Root configuration files, build scripts, and entrypoints.",
		Connections: []string{},
	}
}

func detectEntryPoint(files []domain.RepositoryFile) string {
	if len(files) == 0 {
		return ""
	}

	priorityNames := []string{
		"main.go", "index.ts", "index.js", "app.ts", "app.go",
		"server.ts", "server.go", "page.tsx", "page.jsx",
		"lib.rs", "mod.rs",
	}

	for _, prio := range priorityNames {
		for _, f := range files {
			if path.Base(f.Path) == prio {
				return f.Path
			}
		}
	}

	return files[0].Path
}

func detectDominantLanguage(files []domain.RepositoryFile) string {
	langCounts := make(map[string]int)
	for _, f := range files {
		if f.Language != "" && f.Language != "Plain Text" {
			langCounts[f.Language]++
		}
	}

	maxCount := -1
	dominant := "Plain Text"
	for lang, count := range langCounts {
		if count > maxCount {
			maxCount = count
			dominant = lang
		}
	}
	return dominant
}

func isBeginnerFriendly(id string) bool {
	switch id {
	case "components", "client-lib", "docs", "public-pkg":
		return true
	default:
		return false
	}
}
