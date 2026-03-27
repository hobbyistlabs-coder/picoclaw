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

type SystemMetrics struct {
	Timestamp     time.Time `json:"timestamp"`
	MemoryAllocMB float64   `json:"memory_alloc_mb"`
	Goroutines    int       `json:"goroutines"`
}

type ETLPipeline struct {
	interval      time.Duration
	workspacePath string
	stopCh        chan struct{}
	stopOnce      sync.Once
	mu            sync.Mutex
}

func NewETLPipeline(workspacePath string, interval time.Duration) *ETLPipeline {
	if interval == 0 {
		interval = 10 * time.Second
	}
	return &ETLPipeline{
		interval:      interval,
		workspacePath: workspacePath,
		stopCh:        make(chan struct{}),
	}
}

func (p *ETLPipeline) Start(ctx context.Context) error {
	logDir := filepath.Join(p.workspacePath, "logs", "etl")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create etl log directory: %w", err)
	}

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
	return nil
}

func (p *ETLPipeline) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopCh)
	})
}

func (p *ETLPipeline) extractAndLoad() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := SystemMetrics{
		Timestamp:     time.Now().UTC(),
		MemoryAllocMB: float64(m.Alloc) / 1024 / 1024,
		Goroutines:    runtime.NumGoroutine(),
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		logger.ErrorCF("ETL", "Failed to marshal metrics", map[string]any{"error": err.Error()})
		return
	}
	data = append(data, '\n')

	p.mu.Lock()
	defer p.mu.Unlock()

	logFile := filepath.Join(p.workspacePath, "logs", "etl", "system_metrics.jsonl")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.ErrorCF("ETL", "Failed to open metrics file", map[string]any{"error": err.Error()})
		return
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		logger.ErrorCF("ETL", "Failed to write metrics", map[string]any{"error": err.Error()})
	}
}
