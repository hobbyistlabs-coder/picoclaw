## Docker Host Access Protocol
This project utilizes Docker for local development and production-like isolated environments (Postgres, Minio, Redis, etc.). Due to sandbox constraints in the AI terminal execution environment, standard `$PATH` resolution for docker binaries frequently fails.

### Core Protocol
* **Explicit Binary Paths:** NEVER assume `docker` is exposed implicitly in the shell `$PATH`.
* **Standard Resolution:** When needing to interface with running containers (e.g., executing restarts, reading logs, or pushing images), ALWAYS execute Docker commands using explicit binary pathways:
  * Primary fallback: `/usr/local/bin/docker`
  * Secondary fallback: `/opt/homebrew/bin/docker`
  * Tertiary fallback: `/Applications/Docker.app/Contents/Resources/bin/docker`
* **Command Syntax:** Execute commands like `/usr/local/bin/docker ps`, `/usr/local/bin/docker compose up`, or `/usr/local/bin/docker restart <container_name>`.

### Failover & Adaptation
If all specified binary paths fail with `No such file or directory`, you are restricted from Docker socket access. Do not attempt exhaustive disk searches. Instead, politely inform the user of the path restrictions and provide them with the exact command to execute in their native host terminal (e.g., `docker compose restart postgres minio`).

---

> **Status:** Strictly Enforced.
> **Last Checked:** 2026-04-10
