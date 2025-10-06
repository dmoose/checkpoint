package changelog

import (
	"fmt"
	"os"
	"strings"

	"go-llm/internal/schema"

	"gopkg.in/yaml.v3"
)

// AppendEntry appends the rendered YAML document to the changelog (append-only)
func AppendEntry(path, doc string) error {
	if len(doc) == 0 {
		return fmt.Errorf("empty document")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open changelog: %w", err)
	}
	defer f.Close()
	// Ensure trailing newline
	if doc[len(doc)-1] != '\n' {
		doc += "\n"
	}
	if _, err := f.WriteString(doc); err != nil {
		return fmt.Errorf("append changelog: %w", err)
	}
	return nil
}

// UpdateLastDocument preserves append-only structure by only modifying the last document
func UpdateLastDocument(path string, fn func(*schema.CheckpointEntry) *schema.CheckpointEntry) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read changelog: %w", err)
	}

	// Split content by document separators while preserving them
	contentStr := string(content)
	docs := strings.Split(contentStr, "\n---\n")
	if len(docs) < 2 {
		// Try alternative separator (file might start with ---)
		docs = strings.Split(contentStr, "---\n")
		if len(docs) < 2 {
			return fmt.Errorf("no YAML documents found")
		}
	}

	// Find the last non-empty document
	var lastDocIndex = -1
	for i := len(docs) - 1; i >= 0; i-- {
		if strings.TrimSpace(docs[i]) != "" {
			lastDocIndex = i
			break
		}
	}

	if lastDocIndex == -1 {
		return fmt.Errorf("no non-empty documents found")
	}

	// Parse only the last document
	lastDoc := docs[lastDocIndex]
	// Remove leading --- if present
	lastDoc = strings.TrimPrefix(lastDoc, "---\n")

	var entry schema.CheckpointEntry
	if err := yaml.Unmarshal([]byte(lastDoc), &entry); err != nil {
		return fmt.Errorf("decode last document: %w", err)
	}

	// Apply the function to update the entry
	updatedEntry := fn(&entry)

	// Render only the updated last document
	updatedDoc, err := schema.RenderChangelogDocument(updatedEntry)
	if err != nil {
		return fmt.Errorf("render updated document: %w", err)
	}

	// Remove the leading --- from rendered doc since we'll add it back
	updatedDoc = strings.TrimPrefix(updatedDoc, "---\n")

	// Replace the last document in the original content
	docs[lastDocIndex] = updatedDoc

	// Reconstruct the file content
	var result strings.Builder
	for i, doc := range docs {
		if i == 0 {
			// First document might not have leading ---
			if strings.HasPrefix(contentStr, "---\n") {
				result.WriteString("---\n")
			}
		} else {
			result.WriteString("\n---\n")
		}
		result.WriteString(doc)
	}

	// Ensure file ends with newline
	resultStr := result.String()
	if !strings.HasSuffix(resultStr, "\n") {
		resultStr += "\n"
	}

	// Write atomically
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(resultStr), 0644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename tmp: %w", err)
	}
	return nil
}
