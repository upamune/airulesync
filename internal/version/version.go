package version

import (
	"fmt"
	"runtime"
	"time"
)

// These variables are set during build time using ldflags
var (
	Version   = "dev"
	Commit    = "none"
	BuildTime = ""
	GoVersion = runtime.Version()
	Platform  = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

// BuildInfo represents the build information for the application
type BuildInfo struct {
	Version   string
	Commit    string
	BuildTime string
	GoVersion string
	Platform  string
}

// GetBuildInfo returns the build information
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
		Platform:  Platform,
	}
}

// FormatBuildInfo returns a formatted string with the build information
func FormatBuildInfo() string {
	buildTime := BuildTime
	if buildTime == "" {
		buildTime = time.Now().Format(time.RFC3339)
	}

	return fmt.Sprintf(
		"airulesync version %s\ncommit: %s\nbuilt: %s\ngo version: %s\nplatform: %s",
		Version,
		Commit,
		buildTime,
		GoVersion,
		Platform,
	)
}
