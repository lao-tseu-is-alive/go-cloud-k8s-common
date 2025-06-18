package gohttp

type AppInfo struct {
	App        string `json:"app"`
	Version    string `json:"version"`
	BuildStamp string `json:"buildStamp"`
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
	AuthUrl    string `json:"authUrl"`
}

type VersionReader interface {
	GetVersionInfo() AppInfo
}

// SimpleVersionWriter Create a struct that will implement the VersionReader interface
type SimpleVersionWriter struct {
	Info AppInfo
}

// GetVersionInfo returns the version information of the application.
func (s SimpleVersionWriter) GetVersionInfo() AppInfo {
	return s.Info
}

// NewSimpleVersionReader is a constructor that initializes the VersionReader interface
func NewSimpleVersionReader(app, ver, repo, rev, buildStamp, authUrl string) *SimpleVersionWriter {

	return &SimpleVersionWriter{
		Info: AppInfo{
			App:        app,
			Version:    ver,
			BuildStamp: buildStamp,
			Revision:   rev,
			Repository: repo,
			AuthUrl:    authUrl,
		},
	}
}
