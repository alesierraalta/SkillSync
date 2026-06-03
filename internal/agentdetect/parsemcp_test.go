package agentdetect

import (
	"testing"
)

func TestParseMCPMap_StdioServer(t *testing.T) {
	data := []byte(`{"mcpServers":{"my-server":{"command":"node","args":["server.js"]}}}`)
	servers, err := parseMCPMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	s := servers[0]
	if s.Name != "my-server" {
		t.Errorf("Name: got %q, want %q", s.Name, "my-server")
	}
	if s.Command != "node" {
		t.Errorf("Command: got %q, want %q", s.Command, "node")
	}
	if len(s.Args) != 1 || s.Args[0] != "server.js" {
		t.Errorf("Args: got %v, want [server.js]", s.Args)
	}
	if s.Transport != "stdio" {
		t.Errorf("Transport: got %q, want %q", s.Transport, "stdio")
	}
}

func TestParseMCPMap_HTTPServer(t *testing.T) {
	data := []byte(`{"mcpServers":{"remote":{"url":"http://localhost:3000","type":"http"}}}`)
	servers, err := parseMCPMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	s := servers[0]
	if s.Name != "remote" {
		t.Errorf("Name: got %q, want %q", s.Name, "remote")
	}
	if s.Transport != "http" {
		t.Errorf("Transport: got %q, want %q", s.Transport, "http")
	}
}

func TestParseMCPMap_MissingKey(t *testing.T) {
	// No mcpServers key — should return empty slice, no error
	data := []byte(`{"theme":"dark"}`)
	servers, err := parseMCPMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(servers))
	}
}

func TestParseMCPMap_MalformedJSON(t *testing.T) {
	data := []byte(`{invalid}`)
	_, err := parseMCPMap(data)
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

func TestParseMCPMap_Deterministic(t *testing.T) {
	// Multiple servers — output must be sorted by name
	data := []byte(`{"mcpServers":{"zebra":{"command":"z"},"alpha":{"command":"a"},"middle":{"command":"m"}}}`)
	servers, err := parseMCPMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(servers) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(servers))
	}
	names := []string{servers[0].Name, servers[1].Name, servers[2].Name}
	want := []string{"alpha", "middle", "zebra"}
	for i, n := range names {
		if n != want[i] {
			t.Errorf("index %d: got %q, want %q", i, n, want[i])
		}
	}
}

func TestParseMCPFile_SingleServer(t *testing.T) {
	data := []byte(`{"command":"node","args":["server.js"]}`)
	s, err := parseMCPFile(data, "my-server")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Name != "my-server" {
		t.Errorf("Name: got %q, want %q", s.Name, "my-server")
	}
	if s.Command != "node" {
		t.Errorf("Command: got %q, want %q", s.Command, "node")
	}
	if s.Transport != "stdio" {
		t.Errorf("Transport: got %q, want %q", s.Transport, "stdio")
	}
}

func TestToServer_TransportInference(t *testing.T) {
	tests := []struct {
		name      string
		raw       mcpRaw
		wantTrans string
	}{
		{"command-only is stdio", mcpRaw{Command: "node"}, "stdio"},
		{"url makes it http", mcpRaw{URL: "http://localhost"}, "http"},
		{"type=http makes it http", mcpRaw{Type: "http"}, "http"},
		{"type=sse folds into http", mcpRaw{Type: "sse"}, "http"},
		{"empty raw is stdio", mcpRaw{}, "stdio"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := toServer("test", tt.raw)
			if s.Transport != tt.wantTrans {
				t.Errorf("got %q, want %q", s.Transport, tt.wantTrans)
			}
		})
	}
}
