# Design: Dynamic Footer Keybindings

## Technical Approach

Implement a dynamic keybinding registry. Replace static hint arrays in `View()` with a `KeyBinder` interface. Screens implement `GetKeyBindings() []KeyBinding`. `Update` handlers and `View` footer consume this registry.

## Architecture Decisions

### Decision: Interface-based Binding Registry

**Choice**: `KeyBinder` interface with `GetKeyBindings() []KeyBinding`.
**Alternatives considered**: Static mapping map, switch-based lookup.
**Rationale**: Interface provides type-safe, per-screen binding definitions, enabling `View` to query bindings directly without hardcoding.

### Decision: Centralized Binding Definition

**Choice**: Define `KeyBinding` struct `{Key string, Description string}`.
**Alternatives considered**: Mapping directly to handler functions.
**Rationale**: `KeyBinding` struct decouples display from action while enforcing single source of truth for key descriptions.

## Data Flow

```
    Update Handler ◄───┐
         │             │
   (KeyMapping)        │
         │             │
    KeyBinder ───────→ Footer View
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/ui/types.go` | Create | Define `KeyBinding` and `KeyBinder` interface. |
| `internal/ui/model.go` | Modify | Update `Model` to support binding lookup. |
| `internal/ui/view.go` | Modify | Update `renderFooter` to use `KeyBinder`. |
| `internal/ui/update.go` | Modify | Update handlers to use defined bindings. |

## Interfaces / Contracts

```go
type KeyBinding struct {
    Key         string
    Description string
}

type KeyBinder interface {
    GetKeyBindings() []KeyBinding
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `GetKeyBindings()` | Verify correct bindings for each screen. |
| Integration | Footer rendering | Check footer matches `GetKeyBindings()`. |

## Migration / Rollout

No migration required.

## Open Questions

- [ ] Does `Update` need to *dynamically* use these bindings, or just *refer* to them? (Plan: refer to them for consistency, keep handler logic distinct for performance).
