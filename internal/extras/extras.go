package extras

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// .zshrc sources ~/fzf-git.sh/fzf-git.sh, so the clone has to live at that path
func FzfGit(dryRun bool) error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "fzf-git.sh")
	if _, err := os.Stat(dir); err == nil {
		fmt.Println("  fzf-git.sh already present")
		return nil
	}
	if dryRun {
		fmt.Printf("[dry-run] git clone https://github.com/junegunn/fzf-git.sh.git %s\n", dir)
		return nil
	}
	cmd := exec.Command("git", "clone", "--depth", "1", "https://github.com/junegunn/fzf-git.sh.git", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
