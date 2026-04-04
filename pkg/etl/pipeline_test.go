package etl

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipeline_ExtractAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "logs", "etl", "system_metrics.jsonl")

	p := NewPipeline(tmpDir, 100*time.Millisecond)

	// ensure log dir exists like Start does
	err := os.MkdirAll(filepath.Dir(logFile), 0755)
	require.NoError(t, err)

	p.extractAndLoad()

	// Verify file exists
	_, err = os.Stat(logFile)
	require.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	var metrics SystemMetrics
	err = json.Unmarshal(content, &metrics)
	require.NoError(t, err)

	assert.NotZero(t, metrics.Timestamp)
	assert.Greater(t, metrics.Goroutines, 0)
}

func TestPipeline_StartStop(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "logs", "etl", "system_metrics.jsonl")

	p := NewPipeline(tmpDir, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p.Start(ctx)

	// Wait enough time for a few ticks
	time.Sleep(50 * time.Millisecond)

	p.Stop()

	// Verify file exists
	_, err := os.Stat(logFile)
	require.NoError(t, err)
}
