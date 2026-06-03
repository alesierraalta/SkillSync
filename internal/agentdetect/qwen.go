package agentdetect

import (
	"os"
	"path/filepath"
)

// detectQwen detects Qwen Code by checking ~/.qwen/ and parsing
// ~/.qwen/settings.json for mcpServers (inline map format).
func detectQwen(home string) AgentInfo {
	base := filepath.Join(home, ".qwen")
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return AgentInfo{}
	}

	info := AgentInfo{
		Name:       "Qwen Code",
		Present:    true,
		ConfigPath: base,
		Status:     StatusOK,
	}

	settingsPath := filepath.Join(base, "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return info
	}

	servers, err := parseMCPMap(data)
	if err != nil {
		info.Status = StatusUnreadable
		return info
	}

	info.MCPServers = servers
	return info
}
