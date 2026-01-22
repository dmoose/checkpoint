package guides

import (
	"testing"
)

func TestGetGuide(t *testing.T) {
	t.Run("valid guide returns content", func(t *testing.T) {
		guide, ok := GetGuide("first-time-user")
		if !ok {
			t.Fatal("expected GetGuide to return true for 'first-time-user'")
		}
		if guide.Content == "" {
			t.Error("expected non-empty content for 'first-time-user' guide")
		}
		if guide.Name != "first-time-user" {
			t.Errorf("expected Name 'first-time-user', got %q", guide.Name)
		}
	})

	t.Run("invalid guide returns false", func(t *testing.T) {
		_, ok := GetGuide("nonexistent-guide")
		if ok {
			t.Error("expected GetGuide to return false for invalid guide name")
		}
	})
}

func TestListGuides(t *testing.T) {
	names := ListGuides()
	if len(names) == 0 {
		t.Fatal("expected ListGuides to return a non-empty list")
	}

	expected := map[string]bool{
		"first-time-user": false,
		"llm-workflow":    false,
		"best-practices":  false,
	}

	for _, name := range names {
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected guide %q not found in ListGuides result", name)
		}
	}
}
