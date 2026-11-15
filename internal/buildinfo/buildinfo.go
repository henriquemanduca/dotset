// Package buildinfo exposes version metadata injected at build time.
//
// The values are overridden by the Makefile through -ldflags, e.g.:
//
//	go build -ldflags "-X github.com/henrique/dotset/internal/buildinfo.Version=v1.2.3"
package buildinfo

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version or git description of the build.
	Version = "dev"
	// Commit is the short git commit hash of the build.
	Commit = "none"
	// Date is the UTC build timestamp.
	Date = "unknown"
)

// Info describes the build metadata of the running binary.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns the build metadata for the running binary.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String formats the build metadata as a single human-readable line.
func (i Info) String() string {
	return fmt.Sprintf("dotset %s (commit: %s, built at: %s, %s/%s, %s)",
		i.Version, i.Commit, i.Date, i.OS, i.Arch, i.GoVersion)
}
