package agentdetect

import (
	"os"
	"path/filepath"
	"strings"
)

// detectClaude detects Claude Code installation and reads its MCP servers and plugins.
// Config paths:
//   - ~/.claude/ (presence marker)
//   - ~/.claude/mcp/*.json (individual MCP server files, parseMCPFile format)
//   - ~/.claude/settings.json (enabledPlugins map)
//   - ~/.claude/plugins/cache/<marketplace>/<name>/<version>/.claude-plugin/plugin.json
//   - ~/.claude/plugins/cache/<marketplace>/<name>/<version>/.claude-plugin/.mcp.json
func detectClaude(home string) AgentInfo {
	base := filepath.Join(home, ".claude")
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return AgentInfo{}
	}

	info := AgentInfo{
		Name:       "Claude Code",
		Present:    true,
		ConfigPath: base,
		Status:     StatusOK,
	}

	// Read MCP servers from ~/.claude/mcp/*.json (per-file format)
	mcpDir := filepath.Join(base, "mcp")
	entries, err := os.ReadDir(mcpDir)
	if err == nil {
		seenMCP := map[string]bool{}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			stem := strings.TrimSuffix(e.Name(), ".json")
			data, err := os.ReadFile(filepath.Join(mcpDir, e.Name()))
			if err != nil {
				info.Status = StatusUnreadable
				continue
			}
			srv, err := parseMCPFile(data, stem)
			if err != nil {
				info.Status = StatusUnreadable
				continue
			}
			if !seenMCP[srv.Name] {
				info.MCPServers = append(info.MCPServers, srv)
				seenMCP[srv.Name] = true
			}
		}
	}

	// Read plugins from ~/.claude/settings.json
	settingsPath := filepath.Join(base, "settings.json")
	settingsData, err := os.ReadFile(settingsPath)
	if err == nil {
		plugins, additionalMCPs := parseClaudePlugins(home, settingsData)
		info.Plugins = plugins
		// Merge additional MCP servers from plugin .mcp.json files (dedupe by name)
		seenMCP := map[string]bool{}
		for _, s := range info.MCPServers {
			seenMCP[s.Name] = true
		}
		for _, srv := range additionalMCPs {
			if !seenMCP[srv.Name] {
				info.MCPServers = append(info.MCPServers, srv)
				seenMCP[srv.Name] = true
			}
		}
	}
	// settings absent → Plugins stays empty, no error

	return info
}

// parseClaudePlugins reads settings.json enabledPlugins, cross-references the
// plugin cache directory, and returns Plugin list + any MCP servers from .mcp.json.
// Format of enabledPlugins: map[string]bool where key is "<name>@<marketplace>".
func parseClaudePlugins(home string, data []byte) ([]Plugin, []MCPServer) {
	var settings struct {
		EnabledPlugins map[string]bool `json:"enabledPlugins"`
	}

	if err := parseJSON(data, &settings); err != nil {
		return nil, nil
	}

	cacheBase := filepath.Join(home, ".claude", "plugins", "cache")
	var plugins []Plugin
	var mcpServers []MCPServer

	for key, enabled := range settings.EnabledPlugins {
		// Key format: "<name>@<marketplace>"
		name, marketplace := splitPluginKey(key)

		p := Plugin{
			Name:    name,
			Enabled: enabled,
		}

		// Try to find version dir: cacheBase/<marketplace>/<name>/<version>/
		marketDir := filepath.Join(cacheBase, marketplace, name)
		versionEntries, err := os.ReadDir(marketDir)
		if err != nil || len(versionEntries) == 0 {
			// Cache entry missing — include plugin with empty version
			plugins = append(plugins, p)
			continue
		}

		// Use first version found (there should be only one active version)
		for _, vEntry := range versionEntries {
			if !vEntry.IsDir() {
				continue
			}
			pluginDir := filepath.Join(marketDir, vEntry.Name(), ".claude-plugin")

			// Read plugin.json for version/description
			pluginMeta, err := os.ReadFile(filepath.Join(pluginDir, "plugin.json"))
			if err == nil {
				var meta struct {
					Version     string `json:"version"`
					Description string `json:"description"`
				}
				if parseJSON(pluginMeta, &meta) == nil {
					p.Version = meta.Version
				}
			}

			// Read optional .mcp.json
			mcpData, err := os.ReadFile(filepath.Join(pluginDir, ".mcp.json"))
			if err == nil {
				srvs, err := parseMCPMap(mcpData)
				if err == nil {
					mcpServers = append(mcpServers, srvs...)
				}
			}
			break // only first version
		}

		plugins = append(plugins, p)
	}

	return plugins, mcpServers
}

// splitPluginKey splits "name@marketplace" into (name, marketplace).
// If there's no "@", marketplace defaults to "unknown".
func splitPluginKey(key string) (name, marketplace string) {
	idx := strings.LastIndex(key, "@")
	if idx < 0 {
		return key, "unknown"
	}
	return key[:idx], key[idx+1:]
}
