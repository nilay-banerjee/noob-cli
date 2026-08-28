package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/nilay-banerjee/noob-cli/internal/catalog"
)

type Result struct {
	Installed []string
	Skipped   []string
	Failed    []string
	Manual    []string
}

type Installer interface {
	Bootstrap(dryRun bool) error
	Preinstalled(items []catalog.Item) map[string]bool
	Install(items []catalog.Item, skip map[string]bool, dryRun bool) Result
}

func Detect() (Installer, error) {
	if runtime.GOOS == "darwin" {
		brewOnPath()
		// new brew's ask mode y/n-prompts whenever dependencies are involved
		os.Setenv("HOMEBREW_NO_ASK", "1")
		return &Brew{}, nil
	}
	if _, err := exec.LookPath("apt-get"); err == nil {
		return &Linux{cmd: "apt-get", query: []string{"dpkg", "-s"}, pkgName: func(it catalog.Item) string { return it.Apt }}, nil
	}
	if _, err := exec.LookPath("dnf"); err == nil {
		return &Linux{cmd: "dnf", query: []string{"rpm", "-q"}, pkgName: func(it catalog.Item) string { return it.Dnf }}, nil
	}
	return nil, fmt.Errorf("no supported package manager found (need brew, apt-get, or dnf)")
}

// One password prompt for the whole run. A warm sudo timestamp isn't enough:
// brew wipes it on purpose (brew.sh runs `sudo --reset-timestamp`), so the
// password is kept in a private temp file behind SUDO_ASKPASS, which makes
// brew call `sudo -A` and fetch it instead of prompting. Removed on exit.
func SudoSession(dryRun bool) func() {
	if dryRun {
		return func() {}
	}
	fmt.Println("==> Authorizing sudo (one password prompt for the whole run)")
	password, ok := askPassword()
	if !ok {
		fmt.Println("  couldn't authorize sudo; installs may prompt individually")
		return func() {}
	}

	cleanup := func() {}
	if password != "" {
		if dir, err := os.MkdirTemp("", "noob-cli-sudo-*"); err == nil {
			pwFile := filepath.Join(dir, "pw")
			script := filepath.Join(dir, "askpass.sh")
			if os.WriteFile(pwFile, []byte(password+"\n"), 0o600) == nil &&
				os.WriteFile(script, []byte("#!/bin/sh\nexec cat "+pwFile+"\n"), 0o700) == nil {
				os.Setenv("SUDO_ASKPASS", script)
				cleanup = func() {
					os.Unsetenv("SUDO_ASKPASS")
					os.RemoveAll(dir)
				}
			}
		}
	}

	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				_ = exec.Command("sudo", "-n", "-v").Run()
			}
		}
	}()
	return func() {
		close(stop)
		cleanup()
	}
}

// returns ("", true) when there's no TTY but plain `sudo -v` worked
func askPassword() (string, bool) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", run("sudo", "-v") == nil
	}
	for range 3 {
		fmt.Print("Password: ")
		raw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return "", run("sudo", "-v") == nil
		}
		check := exec.Command("sudo", "-S", "-v")
		check.Stdin = strings.NewReader(string(raw) + "\n")
		if check.Run() == nil {
			return string(raw), true
		}
		fmt.Println("Sorry, try again.")
	}
	return "", false
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

type Brew struct{}

// piping curl into bash would make the installer's stdin the pipe, so it
// couldn't prompt for sudo — download first, then run with the terminal attached
func (b *Brew) Bootstrap(dryRun bool) error {
	if brewOnPath() {
		return nil
	}
	if dryRun {
		fmt.Println("[dry-run] would install Homebrew (and Xcode Command Line Tools with it)")
		return nil
	}
	fmt.Println("Homebrew not found, installing it first (you'll be asked for your password)...")
	script := filepath.Join(os.TempDir(), "brew-install.sh")
	if err := run("curl", "-fsSL", "-o", script,
		"https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh"); err != nil {
		return err
	}
	defer os.Remove(script)
	// safe to skip the installer's own prompts: SudoKeepalive already cached credentials
	cmd := exec.Command("/bin/bash", script)
	cmd.Env = append(os.Environ(), "NONINTERACTIVE=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if !brewOnPath() {
		return fmt.Errorf("brew still not found after install")
	}
	return nil
}

