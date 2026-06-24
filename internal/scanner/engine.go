package scanner

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/username/webmonkey/internal/config"
	"github.com/username/webmonkey/internal/domain"
	"github.com/username/webmonkey/internal/logging"
	"github.com/username/webmonkey/internal/network"
)

type ProgressUpdate struct {
	CurrentIP string
	Scanned   int
	Total     int
	Found     int
}

type Engine struct {
	cfg *config.Config
}

func NewEngine(cfg *config.Config) *Engine {
	return &Engine{cfg: cfg}
}

// GenerateIPs expands a CIDR block into a list of individual IP addresses.
func GenerateIPs(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		// Try parsing as a single IP
		singleIP := net.ParseIP(cidr)
		if singleIP != nil {
			return []string{singleIP.String()}, nil
		}
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		ips = append(ips, ip.String())
	}

	// For standard subnets like /24, optionally trim the network (.0) and broadcast (.255) addresses,
	// but to be fully inventory-compliant and find everything, we can keep them or scan them safely.
	return ips, nil
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// Scan runs the complete discovery engine over the CIDR block.
// Emits real-time progress on the progressChan (if provided) and returns the discovered online devices.
func (e *Engine) Scan(ctx context.Context, progressChan chan<- ProgressUpdate) ([]domain.Device, error) {
	targetCIDR := e.cfg.ScanCIDR
	ips, err := GenerateIPs(targetCIDR)
	if err != nil {
		return nil, err
	}

	total := len(ips)
	logging.Info("Starting scan on CIDR %s (total hosts: %d)", targetCIDR, total)

	// Fetch system ARP cache once before scan, to minimize process spawn / proc reads
	arpCache := network.GetARPCache()

	var discovered []domain.Device
	var mu sync.Mutex

	ipsChan := make(chan string, total)
	for _, ip := range ips {
		ipsChan <- ip
	}
	close(ipsChan)

	scannedCount := 0
	foundCount := 0

	var wg sync.WaitGroup
	concurrency := e.cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 32
	}
	if concurrency > total {
		concurrency = total
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case ip, ok := <-ipsChan:
					if !ok {
						return
					}

					// 1. Host Discovery
					isAlive := network.IsHostAlive(ctx, ip, e.cfg.Timeout)

					mu.Lock()
					scannedCount++
					currentScanned := scannedCount
					currentFound := foundCount
					mu.Unlock()

					// Notify progress channel
					if progressChan != nil {
						select {
						case progressChan <- ProgressUpdate{
							CurrentIP: ip,
							Scanned:   currentScanned,
							Total:     total,
							Found:     currentFound,
						}:
						case <-ctx.Done():
							return
						default:
						}
					}

					if !isAlive {
						continue
					}

					// 2. Resolve Hostname
					hostname := network.ResolveHostname(ctx, ip)

					// 3. Resolve MAC and Vendor
					mac := arpCache[ip]
					vendor := network.GetVendorFromMAC(mac)

					// 4. Port Scan (if enabled)
					var openPorts []domain.Port
					if e.cfg.PortScanEnabled {
						logging.Debug("Scanning ports for %s...", ip)
						// Port scan with smaller internal concurrency to prevent OS limits
						openPorts = network.ScanPorts(ctx, ip, e.cfg.PortsList, e.cfg.Timeout, 10)
					}

					dev := domain.Device{
						IP:         ip,
						MAC:        mac,
						Hostname:   hostname,
						Vendor:     vendor,
						Status:     "online",
						FirstSeen:  time.Now(),
						LastSeen:   time.Now(),
						Tags:       []string{},
						OpenPorts:  openPorts,
						LastScanID: 0,
					}

					mu.Lock()
					foundCount++
					discovered = append(discovered, dev)
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	logging.Info("Scan completed. Discovered %d alive hosts out of %d.", len(discovered), total)
	return discovered, nil
}
