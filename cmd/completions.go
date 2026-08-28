package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

func installCompletions(dryRun bool) error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".oh-my-zsh/custom/completions")
	path := filepath.Join(dir, "_noob-cli")
	if dryRun {
		fmt.Printf("[dry-run] write zsh completion to %s\n", path)
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := rootCmd.GenZshCompletion(f); err != nil {
		return err
	}
	fmt.Printf("  installed to %s (new shells pick it up)\n", path)
	return nil
}
