package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"jane/pkg/providers"
)

// jsonSession mirrors pkg/session.Session for migration purposes.
type jsonSession struct {
	Key      string              `json:"key"`
	Messages []providers.Message `json:"messages"`
	Summary  string              `json:"summary,omitempty"`
	Created  time.Time           `json:"created"`
	Updated  time.Time           `json:"updated"`
}

func MigrateFromJSONL(
	ctx context.Context, sessionsDir string, store Store,
) (int, error) {
	importer, ok := store.(*SQLiteStore)
	if !ok {
		return 0, nil
	}

	entries, err := os.ReadDir(sessionsDir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("memory: read sessions dir: %w", err)
	}

	migrated := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".jsonl") || strings.HasSuffix(name, ".jsonl.migrated") {
			continue
		}

		base := strings.TrimSuffix(name, ".jsonl")
		jsonlPath := filepath.Join(sessionsDir, name)
		metaPath := filepath.Join(sessionsDir, base+".meta.json")
		meta, metaErr := readJSONLMetaFile(metaPath, base)
		if metaErr != nil {
			log.Printf("memory: migrate jsonl: skip %s: %v", name, metaErr)
			continue
		}

		messages, readErr := readMessages(jsonlPath, meta.Skip)
		if readErr != nil {
			log.Printf("memory: migrate jsonl: skip %s: %v", name, readErr)
			continue
		}

		if err = importer.ImportSession(ctx, StoredSession{
			Key:       meta.Key,
			Summary:   meta.Summary,
			Messages:  messages,
			CreatedAt: meta.CreatedAt,
			UpdatedAt: meta.UpdatedAt,
		}); err != nil {
			return migrated, fmt.Errorf("memory: migrate jsonl %s: %w", name, err)
		}

		if renameErr := os.Rename(jsonlPath, jsonlPath+".migrated"); renameErr != nil {
			log.Printf("memory: migrate jsonl: rename %s: %v", name, renameErr)
		}
		if _, statErr := os.Stat(metaPath); statErr == nil {
			if renameErr := os.Rename(metaPath, metaPath+".migrated"); renameErr != nil {
				log.Printf("memory: migrate jsonl: rename %s: %v", filepath.Base(metaPath), renameErr)
			}
		}
		migrated++
	}

	return migrated, nil
}

// MigrateFromJSON reads legacy sessions/*.json files from sessionsDir,
// writes them into the Store, and renames each migrated file to
// .json.migrated as a backup. Returns the number of sessions migrated.
//
// Files that fail to parse are logged and skipped. Already-migrated
// files (.json.migrated) are ignored, making the function idempotent.
func MigrateFromJSON(
	ctx context.Context, sessionsDir string, store Store,
) (int, error) {
	entries, err := os.ReadDir(sessionsDir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("memory: read sessions dir: %w", err)
	}

	migrated := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		// Skip JSONL metadata files. They are part of the new storage format,
		// not legacy session snapshots, and re-importing them would overwrite
		// the paired .jsonl history with an empty message list.
		if strings.HasSuffix(name, ".meta.json") {
			continue
		}
		// Skip already-migrated files.
		if strings.HasSuffix(name, ".migrated") {
			continue
		}

		srcPath := filepath.Join(sessionsDir, name)

		data, readErr := os.ReadFile(srcPath)
		if readErr != nil {
			log.Printf("memory: migrate: skip %s: %v", name, readErr)
			continue
		}

		var sess jsonSession
		if parseErr := json.Unmarshal(data, &sess); parseErr != nil {
			log.Printf("memory: migrate: skip %s: %v", name, parseErr)
			continue
		}

		// Use the key from the JSON content, not the filename.
		// Filenames are sanitized (":" → "_") but keys are not.
		key := sess.Key
		if key == "" {
			key = strings.TrimSuffix(name, ".json")
		}

		// Use SetHistory (atomic replace) instead of per-message
		// AddFullMessage. This makes migration idempotent: if the
		// process crashes after writing messages but before the
		// rename below, a retry replaces the partial data cleanly
		// instead of duplicating messages.
		if setErr := store.SetHistory(ctx, key, sess.Messages); setErr != nil {
			return migrated, fmt.Errorf(
				"memory: migrate %s: set history: %w",
				name, setErr,
			)
		}

		if sess.Summary != "" {
			if sumErr := store.SetSummary(ctx, key, sess.Summary); sumErr != nil {
				return migrated, fmt.Errorf(
					"memory: migrate %s: set summary: %w",
					name, sumErr,
				)
			}
		}

		// Rename to .migrated as backup (not delete).
		renameErr := os.Rename(srcPath, srcPath+".migrated")
		if renameErr != nil {
			log.Printf("memory: migrate: rename %s: %v", name, renameErr)
		}

		migrated++
	}

	return migrated, nil
}

func readJSONLMetaFile(path, fallbackKey string) (sessionMeta, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return sessionMeta{Key: fallbackKey}, nil
	}
	if err != nil {
		return sessionMeta{}, fmt.Errorf("memory: read jsonl meta: %w", err)
	}
	var meta sessionMeta
	if err = json.Unmarshal(data, &meta); err != nil {
		return sessionMeta{}, fmt.Errorf("memory: decode jsonl meta: %w", err)
	}
	if meta.Key == "" {
		meta.Key = fallbackKey
	}
	return meta, nil
}
