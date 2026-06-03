package agentdetect

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// writeFile is a test helper that writes content to a path, creating parent
// directories as needed. Fails the test on any error.
func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
}

// mkdir is a test helper that creates a directory and all parents.
func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("MkdirAll %s: %v", path, err)
	}
}

// ─── Detect() top-level tests ────────────────────────────────────────────────

func TestDetect_HomeUnresolvable(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	_, err := Detect()
	if err == nil {
		t.Error("expected error when HOME is unresolvable, got nil")
	}
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) && err.Error() == "" {
		// Accept any non-nil error (os.UserHomeDir may return PathError or plain error)
	}
}

func TestDetect_MissingDirSkipped(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	result, err := Detect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d agents", len(result))
	}
}

func TestDetect_MalformedConfigUnreadable(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	// Create ~/.gemini/ with bad settings.json
	mkdir(t, filepath.Join(home, ".gemini"))
	writeFile(t, filepath.Join(home, ".gemini", "settings.json"), `{bad json`)

	result, err := Detect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 agent (gemini), got %d", len(result))
	}
	agent := result[0]
	if agent.Status != StatusUnreadable {
		t.Errorf("Status: got %q, want %q", agent.Status, StatusUnreadable)
	}
	if !agent.Present {
		t.Error("Present: expected true")
	}
}

func TestDetect_EmptyConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	// Create ~/.gemini/ with empty-ish settings.json (no mcpServers key)
	mkdir(t, filepath.Join(home, ".gemini"))
	writeFile(t, filepath.Join(home, ".gemini", "settings.json"), `{}`)

	result, err := Detect()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(result))
	}
	if len(result[0].MCPServers) != 0 {
		t.Errorf("expected 0 MCP servers, got %d", len(result[0].MCPServers))
	}
}

// ─── Claude tests ─────────────────────────────────────────────────────────────

