package agentdetect

import (
	"encoding/json"
	"fmt"
	"sort"
)

// mcpRaw is the JSON shape for a single MCP server entry.
// It handles both the inline mcpServers map format and the
// standalone per-file format used by Claude Code.
type mcpRaw struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	URL     string   `json:"url"`
	Type    string   `json:"type"`
}

// toServer converts a raw MCP entry into an MCPServer with normalised Transport.
// Transport rules:
//   - url != ""  OR type == "http" OR type == "sse" → "http"
//   - otherwise → "stdio"
func toServer(name string, r mcpRaw) MCPServer {
	transport := "stdio"
	if r.URL != "" || r.Type == "http" || r.Type == "sse" {
		transport = "http"
	}
	return MCPServer{
		Name:      name,
		Command:   r.Command,
		Args:      r.Args,
		Transport: transport,
	}
}

// parseMCPMap parses the mcpServers map format:
//
//	{ "mcpServers": { "<name>": { "command": ..., "args": [...] }, ... } }
//
// A missing mcpServers key returns an empty slice without error.
// Malformed JSON returns a wrapped error.
// Output is sorted by Name for deterministic ordering.
func parseMCPMap(data []byte) ([]MCPServer, error) {
	var wrap struct {
		MCPServers map[string]mcpRaw `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &wrap); err != nil {
		return nil, fmt.Errorf("parse mcpServers: %w", err)
	}
	out := make([]MCPServer, 0, len(wrap.MCPServers))
	for name, r := range wrap.MCPServers {
		out = append(out, toServer(name, r))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// parseMCPFile parses the single-server per-file format used by Claude Code:
//
//	{ "command": "...", "args": [...] }
//
// The server name is supplied by the caller (typically the file stem).
// Malformed JSON returns a wrapped error.
func parseMCPFile(data []byte, name string) (MCPServer, error) {
	var r mcpRaw
	if err := json.Unmarshal(data, &r); err != nil {
		return MCPServer{}, fmt.Errorf("parse mcp file %s: %w", name, err)
	}
	return toServer(name, r), nil
}
