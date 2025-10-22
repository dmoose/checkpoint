package changelog

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"go-llm/internal/schema"
)

// AppendEntry appends the rendered YAML document to the changelog (append-only)
func AppendEntry(path, doc string) error {
	if len(doc) == 0 {
		return fmt.Errorf("empty document")
	}

	// Check if file exists and get size for consistency check
	var originalSize int64
	if info, err := os.Stat(path); err == nil {
		originalSize = info.Size()
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open changelog: %w", err)
	}
	defer f.Close()

	// Verify we're at the expected position
	pos, err := f.Seek(0, 2) // Seek to end
	if err != nil {
		return fmt.Errorf("seek to end: %w", err)
	}
	if pos != originalSize {
		return fmt.Errorf("file size changed during operation (expected %d, got %d)", originalSize, pos)
	}

	// Ensure trailing newline
	if doc[len(doc)-1] != '\n' {
		doc += "\n"
	}
	if _, err := f.WriteString(doc); err != nil {
		return fmt.Errorf("append changelog: %w", err)
	}

	// Sync to ensure data is written
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync changelog: %w", err)
	}

	return nil
}

// UpdateLastDocument preserves append-only structure by only modifying the last document
func UpdateLastDocument(path string, fn func(*schema.CheckpointEntry) *schema.CheckpointEntry) error {
	// Get file info for consistency check
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat changelog: %w", err)
	}
	originalSize := info.Size()
	originalModTime := info.ModTime()

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read changelog: %w", err)
	}

	// Verify file hasn't changed since we started
	if int64(len(content)) != originalSize {
		return fmt.Errorf("file size changed during read (expected %d, got %d)", originalSize, len(content))
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

	// Write atomically with additional consistency checks
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(resultStr), 0644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}

	// Verify the original file hasn't been modified while we were working
	currentInfo, err := os.Stat(path)
	if err != nil {
		os.Remove(tmp) // Clean up
		return fmt.Errorf("stat original file before rename: %w", err)
	}
	if !currentInfo.ModTime().Equal(originalModTime) || currentInfo.Size() != originalSize {
		os.Remove(tmp) // Clean up
		return fmt.Errorf("original file was modified during update operation")
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp) // Clean up on failure
		return fmt.Errorf("rename tmp: %w", err)
	}

	return nil
}
