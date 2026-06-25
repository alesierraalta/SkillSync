# Spec: Installer Cursor Rendering

## Issue Description
In `internal/ui/installer_model.go`, the navigation cursor highlights both the "Add shell aliases to profile" option and the "[ Execute Install ]" button simultaneously when the cursor is positioned at `9+storageOffset`. This is due to a copy-paste error where `cursorAction` was checking if `m.Cursor == 9+storageOffset` instead of `10+storageOffset`.

## Requirements
- The installer view MUST render the selection cursor (`> `) next to the "Add shell aliases to profile" option when `m.Cursor` is exactly `9+storageOffset`.
- The installer view MUST render the selection cursor (`> `) next to the "[ Execute Install ]" button when `m.Cursor` is exactly `10+storageOffset`.
- The installer view MUST NOT show duplicate selection cursors for these options.

## Scenarios

### Scenario 1: Cursor positioned on Add shell aliases to profile
- **Given** an installer model with `len(m.AllStored)` skills (represented by `storageOffset`)
- **When** `m.Cursor` is equal to `9 + storageOffset`
- **Then** the rendered options view MUST display `> [x] Add shell aliases to profile` (or `[ ]` depending on selection state)
- **And** the rendered options view MUST display `  [ Execute Install ]` (with no selection cursor)

### Scenario 2: Cursor positioned on Execute Install button
- **Given** an installer model with `len(m.AllStored)` skills (represented by `storageOffset`)
- **When** `m.Cursor` is equal to `10 + storageOffset`
- **Then** the rendered options view MUST display `  [x] Add shell aliases to profile` (or `[ ]` depending on selection state, with no selection cursor)
- **And** the rendered options view MUST display `> [ Execute Install ]`
