package etl

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"jane/pkg/logger"
)

// SystemMetrics represents a single snapshot of system health for ETL
type SystemMetrics struct {
	Timestamp      string  `json:"timestamp"`
	Goroutines     int     `json:"goroutines"`
	MemoryAllocMB  float64 `json:"memory_alloc_mb"`
	MemoryTotalMB  float64 `json:"memory_total_mb"`
	MemorySysMB    float64 `json:"memory_sys_mb"`
	NumGC          uint32  `json:"num_gc"`
	GCPauseTotalNs uint64  `json:"gc_pause_total_ns"`
}

// Pipeline orchestrates the periodic extraction, transformation, and loading of metrics.
type Pipeline struct {
	workspacePath string
	interval      time.Duration
	stopCh        chan struct{}
	stopOnce      sync.Once
}

// NewPipeline creates a new ETL Pipeline that runs at the specified interval.
func NewPipeline(workspacePath string, interval time.Duration) *Pipeline {
	if interval == 0 {
		interval = 60 * time.Second
	}
	return &Pipeline{
		workspacePath: workspacePath,
		interval:      interval,
		stopCh:        make(chan struct{}),
	}
}

// Start begins the ETL pipeline background goroutine.
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
				p.runExtraction()
			}
		}
	}()
}

// Stop gracefully stops the pipeline.
func (p *Pipeline) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopCh)
	})
}

func (p *Pipeline) runExtraction() {
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

	logDir := filepath.Join(p.workspacePath, "logs", "etl")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logger.ErrorCF("ETL", "Failed to create log directory", map[string]any{"error": err.Error()})
		return
	}

	logFile := filepath.Join(logDir, "system_metrics.jsonl")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.ErrorCF("ETL", "Failed to open metrics file", map[string]any{"error": err.Error()})
		return
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf("%s\n", data)); err != nil {
		logger.ErrorCF("ETL", "Failed to write metrics", map[string]any{"error": err.Error()})
	}
}
