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

func tmuxTpmTarget() cloneTarget {
	home, _ := os.UserHomeDir()
	return cloneTarget{"https://github.com/tmux-plugins/tpm.git", filepath.Join(home, ".tmux/plugins/tpm")}
}

func RequiredDirs() []string {
	targets := append([]cloneTarget{fzfGitTarget(), tmuxTpmTarget()}, ohMyZshTargets()...)
	dirs := make([]string, len(targets))
	for i, t := range targets {
		dirs[i] = t.dir
	}
	return dirs
}

func NodeViaNvm(dryRun bool) error {
	if _, err := exec.LookPath("node"); err == nil {
		fmt.Println("  node already available")
		return nil
	}
	home, _ := os.UserHomeDir()
	nvmDir := filepath.Join(home, ".nvm")
	if err := clone("https://github.com/nvm-sh/nvm.git", nvmDir, dryRun); err != nil {
		return err
	}
	if dryRun {
		fmt.Println("[dry-run] nvm install --lts")
		return nil
	}
	cmd := exec.Command("bash", "-c", "source "+nvmDir+"/nvm.sh && nvm install --lts")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func CloneTmuxTpm(dryRun bool) error {
	t := tmuxTpmTarget()
	if err := clone(t.url, t.dir, dryRun); err != nil {
		return err
	}
	if dryRun {
		return nil
	}
	installer := filepath.Join(t.dir, "bin/install_plugins")
	cmd := exec.Command(installer)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
