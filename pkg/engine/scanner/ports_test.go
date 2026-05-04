package scanner

import (
	"reflect"
	"testing"
)

func TestParsePorts(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		want    []int
		wantErr bool
	}{
		{
			name: "top100",
			spec: "top100",
			want: Top100,
		},
		{
			name: "single port",
			spec: "80",
			want: []int{80},
		},
		{
			name: "multiple ports",
			spec: "80,443,8080",
			want: []int{80, 443, 8080},
		},
		{
			name: "range",
			spec: "80-85",
			want: []int{80, 81, 82, 83, 84, 85},
		},
		{
			name: "mixed",
			spec: "22,80-82,443,8080",
			want: []int{22, 80, 81, 82, 443, 8080},
		},
		{
			name: "dedup",
			spec: "80,80,443,80",
			want: []int{80, 443},
		},
		{
			name:    "invalid range",
			spec:    "80-70",
			wantErr: true,
		},
		{
			name:    "invalid port",
			spec:    "abc",
			wantErr: true,
		},
		{
			name:    "port too high",
			spec:    "70000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePorts(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePorts(%q) error = %v, wantErr %v", tt.spec, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePorts(%q) = %v, want %v", tt.spec, got, tt.want)
			}
		})
	}
}

func TestParsePorts_All(t *testing.T) {
	ports, err := ParsePorts("all")
	if err != nil {
		t.Fatalf("ParsePorts(\"all\") error = %v", err)
	}
	if len(ports) != 65535 {
		t.Errorf("ParsePorts(\"all\") = %d ports, want 65535", len(ports))
	}
}

func TestDedupPorts(t *testing.T) {
	input := []int{80, 443, 80, 8080, 443, 80}
	want := []int{80, 443, 8080}
	got := dedupPorts(input)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("dedupPorts() = %v, want %v", got, want)
	}
}

func BenchmarkParsePorts(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ParsePorts("22,80,443,3306,5432,6379,8080-8090,9200")
	}
}
