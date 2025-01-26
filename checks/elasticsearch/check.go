package elasticsearch

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Config is the Elasticsearch checker configuration settings container.
type Config struct {
	DSN         string // DSN is the Elasticsearch connection DSN. Required.
	Password    string // Password is the Elasticsearch connection password. Required.
	SSLCertPath string // SSLCertPath is the path to the SSL certificate to use for the connection. Optional.
}

// New creates a new Elasticsearch health check that verifies the status of the cluster.
func New(config Config) func(ctx context.Context) error {
	if config.DSN == "" || config.Password == "" {
		return func(ctx context.Context) error {
			return fmt.Errorf("elasticsearch DSN and password are required")
		}
	}

	client, err := makeHTTPClient(config.SSLCertPath)
	if err != nil {
		return func(ctx context.Context) error {
			return fmt.Errorf("failed to create Elasticsearch HTTP client: %w", err)
		}
	}

	return func(ctx context.Context) error {
		return checkHealth(ctx, client, config.DSN, config.Password)
	}
}

func makeHTTPClient(sslCertPath string) (*http.Client, error) {
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}

	// If SSLCert is set, configure the client to use it.
	// Otherwise, skip TLS verification.

	if sslCertPath != "" {
		cert, err := tls.LoadX509KeyPair(sslCertPath, sslCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load Elasticsearch SSL certificate: %w", err)
		}

		// Configure the client to use the certificate.
		httpTransport := &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		}

		httpClient.Transport = httpTransport
		return &httpClient, nil
	}

	// Configure the client to skip TLS verification.
	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	httpClient.Transport = httpTransport

	return &httpClient, nil
}

func checkHealth(ctx context.Context, client *http.Client, dsn string, password string) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://%s/_cluster/health", dsn),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create Elasticsearch health check request: %w", err)
	}

	req.SetBasicAuth("elastic", password)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Elasticsearch health check request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code from Elasticsearch health check: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Elasticsearch health check response: %w", err)
	}

	healthResp := struct {
		Status string `json:"status"`
	}{}

	if err := json.Unmarshal(body, &healthResp); err != nil {
		return fmt.Errorf("failed to parse Elasticsearch health check response: %w", err)
	}

	if healthResp.Status != "green" {
		return fmt.Errorf("elasticsearch cluster status is not green: %s", healthResp.Status)
	}
	return nil
}
