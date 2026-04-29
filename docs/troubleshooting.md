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

## Install Does Nothing

- **Cause**: No providers selected or the "Install" action was not explicitly triggered.
- **Solution**: Verify that at least one provider is highlighted and selected (checked) in the TUI Installer screen. Ensure you navigate to the execution button and press `enter`.

## Skill Save/Sync Fails

- **Cause**: No skill is highlighted before pressing `s`, or there are network/permission issues during `S` (Sync).
- **Solution**: Ensure the skill you want to save is highlighted in the list before pressing `s`. For Sync (`S`), verify your internet connection and project write permissions.

## Missing Generated Files

- **Cause**: The installer sync loop skipped specific provider indices or failed to create directories like `.github`.
- **Solution**: Check if you are using the latest version of `synck`. Verify that providers like Codex or Copilot were selected during the install flow. Check for hidden folders like `.github` in your project root.
