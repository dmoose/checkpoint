package checkpoint

import (
	"testing"
	"time"
)

func TestCreateSessionTemplate(t *testing.T) {
	session := createSessionTemplate()

	// Verify schema version
	if session.SchemaVersion != "1" {
		t.Errorf("expected schema_version '1', got %q", session.SchemaVersion)
	}

	// Verify created and updated are set and valid RFC3339
	if session.Created == "" {
		t.Error("expected Created to be set")
	}
	if _, err := time.Parse(time.RFC3339, session.Created); err != nil {
		t.Errorf("Created is not valid RFC3339: %v", err)
	}

	if session.Updated == "" {
		t.Error("expected Updated to be set")
	}
	if session.Created != session.Updated {
		t.Error("expected Created and Updated to be equal for a new template")
	}

	// Verify goals are populated with placeholder text
	if len(session.Goals) == 0 {
		t.Error("expected Goals to have at least one entry")
	}
	if session.Goals[0] == "" {
		t.Error("expected Goals[0] to contain placeholder text")
	}

	// Verify approach is populated
	if session.Approach == "" {
		t.Error("expected Approach to contain placeholder text")
	}

	// Verify next_actions are populated
	if len(session.NextActions) == 0 {
		t.Error("expected NextActions to have at least one entry")
	}
	if session.NextActions[0].Summary == "" {
		t.Error("expected NextActions[0].Summary to contain placeholder text")
	}
	if session.NextActions[0].Priority != "high" {
		t.Errorf("expected NextActions[0].Priority 'high', got %q", session.NextActions[0].Priority)
	}
	if session.NextActions[0].Status != "pending" {
		t.Errorf("expected NextActions[0].Status 'pending', got %q", session.NextActions[0].Status)
	}

	// Verify risks and open questions have placeholders
	if len(session.Risks) == 0 {
		t.Error("expected Risks to have at least one entry")
	}
	if len(session.OpenQuestions) == 0 {
		t.Error("expected OpenQuestions to have at least one entry")
	}

	// Verify current focus is populated
	if session.CurrentFocus == "" {
		t.Error("expected CurrentFocus to contain placeholder text")
	}
}
