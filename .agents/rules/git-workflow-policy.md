---
trigger: always_on
---

## Atomic Version Control
This project follows a strict **Feature-Branch** workflow. Every change must be isolated and documented through a granular, explanatory commit history.

### Feature Isolation
* **Branch per Task:** Create a dedicated branch for every new feature, bug fix, or refactor. Never work directly on `main`.
    `git checkout -b <branch-type>/<brief-description>`
* **Meaningful Increments:** Commit after every successful, meaningful addition or logic change. Aim for "atomic" commits—the smallest unit of work that still functions.

### The Explanatory Commit Standard
Every commit message must provide context. Do not use vague messages like "fixed bug" or "update." Follow this structure:
1. **Header:** A brief summary (max 50 chars) in the imperative mood (e.g., "Add user auth" not "Added user auth").
2. **Body:** A concise paragraph explaining **why** the change was made and **what** it accomplishes. 
    * *Example:* `git commit -m "feat: implement JWT token refresh" -m "This adds a middleware check to refresh tokens 5 minutes before expiration to prevent session timeouts during active use."`

### Integration & Cleanup
1. **Sync:** Once the task is complete, pull the latest changes to your local `main`.
2. **Merge:** Merge your feature branch into `main`. For complex features, use `--no-ff` (no-fast-forward) to preserve the feature's history.
3. **Prune:** Delete the local feature branch immediately after a successful merge.
    `git branch -d <branch-name>`

Run `git log --graph --oneline` regularly to ensure the project history remains clean and readable.