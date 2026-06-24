package export

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/username/webmonkey/internal/domain"
)

type ExportData struct {
	Timestamp string          `json:"timestamp"`
	Devices   []domain.Device `json:"devices"`
	Scans     []domain.Scan   `json:"scans"`
	Events    []domain.Event  `json:"events"`
}

// ToJSON exports all network data to a formatted JSON file.
func ToJSON(filePath string, devices []domain.Device, scans []domain.Scan, events []domain.Event) error {
	data := ExportData{
		Timestamp: fmt.Sprintf("%v", os.Getenv("TIME")), // Placeholder or time.Now()
		Devices:   devices,
		Scans:     scans,
		Events:    events,
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
