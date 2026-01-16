package guides

// EmbeddedGuides contains all guide content embedded in the binary
var EmbeddedGuides = map[string]GuideInfo{
	"first-time-user": {
		Name:        "first-time-user",
		Description: "Complete walkthrough for first-time users",
		Content:     firstTimeUserGuide,
	},
	"llm-workflow": {
		Name:        "llm-workflow",
		Description: "LLM integration patterns and workflow",
		Content:     llmWorkflowGuide,
	},
	"best-practices": {
		Name:        "best-practices",
		Description: "Best practices for effective checkpoints",
		Content:     bestPracticesGuide,
	},
}

// GuideInfo contains metadata and content for a guide
type GuideInfo struct {
	Name        string
	Description string
	Content     string
}

// GetGuide returns the guide content by name, or empty string if not found
func GetGuide(name string) (GuideInfo, bool) {
	guide, ok := EmbeddedGuides[name]
	return guide, ok
}

// ListGuides returns all available guide names
func ListGuides() []string {
	names := make([]string, 0, len(EmbeddedGuides))
	for name := range EmbeddedGuides {
		names = append(names, name)
	}
	return names
}
