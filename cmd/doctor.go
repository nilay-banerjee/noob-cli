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
	"github.com/nilay-banerjee/noob-cli/internal/installer"
	"github.com/nilay-banerjee/noob-cli/internal/macos"
)

var doctorFix bool

var doctorCmd = &cobra.Command{
	Use:          "doctor",
	Short:        "Check that this machine matches the expected setup",
	SilenceUsage: true,
	RunE:         runDoctor,
}

func init() {
	doctorCmd.Flags().BoolVarP(&doctorFix, "auto-fix", "a", false, "apply fixes for anything auto-correctable, then re-check")
	rootCmd.AddCommand(doctorCmd)
}

type checkResult struct {
	name   string
	ok     bool
	detail string
	fix    func() error
}

func runDoctor(cmd *cobra.Command, args []string) error {
	checks := gatherChecks()
	failed := printChecks(checks)
	if failed == 0 {
		fmt.Println(okStyle.Render("All good."))
		return nil
	}
	if !doctorFix {
		return fmt.Errorf("%d check(s) failed (re-run with -a to auto-fix)", failed)
	}

	for _, c := range checks {
		if c.ok || c.fix == nil {
			continue
		}
		fmt.Printf("\n==> Fixing %s\n", c.name)
		if err := c.fix(); err != nil {
			fmt.Printf("  fix failed: %v\n", err)
		}
	}

	fmt.Println("\nRe-checking:")
	failed = printChecks(gatherChecks())
	if failed > 0 {
		return fmt.Errorf("%d check(s) still failing", failed)
	}
	fmt.Println(okStyle.Render("All good."))
	return nil
}

func gatherChecks() []checkResult {
	var checks []checkResult
	if runtime.GOOS == "darwin" {
		checks = append(checks, checkBrew(), checkFont())
		checks = append(checks, checkSettings()...)
	}
	checks = append(checks, checkDotfileLinks(), checkDotfilesClean(), checkDotfilesSynced())
	checks = append(checks, checkGitConfigLocal(), checkExtras(), checkShell(), checkNvim())
	return checks
}

func checkGitConfigLocal() checkResult {
	home, _ := os.UserHomeDir()
	tracked, err := os.ReadFile(filepath.Join(dotfiles.RepoDir(), "git/.gitconfig"))
	if err != nil || !strings.Contains(string(tracked), ".gitconfig.local") {
		return checkResult{"gitconfig.local", true, "not referenced, skipped", nil}
	}
	if _, err := os.Stat(filepath.Join(home, ".gitconfig.local")); err != nil {
		return checkResult{"gitconfig.local", false, "missing — create it with your [user] signingkey", nil}
	}
	return checkResult{"gitconfig.local", true, "", nil}
}

func printChecks(checks []checkResult) int {
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
		if !c.ok && c.fix == nil {
			fmt.Printf(" %s", warnStyle.Render("(no auto-fix)"))
		}
		fmt.Println()
	}
	return failed
}

func applySetting(name string) func() error {
	return func() error {
		s, ok := macos.ByName(name)
		if !ok {
			return fmt.Errorf("unknown setting %q", name)
		}
		return s.Apply(false)
	}
}

func checkBrew() checkResult {
	if _, err := exec.LookPath("brew"); err != nil {
		return checkResult{"homebrew", false, "not on PATH", func() error {
			inst, err := installer.Detect()
			if err != nil {
				return err
			}
			return inst.Bootstrap(false)
		}}
	}
	return checkResult{"homebrew", true, "", nil}
}

func checkFont() checkResult {
	home, _ := os.UserHomeDir()
	matches, _ := filepath.Glob(filepath.Join(home, "Library/Fonts", "MesloLGS*NerdFont*"))
	if len(matches) == 0 {
		return checkResult{"meslo nerd font", false, "missing — p10k and Ghostty need it", func() error {
			cmd := exec.Command("brew", "install", "--cask", "font-meslo-lg-nerd-font")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}}
	}
	return checkResult{"meslo nerd font", true, "", nil}
}

