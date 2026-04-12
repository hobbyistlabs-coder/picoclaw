---
trigger: always_on
---

## Codebase Navigation & Comprehension
This project uses `roam` as the primary tool for codebase comprehension. Use `roam` for all exploration tasks; traditional Glob, Grep, or manual Read operations should only be used if `roam` is unable to provide the necessary context.

### Initialization
Run `roam index` at the beginning of **every session** to ensure the mental model of the codebase is fresh and accurate.

### Fallback Protocol
If the `roam` command is not recognized or fails, execute it using the explicit binary path:
`/Users/wadahadlan/Library/Python/3.9/bin/roam <command>`

### Operational Workflow
Before modifying any code, follow these steps:
1.  **Onboarding:** If it's your first time in the repo, run `roam understand` followed by `roam tour`.
2.  **Discovery:** Find symbols or patterns using `roam search <pattern>`.
3.  **Risk Assessment:** Before changing a symbol, run `roam preflight <name>` to analyze blast radius, tests, and fitness.
4.  **Context Gathering:** Use `roam context <name>` to identify relevant files and prioritized line ranges.
5.  **Troubleshooting:** Debug failures using `roam diagnose <name>` for root cause ranking.
6.  **Verification:** After making changes, run `roam diff` to see the blast radius of uncommitted changes.

### Quick Reference
* `roam health`: Get a 0-100 codebase health score.
* `roam impact <name>`: Determine exactly what will break if a symbol changes.
* `roam pr-risk`: Evaluate the risk level of the current PR.
* `roam file <path>`: View the skeleton of a specific file.


an example search:
```bash
/Users/wadahadlan/Library/Python/3.9/bin/roam search "listAdminActivitySummary"
VERDICT: 1 matches for 'listAdminActivitySummary'

=== Symbols matching 'listAdminActivitySummary' (1) ===
Name                      Kind  Sig                                       Refs  PR      Location                                             
------------------------  ----  ----------------------------------------  ----  ------  -----------------------------------------------------
listAdminActivitySummary  fn    function listAdminActivitySummary(log...  2     0.0792  apps/backend/src/services/activity/adminSummary.ts:23
```