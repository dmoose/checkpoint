package changelog

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-llm/internal/language"
	"go-llm/internal/schema"

	"github.com/oklog/ulid/v2"
	"gopkg.in/yaml.v3"
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

// MetaDocument represents the initial metadata document in the changelog
type MetaDocument struct {
	SchemaVersion string              `yaml:"schema_version"`
	DocumentType  string              `yaml:"document_type"` // "meta"
	ProjectID     string              `yaml:"project_id"`
	PathHash      string              `yaml:"path_hash"`
	CreatedAt     string              `yaml:"created_at"`
	ToolVersion   string              `yaml:"tool_version"`
	Languages     []language.Language `yaml:"languages,omitempty"`
}

// InitializeChangelog creates the changelog file with a meta document if it doesn't exist
func InitializeChangelog(changelogPath, toolVersion string) error {
	// Check if file already exists
	if _, err := os.Stat(changelogPath); err == nil {
		// File exists, check if it has a meta document
		content, err := os.ReadFile(changelogPath)
		if err != nil {
			return fmt.Errorf("read existing changelog: %w", err)
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "document_type: meta") {
			// Meta document already exists, nothing to do
			return nil
		}

		// File exists but no meta document, prepend one
		return prependMetaDocument(changelogPath, toolVersion, contentStr)
	}

	// File doesn't exist, create it with meta document
	return createChangelogWithMeta(changelogPath, toolVersion)
}

// createChangelogWithMeta creates a new changelog file with just the meta document
func createChangelogWithMeta(changelogPath, toolVersion string) error {
	meta, err := createMetaDocument(changelogPath, toolVersion)
	if err != nil {
		return fmt.Errorf("create meta document: %w", err)
	}

	metaYAML, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal meta document: %w", err)
	}

	content := "---\n" + string(metaYAML)
	return os.WriteFile(changelogPath, []byte(content), 0644)
}

// prependMetaDocument adds a meta document to the beginning of an existing file
func prependMetaDocument(changelogPath, toolVersion, existingContent string) error {
	meta, err := createMetaDocument(changelogPath, toolVersion)
	if err != nil {
		return fmt.Errorf("create meta document: %w", err)
	}

	metaYAML, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal meta document: %w", err)
	}

	newContent := "---\n" + string(metaYAML) + existingContent
	return os.WriteFile(changelogPath, []byte(newContent), 0644)
}

// createMetaDocument creates a new meta document with project metadata
func createMetaDocument(changelogPath, toolVersion string) (*MetaDocument, error) {
	// Get absolute path of the working directory
	workDir := filepath.Dir(changelogPath)
	absPath, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("get absolute path: %w", err)
	}

	// Generate project ID (ULID)
	projectID := ulid.Make().String()

	// Generate path hash (truncated SHA256 of absolute path)
	hasher := sha256.New()
	hasher.Write([]byte(absPath))
	pathHash := fmt.Sprintf("%x", hasher.Sum(nil))[:16] // First 16 chars

	// Detect project languages
	languages, _ := language.DetectLanguages(workDir) // tolerate errors in language detection

	return &MetaDocument{
		SchemaVersion: "1",
		DocumentType:  "meta",
		ProjectID:     projectID,
		PathHash:      pathHash,
		CreatedAt:     time.Now().Format(time.RFC3339),
		ToolVersion:   toolVersion,
		Languages:     languages,
	}, nil
}

// UpdateLastDocument finds the last commit_hash line and updates it
func UpdateLastDocument(path string, fn func(*schema.CheckpointEntry) *schema.CheckpointEntry) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read changelog: %w", err)
	}

	contentStr := string(content)

	// Parse to get the hash value from the function
	lastSepIndex := strings.LastIndex(contentStr, "\n---\n")
	var lastDocContent string
	if lastSepIndex != -1 {
		lastDocContent = contentStr[lastSepIndex+5:]
	} else if strings.HasPrefix(contentStr, "---\n") {
		lastDocContent = contentStr[4:]
	} else {
		return fmt.Errorf("no YAML document separators found")
	}

	var entry schema.CheckpointEntry
	if err := yaml.Unmarshal([]byte(lastDocContent), &entry); err != nil {
		return fmt.Errorf("decode last document: %w", err)
	}

	updatedEntry := fn(&entry)
	if updatedEntry.CommitHash == "" {
		return nil
	}

	// Find the last occurrence of "commit_hash:" in the file
	lastCommitHashIndex := strings.LastIndex(contentStr, "commit_hash:")
	if lastCommitHashIndex == -1 {
		return fmt.Errorf("no commit_hash field found in changelog")
	}

	// Replace the commit_hash line
	lineEnd := strings.Index(contentStr[lastCommitHashIndex:], "\n")
	if lineEnd == -1 {
		lineEnd = len(contentStr) - lastCommitHashIndex
	}
	newCommitHashLine := "commit_hash: " + updatedEntry.CommitHash
	newContent := contentStr[:lastCommitHashIndex] + newCommitHashLine + contentStr[lastCommitHashIndex+lineEnd:]

	return os.WriteFile(path, []byte(newContent), 0644)
}
