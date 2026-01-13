package changelog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestCreateMetaDocument(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, ".checkpoint-changelog.yaml")

	meta, err := createMetaDocument(changelogPath, "1.0.0")
	if err != nil {
		t.Fatalf("createMetaDocument failed: %v", err)
	}

	// Verify all fields are set
	if meta.SchemaVersion != "1" {
		t.Errorf("expected schema_version '1', got %q", meta.SchemaVersion)
	}
	if meta.DocumentType != "meta" {
		t.Errorf("expected document_type 'meta', got %q", meta.DocumentType)
	}
	if meta.ProjectID == "" {
		t.Error("expected non-empty project_id")
	}
	if len(meta.ProjectID) != 26 { // ULID length
		t.Errorf("expected project_id length 26, got %d", len(meta.ProjectID))
	}
	if meta.PathHash == "" {
		t.Error("expected non-empty path_hash")
	}
	if len(meta.PathHash) != 16 {
		t.Errorf("expected path_hash length 16, got %d", len(meta.PathHash))
	}
	if meta.CreatedAt == "" {
		t.Error("expected non-empty created_at")
	}
	if meta.ToolVersion != "1.0.0" {
		t.Errorf("expected tool_version '1.0.0', got %q", meta.ToolVersion)
	}

	// Verify created_at is valid RFC3339
	if _, err := time.Parse(time.RFC3339, meta.CreatedAt); err != nil {
		t.Errorf("created_at is not valid RFC3339: %v", err)
	}
}

func TestCreateChangelogWithMeta(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, ".checkpoint-changelog.yaml")

	err := createChangelogWithMeta(changelogPath, "1.2.3")
	if err != nil {
		t.Fatalf("createChangelogWithMeta failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		t.Fatal("changelog file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(changelogPath)
	if err != nil {
		t.Fatalf("failed to read changelog: %v", err)
	}

	contentStr := string(content)
	if !strings.HasPrefix(contentStr, "---\n") {
		t.Error("changelog should start with YAML separator")
	}

	// Parse the YAML document
	var meta MetaDocument
	if err := yaml.Unmarshal(content[4:], &meta); err != nil {
		t.Fatalf("failed to parse meta document: %v", err)
	}

	// Verify meta document fields
	if meta.DocumentType != "meta" {
		t.Errorf("expected document_type 'meta', got %q", meta.DocumentType)
	}
	if meta.ToolVersion != "1.2.3" {
		t.Errorf("expected tool_version '1.2.3', got %q", meta.ToolVersion)
	}
}

func TestPrependMetaDocument(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, ".checkpoint-changelog.yaml")

	// Create existing content
	existingContent := `---
schema_version: "1"
timestamp: "2023-01-01T12:00:00Z"
commit_hash: ""
changes:
  - summary: "test change"
    change_type: "feature"
`

	err := prependMetaDocument(changelogPath, "2.0.0", existingContent)
	if err != nil {
		t.Fatalf("prependMetaDocument failed: %v", err)
	}

	// Read and verify content
	content, err := os.ReadFile(changelogPath)
	if err != nil {
		t.Fatalf("failed to read changelog: %v", err)
	}

	contentStr := string(content)

	// Verify meta document is at the beginning
	if !strings.HasPrefix(contentStr, "---\n") {
		t.Error("changelog should start with YAML separator")
	}

	// Verify existing content is preserved
	if !strings.Contains(contentStr, "test change") {
		t.Error("existing content should be preserved")
	}

	// Verify meta document comes before existing content
	metaIndex := strings.Index(contentStr, "document_type: meta")
	testChangeIndex := strings.Index(contentStr, "test change")
	if metaIndex == -1 {
		t.Error("meta document not found")
	}
	if testChangeIndex == -1 {
		t.Error("existing content not found")
	}
	if metaIndex > testChangeIndex {
		t.Error("meta document should come before existing content")
	}
}

