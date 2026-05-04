package scanner

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestConnectScanner(t *testing.T) {
	// Create a test TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	port := addr.Port

	// Start a simple server
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Write([]byte("TEST_BANNER\n"))
			conn.Close()
		}
	}()

	// Give server time to start
	time.Sleep(10 * time.Millisecond)

	t.Run("scan open port", func(t *testing.T) {
		scanner := NewConnectScanner(1, 2, 0, false)
		targets := []Target{
			{Host: "127.0.0.1", IP: net.ParseIP("127.0.0.1"), Port: port},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		results := []ScanResult{}
		for r := range scanner.Scan(ctx, targets) {
			results = append(results, r)
		}

		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}

		if !results[0].Open {
			t.Errorf("Expected port to be open, got closed")
		}

		if results[0].Target.Port != port {
			t.Errorf("Expected port %d, got %d", port, results[0].Target.Port)
		}
	})

	t.Run("scan closed port", func(t *testing.T) {
		scanner := NewConnectScanner(1, 1, 0, false)
		targets := []Target{
			{Host: "127.0.0.1", IP: net.ParseIP("127.0.0.1"), Port: port + 1000},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		results := []ScanResult{}
		for r := range scanner.Scan(ctx, targets) {
			results = append(results, r)
		}

		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}

		if results[0].Open {
			t.Errorf("Expected port to be closed, got open")
		}
	})

	t.Run("banner grab", func(t *testing.T) {
		scanner := NewConnectScanner(1, 2, 0, true)
		targets := []Target{
			{Host: "127.0.0.1", IP: net.ParseIP("127.0.0.1"), Port: port},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		results := []ScanResult{}
		for r := range scanner.Scan(ctx, targets) {
			results = append(results, r)
		}

		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}

		if !results[0].Open {
			t.Errorf("Expected port to be open")
		}

		if results[0].Banner == "" {
			t.Errorf("Expected banner to be captured")
		}
	})

	t.Run("multiple targets", func(t *testing.T) {
		scanner := NewConnectScanner(10, 2, 0, false)
		targets := []Target{
			{Host: "127.0.0.1", IP: net.ParseIP("127.0.0.1"), Port: port},
			{Host: "127.0.0.1", IP: net.ParseIP("127.0.0.1"), Port: port + 1000},
			{Host: "127.0.0.1", IP: net.ParseIP("127.0.0.1"), Port: port + 1001},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		results := []ScanResult{}
		for r := range scanner.Scan(ctx, targets) {
			results = append(results, r)
		}

		if len(results) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(results))
		}

		openCount := 0
		for _, r := range results {
			if r.Open {
				openCount++
			}
		}

		if openCount != 1 {
			t.Errorf("Expected 1 open port, got %d", openCount)
		}
	})
}

func TestProtoName(t *testing.T) {
	tests := []struct {
		port int
		want string
	}{
		{22, "ssh"},
		{80, "http"},
		{443, "https"},
		{3306, "mysql"},
		{5432, "postgresql"},
		{6379, "redis"},
		{99999, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := ProtoName(tt.port)
			if got != tt.want {
				t.Errorf("ProtoName(%d) = %q, want %q", tt.port, got, tt.want)
			}
		})
	}
}

func BenchmarkConnectScanner(b *testing.B) {
	scanner := NewConnectScanner(100, 1, 10000, false)
	targets := []Target{
		{Host: "127.0.0.1", IP: net.ParseIP("127.0.0.1"), Port: 1},
	}

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		for range scanner.Scan(ctx, targets) {
		}
		cancel()
	}
}
