package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/nilay-banerjee/noob-cli/internal/catalog"
	"github.com/nilay-banerjee/noob-cli/internal/macos"
)

type plan struct {
	items      []catalog.Item
	settings   []macos.Setting
	doDotfiles bool
}

func interactivePlan(mac bool, preinstalled map[string]bool) (plan, error) {
	tier, err := askTier()
	if err != nil {
		return plan{}, err
	}
	preselected, err := catalog.Resolve(tier, include, exclude)
	if err != nil {
		return plan{}, err
	}

	pre := map[string]bool{}
	for _, it := range preselected {
		pre[it.Name] = true
	}
	cliOpts, caskOpts, cliDone, caskDone := itemOptions(pre, preinstalled)

	var cliNames, caskNames, settingNames []string
	doDotfiles := !skipDotfiles

	groups := []*huh.Group{
		multiSelectGroup("Step 1/3 — CLI tools", cliOpts, cliDone, &cliNames),
	}
	if mac {
		groups = append(groups, multiSelectGroup("Step 2/3 — Apps (casks)", caskOpts, caskDone, &caskNames))
	}
	if mac && !skipDefaults && tier != catalog.Server {
		groups = append(groups, multiSelectGroup("Step 3/3 — macOS settings", settingOptions(), nil, &settingNames))
	}
	groups = append(groups, huh.NewGroup(huh.NewConfirm().
		Title("Link dotfiles from ~/dotfiles?").
		Value(&doDotfiles)))

	if err := huh.NewForm(groups...).Run(); err != nil {
		return plan{}, err
	}
	items := itemsByName(append(cliNames, caskNames...))
	items = append(items, tierItemsAlreadyInstalled(pre, preinstalled)...)
	return plan{
		items:      items,
		settings:   settingsByName(settingNames),
		doDotfiles: doDotfiles,
	}, nil
}

func askTier() (catalog.Tier, error) {
	choice := "daily"
	err := huh.NewSelect[string]().
		Title("Which tier?").
		Options(
			huh.NewOption("daily — everything I use daily (default)", "daily"),
			huh.NewOption("server — lightweight CLIs only, Linux-friendly", "server"),
			huh.NewOption("ultimate — daily plus the nice-to-haves", "ultimate"),
		).
		Value(&choice).
		Run()
	if err != nil {
		return catalog.Daily, err
	}
	switch choice {
	case "server":
		return catalog.Server, nil
	case "ultimate":
		return catalog.Ultimate, nil
	}
	return catalog.Daily, nil
}

func itemOptions(pre, preinstalled map[string]bool) (clis, casks []huh.Option[string], clisDone, casksDone []string) {
	for _, it := range catalog.Items {
		if preinstalled[it.Name] {
			if it.Cask {
				casksDone = append(casksDone, it.Name)
			} else {
				clisDone = append(clisDone, it.Name)
			}
			continue
		}
		opt := huh.NewOption(fmt.Sprintf("%s — %s", it.Name, it.Desc), it.Name).Selected(pre[it.Name])
		if it.Cask {
			casks = append(casks, opt)
		} else {
			clis = append(clis, opt)
		}
	}
	return clis, casks, clisDone, casksDone
}

func settingOptions() []huh.Option[string] {
	var opts []huh.Option[string]
	for _, s := range macos.Settings {
		opts = append(opts, huh.NewOption(s.Desc, s.Name).Selected(true))
	}
	return opts
}

func multiSelectGroup(title string, opts []huh.Option[string], installed []string, value *[]string) *huh.Group {
	if len(opts) == 0 {
		return huh.NewGroup(huh.NewNote().
			Title(title).
			Description(fmt.Sprintf("Everything here is already installed, nothing to pick: %s",
				strings.Join(installed, ", "))))
	}
	fields := []huh.Field{}
	if len(installed) > 0 {
		fields = append(fields, huh.NewNote().
			Description(fmt.Sprintf("Already installed, skipped: %s", strings.Join(installed, ", "))))
	}
	fields = append(fields, huh.NewMultiSelect[string]().
		Title(title).
		Options(opts...).
		Height(14).
		Value(value))
	return huh.NewGroup(fields...)
}

func tierItemsAlreadyInstalled(pre, preinstalled map[string]bool) []catalog.Item {
	var items []catalog.Item
	for _, it := range catalog.Items {
		if pre[it.Name] && preinstalled[it.Name] {
			items = append(items, it)
		}
	}
	return items
}

func itemsByName(names []string) []catalog.Item {
	var items []catalog.Item
	for _, name := range names {
		if it, ok := catalog.ByName(name); ok {
			items = append(items, it)
		}
	}
	return items
}

func settingsByName(names []string) []macos.Setting {
	var settings []macos.Setting
	for _, name := range names {
		if s, ok := macos.ByName(name); ok {
			settings = append(settings, s)
		}
	}
	return settings
}
