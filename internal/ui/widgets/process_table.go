package widgets

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/brodie/peaktop/internal/types"
)

var (
	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("248")).
				Bold(true).
				Underline(true)
	tableSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("236"))
	tableNormalStyle = lipgloss.NewStyle()
	tablePIDStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	tableNameStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
)

const (
	pidColWidth  = 6
	nameColMin   = 16
	cpuColWidth  = 7
	memColWidth  = 7
	colGap       = 2
)

func RenderProcessTable(processes []types.ProcessInfo, width int, selected int, sortBy string) string {
	if len(processes) == 0 {
		return dimText("no processes")
	}

	nameWidth := width - pidColWidth - cpuColWidth - memColWidth - (3 * colGap) - 1
	if nameWidth < nameColMin {
		nameWidth = nameColMin
	}

	header := buildTableHeader(nameWidth, sortBy)
	var rows []string

	for i, proc := range processes {
		style := tableNormalStyle
		if i == selected {
			style = tableSelectedStyle
		}

		row := buildTableRow(proc, nameWidth, style)
		rows = append(rows, row)
	}

	return header + "\n" + strings.Join(rows, "\n")
}

func buildTableHeader(nameWidth int, sortBy string) string {
	pid := tableHeaderStyle.Render(fmt.Sprintf("%-*s", pidColWidth, colHeader("PID", sortBy, "pid")))
	name := tableHeaderStyle.Render(fmt.Sprintf("%-*s", nameWidth, colHeader("NAME", sortBy, "name")))
	cpu := tableHeaderStyle.Render(fmt.Sprintf("%*s", cpuColWidth, colHeader("CPU%", sortBy, "cpu")))
	mem := tableHeaderStyle.Render(fmt.Sprintf("%*s", memColWidth, colHeader("MEM%", sortBy, "mem")))

	gap := strings.Repeat(" ", colGap)
	return pid + gap + name + gap + cpu + gap + mem
}

func colHeader(label, sortBy, key string) string {
	if sortBy == key {
		return label + "*"
	}
	return label
}

func buildTableRow(proc types.ProcessInfo, nameWidth int, style lipgloss.Style) string {
	pid := tablePIDStyle.Render(fmt.Sprintf("%-*d", pidColWidth, proc.PID))
	truncName := truncateString(proc.Name, nameWidth)
	name := tableNameStyle.Render(fmt.Sprintf("%-*s", nameWidth, truncName))
	cpu := fmt.Sprintf("%*.1f", cpuColWidth, proc.CPUPercent)
	mem := fmt.Sprintf("%*.1f", memColWidth, proc.MemPercent)

	gap := strings.Repeat(" ", colGap)
	row := pid + gap + name + gap + cpu + gap + mem

	return style.Render(row)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 2 {
		return s[:maxLen]
	}
	return s[:maxLen-1] + "…"
}

func dimText(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(s)
}
