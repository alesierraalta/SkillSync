package ui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"skillsync/tui/internal/agentdetect"
	"skillsync/tui/internal/bundle"
	"skillsync/tui/internal/remove"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/syncengine"
	"skillsync/tui/internal/types"
)

// fakeAppService is a minimal AppService stub for Global Skills behavior
// tests. It only implements the surface loadGlobalSkillsCmd / delete / filter
// paths actually exercise; everything else panics if called.
type fakeAppService struct {
	discoverPaths []string
	discoverErr   error
	parseErr      error

	removedPaths []string
	removeErr    error

	exportedNames []string
	exportedDest  string
	exportPath    string
	exportErr     error

	importedPath  string
	importRoot    string
	importResults []bundle.ImportResult
	importErr     error
}

func (f *fakeAppService) DiscoverSkills(string) ([]string, error) {
	return f.discoverPaths, f.discoverErr
}

func (f *fakeAppService) ParseSkill(path string) (*types.Skill, error) {
	if f.parseErr != nil {
		return nil, f.parseErr
	}
	return &types.Skill{
		ID:   path,
		Name: filepath.Base(filepath.Dir(path)),
		Path: path,
		Metadata: types.Metadata{
			Description: "fake",
		},
	}, nil
}

func (f *fakeAppService) RemoveGlobalSkill(path string) error {
	f.removedPaths = append(f.removedPaths, path)
	return f.removeErr
}

// Stubs for the rest of the AppService surface — never reached by these tests.
func (f *fakeAppService) ScanProjects([]string, int) ([]string, error) { panic("unused") }
func (f *fakeAppService) ParseSkillContent(string) (*types.Skill, error) {
	panic("unused")
}
func (f *fakeAppService) SaveSkill(string, *types.Skill) error { panic("unused") }
func (f *fakeAppService) Sync(string, syncengine.SyncOptions) (*runner.SyncReport, error) {
	return nil, errors.New("unused")
}
func (f *fakeAppService) RegisterProject(string) error        { panic("unused") }
func (f *fakeAppService) RegisterProjectInitial(string) error { panic("unused") }
func (f *fakeAppService) ListStoredSkills() ([]storage.StoredSkill, error) {
	return nil, nil
}
func (f *fakeAppService) GetProjects() ([]storage.ProjectInfo, error) {
	return nil, nil
}
func (f *fakeAppService) SaveToStorage(*types.Skill, storage.StoredMetadata) error {
	return nil
}
func (f *fakeAppService) LoadFromStorage(string) (string, error) { return "", nil }
func (f *fakeAppService) CopyStorageExtras(string, string) error { return nil }
func (f *fakeAppService) InstallCoreSkill(string) error          { return nil }
func (f *fakeAppService) RegisterOpenCodeTools() error           { return nil }
func (f *fakeAppService) RegisterSkillManagerAgent() error       { return nil }
func (f *fakeAppService) EnsureAgentsMD(string) error            { return nil }
func (f *fakeAppService) RemoveSkill(string, remove.Options) error {
	return nil
}
func (f *fakeAppService) ExportBundle(names []string, destPath string) (string, error) {
	f.exportedNames = names
	f.exportedDest = destPath
	if f.exportPath != "" {
		return f.exportPath, f.exportErr
	}
	return destPath, f.exportErr
}
func (f *fakeAppService) ImportBundle(bundlePath, projectRoot string) ([]bundle.ImportResult, error) {
	f.importedPath = bundlePath
	f.importRoot = projectRoot
	return f.importResults, f.importErr
}
func (f *fakeAppService) DetectAgentEcosystem() ([]agentdetect.AgentInfo, error) {
	return nil, nil
}

// listModelSelect wraps the list.Model Select call so the test file does
// not need a list import alias.
func listModelSelect(l *list.Model, idx int) {
	l.Select(idx)
}

// executeCmd runs a tea.Cmd and returns the resulting message, panicking on
// anything other than a single non-nil message (matches the
// loadGlobalSkillsCmd contract: one globalSkillsLoadedMsg or
// globalSkillsErrorMsg).
func executeCmd(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	if cmd == nil {
		t.Fatal("cmd is nil")
	}
	return cmd()
}

func TestGlobalSkills_LoadAll_Success(t *testing.T) {
	// Simulate a HOME with three provider skills (Antigravity is under
	// .gemini/antigravity/ — the most common mistake is filtering by
	// ".antigravity" which never matches).
	tmpHome := t.TempDir()
	origHome, hadHome := os.LookupEnv("HOME")
	t.Setenv("HOME", tmpHome)
	if hadHome {
		_ = origHome
	}

	paths := []string{
		filepath.Join(tmpHome, ".claude", "skills", "claude-skill", "SKILL.md"),
		filepath.Join(tmpHome, ".opencode", "skills", "opencode-skill", "SKILL.md"),
		filepath.Join(tmpHome, ".gemini", "antigravity", "skills", "ag-skill", "SKILL.md"),
		filepath.Join(tmpHome, ".qwen", "skills", "qwen-skill", "SKILL.md"),
	}

	backend := &fakeAppService{discoverPaths: paths}
	m := NewModel(backend)

	msg := executeCmd(t, m.loadGlobalSkillsCmd("All"))
	loaded, ok := msg.(globalSkillsLoadedMsg)
	if !ok {
		t.Fatalf("expected globalSkillsLoadedMsg, got %T", msg)
	}
	if len(loaded.items) != 4 {
		t.Errorf("expected 4 skills across all providers, got %d", len(loaded.items))
	}
}

