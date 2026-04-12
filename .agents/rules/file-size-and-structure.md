---
trigger: always_on
---

## File Architecture & Length Constraints
To maintain high readability and modularity, this project enforces a **Strict 70-Line Limit** for all source files. Smaller files facilitate better testing, easier debugging, and more efficient AI context management.

### The 70-Line Rule
* **Target Length:** Files should ideally remain under **70 lines** of code. 
* **The Threshold:** Once a file exceeds ~80 lines, it is considered "technical debt" and must be refactored.
* **Organization:** Logic must be organized into a logical folder hierarchy rather than being centralized in "God files." Prefer flat, descriptive directory structures over deeply nested ones.

### Refactoring Protocol
If you encounter a file that violates this length constraint, do not ignore it. Follow these steps:
1.  **Decompose:** Identify logical sub-components, helper functions, or utility logic that can be extracted.
2.  **Relocate:** Move extracted logic into new files within the same directory or a relevant `utils/`, `components/`, or `services/` folder.
3.  **Abstract:** Use clean exports and imports to re-integrate the logic into the original file, ensuring the main entry point is now a "concise coordinator" rather than a "verbose worker."

### Proper Folder Structure
* **Feature-Based:** Group files by feature rather than just by type (e.g., `features/auth/` containing its own components, hooks, and types).
* **Discoverability:** Filenames must be descriptive. If a folder contains more than 10 files, evaluate if it needs sub-categorization.

### Enforcement
Before every commit, perform a "size check." If your changes push a file over the limit, refactor immediately as part of that task.

---

**Proposed Filename:** `.agents/rules/file-size-and-structure.md`