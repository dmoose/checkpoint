package changelog

import (
	"fmt"
	"io"
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

// UpdateLastDocument reads all YAML documents, applies fn to the last, and rewrites the file
func UpdateLastDocument(path string, fn func(*schema.CheckpointEntry) *schema.CheckpointEntry) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read changelog: %w", err)
	}
	dec := yaml.NewDecoder(strings.NewReader(string(content)))
	var entries []*schema.CheckpointEntry
	for {
		var e schema.CheckpointEntry
		if err := dec.Decode(&e); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("decode changelog: %w", err)
		}
		// Skip zero docs
		if e.SchemaVersion == "" && e.Timestamp == "" && len(e.Changes) == 0 && len(e.NextSteps) == 0 {
			continue
		}
		entries = append(entries, &e)
	}
	if len(entries) == 0 {
		return fmt.Errorf("no documents to update")
	}
	entries[len(entries)-1] = fn(entries[len(entries)-1])

	// Re-encode all docs
	var b strings.Builder
	for _, e := range entries {
		out, err := schema.RenderChangelogDocument(e)
		if err != nil {
			return fmt.Errorf("render doc: %w", err)
		}
		b.WriteString(out)
		if !strings.HasSuffix(out, "\n") {
			b.WriteString("\n")
		}
	}

	// Write atomically
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename tmp: %w", err)
	}
	return nil
}
