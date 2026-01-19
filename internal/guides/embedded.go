package guides

import (
	"embed"
	"path/filepath"
	"strings"
)

//go:embed content/*.md
var contentFS embed.FS

// guideMetadata maps guide names to their descriptions
var guideMetadata = map[string]string{
	"first-time-user": "Complete walkthrough for first-time users",
	"llm-workflow":    "LLM integration patterns and workflow",
	"best-practices":  "Best practices for effective checkpoints",
}

// GuideInfo contains metadata and content for a guide
type GuideInfo struct {
	Name        string
	Description string
	Content     string
}

// EmbeddedGuides is populated at init time from embedded files
var EmbeddedGuides map[string]GuideInfo

func init() {
	EmbeddedGuides = make(map[string]GuideInfo)

	entries, err := contentFS.ReadDir("content")
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		// Convert filename to guide name (remove .md extension)
		name := strings.TrimSuffix(entry.Name(), ".md")

		content, err := contentFS.ReadFile(filepath.Join("content", entry.Name()))
		if err != nil {
			continue
		}

		description := guideMetadata[name]
		if description == "" {
			description = name // fallback to name if no description
		}

		EmbeddedGuides[name] = GuideInfo{
			Name:        name,
			Description: description,
			Content:     string(content),
		}
	}
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
