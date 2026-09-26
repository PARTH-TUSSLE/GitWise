package analysis

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/gitwise/backend/internal/domain"
)

// BuildRepoTree constructs a sorted, hierarchical directory tree with symbol counts from flat repository files.
func BuildRepoTree(files []domain.RepositoryFile, symbols []domain.CodeSymbol) []domain.RepoTreeItem {
	// Map symbols count by file path
	symCounts := make(map[string]int)
	for _, s := range symbols {
		symCounts[s.FilePath]++
	}

	type node struct {
		name        string
		path        string
		isDir       bool
		sizeBytes   int
		language    string
		symbolCount int
		children    map[string]*node
	}

	root := &node{
		name:     "",
		path:     "",
		isDir:    true,
		children: make(map[string]*node),
	}

	for _, f := range files {
		clean := path.Clean(strings.ReplaceAll(f.Path, "\\", "/"))
		parts := strings.Split(clean, "/")

		curr := root
		for i, part := range parts {
			isLast := (i == len(parts)-1)
			childPath := strings.Join(parts[:i+1], "/")

			if isLast {
				curr.children[part] = &node{
					name:        part,
					path:        childPath,
					isDir:       false,
					sizeBytes:   f.SizeBytes,
					language:    f.Language,
					symbolCount: symCounts[clean],
					children:    nil,
				}
			} else {
				if _, exists := curr.children[part]; !exists {
					curr.children[part] = &node{
						name:     part,
						path:     childPath,
						isDir:    true,
						children: make(map[string]*node),
					}
				}
				curr = curr.children[part]
			}
		}
	}

	var convert func(n *node) domain.RepoTreeItem
	convert = func(n *node) domain.RepoTreeItem {
		if !n.isDir {
			return domain.RepoTreeItem{
				Name:        n.name,
				Path:        n.path,
				Type:        "file",
				Size:        formatBytes(n.sizeBytes),
				Language:    n.language,
				SymbolCount: n.symbolCount,
			}
		}

		item := domain.RepoTreeItem{
			Name:     n.name,
			Path:     n.path,
			Type:     "directory",
			Children: make([]domain.RepoTreeItem, 0, len(n.children)),
		}

		totalDirSymbols := 0
		var childNodes []*node
		for _, c := range n.children {
			childNodes = append(childNodes, c)
		}

		// Sort directories first, then alphabetical by name
		sort.Slice(childNodes, func(i, j int) bool {
			if childNodes[i].isDir != childNodes[j].isDir {
				return childNodes[i].isDir // true (dir) comes before false (file)
			}
			return childNodes[i].name < childNodes[j].name
		})

		for _, c := range childNodes {
			convertedChild := convert(c)
			totalDirSymbols += convertedChild.SymbolCount
			item.Children = append(item.Children, convertedChild)
		}

		item.SymbolCount = totalDirSymbols
		return item
	}

	var topLevel []domain.RepoTreeItem
	var topNodes []*node
	for _, c := range root.children {
		topNodes = append(topNodes, c)
	}

	sort.Slice(topNodes, func(i, j int) bool {
		if topNodes[i].isDir != topNodes[j].isDir {
			return topNodes[i].isDir
		}
		return topNodes[i].name < topNodes[j].name
	})

	for _, n := range topNodes {
		topLevel = append(topLevel, convert(n))
	}

	return topLevel
}

func formatBytes(b int) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	} else if b < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(b)/1024.0)
	}
	return fmt.Sprintf("%.1f MB", float64(b)/(1024.0*1024.0))
}
