# Ultimate Visibility ETL Framework Proposal

## Objective
To implement an 'Ultimate Visibility' ETL framework across our application to monitor system health, optimize performance, and make data-driven architectural decisions.

## Framework Overview
The proposed ETL framework leverages the following stack to collect, transform, and visualize system metrics and logs:
- **Vector (`0.34.0-alpine`)**: Acts as the log shipper and metric extractor. It parses raw JSONL logs from `workspace/logs/etl/system_metrics.jsonl`, applies `log_to_metric` transforms (with the `etl_` prefix and explicitly extracted tags), and routes the data to Prometheus (metrics) and Loki (logs).
- **Prometheus**: Time-series database for storing extracted metrics and evaluating alerting rules (e.g., using `deriv()` for gauges and `sum(rate())` for histograms).
- **Loki**: Log aggregation system to store the raw parsed logs for deep diving into specific events.
- **Grafana**: Visualization platform connecting to both Prometheus (`uid: Prometheus`) and Loki (`uid: Loki`) to build comprehensive dashboards.

## Top KPIs (The Four Golden Signals)
Our visibility strategy will be anchored around the Four Golden Signals:
1. **Latency (Response Times):** e.g., `etl_http_request_duration_seconds` (histogram) - Time taken to service a request.
2. **Traffic:** e.g., `etl_http_requests_total` (counter) - Total demand placed on the system.
3. **Errors:** e.g., `etl_http_requests_total{status=~"5.."}` (counter) - Rate of failed requests.
4. **Saturation (Resource Utilization):** e.g., `etl_goroutines` (gauge), `etl_gc_pause_ms` (histogram) - How "full" the system is.

## High-Impact Area Focus: Resource Utilization
For the immediate focus, we will zero in on **Resource Utilization**, specifically tracking memory usage, goroutine counts, and Garbage Collection (GC) pauses.

### Why Resource Utilization?
Tracking these specific metrics provides a direct window into the operational overhead of the application. In Go applications, unchecked goroutine growth or excessive GC pauses can lead to cascading failures, OOM kills, and severely degraded latency for all requests.

### The 'Go/No-Go' Signal for Feature X
When evaluating the implementation of [Feature X] (e.g., a new intensive background processing job or a high-throughput websocket feature), Resource Utilization acts as our strict 'Go/No-Go' signal:

1. **Establish Baseline:** We use the ETL framework to establish a baseline of normal operation (e.g., peak goroutines < 10,000, max GC pause < 5ms).
2. **Deploy Feature X in Staging/Canary:** We introduce the feature under load.
3. **Evaluate KPIs:** We monitor the specific resource metrics:
   - **Go:** If `deriv(etl_goroutines[5m])` remains stable (meaning no goroutine leaks) and the average GC pause `sum(rate(etl_gc_pause_ms_sum[1m])) / sum(rate(etl_gc_pause_ms_count[1m]))` stays below the established threshold, the system is safely absorbing the overhead. The feature is a **Go** for full production rollout.
   - **No-Go:** If `deriv(etl_goroutines[5m])` shows a consistent positive trend (a leak) or GC pauses spike significantly beyond the threshold, the system is saturated. The feature is a **No-Go**; it must be optimized or throttled before broader release.

By rigorously monitoring Resource Utilization, we prevent new features from inadvertently destabilizing the core platform.
