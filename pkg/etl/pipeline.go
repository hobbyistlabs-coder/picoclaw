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

// Pipeline represents an ETL pipeline that periodically extracts system metrics,
// transforms them into a structured format, and loads them into a JSONL file.
type Pipeline struct {
	interval    time.Duration
	logFilePath string
	stopCh      chan struct{}
	stopOnce    sync.Once
}

// MetricPayload represents the transformed structure for the ETL load phase.
type MetricPayload struct {
	Timestamp      string  `json:"timestamp"`
	Goroutines     int     `json:"goroutines"`
	MemoryAllocMB  float64 `json:"memory_alloc_mb"`
	MemoryTotalMB  float64 `json:"memory_total_alloc_mb"`
	MemorySysMB    float64 `json:"memory_sys_mb"`
	NumGC          uint32  `json:"num_gc"`
	GCPauseTotalNs uint64  `json:"gc_pause_total_ns"`
}

// NewPipeline creates a new ETL pipeline writing to the specified path.
func NewPipeline(interval time.Duration, logFilePath string) *Pipeline {
	if interval == 0 {
		interval = 60 * time.Second
	}
	return &Pipeline{
		interval:    interval,
		logFilePath: logFilePath,
		stopCh:      make(chan struct{}),
	}
}

// Start begins the ETL pipeline in a background goroutine.
func (p *Pipeline) Start(ctx context.Context) error {
	// Ensure directory exists
	dir := filepath.Dir(p.logFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create etl log directory: %w", err)
	}

	ticker := time.NewTicker(p.interval)
	go func() {
		defer ticker.Stop()
		// Perform an initial extraction immediately
		p.extractTransformLoad()

		for {
			select {
			case <-ctx.Done():
				return
			case <-p.stopCh:
				return
			case <-ticker.C:
				p.extractTransformLoad()
			}
		}
	}()

	return nil
}

// Stop gracefully stops the ETL pipeline.
func (p *Pipeline) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopCh)
	})
}

// extractTransformLoad captures the current runtime stats and writes them to the JSONL file.
func (p *Pipeline) extractTransformLoad() {
	// Extract
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	goroutines := runtime.NumGoroutine()

	// Transform
	payload := MetricPayload{
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Goroutines:     goroutines,
		MemoryAllocMB:  float64(m.Alloc) / 1024 / 1024,
		MemoryTotalMB:  float64(m.TotalAlloc) / 1024 / 1024,
		MemorySysMB:    float64(m.Sys) / 1024 / 1024,
		NumGC:          m.NumGC,
		GCPauseTotalNs: m.PauseTotalNs,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		logger.ErrorCF("ETL", "Failed to marshal ETL metric payload", map[string]any{"error": err.Error()})
		return
	}
	data = append(data, '\n') // JSONL format

	// Load
	f, err := os.OpenFile(p.logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger.ErrorCF("ETL", "Failed to open ETL log file", map[string]any{"error": err.Error()})
		return
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		logger.ErrorCF("ETL", "Failed to write ETL metric payload", map[string]any{"error": err.Error()})
	}
}
