package gohttp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
	"github.com/stretchr/testify/assert"
)

func TestServer_startHttpsManualServer(t *testing.T) {
	// --- Test Setup ---
	certPath, keyPath, err := createTempCert()
	if err != nil {
		t.Fatalf("Failed to create temp cert: %v", err)
	}
	defer os.Remove(certPath)
	defer os.Remove(keyPath)

	// --- Set environment variables for this test ---
	os.Setenv("TLS_MODE", "manual")
	os.Setenv("TLS_CERT_FILE", certPath)
	os.Setenv("TLS_KEY_FILE", keyPath)
	// Minimal env vars for server to start
	t.Setenv("JWT_SECRET", "a-very-secret-jwt-token-for-testing")
	t.Setenv("JWT_ISSUER_ID", "go-cloud-k8s-common-test-issuer")
	t.Setenv("JWT_CONTEXT_KEY", "testcontext")
	t.Setenv("ADMIN_PASSWORD", "Password123!")
	t.Setenv("JWT_AUTH_URL", "/login")
	defer os.Unsetenv("TLS_MODE")
	defer os.Unsetenv("TLS_CERT_FILE")
	defer os.Unsetenv("TLS_KEY_FILE")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("JWT_ISSUER_ID")
	defer os.Unsetenv("JWT_CONTEXT_KEY")
	defer os.Unsetenv("ADMIN_PASSWORD")
	defer os.Unsetenv("JWT_AUTH_URL")

	// --- Create Mock Dependencies ---
	myJwt := NewJwtChecker(
		config.GetJwtSecretFromEnvOrPanic(),
		config.GetJwtIssuerFromEnvOrPanic(),
		"test-app",
		config.GetJwtContextKeyFromEnvOrPanic(),
		60,
		l)
	myAuthenticator := NewSimpleAdminAuthenticator("admin", "admin", "admin@example.com", 1, myJwt)
	myVersionReader := NewSimpleVersionReader("test-app", "v0.0.1", "", "", "", "/login")

	// --- Run Test ---
	assert.NotPanics(t, func() {
		// Create the server with valid mock dependencies
		server := CreateNewServerFromEnvOrFail(9898, "127.0.0.1", myAuthenticator, myJwt, myVersionReader, l)
		// We're just testing the startup logic, so we start it in a goroutine
		// and give it a moment to initialize before the test finishes.
		go server.StartServer()
		time.Sleep(100 * time.Millisecond)
	})
}

// createTempCert generates a temporary self-signed certificate for testing.
// (This function remains the same as previously provided)
func createTempCert() (certPath, keyPath string, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Co"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", err
	}

	certOut, err := os.CreateTemp("", "cert.pem")
	if err != nil {
		return "", "", err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyOut, err := os.CreateTemp("", "key.pem")
	if err != nil {
		return "", "", err
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return certOut.Name(), keyOut.Name(), nil
}

// Re-using the init() from your original handlers_test.go to setup logger
func init() {
	var err error
	if l, err = golog.NewLogger("simple", golog.DebugLevel, "test_handlers"); err != nil {
		log.Fatalf("💥💥 error golog.NewLogger error: %v'\n", err)
	}
}
