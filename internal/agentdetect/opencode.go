package agentdetect

import (
	"os"
	"path/filepath"
)

// detectOpenCode detects OpenCode by checking ~/.config/opencode/opencode.json.
// MCP servers come from the top-level mcpServers key; plugins come from the
// agent key (map of agent name → agent config).
func detectOpenCode(home string) AgentInfo {
	cfgPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return AgentInfo{}
	}

	info := AgentInfo{
		Name:       "OpenCode",
		Present:    true,
		ConfigPath: cfgPath,
		Status:     StatusOK,
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		info.Status = StatusUnreadable
		return info
	}

	// Parse MCP servers
	servers, err := parseMCPMap(data)
	if err != nil {
		info.Status = StatusUnreadable
		return info
	}
	info.MCPServers = servers

	// Parse agents as plugins
	var cfg struct {
		Agents map[string]interface{} `json:"agent"`
	}
	if parseJSON(data, &cfg) == nil {
		for name := range cfg.Agents {
			info.Plugins = append(info.Plugins, Plugin{
				Name:    name,
				Enabled: true,
			})
		}
	}

	return info
}
