package service

import (
	"context"
	"fmt"
	"time"

	"github.com/username/webmonkey/internal/config"
	"github.com/username/webmonkey/internal/domain"
	"github.com/username/webmonkey/internal/logging"
	"github.com/username/webmonkey/internal/scanner"
	"github.com/username/webmonkey/internal/store"
)

type Manager struct {
	cfg     *config.Config
	store   *store.Store
	scanner *scanner.Engine
}

func NewManager(cfg *config.Config, s *store.Store, engine *scanner.Engine) *Manager {
	return &Manager{
		cfg:     cfg,
		store:   s,
		scanner: engine,
	}
}

// PerformScan executes a network scan, updates the database, and generates events for changes.
func (m *Manager) PerformScan(ctx context.Context, progressChan chan<- scanner.ProgressUpdate) (*domain.Scan, error) {
	startedAt := time.Now()

	// 1. Run the scanning engine
	discovered, err := m.scanner.Scan(ctx, progressChan)
	if err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	finishedAt := time.Now()
	duration := finishedAt.Sub(startedAt)

	// 2. Prepare scan stats
	scan := &domain.Scan{
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		CIDR:       m.cfg.ScanCIDR,
		TotalHosts: 0, // Will be set by engine or calculated
		AliveHosts: len(discovered),
		Duration:   duration,
	}

	// 3. Save scan metadata first to get an ID
	scanID, err := m.store.SaveScan(scan)
	if err != nil {
		return nil, err
	}

	// 4. Process discovered devices and detect changes
	newCount := 0
	lostCount := 0

	// Track which IPs were seen in this scan
	seenIPs := make(map[string]bool)
	for _, d := range discovered {
		seenIPs[d.IP] = true

		// Check if device already exists
		existing, err := m.store.GetDeviceByIP(d.IP)
		if err != nil {
			logging.Error("Error fetching existing device %s: %v", d.IP, err)
			continue
		}

		d.LastScanID = scanID

		if existing == nil {
			// NEW DEVICE
			newCount++
			d.FirstSeen = startedAt
			d.Status = "online"
			_, err = m.store.SaveDevice(&d)
			if err == nil {
				m.store.SaveEvent(&domain.Event{
					Type:      "device_discovered",
					Timestamp: startedAt,
					DeviceIP:  d.IP,
					Message:   fmt.Sprintf("New device discovered: %s (%s)", d.IP, d.Hostname),
				})
			}
		} else {
			// EXISTING DEVICE - Update
			if existing.Status == "offline" {
				m.store.SaveEvent(&domain.Event{
					Type:      "device_reappeared",
					Timestamp: startedAt,
					DeviceIP:  d.IP,
					Message:   fmt.Sprintf("Device came back online: %s", d.IP),
				})
			}
			d.ID = existing.ID
			d.FirstSeen = existing.FirstSeen
			d.Tags = existing.Tags
			d.Notes = existing.Notes
			d.Status = "online"
			_, err = m.store.SaveDevice(&d)
		}

		// Save ports if any were found
		if len(d.OpenPorts) > 0 {
			m.store.ClearPortsForDevice(d.ID)
			for _, p := range d.OpenPorts {
				m.store.SavePort(d.ID, p)
			}
		}
	}

	// 5. Detect devices that went OFFLINE
	// (Devices that were online in DB but NOT seen in this scan)
	allDevices, err := m.store.GetDevices()
	if err == nil {
		for _, ed := range allDevices {
			if !seenIPs[ed.IP] && ed.Status == "online" {
				lostCount++
				m.store.UpdateDeviceStatus(ed.IP, "offline", scanID)
				m.store.SaveEvent(&domain.Event{
					Type:      "device_lost",
					Timestamp: finishedAt,
					DeviceIP:  ed.IP,
					Message:   fmt.Sprintf("Device went offline: %s", ed.IP),
				})
			}
		}
	}

	// 6. Update scan stats with counts
	scan.NewHosts = newCount
	scan.LostHosts = lostCount
	// Optionally update scan record if needed, but we already have totals
	// Re-saving to update new/lost counts
	m.store.SaveScan(scan)

	return scan, nil
}

func (m *Manager) GetDevices() ([]domain.Device, error) {
	return m.store.GetDevices()
}

func (m *Manager) GetEvents() ([]domain.Event, error) {
	return m.store.GetEvents()
}

func (m *Manager) GetScans() ([]domain.Scan, error) {
	return m.store.GetScans()
}

func (m *Manager) UpdateDeviceNotes(id int64, notes string, tags []string) error {
	return m.store.UpdateDeviceNotesAndTags(id, notes, tags)
}
