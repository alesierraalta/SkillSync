# Proposal: Footer Keybindings Expansion

## Intent

Footer currently hardcoded in `internal/ui/view.go`. Inflexible, hard to maintain. Need dynamic keybinding rendering based on active UI state.

## Scope

### In Scope
- Define `KeyBinding` struct (key, action).
- Implement `GetKeyBindings()` method in `Model` to return bindings for current `Screen`.
- Refactor `renderFooter()` in `view.go` to use dynamic bindings.
- Update `handleComponentUpdate` or similar to centralize key mappings.

### Out of Scope
- Global key remapping engine.
- UI theme changes.

## Approach
1. Add `KeyBinding` type: `type KeyBinding struct { Key, Action string }`.
2. Add `func (m Model) GetKeyBindings() []KeyBinding` to `model.go` returning slice based on `m.Screen`.
3. Update `view.go` `renderFooter()` to map `[]KeyBinding` to `key: action` string.
4. Ensure `handle...` functions in `update.go` maintain consistency.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/ui/model.go` | New | Define `KeyBinding` struct. |
| `internal/ui/view.go` | Modified | Update `renderFooter()` to use dynamic bindings. |

## Risks
| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Inconsistency between handling and footer | Med | Unit tests verifying `GetKeyBindings()` aligns with actual key handlers. |

## Rollback Plan
Revert changes to `view.go` `renderFooter()` and remove new `KeyBinding` definitions in `model.go`.

## Dependencies
None.

## Success Criteria
- [ ] Footer renders bindings correctly on all screens.
- [ ] Adding/removing bindings in `GetKeyBindings()` automatically updates footer.
- [ ] Automated tests verify footer output matches expected keys per screen.
