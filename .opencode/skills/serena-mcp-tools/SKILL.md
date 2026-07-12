---
name: serena-mcp-tools
description: >
  Serena MCP for semantic navigation, symbol-based editing, and low-token-cost
  refactors. Trigger: using Serena MCP, searching declarations, references,
  implementations, diagnostics, or editing code by symbol.
metadata:
  author: a.sierra
  version: "1.2"
  scope: [root]
  auto_invoke: "Using Serena MCP, semantic code navigation, symbol editing, or refactors"
allowed-tools: Read, Edit, Write, Glob, Grep, Bash, Task, Serena MCP, Engram, Context7, delegate
---

# Serena MCP Tools

## Mission

Use Serena as a semantic IDE layer: declarations, symbols, references, implementations, diagnostics, and real language refactors.

Do not use it as a universal replacement for `grep`, shell, git, direct reading, or small patches. That is the trap: a powerful tool misused burns tokens, time, and trust.

## Decision Gate

| Question | Yes | No |
| --- | --- | --- |
| Do I need to understand what an identifier means in the language? | Serena | `grep`/direct reading |
| Do I need real references, overrides, or implementations? | Serena | `grep` if it is literal text |
| Does the change depend on a class, method, function, interface, or type? | Serena | direct patch |
| Is it Markdown, JSON, YAML, TOML, `.env`, CI, Dockerfile, or docs? | native tool | Serena only if it contains code with symbols |
| Do I already know the file and it is 1-5 local lines? | direct patch | Serena if there is semantic impact |
| Is it shell, test, build, lint, format, or git? | native shell | no Serena |

Gentle AI Rule: use Serena when it reduces semantic uncertainty. If it only reduces laziness, NO.

## Token Harness

1. Define the semantic goal: symbol, probable file, container, signature, or module.
2. Map before editing: overview, symbol, declaration, or references.
3. Read the minimum: symbol body before the entire file.
4. Edit by symbol if the insertion point or rename is semantic.
5. Verify with diff, diagnostics, and relevant tests.
6. If Serena returns noise, ambiguity, or broad changes, fall back to `grep` + controlled patch.

## Tool Routing

| Need | Preferred Tool |
| --- | --- |
| Top-level map of unknown file | `serena_get_symbols_overview` |
| Find class, function, method, type, or interface | `serena_find_symbol` |
| Jump from use to definition | `serena_find_declaration` |
| Find implementations of abstractions | `serena_find_implementations` |
| Measure real impact of a change | `serena_find_referencing_symbols` |
| Insert near a known entity | `serena_insert_before_symbol` / `serena_insert_after_symbol` |
| Replace a localized function/class | `serena_replace_symbol_body` |
| Rename a real language entity | `serena_rename_symbol` |
| Review language errors | `serena_get_diagnostics_for_file` |
| Literal text, logs, strings, paths | `grep` / `glob` |
| Tests, build, lint, format, git | native `bash` |
| Docs or config | native read/patch |

## Operating Harness

### Bootstrap

- Call `serena_initial_instructions` once per session if you will use Serena.
- Activate the project with `serena_activate_project` when unclear.
- Run `serena_check_onboarding_performed` before relying on repo knowledge.
- If JetBrains is the backend, confirm that the IDE root opened matches the active Serena root.

### Explore

- Start with `get_symbols_overview` on candidate files.
- Use `find_symbol` with low `max_matches` when the name is ambiguous.
- Request `include_body: false` until you need the implementation.
- Use `find_referencing_symbols` before touching public APIs, shared utilities, or interfaces.

### Edit

- Prefer `rename_symbol` for real renames, not textual search.
- Prefer `insert_before_symbol` or `insert_after_symbol` when the insertion point is stable by identity.
- Use `replace_symbol_body` only after exactly locating the correct symbol.
- Do not mix rename, move, and behavior change in a single operation.

### Verify

- Review `git diff --stat` and `git diff -- <files>`.
- Run Serena diagnostics on edited files if the backend is reliable.
- Run minimal relevant tests from the stack.
- If you cannot run tests, state it explicitly and limit the assertion to diff/diagnostics/analysis.

## Task Playbooks

### Understanding a Function or Class

1. Use `find_symbol` to locate it.
2. Read only the signature/body you need.
3. Use `find_referencing_symbols` if impact matters.
4. Explain invariants, dependencies, and risks.

### Changing a Public API

1. Use `find_symbol` or `find_declaration` for the main symbol.
2. Use `find_referencing_symbols` for call sites.
3. Use `rename_symbol` only if it is a pure rename.
4. Edit definition and usages in small units.
5. Run diagnostics and specific tests.

### Adding Functionality

1. Use `get_symbols_overview` on candidate files.
2. Locate existing patterns with `find_symbol` and complementary `grep`.
3. Insert by symbol when the point is semantic.
4. Use native tools for tests, snapshots, docs, and config.

### Large Refactor

1. Map symbols and references before editing.
2. Divide by semantic unit: class, method, module, or interface.
3. Apply one semantic operation at a time.
4. Verify diff and diagnostics between steps.
5. Stop if the backend is not indexed or results are unreliable.

## Stop Conditions

Stop Serena and change strategy if:

- The symbol does not resolve or resolves another entity.
- The result brings too much irrelevant context.
- The backend/LSP is not indexed or its diagnostics are questionable.
- The edit touches unrelated files.
- The task is literal text, configuration, documentation, or git.
- The MCP client does not discover Serena tools consistently.

## Safety Rules

- Do not execute shell via Serena if the agent already has auditable shell.
- Do not activate beta/optional tools unless there is concrete need.
- Treat hooks, scripts, generated files, and dependencies as execution surfaces.
- Keep Serena HTTP transports on localhost unless explicitly justified.
- If the client does not ask for confirmation to execute tools, avoid write or shell operations.

## Verification Commands

```bash
git diff --stat
git diff -- <modified_files>
```

```bash
# Adjust to the real stack.
npm test -- --runInBand
pytest -q
go test ./...
cargo test
mvn test
```

## Agent Prompts

### Navigation

> Use Serena to locate real declaration and references for `<symbol>`. Do not read entire files unless the symbol body is insufficient. Return impact by file and containing symbol.

### Refactor

> Use Serena to rename `<old_symbol>` to `<new_symbol>` as a semantic symbol, not as textual search. Then review diff, diagnostics, and relevant tests. If there is ambiguity, stop before editing.

### Impact Audit

> Use Serena to find implementations and references for `<symbol>`. Classify uses by read, write, override, direct call, indirect call, or test. Do not modify code.

### Safe Insertion

> Use Serena to insert the method/helper next to the closest related symbol. Do not change imports or call sites until you verify overview and references.

## Source Notes

- Serena provides recovery, editing, refactoring, and semantic diagnostics at the symbol level via MCP.
- Its tools may be partially enabled depending on context/mode; do not assume universal availability.
- By default it uses language servers; with JetBrains the backend is decided on startup and the IDE root must match the Serena root.
- The `claude-code`, `codex`, and `ide` contexts are designed to not duplicate native client/agent capabilities.
- Primary documentation: `https://oraios.github.io/serena/`.

## Maintenance

Review this skill when MCP tools, Serena backend, MCP client, or project topology change. Keep stable the central decision: Serena for semantics; native tools for text, shell, git, config, and small patches.
