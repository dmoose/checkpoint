package skills

import (
	"testing"
)

func TestDefaultSkills(t *testing.T) {
	skills := DefaultSkills()
	if len(skills) == 0 {
		t.Fatal("expected DefaultSkills to return a non-empty map")
	}

	expectedNames := []string{"git", "ripgrep", "go"}
	for _, name := range expectedNames {
		content, ok := skills[name]
		if !ok {
			t.Errorf("expected skill %q not found in DefaultSkills", name)
			continue
		}
		if content == "" {
			t.Errorf("expected non-empty content for skill %q", name)
		}
	}
}

func TestGetSkill(t *testing.T) {
	t.Run("valid skill returns content", func(t *testing.T) {
		content, ok := GetSkill("git")
		if !ok {
			t.Fatal("expected GetSkill to return true for 'git'")
		}
		if content == "" {
			t.Error("expected non-empty content for 'git' skill")
		}
	})

	t.Run("invalid skill returns false", func(t *testing.T) {
		_, ok := GetSkill("nonexistent-skill")
		if ok {
			t.Error("expected GetSkill to return false for invalid skill name")
		}
	})
}

func TestListSkills(t *testing.T) {
	names := ListSkills()
	if len(names) == 0 {
		t.Fatal("expected ListSkills to return a non-empty list")
	}

	expected := map[string]bool{
		"git":     false,
		"ripgrep": false,
		"go":      false,
	}

	for _, name := range names {
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected skill %q not found in ListSkills result", name)
		}
	}
}
