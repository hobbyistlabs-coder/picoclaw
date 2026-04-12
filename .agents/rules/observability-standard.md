## Observability & Structured Logging
All systems must emit comprehensive logs to ensure full behavioral reconstruction. Every log entry must answer three questions: **What happened? When did it happen? Why did it fail?**

### The Structured Logging Standard
All logs must be emitted in **JSON** format to facilitate machine readability and automated analysis.

```json
{
  "timestamp": "2026-04-10T20:42:00Z",
  "level": "INFO",
  "component": "image-pipeline",
  "event": "upload.started",
  "requestId": "req_123",
  "message": "Starting upload"
}
```

### Log Levels & Usage
| Level     | Usage Case                                                            |
| :-------- | :-------------------------------------------------------------------- |
| **TRACE** | Fine-grained execution flow (internal logic steps).                   |
| **DEBUG** | Diagnostics and high-volume data for troubleshooting.                 |
| **INFO**  | Normal, expected state changes and milestones.                        |
| **WARN**  | Recoverable anomalies or unexpected but non-breaking behavior.        |
| **ERROR** | Operation failures that require attention but don't kill the process. |
| **FATAL** | Unrecoverable failures causing immediate system shutdown.             |

### Mandatory Log Events
Systems are required to log the following lifecycle events:
* **Initialization:** Process start/stop and configuration loading.
* **Boundaries:** All external API calls and database writes.
* **Resiliency:** Retries, circuit-breaker triggers, and failures.
* **Workflows:** Background job lifecycles and critical state transitions.

### Security Guardrail
**Strict Prohibition:** Never log secrets, PII (Personally Identifiable Information), API keys, or authentication tokens. All loggers should implement a redaction layer if possible.

---