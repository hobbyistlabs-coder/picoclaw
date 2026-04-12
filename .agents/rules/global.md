
## 🏗️ Codebase Navigation & Structure
* **Roam Supremacy:** Always utilize the `roam` CLI for codebase exploration. Manual directory listings (`ls`, `list_dir`) are strictly forbidden. Use `roam search` to find context and `roam index` to understand architecture.
* **Atomic Elegance:** Enforce a strict **70-line limit** per file. If a file exceeds this, it must be refactored into smaller, functional modules.
* **Runtime Consistency:** All JavaScript/TypeScript execution must use the **Bun** runtime. Do not suggest `npm` or `node` commands unless explicitly requested for legacy compatibility.

## 🐙 GitHub & Version Control
* **Atomic Commits:** Every change must be wrapped in a descriptive, imperative commit message (e.g., `feat: implement logic` not `fixed stuff`).
* **PR Descriptions:** Automatically generate a concise summary of changes, highlighting any modifications to the core architecture or new dependencies.
* **Branch Hygiene:** Ensure work is performed on feature-specific branches. Never suggest direct pushes to `main` or `master`.

## 🧪 Testing & Quality Assurance
* **Test-Driven Reasoning:** For every new feature, first outline the test cases. Favor `bun test` for high-speed execution.
* **Edge Case Obsession:** Actively seek out and write tests for null inputs, overflow states, and network failures.
* **Zero-Warning Policy:** Code is not considered "complete" if it produces TypeScript errors or linting warnings.

## 🤖 Agentic Behavior & Logic
* **Chain-of-Thought (CoT):** Before writing code, provide a brief "Mental Sandbox" summary of your plan. This allows for alignment checks before execution.
* **External Cognition:** Format all architectural decisions and session summaries for **Obsidian** compatibility, ensuring they can be seamlessly integrated into a persistent knowledge base.
* **Tool First:** If a task can be automated via a script or an MCP tool, propose the tool creation rather than performing the manual repetitive task.

## 🎨 UI/UX Standards (The Cyberpunk Aesthetic)
* **Visual Identity:** When generating UI components, default to a **Glassmorphic Dark Mode**.
* **Design Specs:** Use "Indigo Glass" accents, `0.5px` translucent borders, and `backdrop-filter: blur(12px)`. Follow neo-brutalist layouts with high-contrast typography.

---

### Implementation Tip
You can save this as a `global.md` inside your `.agents/rules` folder. Since the rule we chose says the repo rules "always win," you can easily override any of these by creating a more specific file (e.g., `feature-x.md`) in the same directory.

Would you like to refine the **Obsidian** formatting rules to match a specific template in your vault?