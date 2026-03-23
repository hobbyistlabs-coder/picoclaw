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

// ETLPipeline represents the background task responsible for extracting
// system metrics (memory, goroutines) and writing them to a JSONL file.
// This supports the 'Ultimate Visibility' Go/No-Go signals framework.
type ETLPipeline struct {
	workspacePath string
	interval      time.Duration
	stopCh        chan struct{}
	stopOnce      sync.Once
}

// SystemMetrics represents a point-in-time snapshot of application health.
type SystemMetrics struct {
	Timestamp      string  `json:"timestamp"`
	Goroutines     int     `json:"goroutines"`
	MemoryAllocMB  float64 `json:"memory_alloc_mb"`
	MemoryTotalMB  float64 `json:"memory_total_mb"`
	MemorySysMB    float64 `json:"memory_sys_mb"`
}

// NewETLPipeline creates a new pipeline instance.
func NewETLPipeline(workspacePath string, interval time.Duration) *ETLPipeline {
	if interval <= 0 {
		interval = 1 * time.Minute
	}
	return &ETLPipeline{
		workspacePath: workspacePath,
		interval:      interval,
		stopCh:        make(chan struct{}),
	}
}

// Start initiates the background metrics collection loop.
func (p *ETLPipeline) Start(ctx context.Context) {
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
}

// Stop halts the metrics collection loop.
func (p *ETLPipeline) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopCh)
	})
}

// extractAndLoad gathers runtime metrics and appends them to the JSONL file.
func (p *ETLPipeline) extractAndLoad() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := SystemMetrics{
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Goroutines:     runtime.NumGoroutine(),
		MemoryAllocMB:  float64(m.Alloc) / 1024 / 1024,
		MemoryTotalMB:  float64(m.TotalAlloc) / 1024 / 1024,
		MemorySysMB:    float64(m.Sys) / 1024 / 1024,
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		logger.ErrorCF("ETLPipeline", "Failed to marshal metrics", map[string]any{"error": err.Error()})
		return
	}
	data = append(data, '\n')

	dir := filepath.Join(p.workspacePath, "logs", "etl")
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.ErrorCF("ETLPipeline", "Failed to create etl log directory", map[string]any{"error": err.Error(), "dir": dir})
		return
	}

	filePath := filepath.Join(dir, "system_metrics.jsonl")
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.ErrorCF("ETLPipeline", "Failed to open metrics file", map[string]any{"error": err.Error(), "file": filePath})
		return
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		logger.ErrorCF("ETLPipeline", "Failed to write metrics", map[string]any{"error": err.Error(), "file": filePath})
	}
}
