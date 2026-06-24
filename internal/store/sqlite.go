package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/username/webmonkey/internal/domain"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

// NewStore opens an SQLite database and initializes the tables.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database %s: %w", dbPath, err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create tables
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to execute schema: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the underlying SQLite database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// ==========================================
// SCANS REPOSITORY
// ==========================================

func (s *Store) SaveScan(scan *domain.Scan) (int64, error) {
	query := `
		INSERT INTO scans (started_at, finished_at, cidr, total_hosts, alive_hosts, new_hosts, lost_hosts, duration)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := s.db.Exec(query,
		scan.StartedAt,
		scan.FinishedAt,
		scan.CIDR,
		scan.TotalHosts,
		scan.AliveHosts,
		scan.NewHosts,
		scan.LostHosts,
		scan.Duration.Nanoseconds(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to save scan: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get scan last insert id: %w", err)
	}
	scan.ID = id
	return id, nil
}

func (s *Store) GetScans() ([]domain.Scan, error) {
	query := `SELECT id, started_at, finished_at, cidr, total_hosts, alive_hosts, new_hosts, lost_hosts, duration FROM scans ORDER BY id DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list scans: %w", err)
	}
	defer rows.Close()

	var scans []domain.Scan
	for rows.Next() {
		var sc domain.Scan
		var durationNS int64
		err := rows.Scan(
			&sc.ID,
			&sc.StartedAt,
			&sc.FinishedAt,
			&sc.CIDR,
			&sc.TotalHosts,
			&sc.AliveHosts,
			&sc.NewHosts,
			&sc.LostHosts,
			&durationNS,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scan row: %w", err)
		}
		sc.Duration = time.Duration(durationNS)
		scans = append(scans, sc)
	}
	return scans, nil
}

// ==========================================
// DEVICES REPOSITORY
// ==========================================

func (s *Store) SaveDevice(d *domain.Device) (int64, error) {
	tagsStr := strings.Join(d.Tags, ",")
	query := `
		INSERT INTO devices (ip, mac, hostname, vendor, status, first_seen, last_seen, tags, notes, last_scan_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ip) DO UPDATE SET
			mac = CASE WHEN excluded.mac != '' THEN excluded.mac ELSE devices.mac END,
			hostname = CASE WHEN excluded.hostname != '' THEN excluded.hostname ELSE devices.hostname END,
			vendor = CASE WHEN excluded.vendor != '' THEN excluded.vendor ELSE devices.vendor END,
			status = excluded.status,
			last_seen = excluded.last_seen,
			last_scan_id = excluded.last_scan_id
	`
	_, err := s.db.Exec(query,
		d.IP,
		d.MAC,
		d.Hostname,
		d.Vendor,
		d.Status,
		d.FirstSeen,
		d.LastSeen,
		tagsStr,
		d.Notes,
		d.LastScanID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to save device %s: %w", d.IP, err)
	}

	// Fetch current ID if updated/inserted
	var id int64
	err = s.db.QueryRow("SELECT id FROM devices WHERE ip = ?", d.IP).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to query device id for %s: %w", d.IP, err)
	}
	d.ID = id
	return id, nil
}

