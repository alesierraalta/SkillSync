package bundle

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
)

// ReadManifest opens a .skillsync bundle and reads its manifest.json.
// It validates the version and returns the parsed Manifest.
func ReadManifest(bundlePath string) (*Manifest, error) {
	r, err := zip.OpenReader(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("open bundle: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name != "manifest.json" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open manifest: %w", err)
		}
		defer rc.Close()

		data, err := io.ReadAll(io.LimitReader(rc, MaxManifestBytes+1))
		if err != nil {
			return nil, fmt.Errorf("read manifest: %w", err)
		}
		if len(data) > MaxManifestBytes {
			return nil, fmt.Errorf("manifest exceeds max size %d bytes", MaxManifestBytes)
		}

		var m Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse manifest: %w", err)
		}

		if m.Version != ManifestVersion {
			return nil, fmt.Errorf("unsupported manifest version %d (supported: %d)", m.Version, ManifestVersion)
		}

		return &m, nil
	}

	return nil, fmt.Errorf("manifest.json not found in bundle")
}

// writeManifestJSON encodes and writes the manifest to the zip writer.
func writeManifestJSON(zw *zip.Writer, m Manifest) error {
	w, err := zw.Create("manifest.json")
	if err != nil {
		return fmt.Errorf("create manifest entry: %w", err)
	}
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	return nil
}

// deriveOriginProvider attempts to detect the provider ecosystem from an
// origin path (e.g. ".claude/skills/foo" → "claude").
func deriveOriginProvider(originPath string) string {
	// Known provider patterns in order of specificity.
	patterns := []struct {
		keyword  string
		provider string
	}{
		{".claude/skills/", "claude"},
		{".opencode/skills/", "opencode"},
		{".gemini/skills/", "gemini"},
		{".cursor/skills/", "cursor"},
		{".copilot/skills/", "copilot"},
		{".qwen/skills/", "qwen"},
		{".agents/skills/", "agents"},
	}

	for _, p := range patterns {
		if containsIgnorePath(originPath, p.keyword) {
			return p.provider
		}
	}
	return ""
}

// containsIgnorePath checks if s contains substr using filepath-separator-aware
// comparison. On Windows we compare with normalized separators.
func containsIgnorePath(s, substr string) bool {
	// Normalize both to forward slashes for comparison
	normS := normPath(s)
	normSub := normPath(substr)
	for i := 0; i <= len(normS)-len(normSub); i++ {
		if normS[i:i+len(normSub)] == normSub {
			return true
		}
	}
	return false
}

func normPath(p string) string {
	result := make([]byte, len(p))
	for i := 0; i < len(p); i++ {
		if p[i] == '\\' {
			result[i] = '/'
		} else {
			result[i] = p[i]
		}
	}
	return string(result)
}
