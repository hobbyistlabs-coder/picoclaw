package main

import "os"

const (
	launcherUIModeStatic  = "static"
	launcherUIModeAPIOnly = "api-only"
)

func launcherUIMode() string {
	switch os.Getenv("JANE_AI_LAUNCHER_UI_MODE") {
	case "", launcherUIModeStatic:
		return launcherUIModeStatic
	case launcherUIModeAPIOnly:
		return launcherUIModeAPIOnly
	default:
		return launcherUIModeStatic
	}
}
