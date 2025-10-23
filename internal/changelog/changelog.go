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

// UpdateLastDocument modifies the last document by re-rendering it
func UpdateLastDocument(path string, fn func(*schema.CheckpointEntry) *schema.CheckpointEntry) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read changelog: %w", err)
	}

	contentStr := string(content)

	// Find the start of the last document
	lastSepIndex := strings.LastIndex(contentStr, "\n---\n")
	var beforeLastDoc, lastDocContent string

	if lastSepIndex != -1 {
		// There's a separator, so split there
		beforeLastDoc = contentStr[:lastSepIndex+1]  // Keep the newline before ---
		lastDocContent = contentStr[lastSepIndex+5:] // Skip "\n---\n"
	} else if strings.HasPrefix(contentStr, "---\n") {
		// File starts with ---, this is the only/first document
		beforeLastDoc = ""
		lastDocContent = contentStr[4:] // Skip "---\n"
	} else {
		return fmt.Errorf("no YAML document separators found")
	}

	// Parse the last document
	var entry schema.CheckpointEntry
	if err := yaml.Unmarshal([]byte(lastDocContent), &entry); err != nil {
		return fmt.Errorf("decode last document: %w", err)
	}

	// Apply the update function
	updatedEntry := fn(&entry)

	// Re-render the updated document
	updatedDoc, err := schema.RenderChangelogDocument(updatedEntry)
	if err != nil {
		return fmt.Errorf("render updated document: %w", err)
	}

	// Reconstruct the file: everything before + updated last document
	var newContent string
	if beforeLastDoc == "" {
		// This was the only document
		newContent = updatedDoc
	} else {
		// Remove trailing newline from updatedDoc since RenderChangelogDocument includes "---\n"
		newContent = beforeLastDoc + strings.TrimSuffix(updatedDoc, "\n")
	}

	// Ensure file ends with newline
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}

	// Write atomically
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp) // Clean up on failure
		return fmt.Errorf("rename tmp: %w", err)
	}

	return nil
}
