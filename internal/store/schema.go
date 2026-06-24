package store

const schema = `
CREATE TABLE IF NOT EXISTS scans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    started_at DATETIME NOT NULL,
    finished_at DATETIME NOT NULL,
    cidr TEXT NOT NULL,
    total_hosts INTEGER NOT NULL,
    alive_hosts INTEGER NOT NULL,
    new_hosts INTEGER NOT NULL,
    lost_hosts INTEGER NOT NULL,
    duration INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS devices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip TEXT UNIQUE NOT NULL,
    mac TEXT NOT NULL DEFAULT '',
    hostname TEXT NOT NULL DEFAULT '',
    vendor TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'offline',
    first_seen DATETIME NOT NULL,
    last_seen DATETIME NOT NULL,
    tags TEXT NOT NULL DEFAULT '', -- Comma-separated tags
    notes TEXT NOT NULL DEFAULT '',
    last_scan_id INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS ports (
    device_id INTEGER NOT NULL,
    number INTEGER NOT NULL,
    protocol TEXT NOT NULL,
    state TEXT NOT NULL,
    service_name TEXT NOT NULL,
    PRIMARY KEY (device_id, number, protocol),
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    device_ip TEXT NOT NULL,
    message TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_devices_ip ON devices(ip);
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_ports_device_id ON ports(device_id);
`
