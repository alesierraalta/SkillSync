// Package agentdetect provides read-only detection of installed AI agent CLI/TUI tools.
// It enumerates each tool's MCP servers and plugins from well-known config paths
// under the user's HOME directory. Detection is resilient: missing directories
// are silently skipped, malformed configs set Status=unreadable and never panic.
package agentdetect

import (
	"fmt"
	"os"
)

// Status describes the readability of a detected agent's configuration.
type Status string

const (
	// StatusOK means the agent directory exists and its config was parsed successfully.
	StatusOK Status = "ok"
	// StatusUnreadable means the directory exists but config is malformed or unreadable.
	StatusUnreadable Status = "unreadable"
	// StatusPresentOnly means the directory exists but no known config format applies.
	StatusPresentOnly Status = "present-only"
)

// MCPServer describes a single MCP server entry from an agent's config.
type MCPServer struct {
	Name      string
	Command   string
	Args      []string
	Transport string // "stdio" or "http" (sse folds into http)
	Enabled   bool
}

// Plugin describes a plugin/extension registered in an agent's config.
type Plugin struct {
	Name    string
	Version string
	Enabled bool
}

// AgentInfo holds the detection result for a single AI agent tool.
type AgentInfo struct {
	Name       string
	Present    bool
	ConfigPath string
	Status     Status
	MCPServers []MCPServer
	Plugins    []Plugin
}

// detector is a registry entry pairing a display name with its detection function.
// detectFn receives the resolved home directory and returns an AgentInfo.
// If the agent is not present, AgentInfo.Present must be false.
// detectFn must never panic and must never return an error; any IO/parse
// failures are folded into AgentInfo.Status.
type detector struct {
	name     string
	detectFn func(home string) AgentInfo
}

// registry lists all supported tools. Order determines display order in the TUI.
// Stubs are replaced by real implementations in their respective source files.
var registry = []detector{
	{"Claude Code", detectClaude},
	{"Gemini CLI", detectGemini},
	{"Antigravity", detectAntigravity},
	{"OpenCode", detectOpenCode},
	{"Codex", detectCodex},
	{"Qwen Code", detectQwen},
	{"Pi", detectPi},
	{"Cursor", detectCursor},
	{"Copilot", detectCopilot},
}

// Detect runs all registered detectors and returns info for present tools.
// The only error condition is an unresolvable HOME directory; all per-tool
// errors are folded into AgentInfo.Status.
func Detect() ([]AgentInfo, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("agentdetect: resolve home: %w", err)
	}
	out := make([]AgentInfo, 0, len(registry))
	for _, d := range registry {
		info := d.detectFn(home)
		if info.Present {
			out = append(out, info)
		}
	}
	return out, nil
}
