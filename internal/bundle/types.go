package bundle

import (
	"errors"
	"time"
)

// ManifestVersion is the only supported manifest version.
// v1 limitation: no bundle signing or integrity verification.
const ManifestVersion = 1

// MaxEntryBytes limits each individual ZIP entry to 64 MiB to prevent OOM
// from malicious or malformed bundles. Entries exceeding this size are rejected
// before any data is copied.
const MaxEntryBytes = 64 << 20 // 64 MiB

// MaxManifestBytes limits the manifest.json entry to 1 MiB to prevent
// oversized manifest attacks.
const MaxManifestBytes = 1 << 20 // 1 MiB

// MaxTotalBytes limits the sum of all uncompressed entry sizes in a bundle
// to 1 GiB to prevent zip-bomb-style attacks during export pre-checks.
const MaxTotalBytes = 1 << 30 // 1 GiB

// MaxFileCount limits the total number of files in a bundle to prevent
// excessive file count attacks.
const MaxFileCount = 10000

// Manifest describes the contents of a .skillsync bundle.
type Manifest struct {
	Version   int             `json:"version"`
	CreatedAt time.Time       `json:"created_at"`
	CreatedBy string          `json:"created_by"`
	Skills    []ManifestSkill `json:"skills"`
}

// ManifestSkill represents a single skill entry in the bundle manifest.
type ManifestSkill struct {
	Name           string `json:"name"`
	OriginProvider string `json:"origin_provider,omitempty"`
	Description    string `json:"description,omitempty"`
}

// DuplicateAction controls behavior when importing a skill that already exists.
type DuplicateAction int

const (
	// DuplicateSkip is the default: existing skills are not overwritten.
	DuplicateSkip DuplicateAction = iota
	// DuplicateOverwrite replaces existing skill files.
	DuplicateOverwrite
)

// ImportOptions controls bundle import behavior.
type ImportOptions struct {
	// Targets is the list of provider directories to install into.
	// If empty, defaults are derived from active project providers.
	Targets []string

	// OnDuplicate controls whether existing skills are skipped or overwritten.
	OnDuplicate DuplicateAction

	// ProjectRoot is the project root directory for install and sync.
	ProjectRoot string
}

// ImportResult describes the outcome of importing a single skill.
type ImportResult struct {
	Skill  string `json:"skill"`
	Status string `json:"status"` // "installed" | "skipped" | "overwritten" | "failed"
	Error  error  `json:"error,omitempty"`
}

// ErrSkillNotFound is returned when a requested skill is not in the bundle.
var ErrSkillNotFound = errors.New("skill not found in bundle")
