package gohttp

type AppInfo struct {
	App        string `json:"app"`
	Version    string `json:"version"`
	Repository string `json:"repository"`
}

type VersionWriter interface {
	GetVersionInfo() AppInfo
}

// SimpleVersionWriter Create a struct that will implement the VersionWriter interface
type SimpleVersionWriter struct {
	Version AppInfo
}

// GetVersionInfo returns the version information of the application.
func (s SimpleVersionWriter) GetVersionInfo() AppInfo {
	return s.Version
}

// NewSimpleVersionWriter is a constructor that initializes the VersionWriter interface
func NewSimpleVersionWriter(app, ver, repo string) *SimpleVersionWriter {

	return &SimpleVersionWriter{
		Version: AppInfo{
			App:        app,
			Version:    ver,
			Repository: repo,
		},
	}
}
