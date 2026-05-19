package storage

import "time"

// StoredMetadata represents sidecar metadata for a stored skill
type StoredMetadata struct {
	SkillName     string    `json:"skill_name"`
	Description   string    `json:"description"`
	OriginProject string    `json:"origin_project"`
	OriginPath    string    `json:"origin_path"`
	SavedAt       time.Time `json:"saved_at"`
}

// StoredSkill combines a skill ID with its metadata
type StoredSkill struct {
	ID       string
	Metadata StoredMetadata
}

// ProjectInfo represents a registered project
type ProjectInfo struct {
	Path       string    `json:"path"`
	LastSynced time.Time `json:"last_synced"`
}

// ProjectRegistry holds all registered projects
type ProjectRegistry struct {
	Projects []ProjectInfo `json:"projects"`
}
