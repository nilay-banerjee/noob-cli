package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/nilay-banerjee/noob-cli/internal/installer"
)

var (
	headStyle = lipgloss.NewStyle().Bold(true)
	okStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	dimStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	warnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	failStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

func printPlan(p plan, preinstalled map[string]bool, mac bool) {
	var clis, casks []string
	skipped := 0
	for _, it := range p.items {
		name := it.Name
		if preinstalled[it.Name] {
			name = dimStyle.Render(name + " ✓")
			skipped++
		}
		if it.Cask {
			casks = append(casks, name)
		} else {
			clis = append(clis, name)
		}
	}
	fmt.Println(headStyle.Render("Plan"))
	fmt.Printf("  CLIs:  %s\n", strings.Join(clis, ", "))
	if mac && len(casks) > 0 {
		fmt.Printf("  Casks: %s\n", strings.Join(casks, ", "))
	}
	if len(p.settings) > 0 {
		var names []string
		for _, s := range p.settings {
			names = append(names, s.Name)
		}
		fmt.Printf("  Settings: %s\n", strings.Join(names, ", "))
	}
	fmt.Printf("  Dotfiles: %v\n", p.doDotfiles)
	if skipped > 0 {
		fmt.Println(dimStyle.Render(fmt.Sprintf("  ✓ = already installed, %d will be skipped", skipped)))
	}
}

func printSummary(res installer.Result) {
	fmt.Println()
	fmt.Println(headStyle.Render("Summary"))
	if len(res.Installed) > 0 {
		fmt.Printf("  %s %s\n",
			okStyle.Render(fmt.Sprintf("✓ installed (%d):", len(res.Installed))),
			strings.Join(res.Installed, ", "))
	}
	if len(res.Skipped) > 0 {
		fmt.Printf("  %s\n",
			dimStyle.Render(fmt.Sprintf("• already there (%d): %s", len(res.Skipped), strings.Join(res.Skipped, ", "))))
	}
	if len(res.Manual) > 0 {
		fmt.Printf("  %s\n", warnStyle.Render(fmt.Sprintf("! needs manual install (%d):", len(res.Manual))))
		for _, m := range res.Manual {
			fmt.Printf("      - %s\n", m)
		}
	}
	if len(res.Failed) > 0 {
		fmt.Printf("  %s %s\n",
			failStyle.Render(fmt.Sprintf("✗ failed (%d):", len(res.Failed))),
			strings.Join(res.Failed, ", "))
		fmt.Println(dimStyle.Render("    re-run noob-cli to retry just these — everything else is skipped"))
	}
	if len(res.Installed)+len(res.Skipped)+len(res.Manual)+len(res.Failed) == 0 {
		fmt.Println(dimStyle.Render("  nothing to do"))
	}
}
