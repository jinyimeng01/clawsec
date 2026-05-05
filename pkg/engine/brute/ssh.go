package brute

import (
	"context"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHProtocol implements SSH brute force
type SSHProtocol struct {
	Timeout time.Duration
}

func (p *SSHProtocol) Name() string {
	return "ssh"
}

func (p *SSHProtocol) Try(ctx context.Context, target, username, password string) (Result, error) {
	if p.Timeout <= 0 {
		p.Timeout = 5 * time.Second
	}

	// Ensure target has port
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		host = target
		port = "22"
	}
	addr := net.JoinHostPort(host, port)

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         p.Timeout,
		BannerCallback:  ssh.BannerDisplayStderr(),
	}

	result := Result{
		Target:   target,
		Protocol: "ssh",
		Username: username,
		Password: password,
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return result, nil // Failed auth, not a real error
	}
	defer client.Close()

	result.Success = true
	result.Banner = string(client.ServerVersion())
	return result, nil
}

// SSHKeyProtocol implements SSH key-based brute force
type SSHKeyProtocol struct {
	Timeout time.Duration
	Keys    []string // List of private key strings
}

func (p *SSHKeyProtocol) Name() string {
	return "ssh-key"
}

func (p *SSHKeyProtocol) Try(ctx context.Context, target, username, password string) (Result, error) {
	if p.Timeout <= 0 {
		p.Timeout = 5 * time.Second
	}

	host, port, err := net.SplitHostPort(target)
	if err != nil {
		host = target
		port = "22"
	}
	addr := net.JoinHostPort(host, port)

	result := Result{
		Target:   target,
		Protocol: "ssh-key",
		Username: username,
		Password: password,
	}

	// Try each key
	for _, keyStr := range p.Keys {
		signer, err := ssh.ParsePrivateKey([]byte(keyStr))
		if err != nil {
			continue
		}

		config := &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         p.Timeout,
		}

		client, err := ssh.Dial("tcp", addr, config)
		if err != nil {
			continue
		}
		client.Close()

		result.Success = true
		return result, nil
	}

	return result, nil
}
