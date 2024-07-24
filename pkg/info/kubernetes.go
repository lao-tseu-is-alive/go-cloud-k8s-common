package info

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/go_http"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	fmtErrK8sServiceHostEnvNotFound = "KUBERNETES_SERVICE_HOST ENV variable does not exist (🤔 maybe because not inside K8s ??).Err :  %v"
	caCertPath                      = "certificates/isrg-root-x1-cross-signed.pem"
	defaultReadTimeout              = 10 * time.Second
)

type K8sInfo struct {
	CurrentNamespace string `json:"current_namespace"`
	Version          string `json:"version"`
	Token            string `json:"token"`
	CaCert           string `json:"ca_cert"`
}

// GetKubernetesApiUrlFromEnv returns the k8s api url based on the content of standard env var :
//
//	KUBERNETES_SERVICE_HOST
//	KUBERNETES_SERVICE_PORT
//	in case the above ENV variables doesn't  exist the function returns an empty string and an error
func GetKubernetesApiUrlFromEnv() (string, error) {
	srvPort := 443
	k8sApiUrl := "https://"

	var err error
	val, exist := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	if !exist {
		return "", fmt.Errorf(fmtErrK8sServiceHostEnvNotFound, err)
	}
	k8sApiUrl = fmt.Sprintf("%s%s", k8sApiUrl, val)
	val, exist = os.LookupEnv("KUBERNETES_SERVICE_PORT")
	if exist {
		srvPort, err = strconv.Atoi(val)
		if err != nil {
			return "", fmt.Errorf("ERROR: CONFIG ENV PORT should contain a valid integer. %v", err)
		}
		if srvPort < 1 || srvPort > 65535 {
			return "", fmt.Errorf("ERROR: CONFIG ENV PORT should contain an integer between 1 and 65535. Err: %v", err)
		}
	}
	return fmt.Sprintf("%s:%d", k8sApiUrl, srvPort), nil
}

func GetKubernetesConnInfo(logger *log.Logger) (*K8sInfo, error) {
	const K8sServiceAccountPath = "/var/run/secrets/kubernetes.io/serviceaccount"
	K8sNamespacePath := fmt.Sprintf("%s/namespace", K8sServiceAccountPath)
	K8sTokenPath := fmt.Sprintf("%s/token", K8sServiceAccountPath)
	K8sCaCertPath := fmt.Sprintf("%s/ca.crt", K8sServiceAccountPath)

	k8sInfo := K8sInfo{
		CurrentNamespace: "",
		Version:          "",
		Token:            "",
		CaCert:           "",
	}

	K8sNamespace, err := os.ReadFile(K8sNamespacePath)
	if err != nil {
		return &k8sInfo, fmt.Errorf("GetKubernetesConnInfo: error reading namespace in %s. Err: %v", K8sNamespacePath, err)
	}
	k8sInfo.CurrentNamespace = string(K8sNamespace)

	K8sToken, err := os.ReadFile(K8sTokenPath)
	if err != nil {
		return &k8sInfo, fmt.Errorf("GetKubernetesConnInfo: error reading token in %s. Err: %v", K8sTokenPath, err)
	}
	k8sInfo.Token = string(K8sToken)

	K8sCaCert, err := os.ReadFile(K8sCaCertPath)
	if err != nil {
		return &k8sInfo, fmt.Errorf("GetKubernetesConnInfo: error reading Ca Cert in %s. Err: %v", K8sCaCertPath, err)
	}
	k8sInfo.CaCert = string(K8sCaCert)

	k8sUrl, err := GetKubernetesApiUrlFromEnv()
	if err != nil {
		return &k8sInfo, fmt.Errorf("GetKubernetesConnInfo: error reading GetKubernetesApiUrlFromEnv. Err: %v", err)
	}
	urlVersion := fmt.Sprintf("%s/openapi/v2", k8sUrl)
	res, err := GetJsonFromUrl(urlVersion, k8sInfo.Token, K8sCaCert, true, defaultReadTimeout, logger)
	if err != nil {

		logger.Printf("GetKubernetesConnInfo: error in GetJsonFromUrl(url:%s) err:%v", urlVersion, err)
		//return &k8sInfo, ErrorConfig{
		//	err: err,
		//	msg: fmt.Sprintf("GetKubernetesConnInfo: error doing GetJsonFromUrl(url:%s)", urlVersion),
		//}
	} else {
		logger.Printf("GetKubernetesConnInfo: successfully returned from GetJsonFromUrl(url:%s)", urlVersion)
		var myVersionRegex = regexp.MustCompile("{\"title\":\"(?P<title>.+)\",\"version\":\"(?P<version>.+)\"}")
		match := myVersionRegex.FindStringSubmatch(strings.TrimSpace(res[:150]))
		k8sVersionFields := make(map[string]string)
		for i, name := range myVersionRegex.SubexpNames() {
			if i != 0 && name != "" {
				k8sVersionFields[name] = match[i]
			}
		}
		k8sInfo.Version = fmt.Sprintf("%s, %s", k8sVersionFields["title"], k8sVersionFields["version"])
	}

	return &k8sInfo, nil
}

