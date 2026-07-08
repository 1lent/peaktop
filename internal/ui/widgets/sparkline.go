package widgets

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

var sparkGradient = []lipgloss.Color{
	"39",
	"33",
	"39",
	"45",
	"51",
	"87",
	"226",
	"220",
	"208",
	"196",
}

func RenderSparkline(data []float64, width int, height int) string {
	if len(data) == 0 {
		return ""
	}

	if width <= 0 {
		width = len(data)
	}
	if height <= 0 {
		height = 3
	}

	min, max := sparkMinMax(data)
	rangeSize := max - min
	if rangeSize <= 0 {
		rangeSize = 1
	}

	columns := make([]string, width)
	step := 1
	if len(data) > width {
		step = len(data) / width
		if step < 1 {
			step = 1
		}
	}

	for col := 0; col < width; col++ {
		srcIdx := col * step
		if srcIdx >= len(data) {
			srcIdx = len(data) - 1
		}

		normalized := (data[srcIdx] - min) / rangeSize
		if normalized < 0 {
			normalized = 0
		}
		if normalized > 1 {
			normalized = 1
		}

		charIdx := int(normalized * float64(len(sparkChars)-1))
		if charIdx < 0 {
			charIdx = 0
		}
		if charIdx >= len(sparkChars) {
			charIdx = len(sparkChars) - 1
		}

		colorIdx := int(normalized * float64(len(sparkGradient)-1))
		if colorIdx < 0 {
			colorIdx = 0
		}
		if colorIdx >= len(sparkGradient) {
			colorIdx = len(sparkGradient) - 1
		}

		style := lipgloss.NewStyle().Foreground(sparkGradient[colorIdx])
		columns[col] = style.Render(string(sparkChars[charIdx]))
	}

	return strings.Join(columns, "")
}

func sparkMinMax(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 1
	}
	min := values[0]
	max := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}
