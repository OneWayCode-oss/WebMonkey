package network

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/username/webmonkey/internal/logging"
)

var (
	// Regex patterns for parsing "arp -a" outputs
	ipMacRegex = regexp.MustCompile(`(?:(?:\d{1,3}\.){3}\d{1,3}).*?(?:[0-9a-fA-F]{2}[:-]){5}[0-9a-fA-F]{2}`)
	macRegex   = regexp.MustCompile(`(?:[0-9a-fA-F]{2}[:-]){5}[0-9a-fA-F]{2}`)
	ipRegex    = regexp.MustCompile(`(?:\d{1,3}\.){3}\d{1,3}`)
)

// GetARPCache queries the system ARP cache and returns a map of IP -> MAC.
func GetARPCache() map[string]string {
	cache := make(map[string]string)

	// Try Linux /proc/net/arp first
	if runtime.GOOS == "linux" {
		if err := parseProcNetArp(cache); err == nil {
			return cache
		}
	}

	// Fallback to running "arp -a" command
	parseArpCommand(cache)
	return cache
}

func parseProcNetArp(cache map[string]string) error {
	file, err := os.Open("/proc/net/arp")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Skip the header row
	if scanner.Scan() {
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) >= 4 {
				ip := fields[0]
				mac := fields[3]
				// Ignore blank/unresolved MACs (e.g., 00:00:00:00:00:00)
				if mac != "00:00:00:00:00:00" && mac != "00-00-00-00-00-00" {
					cache[ip] = normalizeMAC(mac)
				}
			}
		}
	}
	return scanner.Err()
}

func parseArpCommand(cache map[string]string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "arp", "-a")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		logging.Debug("Failed to run 'arp -a': %v", err)
		return
	}

	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		match := ipMacRegex.FindString(line)
		if match != "" {
			ip := ipRegex.FindString(match)
			mac := macRegex.FindString(match)
			if ip != "" && mac != "" {
				cache[ip] = normalizeMAC(mac)
			}
			continue
		}

		// Fallback simple line scan for simpler outputs
		fields := strings.Fields(line)
		var currentIP, currentMAC string
		for _, field := range fields {
			field = strings.Trim(field, "()")
			if ipRegex.MatchString(field) {
				currentIP = field
			} else if macRegex.MatchString(field) {
				currentMAC = field
			}
		}
		if currentIP != "" && currentMAC != "" {
			cache[currentIP] = normalizeMAC(currentMAC)
		}
	}
}

func normalizeMAC(mac string) string {
	mac = strings.ReplaceAll(mac, "-", ":")
	return strings.ToLower(mac)
}

// GetVendorFromMAC returns a vendor name based on the MAC prefix (OUI).
func GetVendorFromMAC(mac string) string {
	if mac == "" {
		return "Unknown"
	}
	mac = normalizeMAC(mac)
	prefix := mac
	if len(mac) >= 8 {
		prefix = mac[:8]
	}

	// Simple OUI lookup table for local/common devices
	ouiDatabase := map[string]string{
		"00:00:0c": "Cisco Systems",
		"00:05:cd": "Askey Computer Corp.",
		"00:0c:29": "VMware",
		"00:11:22": "CUI Inc.",
		"00:15:5d": "Microsoft",
		"00:1a:11": "Google",
		"00:1c:42": "Parallels",
		"00:21:ccc": "Intel",
		"00:50:56": "VMware",
		"04:18:d6": "Ubiquiti Networks",
		"08:00:27": "Oracle (VirtualBox)",
		"10:dd:b1": "Apple",
		"2c:f0:ee": "Apple",
		"3c:5a:37": "Xiaomi",
		"3c:a6:2f": "Apple",
		"48:d7:05": "Apple",
		"50:ec:50": "Ubiquiti Networks",
		"70:8b:cd": "ASUSTek Computer",
		"74:ac:5f": "Intel",
		"80:ea:96": "Huawei",
		"a4:77:33": "Google",
		"b4:b5:b3": "Apple",
		"b8:27:eb": "Raspberry Pi Foundation",
		"c0:56:27": "Huawei",
		"dc:a6:32": "Raspberry Pi Foundation",
		"e4:5f:01": "Raspberry Pi Foundation",
		"fc:ec:da": "Ubiquiti Networks",
	}

	if vendor, found := ouiDatabase[prefix]; found {
		return vendor
	}
	return "Generic Device"
}
