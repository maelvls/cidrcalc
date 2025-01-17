package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
)

func main() {
	// Define flags
	hostname := flag.String("hostname", "", "Hostname to resolve and calculate the CIDR for its IPs")
	debug := flag.Bool("debug", false, "Enable debug output")
	flag.Parse()

	var ips []net.IP

	if *hostname != "" {
		// Resolve the hostname to IPs
		resolvedIPs, err := net.LookupIP(*hostname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving hostname %s: %v\n", *hostname, err)
			return
		}
		ips = append(ips, resolvedIPs...)
		if *debug {
			debugLog(fmt.Sprintf("Resolved IPs for %s: %v", *hostname, resolvedIPs))
		}
	} else {
		// Read IPs from standard input
		scanner := bufio.NewScanner(os.Stdin)
		if *debug {
			debugLog("Enter IPs, one per line. Press Ctrl+D (Unix) or Ctrl+Z (Windows) to end:")
		}
		for scanner.Scan() {
			ip := net.ParseIP(scanner.Text())
			if ip == nil {
				if *debug {
					debugLog(fmt.Sprintf("Invalid IP: %s", scanner.Text()))
				}
				continue
			}
			ips = append(ips, ip)
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			return
		}
	}

	if len(ips) == 0 {
		fmt.Fprintf(os.Stderr, "No valid IPs provided.\n")
		return
	}

	// Sort IPs
	sort.Slice(ips, func(i, j int) bool {
		return compareIPs(ips[i], ips[j]) < 0
	})

	// Calculate the largest CIDR block
	minIP := ips[0]
	maxIP := ips[len(ips)-1]

	// Convert minIP and maxIP to uint32 for calculations
	minUint := ipToUint32(minIP)
	maxUint := ipToUint32(maxIP)

	// Calculate the CIDR prefix
	prefixLen := 32
	for prefixLen > 0 {
		mask := uint32((1<<prefixLen)-1) << (32 - prefixLen)
		if minUint&mask == maxUint&mask {
			break
		}
		prefixLen--
	}

	// Print the largest CIDR block to stdout
	cidr := fmt.Sprintf("%s/%d", minIP.Mask(net.CIDRMask(prefixLen, 32)), prefixLen)
	fmt.Println(cidr)
}

// debugLog prints debug messages to stderr with a yellow "debug:" prefix
func debugLog(message string) {
	// Yellow ANSI color code
	yellow := "\033[33m"
	reset := "\033[0m"
	fmt.Fprintf(os.Stderr, "%sdebug:%s %s\n", yellow, reset, message)
}

// ipToUint32 converts an IPv4 address to a uint32.
func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// compareIPs compares two IP addresses. Returns -1, 0, or 1.
func compareIPs(ip1, ip2 net.IP) int {
	ip1 = ip1.To4()
	ip2 = ip2.To4()
	for i := 0; i < 4; i++ {
		if ip1[i] < ip2[i] {
			return -1
		}
		if ip1[i] > ip2[i] {
			return 1
		}
	}
	return 0
}
