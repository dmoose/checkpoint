package skills

import (
	"embed"
	"path/filepath"
	"strings"
)

//go:embed content/*.md
var contentFS embed.FS

// DefaultSkills returns the default skills that ship with checkpoint.
// Returns a map of skill name -> content.
func DefaultSkills() map[string]string {
	skills := make(map[string]string)

	entries, err := contentFS.ReadDir("content")
	if err != nil {
		return skills
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		// Convert filename to skill name (remove .md extension)
		name := strings.TrimSuffix(entry.Name(), ".md")

		content, err := contentFS.ReadFile(filepath.Join("content", entry.Name()))
		if err != nil {
			continue
		}

		skills[name] = string(content)
	}

	return skills
}

// GetSkill returns the content of a default skill by name.
func GetSkill(name string) (string, bool) {
	skills := DefaultSkills()
	content, ok := skills[name]
	return content, ok
}

// ListSkills returns the names of all default skills.
func ListSkills() []string {
	skills := DefaultSkills()
	names := make([]string, 0, len(skills))
	for name := range skills {
		names = append(names, name)
	}
	return names
}
