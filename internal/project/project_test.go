package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendRecommendations(t *testing.T) {
	t.Run("appends recommendations with all addition types", func(t *testing.T) {
		dir := t.TempDir()
		projectFile := filepath.Join(dir, "project.yaml")

		initial := "schema_version: \"1\"\nname: test\n"
		if err := os.WriteFile(projectFile, []byte(initial), 0644); err != nil {
			t.Fatalf("write initial file: %v", err)
		}

		additions := ProjectAdditions{
			KeyInsights: []Insight{
				{Insight: "test insight", Rationale: "because testing"},
			},
			FailedApproaches: []FailedApproach{
				{Approach: "bad approach", WhyFailed: "did not work"},
			},
			DesignPrinciples: []Principle{
				{Principle: "keep it simple", Rationale: "simplicity wins"},
			},
		}

		err := AppendRecommendations(projectFile, "2026-03-14T00:00:00Z", additions, nil)
		if err != nil {
			t.Fatalf("AppendRecommendations returned error: %v", err)
		}

		data, err := os.ReadFile(projectFile)
		if err != nil {
			t.Fatalf("read file: %v", err)
		}
		content := string(data)

		// Verify initial content is preserved
		if !strings.Contains(content, "name: test") {
			t.Error("initial content not preserved")
		}

		// Verify recommendations document was appended
		if !strings.Contains(content, "document_type: recommendations") {
			t.Error("missing document_type: recommendations")
		}
		if !strings.Contains(content, "timestamp: \"2026-03-14T00:00:00Z\"") {
			t.Errorf("missing timestamp, got:\n%s", content)
		}
		if !strings.Contains(content, "---") {
			t.Error("missing YAML document separator")
		}

		// Verify additions
		if !strings.Contains(content, "test insight") {
			t.Error("missing insight")
		}
		if !strings.Contains(content, "bad approach") {
			t.Error("missing failed approach")
		}
		if !strings.Contains(content, "keep it simple") {
			t.Error("missing design principle")
		}
	})

	t.Run("appends with empty additions", func(t *testing.T) {
		dir := t.TempDir()
		projectFile := filepath.Join(dir, "project.yaml")

		initial := "schema_version: \"1\"\nname: test\n"
		if err := os.WriteFile(projectFile, []byte(initial), 0644); err != nil {
			t.Fatalf("write initial file: %v", err)
		}

		err := AppendRecommendations(projectFile, "2026-03-14T00:00:00Z", ProjectAdditions{}, nil)
		if err != nil {
			t.Fatalf("AppendRecommendations returned error: %v", err)
		}

		data, err := os.ReadFile(projectFile)
		if err != nil {
			t.Fatalf("read file: %v", err)
		}
		content := string(data)

		if !strings.Contains(content, "document_type: recommendations") {
			t.Error("empty additions should still append a recommendations doc")
		}
	})

	t.Run("errors on non-existent file", func(t *testing.T) {
		err := AppendRecommendations("/nonexistent/path/project.yaml", "2026-03-14T00:00:00Z", ProjectAdditions{}, nil)
		if err == nil {
			t.Fatal("expected error for non-existent file, got nil")
		}
	})
}
