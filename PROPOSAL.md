# Ultimate Visibility: ETL Framework Proposal

## Architecture
To achieve "Ultimate Visibility" across our application, we propose an ETL (Extract, Transform, Load) framework leveraging industry-standard observability tools. The architecture will consist of:

1.  **Vector (Log Parsing & Metrics Extraction):** We will use Vector (specifically image `0.34.0-alpine`) to parse our structured JSONL logs (`workspace/logs/etl/system_metrics.jsonl`). Vector acts as the core pipeline, transforming log data into usable metrics (via `log_to_metric` using the `etl_` prefix) and forwarding the raw logs.
2.  **Prometheus (Time-Series Database & Alerting):** Prometheus will scrape the metrics exposed by Vector, storing them efficiently for time-series analysis and evaluating alerting rules (e.g., deriv thresholds for gauge metrics and histogram averages).
3.  **Loki (Log Storage):** Loki will receive and store the raw structured logs from Vector, allowing for deep-dive investigation when metrics indicate an anomaly.
4.  **Grafana (Visualization):** Grafana will sit on top of both Prometheus and Loki (configured via explicit UIDs `Prometheus` and `Loki`), providing dynamic dashboards to visualize the system's health.

## Key Performance Indicators (KPIs)

Our primary focus will be on the **Four Golden Signals**:

1.  **Latency:** The time it takes to service a request.
2.  **Traffic:** A measure of how much demand is being placed on the system.
3.  **Errors:** The rate of requests that fail.
4.  **Saturation:** How "full" the system is.

## High-Impact Area Focus: Resource Utilization

Resource utilization is a critical component of the **Saturation** signal. By closely monitoring metrics like **Goroutines** and **Go Runtime Garbage Collection (GC) Pauses**, we can establish a reliable baseline of system overhead.

### The 'Go/No-Go' Signal for Feature Implementation

When evaluating the implementation of a new feature, Resource Utilization metrics provide a definitive 'Go/No-Go' signal.

For instance, if we implement a new background processing queue:
*   **Go:** If the number of Goroutines remains stable under load and GC pause times stay consistently below our target threshold (e.g., < 1ms), the feature's performance overhead is acceptable.
*   **No-Go:** If we observe a steep, continuous derivative in the Goroutine count (`deriv(etl_goroutines[5m]) > 0`)—indicating a Goroutine leak—or if GC pause averages spike significantly, the feature introduces unacceptable saturation and must be blocked or rewritten before release.

This proactive observability ensures that new features do not silently degrade the overall performance and reliability of the application.
