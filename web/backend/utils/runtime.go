package utils

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"jane/pkg/runtimepaths"
)

// GetDefaultConfigPath returns the default path to the picoclaw config file.
func GetDefaultConfigPath() string {
	if configPath := os.Getenv("JANE_AI_CONFIG"); configPath != "" {
		return configPath
	}
	if configPath := os.Getenv("PICOCLAW_CONFIG"); configPath != "" {
		return configPath
	}
	if home := os.Getenv("JANE_AI_HOME"); home != "" {
		return filepath.Join(home, "config.json")
	}
	if home := os.Getenv("PICOCLAW_HOME"); home != "" {
		return filepath.Join(home, "config.json")
	}
	return runtimepaths.ConfigPath()
}

// FindPicoclawBinary locates the picoclaw executable.
// Search order:
//  1. JANE_AI_BINARY or PICOCLAW_BINARY environment variable (explicit override)
//  2. Same directory as the current executable
//  3. Falls back to "jane-ai" or "picoclaw" on $PATH
func FindPicoclawBinary() string {
	if p := os.Getenv("JANE_AI_BINARY"); p != "" {
		if info, _ := os.Stat(p); info != nil && !info.IsDir() {
			return p
		}
	}
	if p := os.Getenv("PICOCLAW_BINARY"); p != "" {
		if info, _ := os.Stat(p); info != nil && !info.IsDir() {
			return p
		}
	}

	binaryNames := []string{"jane-ai", "picoclaw"}
	if runtime.GOOS == "windows" {
		binaryNames = []string{"jane-ai.exe", "picoclaw.exe"}
	}
	if exe, err := os.Executable(); err == nil {
		for _, binaryName := range binaryNames {
			candidate := filepath.Join(filepath.Dir(exe), binaryName)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate
			}
		}
	}

	if path, err := exec.LookPath(binaryNames[0]); err == nil {
		return path
	}
	return "picoclaw"
}

// GetLocalIP returns the local IP address of the machine.
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}

// OpenBrowser automatically opens the given URL in the default browser.
func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