func (s *Store) GetDevices() ([]domain.Device, error) {
	query := `SELECT id, ip, mac, hostname, vendor, status, first_seen, last_seen, tags, notes, last_scan_id FROM devices ORDER BY ip ASC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}
	defer rows.Close()

	var devices []domain.Device
	for rows.Next() {
		var d domain.Device
		var tagsStr string
		err := rows.Scan(
			&d.ID,
			&d.IP,
			&d.MAC,
			&d.Hostname,
			&d.Vendor,
			&d.Status,
			&d.FirstSeen,
			&d.LastSeen,
			&tagsStr,
			&d.Notes,
			&d.LastScanID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device row: %w", err)
		}
		if tagsStr != "" {
			d.Tags = strings.Split(tagsStr, ",")
		} else {
			d.Tags = []string{}
		}

		// Hydrate open ports
		ports, err := s.GetPortsForDevice(d.ID)
		if err != nil {
			return nil, err
		}
		d.OpenPorts = ports

		devices = append(devices, d)
	}
	return devices, nil
}

func (s *Store) GetDeviceByIP(ip string) (*domain.Device, error) {
	query := `SELECT id, ip, mac, hostname, vendor, status, first_seen, last_seen, tags, notes, last_scan_id FROM devices WHERE ip = ?`
	row := s.db.QueryRow(query, ip)

	var d domain.Device
	var tagsStr string
	err := row.Scan(
		&d.ID,
		&d.IP,
		&d.MAC,
		&d.Hostname,
		&d.Vendor,
		&d.Status,
		&d.FirstSeen,
		&d.LastSeen,
		&tagsStr,
		&d.Notes,
		&d.LastScanID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get device by ip %s: %w", ip, err)
	}

	if tagsStr != "" {
		d.Tags = strings.Split(tagsStr, ",")
	} else {
		d.Tags = []string{}
	}

	ports, err := s.GetPortsForDevice(d.ID)
	if err != nil {
		return nil, err
	}
	d.OpenPorts = ports

	return &d, nil
}

func (s *Store) GetDeviceByID(id int64) (*domain.Device, error) {
	query := `SELECT id, ip, mac, hostname, vendor, status, first_seen, last_seen, tags, notes, last_scan_id FROM devices WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var d domain.Device
	var tagsStr string
	err := row.Scan(
		&d.ID,
		&d.IP,
		&d.MAC,
		&d.Hostname,
		&d.Vendor,
		&d.Status,
		&d.FirstSeen,
		&d.LastSeen,
		&tagsStr,
		&d.Notes,
		&d.LastScanID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get device by id %d: %w", id, err)
	}

	if tagsStr != "" {
		d.Tags = strings.Split(tagsStr, ",")
	} else {
		d.Tags = []string{}
	}

	ports, err := s.GetPortsForDevice(d.ID)
	if err != nil {
		return nil, err
	}
	d.OpenPorts = ports

	return &d, nil
}

func (s *Store) UpdateDeviceStatus(ip string, status string, lastScanID int64) error {
	query := `UPDATE devices SET status = ?, last_scan_id = ?, last_seen = ? WHERE ip = ?`
	_, err := s.db.Exec(query, status, lastScanID, time.Now(), ip)
	if err != nil {
		return fmt.Errorf("failed to update device status for %s: %w", ip, err)
	}
	return nil
}

func (s *Store) UpdateDeviceNotesAndTags(id int64, notes string, tags []string) error {
	tagsStr := strings.Join(tags, ",")
	query := `UPDATE devices SET notes = ?, tags = ? WHERE id = ?`
	_, err := s.db.Exec(query, notes, tagsStr, id)
	if err != nil {
		return fmt.Errorf("failed to update notes/tags for device %d: %w", id, err)
	}
	return nil
}

// ==========================================
// PORTS REPOSITORY
// ==========================================

func (s *Store) SavePort(deviceID int64, p domain.Port) error {
	query := `
		INSERT OR REPLACE INTO ports (device_id, number, protocol, state, service_name)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, deviceID, p.Number, p.Protocol, p.State, p.ServiceName)
	if err != nil {
		return fmt.Errorf("failed to save port %d for device %d: %w", p.Number, deviceID, err)
	}
	return nil
}

func (s *Store) GetPortsForDevice(deviceID int64) ([]domain.Port, error) {
	query := `SELECT number, protocol, state, service_name FROM ports WHERE device_id = ? ORDER BY number ASC`
	rows, err := s.db.Query(query, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list ports for device %d: %w", deviceID, err)
	}
	defer rows.Close()

	var ports []domain.Port
	for rows.Next() {
		var p domain.Port
		err := rows.Scan(&p.Number, &p.Protocol, &p.State, &p.ServiceName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan port row: %w", err)
		}
		ports = append(ports, p)
	}
	return ports, nil
}

func (s *Store) ClearPortsForDevice(deviceID int64) error {
	_, err := s.db.Exec("DELETE FROM ports WHERE device_id = ?", deviceID)
	if err != nil {
		return fmt.Errorf("failed to clear ports for device %d: %w", deviceID, err)
	}
	return nil
}

// ==========================================
// EVENTS REPOSITORY
// ==========================================

func (s *Store) SaveEvent(e *domain.Event) error {
	query := `INSERT INTO events (type, timestamp, device_ip, message) VALUES (?, ?, ?, ?)`
	res, err := s.db.Exec(query, e.Type, e.Timestamp, e.DeviceIP, e.Message)
	if err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		e.ID = id
	}
	return nil
}

func (s *Store) GetEvents() ([]domain.Event, error) {
	query := `SELECT id, type, timestamp, device_ip, message FROM events ORDER BY timestamp DESC, id DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var e domain.Event
		err := rows.Scan(&e.ID, &e.Type, &e.Timestamp, &e.DeviceIP, &e.Message)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}