func TestGlobalSkills_FilterAntigravity_MatchesGeminiAntigravity(t *testing.T) {
	// Antigravity skills live at ~/.gemini/antigravity/<name>/SKILL.md, NOT
	// at ~/.antigravity/. A naive `strings.Contains(path, "/.antigravity/")`
	// filter returns nothing — guard the regression here.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	ag := filepath.Join(tmpHome, ".gemini", "antigravity", "skills", "ag-skill", "SKILL.md")
	other := filepath.Join(tmpHome, ".claude", "skills", "claude-skill", "SKILL.md")

	backend := &fakeAppService{discoverPaths: []string{ag, other}}
	m := NewModel(backend)

	msg := executeCmd(t, m.loadGlobalSkillsCmd("Antigravity"))
	loaded, ok := msg.(globalSkillsLoadedMsg)
	if !ok {
		t.Fatalf("expected globalSkillsLoadedMsg, got %T", msg)
	}
	if len(loaded.items) != 1 {
		t.Fatalf("expected exactly 1 Antigravity skill, got %d", len(loaded.items))
	}
	// Path-aware assertion: on Windows, paths use \ as separator; normalize
	// to forward slashes so the suffix check is portable.
	normalized := filepath.ToSlash(loaded.items[0].skill.Path)
	if !strings.HasSuffix(normalized, "ag-skill/SKILL.md") {
		t.Errorf("expected the .gemini/antigravity skill, got %s", loaded.items[0].skill.Path)
	}
}

func TestGlobalSkills_FilterClaude_ExcludesOtherProviders(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	claude := filepath.Join(tmpHome, ".claude", "skills", "claude-skill", "SKILL.md")
	opencode := filepath.Join(tmpHome, ".opencode", "skills", "opencode-skill", "SKILL.md")

	backend := &fakeAppService{discoverPaths: []string{claude, opencode}}
	m := NewModel(backend)

	msg := executeCmd(t, m.loadGlobalSkillsCmd("Claude"))
	loaded, ok := msg.(globalSkillsLoadedMsg)
	if !ok {
		t.Fatalf("expected globalSkillsLoadedMsg, got %T", msg)
	}
	if len(loaded.items) != 1 {
		t.Fatalf("expected exactly 1 Claude skill, got %d", len(loaded.items))
	}
	// Path-aware assertion: on Windows, paths use \ as separator; normalize
	// to forward slashes so the suffix check is portable.
	normalized := filepath.ToSlash(loaded.items[0].skill.Path)
	if !strings.HasSuffix(normalized, "claude-skill/SKILL.md") {
		t.Errorf("expected the .claude skill, got %s", loaded.items[0].skill.Path)
	}
}

func TestGlobalSkills_DiscoveryError_SurfacesInUI(t *testing.T) {
	// A discovery failure must transition the screen out of the loading state
	// (otherwise the user is stuck on "Buscando..." forever) AND carry the
	// error in the model so the view can render it.
	backend := &fakeAppService{discoverErr: errors.New("home dir unreadable")}
	m := NewModel(backend)
	m.Screen = ScreenGlobalSkillsList
	m.globalCategory = "Claude"

	// Fire the cmd and dispatch the message through Update.
	cmd := m.loadGlobalSkillsCmd("Claude")
	newModel, _ := m.Update(executeCmd(t, cmd))
	m = newModel.(Model)

	if !m.globalSkillsLoaded {
		t.Fatal("expected globalSkillsLoaded=true after error, still false (loading state)")
	}
	if m.globalSkillsErr == nil {
		t.Fatal("expected globalSkillsErr to be set, got nil")
	}
	if !strings.Contains(m.globalSkillsErr.Error(), "home dir unreadable") {
		t.Errorf("expected error to surface, got %v", m.globalSkillsErr)
	}

	// The view must render the error and not the loading string.
	view := m.View()
	if strings.Contains(view, "Buscando...") {
		t.Errorf("error view should not show loading state, got:\n%s", view)
	}
	if !strings.Contains(view, "home dir unreadable") {
		t.Errorf("error view should show error message, got:\n%s", view)
	}
}

