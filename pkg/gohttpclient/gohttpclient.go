package gohttpclient

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"io"
	"net/http"
	"time"
)

// WaitForHttpServer attempts to establish a http TCP connection to listenAddress
// in a given amount of time. It returns upon a successful connection;
// otherwise exits with an error.
func WaitForHttpServer(url string, waitDuration time.Duration, numRetries int, l golog.MyLogger) {
	l.Debug("INFO: 'WaitForHttpServer Will wait for server to be up at %s for %v seconds, with %d retries'\n", url, waitDuration.Seconds(), numRetries)
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	for i := 0; i < numRetries; i++ {
		resp, err := httpClient.Get(url)

		if err != nil {
			if i > 0 {
				l.Info("\nWaitForHttpServer: httpClient.Get(%s) retry:[%d], %v\n", url, i, err)
			}
			time.Sleep(waitDuration)
			continue
		}
		// All seems is good
		l.Info("OK: Server responded after %d retries, with status code %d ", i, resp.StatusCode)
		return
	}
	l.Fatal("Server %s not ready up after %d attempts", url, numRetries)
}

func GetJsonFromUrlWithBearerAuth(url string, token string, caCert []byte, allowInsecure bool, readTimeout time.Duration, l golog.MyLogger) (string, error) {
	// Create a Bearer string by appending string access token
	var bearer = "Bearer " + token

	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		l.Error("Error on http.NewRequest [ERROR: %v]\n", err)
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
		l.Error("GetJsonFromUrlWithBearerAuth: Error on response.\n[ERROR] -", err)
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			l.Error("GetJsonFromUrlWithBearerAuth: Error on Body.Close().\n[ERROR] -", err)
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		l.Error("Error on response StatusCode is not OK Received StatusCode:%d\n", resp.StatusCode)
		return "", errors.New(fmt.Sprintf("Error on response StatusCode:%d\n", resp.StatusCode))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error("GetJsonFromUrlWithBearerAuth: Error while reading the response bytes:", err)
		return "", err
	}
	return string(body), nil
}
