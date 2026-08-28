package extras

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type cloneTarget struct {
	url string
	dir string
}

func fzfGitTarget() cloneTarget {
	home, _ := os.UserHomeDir()
	return cloneTarget{"https://github.com/junegunn/fzf-git.sh.git", filepath.Join(home, "fzf-git.sh")}
}

func ohMyZshTargets() []cloneTarget {
	home, _ := os.UserHomeDir()
	omz := filepath.Join(home, ".oh-my-zsh")
	custom := filepath.Join(omz, "custom")
	return []cloneTarget{
		{"https://github.com/ohmyzsh/ohmyzsh.git", omz},
		{"https://github.com/zsh-users/zsh-autosuggestions.git", filepath.Join(custom, "plugins/zsh-autosuggestions")},
		{"https://github.com/zsh-users/zsh-syntax-highlighting.git", filepath.Join(custom, "plugins/zsh-syntax-highlighting")},
		{"https://github.com/romkatv/powerlevel10k.git", filepath.Join(custom, "themes/powerlevel10k")},
	}
}

func RequiredDirs() []string {
	targets := append([]cloneTarget{fzfGitTarget()}, ohMyZshTargets()...)
	dirs := make([]string, len(targets))
	for i, t := range targets {
		dirs[i] = t.dir
	}
	return dirs
}

func CloneZshrcFzfGit(dryRun bool) error {
	t := fzfGitTarget()
	return clone(t.url, t.dir, dryRun)
}

func CloneZshrcOhMyZsh(dryRun bool) error {
	for _, t := range ohMyZshTargets() {
		if err := clone(t.url, t.dir, dryRun); err != nil {
			return err
		}
	}
	return nil
}

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
