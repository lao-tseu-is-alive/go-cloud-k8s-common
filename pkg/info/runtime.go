package info

import (
	"log"
	"os"
	"runtime"
	"strconv"
)

type RuntimeInfo struct {
	Hostname            string              `json:"hostname"`              // host name reported by the kernel.
	Pid                 int                 `json:"pid"`                   // process id of the caller.
	PPid                int                 `json:"ppid"`                  // process id of the caller's parent.
	Uid                 int                 `json:"uid"`                   // numeric user id of the caller.
	Appname             string              `json:"appname"`               // name of this application
	Version             string              `json:"version"`               // version of this application
	ParamName           string              `json:"param_name"`            // value of the name parameter (_NO_PARAMETER_NAME_ if name was not set)
	RemoteAddr          string              `json:"remote_addr"`           // remote client ip address
	RequestId           string              `json:"request_id"`            // globally unique request id
	GOOS                string              `json:"goos"`                  // operating system
	GOARCH              string              `json:"goarch"`                // architecture
	Runtime             string              `json:"runtime"`               // go runtime at compilation time
	NumGoroutine        string              `json:"num_goroutine"`         // number of go routines
	OsReleaseName       string              `json:"os_release_name"`       // Linux release Name or _UNKNOWN_
	OsReleaseVersion    string              `json:"os_release_version"`    // Linux release Version or _UNKNOWN_
	OsReleaseVersionId  string              `json:"os_release_version_id"` // Linux release VersionId or _UNKNOWN_
	NumCPU              string              `json:"num_cpu"`               // number of cpu
	Uptime              string              `json:"uptime"`                // tells how long this service was started based on an internal variable
	UptimeOs            string              `json:"uptime_os"`             // tells how long system was started based on /proc/uptime
	K8sApiUrl           string              `json:"k8s_api_url"`           // url for k8s api based KUBERNETES_SERVICE_HOST
	K8sVersion          string              `json:"k8s_version"`           // version of k8s cluster
	K8sLatestVersion    string              `json:"k8s_latest_version"`    // latest version announced in https://kubernetes.io/
	K8sCurrentNamespace string              `json:"k8s_current_namespace"` // k8s namespace of this container
	EnvVars             []string            `json:"env_vars"`              // environment variables
	Headers             map[string][]string `json:"headers"`               // received headers
}

func CollectRuntimeInfo(appName, version string, l *log.Logger) RuntimeInfo {
	hostName, err := os.Hostname()
	if err != nil {
		l.Printf("💥💥 ERROR: 'os.Hostname() returned an error : %v'", err)
		hostName = "#unknown#"
	}

	osReleaseInfo, err := GetOsInfo()
	if err != nil {
		l.Printf("💥💥 ERROR: 'GetOsInfo() returned an error : %+#v'", err)
	}

	uptimeOS, err := GetOsUptime()
	if err != nil {
		l.Printf("💥💥 ERROR: 'GetOsUptime() returned an error : %+#v'", err)
	}

	k8sApiUrl, k8sVersion, k8sCurrentNameSpace := GetKubernetesInfo(l)

	latestK8sVersion, err := GetKubernetesLatestVersion(l)
	if err != nil {
		l.Printf("💥💥 ERROR: 'GetKubernetesLatestVersion() returned an error : %+#v'", err)
	}

	return RuntimeInfo{
		Hostname:            hostName,
		Pid:                 os.Getpid(),
		PPid:                os.Getppid(),
		Uid:                 os.Getuid(),
		Appname:             appName,
		Version:             version,
		ParamName:           "_NO_PARAMETER_NAME_",
		RemoteAddr:          "",
		RequestId:           "",
		GOOS:                runtime.GOOS,
		GOARCH:              runtime.GOARCH,
		Runtime:             runtime.Version(),
		NumGoroutine:        strconv.FormatInt(int64(runtime.NumGoroutine()), 10),
		OsReleaseName:       osReleaseInfo.Name,
		OsReleaseVersion:    osReleaseInfo.Version,
		OsReleaseVersionId:  osReleaseInfo.VersionId,
		NumCPU:              strconv.FormatInt(int64(runtime.NumCPU()), 10),
		Uptime:              "",
		UptimeOs:            uptimeOS,
		K8sApiUrl:           k8sApiUrl,
		K8sVersion:          k8sVersion,
		K8sLatestVersion:    latestK8sVersion,
		K8sCurrentNamespace: k8sCurrentNameSpace,
		EnvVars:             os.Environ(),
		Headers:             map[string][]string{},
	}
}
