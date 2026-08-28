package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/nilay-banerjee/noob-cli/internal/dotfiles"
	"github.com/nilay-banerjee/noob-cli/internal/extras"
	"github.com/nilay-banerjee/noob-cli/internal/macos"
)

var doctorCmd = &cobra.Command{
	Use:          "doctor",
	Short:        "Check that this machine matches the expected setup",
	SilenceUsage: true,
	RunE:         runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

type checkResult struct {
	name   string
	ok     bool
	detail string
}

func runDoctor(cmd *cobra.Command, args []string) error {
	var checks []checkResult
	if runtime.GOOS == "darwin" {
		checks = append(checks, checkBrew(), checkFont())
		checks = append(checks, checkSettings()...)
	}
	checks = append(checks, checkDotfileLinks(), checkDotfilesClean(), checkDotfilesSynced())
	checks = append(checks, checkExtras(), checkShell(), checkNvim())

	failed := 0
	for _, c := range checks {
		if c.ok {
			fmt.Printf("  %s %s", okStyle.Render("✓"), c.name)
		} else {
			failed++
			fmt.Printf("  %s %s", failStyle.Render("✗"), c.name)
		}
		if c.detail != "" {
			fmt.Printf(" %s", dimStyle.Render("— "+c.detail))
		}
		fmt.Println()
	}
	if failed > 0 {
		return fmt.Errorf("%d check(s) failed", failed)
	}
	fmt.Println(okStyle.Render("All good."))
	return nil
}

func checkBrew() checkResult {
	if _, err := exec.LookPath("brew"); err != nil {
		return checkResult{"homebrew", false, "not on PATH — run noob-cli to install"}
	}
	return checkResult{"homebrew", true, ""}
}

func checkFont() checkResult {
	home, _ := os.UserHomeDir()
	matches, _ := filepath.Glob(filepath.Join(home, "Library/Fonts", "MesloLGS*NerdFont*"))
	if len(matches) == 0 {
		return checkResult{"meslo nerd font", false, "missing — p10k and Ghostty need it"}
	}
	return checkResult{"meslo nerd font", true, ""}
}

func checkSettings() []checkResult {
	var out []checkResult
	out = append(out, checkCapsRemap())

	repeat, _ := exec.Command("defaults", "read", "-g", "KeyRepeat").Output()
	if strings.TrimSpace(string(repeat)) != "2" {
		out = append(out, checkResult{"fast key repeat", false, "KeyRepeat is not 2"})
	} else {
		out = append(out, checkResult{"fast key repeat", true, ""})
	}

	if _, err := os.Stat("/Applications/Raycast.app"); err == nil {
		hotkey, _ := exec.Command("defaults", "read", "com.raycast.macos", "raycastGlobalHotkey").Output()
		if strings.TrimSpace(string(hotkey)) != "Command-49" {
			out = append(out, checkResult{"raycast hotkey", false, "not set to Cmd-Space"})
		} else {
			out = append(out, checkResult{"raycast hotkey", true, ""})
		}
	}
	return out
}

const (
	capsLockKeyUsage  = "30064771129"
	leftCtrlKeyUsage  = "30064771296"
	rightCtrlKeyUsage = "30064771300"
)

func checkCapsRemap() checkResult {
	if _, err := os.Stat(macos.CapsAgentPlist()); err == nil {
		return checkResult{"caps-to-ctrl", true, "via noob-cli LaunchAgent"}
	}
	if capsRemappedInSystemSettings() {
		return checkResult{"caps-to-ctrl", true, "via System Settings modifier keys"}
	}
	return checkResult{"caps-to-ctrl", false, "not remapped — run noob-cli or set it in System Settings > Keyboard"}
}

func capsRemappedInSystemSettings() bool {
	out, _ := exec.Command("defaults", "-currentHost", "read", "-g").Output()
	mappings := string(out)
	if !strings.Contains(mappings, "modifiermapping") || !strings.Contains(mappings, capsLockKeyUsage) {
		return false
	}
	return strings.Contains(mappings, leftCtrlKeyUsage) || strings.Contains(mappings, rightCtrlKeyUsage)
}

func checkDotfileLinks() checkResult {
	if _, err := os.Stat(dotfiles.RepoDir()); err != nil {
		return checkResult{"dotfile links", false, "~/dotfiles missing — run noob-cli dotfiles link"}
	}
	home, _ := os.UserHomeDir()
	var broken []string
	linked := 0
	for _, m := range dotfiles.Mappings {
		if _, err := os.Stat(filepath.Join(dotfiles.RepoDir(), m.Repo)); err != nil {
			continue
		}
		live := filepath.Join(home, m.Home)
		target, err := os.Readlink(live)
		if err != nil || !strings.HasPrefix(target, dotfiles.RepoDir()) {
			broken = append(broken, "~/"+m.Home)
			continue
		}
		linked++
	}
	if len(broken) > 0 {
		return checkResult{"dotfile links", false, "not linked: " + strings.Join(broken, ", ")}
	}
	return checkResult{"dotfile links", true, fmt.Sprintf("%d linked", linked)}
}

func checkDotfilesClean() checkResult {
	out, err := gitInDotfiles("status", "--porcelain")
	if err != nil {
		return checkResult{"dotfiles clean", false, err.Error()}
	}
	if strings.TrimSpace(out) != "" {
		return checkResult{"dotfiles clean", false, "uncommitted changes in ~/dotfiles"}
	}
	return checkResult{"dotfiles clean", true, ""}
}

func checkDotfilesSynced() checkResult {
	if _, err := gitInDotfiles("fetch", "--quiet"); err != nil {
		return checkResult{"dotfiles synced", true, "couldn't reach remote, skipped"}
	}
	behind, _ := gitInDotfiles("rev-list", "--count", "HEAD..@{u}")
	ahead, _ := gitInDotfiles("rev-list", "--count", "@{u}..HEAD")
	behind, ahead = strings.TrimSpace(behind), strings.TrimSpace(ahead)
	if behind != "0" {
		return checkResult{"dotfiles synced", false, behind + " commit(s) behind — cd ~/dotfiles && git pull"}
	}
	if ahead != "0" {
		return checkResult{"dotfiles synced", false, ahead + " commit(s) unpushed"}
	}
	return checkResult{"dotfiles synced", true, ""}
}

func gitInDotfiles(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dotfiles.RepoDir()
	out, err := cmd.Output()
	return string(out), err
}

func checkExtras() checkResult {
	var missing []string
	for _, dir := range extras.RequiredDirs() {
		if _, err := os.Stat(dir); err != nil {
			missing = append(missing, filepath.Base(dir))
		}
	}
	if len(missing) > 0 {
		return checkResult{"zshrc clones", false, "missing: " + strings.Join(missing, ", ")}
	}
	return checkResult{"zshrc clones", true, ""}
}

func checkShell() checkResult {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, "zsh", "-ic", "exit 0").Run(); err != nil {
		return checkResult{"zsh loads", false, "zsh -ic exited non-zero — check ~/.zshrc"}
	}
	return checkResult{"zsh loads", true, ""}
}

func checkNvim() checkResult {
	if _, err := exec.LookPath("nvim"); err != nil {
		return checkResult{"nvim loads", true, "nvim not installed, skipped"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, "nvim", "--headless", "+q").Run(); err != nil {
		return checkResult{"nvim loads", false, "nvim --headless +q failed — open nvim and check :messages"}
	}
	return checkResult{"nvim loads", true, ""}
}
