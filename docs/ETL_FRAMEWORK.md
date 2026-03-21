# ETL Framework for Ultimate Visibility

## Overview

As PicoClaw targets an ultra-lightweight environment (<10MB RAM, $10 hardware), implementing a robust ETL (Extract, Transform, Load) framework for 'Ultimate Visibility' is critical. This framework monitors system health and establishes objective 'Go/No-Go' criteria for new feature implementations based on strict performance budgets.

## Architecture

1.  **Extract**: Data is extracted using the `ResourceTracker` and `runtime.ReadMemStats()`.
2.  **Transform**: Data is normalized into time-series structs `MetricEvent`, and evaluated against predefined Go/No-Go thresholds.
3.  **Load**: JSONL formatted metrics are written out to `{workspacePath}/logs/etl/system_metrics.jsonl`.

## KPIs

To maintain the rigorous constraints of the PicoClaw architecture, the following KPIs are Paramount:

1.  **Memory Allocation (`sys.memory.alloc_mb`)**: The absolute memory footprint of the application. **Target: strictly <10MB.**
2.  **Active Goroutine Count (`sys.goroutines`)**: Indicates the level of concurrency. Steadily increasing counts signify potential goroutine leaks.

## Resource Utilization as a Go/No-Go Signal

Resource utilization serves as our definitive indicator for deploying or rolling back features.

*   **No-Go Signal**: If a feature causes memory to consistently exceed **10MB** or goroutines to leak, the implementation receives a strict 'NO-GO' and will trigger an error log, indicating it must be optimized.
*   **Go Signal**: If the system remains predictably stable and `<10MB`, it is ready for production.
