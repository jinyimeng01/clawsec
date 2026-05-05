package crawler

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/clawsec/clawsec/internal/logger"
)

// DirResult represents a directory busting result
type DirResult struct {
	URL        string
	StatusCode int
	Size       int
	Redirect   string
	Title      string
	Found      bool
}

// DirBuster performs directory enumeration
type DirBuster struct {
	client  *http.Client
	threads int
	statusFilter []int
	sizeFilter   map[int]bool
}

// NewDirBuster creates a new directory buster
func NewDirBuster(threads int, timeout int) *DirBuster {
	if threads <= 0 {
		threads = 20
	}
	if timeout <= 0 {
		timeout = 10
	}

	return &DirBuster{
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		threads:      threads,
		statusFilter: []int{200, 201, 204, 301, 302, 307, 308, 401, 403, 405, 500},
	}
}

// Scan scans a target for directories/files
func (d *DirBuster) Scan(ctx context.Context, baseURL string, wordlist []string) <-chan DirResult {
	results := make(chan DirResult, d.threads*2)

	go func() {
		defer close(results)

		var wg sync.WaitGroup
		workChan := make(chan string, d.threads)

		// Start workers
		for i := 0; i < d.threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for path := range workChan {
					select {
					case <-ctx.Done():
						return
					default:
					}

					result := d.checkURL(ctx, baseURL, path)
					if result.Found {
						select {
						case results <- result:
						case <-ctx.Done():
							return
						}
					}
				}
			}()
		}

		// Feed work
		go func() {
			defer close(workChan)
			for _, path := range wordlist {
				select {
				case workChan <- path:
				case <-ctx.Done():
					return
				}
			}
		}()

		wg.Wait()
	}()

	return results
}

func (d *DirBuster) checkURL(ctx context.Context, baseURL, path string) DirResult {
	url := baseURL
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	url += path

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return DirResult{URL: url}
	}

	req.Header.Set("User-Agent", "ClawSec/0.1.0")

	resp, err := d.client.Do(req)
	if err != nil {
		return DirResult{URL: url}
	}
	defer resp.Body.Close()

	result := DirResult{
		URL:        url,
		StatusCode: resp.StatusCode,
	}

	// Check if status is in filter
	found := false
	for _, status := range d.statusFilter {
		if resp.StatusCode == status {
			found = true
			break
		}
	}

	if !found {
		return result
	}

	// Get content length
	result.Size = int(resp.ContentLength)

	// Get redirect location
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		result.Redirect = resp.Header.Get("Location")
	}

	// Try to extract title
	if resp.StatusCode == 200 {
		// Simplified title extraction - read first 4KB
		body := make([]byte, 4096)
		n, _ := resp.Body.Read(body)
		bodyStr := string(body[:n])
		if start := strings.Index(bodyStr, "<title>"); start != -1 {
			if end := strings.Index(bodyStr[start:], "</title>"); end != -1 {
				result.Title = strings.TrimSpace(bodyStr[start+7 : start+end])
			}
		}
	}

	result.Found = true
	return result
}

// LoadWordlist loads a wordlist from file
func LoadWordlist(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" && !strings.HasPrefix(word, "#") {
			words = append(words, word)
		}
	}

	return words, scanner.Err()
}

// DefaultWordlist returns a small default wordlist
func DefaultWordlist() []string {
	return []string{
		"admin", "login", "api", "test", "dev", "staging", "prod",
		"backup", "config", "db", "database", "debug", "env",
		"phpmyadmin", "wp-admin", "wp-login", "administrator",
		"jenkins", "swagger", "api-docs", "graphql", "actuator",
		".env", ".git", ".svn", ".htaccess", "robots.txt",
		"sitemap.xml", "crossdomain.xml", "server-status",
		"phpinfo.php", "info.php", "test.php", "shell.php",
		"uploads", "images", "assets", "static", "public",
		"private", "secret", "internal", "dashboard",
		"console", "management", "manager", "tomcat",
		"jboss", "weblogic", "solr", "elasticsearch",
		"kibana", "grafana", "prometheus", "nagios",
		"zabbix", "cacti", "munin", "awstats",
		"webalizer", "phpMyAdmin", "myadmin", "pma",
	}
}

// SmartExtensions adds common extensions to paths
func SmartExtensions(paths []string) []string {
	extensions := []string{"", ".php", ".asp", ".aspx", ".jsp", ".html", ".txt", ".bak", ".old", ".zip", ".tar.gz", ".sql"}
	var result []string
	for _, path := range paths {
		for _, ext := range extensions {
			result = append(result, path+ext)
		}
	}
	return result
}
