package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Config holds all application configuration
type Config struct {
	mu sync.RWMutex

	// Core settings
	ConfigPath  string
	OutputFormat string
	OutputFile  string
	Silent      bool
	Verbose     bool
	Debug       bool
	NoColor     bool

	// Scan settings
	Threads     int
	Timeout     int
	Retries     int
	RateLimit   int
	UserAgent   string
	RandomUA    bool
	Proxy       string
	Proxies     []string
	ForceProxy  bool

	// Network settings
	DNSResolver   string
	MaxRedirects  int
	InsecureSSL   bool
	FollowRedirects bool

	// Attack settings
	Authorized    bool
	Stealth       bool
	EncryptOutput bool
	EncryptKey    string

	// AI settings
	AIEnabled     bool
	AIEndpoint    string
	AIModel       string
	AIAPIKey      string

	// Product credentials (injected at runtime)
	Products map[string]ProductConfig

	// Raw config for extensibility
	raw map[string]interface{}
}

type ProductConfig struct {
	URL      string `yaml:"url" json:"url"`
	APIKey   string `yaml:"api_key" json:"api_key"`
	Token    string `yaml:"token" json:"token"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Insecure bool   `yaml:"insecure" json:"insecure"`
}

var (
	globalConfig *Config
	once         sync.Once
)

// Get returns the global config singleton
func Get() *Config {
	once.Do(func() {
		globalConfig = &Config{
			Products: make(map[string]ProductConfig),
			raw:      make(map[string]interface{}),
		}
	})
	return globalConfig
}

// SetRaw sets a raw config value
func (c *Config) SetRaw(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.raw[key] = value
}

// GetRaw gets a raw config value
func (c *Config) GetRaw(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.raw[key]
	return v, ok
}

// SetProduct sets product configuration
func (c *Config) SetProduct(name string, pc ProductConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Products[name] = pc
}

// GetProduct gets product configuration
func (c *Config) GetProduct(name string) (ProductConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	pc, ok := c.Products[name]
	return pc, ok
}

// DefaultConfigDir returns the default config directory
func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".clawsec"
	}
	return filepath.Join(home, ".clawsec")
}

// DefaultConfigPath returns the default config file path
func DefaultConfigPath() string {
	return filepath.Join(DefaultConfigDir(), "config.yaml")
}

// EnsureDir creates the config directory if it doesn't exist
func EnsureDir() error {
	dir := DefaultConfigDir()
	return os.MkdirAll(dir, 0755)
}

// LoadEnv loads configuration from environment variables
func LoadEnv() {
	cfg := Get()
	if v := os.Getenv("CLAWSEC_THREADS"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Threads)
	}
	if v := os.Getenv("CLAWSEC_TIMEOUT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Timeout)
	}
	if v := os.Getenv("CLAWSEC_PROXY"); v != "" {
		cfg.Proxy = v
	}
	if v := os.Getenv("CLAWSEC_USER_AGENT"); v != "" {
		cfg.UserAgent = v
	}
	if v := os.Getenv("CLAWSEC_AI_ENDPOINT"); v != "" {
		cfg.AIEndpoint = v
	}
	if v := os.Getenv("CLAWSEC_AI_API_KEY"); v != "" {
		cfg.AIAPIKey = v
	}

	// Product configs from env: CLAWSEC_<PRODUCT>_URL, CLAWSEC_<PRODUCT>_API_KEY
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "CLAWSEC_") {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimPrefix(parts[0], "CLAWSEC_")
			upperKey := strings.ToUpper(key)
			
			// Check if it's a product config
			for _, prod := range []string{"SAFELINE", "XRAY", "CLOUDWALKER", "TANSWER", "DDR"} {
				if strings.HasPrefix(upperKey, prod+"_") {
					field := strings.TrimPrefix(upperKey, prod+"_")
					prodName := strings.ToLower(prod)
					pc := cfg.Products[prodName]
					switch field {
					case "URL":
						pc.URL = parts[1]
					case "API_KEY":
						pc.APIKey = parts[1]
					case "TOKEN":
						pc.Token = parts[1]
					case "USERNAME":
						pc.Username = parts[1]
					case "PASSWORD":
						pc.Password = parts[1]
					}
					cfg.Products[prodName] = pc
					break
				}
			}
		}
	}
}

// InitDefaultConfig creates a default config file if it doesn't exist
func InitDefaultConfig() error {
	path := DefaultConfigPath()
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}

	if err := EnsureDir(); err != nil {
		return err
	}

	defaultCfg := `# ClawSec Configuration File
# Place this file at ~/.clawsec/config.yaml

# Core settings
output_format: color
timeout: 5
threads: 50
rate_limit: 150
user_agent: ""
random_ua: false

# Network
proxy: ""
force_proxy: false
insecure_ssl: false
follow_redirects: true
max_redirects: 10

# AI settings
ai:
  enabled: false
  endpoint: ""
  model: "claude-sonnet-4-20250514"
  api_key: ""

# Product configurations
# Uncomment and fill in the products you use:

# safeline:
#   url: "https://safeline.example.com"
#   api_key: "your-api-key"
#   insecure: true

# xray:
#   url: "https://xray.example.com"
#   api_key: "your-api-key"
#   insecure: true

# cloudwalker:
#   url: "https://cw.example.com"
#   token: "your-token"
#   insecure: true
`
	return os.WriteFile(path, []byte(defaultCfg), 0644)
}