func brewOnPath() bool {
	if _, err := exec.LookPath("brew"); err == nil {
		return true
	}
	for _, p := range []string{"/opt/homebrew/bin", "/usr/local/bin"} {
		if _, err := os.Stat(p + "/brew"); err == nil {
			os.Setenv("PATH", p+":"+os.Getenv("PATH"))
			return true
		}
	}
	return false
}

func (b *Brew) installedSet() map[string]bool {
	set := map[string]bool{}
	for _, args := range [][]string{{"list", "--formula", "-1"}, {"list", "--cask", "-1"}} {
		out, err := exec.Command("brew", args...).Output()
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(out), "\n") {
			if line = strings.TrimSpace(line); line != "" {
				set[line] = true
			}
		}
	}
	return set
}

// covers apps installed outside brew (manual downloads), which would make
// `brew install --cask` fail with "already an App at ..."
func appExists(app string) bool {
	if app == "" {
		return false
	}
	home, _ := os.UserHomeDir()
	for _, dir := range []string{"/Applications", filepath.Join(home, "Applications")} {
		if _, err := os.Stat(filepath.Join(dir, app)); err == nil {
			return true
		}
	}
	return false
}

func shortBrewName(name string) string {
	// tap-qualified and versioned names show up unqualified in `brew list`
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	return strings.SplitN(name, "@", 2)[0]
}

func (b *Brew) Preinstalled(items []catalog.Item) map[string]bool {
	set := b.installedSet()
	out := map[string]bool{}
	for _, it := range items {
		if set[it.Brew] || set[shortBrewName(it.Brew)] || appExists(it.App) {
			out[it.Name] = true
		}
	}
	return out
}

func (b *Brew) Install(items []catalog.Item, skip map[string]bool, dryRun bool) Result {
	var res Result
	for _, it := range items {
		if skip[it.Name] {
			res.Skipped = append(res.Skipped, it.Name)
			continue
		}
		args := []string{"install"}
		if it.Cask {
			args = append(args, "--cask")
		}
		args = append(args, it.Brew)
		if dryRun {
			fmt.Printf("[dry-run] brew %s\n", strings.Join(args, " "))
			res.Installed = append(res.Installed, it.Name)
			continue
		}
		fmt.Printf("\n==> Installing %s\n", it.Name)
		if err := run("brew", args...); err != nil {
			res.Failed = append(res.Failed, it.Name)
		} else {
			res.Installed = append(res.Installed, it.Name)
		}
	}
	return res
}

type Linux struct {
	cmd     string
	query   []string
	pkgName func(catalog.Item) string
	updated bool
}

func (l *Linux) Bootstrap(dryRun bool) error {
	return nil
}

func (l *Linux) isInstalled(pkg string) bool {
	args := append(l.query[1:], pkg)
	return exec.Command(l.query[0], args...).Run() == nil
}

func (l *Linux) Preinstalled(items []catalog.Item) map[string]bool {
	out := map[string]bool{}
	for _, it := range items {
		if pkg := l.pkgName(it); pkg != "" && l.isInstalled(pkg) {
			out[it.Name] = true
		}
	}
	return out
}

func (l *Linux) Install(items []catalog.Item, skip map[string]bool, dryRun bool) Result {
	var res Result
	for _, it := range items {
		if it.Cask {
			continue
		}
		pkg := l.pkgName(it)
		if pkg == "" {
			hint := it.Hint
			if hint == "" {
				hint = "no " + l.cmd + " package"
			}
			res.Manual = append(res.Manual, fmt.Sprintf("%s (%s)", it.Name, hint))
			continue
		}
		if skip[it.Name] {
			res.Skipped = append(res.Skipped, it.Name)
			continue
		}
		if dryRun {
			fmt.Printf("[dry-run] sudo %s install -y %s\n", l.cmd, pkg)
			res.Installed = append(res.Installed, it.Name)
			continue
		}
		if l.cmd == "apt-get" && !l.updated {
			_ = run("sudo", "apt-get", "update")
			l.updated = true
		}
		fmt.Printf("\n==> Installing %s\n", it.Name)
		if err := run("sudo", l.cmd, "install", "-y", pkg); err != nil {
			res.Failed = append(res.Failed, it.Name)
		} else {
			res.Installed = append(res.Installed, it.Name)
		}
	}
	return res
}
