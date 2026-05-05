package brute

import (
	"context"
	"fmt"
	"net"
	"time"
)

// FTPProtocol implements FTP brute force
type FTPProtocol struct {
	Timeout time.Duration
}

func (p *FTPProtocol) Name() string {
	return "ftp"
}

func (p *FTPProtocol) Try(ctx context.Context, target, username, password string) (Result, error) {
	if p.Timeout <= 0 {
		p.Timeout = 5 * time.Second
	}

	host, port, err := net.SplitHostPort(target)
	if err != nil {
		host = target
		port = "21"
	}
	addr := net.JoinHostPort(host, port)

	result := Result{
		Target:   target,
		Protocol: "ftp",
		Username: username,
		Password: password,
	}

	// Simple FTP handshake
	conn, err := net.DialTimeout("tcp", addr, p.Timeout)
	if err != nil {
		return result, nil
	}
	defer conn.Close()

	// Set deadlines
	conn.SetDeadline(time.Now().Add(p.Timeout))
	defer conn.SetDeadline(time.Time{})

	// Read banner
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	banner := string(buf[:n])
	result.Banner = banner

	// Send USER
	fmt.Fprintf(conn, "USER %s\r\n", username)
	n, _ = conn.Read(buf)
	if n == 0 {
		return result, nil
	}
	userResp := string(buf[:n])

	// Check if user exists (331 = need password, 530 = no login)
	if len(userResp) >= 3 && userResp[:3] == "530" {
		return result, nil
	}

	// Send PASS
	fmt.Fprintf(conn, "PASS %s\r\n", password)
	n, _ = conn.Read(buf)
	if n == 0 {
		return result, nil
	}
	passResp := string(buf[:n])

	// 230 = login successful
	if len(passResp) >= 3 && passResp[:3] == "230" {
		result.Success = true
	}

	return result, nil
}
