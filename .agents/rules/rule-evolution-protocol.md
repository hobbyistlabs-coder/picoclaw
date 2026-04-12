---
trigger: always_on
---

## Agent Autonomy & Rule Evolution
This project uses a self-documenting workflow to align AI behavior with developer preferences. If you identify a recurring pattern, preference, or specific technical requirement in how work is performed, you are authorized to codify it.

### Rule Creation Trigger
* **Pattern Recognition:** If you notice a specific preference for library usage, code styling, or architectural patterns not already documented.
* **Workflow Refinement:** If a specific sequence of commands or a "way of doing things" leads to better results.
* **Correction Loops:** If I have to correct your approach more than once for the same type of task.

### Implementation Protocol
1. **Path:** All rules must be stored in `.agents/rules/`.
2. **Naming:** Use `kebab-case` and a descriptive prefix (e.g., `ui-styling-standards.md`, `error-handling-patterns.md`).
3. **Format:** Use Markdown with clear headers, bolded key actions, and code blocks for command examples.
4. **Action:**
   * Create the file: `touch .agents/rules/new-rule-name.md`
   * Populate it with the logic, "why" it exists, and "how" to execute it.
   * Notify me that a new rule has been established so I can review or adjust it.

### Maintenance
* **Avoid Redundancy:** Before creating a new rule, check existing files in `.agents/rules/` to see if an update to an existing policy is more appropriate than a new file.
* **Strictness:** Once a rule is written to this directory, it is considered a "Hard Constraint" for all future sessions.

> **Current Rule Inventory:** Run `ls .agents/rules/` to see the current active constraints.