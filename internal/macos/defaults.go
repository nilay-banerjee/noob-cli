package macos

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Setting struct {
	Name  string
	Desc  string
	Apply func(dryRun bool) error
}

var Settings = []Setting{
	{Name: "caps-to-ctrl", Desc: "Remap Caps Lock to Ctrl (native modifier mapping, instant via hidutil)", Apply: capsToCtrl},
	{Name: "raycast-hotkey", Desc: "Give Cmd-Space to Raycast: disable Spotlight hotkey, set Raycast's, launch it", Apply: raycastHotkey},
	{Name: "fast-key-repeat", Desc: "Key repeat faster than System Settings allows", Apply: fastKeyRepeat},
	{Name: "finder-dock", Desc: "Finder/Dock sanity: extensions, hidden files, path bar, snappier Dock", Apply: finderDock},
	{Name: "dock-apps", Desc: "Pin the usual apps to the Dock (Arc, Spotify, Ghostty, Discord, Obsidian, Slack)", Apply: dockApps},
	{Name: "default-browser", Desc: "Make Arc the default browser (Chrome if Arc is missing; macOS confirms once)", Apply: defaultBrowser},
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

const cmdSpaceHotkey = "Command-49"

const capsMapping = `{"UserKeyMapping":[{"HIDKeyboardModifierMappingSrc":0x700000039,"HIDKeyboardModifierMappingDst":0x7000000E4}]}`

const modifierMappingEntry = `<dict>` +
	`<key>HIDKeyboardModifierMappingSrc</key><integer>30064771129</integer>` +
	`<key>HIDKeyboardModifierMappingDst</key><integer>30064771300</integer>` +
	`</dict>`

func CapsAgentPlist() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library/LaunchAgents/com.noob-cli.caps-to-ctrl.plist")
}

func capsToCtrl(dryRun bool) error {
	if err := run(dryRun, "hidutil", "property", "--set", capsMapping); err != nil {
		return err
	}
	for _, id := range keyboardMappingIDs() {
		if err := run(dryRun, "defaults", "-currentHost", "write", "-g",
			"com.apple.keyboard.modifiermapping."+id, "-array", modifierMappingEntry); err != nil {
			return err
		}
	}
	if !dryRun {
		os.Remove(CapsAgentPlist())
	}
	return nil
}

func keyboardMappingIDs() []string {
	ids := map[string]bool{"0-0-0": true}
	out, _ := exec.Command("hidutil", "list", "--matching", `{"DeviceUsagePage":1,"DeviceUsage":6}`).Output()
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || !strings.HasPrefix(fields[0], "0x") {
			continue
		}
		vendor, vendorErr := strconv.ParseInt(fields[0][2:], 16, 64)
		product, productErr := strconv.ParseInt(fields[1][2:], 16, 64)
		if vendorErr != nil || productErr != nil {
			continue
		}
		ids[fmt.Sprintf("%d-%d-0", vendor, product)] = true
	}
	sorted := make([]string, 0, len(ids))
	for id := range ids {
		sorted = append(sorted, id)
	}
	sort.Strings(sorted)
	return sorted
}

func raycastHotkey(dryRun bool) error {
	if err := run(dryRun, "defaults", "write", "com.apple.symbolichotkeys", "AppleSymbolicHotKeys",
		"-dict-add", "64", "<dict><key>enabled</key><false/></dict>"); err != nil {
		return err
	}
	if err := run(dryRun, "defaults", "write", "com.raycast.macos", "raycastGlobalHotkey", "-string", cmdSpaceHotkey); err != nil {
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

const (
	arcBrowserHandler    = "browser"
	chromeBrowserHandler = "chrome"
)

func defaultBrowser(dryRun bool) error {
	handler := arcBrowserHandler
	if _, err := os.Stat("/Applications/Arc.app"); err != nil {
		handler = chromeBrowserHandler
	}
	if _, err := exec.LookPath("defaultbrowser"); err != nil {
		if err := run(dryRun, "brew", "install", "defaultbrowser"); err != nil {
			return err
		}
	}
	if err := run(dryRun, "defaultbrowser", handler); err != nil {
		return err
	}
	if !dryRun {
		fmt.Println("  approve the macOS dialog to finish switching the default browser")
	}
	return nil
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
