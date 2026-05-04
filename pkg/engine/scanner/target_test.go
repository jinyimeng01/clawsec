package scanner

import (
	"testing"
)

func TestParseTargets(t *testing.T) {
	tests := []struct {
		name    string
		specs   []string
		ports   []int
		wantMin int
		wantErr bool
	}{
		{
			name:    "single IP single port",
			specs:   []string{"127.0.0.1"},
			ports:   []int{80},
			wantMin: 1,
		},
		{
			name:    "single IP multiple ports",
			specs:   []string{"127.0.0.1"},
			ports:   []int{80, 443},
			wantMin: 2,
		},
		{
			name:    "CIDR /30",
			specs:   []string{"192.168.1.0/30"},
			ports:   []int{80},
			wantMin: 2, // .1 and .2 (skips .0 network address)
		},
		{
			name:    "host:port",
			specs:   []string{"127.0.0.1:8080"},
			ports:   []int{80},
			wantMin: 1,
		},
		{
			name:    "multiple targets",
			specs:   []string{"127.0.0.1", "127.0.0.2"},
			ports:   []int{22, 80},
			wantMin: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTargets(tt.specs, tt.ports)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTargets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) < tt.wantMin {
				t.Errorf("ParseTargets() = %d targets, want at least %d", len(got), tt.wantMin)
			}
		})
	}
}

func TestResolveHost(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{
			name:    "localhost",
			host:    "localhost",
			wantErr: false,
		},
		{
			name:    "IP address",
			host:    "127.0.0.1",
			wantErr: false,
		},
		{
			name:    "invalid host",
			host:    "not-a-valid-host-123456789.local",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ips, err := resolveHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveHost(%q) error = %v, wantErr %v", tt.host, err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(ips) == 0 {
				t.Errorf("resolveHost(%q) returned no IPs", tt.host)
			}
		})
	}
}

func TestParseCIDR(t *testing.T) {
	tests := []struct {
		name    string
		cidr    string
		ports   []int
		wantMin int
		wantErr bool
	}{
		{
			name:    "/30 network",
			cidr:    "10.0.0.0/30",
			ports:   []int{80},
			wantMin: 2,
			wantErr: false,
		},
		{
			name:    "/31 network",
			cidr:    "10.0.0.0/31",
			ports:   []int{80},
			wantMin: 2,
			wantErr: false,
		},
		{
			name:    "invalid CIDR",
			cidr:    "invalid",
			ports:   []int{80},
			wantMin: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCIDR(tt.cidr, tt.ports)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCIDR(%q) error = %v, wantErr %v", tt.cidr, err, tt.wantErr)
				return
			}
			if len(got) < tt.wantMin {
				t.Errorf("parseCIDR(%q) = %d targets, want at least %d", tt.cidr, len(got), tt.wantMin)
			}
		})
	}
}
