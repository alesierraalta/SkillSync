# Proposal: Compact Detail View and Cursor Indicator

## Intent
Make detail view more compact and provide visual feedback for focused input.

## Scope
- `internal/ui/view.go`

## Approach
1. Modify `detailView` function.
2. Check `m.inputs[i].Focused()`.
3. Prepend `> ` to focused input string.
4. Reduce `\n` count in `detailView` for compact layout.

## Alternatives
- Use color styling instead of `> ` cursor (might be less accessible).
