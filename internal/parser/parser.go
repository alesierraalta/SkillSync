package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"skillsync/tui/internal/types"
)

// Parse reads SKILL.md and extracts YAML frontmatter + RawBody.
func Parse(path string) (*types.Skill, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s, err := ParseContent(string(raw))
	if err != nil {
		return nil, err
	}
	s.ID = filepath.ToSlash(path)
	s.Path = path
	if s.Name == "" {
		s.Name = filepath.Base(filepath.Dir(path))
	}
	return s, nil
}

// ParseContent extracts skill info from raw string content.
func ParseContent(content string) (*types.Skill, error) {
	parts := strings.SplitN(content, "---", 3)

	if len(parts) < 3 {
		return &types.Skill{
			RawBody: content,
		}, nil
	}

	prefix := parts[0]
	yamlStr := parts[1]
	body := strings.TrimSpace(parts[2])

	var flexible struct {
		Description string      `yaml:"description"`
		AutoInvoke  interface{} `yaml:"auto_invoke"`
		Scope       interface{} `yaml:"scope"`
		LocalOnly   bool        `yaml:"local_only"`
		Metadata    struct {
			Scope      interface{} `yaml:"scope"`
			AutoInvoke interface{} `yaml:"auto_invoke"`
		} `yaml:"metadata"`
	}

	err := yaml.Unmarshal([]byte(yamlStr), &flexible)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	var meta types.Metadata
	meta.Description = flexible.Description
	meta.LocalOnly = flexible.LocalOnly

	// Handle AutoInvoke safely (bool or non-empty string)
	if flexible.Metadata.AutoInvoke != nil {
		meta.AutoInvoke = isTrue(flexible.Metadata.AutoInvoke)
	} else if flexible.AutoInvoke != nil {
		meta.AutoInvoke = isTrue(flexible.AutoInvoke)
	}

	if flexible.Metadata.Scope != nil {
		meta.Scope = normalizeScope(flexible.Metadata.Scope)
	} else if flexible.Scope != nil {
		meta.Scope = normalizeScope(flexible.Scope)
	}

	return &types.Skill{
		Prefix:   prefix,
		Metadata: meta,
		RawBody:  body,
	}, nil
}

func isTrue(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val != "" && val != "false"
	}
	return false
}

// Save writes SKILL.md back with comments preserved.
// Note: Metadata fields are updated into the YAML Node structure.
func Format(skill *types.Skill) (string, error) {
	var yamlStr string
	if skill.Path != "" {
		raw, err := os.ReadFile(skill.Path)
		if err == nil {
			content := string(raw)
			parts := strings.SplitN(content, "---", 3)
			if len(parts) >= 3 {
				yamlStr = parts[1]
			}
		}
	}

	if yamlStr == "" {
		// No existing file or no frontmatter, create minimal YAML
		return skill.Prefix + "---\nscope: " + skill.Metadata.Scope + "\n---\n" + skill.RawBody, nil
	}

	var node yaml.Node
	err := yaml.Unmarshal([]byte(yamlStr), &node)
	if err != nil {
		return "", err
	}

	// Update metadata into node while preserving comments
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		mapping := node.Content[0]
		if mapping.Kind == yaml.MappingNode {
			updateMappingNode(mapping, skill.Metadata)
		}
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	err = enc.Encode(&node)
	if err != nil {
		return "", err
	}

	// yaml.v3 appends its own "---" at start if not present, check
	yamlResult := buf.String()
	if !strings.HasPrefix(yamlResult, "---") {
		yamlResult = "---\n" + yamlResult
	}
	if !strings.HasSuffix(yamlResult, "---\n") {
		if strings.HasSuffix(yamlResult, "---") {
			yamlResult = yamlResult + "\n"
		} else {
			yamlResult = yamlResult + "---\n"
		}
	}

	return skill.Prefix + yamlResult + skill.RawBody, nil
}

// Save writes SKILL.md back with comments preserved.
// Note: Metadata fields are updated into the YAML Node structure.
func Save(path string, skill *types.Skill) error {
	finalContent, err := Format(skill)
	if err != nil {
		return err
	}

	// Atomic write
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(finalContent), 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func updateMappingNode(mapping *yaml.Node, meta types.Metadata) {
	// Simple map for fields
	fields := map[string]string{
		"scope":       meta.Scope,
		"description": meta.Description,
	}
	
	boolFields := map[string]bool{
		"auto_invoke": meta.AutoInvoke,
		"local_only":  meta.LocalOnly,
	}

	for i := 0; i < len(mapping.Content); i += 2 {
		key := mapping.Content[i].Value
		if val, ok := fields[key]; ok {
			mapping.Content[i+1].Value = val
			mapping.Content[i+1].Tag = "!!str"
			delete(fields, key)
		} else if bVal, ok := boolFields[key]; ok {
			mapping.Content[i+1].Value = fmt.Sprintf("%v", bVal)
			mapping.Content[i+1].Tag = "!!bool"
			delete(boolFields, key)
		}
	}

	// Add missing fields
	for k, v := range fields {
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: k},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v},
		)
	}
	for k, v := range boolFields {
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: k},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: fmt.Sprintf("%v", v)},
		)
	}
}


func normalizeScope(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []interface{}:
		var parts []string
		for _, p := range val {
			if s, ok := p.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ", ")
	}
	return ""
}
