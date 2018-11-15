package version

var (
	// Version is overridden by go build flags
	Version = "UNKNOWN"
	// BuildDate is overridden by go build flags
	BuildDate = "UNKNOWN"
	// Package is the package name used to build this project, overridden by go build flags
	Package = "UNKNOWN"
	// Branch this was built on
	Branch = "???"
	// Commit this was built at
	Commit = "???"
)
