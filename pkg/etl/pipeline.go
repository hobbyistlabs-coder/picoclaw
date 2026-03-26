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

// Pipeline represents the ETL framework pipeline for system metrics.
type Pipeline struct {
	workspacePath string
	interval      time.Duration
	stopCh        chan struct{}
	stopOnce      sync.Once
	mu            sync.Mutex
}

// SystemMetrics represents the data collected per interval.
type SystemMetrics struct {
	Timestamp    string  `json:"timestamp"`
	Goroutines   int     `json:"sys.goroutines"`
	AllocMB      float64 `json:"sys.memory.alloc_mb"`
	TotalAllocMB float64 `json:"sys.memory.total_alloc_mb"`
	SysMB        float64 `json:"sys.memory.sys_mb"`
	NumGC        uint32  `json:"sys.memory.num_gc"`
	PauseTotalNs uint64  `json:"sys.memory.gc_pause_total_ns"`
}

// NewPipeline initializes a new ETL pipeline.
func NewPipeline(workspacePath string, interval time.Duration) *Pipeline {
	if interval == 0 {
		interval = time.Minute
	}
	return &Pipeline{
		workspacePath: workspacePath,
		interval:      interval,
		stopCh:        make(chan struct{}),
	}
}

// Start begins the ETL pipeline background process.
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
}

// Stop halts the ETL pipeline gracefully.
func (p *Pipeline) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopCh)
	})
}

// extractAndLoad gathers metrics and writes them to the JSONL file.
func (p *Pipeline) extractAndLoad() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := SystemMetrics{
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Goroutines:   runtime.NumGoroutine(),
		AllocMB:      float64(m.Alloc) / 1024 / 1024,
		TotalAllocMB: float64(m.TotalAlloc) / 1024 / 1024,
		SysMB:        float64(m.Sys) / 1024 / 1024,
		NumGC:        m.NumGC,
		PauseTotalNs: m.PauseTotalNs,
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		logger.ErrorCF("ETL", "Failed to marshal metrics", map[string]any{"error": err.Error()})
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	logsDir := filepath.Join(p.workspacePath, "logs", "etl")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		logger.ErrorCF("ETL", "Failed to create logs directory", map[string]any{"error": err.Error(), "path": logsDir})
		return
	}

	logFile := filepath.Join(logsDir, "system_metrics.jsonl")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.ErrorCF("ETL", "Failed to open metrics file", map[string]any{"error": err.Error(), "path": logFile})
		return
	}
	defer f.Close()

	data = append(data, '\n')
	if _, err := f.Write(data); err != nil {
		logger.ErrorCF("ETL", "Failed to write metrics", map[string]any{"error": err.Error()})
	}
}
