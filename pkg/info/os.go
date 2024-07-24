package info

import (
	"fmt"
	"os"
	"regexp"
)

const (
	defaultUnknown = "_UNKNOWN_"
	// defaultUnknown         = "¯\\_( ͡° ͜ʖ ͡°)_/¯"
)

type OsInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	VersionId string `json:"versionId"`
}

func GetOsInfo() (*OsInfo, error) {
	const (
		OsReleasePath          = "/etc/os-release"
		regexFindOsNameVersion = `(?m)^NAME="(?P<name>[^"]+)"\s?|^VERSION="(?P<version>[^"]+)"|^VERSION_ID="?(?P<versid>[^"]+)"?\s`
	)
	info := OsInfo{
		Name:      defaultUnknown,
		Version:   defaultUnknown,
		VersionId: defaultUnknown,
	}
	content, err := os.ReadFile(OsReleasePath)
	if err != nil {
		return &info, fmt.Errorf("GetOsInfo: error reading %s. Err: %v", OsReleasePath, err)
	}
	r := regexp.MustCompile(regexFindOsNameVersion)
	matches := r.FindAllStringSubmatch(string(content), -1)

	for _, match := range matches {
		for i, name := range r.SubexpNames() {
			if i > 0 && match[i] != "" {
				switch name {
				case "name":
					info.Name = match[i]
				case "version":
					info.Version = match[i]
				case "versid":
					info.VersionId = match[i]
				}
			}
		}
	}
	return &info, nil
}

func GetOsUptime() (string, error) {
	uptimeResult := defaultUnknown
	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return uptimeResult, err
	}
	uptimeResult = string(content)
	return uptimeResult, nil
}
