package main

import (
	"bytes"
	"net"
	"strings"
	"testing"
)

func TestCalculateCIDR(t *testing.T) {
	tests := []struct {
		name    string
		ips     []string
		want    string
		wantErr bool
	}{
		{
			name: "single IP",
			ips:  []string{"192.168.1.1"},
			want: "192.168.1.1/32",
		},
		{
			name: "two adjacent IPs",
			ips:  []string{"192.168.1.1", "192.168.1.2"},
			want: "192.168.1.0/30",
		},
		{
			name: "multiple IPs in /24",
			ips:  []string{"192.168.1.1", "192.168.1.50", "192.168.1.200"},
			want: "192.168.1.0/24",
		},
		{
			name: "IPs across /16",
			ips:  []string{"192.168.1.1", "192.168.50.1"},
			want: "192.168.0.0/18",
		},
		{
			name: "unsorted IPs",
			ips:  []string{"192.168.1.100", "192.168.1.1", "192.168.1.50"},
			want: "192.168.1.0/25",
		},
		{
			name:    "empty IP list",
			ips:     []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ips []net.IP
			for _, ipStr := range tt.ips {
				ips = append(ips, net.ParseIP(ipStr))
			}

			got, err := calculateCIDR(ips)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateCIDR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("calculateCIDR() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseIPsFromReader(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		debug     bool
		wantCount int
		wantIPs   []string
	}{
		{
			name:      "valid IPs",
			input:     "192.168.1.1\n192.168.1.2\n192.168.1.3",
			debug:     false,
			wantCount: 3,
			wantIPs:   []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
		},
		{
			name:      "mixed valid and invalid IPs",
			input:     "192.168.1.1\ninvalid\n192.168.1.2",
			debug:     false,
			wantCount: 2,
			wantIPs:   []string{"192.168.1.1", "192.168.1.2"},
		},
		{
			name:      "empty input",
			input:     "",
			debug:     false,
			wantCount: 0,
			wantIPs:   []string{},
		},
		{
			name:      "only invalid IPs",
			input:     "invalid1\ninvalid2",
			debug:     false,
			wantCount: 0,
			wantIPs:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			ips, err := parseIPsFromReader(reader, tt.debug)
			if err != nil {
				t.Errorf("parseIPsFromReader() error = %v", err)
				return
			}
			if len(ips) != tt.wantCount {
				t.Errorf("parseIPsFromReader() got %d IPs, want %d", len(ips), tt.wantCount)
			}
			for i, want := range tt.wantIPs {
				if i >= len(ips) {
					t.Errorf("parseIPsFromReader() missing IP at index %d", i)
					continue
				}
				if ips[i].String() != want {
					t.Errorf("parseIPsFromReader() IP[%d] = %v, want %v", i, ips[i], want)
				}
			}
		})
	}
}

func TestIPToUint32(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want uint32
	}{
		{
			name: "192.168.1.1",
			ip:   "192.168.1.1",
			want: 3232235777, // 192*2^24 + 168*2^16 + 1*2^8 + 1
		},
		{
			name: "0.0.0.0",
			ip:   "0.0.0.0",
			want: 0,
		},
		{
			name: "255.255.255.255",
			ip:   "255.255.255.255",
			want: 4294967295,
		},
		{
			name: "10.0.0.1",
			ip:   "10.0.0.1",
			want: 167772161,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if got := ipToUint32(ip); got != tt.want {
				t.Errorf("ipToUint32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareIPs(t *testing.T) {
	tests := []struct {
		name string
		ip1  string
		ip2  string
		want int
	}{
		{
			name: "equal IPs",
			ip1:  "192.168.1.1",
			ip2:  "192.168.1.1",
			want: 0,
		},
		{
			name: "ip1 < ip2",
			ip1:  "192.168.1.1",
			ip2:  "192.168.1.2",
			want: -1,
		},
		{
			name: "ip1 > ip2",
			ip1:  "192.168.1.2",
			ip2:  "192.168.1.1",
			want: 1,
		},
		{
			name: "different octets",
			ip1:  "192.168.1.1",
			ip2:  "192.169.1.1",
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip1 := net.ParseIP(tt.ip1)
			ip2 := net.ParseIP(tt.ip2)
			if got := compareIPs(ip1, ip2); got != tt.want {
				t.Errorf("compareIPs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculatePrefixLength(t *testing.T) {
	tests := []struct {
		name    string
		minIP   string
		maxIP   string
		want    int
	}{
		{
			name:  "same IP",
			minIP: "192.168.1.1",
			maxIP: "192.168.1.1",
			want:  32,
		},
		{
			name:  "adjacent IPs",
			minIP: "192.168.1.0",
			maxIP: "192.168.1.1",
			want:  31,
		},
		{
			name:  "full /24",
			minIP: "192.168.1.0",
			maxIP: "192.168.1.255",
			want:  24,
		},
		{
			name:  "full /16",
			minIP: "192.168.0.0",
			maxIP: "192.168.255.255",
			want:  16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minUint := ipToUint32(net.ParseIP(tt.minIP))
			maxUint := ipToUint32(net.ParseIP(tt.maxIP))
			if got := calculatePrefixLength(minUint, maxUint); got != tt.want {
				t.Errorf("calculatePrefixLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkCalculateCIDR(b *testing.B) {
	ips := []net.IP{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("192.168.1.50"),
		net.ParseIP("192.168.1.100"),
		net.ParseIP("192.168.1.200"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calculateCIDR(ips)
	}
}

func BenchmarkParseIPsFromReader(b *testing.B) {
	input := "192.168.1.1\n192.168.1.2\n192.168.1.3\n192.168.1.4\n192.168.1.5"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader([]byte(input))
		_, _ = parseIPsFromReader(reader, false)
	}
}
