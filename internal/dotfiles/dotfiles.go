package dotfiles

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Link struct {
	Repo string
	Home string
}

var Links = []Link{
	{Repo: "zsh/.zshrc", Home: ".zshrc"},
	{Repo: "zsh/.p10k.zsh", Home: ".p10k.zsh"},
	{Repo: "tmux/.tmux.conf", Home: ".tmux.conf"},
	{Repo: "git/.gitconfig", Home: ".gitconfig"},
	{Repo: "config/nvim", Home: ".config/nvim"},
	{Repo: "config/aerospace", Home: ".config/aerospace"},
	{Repo: "config/ghostty", Home: ".config/ghostty"},
	{Repo: "ghostty/config", Home: "Library/Application Support/com.mitchellh.ghostty/config"},
	{Repo: "config/lazygit", Home: ".config/lazygit"},
}

func RepoDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "dotfiles")
}

func homePath(l Link) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, l.Home)
}

func repoPath(l Link) string {
	return filepath.Join(RepoDir(), l.Repo)
}

func pointsIntoRepo(path string) bool {
	target, err := os.Readlink(path)
	if err != nil {
		return false
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	rel, err := filepath.Rel(RepoDir(), target)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, "../")
}

func git(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = RepoDir()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Init harvests the machine's live configs into ~/dotfiles and symlinks them back.
func Init(dryRun bool) error {
	for _, l := range Links {
		src := homePath(l)
		dst := repoPath(l)
		info, err := os.Lstat(src)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if pointsIntoRepo(src) {
				fmt.Printf("  already linked: ~/%s\n", l.Home)
			} else {
				fmt.Printf("  skipping ~/%s: symlink to somewhere else\n", l.Home)
			}
			continue
		}
		if _, err := os.Stat(dst); err == nil {
			fmt.Printf("  skipping ~/%s: %s already exists in repo\n", l.Home, l.Repo)
			continue
		}
		if dryRun {
			fmt.Printf("[dry-run] move ~/%s -> ~/dotfiles/%s and symlink back\n", l.Home, l.Repo)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.Rename(src, dst); err != nil {
			return err
		}
		if err := os.Symlink(dst, src); err != nil {
			return err
		}
		fmt.Printf("  harvested ~/%s -> ~/dotfiles/%s\n", l.Home, l.Repo)
	}
	if dryRun {
		return nil
	}
	if _, err := os.Stat(filepath.Join(RepoDir(), ".git")); os.IsNotExist(err) {
		if err := git("init"); err != nil {
			return err
		}
		if err := git("add", "-A"); err != nil {
			return err
		}
		if err := git("commit", "-m", "Initial dotfiles harvest"); err != nil {
			return err
		}
	}
	return nil
}

const DefaultRepo = "https://github.com/nilay-banerjee/dotfiles.git"

// Setup makes ~/dotfiles configs live on this machine: clone if needed, then symlink.
func Setup(repoURL string, dryRun bool) error {
	if repoURL == "" {
		repoURL = DefaultRepo
	}
	if _, err := os.Stat(RepoDir()); os.IsNotExist(err) {
		if dryRun {
			fmt.Printf("[dry-run] git clone %s ~/dotfiles\n", repoURL)
		} else {
			cmd := exec.Command("git", "clone", repoURL, RepoDir())
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}
	backupDir := filepath.Join(os.TempDir(), "dotfiles-backup-"+time.Now().Format("20060102-150405"))
	for _, l := range Links {
		src := repoPath(l)
		dst := homePath(l)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}
		if info, err := os.Lstat(dst); err == nil {
			if info.Mode()&os.ModeSymlink != 0 && pointsIntoRepo(dst) {
				fmt.Printf("  already linked: ~/%s\n", l.Home)
				continue
			}
			if dryRun {
				fmt.Printf("[dry-run] back up ~/%s, then symlink ~/dotfiles/%s\n", l.Home, l.Repo)
				continue
			}
			if err := os.MkdirAll(backupDir, 0o755); err != nil {
				return err
			}
			backup := filepath.Join(backupDir, filepath.Base(l.Home))
			if err := os.Rename(dst, backup); err != nil {
				return err
			}
			fmt.Printf("  backed up ~/%s -> %s\n", l.Home, backup)
		} else if dryRun {
			fmt.Printf("[dry-run] symlink ~/%s -> ~/dotfiles/%s\n", l.Home, l.Repo)
			continue
		}
		if dryRun {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.Symlink(src, dst); err != nil {
			return err
		}
		fmt.Printf("  linked ~/%s -> ~/dotfiles/%s\n", l.Home, l.Repo)
	}
	return nil
}
