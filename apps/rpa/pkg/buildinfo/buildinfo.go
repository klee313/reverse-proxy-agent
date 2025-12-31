// Package buildinfo exposes build metadata for logs and diagnostics.
package buildinfo

import (
	"runtime"
	"runtime/debug"
)

// These are intended to be set via -ldflags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

type Info struct {
	Version   string
	Commit    string
	Date      string
	GoVersion string
}

func Current() Info {
	info := Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: runtime.Version(),
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if info.Version == "dev" && bi.Main.Version != "" {
			info.Version = bi.Main.Version
		}
	}
	return info
}

func Fields() map[string]any {
	info := Current()
	return map[string]any{
		"version":    info.Version,
		"commit":     info.Commit,
		"build_time": info.Date,
		"go_version": info.GoVersion,
	}
}
