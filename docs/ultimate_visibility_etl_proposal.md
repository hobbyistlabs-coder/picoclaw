# Ultimate Visibility: ETL Framework Proposal

## Executive Summary
This document proposes a comprehensive ETL (Extract, Transform, Load) framework designed to provide "Ultimate Visibility" into our system's health and performance. By leveraging a robust stack of observability tools, we can proactively monitor our infrastructure, diagnose issues rapidly, and make data-driven decisions regarding feature deployments.

## The ETL Architecture

Our observability pipeline will be orchestrated using Docker Compose and consist of the following components:

1. **Extract (Vector):** We will utilize Vector as our lightweight, high-performance data collection agent. Vector will read raw log entries from our application's log files (e.g., `workspace/logs/etl/system_metrics.jsonl`).
2. **Transform (Vector):** Before routing data, Vector will transform these raw logs into structured metrics. We will use Vector's `log_to_metric` transform to extract key numerical values and tag them appropriately. All generated metrics will use the `etl_` prefix for consistency (e.g., `etl_http_requests_total`).
3. **Load (Prometheus & Loki):**
   - **Metrics:** Time-series data (gauges, counters, histograms) will be exported by Vector and scraped by Prometheus.
   - **Logs:** The original, structured log lines will be forwarded to Loki for efficient, label-based log aggregation and querying.
4. **Visualize (Grafana):** Grafana will sit on top of Prometheus and Loki, providing unified dashboards that visualize our key performance indicators (KPIs) and alert us to anomalies. Datasources will be configured with explicit UIDs (`uid: Prometheus`, `uid: Loki`) for seamless dashboard provisioning.

## Top KPIs: The Four Golden Signals

To achieve true visibility, our dashboards and alerts will focus on the "Four Golden Signals" of monitoring:

1. **Latency:** The time it takes to service a request (e.g., API response times).
2. **Traffic:** A measure of how much demand is being placed on the system (e.g., HTTP requests per second).
3. **Errors:** The rate of requests that fail (e.g., HTTP 5xx errors).
4. **Saturation:** How "full" the service is (e.g., CPU utilization, memory usage, or gauge metrics like `etl_goroutines`).

*Note on Alerts:* When creating Prometheus alerts for gauge metrics like `etl_goroutines`, we will use `deriv(metric[5m]) > 0` to detect positive trends (leaks) rather than `rate()`, which is intended for counters.

## High-Impact Focus: API Response Times

**Why API Response Time?**
API Response Time (Latency) is often the most direct indicator of user experience. Even if the system is up and not throwing errors, a slow system is perceived as a broken system.

**The "Go/No-Go" Signal for Feature Implementation**
Tracking `etl_http_request_duration_seconds_bucket` allows us to establish a baseline for acceptable performance (e.g., 95th percentile response time < 200ms).

When proposing or testing a new feature (Feature X):
1. **Baseline Measurement:** We verify the current p95 latency.
2. **Canary/Staging Deployment:** We deploy Feature X to a subset of traffic or a staging environment.
3. **Overhead Analysis:** We compare the new p95 latency against the baseline.
4. **The Signal:**
   - **GO:** If Feature X adds negligible overhead or stays within our latency budget, it is safe to proceed to a full rollout.
   - **NO-GO:** If Feature X causes latency to spike beyond our acceptable threshold, the deployment is halted. The feature must be optimized before it can be released, preventing widespread degradation of the user experience.

By making API Response Time our primary gatekeeper, we ensure that new functionality never comes at the cost of core system responsiveness.
