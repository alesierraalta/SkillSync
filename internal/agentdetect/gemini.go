package agentdetect

import (
	"os"
	"path/filepath"
)

// detectGemini detects Gemini CLI by checking ~/.gemini/ and parsing
// ~/.gemini/settings.json for mcpServers (inline map format).
func detectGemini(home string) AgentInfo {
	base := filepath.Join(home, ".gemini")
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return AgentInfo{}
	}

	info := AgentInfo{
		Name:       "Gemini CLI",
		Present:    true,
		ConfigPath: base,
		Status:     StatusOK,
	}

	settingsPath := filepath.Join(base, "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		// Settings absent — present but no config
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

// detectAntigravity detects the Antigravity tool by checking
// ~/.gemini/antigravity/settings.json for mcpServers.
func detectAntigravity(home string) AgentInfo {
	base := filepath.Join(home, ".gemini", "antigravity")
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return AgentInfo{}
	}

	info := AgentInfo{
		Name:       "Antigravity",
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
