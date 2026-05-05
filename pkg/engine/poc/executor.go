package poc

import (
	"context"
	"sync"
	"time"

	"github.com/clawsec/clawsec/internal/logger"
)

// Executor executes PoC templates
type Executor struct {
	httpClient *HTTPClient
	threads    int
	timeout    int
	results    chan *Result
}

// NewExecutor creates a new PoC executor
func NewExecutor(threads, timeout int) *Executor {
	if threads <= 0 {
		threads = 25
	}
	if timeout <= 0 {
		timeout = 10
	}

	return &Executor{
		httpClient: NewHTTPClient(timeout, true, 10),
		threads:    threads,
		timeout:    timeout,
		results:    make(chan *Result, threads*2),
	}
}

// Execute runs a template against a list of targets
func (e *Executor) Execute(ctx context.Context, template *Template, targets []string) <-chan *Result {
	go func() {
		defer close(e.results)

		var wg sync.WaitGroup
		targetChan := make(chan string, e.threads)

		// Start workers
		for i := 0; i < e.threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for target := range targetChan {
					select {
					case <-ctx.Done():
						return
					default:
					}

					results := e.executeTemplate(ctx, template, target)
					for _, result := range results {
						select {
						case e.results <- result:
						case <-ctx.Done():
							return
						}
					}
				}
			}()
		}

		// Feed targets
		go func() {
			defer close(targetChan)
			for _, target := range targets {
				select {
				case targetChan <- target:
				case <-ctx.Done():
					return
				}
			}
		}()

		wg.Wait()
	}()

	return e.results
}

func (e *Executor) executeTemplate(ctx context.Context, template *Template, target string) []*Result {
	var results []*Result

	// Initialize variables
	vars := InitTemplateVariables(target)

	// Add template variables
	for k, v := range template.Variables {
		vars[k] = ExpandVariables(v, vars)
	}

	// Execute HTTP requests
	for _, httpReq := range template.HTTP {
		resp, err := e.httpClient.ExecuteHTTPRequest(ctx, httpReq, target, vars)
		if err != nil {
			logger.Debugf("HTTP request failed for %s: %v", target, err)
			continue
		}

		// Match response
		matched, matcherResults := MatchResponseWithCondition(httpReq.Matchers, httpReq.MatchersCondition, resp, vars)
		if matched {
			// Extract data
			extracted := ExtractData(httpReq.Extractors, resp, vars)
			for k, v := range extracted {
				vars[k] = v
			}

			// Build matcher name
			matcherName := ""
			for name, m := range matcherResults {
				if m {
					matcherName = name
					break
				}
			}

			// Build extracted results
			var extractedResults []string
			for _, v := range extracted {
				switch val := v.(type) {
				case string:
					extractedResults = append(extractedResults, val)
				case []string:
					extractedResults = append(extractedResults, val...)
				}
			}

			result := &Result{
				TemplateID:       template.ID,
				TemplatePath:     template.Path,
				Info:             template.Info,
				Type:             "http",
				Host:             extractHostname(target),
				URL:              target,
				MatchedAt:        time.Now(),
				MatcherName:      matcherName,
				ExtractedResults: extractedResults,
				Meta: map[string]interface{}{
					"status_code": resp.StatusCode,
					"duration_ms": resp.Duration,
					"size":        resp.Size,
				},
			}
			results = append(results, result)

			if httpReq.StopAtFirstMatch {
				break
			}
		}
	}

	return results
}

// ExecuteMultiple runs multiple templates against targets
func (e *Executor) ExecuteMultiple(ctx context.Context, templates []*Template, targets []string) <-chan *Result {
	results := make(chan *Result, e.threads*2)

	go func() {
		defer close(results)
		var wg sync.WaitGroup

		for _, tmpl := range templates {
			tmpl := tmpl // capture
			wg.Add(1)
			go func() {
				defer wg.Done()
				executor := NewExecutor(e.threads, e.timeout)
				for result := range executor.Execute(ctx, tmpl, targets) {
					select {
					case results <- result:
					case <-ctx.Done():
						return
					}
				}
			}()
		}

		wg.Wait()
	}()

	return results
}
