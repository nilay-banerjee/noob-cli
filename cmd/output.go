package cmd

import (
	"fmt"
	"strings"

	"github.com/nilay-banerjee/noob-cli/internal/installer"
)

func printPlan(p plan, preinstalled map[string]bool, mac bool) {
	var clis, casks []string
	skipped := 0
	for _, it := range p.items {
		name := it.Name
		if preinstalled[it.Name] {
			name += " ✓"
			skipped++
		}
		if it.Cask {
			casks = append(casks, name)
		} else {
			clis = append(clis, name)
		}
	}
	fmt.Printf("Plan:\n  CLIs:  %s\n", strings.Join(clis, ", "))
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
		fmt.Printf("  ✓ = already installed, %d will be skipped\n", skipped)
	}
}

func printSummary(res installer.Result) {
	fmt.Println("\nDone.")
	if len(res.Installed) > 0 {
		fmt.Printf("  installed: %s\n", strings.Join(res.Installed, ", "))
	}
	if len(res.Skipped) > 0 {
		fmt.Printf("  already there: %s\n", strings.Join(res.Skipped, ", "))
	}
	if len(res.Manual) > 0 {
		fmt.Println("  needs manual install:")
		for _, m := range res.Manual {
			fmt.Printf("    - %s\n", m)
		}
	}
	if len(res.Failed) > 0 {
		fmt.Printf("  FAILED: %s\n", strings.Join(res.Failed, ", "))
	}
}