func TestDetectClaude_DualMCPFormat(t *testing.T) {
	home := t.TempDir()

	// Create ~/.claude/mcp/srv1.json (single-server format)
	writeFile(t, filepath.Join(home, ".claude", "mcp", "srv1.json"),
		`{"command":"node","args":["server1.js"]}`)
	// Create ~/.claude/mcp/srv2.json
	writeFile(t, filepath.Join(home, ".claude", "mcp", "srv2.json"),
		`{"url":"http://localhost:3000","type":"http"}`)

	result := detectClaude(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if len(result.MCPServers) != 2 {
		t.Fatalf("expected 2 MCP servers, got %d", len(result.MCPServers))
	}
	names := map[string]bool{}
	for _, s := range result.MCPServers {
		names[s.Name] = true
	}
	if !names["srv1"] || !names["srv2"] {
		t.Errorf("expected srv1 and srv2, got %v", names)
	}
}

func TestDetectClaude_MCPDirAbsent(t *testing.T) {
	home := t.TempDir()
	// ~/.claude exists but no mcp/ dir
	mkdir(t, filepath.Join(home, ".claude"))

	result := detectClaude(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if result.Status != StatusOK {
		t.Errorf("Status: got %q, want %q", result.Status, StatusOK)
	}
	if len(result.MCPServers) != 0 {
		t.Errorf("expected 0 MCP servers, got %d", len(result.MCPServers))
	}
}

func TestDetectClaude_MalformedMCPFile(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".claude", "mcp", "bad.json"), `{broken`)

	result := detectClaude(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if result.Status != StatusUnreadable {
		t.Errorf("Status: got %q, want %q", result.Status, StatusUnreadable)
	}
}

func TestDetectClaude_SettingsAbsent(t *testing.T) {
	home := t.TempDir()
	mkdir(t, filepath.Join(home, ".claude"))
	// No settings.json

	result := detectClaude(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if len(result.Plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(result.Plugins))
	}
}

func TestDetectClaude_EnabledPluginsCrossRef(t *testing.T) {
	home := t.TempDir()

	// settings.json with enabledPlugins
	writeFile(t, filepath.Join(home, ".claude", "settings.json"),
		`{"enabledPlugins":{"git-tools@github":true,"orphan@npm":true}}`)

	// Cache entry for git-tools
	cacheDir := filepath.Join(home, ".claude", "plugins", "cache", "github", "git-tools", "1.2.0", ".claude-plugin")
	writeFile(t, filepath.Join(cacheDir, "plugin.json"),
		`{"version":"1.2.0","description":"Git tools plugin"}`)
	writeFile(t, filepath.Join(cacheDir, ".mcp.json"),
		`{"mcpServers":{"git-mcp":{"command":"node","args":["git-mcp.js"]}}}`)
	// No cache for orphan

	result := detectClaude(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}

	pluginMap := map[string]Plugin{}
	for _, p := range result.Plugins {
		pluginMap[p.Name] = p
	}

	gitTools, ok := pluginMap["git-tools"]
	if !ok {
		t.Fatal("expected git-tools plugin")
	}
	if gitTools.Version != "1.2.0" {
		t.Errorf("git-tools version: got %q, want %q", gitTools.Version, "1.2.0")
	}

	orphan, ok := pluginMap["orphan"]
	if !ok {
		t.Fatal("expected orphan plugin")
	}
	if orphan.Version != "" {
		t.Errorf("orphan version: expected empty, got %q", orphan.Version)
	}

	// MCP from plugin cache
	mcpNames := map[string]bool{}
	for _, s := range result.MCPServers {
		mcpNames[s.Name] = true
	}
	if !mcpNames["git-mcp"] {
		t.Errorf("expected git-mcp server, got %v", mcpNames)
	}
}

// ─── Gemini tests ──────────────────────────────────────────────────────────────

func TestDetectGemini_InlineMCP(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".gemini", "settings.json"),
		`{"mcpServers":{"fetch":{"command":"node","args":["fetch.js"]}}}`)

	result := detectGemini(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if len(result.MCPServers) != 1 {
		t.Fatalf("expected 1 MCP server, got %d", len(result.MCPServers))
	}
	if result.MCPServers[0].Name != "fetch" {
		t.Errorf("Name: got %q, want %q", result.MCPServers[0].Name, "fetch")
	}
}

func TestDetectGemini_MCPKeyAbsent(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".gemini", "settings.json"), `{"theme":"dark"}`)

	result := detectGemini(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if result.Status != StatusOK {
		t.Errorf("Status: got %q, want ok", result.Status)
	}
	if len(result.MCPServers) != 0 {
		t.Errorf("expected 0 MCP servers, got %d", len(result.MCPServers))
	}
}

func TestDetectGemini_MalformedSettings(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".gemini", "settings.json"), `{bad`)

	result := detectGemini(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if result.Status != StatusUnreadable {
		t.Errorf("Status: got %q, want unreadable", result.Status)
	}
}

// ─── Antigravity tests ─────────────────────────────────────────────────────────

func TestDetectAntigravity(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".gemini", "antigravity", "settings.json"),
		`{"mcpServers":{"ag-server":{"command":"ag","args":[]}}}`)

	result := detectAntigravity(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if len(result.MCPServers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(result.MCPServers))
	}
	if result.MCPServers[0].Name != "ag-server" {
		t.Errorf("Name: got %q, want ag-server", result.MCPServers[0].Name)
	}
}

// ─── OpenCode tests ────────────────────────────────────────────────────────────

func TestDetectOpenCode_MCPAndAgents(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".config", "opencode", "opencode.json"),
		`{"mcpServers":{"oc-srv":{"command":"oc","args":[]}},"agent":{"my-agent":{}}}`)

	result := detectOpenCode(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if len(result.MCPServers) != 1 {
		t.Fatalf("expected 1 MCP server, got %d", len(result.MCPServers))
	}
	if result.MCPServers[0].Name != "oc-srv" {
		t.Errorf("MCP name: got %q, want oc-srv", result.MCPServers[0].Name)
	}
	if len(result.Plugins) != 1 {
		t.Fatalf("expected 1 plugin (agent), got %d", len(result.Plugins))
	}
	if result.Plugins[0].Name != "my-agent" {
		t.Errorf("Plugin name: got %q, want my-agent", result.Plugins[0].Name)
	}
}

