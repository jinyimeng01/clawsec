package products

import (
	"bytes"
	"context"
	"crypto/tls"
	"net/http"
	"time"
)

// Product defines the interface for security product adapters
type Product interface {
	Name() string
	Connect(config Config) error
	IsConnected() bool
	Query(ctx context.Context, queryType string, params map[string]interface{}) ([]map[string]interface{}, error)
	Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error)
}

// Config holds product connection settings
type Config struct {
	URL      string
	APIKey   string
	Token    string
	Username string
	Password string
	Insecure bool
	Timeout  time.Duration
}

// BaseProduct provides common HTTP client functionality
type BaseProduct struct {
	Name_      string
	Config     Config
	Client     *http.Client
	Connected  bool
	Headers    map[string]string
}

func (b *BaseProduct) Name() string {
	return b.Name_
}

func (b *BaseProduct) IsConnected() bool {
	return b.Connected
}

func (b *BaseProduct) InitHTTPClient() {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: b.Config.Insecure},
	}
	b.Client = &http.Client{
		Transport: transport,
		Timeout:   b.Config.Timeout,
	}
	if b.Config.Timeout == 0 {
		b.Client.Timeout = 30 * time.Second
	}
}

func (b *BaseProduct) DoRequest(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	url := b.Config.URL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	for k, v := range b.Headers {
		req.Header.Set(k, v)
	}

	return b.Client.Do(req)
}

// Registry holds all product adapters
var Registry = make(map[string]Product)

// Register registers a product adapter
func Register(name string, product Product) {
	Registry[name] = product
}

// Get gets a product adapter by name
func Get(name string) (Product, bool) {
	p, ok := Registry[name]
	return p, ok
}

// List returns all registered product names
func List() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	return names
}

// LoadConfig loads product config from environment/global config
func LoadConfig(name string, cfg Config) Product {
	if p, ok := Get(name); ok {
		p.Connect(cfg)
		return p
	}
	return nil
}