func TestGlobalSkills_DeleteGlobalSkill_DispatchesToBackend(t *testing.T) {
	// 'd' on the global list must populate the DeleteConfirmModel with the
	// selected skill's path so the eventual RemoveGlobalSkill call hits the
	// filesystem path the user actually selected.
	backend := &fakeAppService{}
	m := NewModel(backend)
	m.Screen = ScreenGlobalSkillsList
	m.globalSkillsLoaded = true
	agPath := filepath.FromSlash("/home/u/.gemini/antigravity/skills/ag-skill/SKILL.md")
	m.globalSkillsList.SetItems([]list.Item{
		globalSkillItem{
			skill:    types.Skill{Name: "ag-skill", Path: agPath},
			category: "Antigravity",
		},
	})
	listModelSelect(&m.globalSkillsList, 0)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	res := updated.(Model)
	if res.Screen != ScreenDeleteConfirm {
		t.Fatalf("expected ScreenDeleteConfirm after 'd', got %d", res.Screen)
	}
	if res.deleteConfirm.globalPath != agPath {
		t.Errorf("expected deleteConfirm.globalPath=%q, got %q", agPath, res.deleteConfirm.globalPath)
	}
	if res.deleteConfirm.skillName != "ag-skill" {
		t.Errorf("expected deleteConfirm.skillName=ag-skill, got %q", res.deleteConfirm.skillName)
	}
}

func TestGlobalSkills_DeleteSuccess_ReloadsList(t *testing.T) {
	// After a successful global delete, the model must return to the
	// GlobalSkillsList screen, set globalSkillsLoaded=false so a reload
	// happens, and the backend must have been asked to remove the selected
	// path.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	backend := &fakeAppService{}
	m := NewModel(backend)
	m.Screen = ScreenDeleteConfirm
	m.PrevScreen = ScreenGlobalSkillsList
	m.globalCategory = "Antigravity"
	agPath := filepath.Join(tmpHome, ".gemini", "antigravity", "skills", "ag-skill", "SKILL.md")
	m.deleteConfirm.skillName = "ag-skill"
	m.deleteConfirm.globalPath = agPath

	// Step 1: user presses 'y' on the confirm screen. The returned cmd
	// captures the delete; we execute it inline (real DeleteConfirmModel
	// hands the cmd to the bubbletea runtime).
	updated, deleteCmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	afterConfirm := updated.(Model)
	if !afterConfirm.deleteConfirm.deleting {
		t.Fatal("expected deleteConfirm.deleting=true after y")
	}
	if deleteCmd == nil {
		t.Fatal("expected deleteCmd to be returned after y")
	}

	// Step 2: run the deleteCmd; it should call RemoveGlobalSkill and
	// return a deleteSkillFinishedMsg.
	finishedMsg, ok := deleteCmd().(deleteSkillFinishedMsg)
	if !ok {
		t.Fatalf("expected deleteSkillFinishedMsg from deleteCmd, got %T", finishedMsg)
	}
	if finishedMsg.err != nil {
		t.Fatalf("expected nil err, got %v", finishedMsg.err)
	}
	if len(backend.removedPaths) != 1 || backend.removedPaths[0] != agPath {
		t.Errorf("expected backend to receive RemoveGlobalSkill(%q), got %v",
			agPath, backend.removedPaths)
	}

	// Step 3: dispatch the finished message back through the parent Update.
	// On success we return to GlobalSkillsList with globalSkillsLoaded=false
	// (so the next render triggers a reload) and a status message naming the
	// deleted skill.
	finalModel, _ := afterConfirm.Update(finishedMsg)
	res := finalModel.(Model)
	if res.Screen != ScreenGlobalSkillsList {
		t.Errorf("expected return to ScreenGlobalSkillsList, got %d", res.Screen)
	}
	if res.globalSkillsLoaded {
		t.Error("expected globalSkillsLoaded=false so the next render triggers a reload")
	}
	if !strings.Contains(res.StatusMsg, "ag-skill") {
		t.Errorf("expected StatusMsg to mention deleted skill, got %q", res.StatusMsg)
	}
}

func TestGlobalSkills_DeleteError_StaysOnConfirmWithError(t *testing.T) {
	// A failed delete must keep the user on the confirm screen so they can
	// read the error and retry, NOT bounce them back to an empty list.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	backend := &fakeAppService{removeErr: errors.New("permission denied")}
	m := NewModel(backend)
	m.Screen = ScreenDeleteConfirm
	m.PrevScreen = ScreenGlobalSkillsList
	m.globalCategory = "Antigravity"
	agPath := filepath.Join(tmpHome, ".gemini", "antigravity", "skills", "ag-skill", "SKILL.md")
	m.deleteConfirm.skillName = "ag-skill"
	m.deleteConfirm.globalPath = agPath

	updated, deleteCmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	afterConfirm := updated.(Model)
	if deleteCmd == nil {
		t.Fatal("expected deleteCmd to be returned after y")
	}

	finishedMsg, ok := deleteCmd().(deleteSkillFinishedMsg)
	if !ok {
		t.Fatalf("expected deleteSkillFinishedMsg, got %T", finishedMsg)
	}
	if finishedMsg.err == nil {
		t.Fatal("expected non-nil err from failing RemoveGlobalSkill")
	}

	finalModel, _ := afterConfirm.Update(finishedMsg)
	res := finalModel.(Model)

	if res.Screen != ScreenDeleteConfirm {
		t.Errorf("expected to stay on ScreenDeleteConfirm on error, got %d", res.Screen)
	}
	if res.deleteConfirm.err == nil {
		t.Error("expected deleteConfirm.err to be set on failure")
	}
}
