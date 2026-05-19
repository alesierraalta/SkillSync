---
name: context7
description: >
  Disciplined use of Context7 docs lookup: resolve library ID first, query docs with precise questions, version selection, usage limits, not sending secrets, when to use vs local docs/web search.
  Trigger: looking for documentation for external libraries, framework features, or API references.
metadata:
  version: "1.0"
  scope: [root]
  auto_invoke: "Searching external documentation"
---

# Context7 Documentation Lookup

## When to Use

Use this skill when:
- Needing up-to-date documentation for a specific library or framework.
- Looking for code examples for external APIs.
- Comparing different versions of a library's documentation.
- Troubleshooting errors related to external dependencies.

---

## Critical Patterns

### Pattern 1: Resolve Library ID First
You MUST call `resolve-library-id` before `query-docs` unless the user provides an ID in `/org/project` format. This ensures you target the correct documentation set.

### Pattern 2: Specific and Precise Queries
Avoid vague queries like "auth". Use specific tasks: "How to set up JWT authentication in Express.js".

### Pattern 3: Version Selection
If the user specifies a version, use it. When multiple versions are available from `resolve-library-id`, pick the one that matches the project's dependencies.

### Pattern 4: Safety and Limits
- NEVER include secrets, API keys, or proprietary code in queries.
- Do not call `query-docs` more than 3 times per question to respect usage limits.

### Pattern 5: Tool Choice
Prefer `context7` for specialized library docs. Use `web_search` for general questions or `read_file` for local documentation.

---

## Decision Tree

```text
Need library-specific docs? → Resolve ID → Query Docs
User provided /org/project ID? → Query Docs directly
Querying general knowledge? → Use Web Search / Internet Search
Need local codebase context? → Use Grep / Glob / Read
```

---

## Code Examples

### Example 1: Resolving and Querying

```javascript
// Step 1: Resolve
await context7_resolve_library_id({ libraryName: "Next.js", query: "middleware" });

// Step 2: Query
await context7_query_docs({ libraryId: "/vercel/next.js", query: "How to use middleware for auth" });
```

---

## Commands

```bash
# No specific shell commands, but use the tools via the MCP interface.
```
