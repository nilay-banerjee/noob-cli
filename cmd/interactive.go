package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/nilay-banerjee/noob-cli/internal/catalog"
	"github.com/nilay-banerjee/noob-cli/internal/macos"
)

type plan struct {
	items      []catalog.Item
	settings   []macos.Setting
	doDotfiles bool
}

func interactivePlan(mac bool) (plan, error) {
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
	cliOpts, caskOpts := itemOptions(pre)

	var cliNames, caskNames, settingNames []string
	doDotfiles := !skipDotfiles

	groups := []*huh.Group{
		multiSelectGroup("Step 1/3 — CLI tools", cliOpts, &cliNames),
	}
	if mac {
		groups = append(groups, multiSelectGroup("Step 2/3 — Apps (casks)", caskOpts, &caskNames))
	}
	if mac && !skipDefaults && tier != catalog.Server {
		groups = append(groups, multiSelectGroup("Step 3/3 — macOS settings", settingOptions(), &settingNames))
	}
	groups = append(groups, huh.NewGroup(huh.NewConfirm().
		Title("Link dotfiles from ~/dotfiles?").
		Value(&doDotfiles)))

	if err := huh.NewForm(groups...).Run(); err != nil {
		return plan{}, err
	}
	return plan{
		items:      itemsByName(append(cliNames, caskNames...)),
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

func itemOptions(pre map[string]bool) (clis, casks []huh.Option[string]) {
	for _, it := range catalog.Items {
		label := fmt.Sprintf("%s — %s", it.Name, it.Desc)
		opt := huh.NewOption(label, it.Name).Selected(pre[it.Name])
		if it.Cask {
			casks = append(casks, opt)
		} else {
			clis = append(clis, opt)
		}
	}
	return clis, casks
}

func settingOptions() []huh.Option[string] {
	var opts []huh.Option[string]
	for _, s := range macos.Settings {
		opts = append(opts, huh.NewOption(s.Desc, s.Name).Selected(true))
	}
	return opts
}

func multiSelectGroup(title string, opts []huh.Option[string], value *[]string) *huh.Group {
	return huh.NewGroup(huh.NewMultiSelect[string]().
		Title(title).
		Options(opts...).
		Height(14).
		Value(value))
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
