package tui

import (
	"fmt"
)

func (m Model) renderDashboard() string {
	logo := `
      .---.
     /     \
    | () () |
     \  ^  /
      |||||
      |||||
`
	
	stats := "Total Devices: 0\nOnline: 0\nLast Scan: Never"
	
	if m.isScanning {
		return BannerStyle.Render(logo) + "\n\n" +
			TitleStyle.Render("Scanning Network...") + "\n" +
			m.spinner.View() + " " + fmt.Sprintf("Scanned: %d/%d", m.progress.Scanned, m.progress.Total)
	}

	return BannerStyle.Render(logo) + "\n\n" +
		BoxStyle.Render(
			TitleStyle.Render("Network Summary") + "\n\n" +
				stats,
		)
}

func (m Model) renderDevices() string {
	return BoxStyle.Render("Devices Table (Coming Soon...)")
}

func (m Model) renderHistory() string {
	return BoxStyle.Render("Scan History (Coming Soon...)")
}

func (m Model) renderLogs() string {
	return BoxStyle.Render("Application Logs (Coming Soon...)")
}

func (m Model) renderHelp() string {
	return BoxStyle.Render("Keyboard Shortcuts:\n" +
		" [s] - Start Scan\n" +
		" [q] - Quit\n" +
		" [tab] - Next Screen\n" +
		" [1-5] - Navigate")
}
