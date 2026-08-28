package installer

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/nilay-banerjee/noob-cli/internal/catalog"
)

type Result struct {
	Installed []string
	Skipped   []string
	Failed    []string
	Manual    []string
}

type Installer interface {
	Install(items []catalog.Item, dryRun bool) Result
}

func Detect() (Installer, error) {
	if runtime.GOOS == "darwin" {
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

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

type Brew struct{}

func (b *Brew) ensure(dryRun bool) error {
	if _, err := exec.LookPath("brew"); err == nil {
		return nil
	}
	for _, p := range []string{"/opt/homebrew/bin", "/usr/local/bin"} {
		if _, err := os.Stat(p + "/brew"); err == nil {
			os.Setenv("PATH", p+":"+os.Getenv("PATH"))
			return nil
		}
	}
	if dryRun {
		fmt.Println("[dry-run] would install Homebrew")
		return nil
	}
	fmt.Println("Homebrew not found, installing it first...")
	return run("/bin/bash", "-c",
		`curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh | /bin/bash`)
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

func (b *Brew) Install(items []catalog.Item, dryRun bool) Result {
	var res Result
	if err := b.ensure(dryRun); err != nil {
		res.Failed = append(res.Failed, "homebrew: "+err.Error())
		return res
	}
	installed := map[string]bool{}
	if !dryRun {
		installed = b.installedSet()
	}
	for _, it := range items {
		// tap-qualified and versioned names show up unqualified in `brew list`
		short := it.Brew
		if i := strings.LastIndex(short, "/"); i >= 0 {
			short = short[i+1:]
		}
		short = strings.SplitN(short, "@", 2)[0]
		if installed[it.Brew] || installed[short] {
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

func (l *Linux) isInstalled(pkg string) bool {
	args := append(l.query[1:], pkg)
	return exec.Command(l.query[0], args...).Run() == nil
}

func (l *Linux) Install(items []catalog.Item, dryRun bool) Result {
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
		if !dryRun && l.isInstalled(pkg) {
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
