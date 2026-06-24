package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/username/webmonkey/internal/config"
	"github.com/username/webmonkey/internal/logging"
	"github.com/username/webmonkey/internal/scanner"
	"github.com/username/webmonkey/internal/service"
	"github.com/username/webmonkey/internal/store"
	"github.com/username/webmonkey/internal/tui"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Logging
	err = logging.Init("webmonkey.log", cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}

	// 3. Initialize Store (Database)
	dbStore, err := store.NewStore(cfg.DBPath)
	if err != nil {
		logging.Error("Failed to initialize database: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer dbStore.Close()

	// 4. Initialize Scanner & Service Manager
	scannerEngine := scanner.NewEngine(cfg)
	mgr := service.NewManager(cfg, dbStore, scannerEngine)

	// 5. Run TUI
	p := tea.NewProgram(tui.NewModel(cfg, mgr), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		logging.Error("Failed to run TUI: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to run TUI: %v\n", err)
		os.Exit(1)
	}
}
