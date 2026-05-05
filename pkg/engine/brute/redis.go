package brute

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// RedisProtocol implements Redis brute force
type RedisProtocol struct {
	Timeout time.Duration
}

func (p *RedisProtocol) Name() string {
	return "redis"
}

func (p *RedisProtocol) Try(ctx context.Context, target, username, password string) (Result, error) {
	if p.Timeout <= 0 {
		p.Timeout = 5 * time.Second
	}

	host, port, err := net.SplitHostPort(target)
	if err != nil {
		host = target
		port = "6379"
	}
	addr := net.JoinHostPort(host, port)

	result := Result{
		Target:   target,
		Protocol: "redis",
		Username: username,
		Password: password,
	}

	conn, err := net.DialTimeout("tcp", addr, p.Timeout)
	if err != nil {
		return result, nil
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(p.Timeout))
	defer conn.SetDeadline(time.Time{})

	reader := bufio.NewReader(conn)

	// Try AUTH command
	// Redis 6+ supports ACL with username
	if username != "" && username != "default" {
		fmt.Fprintf(conn, "AUTH %s %s\r\n", username, password)
	} else {
		fmt.Fprintf(conn, "AUTH %s\r\n", password)
	}

	resp, err := reader.ReadString('\n')
	if err != nil {
		return result, nil
	}

	// +OK = success
	if strings.HasPrefix(resp, "+OK") {
		result.Success = true
		// Get info
		fmt.Fprint(conn, "INFO server\r\n")
		info, _ := reader.ReadString('\n')
		result.Banner = strings.TrimSpace(info)
	}

	return result, nil
}
