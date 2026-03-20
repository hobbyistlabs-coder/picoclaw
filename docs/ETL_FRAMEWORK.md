# ETL Framework Proposal: Ultimate Visibility

## 1. Executive Summary

To achieve "Ultimate Visibility" across the PicoClaw application, we need a robust ETL (Extract, Transform, Load) framework. This framework will passively collect system telemetry and application health metrics, transform them into a standardized schema, and load them into a centralized, queryable format (e.g., JSON Lines for easy ingestion into Data Warehouses or Log Aggregation tools like Elasticsearch/Datadog/BigQuery).

## 2. Top KPIs for System Health

To maintain system reliability and performance, the following Key Performance Indicators (KPIs) are critical:

1.  **Memory Utilization (Allocated vs. Sys):** Detects memory leaks and inefficient garbage collection.
2.  **Active Goroutines:** Identifies concurrency leaks or unhandled blocking operations.
3.  **API Response Times (P50, P90, P99):** Measures system responsiveness and user experience.
4.  **Error Rates (Model vs. Infrastructure vs. Logic):** Classifies failures to direct debugging efforts.
5.  **CPU Usage/GC Pauses:** Indicates computational bottlenecks or excessive allocation overhead.

## 3. High-Impact Area Focus: Resource Utilization (Memory & Goroutines)

We will focus our initial ETL implementation on **Resource Utilization**. In a long-running Go application like PicoClaw, resource leaks (specifically goroutines and memory) are the most common cause of silent degradation and eventual Out-Of-Memory (OOM) crashes.

### The "Go/No-Go" Signal for Feature X

**Hypothetical Feature:** Implementing a *High-Concurrency WebSocket Feature* for real-time agent streaming.

**How Tracking Resource Utilization Provides the Signal:**

1.  **Baseline Extraction:** The ETL pipeline establishes a baseline of memory and goroutine usage during normal operation over a 24-hour period.
2.  **Canary Deployment:** Feature X is deployed to a subset of users or subjected to a load test.
3.  **Transformation & Analysis:** The ETL pipeline continuously extracts metrics. We monitor the rate of change (slope) of `sys.goroutines` and `sys.memory.alloc_mb`.
4.  **Go/No-Go Decision:**
    *   **Go:** If the goroutine count stabilizes after connections are established, and memory allocation plateaus within acceptable limits, the feature is safe for general release.
    *   **No-Go:** If the ETL data shows a linear upward trend in goroutines (a leak) or unbounded memory growth when WebSocket clients disconnect, it provides an immediate, empirical **No-Go** signal. The feature must be rolled back and the connection teardown logic debugged before a full rollout.

## 4. Proposed ETL Architecture

*   **Extract:** A new `etl.Pipeline` service will hook into the Go runtime (similar to the existing `ResourceTracker`) and periodically sample metrics (e.g., every 60 seconds).
*   **Transform:** The raw runtime stats will be transformed into a structured JSON schema, tagging each event with a timestamp, node ID, and environment.
*   **Load:** The structured JSON will be written to an append-only JSONL file (`workspace/logs/etl/system_metrics.jsonl`). External log forwarders (like FluentBit or Promtail) can then tail this file and load it into the final Data Warehouse.

This local-first approach ensures the agent remains lightweight while providing the necessary hooks for enterprise-grade observability.