// ─── Codex tests ──────────────────────────────────────────────────────────────

func TestDetectCodex_PerPluginWalk(t *testing.T) {
	home := t.TempDir()
	// Create ~/.codex/.tmp/plugins/plugins/my-plugin/
	pluginDir := filepath.Join(home, ".codex", ".tmp", "plugins", "plugins", "my-plugin")
	writeFile(t, filepath.Join(pluginDir, ".codex-plugin", "plugin.json"),
		`{"version":"2.0.0","description":"My plugin"}`)
	writeFile(t, filepath.Join(pluginDir, ".mcp.json"),
		`{"mcpServers":{"codex-srv":{"command":"cx","args":[]}}}`)

	result := detectCodex(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if len(result.Plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(result.Plugins))
	}
	if result.Plugins[0].Name != "my-plugin" {
		t.Errorf("Plugin name: got %q, want my-plugin", result.Plugins[0].Name)
	}
	if result.Plugins[0].Version != "2.0.0" {
		t.Errorf("Plugin version: got %q, want 2.0.0", result.Plugins[0].Version)
	}
	if len(result.MCPServers) != 1 {
		t.Fatalf("expected 1 MCP server from plugin, got %d", len(result.MCPServers))
	}
	if result.MCPServers[0].Name != "codex-srv" {
		t.Errorf("MCP name: got %q, want codex-srv", result.MCPServers[0].Name)
	}
}

func TestDetectCodex_PluginDirAbsent(t *testing.T) {
	home := t.TempDir()
	// ~/.codex exists but no .tmp/plugins/plugins
	mkdir(t, filepath.Join(home, ".codex"))

	result := detectCodex(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if result.Status != StatusOK {
		t.Errorf("Status: got %q, want ok", result.Status)
	}
	if len(result.Plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(result.Plugins))
	}
}

// ─── Qwen tests ───────────────────────────────────────────────────────────────

func TestDetectQwen(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".qwen", "settings.json"),
		`{"mcpServers":{"qwen-mcp":{"command":"qwen","args":[]}}}`)

	result := detectQwen(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if len(result.MCPServers) != 1 {
		t.Fatalf("expected 1 MCP server, got %d", len(result.MCPServers))
	}
	if result.MCPServers[0].Name != "qwen-mcp" {
		t.Errorf("Name: got %q, want qwen-mcp", result.MCPServers[0].Name)
	}
}

// ─── Best-effort tests ────────────────────────────────────────────────────────

func TestDetectPi_PresentOnly(t *testing.T) {
	home := t.TempDir()
	mkdir(t, filepath.Join(home, ".pi", "agent"))

	result := detectPi(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if result.Status != StatusPresentOnly {
		t.Errorf("Status: got %q, want %q", result.Status, StatusPresentOnly)
	}
	if len(result.MCPServers) != 0 {
		t.Errorf("expected 0 MCPs, got %d", len(result.MCPServers))
	}
}

func TestDetectCursor(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, ".cursor", "mcp.json"),
		`{"mcpServers":{"cursor-mcp":{"command":"cursor","args":[]}}}`)

	result := detectCursor(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if len(result.MCPServers) != 1 {
		t.Fatalf("expected 1 MCP server, got %d", len(result.MCPServers))
	}
	if result.MCPServers[0].Name != "cursor-mcp" {
		t.Errorf("Name: got %q, want cursor-mcp", result.MCPServers[0].Name)
	}
}

func TestDetectCopilot(t *testing.T) {
	home := t.TempDir()
	mkdir(t, filepath.Join(home, ".copilot"))

	result := detectCopilot(home)
	if !result.Present {
		t.Fatal("expected Present=true")
	}
	if result.Status != StatusPresentOnly {
		t.Errorf("Status: got %q, want %q", result.Status, StatusPresentOnly)
	}
}
