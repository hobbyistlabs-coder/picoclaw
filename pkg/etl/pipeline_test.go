package etl

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestETLPipeline_StartStop(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test_metrics.jsonl")

	pipeline := NewPipeline(100*time.Millisecond, logPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := pipeline.Start(ctx)
	assert.NoError(t, err)

	// Wait for at least one tick
	time.Sleep(250 * time.Millisecond)

	pipeline.Stop()

	// Verify file was created
	stat, err := os.Stat(logPath)
	assert.NoError(t, err)
	assert.True(t, stat.Size() > 0)

	// Read content and parse JSON
	content, err := os.ReadFile(logPath)
	assert.NoError(t, err)

	lines := 0
	for _, c := range content {
		if c == '\n' {
			lines++
		}
	}
	assert.GreaterOrEqual(t, lines, 2) // Should have at least the initial extract + 1 tick

	var payload MetricPayload
	err = json.Unmarshal(content[:len(content)-1], &payload) // Trying to parse the whole might fail if multiple lines
	if err != nil {
		// Just parse first line to test
		firstLineEnd := 0
		for i, c := range content {
			if c == '\n' {
				firstLineEnd = i
				break
			}
		}
		err = json.Unmarshal(content[:firstLineEnd], &payload)
		assert.NoError(t, err)
	}

	assert.NotZero(t, payload.Goroutines)
	assert.NotZero(t, payload.MemoryAllocMB)
}
