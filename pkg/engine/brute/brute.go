package brute

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/clawsec/clawsec/internal/logger"
)

// Result represents a brute force attempt result
type Result struct {
	Target   string
	Protocol string
	Username string
	Password string
	Success  bool
	Error    error
	Banner   string
	RTT      time.Duration
}

// Options configures the brute force attack
type Options struct {
	Targets       []string
	Usernames     []string
	Passwords     []string
	Threads       int
	Timeout       time.Duration
	Delay         time.Duration
	StopOnSuccess bool
	Mode          string // cartesian, pair, cycle-user, cycle-pass
}

// Protocol defines the interface for protocol-specific brute forcers
type Protocol interface {
	Name() string
	Try(ctx context.Context, target, username, password string) (Result, error)
}

// Runner manages brute force attacks across protocols
type Runner struct {
	protocol Protocol
	opts     Options
	results  chan Result
	progress *Progress
}

// Progress tracks brute force progress
type Progress struct {
	Total     int64
	Attempted int64
	Success   int64
	Failed    int64
	StartTime time.Time
}

// NewRunner creates a new brute force runner
func NewRunner(protocol Protocol, opts Options) *Runner {
	if opts.Threads <= 0 {
		opts.Threads = 10
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Second
	}

	return &Runner{
		protocol: protocol,
		opts:     opts,
		results:  make(chan Result, opts.Threads*2),
		progress: &Progress{
			StartTime: time.Now(),
		},
	}
}

// Run starts the brute force attack
func (r *Runner) Run(ctx context.Context) <-chan Result {
	// Generate credential combinations
	combinations := r.generateCombinations()
	atomic.StoreInt64(&r.progress.Total, int64(len(combinations)))

	logger.Infof("Starting %s brute force - targets: %d, users: %d, passwords: %d, total: %d, threads: %d",
		r.protocol.Name(), len(r.opts.Targets), len(r.opts.Usernames), len(r.opts.Passwords),
		len(combinations), r.opts.Threads)

	go func() {
		defer close(r.results)

		var wg sync.WaitGroup
		workChan := make(chan [3]string, r.opts.Threads) // [target, username, password]
		stopFlag := int32(0)

		// Start workers
		for i := 0; i < r.opts.Threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for work := range workChan {
					if ctx.Err() != nil {
						return
					}
					if r.opts.StopOnSuccess && atomic.LoadInt32(&stopFlag) == 1 {
						return
					}

					target, username, password := work[0], work[1], work[2]

					start := time.Now()
					result, err := r.protocol.Try(ctx, target, username, password)
					result.RTT = time.Since(start)

					if err != nil {
						result.Error = err
						result.Success = false
						atomic.AddInt64(&r.progress.Failed, 1)
					} else if result.Success {
						atomic.AddInt64(&r.progress.Success, 1)
						if r.opts.StopOnSuccess {
							atomic.StoreInt32(&stopFlag, 1)
						}
					}

					atomic.AddInt64(&r.progress.Attempted, 1)

					select {
					case r.results <- result:
					case <-ctx.Done():
						return
					}

					// Delay between attempts
					if r.opts.Delay > 0 {
						time.Sleep(r.opts.Delay)
					}
				}
			}()
		}

		// Feed work
		go func() {
			defer close(workChan)
			for _, combo := range combinations {
				if r.opts.StopOnSuccess && atomic.LoadInt32(&stopFlag) == 1 {
					return
				}
				select {
				case workChan <- combo:
				case <-ctx.Done():
					return
				}
			}
		}()

		wg.Wait()
		logger.Infof("Brute force complete - attempted: %d, success: %d, failed: %d, duration: %v",
			atomic.LoadInt64(&r.progress.Attempted),
			atomic.LoadInt64(&r.progress.Success),
			atomic.LoadInt64(&r.progress.Failed),
			time.Since(r.progress.StartTime).Round(time.Second))
	}()

	return r.results
}

func (r *Runner) generateCombinations() [][3]string {
	var combos [][3]string

	switch r.opts.Mode {
	case "pair":
		// One-to-one pairing
		minLen := len(r.opts.Usernames)
		if len(r.opts.Passwords) < minLen {
			minLen = len(r.opts.Passwords)
		}
		for i := 0; i < minLen; i++ {
			for _, target := range r.opts.Targets {
				combos = append(combos, [3]string{target, r.opts.Usernames[i], r.opts.Passwords[i]})
			}
		}

	case "cycle-user":
		// Fixed password, cycle through users
		for _, target := range r.opts.Targets {
			for _, password := range r.opts.Passwords {
				for _, username := range r.opts.Usernames {
					combos = append(combos, [3]string{target, username, password})
				}
			}
		}

	case "cycle-pass":
		// Fixed user, cycle through passwords
		for _, target := range r.opts.Targets {
			for _, username := range r.opts.Usernames {
				for _, password := range r.opts.Passwords {
					combos = append(combos, [3]string{target, username, password})
				}
			}
		}

	default: // cartesian
		// Cartesian product: all combinations
		for _, target := range r.opts.Targets {
			for _, username := range r.opts.Usernames {
				for _, password := range r.opts.Passwords {
					combos = append(combos, [3]string{target, username, password})
				}
			}
		}
	}

	return combos
}

// LoadDictionary loads a dictionary from file
func LoadDictionary(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// DefaultUsernames returns common default usernames
func DefaultUsernames() []string {
	return []string{
		"admin", "root", "user", "test", "guest", "oracle", "postgres",
		"mysql", "administrator", "web", "www", "ftp", "ldap", "sa",
	}
}

// DefaultPasswords returns common default passwords
func DefaultPasswords() []string {
	return []string{
		"admin", "root", "123456", "password", "12345678", "1234",
		"12345", "passwd", "123", "test", "guest", "oracle", "mysql",
		"postgres", "web", "www", "ftp", "1234567890", "admin123",
		"123456789", "123123", "000000", "111111", "qwerty", "password123",
	}
}

// Progress returns current progress
func (r *Runner) Progress() Progress {
	return *r.progress
}

// FormatResult formats a result for output
func FormatResult(r Result) string {
	if r.Success {
		return fmt.Sprintf("[SUCCESS] %s://%s - %s:%s (RTT: %v)",
			r.Protocol, r.Target, r.Username, r.Password, r.RTT)
	}
	if r.Error != nil {
		return fmt.Sprintf("[FAILED] %s://%s - %s:%s - %v",
			r.Protocol, r.Target, r.Username, r.Password, r.Error)
	}
	return fmt.Sprintf("[FAILED] %s://%s - %s:%s",
		r.Protocol, r.Target, r.Username, r.Password)
}
