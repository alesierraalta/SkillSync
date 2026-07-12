---
managed_by: skillsync
description: Syncs skill metadata to AGENTS.md Auto-invoke sections. Trigger: When updating skill metadata (metadata.scope/metadata.auto_invoke), regenerating Auto-invoke tables, or running ./.agent/skills/skill-sync/assets/sync.sh (including --dry-run/--scope).

---

/synck skill-sync "$!1"
