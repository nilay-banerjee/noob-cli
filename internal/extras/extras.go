package extras

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func clone(url, dir string, dryRun bool) error {
	if _, err := os.Stat(dir); err == nil {
		fmt.Printf("  already present: %s\n", dir)
		return nil
	}
	if dryRun {
		fmt.Printf("[dry-run] git clone %s %s\n", url, dir)
		return nil
	}
	cmd := exec.Command("git", "clone", "--depth", "1", url, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// .zshrc sources ~/fzf-git.sh/fzf-git.sh, so the clone has to live at that path
func FzfGit(dryRun bool) error {
	home, _ := os.UserHomeDir()
	return clone("https://github.com/junegunn/fzf-git.sh.git", filepath.Join(home, "fzf-git.sh"), dryRun)
}

// .zshrc expects oh-my-zsh with its plugins and the p10k theme as custom clones,
// not the brew formulas
func ZshEnv(dryRun bool) error {
	home, _ := os.UserHomeDir()
	omz := filepath.Join(home, ".oh-my-zsh")
	custom := filepath.Join(omz, "custom")
	repos := []struct{ url, dir string }{
		{"https://github.com/ohmyzsh/ohmyzsh.git", omz},
		{"https://github.com/zsh-users/zsh-autosuggestions.git", filepath.Join(custom, "plugins/zsh-autosuggestions")},
		{"https://github.com/zsh-users/zsh-syntax-highlighting.git", filepath.Join(custom, "plugins/zsh-syntax-highlighting")},
		{"https://github.com/romkatv/powerlevel10k.git", filepath.Join(custom, "themes/powerlevel10k")},
	}
	for _, r := range repos {
		if err := clone(r.url, r.dir, dryRun); err != nil {
			return err
		}
	}
	return nil
}
