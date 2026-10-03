package pr

import (
	"bufio"
	"path/filepath"
	"strings"

	"github.com/gitwise/backend/internal/domain"
)

// ParseUnifiedDiff parses unified diff text into a slice of PRFileDiff models.
func ParseUnifiedDiff(rawDiff string) []domain.PRFileDiff {
	var fileDiffs []domain.PRFileDiff
	if strings.TrimSpace(rawDiff) == "" {
		return fileDiffs
	}

	scanner := bufio.NewScanner(strings.NewReader(rawDiff))
	var currentFile *domain.PRFileDiff
	var beforeLines []string
	var afterLines []string

	finalizeCurrent := func() {
		if currentFile != nil {
			currentFile.DiffSnippet = domain.DiffSnippet{
				Target: currentFile.Path,
				Before: strings.Join(beforeLines, "\n"),
				After:  strings.Join(afterLines, "\n"),
			}
			if currentFile.BehavioralSummary == "" {
				currentFile.BehavioralSummary = generateBehavioralSummary(currentFile)
			}
			fileDiffs = append(fileDiffs, *currentFile)
			currentFile = nil
			beforeLines = nil
			afterLines = nil
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		// New file header
		if strings.HasPrefix(line, "diff --git ") {
			finalizeCurrent()
			parts := strings.Fields(line)
			filePath := "unknown"
			if len(parts) >= 4 {
				filePath = strings.TrimPrefix(parts[3], "b/")
			}
			currentFile = &domain.PRFileDiff{
				Path:      filePath,
				Status:    "modified",
				Subsystem: classifySubsystem(filePath),
			}
			continue
		}

		if currentFile == nil {
			continue
		}

		if strings.HasPrefix(line, "new file mode") {
			currentFile.Status = "added"
		} else if strings.HasPrefix(line, "deleted file mode") {
			currentFile.Status = "deleted"
		} else if strings.HasPrefix(line, "--- /dev/null") {
			currentFile.Status = "added"
		} else if strings.HasPrefix(line, "+++ /dev/null") {
			currentFile.Status = "deleted"
		} else if strings.HasPrefix(line, "+++ b/") {
			currentFile.Path = strings.TrimPrefix(line, "+++ b/")
			currentFile.Subsystem = classifySubsystem(currentFile.Path)
		} else if strings.HasPrefix(line, "@@") {
			// Hunk header
			continue
		} else if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			currentFile.Additions++
			afterLines = append(afterLines, strings.TrimPrefix(line, "+"))
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			currentFile.Deletions++
			beforeLines = append(beforeLines, strings.TrimPrefix(line, "-"))
		}
	}

	finalizeCurrent()
	return fileDiffs
}

func generateBehavioralSummary(f *domain.PRFileDiff) string {
	base := filepath.Base(f.Path)
	if f.Status == "added" {
		return "Introduces new module " + base + " with corresponding types and interfaces."
	}
	if f.Status == "deleted" {
		return "Deprecates and removes redundant module " + base + "."
	}
	if strings.Contains(f.Path, "test") || strings.Contains(f.Path, "_test.go") {
		return "Extends regression assertions and mock harness in " + base + "."
	}
	return "Refactors internal execution logic and boundary constraints in " + base + "."
}

func classifySubsystem(path string) string {
	clean := filepath.ToSlash(path)
	parts := strings.Split(clean, "/")
	if len(parts) > 1 {
		return strings.ToUpper(parts[0][:1]) + parts[0][1:]
	}
	return "Core"
}
