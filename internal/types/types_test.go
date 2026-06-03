package types

import "testing"

func TestTypes(t *testing.T) {
	s := Skill{
		ID: "test",
		Metadata: Metadata{
			Description: "desc",
			AutoInvoke:  []string{"test"},
		},
	}

	if s.ID != "test" {
		t.Errorf("Expected ID test, got %s", s.ID)
	}

	if s.Metadata.Description != "desc" {
		t.Error("Metadata Description mismatch")
	}

	if len(s.Metadata.AutoInvoke) == 0 || s.Metadata.AutoInvoke[0] != "test" {
		t.Error("Metadata AutoInvoke mismatch")
	}
}