func GetJsonFromUrl(url string, bearerToken string, caCert []byte, allowInsecure bool, readTimeout time.Duration, logger *log.Logger) (string, error) {
	// Create a Bearer string by appending string access token
	var bearer = "Bearer " + bearerToken

	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Printf("Error on http.NewRequest [ERROR: %v]\n", err)
		return "", err
	}

	// add authorization header to the req
	req.Header.Add("Authorization", bearer)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:            caCertPool,
			InsecureSkipVerify: allowInsecure,
		},
	}
	// Send req using http Client
	client := &http.Client{
		Transport: tr,
		Timeout:   readTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Println("Error on sending request.\n[ERROR] -", err)
		return "", err
	}
	defer go_http.CloseBody(resp.Body, "GetJsonFromUrl", logger)
	if resp.StatusCode != http.StatusOK {
		logger.Printf("Error on response StatusCode is not OK Received StatusCode:%d\n", resp.StatusCode)
		return "", errors.New(fmt.Sprintf("Error on response StatusCode:%d\n", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Println("Error while reading the response bytes:", err)
		return "", err
	}
	return string(body), nil
}
func GetKubernetesInfo(l *log.Logger) (string, string, string) {
	k8sVersion := ""
	k8sCurrentNameSpace := ""
	k8sUrl := ""

	k8sUrl, err := GetKubernetesApiUrlFromEnv()
	if err != nil {
		l.Printf("💥💥 ERROR: 'GetKubernetesApiUrlFromEnv() returned an error : %+#v'", err)
	} else {
		kubernetesConnInfo, err := GetKubernetesConnInfo(l)
		if err != nil {
			l.Printf("💥💥 ERROR: 'GetKubernetesConnInfo() returned an error : %v'", err)
		}
		k8sVersion = kubernetesConnInfo.Version
		k8sCurrentNameSpace = kubernetesConnInfo.CurrentNamespace
	}

	return k8sUrl, k8sVersion, k8sCurrentNameSpace
}

func GetKubernetesLatestVersion(logger *log.Logger) (string, error) {
	k8sUrl := "https://kubernetes.io/"
	// Make an HTTP GET request to the Kubernetes releases page
	// Create a new request using http
	req, err := http.NewRequest("GET", k8sUrl, nil)
	if err != nil {
		logger.Printf("Error on http.NewRequest [ERROR: %v]\n", err)
		return "", err
	}
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		logger.Printf("Error on ReadFile(caCertPath) [ERROR: %v]\n", err)
		return "", err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}

	//tr := &http.Transport{ TLSClientConfig: &tls.Config{InsecureSkipVerify: true} }

	// add authorization header to the req
	// req.Header.Add("Authorization", bearer)
	// Send req using http Client
	client := &http.Client{
		Timeout:   defaultReadTimeout,
		Transport: tr,
	}

	resp, err := client.Do(req)

	if err != nil {
		logger.Println("Error on response.\n[ERROR] -", err)
		return fmt.Sprintf("GetKubernetesLatestVersion was unable to get content from %s, Error: %v", k8sUrl, err), err
	}
	defer go_http.CloseBody(resp.Body, "GetKubernetesLatestVersion", logger)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Println("Error while reading the response bytes:", err)
		return fmt.Sprintf("GetKubernetesLatestVersion got a problem reading the response from %s, Error: %v", k8sUrl, err), err
	}
	// Use a regular expression to extract the latest release number from the page
	re := regexp.MustCompile(`(?m)href=.+?>v(\d+\.\d+)`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	if matches == nil {
		return fmt.Sprintf("GetKubernetesLatestVersion was unable to find latest release number from %s", k8sUrl), nil
	}
	// Print only the release numbers
	maxVersion := 0.0
	for _, match := range matches {
		// fmt.Println(match[1])
		if val, err := strconv.ParseFloat(match[1], 32); err == nil {
			if val > maxVersion {
				maxVersion = val
			}
		}
	}
	// latestRelease := matches[0]
	// fmt.Printf("\nThe latest major release of Kubernetes is %T : %v+", latestRelease, latestRelease)
	return fmt.Sprintf("%2.2f", maxVersion), nil
}
