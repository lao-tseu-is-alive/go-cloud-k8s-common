package config

import (
	"fmt"
	"os"
	"strings"
)

// GetTlsMode returns the configured TLS mode ("none", "manual", "autocert").
// It panics if the mode is invalid.
func GetTlsMode() string {
	mode := os.Getenv("TLS_MODE")
	switch mode {
	case "", "none":
		return "none"
	case "manual":
		return "manual"
	case "autocert":
		return "autocert"
	default:
		panic(fmt.Sprintf("💥💥 ERROR: Invalid TLS_MODE '%s'. Must be one of 'none', 'manual', or 'autocert'.", mode))
	}
}

// GetTlsCertFile returns the path to the TLS certificate file.
func GetTlsCertFile() string {
	return os.Getenv("TLS_CERT_FILE")
}

// GetTlsKeyFile returns the path to the TLS key file.
func GetTlsKeyFile() string {
	return os.Getenv("TLS_KEY_FILE")
}

// GetAutocertHosts returns the list of hosts for autocert.
func GetAutocertHosts() []string {
	hosts := os.Getenv("AUTOCERT_HOSTS")
	if hosts == "" {
		return nil
	}
	return strings.Split(hosts, ",")
}

// GetAutocertDir returns the directory to cache certs.
func GetAutocertDir() string {
	dir := os.Getenv("AUTOCERT_DIR")
	if dir == "" {
		return "/certs" // A sensible default
	}
	return dir
}
