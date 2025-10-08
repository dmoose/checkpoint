package config

import (
	"testing"
)

func TestConfigConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"InputFileName", InputFileName, ".checkpoint-input"},
		{"DiffFileName", DiffFileName, ".checkpoint-diff"},
		{"ChangelogFileName", ChangelogFileName, ".checkpoint-changelog.yaml"},
		{"StatusFileName", StatusFileName, ".checkpoint-status.yaml"},
		{"LockFileName", LockFileName, ".checkpoint-lock"},
		{"CheckpointMdFileName", CheckpointMdFileName, "CHECKPOINT.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("expected %s to be %q, got %q", tt.name, tt.expected, tt.constant)
			}
		})
	}
}

func TestFileNameConsistency(t *testing.T) {
	// Ensure all checkpoint files start with .checkpoint prefix (except CHECKPOINT.md)
	checkpointFiles := []struct {
		name string
		file string
	}{
		{"InputFileName", InputFileName},
		{"DiffFileName", DiffFileName},
		{"ChangelogFileName", ChangelogFileName},
		{"StatusFileName", StatusFileName},
		{"LockFileName", LockFileName},
	}

	for _, file := range checkpointFiles {
		t.Run(file.name, func(t *testing.T) {
			if file.file[0] != '.' {
				t.Errorf("expected %s to start with '.', got %q", file.name, file.file)
			}
			if len(file.file) < 2 || file.file[1:12] != "checkpoint-" {
				t.Errorf("expected %s to have '.checkpoint-' prefix, got %q", file.name, file.file)
			}
		})
	}
}

func TestYamlExtensions(t *testing.T) {
	yamlFiles := []struct {
		name string
		file string
	}{
		{"ChangelogFileName", ChangelogFileName},
		{"StatusFileName", StatusFileName},
	}

	for _, file := range yamlFiles {
		t.Run(file.name, func(t *testing.T) {
			if len(file.file) < 5 || file.file[len(file.file)-5:] != ".yaml" {
				t.Errorf("expected %s to end with '.yaml', got %q", file.name, file.file)
			}
		})
	}
}

func TestUniqueFilenames(t *testing.T) {
	files := []string{
		InputFileName,
		DiffFileName,
		ChangelogFileName,
		StatusFileName,
		LockFileName,
		CheckpointMdFileName,
	}

	seen := make(map[string]bool)
	for _, file := range files {
		if seen[file] {
			t.Errorf("duplicate filename found: %q", file)
		}
		seen[file] = true
	}

	if len(seen) != len(files) {
		t.Errorf("expected %d unique filenames, got %d", len(files), len(seen))
	}
}
