package ui

import (
	"fmt"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/syncengine"
	"strings"
	"testing"
)

func TestSyncReportModelFoundation(t *testing.T) {
	// Task 1.1: syncReportMsg should exist
	_ = syncReportMsg{
		report: &runner.SyncReport{},
		err:    nil,
	}

	// Task 1.2: Model should have syncReport field
	m := Model{}
	m.syncReport = &runner.SyncReport{}

	if m.syncReport == nil {
		t.Error("syncReport field should be assignable")
	}
}

func TestSyncReportUpdateHandler(t *testing.T) {
	m := Model{
		SyncFinished: false,
	}

	report := &runner.SyncReport{
		Changes: []runner.FileChange{
			{Path: "test.txt", Status: "modified"},
		},
	}

	msg := syncReportMsg{
		report: report,
		err:    nil,
	}

	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if !updatedModel.SyncFinished {
		t.Error("expected SyncFinished to be true")
	}

	if updatedModel.syncReport != report {
		t.Error("expected syncReport to be updated")
	}

	if updatedModel.err != nil {
		t.Error("expected err to be nil")
	}

	t.Run("Failure", func(t *testing.T) {
		m := Model{
			SyncFinished: false,
			SyncFailed:   false,
		}

		syncErr := fmt.Errorf("sync failed")
		msg := syncReportMsg{
			report: nil,
			err:    syncErr,
		}

		newModel, _ := m.Update(msg)
		updatedModel := newModel.(Model)

		if !updatedModel.SyncFinished {
			t.Error("expected SyncFinished to be true even on failure")
		}

		if !updatedModel.SyncFailed {
			t.Error("expected SyncFailed to be true")
		}

		if updatedModel.err != syncErr {
			t.Errorf("expected err to be %v, got %v", syncErr, updatedModel.err)
		}
	})
}

func TestStartSyncCommand(t *testing.T) {
	mockBackend := &MockAppService{
		SyncFunc: func(root string, opts syncengine.SyncOptions) (*runner.SyncReport, error) {
			return &runner.SyncReport{}, nil
		},
	}
	m := NewModel(mockBackend)
	m.rootPath = "."

	cmd := m.startSync()
	msg := cmd()

	_, ok := msg.(syncReportMsg)
	if !ok {
		t.Errorf("expected msg of type syncReportMsg, got %T", msg)
	}

	t.Run("Failure", func(t *testing.T) {
		syncErr := fmt.Errorf("sync error")
		mockBackend := &MockAppService{
			SyncFunc: func(root string, opts syncengine.SyncOptions) (*runner.SyncReport, error) {
				return nil, syncErr
			},
		}

		m := NewModel(mockBackend)
		m.rootPath = "."
		cmd := m.startSync()
		msg := cmd()

		res, ok := msg.(syncReportMsg)
		if !ok {
			t.Fatalf("expected msg of type syncReportMsg, got %T", msg)
		}

		if res.err != syncErr {
			t.Errorf("expected error %v, got %v", syncErr, res.err)
		}
	})
}

func TestSyncingViewRendering(t *testing.T) {
	t.Run("Success Report", func(t *testing.T) {
		m := NewModel(NewBackend(storage.NewService("")))
		m.Screen = ScreenSyncing
		m.SyncFinished = true
		m.syncReport = &runner.SyncReport{
			Changes: []runner.FileChange{
				{Path: "AGENTS.md", Status: "modified"},
			},
		}

		view := m.syncingView()
		if !strings.Contains(view, "Skills synced: 1") {
			t.Errorf("expected view to contain 'Skills synced: 1', got:\n%s", view)
		}
	})

	t.Run("Success Report Multiple Changes", func(t *testing.T) {
		m := NewModel(NewBackend(storage.NewService("")))
		m.Screen = ScreenSyncing
		m.SyncFinished = true
		m.syncReport = &runner.SyncReport{
			Changes: []runner.FileChange{
				{Path: "AGENTS.md", Status: "modified"},
				{Path: "CLAUDE.md", Status: "created"},
			},
		}

		view := m.syncingView()
		if !strings.Contains(view, "Skills synced: 2") {
			t.Errorf("expected view to contain 'Skills synced: 2', got:\n%s", view)
		}
		if !strings.Contains(view, "AGENTS.md (modified)") {
			t.Errorf("expected view to contain AGENTS.md, got:\n%s", view)
		}
		if !strings.Contains(view, "CLAUDE.md (created)") {
			t.Errorf("expected view to contain CLAUDE.md, got:\n%s", view)
		}
	})

	t.Run("Failure Error", func(t *testing.T) {
		m := NewModel(NewBackend(storage.NewService("")))
		m.Screen = ScreenSyncing
		m.SyncFinished = true
		m.SyncFailed = true
		m.err = fmt.Errorf("critical sync error")

		view := m.syncingView()
		if !strings.Contains(view, "Error Details:") {
			t.Errorf("expected view to contain 'Error Details:', got:\n%s", view)
		}
		if !strings.Contains(view, "critical sync error") {
			t.Errorf("expected view to contain error message, got:\n%s", view)
		}
	})
}
