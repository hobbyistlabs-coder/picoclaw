# Ultimate Visibility: ETL Framework Proposal

## Overview
This proposal outlines an ETL framework designed to provide "Ultimate Visibility" into the health and performance of our application. We achieve this by extracting crucial metrics and logs from the application (via `pkg/etl/pipeline.go`), transforming them using Vector, and loading them into a centralized monitoring stack composed of Prometheus, Loki, and Grafana.

## Key Performance Indicators (KPIs)
To ensure comprehensive system health monitoring, our framework focuses on the **Four Golden Signals**:

1. **Latency:** The time it takes to service a request (e.g., API Response Times). Tracked via specific span durations and processing delays.
2. **Traffic:** A measure of how much demand is being placed on the system (e.g., HTTP requests per second). Tracked via API hit counts.
3. **Errors:** The rate of requests that fail (e.g., 5xx HTTP responses, application exceptions). Tracked via standard error logging and metric counts.
4. **Saturation:** How "full" the system is, measuring constrained resources like Memory, CPU, and Goroutines. Tracked via extracted runtime metrics (e.g., Memory Allocated, GC Pause Times, Goroutine Count).

## High-Impact Focus: Resource Utilization
For this iteration, we focus specifically on **Resource Saturation**, tracking underlying system resources:
* **Goroutine Count (`etl_goroutines`)**
* **Garbage Collection Pauses (`etl_gc_pause_ms`)**
* **Memory Allocation (`etl_memory_alloc_mb`)**

### Go/No-Go Signal for Feature X
By tracking Resource Utilization specifically—like memory allocation overhead and GC pause times—we gain a clear "Go/No-Go" signal for implementing large, resource-intensive features (Feature X).

If deploying Feature X to a staging environment causes the `etl_gc_pause_ms` to consistently average over our target threshold (e.g., 10ms), or if we see a continuous, unbounded positive derivative in `etl_goroutines` (indicating a leak), these KPIs immediately trigger a **"No-Go"** alert. We would block the deployment of Feature X until the performance overhead or leaks are addressed, ensuring overall application stability is preserved.

## Architecture
1. **Application (pkg/etl):** Periodically extracts runtime metrics (memory, GC, goroutines) and writes them to a JSONL log file.
2. **Vector:** Tails the JSONL log file, parses the JSON, and converts the data points into Prometheus metrics. Logs are forwarded to Loki.
3. **Prometheus & Alertmanager:** Scrapes Vector for metrics and evaluates alerting rules (e.g., high GC pauses, goroutine leaks).
4. **Loki:** Stores raw JSON logs for deep-dive debugging.
5. **Grafana:** Visualizes the metrics from Prometheus and Loki via pre-configured datasources.
