## Obsidian External Cognition Protocol
This project treats the Obsidian Vault as the **External Cognition Layer**. Documentation is not a byproduct of work; it is the system’s long-term memory and a mandatory requirement for all operations.

### Core Principle
**Documentation is not optional.** Every significant discovery, architectural decision, or complex bug resolution must be mirrored in the vault to ensure the system’s "long-term memory" remains intact.

### Operational Requirements
* **Active State:** Ensure the Obsidian desktop app is running before executing any `obsidian` CLI commands.
* **Targeting:** Use the `vault="<name>"` flag when performing actions on a background vault or one that is not the current default.
* **Output:** Append `--copy` to any command to immediately send the resulting output to the system clipboard for quick pasting.

### Standard Workflows
| Action               | Command Pattern                                                        |
| :------------------- | :--------------------------------------------------------------------- |
| **Interactive Mode** | `obsidian` (Opens TUI with autocomplete and history)                   |
| **Find Content**     | `obsidian search query="<pattern>"`                                    |
| **Read Note**        | `obsidian read file="<name>"`                                          |
| **Create Note**      | `obsidian create name="<name>" template="<template_name>"`             |
| **Quick Capture**    | `obsidian daily:append content="<text>"`                               |
| **Task Management**  | `obsidian tasks todo` OR `obsidian task file="<name>" line=<n> toggle` |

### System Maintenance & Utility
* **Note Integrity:** Use `obsidian diff file="<name>"` to track changes within the vault.
* **Organization:** Use `obsidian tags` to audit the current taxonomy.
* **Developer Tools:** `obsidian devtools` and `obsidian plugin:reload id="<id>"` for troubleshooting the Obsidian environment itself.

---