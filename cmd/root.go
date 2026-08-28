package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/nilay-banerjee/noob-cli/internal/catalog"
	"github.com/nilay-banerjee/noob-cli/internal/dotfiles"
	"github.com/nilay-banerjee/noob-cli/internal/extras"
	"github.com/nilay-banerjee/noob-cli/internal/installer"
	"github.com/nilay-banerjee/noob-cli/internal/macos"
)

var (
	serverFlag   bool
	dailyFlag    bool
	ultimateFlag bool
	include      []string
	exclude      []string
	dryRun       bool
	skipDefaults bool
	skipDotfiles bool
	dotfilesRepo string
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "noob-cli",
	Version: version,
	Short:   "Set up a fresh machine the way Nilay likes it",
	Long: `noob-cli installs the apps and CLIs for a fresh macOS (or Linux server) machine,
links dotfiles from ~/dotfiles, and applies sane macOS defaults.

Run without flags for an interactive step-through. Pass a tier flag to skip the prompts.`,
	SilenceUsage: true,
	RunE:         runSetup,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&serverFlag, "server", false, "lightweight CLI-only tier, works on Linux too")
	rootCmd.Flags().BoolVar(&dailyFlag, "daily", false, "everything used daily (default)")
	rootCmd.Flags().BoolVar(&ultimateFlag, "ultimate", false, "daily plus the nice-to-haves")
	rootCmd.Flags().StringSliceVar(&include, "include", nil, "extra apps to install on top of the tier")
	rootCmd.Flags().StringSliceVar(&exclude, "exclude", nil, "apps to drop from the tier")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "print what would happen without changing anything")
	rootCmd.Flags().BoolVar(&skipDefaults, "skip-defaults", false, "don't touch macOS settings")
	rootCmd.Flags().BoolVar(&skipDotfiles, "skip-dotfiles", false, "don't link dotfiles")
	rootCmd.Flags().StringVar(&dotfilesRepo, "dotfiles-repo", "", "git URL to clone into ~/dotfiles if it's missing (default "+dotfiles.DefaultRepo+")")
}

func runSetup(cmd *cobra.Command, args []string) error {
	inst, err := installer.Detect()
	if err != nil {
		return err
	}
	preinstalled := inst.Preinstalled(catalog.Items)

	p, mac, err := buildPlan(preinstalled)
	if err != nil {
		return err
	}
	printPlan(p, preinstalled, mac)

	stopSudo := installer.SudoSession(dryRun)
	defer stopSudo()
	if err := inst.Bootstrap(dryRun); err != nil {
		return fmt.Errorf("homebrew bootstrap failed: %w\nFix that (the account needs to be an Administrator), then re-run noob-cli — everything is safe to re-run", err)
	}
	res := inst.Install(p.items, preinstalled, dryRun)
	if len(res.Failed) > 0 && !dryRun {
		fmt.Printf("\n==> Retrying %d failed install(s)\n", len(res.Failed))
		retry := inst.Install(itemsByName(res.Failed), nil, false)
		res.Installed = append(res.Installed, retry.Installed...)
		res.Failed = retry.Failed
	}

	if p.doDotfiles {
		fmt.Println("\n==> Dotfiles")
		if err := dotfiles.Setup(dotfilesRepo, dryRun); err != nil {
			fmt.Printf("  dotfiles skipped: %v\n", err)
		}
	}

	if hasItem(p.items, "fzf") {
		fmt.Println("\n==> fzf-git.sh")
		if err := extras.FzfGit(dryRun); err != nil {
			fmt.Printf("  failed: %v\n", err)
		}
	}

	if hasItem(p.items, "zsh") {
		fmt.Println("\n==> oh-my-zsh + plugins + powerlevel10k")
		if err := extras.ZshEnv(dryRun); err != nil {
			fmt.Printf("  failed: %v\n", err)
		}
	}

	for _, s := range p.settings {
		fmt.Printf("\n==> Setting: %s\n", s.Desc)
		if err := s.Apply(dryRun); err != nil {
			fmt.Printf("  failed: %v\n", err)
		}
	}

	printSummary(res)
	return nil
}

func buildPlan(preinstalled map[string]bool) (plan, bool, error) {
	mac := runtime.GOOS == "darwin"
	tier, n, err := pickedTier()
	if err != nil {
		return plan{}, mac, err
	}
	if n == 0 && isTTY() {
		p, err := interactivePlan(mac, preinstalled)
		return p, mac, err
	}
	items, err := catalog.Resolve(tier, include, exclude)
	if err != nil {
		return plan{}, mac, err
	}
	return plan{
		items:      items,
		settings:   defaultSettings(tier, items, mac),
		doDotfiles: !skipDotfiles,
	}, mac, nil
}

func pickedTier() (catalog.Tier, int, error) {
	n := 0
	tier := catalog.Daily
	if serverFlag {
		n++
		tier = catalog.Server
	}
	if dailyFlag {
		n++
		tier = catalog.Daily
	}
	if ultimateFlag {
		n++
		tier = catalog.Ultimate
	}
	if n > 1 {
		return tier, n, fmt.Errorf("pick one of --server, --daily, --ultimate")
	}
	return tier, n, nil
}

func defaultSettings(tier catalog.Tier, items []catalog.Item, mac bool) []macos.Setting {
	if !mac || skipDefaults || tier == catalog.Server {
		return nil
	}
	var out []macos.Setting
	for _, s := range macos.Settings {
		if s.Name == "raycast-hotkey" && !hasItem(items, "raycast") {
			continue
		}
		out = append(out, s)
	}
	return out
}

func hasItem(items []catalog.Item, name string) bool {
	for _, it := range items {
		if it.Name == name {
			return true
		}
	}
	return false
}

func isTTY() bool {
	info, err := os.Stdin.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