func TestInitializeChangelog_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, ".checkpoint-changelog.yaml")

	err := InitializeChangelog(changelogPath, "3.0.0")
	if err != nil {
		t.Fatalf("InitializeChangelog failed: %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(changelogPath)
	if err != nil {
		t.Fatalf("failed to read changelog: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "document_type: meta") {
		t.Error("meta document should be present")
	}
	if !strings.Contains(contentStr, "tool_version: 3.0.0") {
		t.Error("tool version should be set correctly")
	}
}

func TestInitializeChangelog_ExistingFileWithMeta(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, ".checkpoint-changelog.yaml")

	// Create file with existing meta document
	existingContent := `---
schema_version: "1"
document_type: meta
project_id: "01ARZ3NDEKTSV4RRFFQ69G5FAV"
path_hash: "abcdef1234567890"
created_at: "2023-01-01T12:00:00Z"
tool_version: "1.0.0"
---
schema_version: "1"
timestamp: "2023-01-01T13:00:00Z"
commit_hash: ""
changes:
  - summary: "existing change"
    change_type: "feature"
`

	err := os.WriteFile(changelogPath, []byte(existingContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Initialize changelog - should do nothing
	err = InitializeChangelog(changelogPath, "4.0.0")
	if err != nil {
		t.Fatalf("InitializeChangelog failed: %v", err)
	}

	// Verify content is unchanged
	content, err := os.ReadFile(changelogPath)
	if err != nil {
		t.Fatalf("failed to read changelog: %v", err)
	}

	contentStr := string(content)
	if contentStr != existingContent {
		t.Error("existing file with meta should not be modified")
	}
}

func TestInitializeChangelog_ExistingFileWithoutMeta(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, ".checkpoint-changelog.yaml")

	// Create file without meta document
	existingContent := `---
schema_version: "1"
timestamp: "2023-01-01T12:00:00Z"
commit_hash: ""
changes:
  - summary: "existing change"
    change_type: "feature"
`

	err := os.WriteFile(changelogPath, []byte(existingContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Initialize changelog - should prepend meta
	err = InitializeChangelog(changelogPath, "5.0.0")
	if err != nil {
		t.Fatalf("InitializeChangelog failed: %v", err)
	}

	// Verify meta document was added
	content, err := os.ReadFile(changelogPath)
	if err != nil {
		t.Fatalf("failed to read changelog: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "document_type: meta") {
		t.Error("meta document should have been added")
	}
	if !strings.Contains(contentStr, "tool_version: 5.0.0") {
		t.Error("tool version should be set correctly")
	}
	if !strings.Contains(contentStr, "existing change") {
		t.Error("existing content should be preserved")
	}

	// Verify meta comes before existing content
	metaIndex := strings.Index(contentStr, "document_type: meta")
	existingIndex := strings.Index(contentStr, "existing change")
	if metaIndex > existingIndex {
		t.Error("meta document should come before existing content")
	}
}

func TestPathHashConsistency(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath1 := filepath.Join(tmpDir, "project1", ".checkpoint-changelog.yaml")
	changelogPath2 := filepath.Join(tmpDir, "project1", ".checkpoint-changelog.yaml")
	changelogPath3 := filepath.Join(tmpDir, "project2", ".checkpoint-changelog.yaml")

	// Ensure directories exist
	_ = os.MkdirAll(filepath.Dir(changelogPath1), 0755)
	_ = os.MkdirAll(filepath.Dir(changelogPath3), 0755)

	// Create meta documents for same path - should have same hash
	meta1, err := createMetaDocument(changelogPath1, "1.0.0")
	if err != nil {
		t.Fatalf("createMetaDocument failed: %v", err)
	}

	meta2, err := createMetaDocument(changelogPath2, "1.0.0")
	if err != nil {
		t.Fatalf("createMetaDocument failed: %v", err)
	}

	if meta1.PathHash != meta2.PathHash {
		t.Error("same paths should produce same path hash")
	}

	// Create meta document for different path - should have different hash
	meta3, err := createMetaDocument(changelogPath3, "1.0.0")
	if err != nil {
		t.Fatalf("createMetaDocument failed: %v", err)
	}

	if meta1.PathHash == meta3.PathHash {
		t.Error("different paths should produce different path hashes")
	}
}

func TestProjectIDUniqueness(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, ".checkpoint-changelog.yaml")

	// Create multiple meta documents - should have different project IDs
	meta1, err := createMetaDocument(changelogPath, "1.0.0")
	if err != nil {
		t.Fatalf("createMetaDocument failed: %v", err)
	}

	meta2, err := createMetaDocument(changelogPath, "1.0.0")
	if err != nil {
		t.Fatalf("createMetaDocument failed: %v", err)
	}

	if meta1.ProjectID == meta2.ProjectID {
		t.Error("different calls should produce different project IDs")
	}
}

func TestReadMetaDocument_ValidMeta(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, ".checkpoint-changelog.yaml")

	// Create changelog with meta document
	err := createChangelogWithMeta(changelogPath, "1.0.0")
	if err != nil {
		t.Fatalf("createChangelogWithMeta failed: %v", err)
	}

	// Read meta document
	meta, err := ReadMetaDocument(changelogPath)
	if err != nil {
		t.Fatalf("ReadMetaDocument failed: %v", err)
	}

	if meta == nil {
		t.Fatal("expected meta document, got nil")
	}

	// Verify fields
	if meta.DocumentType != "meta" {
		t.Errorf("expected document_type 'meta', got %q", meta.DocumentType)
	}
	if meta.SchemaVersion != "1" {
		t.Errorf("expected schema_version '1', got %q", meta.SchemaVersion)
	}
	if meta.ProjectID == "" {
		t.Error("expected non-empty project_id")
	}
	if meta.PathHash == "" {
		t.Error("expected non-empty path_hash")
	}
	if meta.ToolVersion != "1.0.0" {
		t.Errorf("expected tool_version '1.0.0', got %q", meta.ToolVersion)
	}
}

func TestReadMetaDocument_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, ".checkpoint-changelog.yaml")

	// Read from non-existent file
	meta, err := ReadMetaDocument(changelogPath)
	if err != nil {
		t.Fatalf("ReadMetaDocument should not error on missing file: %v", err)
	}

	if meta != nil {
		t.Error("expected nil meta for missing file")
	}
}

func TestReadMetaDocument_NoMeta(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, ".checkpoint-changelog.yaml")

	// Create changelog without meta document
	content := `---
schema_version: "1"
timestamp: "2023-01-01T12:00:00Z"
commit_hash: "abc123"
changes:
  - summary: "test change"
    change_type: "feature"
`
	err := os.WriteFile(changelogPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Read meta document
	meta, err := ReadMetaDocument(changelogPath)
	if err != nil {
		t.Fatalf("ReadMetaDocument failed: %v", err)
	}

	if meta != nil {
		t.Error("expected nil meta for changelog without meta document")
	}
}

func TestReadMetaDocument_WithMultipleDocuments(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, ".checkpoint-changelog.yaml")

	// Create changelog with meta and checkpoint documents
	content := `---
schema_version: "1"
document_type: meta
project_id: "01ARZ3NDEKTSV4RRFFQ69G5FAV"
path_hash: "abcdef1234567890"
created_at: "2023-01-01T12:00:00Z"
tool_version: "2.5.0"
---
schema_version: "1"
timestamp: "2023-01-01T13:00:00Z"
commit_hash: "abc123"
changes:
  - summary: "first change"
    change_type: "feature"
---
schema_version: "1"
timestamp: "2023-01-02T14:00:00Z"
commit_hash: "def456"
changes:
  - summary: "second change"
    change_type: "fix"
`
	err := os.WriteFile(changelogPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Read meta document
	meta, err := ReadMetaDocument(changelogPath)
	if err != nil {
		t.Fatalf("ReadMetaDocument failed: %v", err)
	}

	if meta == nil {
		t.Fatal("expected meta document, got nil")
	}

	// Verify it read the first document (meta)
	if meta.DocumentType != "meta" {
		t.Errorf("expected document_type 'meta', got %q", meta.DocumentType)
	}
	if meta.ProjectID != "01ARZ3NDEKTSV4RRFFQ69G5FAV" {
		t.Errorf("expected project_id '01ARZ3NDEKTSV4RRFFQ69G5FAV', got %q", meta.ProjectID)
	}
	if meta.PathHash != "abcdef1234567890" {
		t.Errorf("expected path_hash 'abcdef1234567890', got %q", meta.PathHash)
	}
	if meta.ToolVersion != "2.5.0" {
		t.Errorf("expected tool_version '2.5.0', got %q", meta.ToolVersion)
	}
}

func TestReadMetaDocument_CorruptedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, ".checkpoint-changelog.yaml")

	// Create changelog with corrupted YAML
	content := `---
schema_version: "1"
document_type: meta
project_id: "01ARZ3NDEKTSV4RRFFQ69G5FAV
path_hash: "abcdef1234567890"
this is not valid yaml: [[[
`
	err := os.WriteFile(changelogPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Read meta document - should return error
	meta, err := ReadMetaDocument(changelogPath)
	if err == nil {
		t.Error("expected error for corrupted YAML")
	}
	if meta != nil {
		t.Error("expected nil meta for corrupted YAML")
	}
}

func TestReadMetaDocument_NonMetaFirstDocument(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, ".checkpoint-changelog.yaml")

	// Create changelog where first document is not meta
	content := `---
schema_version: "1"
timestamp: "2023-01-01T13:00:00Z"
commit_hash: "abc123"
changes:
  - summary: "first change"
    change_type: "feature"
---
schema_version: "1"
document_type: meta
project_id: "01ARZ3NDEKTSV4RRFFQ69G5FAV"
path_hash: "abcdef1234567890"
created_at: "2023-01-01T12:00:00Z"
tool_version: "1.0.0"
`
	err := os.WriteFile(changelogPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Read meta document - should return nil since first doc is not meta
	meta, err := ReadMetaDocument(changelogPath)
	if err != nil {
		t.Fatalf("ReadMetaDocument failed: %v", err)
	}

	if meta != nil {
		t.Error("expected nil meta when first document is not meta type")
	}
}
