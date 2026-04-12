---
trigger: always_on
---

## Dependency Management & Runtime
This project is standardized on **Bun**. To ensure performance consistency and lockfile integrity, avoid using `npm`, `yarn`, or `pnpm` for any package management or execution tasks.

### Core Protocol
* **Package Management:** Use `bun install` for adding or updating dependencies.
* **Script Execution:** Run all project scripts via `bun run <script-name>`.
* **Binary Execution:** Use `bunx` instead of `npx` for one-off command executions.
* **Environment:** Prefer `bun` as the runtime for all TypeScript/JavaScript files.

### Failover & Adaptation
In the event that the environment lacks `bun` or a specific command fails due to runtime incompatibilities:
1. **Identify the Gap:** Determine if it is a missing binary, a version mismatch, or a restricted environment.
2. **Execute Workaround:** Use the most stable alternative available (e.g., `npm` or `node`) only as a temporary bridge to restore functionality.
3. **Document & Update:** Immediately modify this file (`.agents/rules/bun-rule.md`) to include a **"Workarounds"** section. Document the specific error encountered and the exact steps taken to bypass it.

### Verification
Before committing changes to `package.json`, ensure the `bun.lockb` file has been updated and no `package-lock.json` or `yarn.lock` files have been accidentally generated. If they have, remove them before merging.

### Workarounds
* **Missing binary in PATH:** If `bun` or `bunx` is not found, use the full path: `~/.bun/bin/bun` or `~/.bun/bin/bunx`.
* **Environment:** The current shell environment lacks several standard binaries in `$PATH`. Explicit paths are preferred for reliability.

---

> **Status:** Strictly Enforced.
> **Last Checked:** 2026-04-11