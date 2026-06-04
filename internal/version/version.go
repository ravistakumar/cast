// Package version exposes the build version string.
package version

// Version is overridden at build time via -ldflags.
var Version = "dev"

// String returns the current version.
func String() string { return Version }
