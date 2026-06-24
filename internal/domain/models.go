package domain

import "time"

// Device represents a network host discovered during scanning.
type Device struct {
	ID         int64     `json:"id"`
	IP         string    `json:"ip"`
	MAC        string    `json:"mac"`
	Hostname   string    `json:"hostname"`
	Vendor     string    `json:"vendor"`
	Status     string    `json:"status"` // "online", "offline"
	FirstSeen  time.Time `json:"first_seen"`
	LastSeen   time.Time `json:"last_seen"`
	Tags       []string  `json:"tags"`
	Notes      string    `json:"notes"`
	OpenPorts  []Port    `json:"open_ports"`
	LastScanID int64     `json:"last_scan_id"`
}

// Scan holds metadata and statistics about a network discovery run.
type Scan struct {
	ID         int64         `json:"id"`
	StartedAt  time.Time     `json:"started_at"`
	FinishedAt time.Time     `json:"finished_at"`
	CIDR       string        `json:"cidr"`
	TotalHosts int           `json:"total_hosts"`
	AliveHosts int           `json:"alive_hosts"`
	NewHosts   int           `json:"new_hosts"`
	LostHosts  int           `json:"lost_hosts"`
	Duration   time.Duration `json:"duration"`
}

// Port represents an open TCP port on a device.
type Port struct {
	Number      int    `json:"number"`
	Protocol    string `json:"protocol"` // e.g., "tcp"
	State       string `json:"state"`    // "open"
	ServiceName string `json:"service_name"`
}

// Event represents a significant network change (e.g., host joined/left, port opened).
type Event struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"` // "device_discovered", "device_lost", "device_reappeared", "port_opened"
	Timestamp time.Time `json:"timestamp"`
	DeviceIP  string    `json:"device_ip"`
	Message   string    `json:"message"`
}
