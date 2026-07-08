package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Name  string
	Label string

	FG          lipgloss.Color
	BG          lipgloss.Color
	CPU         lipgloss.Color
	GPU         lipgloss.Color
	Memory      lipgloss.Color
	ANE         lipgloss.Color
	Network     lipgloss.Color
	Power       lipgloss.Color
	Thermal     lipgloss.Color
	Battery     lipgloss.Color
	Alert       lipgloss.Color
	Dim         lipgloss.Color
	Selected    lipgloss.Color
	Green       lipgloss.Color
	Yellow      lipgloss.Color
	Red         lipgloss.Color
	Header      lipgloss.Color
	Border      lipgloss.Color
}

var DarkTheme = Theme{
	Name:     "dark",
	Label:    "Dark",
	FG:       "252",
	BG:       "0",
	CPU:      "75",
	GPU:      "213",
	Memory:   "42",
	ANE:      "183",
	Network:  "39",
	Power:    "226",
	Thermal:  "208",
	Battery:  "120",
	Alert:    "196",
	Dim:      "240",
	Selected: "236",
	Green:    "42",
	Yellow:   "220",
	Red:      "196",
	Header:   "248",
	Border:   "238",
}

var LightTheme = Theme{
	Name:     "light",
	Label:    "Light",
	FG:       "235",
	BG:       "255",
	CPU:      "25",
	GPU:      "127",
	Memory:   "28",
	ANE:      "54",
	Network:  "20",
	Power:    "178",
	Thermal:  "166",
	Battery:  "29",
	Alert:    "160",
	Dim:      "248",
	Selected: "252",
	Green:    "28",
	Yellow:   "136",
	Red:      "160",
	Header:   "240",
	Border:   "250",
}

var DraculaTheme = Theme{
	Name:     "dracula",
	Label:    "Dracula",
	FG:       "252",
	BG:       "235",
	CPU:      "84",
	GPU:      "212",
	Memory:   "120",
	ANE:      "141",
	Network:  "117",
	Power:    "228",
	Thermal:  "215",
	Battery:  "84",
	Alert:    "203",
	Dim:      "240",
	Selected: "237",
	Green:    "84",
	Yellow:   "221",
	Red:      "203",
	Header:   "246",
	Border:   "239",
}

func AllThemes() []Theme {
	return []Theme{DarkTheme, LightTheme, DraculaTheme}
}

func ThemeByName(name string) (Theme, bool) {
	for _, t := range AllThemes() {
		if t.Name == name {
			return t, true
		}
	}
	return DarkTheme, false
}

const configDir = ".peaktop"
const configFile = "config.json"

type configData struct {
	Theme string `json:"theme"`
}

func LoadTheme() Theme {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DarkTheme
	}

	configPath := filepath.Join(homeDir, configDir, configFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return DarkTheme
	}

	var config configData
	if err := json.Unmarshal(data, &config); err != nil {
		return DarkTheme
	}

	theme, found := ThemeByName(config.Theme)
	if !found {
		fmt.Fprintf(os.Stderr, "peaktop: unknown theme %q, using dark\n", config.Theme)
		return DarkTheme
	}

	return theme
}
