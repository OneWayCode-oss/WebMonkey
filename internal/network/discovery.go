package network

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/username/webmonkey/internal/logging"
)

// CommonPortsForDiscovery are ports commonly open on network appliances, PCs, and servers.
// Probing these helps detect live hosts without root permissions.
var CommonPortsForDiscovery = []int{21, 22, 23, 25, 80, 110, 135, 139, 443, 445, 3389, 8080}

// IsHostAlive checks if an IP is active using TCP dials and an OS ping fallback.
func IsHostAlive(ctx context.Context, ip string, timeout time.Duration) bool {
	// 1. First, try TCP dial on common ports
	for _, port := range CommonPortsForDiscovery {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		address := fmt.Sprintf("%s:%d", ip, port)
		dialer := net.Dialer{Timeout: timeout}
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err == nil {
			conn.Close()
			logging.Debug("Host %s alive: TCP connection to port %d succeeded", ip, port)
			return true
		}

		// Important: If connection is explicitly refused, the host is ALIVE!
		// It replied with a TCP RST, meaning the IP stack is active.
		if err != nil && (strings.Contains(err.Error(), "refused") || strings.Contains(err.Error(), "reset")) {
			logging.Debug("Host %s alive: TCP port %d refused connection (RST received)", ip, port)
			return true
		}
	}

	// 2. Fallback to OS-native ping CLI
	return osPing(ctx, ip, timeout)
}

// osPing runs the system ping CLI which has setuid/raw sockets permission from the OS.
func osPing(ctx context.Context, ip string, timeout time.Duration) bool {
	var cmd *exec.Cmd

	timeoutSecs := int(timeout.Seconds())
	if timeoutSecs < 1 {
		timeoutSecs = 1
	}

	if runtime.GOOS == "windows" {
		// -n 1: 1 packet, -w timeout (ms)
		timeoutMS := int(timeout.Milliseconds())
		cmd = exec.CommandContext(ctx, "ping", "-n", "1", "-w", fmt.Sprintf("%d", timeoutMS), ip)
	} else if runtime.GOOS == "darwin" {
		// -c 1: 1 packet, -t timeout (s)
		cmd = exec.CommandContext(ctx, "ping", "-c", "1", "-t", fmt.Sprintf("%d", timeoutSecs), ip)
	} else {
		// Linux: -c 1: 1 packet, -W timeout (s)
		cmd = exec.CommandContext(ctx, "ping", "-c", "1", "-W", fmt.Sprintf("%d", timeoutSecs), ip)
	}

	err := cmd.Run()
	if err == nil {
		logging.Debug("Host %s alive: OS Ping CLI succeeded", ip)
		return true
	}

	return false
}

// ResolveHostname does a reverse DNS lookup to resolve hostnames.
func ResolveHostname(ctx context.Context, ip string) string {
	resolver := &net.Resolver{}
	names, err := resolver.LookupAddr(ctx, ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	// Return the first name and trim trailing dot
	return strings.TrimSuffix(names[0], ".")
}
