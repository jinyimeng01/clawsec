package scanner

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// Target represents a scan target
type Target struct {
	Host string
	IP   net.IP
	Port int
}

// ParseTargets parses target specification into individual targets
func ParseTargets(specs []string, ports []int) ([]Target, error) {
	var targets []Target

	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}

		// Check if it's a file
		if strings.HasPrefix(spec, "file:") {
			fileTargets, err := parseTargetFile(strings.TrimPrefix(spec, "file:"), ports)
			if err != nil {
				return nil, err
			}
			targets = append(targets, fileTargets...)
			continue
		}

		// Check if it's already a host:port combination
		if strings.Contains(spec, ":") {
			host, portStr, err := net.SplitHostPort(spec)
			if err == nil {
				port, _ := parsePort(portStr)
				if port > 0 {
					ips, err := resolveHost(host)
					if err != nil {
						return nil, fmt.Errorf("failed to resolve %s: %w", host, err)
					}
					for _, ip := range ips {
						targets = append(targets, Target{Host: host, IP: ip, Port: port})
					}
					continue
				}
			}
		}

		// Check if it's a CIDR
		if strings.Contains(spec, "/") {
			cidrTargets, err := parseCIDR(spec, ports)
			if err != nil {
				return nil, err
			}
			targets = append(targets, cidrTargets...)
			continue
		}

		// Single host/IP
		ips, err := resolveHost(spec)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %s: %w", spec, err)
		}
		for _, ip := range ips {
			for _, port := range ports {
				targets = append(targets, Target{Host: spec, IP: ip, Port: port})
			}
		}
	}

	return targets, nil
}

func parseTargetFile(path string, ports []int) ([]Target, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open target file: %w", err)
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

	return ParseTargets(lines, ports)
}

func parseCIDR(cidr string, ports []int) ([]Target, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %s", cidr)
	}

	var targets []Target
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		// Skip network and broadcast addresses for /24 to /30
		// For /31, both addresses are usable (RFC 3021)
		if ones, bits := ipnet.Mask.Size(); bits == 32 && ones >= 24 && ones <= 30 {
			lastOctet := ip[3]
			if lastOctet == 0 || lastOctet == 255 {
				continue
			}
		}

		for _, port := range ports {
			targets = append(targets, Target{
				Host: ip.String(),
				IP:   dupIP(ip),
				Port: port,
			})
		}
	}

	return targets, nil
}

func resolveHost(host string) ([]net.IP, error) {
	// Check if it's already an IP
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}

	// DNS lookup
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}

	var result []net.IP
	for _, ip := range ips {
		result = append(result, ip)
	}
	return result, nil
}

func parsePort(s string) (int, error) {
	var port int
	_, err := fmt.Sscanf(s, "%d", &port)
	return port, err
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func dupIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}
