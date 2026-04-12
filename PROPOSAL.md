# Ultimate Visibility ETL Framework Proposal

## Executive Summary
This proposal outlines an ETL (Extract, Transform, Load) framework designed to provide "Ultimate Visibility" into the health, performance, and operational status of our application. We will achieve this through a robust, scalable metrics and logging pipeline orchestrated by Docker Compose.

## Architecture & Framework
We propose an infrastructure built on industry-standard observability tools:
- **Vector:** Acts as the high-performance log parser and routing backbone. It will ingest raw JSONL metrics output by our application (from `workspace/logs/etl/system_metrics.jsonl`), apply structural transformations, and route data to appropriate sinks.
- **Prometheus:** The primary time-series database. It will scrape metrics exposed by Vector, allowing for powerful querying (PromQL) and alerting based on system performance.
- **Loki:** A horizontally-scalable, highly-available, multi-tenant log aggregation system. It will store raw logs for deep-dive troubleshooting and historical analysis.
- **Grafana:** The visualization layer. It will connect to both Prometheus and Loki to provide unified dashboards, bringing together the "Four Golden Signals" and custom application metrics into a single pane of glass.

## Top KPIs for System Health
To effectively monitor system health, we focus on the following Key Performance Indicators (extracted natively by our pipeline):
1.  **Goroutine Count (`etl_goroutines`):** A critical indicator of concurrency health. Rapid spikes or sustained growth without matching drops can indicate goroutine leaks, leading to memory exhaustion and system crashes.
2.  **Memory Allocations (`etl_memory_alloc_mb`, `etl_memory_total_mb`, `etl_memory_sys_mb`):** Tracking active memory usage, cumulative allocations, and OS-level memory requests helps identify memory leaks and inefficient object lifecycle management.
3.  **Garbage Collection Frequency (`etl_num_gc`):** High GC frequency can introduce significant application pauses, severely impacting latency.

## High-Impact Focus: Resource Utilization as a Go/No-Go Signal
**Focus Area:** Resource Utilization (specifically Goroutines and Memory footprint).

**Why this matters for Feature X:**
When evaluating the implementation of a new feature (e.g., "Feature X"—a new high-throughput streaming endpoint or intensive background processing job), understanding the current baseline Resource Utilization is paramount.

By closely tracking `etl_goroutines` and memory metrics *before* and *during* a canary rollout of Feature X, we gain a definitive **Go/No-Go** signal:

*   **No-Go Signal:** If the deployment of Feature X causes `deriv(etl_goroutines[5m]) > 0` to trigger consistently (indicating unchecked concurrent growth), or if memory allocations continuously trend upward without plateauing, the feature introduces an unsustainable overhead. The rollout must be halted and the feature optimized before proceeding, preventing wider system degradation.
*   **Go Signal:** If Feature X operates within an acceptable delta of our established baseline metrics—meaning goroutines spin up and spin down efficiently and memory remains stable after initial allocation—we have confidence that the feature can safely scale to our broader user base.

By establishing this data-driven framework, we move from reactive troubleshooting to proactive capacity planning and safer feature deployments.
