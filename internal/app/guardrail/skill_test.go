package guardrail

import (
	"testing"

	"github.com/dmoose/checkpoint/internal/explain"
)

func TestSkillLoaded_Match(t *testing.T) {
	skills := []explain.Skill{
		{Name: "go"},
		{Name: "docker"},
		{Name: "testing"},
	}

	if !skillLoaded(skills, "go") {
		t.Error("expected skillLoaded to return true for 'go'")
	}
	if !skillLoaded(skills, "docker") {
		t.Error("expected skillLoaded to return true for 'docker'")
	}
	if !skillLoaded(skills, "testing") {
		t.Error("expected skillLoaded to return true for 'testing'")
	}
}

func TestSkillLoaded_NoMatch(t *testing.T) {
	skills := []explain.Skill{
		{Name: "go"},
		{Name: "docker"},
	}

	if skillLoaded(skills, "python") {
		t.Error("expected skillLoaded to return false for 'python'")
	}
	if skillLoaded(skills, "") {
		t.Error("expected skillLoaded to return false for empty string")
	}
}

func TestSkillLoaded_EmptySlice(t *testing.T) {
	var skills []explain.Skill

	if skillLoaded(skills, "go") {
		t.Error("expected skillLoaded to return false with nil slice")
	}

	skills = []explain.Skill{}
	if skillLoaded(skills, "go") {
		t.Error("expected skillLoaded to return false with empty slice")
	}
}
