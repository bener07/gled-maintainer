package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ClientConfig holds connection parameters for a remote equipment server.
type ClientConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	Key  string `json:"key"`
}

// NewConnection creates and returns a new ClientConfig.
func NewConnection(host string, port int, key string) (*ClientConfig, error) {
	if host == "" {
		return nil, fmt.Errorf("host não pode estar vazio")
	}
	return &ClientConfig{
		Host: host,
		Port: port,
		Key:  key,
	}, nil
}

// baseURL returns the base URL for the client.
func (c *ClientConfig) baseURL() string {
	return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
}

// httpClient returns an HTTP client with a timeout.
func (c *ClientConfig) httpClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

// Ping checks if the remote client is reachable by calling its /health endpoint.
func (c *ClientConfig) Ping() error {
	resp, err := c.httpClient().Get(c.baseURL() + "/health")
	if err != nil {
		return fmt.Errorf("erro de conectividade: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cliente respondeu com status %s", resp.Status)
	}
	return nil
}

// GetStatus fetches status information from the remote client.
func (c *ClientConfig) GetStatus() (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", c.baseURL()+"/status", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.Key)

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("erro ao descodificar resposta: %w", err)
	}
	return result, nil
}
