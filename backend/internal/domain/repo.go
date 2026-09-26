package domain

// SubsystemNode represents a distinct functional subsystem within a repository snapshot.
type SubsystemNode struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	FileCount        int      `json:"fileCount"`
	SymbolCount      int      `json:"symbolCount,omitempty"`
	EntryPoint       string   `json:"entryPoint"`
	Description      string   `json:"description"`
	Language         string   `json:"language"`
	Connections      []string `json:"connections"`
	BeginnerFriendly bool     `json:"beginnerFriendly"`
	PathPrefix       string   `json:"pathPrefix,omitempty"`
}

// FeatureTraceStep captures a single execution step in a code path trace.
type FeatureTraceStep struct {
	Step        int    `json:"step"`
	Title       string `json:"title"`
	Subsystem   string `json:"subsystem"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Description string `json:"description"`
	CodeSnippet string `json:"codeSnippet"`
}

// FeatureTrace represents an end-to-end architectural flow through subsystems.
type FeatureTrace struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Steps       []FeatureTraceStep `json:"steps"`
}

// RepoTreeItem models a node in the repository file tree.
type RepoTreeItem struct {
	Name        string         `json:"name"`
	Path        string         `json:"path"`
	Type        string         `json:"type"` // "file" | "directory"
	Size        string         `json:"size,omitempty"`
	Language    string         `json:"language,omitempty"`
	SymbolCount int            `json:"symbolCount,omitempty"`
	Owner       string         `json:"owner,omitempty"`
	Children    []RepoTreeItem `json:"children,omitempty"`
}

// BeginnerFile provides onboarding orientation file pointers.
type BeginnerFile struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// ContributorGuide details onboarding prerequisites and starting files.
type ContributorGuide struct {
	StepsToStart  []string       `json:"stepsToStart"`
	Prerequisites []string       `json:"prerequisites"`
	BeginnerFiles []BeginnerFile `json:"beginnerFiles"`
}

// ChatCitation represents a verified source code reference in chat.
type ChatCitation struct {
	File    string  `json:"file"`
	Line    *int    `json:"line,omitempty"`
	Snippet *string `json:"snippet,omitempty"`
}

// ChatMessage represents a conversation turn in the terminal mentor chat.
type ChatMessage struct {
	Sender    string         `json:"sender"` // "user" | "gitwise"
	Message   string         `json:"message"`
	Citations []ChatCitation `json:"citations,omitempty"`
}

// RepoModel is the complete domain contract consumed by the repository explorer UI.
type RepoModel struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Owner            string           `json:"owner"`
	Branch           string           `json:"branch"`
	Stars            int              `json:"stars"`
	Forks            int              `json:"forks"`
	IndexedFiles     int              `json:"indexedFiles"`
	PrimaryLanguage  string           `json:"primaryLanguage"`
	Description      string           `json:"description"`
	Subsystems       []SubsystemNode  `json:"subsystems"`
	FeatureTraces    []FeatureTrace   `json:"featureTraces"`
	ContributorGuide ContributorGuide `json:"contributorGuide"`
	FileTree         []RepoTreeItem   `json:"fileTree"`
	ChatHistory      []ChatMessage    `json:"chatHistory"`
}
