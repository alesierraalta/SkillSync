# Troubleshooting

Common issues and how to solve them.

## TUI Crashes on Startup

- **Cause**: Corrupted `skills-lock.json` or malformed YAML in a `SKILL.md` file.
- **Solution**: Check the console output for parsing errors. You can try deleting `skills-lock.json` and restarting to force a full re-discovery.

## Skill Not Showing Up

- **Cause**: `SKILL.md` is missing the required YAML frontmatter or is in an unmonitored directory.
- **Solution**: Ensure your skill is inside a directory with a `SKILL.md` file and that the metadata section is correctly formatted. Run `./scripts/sync.sh` to check for discovery errors.

## Sync Fails

- **Cause**: Permission issues on `AGENTS.md` or a syntax error in the template.
- **Solution**: Verify that `AGENTS.md` is not being held open by another process and that you have write permissions in the root directory.

## Sandbox Errors (Access Denied)

- **Cause**: The skill is trying to access a file outside the project root or a hidden system directory.
- **Solution**: Adjust the skill's instructions or the input path to stay within the workspace boundaries. Check `--verbose` logs in `skill-sandbox`.
