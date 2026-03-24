package etl

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"jane/pkg/logger"
)

// SystemMetrics represents a snapshot of the system's resource utilization.
type SystemMetrics struct {
	Timestamp      string  `json:"timestamp"`
	Goroutines     int     `json:"goroutines"`
	MemoryAllocMB  float64 `json:"memory_alloc_mb"`
	MemoryTotalMB  float64 `json:"memory_total_mb"`
	MemorySysMB    float64 `json:"memory_sys_mb"`
	NumGC          uint32  `json:"num_gc"`
	GCPauseTotalNs uint64  `json:"gc_pause_total_ns"`
}

// Pipeline orchestrates the periodic extraction of system metrics
// and writes them as JSONL to the designated log file.
type Pipeline struct {
	workspacePath string
	interval      time.Duration
	stopCh        chan struct{}
	stopOnce      sync.Once
}

// NewPipeline initializes a new ETL pipeline for system metrics.
func NewPipeline(workspacePath string, interval time.Duration) *Pipeline {
	if interval <= 0 {
		interval = 1 * time.Minute
	}
	return &Pipeline{
		workspacePath: workspacePath,
		interval:      interval,
		stopCh:        make(chan struct{}),
	}
}

// Start begins the periodic extraction and loading of metrics.
func (p *Pipeline) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-p.stopCh:
				return
			case <-ticker.C:
				p.extractAndLoad()
			}
		}
	}()
	logger.InfoCF("ETL", "Pipeline started", map[string]any{
		"interval": p.interval.String(),
	})
}

// Stop gracefully halts the pipeline.
func (p *Pipeline) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopCh)
		logger.InfoCF("ETL", "Pipeline stopped", nil)
	})
}

// extractAndLoad gathers current metrics and appends them to the JSONL file.
func (p *Pipeline) extractAndLoad() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := SystemMetrics{
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Goroutines:     runtime.NumGoroutine(),
		MemoryAllocMB:  float64(m.Alloc) / 1024 / 1024,
		MemoryTotalMB:  float64(m.TotalAlloc) / 1024 / 1024,
		MemorySysMB:    float64(m.Sys) / 1024 / 1024,
		NumGC:          m.NumGC,
		GCPauseTotalNs: m.PauseTotalNs,
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		logger.ErrorCF("ETL", "Failed to marshal metrics", map[string]any{"error": err.Error()})
		return
	}

	// Ensure target directory exists
	targetDir := filepath.Join(p.workspacePath, "logs", "etl")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		logger.ErrorCF("ETL", "Failed to create metrics directory", map[string]any{"error": err.Error(), "path": targetDir})
		return
	}

	targetFile := filepath.Join(targetDir, "system_metrics.jsonl")

	f, err := os.OpenFile(targetFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.ErrorCF("ETL", "Failed to open metrics file", map[string]any{"error": err.Error(), "file": targetFile})
		return
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		logger.ErrorCF("ETL", "Failed to write metrics", map[string]any{"error": err.Error(), "file": targetFile})
	}
}
