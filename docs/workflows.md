# Standard Workflows

To maintain a high quality of code and documentation, we follow specific workflows.

## SDD (Spec-Driven Development)

All significant changes should follow the SDD cycle:
1. **Explore**: Investigate the codebase and requirements.
2. **Propose**: Define the intent and scope of the change.
3. **Spec**: Write detailed functional requirements (delta specs).
4. **Design**: Plan the technical implementation.
5. **Tasks**: Break down the work into a checklist.
6. **Apply**: Implement the changes.
7. **Verify**: Test against the specs.
8. **Archive**: Persist the final state and cleanup.

## Token Optimization with RTK

We use `rtk` (Rust Token Killer) to keep our AI context clean and efficient.
- It automatically injects instructions into `CLAUDE.md`.
- It helps manage long conversations by suggesting compactions when the token limit is approached.

## Skill Synchronization

Whenever a skill's metadata or instructions are updated:
1. Run the TUI and press `y`.
2. This triggers the `skill-sync` skill.
3. Metadata is extracted from all `SKILL.md` files.
4. `AGENTS.md` is updated with fresh tables and auto-invoke rules.
5. The `skills-lock.json` file is updated to track versions and locations.

## Global Skill Storage

Skillsync allows you to save and reuse skills across different projects.

### Workflow
1. **Save a Skill**: In the TUI list view, highlight a skill and press `s` (lowercase). This saves the skill and its metadata to `~/.skillsync/storage`.
2. **Sync Storage**: Press `y` while in the Storage screen to synchronize your local project with the global storage.
3. **Install from Storage**: Navigate to the "Almacenamiento de skills" section in the TUI to browse and install skills previously saved from other projects.
