package alpaca

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"jane/pkg/runtimepaths"
)

func resolveConfig(keyID, secretKey, baseURL string) (string, string, string) {
	keyID = firstNonEmpty(keyID, os.Getenv("ALPACA_API_KEY"), envFileValue("ALPACA_API_KEY"))
	secretKey = firstNonEmpty(secretKey, os.Getenv("ALPACA_SECRET_KEY"), envFileValue("ALPACA_SECRET_KEY"))
	baseURL = firstNonEmpty(baseURL, os.Getenv("ALPACA_API_URL"), envFileValue("ALPACA_API_URL"))
	return keyID, secretKey, baseURL
}

// resolveDataURL returns the market data base URL from env/config.
// ALPACA_DATA_URL is intentionally separate from ALPACA_API_URL so that the
// trading endpoint (e.g. paper-api.alpaca.markets) is never accidentally sent
// to the market data client, which would cause HTTP 404 for all data calls.
// When empty the Alpaca SDK defaults to https://data.alpaca.markets.
func resolveDataURL() string {
	return firstNonEmpty(os.Getenv("ALPACA_DATA_URL"), envFileValue("ALPACA_DATA_URL"))
}

func envFileValue(key string) string {
	for _, path := range candidateEnvFiles() {
		values, err := loadEnvFile(path)
		if err == nil && strings.TrimSpace(values[key]) != "" {
			return strings.TrimSpace(values[key])
		}
	}
	return ""
}

func candidateEnvFiles() []string {
	var paths []string
	add := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		for _, existing := range paths {
			if existing == path {
				return
			}
		}
		paths = append(paths, path)
	}
	add(firstNonEmpty(os.Getenv("JANE_AI_ALPACA_ENV_FILE"), os.Getenv("PICOCLAW_ALPACA_ENV_FILE")))
	if wd, err := os.Getwd(); err == nil {
		for _, dir := range walkUpDirs(wd) {
			add(filepath.Join(dir, ".env.alpaca"))
		}
	}
	if exe, err := os.Executable(); err == nil {
		for _, dir := range walkUpDirs(filepath.Dir(exe)) {
			add(filepath.Join(dir, ".env.alpaca"))
		}
	}
	add(filepath.Join(runtimepaths.HomeDir(), ".env.alpaca"))
	return paths
}

func walkUpDirs(start string) []string {
	var dirs []string
	current := filepath.Clean(start)
	for {
		dirs = append(dirs, current)
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return dirs
}

func loadEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return values, scanner.Err()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
