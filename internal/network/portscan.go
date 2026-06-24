package network

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/username/webmonkey/internal/domain"
)

// LookupPortService returns a standard friendly name for common ports.
func LookupPortService(port int) string {
	services := map[int]string{
		21:   "FTP",
		22:   "SSH",
		23:   "Telnet",
		25:   "SMTP",
		53:   "DNS",
		80:   "HTTP",
		110:  "POP3",
		115:  "SFTP",
		123:  "NTP",
		135:  "MSRPC",
		139:  "NetBIOS",
		143:  "IMAP",
		443:  "HTTPS",
		445:  "SMB",
		993:  "IMAPS",
		995:  "POP3S",
		1433: "MS-SQL",
		3306: "MySQL",
		3389: "RDP",
		5432: "PostgreSQL",
		5900: "VNC",
		8080: "HTTP-Alt",
		8443: "HTTPS-Alt",
	}

	if name, found := services[port]; found {
		return name
	}
	return "unknown"
}

// ScanPorts scans specific TCP ports on an IP with controlled concurrency.
func ScanPorts(ctx context.Context, ip string, ports []int, timeout time.Duration, concurrency int) []domain.Port {
	if len(ports) == 0 {
		return nil
	}
	if concurrency <= 0 {
		concurrency = 10
	}

	var openPorts []domain.Port
	var mu sync.Mutex

	portsChan := make(chan int, len(ports))
	for _, p := range ports {
		portsChan <- p
	}
	close(portsChan)

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case port, ok := <-portsChan:
					if !ok {
						return
					}

					addr := fmt.Sprintf("%s:%d", ip, port)
					dialer := net.Dialer{Timeout: timeout}
					conn, err := dialer.DialContext(ctx, "tcp", addr)
					if err == nil {
						conn.Close()

						mu.Lock()
						openPorts = append(openPorts, domain.Port{
							Number:      port,
							Protocol:    "tcp",
							State:       "open",
							ServiceName: LookupPortService(port),
						})
						mu.Unlock()
					}
				}
			}
		}()
	}

	wg.Wait()
	return openPorts
}
