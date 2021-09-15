package cmd

import (
	"github.com/charmbracelet/lipgloss"
)

// italic, bright; use to call out instructions or things of interest
var styleNotice = lipgloss.NewStyle().
	Italic(true).
	Foreground(lipgloss.AdaptiveColor{Light: "15", Dark: "15"})

// bold and green, like Kermit, for when things went right
var styleSuccess = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.AdaptiveColor{Light: "34", Dark: "76"})

// oh dear
var styleFailure = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.AdaptiveColor{Light: "160", Dark: "196"})
