package types

// Metadata represents YAML frontmatter of SKILL.md
type Metadata struct {
	Description string `yaml:"description"`
	AutoInvoke  bool   `yaml:"auto_invoke"`
	Scope       string `yaml:"scope"`
	LocalOnly   bool   `yaml:"local_only"`
}

// Skill represents a loaded skill with its path and content
type Skill struct {
	ID       string
	Name     string
	Path     string
	Prefix   string // New: text before first ---
	Metadata Metadata
	RawBody  string
}
