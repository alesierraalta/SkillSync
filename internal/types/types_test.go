package types

import "testing"

func TestTypes(t *testing.T) {
	s := Skill{
		ID: "test",
		Metadata: Metadata{
			Description: "desc",
			AutoInvoke:  true,
		},
	}

	if s.ID != "test" {
		t.Errorf("Expected ID test, got %s", s.ID)
	}

	if s.Metadata.Description != "desc" {
		t.Error("Metadata Description mismatch")
	}

	if !s.Metadata.AutoInvoke {
		t.Error("Metadata AutoInvoke mismatch")
	}
}
