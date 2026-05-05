package scanner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/clawsec/clawsec/internal/logger"
	"github.com/clawsec/clawsec/internal/output"
)

// ConnectScanner performs TCP connect scanning
type ConnectScanner struct {
	threads    int
	timeout    time.Duration
	rateLimit  int
	bannerGrab bool

	results chan ScanResult
	done    chan struct{}
}

// ScanResult represents a single scan result
type ScanResult struct {
	Target    Target
	Open      bool
	Banner    string
	Err       error
	RTT       time.Duration
	Timestamp time.Time
}

// NewConnectScanner creates a new connect scanner
func NewConnectScanner(threads, timeout, rateLimit int, bannerGrab bool) *ConnectScanner {
	if threads <= 0 {
		threads = 50
	}
	if timeout <= 0 {
		timeout = 3
	}
	return &ConnectScanner{
		threads:    threads,
		timeout:    time.Duration(timeout) * time.Second,
		rateLimit:  rateLimit,
		bannerGrab: bannerGrab,
		results:    make(chan ScanResult, threads*2),
		done:       make(chan struct{}),
	}
}

// Scan performs the scan and returns results via channel
func (s *ConnectScanner) Scan(ctx context.Context, targets []Target) <-chan ScanResult {
	go func() {
		defer close(s.results)

		// Create rate limiter
		var limiter <-chan time.Time
		if s.rateLimit > 0 {
			limiter = time.Tick(time.Second / time.Duration(s.rateLimit))
		}

		// Worker pool
		var wg sync.WaitGroup
		targetChan := make(chan Target, s.threads)

		// Start workers
		for i := 0; i < s.threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for target := range targetChan {
					select {
					case <-ctx.Done():
						return
					default:
					}

					if limiter != nil {
						<-limiter
					}

					result := s.scanTarget(ctx, target)
					select {
					case s.results <- result:
					case <-ctx.Done():
						return
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

	return s.results
}

func (s *ConnectScanner) scanTarget(ctx context.Context, target Target) ScanResult {
	start := time.Now()
	addr := fmt.Sprintf("%s:%d", target.IP.String(), target.Port)

	result := ScanResult{
		Target:    target,
		Timestamp: start,
	}

	// Create dialer with timeout
	dialer := &net.Dialer{
		Timeout: s.timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		result.RTT = time.Since(start)
		// Connection refused means port is closed, other errors might mean filtered
		if opErr, ok := err.(*net.OpError); ok {
			if sysErr, ok := opErr.Err.(*net.AddrError); ok && sysErr != nil {
				result.Err = err
			}
		}
		return result
	}
	defer conn.Close()

	result.Open = true
	result.RTT = time.Since(start)

	// Banner grabbing
	if s.bannerGrab {
		banner := s.grabBanner(conn)
		result.Banner = banner
	}

	logger.Debugf("Open: %s (RTT: %v)", addr, result.RTT)
	return result
}

func (s *ConnectScanner) grabBanner(conn net.Conn) string {
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		// Some services need a probe first
		_, _ = conn.Write([]byte("\r\n"))
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, err = conn.Read(buf)
		if err != nil || n == 0 {
			return ""
		}
	}
	return string(buf[:n])
}

// ResultsToOutput converts scan results to output format
func ResultsToOutput(results []ScanResult) []output.Result {
	var out []output.Result
	for _, r := range results {
		if !r.Open {
			continue
		}
		level := "info"
		if r.Banner != "" {
			level = "low"
		}
		msg := fmt.Sprintf("Port open: %d/%s", r.Target.Port, ProtoName(r.Target.Port))
		if r.Banner != "" {
			msg += fmt.Sprintf(" - Banner: %s", truncate(r.Banner, 100))
		}
		out = append(out, output.Result{
			Timestamp: r.Timestamp,
			Type:      "port",
			Level:     level,
			Host:      r.Target.Host,
			Port:      r.Target.Port,
			Message:   msg,
			Metadata: map[string]interface{}{
				"rtt_ms": r.RTT.Milliseconds(),
				"banner": r.Banner,
				"ip":     r.Target.IP.String(),
			},
		})
	}
	return out
}

func ProtoName(port int) string {
	protos := map[int]string{
		21: "ftp", 22: "ssh", 23: "telnet", 25: "smtp", 53: "dns",
		80: "http", 110: "pop3", 111: "rpcbind", 135: "msrpc", 139: "netbios-ssn",
		143: "imap", 443: "https", 445: "microsoft-ds", 993: "imaps", 995: "pop3s",
		1723: "pptp", 3306: "mysql", 3389: "rdp", 5432: "postgresql",
		5900: "vnc", 5985: "winrm", 6379: "redis", 8080: "http-proxy",
		8443: "https-alt", 9200: "elasticsearch", 27017: "mongodb",
	}
	if name, ok := protos[port]; ok {
		return name
	}
	return "unknown"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
