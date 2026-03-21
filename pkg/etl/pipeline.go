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

const (
	MemoryLimitMB = 10.0 // Strict Go/No-Go memory limit
)

// Pipeline represents the ETL framework for Ultimate Visibility.
// It tracks system KPIs (memory, goroutines) and provides objective
// Go/No-Go signals based on strict resource utilization constraints.
type Pipeline struct {
	workspacePath string
	interval      time.Duration
	stopCh        chan struct{}
	stopOnce      sync.Once
}

// MetricEvent represents a single time-series metric entry.
type MetricEvent struct {
	Timestamp      time.Time `json:"timestamp"`
	Goroutines     int       `json:"goroutines"`
	MemoryAllocMB  float64   `json:"memory_alloc_mb"`
	MemoryTotalMB  float64   `json:"memory_total_mb"`
	MemorySysMB    float64   `json:"memory_sys_mb"`
	NumGC          uint32    `json:"num_gc"`
	GoNoGoSignal   string    `json:"go_no_go_signal"`
	SignalMessage  string    `json:"signal_message,omitempty"`
}

// NewPipeline initializes a new ETL pipeline.
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

// Start begins the ETL background worker.
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
				p.Run()
			}
		}
	}()
}

// Stop gracefully stops the ETL worker.
func (p *Pipeline) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopCh)
	})
}

// Run executes a single Extract, Transform, Load cycle.
func (p *Pipeline) Run() {
	rawMetrics := p.Extract()
	event := p.Transform(rawMetrics)
	p.Load(event)
}

// Extract retrieves the raw runtime metrics.
func (p *Pipeline) Extract() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

// Transform converts raw metrics into a normalized structured format
// and applies the Go/No-Go threshold evaluation logic.
func (p *Pipeline) Transform(m runtime.MemStats) MetricEvent {
	allocMB := float64(m.Alloc) / 1024 / 1024
	totalAllocMB := float64(m.TotalAlloc) / 1024 / 1024
	sysMB := float64(m.Sys) / 1024 / 1024
	goroutines := runtime.NumGoroutine()

	event := MetricEvent{
		Timestamp:     time.Now(),
		Goroutines:    goroutines,
		MemoryAllocMB: allocMB,
		MemoryTotalMB: totalAllocMB,
		MemorySysMB:   sysMB,
		NumGC:         m.NumGC,
		GoNoGoSignal:  "GO",
	}

	// Go/No-Go Logic: Evaluate against PicoClaw's rigorous budget
	if allocMB > MemoryLimitMB {
		event.GoNoGoSignal = "NO-GO"
		event.SignalMessage = fmt.Sprintf("Memory threshold breached: %.2fMB > %.2fMB target", allocMB, MemoryLimitMB)
		logger.ErrorCF("ETL Framework", "NO-GO SIGNAL TRIGGERED", map[string]any{
			"memory_alloc_mb": allocMB,
			"threshold_mb":    MemoryLimitMB,
		})
	} else {
		event.SignalMessage = "Resource utilization within budget."
	}

	return event
}

// Load writes the structured event to the Time-Series database (JSONL file).
func (p *Pipeline) Load(event MetricEvent) {
	etlDir := filepath.Join(p.workspacePath, "logs", "etl")
	if err := os.MkdirAll(etlDir, 0755); err != nil {
		logger.ErrorCF("ETL Framework", "Failed to create ETL directory", map[string]any{"error": err.Error()})
		return
	}

	filePath := filepath.Join(etlDir, "system_metrics.jsonl")
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.ErrorCF("ETL Framework", "Failed to open metrics file", map[string]any{"error": err.Error()})
		return
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		logger.ErrorCF("ETL Framework", "Failed to marshal metric event", map[string]any{"error": err.Error()})
		return
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		logger.ErrorCF("ETL Framework", "Failed to write metric event", map[string]any{"error": err.Error()})
	}
}
