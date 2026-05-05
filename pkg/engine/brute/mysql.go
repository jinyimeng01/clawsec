package brute

import (
	"context"
	"net"
	"time"
)

// MySQLProtocol implements MySQL brute force (simplified handshake)
type MySQLProtocol struct {
	Timeout time.Duration
}

func (p *MySQLProtocol) Name() string {
	return "mysql"
}

func (p *MySQLProtocol) Try(ctx context.Context, target, username, password string) (Result, error) {
	if p.Timeout <= 0 {
		p.Timeout = 5 * time.Second
	}

	host, port, err := net.SplitHostPort(target)
	if err != nil {
		host = target
		port = "3306"
	}
	addr := net.JoinHostPort(host, port)

	result := Result{
		Target:   target,
		Protocol: "mysql",
		Username: username,
		Password: password,
	}

	// For a proper MySQL implementation, we'd need to implement the full MySQL protocol
	// This is a simplified version that just checks connectivity
	conn, err := net.DialTimeout("tcp", addr, p.Timeout)
	if err != nil {
		return result, nil
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(p.Timeout))
	defer conn.SetDeadline(time.Time{})

	// Read server greeting
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return result, nil
	}

	// The first byte should be the protocol version (0x0a for MySQL 4.1+)
	if n > 0 && buf[0] == 0x0a {
		result.Banner = "MySQL Server"
	}

	// Note: Full MySQL authentication requires implementing the handshake protocol
	// including scramble parsing and SHA1 hashing. For production use, consider
	// using a proper MySQL driver.

	return result, nil
}
