---
name: git-github-branches
description: >
  Guidelines for working with Git branches, commits, pushes, GitHub PRs, or branch-based development.
  Trigger: When working with Git branches, commits, pushes, GitHub PRs, or branch-based development.
metadata:
  version: '1.0'
  scope: [root]
  auto_invoke:
    - 'Working with Git branches'
    - 'Conventional commits'
    - 'GitHub PR preparation'
    - 'Branch-based development'
---

## When to Use

Use this skill when:
- Creating or managing Git branches
- Writing commit messages
- Preparing a branch for a Pull Request
- Pushing changes to a remote repository
- Working with GitHub PRs and issues

---

## Critical Rules

1. **Branch Naming**: MUST match `^(feat|fix|chore|docs|style|refactor|perf|test|build|ci|revert)\/[a-z0-9._-]+$`.
2. **Conventional Commits**: MUST follow `type(scope): description` format.
3. **No AI Attribution**: NEVER add `Co-Authored-By` or AI attribution trailers.
4. **Git Safety**: No destructive commands or force push unless explicit. NEVER force-push to `main` or `master`.
5. **Issue Linkage**: Every PR MUST link an approved issue (e.g., `Closes #123`).
6. **PR Labels**: Every PR MUST have exactly one `type:*` label (e.g., `type:feature`).

---

## Workflow

### 1. Branch Management
- Check current status and branch: `git status`
- Create branch from `main`: `git checkout -b type/description main`

### 2. Conventional Commits
Format: `type(scope): description`
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Formatting, missing semi colons, etc.
- `refactor`: Refactoring code
- `perf`: Performance improvements
- `test`: Adding missing tests
- `chore`: Maintenance tasks

Example:
```bash
git commit -m "feat(ui): add new navigation menu"
```

### 3. Safety & Verification
Before committing:
- Verify diff: `git diff --cached`
- Run relevant tests or linters.
- Ensure no accidental changes to sensitive files.

Before pushing:
- Check current branch and log: `git log -n 5 --oneline`
- Pull latest changes: `git pull origin main --rebase`

### 4. Pull Requests
- Push to remote: `git push -u origin type/description`
- Create PR using GitHub CLI: `gh pr create --title "type(scope): description" --body "Closes #N"`
- Add label: `gh pr edit --add-label "type:feature"`

---

## Commands

```bash
# Branching
git checkout -b feat/my-feature main

# Committing
git add .
git commit -m "feat(scope): my description"

# Pushing
git push -u origin feat/my-feature

# PRs
gh pr create --title "feat(ui): add menu" --body "Closes #1"
gh pr edit --add-label "type:feature"
```
