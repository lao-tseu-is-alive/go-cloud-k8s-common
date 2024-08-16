package gohttp

type AppInfo struct {
	App        string `json:"app"`
	Version    string `json:"version"`
	Repository string `json:"repository"`
	Build      string `json:"build"`
}

type VersionReader interface {
	GetVersionInfo() AppInfo
}

// SimpleVersionWriter Create a struct that will implement the VersionReader interface
type SimpleVersionWriter struct {
	Version AppInfo
}

// GetVersionInfo returns the version information of the application.
func (s SimpleVersionWriter) GetVersionInfo() AppInfo {
	return s.Version
}

// NewSimpleVersionReader is a constructor that initializes the VersionReader interface
func NewSimpleVersionReader(app, ver, repo, build string) *SimpleVersionWriter {

	return &SimpleVersionWriter{
		Version: AppInfo{
			App:        app,
			Version:    ver,
			Repository: repo,
			Build:      build,
		},
	}
}
