package widgets

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	gaugeBarWidth    = 40
	gaugeLabelWidth  = 10
	gaugeLabelPad    = 1
	gaugeBlockFull   = "█"
	gaugeBlockEmpty  = "░"
	gaugeBlockMedium = "▓"
	gaugeBlockLight  = "▒"
	greenThreshold   = 50.0
	yellowThreshold  = 80.0
)

var (
	gaugeGreenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	gaugeYellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	gaugeRedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	gaugeDimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func RenderGauge(label string, percent float64, width int) string {
	if width < gaugeLabelWidth+gaugeLabelPad+10 {
		width = gaugeLabelWidth + gaugeLabelPad + 10
	}

	barWidth := width - gaugeLabelWidth - gaugeLabelPad - 3
	if barWidth < 4 {
		barWidth = 4
	}

	labelText := fmt.Sprintf("%-*s", gaugeLabelWidth, label)
	filled := int(float64(barWidth) * percent / 100.0)
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}

	empty := barWidth - filled

	var barStyle lipgloss.Style
	switch {
	case percent >= yellowThreshold:
		barStyle = gaugeRedStyle
	case percent >= greenThreshold:
		barStyle = gaugeYellowStyle
	default:
		barStyle = gaugeGreenStyle
	}

	filledBar := barStyle.Render(strings.Repeat(gaugeBlockFull, filled))
	emptyBar := gaugeDimStyle.Render(strings.Repeat(gaugeBlockEmpty, empty))

	pctText := fmt.Sprintf("%5.2f%%", percent)

	return fmt.Sprintf("%s %s%s %s", labelText, filledBar, emptyBar, pctText)
}
