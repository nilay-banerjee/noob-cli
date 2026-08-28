package macos

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Setting struct {
	Name  string
	Desc  string
	Apply func(dryRun bool) error
}

var Settings = []Setting{
	{Name: "caps-to-ctrl", Desc: "Remap Caps Lock to Ctrl (survives reboot via LaunchAgent)", Apply: capsToCtrl},
	{Name: "raycast-hotkey", Desc: "Give Cmd-Space to Raycast: disable Spotlight hotkey, set Raycast's, launch it", Apply: raycastHotkey},
	{Name: "fast-key-repeat", Desc: "Key repeat faster than System Settings allows", Apply: fastKeyRepeat},
	{Name: "finder-dock", Desc: "Finder/Dock sanity: extensions, hidden files, path bar, snappier Dock", Apply: finderDock},
	{Name: "dock-apps", Desc: "Pin the usual apps to the Dock (Arc, Spotify, Ghostty, Discord, Obsidian, Slack)", Apply: dockApps},
}

func ByName(name string) (Setting, bool) {
	for _, s := range Settings {
		if s.Name == name {
			return s, true
		}
	}
	return Setting{}, false
}

func run(dryRun bool, name string, args ...string) error {
	if dryRun {
		fmt.Printf("[dry-run] %s %v\n", name, args)
		return nil
	}
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

const capsMapping = `{"UserKeyMapping":[{"HIDKeyboardModifierMappingSrc":0x700000039,"HIDKeyboardModifierMappingDst":0x7000000E0}]}`

const capsAgentPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.noob-cli.caps-to-ctrl</string>
	<key>ProgramArguments</key>
	<array>
		<string>/usr/bin/hidutil</string>
		<string>property</string>
		<string>--set</string>
		<string>` + capsMapping + `</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>
`

func capsToCtrl(dryRun bool) error {
	if err := run(dryRun, "hidutil", "property", "--set", capsMapping); err != nil {
		return err
	}
	home, _ := os.UserHomeDir()
	plist := filepath.Join(home, "Library/LaunchAgents/com.noob-cli.caps-to-ctrl.plist")
	if dryRun {
		fmt.Printf("[dry-run] write LaunchAgent %s\n", plist)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(plist), 0o755); err != nil {
		return err
	}
	return os.WriteFile(plist, []byte(capsAgentPlist), 0o644)
}

func raycastHotkey(dryRun bool) error {
	if err := run(dryRun, "defaults", "write", "com.apple.symbolichotkeys", "AppleSymbolicHotKeys",
		"-dict-add", "64", "<dict><key>enabled</key><false/></dict>"); err != nil {
		return err
	}
	// Command-49 is Cmd-Space; must be written before Raycast launches to be picked up
	if err := run(dryRun, "defaults", "write", "com.raycast.macos", "raycastGlobalHotkey", "-string", "Command-49"); err != nil {
		return err
	}
	if err := run(dryRun, "killall", "cfprefsd"); err != nil {
		return err
	}
	if err := run(dryRun, "open", "-ga", "Raycast"); err != nil {
		return err
	}
	if !dryRun {
		fmt.Println("  Raycast launched with Cmd-Space configured; Spotlight frees the key after logout/login.")
	}
	return nil
}

func fastKeyRepeat(dryRun bool) error {
	if err := run(dryRun, "defaults", "write", "-g", "KeyRepeat", "-int", "2"); err != nil {
		return err
	}
	return run(dryRun, "defaults", "write", "-g", "InitialKeyRepeat", "-int", "15")
}

var dockPins = []string{
	"/Applications/Arc.app",
	"/Applications/Spotify.app",
	"/Applications/Ghostty.app",
	"/Applications/Discord.app",
	"/Applications/Obsidian.app",
	"/Applications/Slack.app",
}

func dockApps(dryRun bool) error {
	out, _ := exec.Command("defaults", "read", "com.apple.dock", "persistent-apps").Output()
	current := string(out)
	added := false
	for _, app := range dockPins {
		if _, err := os.Stat(app); err != nil {
			continue
		}
		if strings.Contains(current, app) {
			fmt.Printf("  already in Dock: %s\n", filepath.Base(app))
			continue
		}
		tile := fmt.Sprintf(`<dict><key>tile-data</key><dict><key>file-data</key><dict>`+
			`<key>_CFURLString</key><string>%s</string>`+
			`<key>_CFURLStringType</key><integer>0</integer></dict></dict></dict>`, app)
		if err := run(dryRun, "defaults", "write", "com.apple.dock", "persistent-apps", "-array-add", tile); err != nil {
			return err
		}
		added = true
	}
	if added {
		return run(dryRun, "killall", "Dock")
	}
	return nil
}

func finderDock(dryRun bool) error {
	cmds := [][]string{
		{"defaults", "write", "NSGlobalDomain", "AppleShowAllExtensions", "-bool", "true"},
		{"defaults", "write", "com.apple.finder", "AppleShowAllFiles", "-bool", "true"},
		{"defaults", "write", "com.apple.finder", "ShowPathbar", "-bool", "true"},
		{"defaults", "write", "com.apple.desktopservices", "DSDontWriteNetworkStores", "-bool", "true"},
		{"defaults", "write", "com.apple.dock", "autohide-delay", "-float", "0"},
		{"defaults", "write", "com.apple.dock", "autohide-time-modifier", "-float", "0.4"},
		{"killall", "Finder"},
		{"killall", "Dock"},
	}
	for _, c := range cmds {
		if err := run(dryRun, c[0], c[1:]...); err != nil {
			return err
		}
	}
	return nil
}
