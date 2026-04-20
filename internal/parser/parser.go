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

	content := string(raw)
	parts := strings.SplitN(content, "---", 3)
	
	if len(parts) < 3 {
		return &types.Skill{
			Path:    path,
			Name:    filepath.Base(filepath.Dir(path)),
			RawBody: content,
		}, nil
	}

	prefix := parts[0]
	yamlStr := parts[1]
	body := strings.TrimSpace(parts[2])

	var meta types.Metadata
	err = yaml.Unmarshal([]byte(yamlStr), &meta)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &types.Skill{
		Path:     path,
		Name:     filepath.Base(filepath.Dir(path)),
		Prefix:   prefix,
		Metadata: meta,
		RawBody:  body,
	}, nil
}

// Save writes SKILL.md back with comments preserved.
// Note: Metadata fields are updated into the YAML Node structure.
func Save(path string, skill *types.Skill) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(raw)
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		// No frontmatter, just write body
		return os.WriteFile(path, []byte(skill.RawBody), 0644)
	}

	yamlStr := parts[1]
	body := parts[2]

	var node yaml.Node
	err = yaml.Unmarshal([]byte(yamlStr), &node)
	if err != nil {
		return err
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
		return err
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

		finalContent := skill.Prefix + yamlResult + strings.TrimPrefix(body, "\n")

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
