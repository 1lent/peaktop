package widgets

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/1lent/peaktop/internal/types"
)

var (
	alertBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")).
			Background(lipgloss.Color("248")).
			Padding(0, 1)
	alertBadgeWarn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("235")).
			Background(lipgloss.Color("220")).
			Padding(0, 1)
	alertBadgeCritOn = lipgloss.NewStyle().
				Foreground(lipgloss.Color("235")).
				Background(lipgloss.Color("196")).
				Bold(true).
				Padding(0, 1)
	alertBadgeCritOff = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Background(lipgloss.Color("52")).
				Bold(true).
				Padding(0, 1)
	alertBadgeEmpty = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
	alertBlinkPeriod = 800 * time.Millisecond
)

func RenderAlertBadge(alerts []types.AlertEvent) string {
	if len(alerts) == 0 {
		return alertBadgeEmpty.Render("--")
	}

	total := len(alerts)
	critCount := 0
	warnCount := 0

	for _, a := range alerts {
		switch a.Level {
		case "critical":
			critCount++
		case "warning":
			warnCount++
		}
	}

	if critCount > 0 {
		return renderCriticalBadge(total, critCount)
	}

	if warnCount > 0 {
		return alertBadgeWarn.Render(fmt.Sprintf("%d warn", warnCount))
	}

	return alertBadge.Render(fmt.Sprintf("%d info", total))
}

func renderCriticalBadge(total, critCount int) string {
	on := (time.Now().UnixMilli() / alertBlinkPeriod.Milliseconds() % 2) == 0

	if on {
		return alertBadgeCritOn.Render(fmt.Sprintf("%d!!", critCount))
	}
	return alertBadgeCritOff.Render(fmt.Sprintf("%d!!", critCount))
}
