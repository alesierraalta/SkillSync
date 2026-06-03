package agentdetect

import (
	"os"
	"path/filepath"
)

// detectPi detects Pi agent by checking ~/.pi/agent/ directory.
// Pi is a best-effort, present-only detector (no known config format).
func detectPi(home string) AgentInfo {
	base := filepath.Join(home, ".pi", "agent")
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return AgentInfo{}
	}
	return AgentInfo{
		Name:       "Pi",
		Present:    true,
		ConfigPath: base,
		Status:     StatusPresentOnly,
	}
}

// detectCursor detects Cursor by checking ~/.cursor/ and attempting to parse
// ~/.cursor/mcp.json for mcpServers. On any miss or parse failure, the tool
// is reported as present with empty MCPServers (Status: ok, not unreadable).
func detectCursor(home string) AgentInfo {
	base := filepath.Join(home, ".cursor")
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return AgentInfo{}
	}

	info := AgentInfo{
		Name:       "Cursor",
		Present:    true,
		ConfigPath: base,
		Status:     StatusOK,
	}

	mcpPath := filepath.Join(base, "mcp.json")
	data, err := os.ReadFile(mcpPath)
	if err != nil {
		// Config absent — present-only is fine for best-effort tool
		return info
	}

	servers, err := parseMCPMap(data)
	if err != nil {
		// Best-effort: parse failure keeps Status:ok, empty servers
		return info
	}

	info.MCPServers = servers
	return info
}

// detectCopilot detects GitHub Copilot by checking ~/.copilot/.
// Copilot is a best-effort, present-only detector.
func detectCopilot(home string) AgentInfo {
	base := filepath.Join(home, ".copilot")
	if _, err := os.Stat(base); os.IsNotExist(err) {
		return AgentInfo{}
	}
	return AgentInfo{
		Name:       "Copilot",
		Present:    true,
		ConfigPath: base,
		Status:     StatusPresentOnly,
	}
}
