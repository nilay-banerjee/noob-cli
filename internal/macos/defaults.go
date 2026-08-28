package macos

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Setting struct {
	Name  string
	Desc  string
	Apply func(dryRun bool) error
}

var Settings = []Setting{
	{Name: "caps-to-ctrl", Desc: "Remap Caps Lock to Ctrl (survives reboot via LaunchAgent)", Apply: capsToCtrl},
	{Name: "raycast-hotkey", Desc: "Free up Cmd-Space for Raycast by disabling the Spotlight hotkey", Apply: raycastHotkey},
	{Name: "fast-key-repeat", Desc: "Key repeat faster than System Settings allows", Apply: fastKeyRepeat},
	{Name: "finder-dock", Desc: "Finder/Dock sanity: extensions, hidden files, path bar, snappier Dock", Apply: finderDock},
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
	if err := run(dryRun, "killall", "cfprefsd"); err != nil {
		return err
	}
	if !dryRun {
		fmt.Println("  Spotlight hotkey disabled; takes effect after logout. Set Raycast's hotkey to Cmd-Space in its settings.")
	}
	return nil
}

func fastKeyRepeat(dryRun bool) error {
	if err := run(dryRun, "defaults", "write", "-g", "KeyRepeat", "-int", "2"); err != nil {
		return err
	}
	return run(dryRun, "defaults", "write", "-g", "InitialKeyRepeat", "-int", "15")
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
