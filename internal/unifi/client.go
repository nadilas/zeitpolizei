package unifi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"
)

// Config holds UniFi controller configuration
type Config struct {
	BaseURL  string
	Username string
	Password string
	Site     string
	IsUDM    bool
	Insecure bool
}

// Client is a UniFi controller API client
type Client struct {
	config     Config
	httpClient *http.Client
	csrfToken  string
	mu         sync.RWMutex
}

// Client represents a network client from UniFi
type ClientInfo struct {
	MAC       string `json:"mac"`
	Name      string `json:"name,omitempty"`
	Hostname  string `json:"hostname,omitempty"`
	IP        string `json:"ip,omitempty"`
	TxBytes   int64  `json:"tx_bytes"`
	RxBytes   int64  `json:"rx_bytes"`
	Blocked   bool   `json:"blocked"`
	IsWired   bool   `json:"is_wired"`
	LastSeen  int64  `json:"last_seen"`
	Uptime    int64  `json:"uptime"`
	AssocTime int64  `json:"assoc_time"`
}

// NewClient creates a new UniFi API client
func NewClient(cfg Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.Insecure,
		},
	}

	httpClient := &http.Client{
		Jar:       jar,
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	// Normalize base URL
	cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")

	return &Client{
		config:     cfg,
		httpClient: httpClient,
	}, nil
}

// apiPrefix returns the API prefix based on controller type
func (c *Client) apiPrefix() string {
	if c.config.IsUDM {
		return "/proxy/network"
	}
	return ""
}

// buildURL constructs the full URL for an API endpoint
func (c *Client) buildURL(endpoint string) string {
	return c.config.BaseURL + c.apiPrefix() + endpoint
}

// doRequest performs an HTTP request with proper headers
func (c *Client) doRequest(method, url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add CSRF token for UDM if we have one
	c.mu.RLock()
	csrfToken := c.csrfToken
	c.mu.RUnlock()

	if csrfToken != "" {
		req.Header.Set("X-CSRF-Token", csrfToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Extract and store CSRF token from response
	if token := resp.Header.Get("X-CSRF-Token"); token != "" {
		c.mu.Lock()
		c.csrfToken = token
		c.mu.Unlock()
	}

	return resp, nil
}

// Login authenticates with the UniFi controller
func (c *Client) Login() error {
	loginPayload := map[string]string{
		"username": c.config.Username,
		"password": c.config.Password,
	}

	// UDM uses different login endpoint
	var loginURL string
	if c.config.IsUDM {
		loginURL = c.config.BaseURL + "/api/auth/login"
	} else {
		loginURL = c.config.BaseURL + "/api/login"
	}

	resp, err := c.doRequest("POST", loginURL, loginPayload)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetClients retrieves all connected clients from UniFi
func (c *Client) GetClients() ([]ClientInfo, error) {
	url := c.buildURL(fmt.Sprintf("/api/s/%s/stat/sta", c.config.Site))

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get clients: status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []ClientInfo `json:"data"`
		Meta struct {
			RC string `json:"rc"`
		} `json:"meta"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}

// GetClient retrieves a specific client by MAC address
func (c *Client) GetClient(mac string) (*ClientInfo, error) {
	clients, err := c.GetClients()
	if err != nil {
		return nil, err
	}

	mac = strings.ToLower(mac)
	for _, client := range clients {
		if strings.ToLower(client.MAC) == mac {
			return &client, nil
		}
	}

	return nil, nil // Client not found (not currently connected)
}

// BlockClient blocks a client by MAC address
func (c *Client) BlockClient(mac string) error {
	url := c.buildURL(fmt.Sprintf("/api/s/%s/cmd/stamgr", c.config.Site))

	payload := map[string]string{
		"cmd": "block-sta",
		"mac": strings.ToLower(mac),
	}

	resp, err := c.doRequest("POST", url, payload)
	if err != nil {
		return fmt.Errorf("block request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to block client: status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// UnblockClient unblocks a client by MAC address
func (c *Client) UnblockClient(mac string) error {
	url := c.buildURL(fmt.Sprintf("/api/s/%s/cmd/stamgr", c.config.Site))

	payload := map[string]string{
		"cmd": "unblock-sta",
		"mac": strings.ToLower(mac),
	}

	resp, err := c.doRequest("POST", url, payload)
	if err != nil {
		return fmt.Errorf("unblock request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to unblock client: status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetBlockedClients retrieves all blocked clients
func (c *Client) GetBlockedClients() ([]ClientInfo, error) {
	// Get all known clients (including blocked ones)
	url := c.buildURL(fmt.Sprintf("/api/s/%s/rest/user", c.config.Site))

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get users: status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []ClientInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Filter to blocked clients only
	var blocked []ClientInfo
	for _, client := range result.Data {
		if client.Blocked {
			blocked = append(blocked, client)
		}
	}

	return blocked, nil
}

// IsClientBlocked checks if a client is currently blocked
func (c *Client) IsClientBlocked(mac string) (bool, error) {
	blocked, err := c.GetBlockedClients()
	if err != nil {
		return false, err
	}

	mac = strings.ToLower(mac)
	for _, client := range blocked {
		if strings.ToLower(client.MAC) == mac {
			return true, nil
		}
	}

	return false, nil
}

// GetAllKnownClients retrieves all known clients (connected and historical)
func (c *Client) GetAllKnownClients() ([]ClientInfo, error) {
	url := c.buildURL(fmt.Sprintf("/api/s/%s/rest/user", c.config.Site))

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get users: status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []ClientInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}
