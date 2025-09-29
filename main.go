package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	apiURL   string
	username string
	password string
	token    string
)

// APIClient represents the Nginx Proxy Manager API client
type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

// AuthRequest represents the authentication request structure
type AuthRequest struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response structure
type AuthResponse struct {
	Token string `json:"token"`
}

// ProxyHost represents a proxy host configuration
type ProxyHost struct {
	ID                int      `json:"id"`
	DomainNames       []string `json:"domain_names"`
	ForwardScheme     string   `json:"forward_scheme"`
	ForwardHost       string   `json:"forward_host"`
	ForwardPort       int      `json:"forward_port"`
	AccessListID      int      `json:"access_list_id"`
	CertificateID     int      `json:"certificate_id"`
	SslForced         bool     `json:"ssl_forced"`
	CachingEnabled    bool     `json:"caching_enabled"`
	BlockExploits     bool     `json:"block_exploits"`
	AdvancedConfig    string   `json:"advanced_config"`
	Enabled           bool     `json:"enabled"`
	CreatedOn         string   `json:"created_on"`
	ModifiedOn        string   `json:"modified_on"`
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Authenticate performs authentication and stores the token
func (c *APIClient) Authenticate(username, password string) error {
	authReq := AuthRequest{
		Identity: username,
		Password: password,
	}

	jsonData, err := json.Marshal(authReq)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/tokens", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to make auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	c.Token = authResp.Token
	return nil
}

// makeAuthenticatedRequest makes an authenticated request to the API
func (c *APIClient) makeAuthenticatedRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.BaseURL+endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	return c.HTTPClient.Do(req)
}

// ListProxyHosts lists all proxy hosts
func (c *APIClient) ListProxyHosts() ([]ProxyHost, error) {
	resp, err := c.makeAuthenticatedRequest("GET", "/nginx/proxy-hosts", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list proxy hosts, status: %d", resp.StatusCode)
	}

	var hosts []ProxyHost
	if err := json.NewDecoder(resp.Body).Decode(&hosts); err != nil {
		return nil, fmt.Errorf("failed to decode proxy hosts: %w", err)
	}

	return hosts, nil
}

// CreateProxyHost creates a new proxy host
func (c *APIClient) CreateProxyHost(host ProxyHost) (*ProxyHost, error) {
	jsonData, err := json.Marshal(host)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proxy host: %w", err)
	}

	resp, err := c.makeAuthenticatedRequest("POST", "/nginx/proxy-hosts", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create proxy host, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var createdHost ProxyHost
	if err := json.NewDecoder(resp.Body).Decode(&createdHost); err != nil {
		return nil, fmt.Errorf("failed to decode created proxy host: %w", err)
	}

	return &createdHost, nil
}

// DeleteProxyHost deletes a proxy host by ID
func (c *APIClient) DeleteProxyHost(id int) error {
	resp, err := c.makeAuthenticatedRequest("DELETE", fmt.Sprintf("/nginx/proxy-hosts/%d", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete proxy host, status: %d", resp.StatusCode)
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "nginxproxymanager-cli",
	Short: "A CLI tool for managing Nginx Proxy Manager",
	Long:  `A command line interface for interacting with Nginx Proxy Manager API.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all proxy hosts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := NewAPIClient(apiURL)
		
		if err := client.Authenticate(username, password); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		hosts, err := client.ListProxyHosts()
		if err != nil {
			return fmt.Errorf("failed to list proxy hosts: %w", err)
		}

		fmt.Printf("Found %d proxy hosts:\n\n", len(hosts))
		for _, host := range hosts {
			fmt.Printf("ID: %d\n", host.ID)
			fmt.Printf("Domain Names: %v\n", host.DomainNames)
			fmt.Printf("Forward: %s://%s:%d\n", host.ForwardScheme, host.ForwardHost, host.ForwardPort)
			fmt.Printf("Enabled: %t\n", host.Enabled)
			fmt.Printf("SSL Forced: %t\n", host.SslForced)
			fmt.Println("---")
		}

		return nil
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new proxy host",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get required parameters first
		domainName, _ := cmd.Flags().GetString("domain")
		forwardHost, _ := cmd.Flags().GetString("forward-host")
		forwardPort, _ := cmd.Flags().GetInt("forward-port")
		forwardScheme, _ := cmd.Flags().GetString("forward-scheme")

		// Validate required parameters before authentication
		if domainName == "" || forwardHost == "" || forwardPort == 0 {
			return fmt.Errorf("domain, forward-host, and forward-port are required")
		}

		client := NewAPIClient(apiURL)
		
		if err := client.Authenticate(username, password); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		host := ProxyHost{
			DomainNames:   []string{domainName},
			ForwardScheme: forwardScheme,
			ForwardHost:   forwardHost,
			ForwardPort:   forwardPort,
			Enabled:       true,
			BlockExploits: true,
		}

		createdHost, err := client.CreateProxyHost(host)
		if err != nil {
			return fmt.Errorf("failed to create proxy host: %w", err)
		}

		fmt.Printf("Successfully created proxy host with ID: %d\n", createdHost.ID)
		fmt.Printf("Domain: %v\n", createdHost.DomainNames)
		fmt.Printf("Forward: %s://%s:%d\n", createdHost.ForwardScheme, createdHost.ForwardHost, createdHost.ForwardPort)

		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a proxy host by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate required parameters before authentication
		id, _ := cmd.Flags().GetInt("id")
		if id == 0 {
			return fmt.Errorf("id is required")
		}

		client := NewAPIClient(apiURL)
		
		if err := client.Authenticate(username, password); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		if err := client.DeleteProxyHost(id); err != nil {
			return fmt.Errorf("failed to delete proxy host: %w", err)
		}

		fmt.Printf("Successfully deleted proxy host with ID: %d\n", id)
		return nil
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&apiURL, "api-url", "a", "http://dockernuc:81/api", "Nginx Proxy Manager API URL")
	rootCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "Username for authentication")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password for authentication")

	// Create command flags
	createCmd.Flags().String("domain", "", "Domain name for the proxy host")
	createCmd.Flags().String("forward-host", "", "Forward host")
	createCmd.Flags().Int("forward-port", 0, "Forward port")
	createCmd.Flags().String("forward-scheme", "http", "Forward scheme (http or https)")

	// Delete command flags
	deleteCmd.Flags().Int("id", 0, "ID of the proxy host to delete")

	// Add commands
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)
}

func main() {
	// Check for environment variables
	if apiURL == "http://dockernuc:81/api" {
		if envURL := os.Getenv("NPM_API_URL"); envURL != "" {
			apiURL = envURL
		}
	}
	
	if username == "" {
		if envUsername := os.Getenv("NPM_USERNAME"); envUsername != "" {
			username = envUsername
		}
	}
	
	if password == "" {
		if envPassword := os.Getenv("NPM_PASSWORD"); envPassword != "" {
			password = envPassword
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}