func checkSettings() []checkResult {
	var out []checkResult
	out = append(out, checkCapsRemap())

	repeat, _ := exec.Command("defaults", "read", "-g", "KeyRepeat").Output()
	if strings.TrimSpace(string(repeat)) != "2" {
		out = append(out, checkResult{"fast key repeat", false, "KeyRepeat is not 2", applySetting("fast-key-repeat")})
	} else {
		out = append(out, checkResult{"fast key repeat", true, "", nil})
	}

	if _, err := os.Stat("/Applications/Raycast.app"); err == nil {
		hotkey, _ := exec.Command("defaults", "read", "com.raycast.macos", "raycastGlobalHotkey").Output()
		if strings.TrimSpace(string(hotkey)) != "Command-49" {
			out = append(out, checkResult{"raycast hotkey", false, "not set to Cmd-Space", applySetting("raycast-hotkey")})
		} else {
			out = append(out, checkResult{"raycast hotkey", true, "", nil})
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
		return checkResult{"caps-to-ctrl", true, "via noob-cli LaunchAgent", nil}
	}
	if capsRemappedInSystemSettings() {
		return checkResult{"caps-to-ctrl", true, "via System Settings modifier keys", nil}
	}
	return checkResult{"caps-to-ctrl", false, "not remapped", applySetting("caps-to-ctrl")}
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
	relink := func() error { return dotfiles.Link("", false) }
	if _, err := os.Stat(dotfiles.RepoDir()); err != nil {
		return checkResult{"dotfile links", false, "~/dotfiles missing", relink}
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
		return checkResult{"dotfile links", false, "not linked: " + strings.Join(broken, ", "), relink}
	}
	return checkResult{"dotfile links", true, fmt.Sprintf("%d linked", linked), nil}
}

func checkDotfilesClean() checkResult {
	out, err := gitInDotfiles("status", "--porcelain")
	if err != nil {
		return checkResult{"dotfiles clean", false, err.Error(), nil}
	}
	if strings.TrimSpace(out) != "" {
		return checkResult{"dotfiles clean", false, "uncommitted changes in ~/dotfiles", nil}
	}
	return checkResult{"dotfiles clean", true, "", nil}
}

func checkDotfilesSynced() checkResult {
	if _, err := gitInDotfiles("fetch", "--quiet"); err != nil {
		return checkResult{"dotfiles synced", true, "couldn't reach remote, skipped", nil}
	}
	behind, _ := gitInDotfiles("rev-list", "--count", "HEAD..@{u}")
	ahead, _ := gitInDotfiles("rev-list", "--count", "@{u}..HEAD")
	behind, ahead = strings.TrimSpace(behind), strings.TrimSpace(ahead)
	if behind != "0" {
		return checkResult{"dotfiles synced", false, behind + " commit(s) behind", func() error {
			_, err := gitInDotfiles("pull", "--ff-only")
			return err
		}}
	}
	if ahead != "0" {
		return checkResult{"dotfiles synced", false, ahead + " commit(s) unpushed", nil}
	}
	return checkResult{"dotfiles synced", true, "", nil}
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
		return checkResult{"zshrc clones", false, "missing: " + strings.Join(missing, ", "), func() error {
			if err := extras.CloneZshrcFzfGit(false); err != nil {
				return err
			}
			if err := extras.CloneZshrcOhMyZsh(false); err != nil {
				return err
			}
			return extras.CloneTmuxTpm(false)
		}}
	}
	return checkResult{"zshrc clones", true, "", nil}
}

func checkShell() checkResult {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, "zsh", "-ic", "exit 0").Run(); err != nil {
		return checkResult{"zsh loads", false, "zsh -ic exited non-zero — check ~/.zshrc", nil}
	}
	return checkResult{"zsh loads", true, "", nil}
}

func checkNvim() checkResult {
	if _, err := exec.LookPath("nvim"); err != nil {
		return checkResult{"nvim loads", true, "nvim not installed, skipped", nil}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, "nvim", "--headless", "+q").Run(); err != nil {
		return checkResult{"nvim loads", false, "nvim --headless +q failed — open nvim and check :messages", nil}
	}
	return checkResult{"nvim loads", true, "", nil}
}
