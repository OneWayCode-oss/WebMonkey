
# WebMonkey

See your network. Know your network.

WebMonkey is a modern, cross-platform terminal-based network scanning and inventory tool written in Go.

## Disclaimer
This project is intended **only** for analysis and monitoring of **your own networks** or networks where you have explicit authorization to perform such actions. Never use this tool for unauthorized network scanning.

## Features
- **Network Scanning**: Safe, rootless host discovery.
- **Inventory**: Tracks devices, MAC addresses, and vendors.
- **Monitoring**: Detects new, offline, and reappearing devices.
- **TUI**: A beautiful terminal dashboard.
- **Persistence**: SQLite storage for devices, scans, and events.
- **Export**: JSON report generation.

## Installation
Ensure you have Go 1.20+ installed.

```bash
git clone https://github.com/username/webmonkey
cd webmonkey
go build -o webmonkey ./cmd/webmonkey
```

## Usage
Copy `config.example.yaml` to `config.yaml` and adjust settings.

```bash
# Run the TUI
./webmonkey
```

## License
MIT
=======
# WebMonkey

