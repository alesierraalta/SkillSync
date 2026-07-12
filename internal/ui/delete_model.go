package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"skillsync/tui/internal/remove"
)

// deleteSkillFinishedMsg is sent when the deletion operation completes.
type deleteSkillFinishedMsg struct {
	name string
	err  error
}

// DeleteConfirmModel handles the state for the delete confirmation dialog.
type DeleteConfirmModel struct {
	backend    AppService
	skillName  string
	globalPath string
	confirmed  bool
	deleting   bool
	success    bool
	localOnly  bool
	err        error
}

// NewDeleteConfirmModel creates a new DeleteConfirmModel with the given backend.
func NewDeleteConfirmModel(backend AppService) DeleteConfirmModel {
	return DeleteConfirmModel{
		backend: backend,
	}
}

// Init returns no command for the delete confirm model.
func (m DeleteConfirmModel) Init() tea.Cmd {
	return nil
}

// Update handles key events and async deletion results for the delete confirm dialog.
func (m DeleteConfirmModel) Update(msg tea.Msg) (DeleteConfirmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.confirmed = true
			m.deleting = true
			return m, m.deleteCmd()
		case "n", "N", "esc":
			return m, nil
		}
	case deleteSkillFinishedMsg:
		m.deleting = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.success = true
		}
	}
	return m, nil
}

// deleteCmd returns a tea.Cmd that calls the backend RemoveSkill and returns
// a deleteSkillFinishedMsg with the result.
func (m DeleteConfirmModel) deleteCmd() tea.Cmd {
	return func() tea.Msg {
		var err error
		if m.globalPath != "" {
			err = m.backend.RemoveGlobalSkill(m.globalPath)
		} else {
			opts := remove.Options{Local: m.localOnly}
			err = m.backend.RemoveSkill(m.skillName, opts)
		}
		return deleteSkillFinishedMsg{name: m.skillName, err: err}
	}
}

// View renders the delete confirmation dialog based on current state.
func (m DeleteConfirmModel) View() string {
	if m.deleting {
		return fmt.Sprintf("Deleting skill '%s'...", m.skillName)
	}
	if m.err != nil {
		return fmt.Sprintf("Error deleting skill '%s': %v", m.skillName, m.err)
	}
	if m.success {
		return fmt.Sprintf("Skill '%s' deleted.", m.skillName)
	}
	return fmt.Sprintf("Remove skill '%s'? This cannot be undone. [y/N]", m.skillName)
}
