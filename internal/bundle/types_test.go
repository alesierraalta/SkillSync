package bundle

import (
	"testing"
	"time"
)

func TestManifestVersion(t *testing.T) {
	if ManifestVersion != 1 {
		t.Errorf("ManifestVersion = %d, want 1", ManifestVersion)
	}
}

func TestDuplicateActionValues(t *testing.T) {
	if DuplicateSkip != 0 {
		t.Errorf("DuplicateSkip = %d, want 0", DuplicateSkip)
	}
	if DuplicateOverwrite != 1 {
		t.Errorf("DuplicateOverwrite = %d, want 1", DuplicateOverwrite)
	}
}

func TestManifestStruct(t *testing.T) {
	m := Manifest{
		Version:   1,
		CreatedAt: time.Now(),
		CreatedBy: "synck",
		Skills: []ManifestSkill{
			{Name: "test-skill", OriginProvider: "claude", Description: "A test skill"},
		},
	}
	if m.Version != 1 {
		t.Errorf("Manifest.Version = %d, want 1", m.Version)
	}
	if m.CreatedBy != "synck" {
		t.Errorf("Manifest.CreatedBy = %s, want synck", m.CreatedBy)
	}
	if len(m.Skills) != 1 {
		t.Fatalf("len(Manifest.Skills) = %d, want 1", len(m.Skills))
	}
	if m.Skills[0].Name != "test-skill" {
		t.Errorf("Manifest.Skills[0].Name = %s, want test-skill", m.Skills[0].Name)
	}
	if m.Skills[0].OriginProvider != "claude" {
		t.Errorf("Manifest.Skills[0].OriginProvider = %s, want claude", m.Skills[0].OriginProvider)
	}
}

func TestImportOptionsDefaults(t *testing.T) {
	opts := ImportOptions{}
	if opts.OnDuplicate != DuplicateSkip {
		t.Errorf("default OnDuplicate = %d, want DuplicateSkip (0)", opts.OnDuplicate)
	}
	if len(opts.Targets) != 0 {
		t.Errorf("default Targets should be nil, got %v", opts.Targets)
	}
}

func TestImportResultStruct(t *testing.T) {
	r := ImportResult{
		Skill:  "test-skill",
		Status: "installed",
	}
	if r.Skill != "test-skill" {
		t.Errorf("ImportResult.Skill = %s, want test-skill", r.Skill)
	}
	if r.Status != "installed" {
		t.Errorf("ImportResult.Status = %s, want installed", r.Status)
	}
	if r.Error != nil {
		t.Errorf("ImportResult.Error should be nil, got %v", r.Error)
	}
}

func TestErrSkillNotFound(t *testing.T) {
	if ErrSkillNotFound == nil {
		t.Fatal("ErrSkillNotFound should not be nil")
	}
	if ErrSkillNotFound.Error() == "" {
		t.Error("ErrSkillNotFound should have non-empty message")
	}
}
