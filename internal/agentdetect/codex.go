package agentdetect

import (
	"os"
	"path/filepath"
)

// codexPluginRoot is the path under HOME where Codex stores plugin subdirectories.
const codexPluginRoot = ".codex/.tmp/plugins/plugins"

// detectCodex detects Codex by checking ~/.codex/ and enumerating plugins
// under ~/.codex/.tmp/plugins/plugins/. Each plugin subdirectory may contain
// .codex-plugin/plugin.json and .mcp.json.
func detectCodex(home string) AgentInfo {
	base := filepath.Join(home, ".codex")
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return AgentInfo{}
	}

	info := AgentInfo{
		Name:       "Codex",
		Present:    true,
		ConfigPath: base,
		Status:     StatusOK,
	}

	pluginDir := filepath.Join(home, codexPluginRoot)
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		// Plugin directory absent — present but no plugins
		return info
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pluginName := e.Name()
		pluginBase := filepath.Join(pluginDir, pluginName)
		codexPluginDir := filepath.Join(pluginBase, ".codex-plugin")

		p := Plugin{
			Name:    pluginName,
			Enabled: true,
		}

		// Read plugin.json for metadata
		metaData, err := os.ReadFile(filepath.Join(codexPluginDir, "plugin.json"))
		if err == nil {
			var meta struct {
				Version     string `json:"version"`
				Description string `json:"description"`
			}
			if parseJSON(metaData, &meta) == nil {
				p.Version = meta.Version
			}
		}

		info.Plugins = append(info.Plugins, p)

		// Read optional .mcp.json at plugin root
		mcpData, err := os.ReadFile(filepath.Join(pluginBase, ".mcp.json"))
		if err == nil {
			srvs, err := parseMCPMap(mcpData)
			if err == nil {
				info.MCPServers = append(info.MCPServers, srvs...)
			}
		}
	}

	return info
}
