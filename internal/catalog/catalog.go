package catalog

import (
	"fmt"
	"sort"
	"strings"
)

type Tier int

const (
	Server Tier = iota
	Daily
	Ultimate
)

func (t Tier) String() string {
	switch t {
	case Server:
		return "server"
	case Daily:
		return "daily"
	case Ultimate:
		return "ultimate"
	}
	return "unknown"
}

type Item struct {
	Name      string
	Desc      string
	Cask      bool
	Tier      Tier
	Brew      string
	Apt       string
	Dnf       string
	Hint      string
	AppBundle string
	Bin       string
}

func (it Item) Binary() string {
	if it.Bin != "" {
		return it.Bin
	}
	if it.Cask {
		return ""
	}
	return it.Name
}

var Items = []Item{
	{Name: "git", Desc: "version control", Tier: Server, Brew: "git", Apt: "git", Dnf: "git"},
	{Name: "zsh", Desc: "shell", Tier: Server, Brew: "zsh", Apt: "zsh", Dnf: "zsh"},
	{Name: "zsh-autosuggestions", Desc: "fish-style suggestions for zsh", Tier: Server, Brew: "zsh-autosuggestions", Apt: "zsh-autosuggestions", Dnf: "zsh-autosuggestions"},
	{Name: "zsh-syntax-highlighting", Desc: "command highlighting for zsh", Tier: Server, Brew: "zsh-syntax-highlighting", Apt: "zsh-syntax-highlighting", Dnf: "zsh-syntax-highlighting"},
	{Name: "neovim", Desc: "editor", Tier: Server, Brew: "neovim", Apt: "neovim", Dnf: "neovim", Bin: "nvim"},
	{Name: "tmux", Desc: "terminal multiplexer", Tier: Server, Brew: "tmux", Apt: "tmux", Dnf: "tmux"},
	{Name: "fzf", Desc: "fuzzy finder", Tier: Server, Brew: "fzf", Apt: "fzf", Dnf: "fzf"},
	{Name: "fd", Desc: "fast find, used by the fzf config", Tier: Server, Brew: "fd", Apt: "fd-find", Dnf: "fd-find", Hint: "apt names the binary fdfind"},
	{Name: "ripgrep", Desc: "fast grep", Tier: Server, Brew: "ripgrep", Apt: "ripgrep", Dnf: "ripgrep", Bin: "rg"},
	{Name: "bat", Desc: "cat with syntax highlighting", Tier: Server, Brew: "bat", Apt: "bat", Dnf: "bat", Hint: "installed as `batcat` on Debian/Ubuntu"},
	{Name: "eza", Desc: "modern ls", Tier: Server, Brew: "eza", Apt: "eza", Dnf: "eza", Hint: "needs Ubuntu 24.04+ / Debian 13+, otherwise see eza.rocks"},
	{Name: "zoxide", Desc: "smarter cd", Tier: Server, Brew: "zoxide", Apt: "zoxide", Dnf: "zoxide"},
	{Name: "fastfetch", Desc: "system info", Tier: Server, Brew: "fastfetch", Apt: "fastfetch", Dnf: "fastfetch", Hint: "needs Ubuntu 24.10+, otherwise ppa:zhangsongcui3371/fastfetch"},
	{Name: "lazygit", Desc: "git TUI", Tier: Server, Brew: "lazygit", Dnf: "lazygit", Hint: "no apt package; grab a release from github.com/jesseduffield/lazygit"},
	{Name: "gh", Desc: "GitHub CLI", Tier: Server, Brew: "gh", Apt: "gh", Dnf: "gh"},
	{Name: "git-delta", Desc: "better git diffs", Tier: Server, Brew: "git-delta", Apt: "git-delta", Dnf: "git-delta", Bin: "delta"},
	{Name: "mise", Desc: "runtime version manager", Tier: Server, Brew: "mise", Hint: "on Linux: curl https://mise.run | sh"},
	{Name: "tree", Desc: "directory trees", Tier: Server, Brew: "tree", Apt: "tree", Dnf: "tree"},
	{Name: "make", Desc: "build tool", Tier: Server, Brew: "make", Apt: "make", Dnf: "make"},

	{Name: "yarn", Desc: "JS package manager", Tier: Ultimate, Brew: "yarn", Hint: "on Linux: corepack enable"},
	{Name: "watchman", Desc: "file watcher", Tier: Ultimate, Brew: "watchman"},
	{Name: "arc", Desc: "Arc browser", Cask: true, Tier: Daily, Brew: "arc", AppBundle: "Arc.app"},
	{Name: "google-chrome", Desc: "Chrome", Cask: true, Tier: Daily, Brew: "google-chrome", AppBundle: "Google Chrome.app"},
	{Name: "ghostty", Desc: "terminal", Cask: true, Tier: Daily, Brew: "ghostty", AppBundle: "Ghostty.app"},
	{Name: "meslo-nerd-font", Desc: "font for the p10k prompt and Ghostty", Cask: true, Tier: Daily, Brew: "font-meslo-lg-nerd-font"},
	{Name: "raycast", Desc: "launcher", Cask: true, Tier: Daily, Brew: "raycast", AppBundle: "Raycast.app"},
	{Name: "spotify", Desc: "music", Cask: true, Tier: Daily, Brew: "spotify", AppBundle: "Spotify.app"},
	{Name: "whatsapp", Desc: "WhatsApp", Cask: true, Tier: Daily, Brew: "whatsapp", AppBundle: "WhatsApp.app"},
	{Name: "free-download-manager", Desc: "FDM", Cask: true, Tier: Daily, Brew: "free-download-manager", AppBundle: "Free Download Manager.app"},
	{Name: "obsidian", Desc: "notes", Cask: true, Tier: Daily, Brew: "obsidian", AppBundle: "Obsidian.app"},
	{Name: "1password", Desc: "password manager", Cask: true, Tier: Daily, Brew: "1password", AppBundle: "1Password.app"},
	{Name: "slack", Desc: "work chat", Cask: true, Tier: Ultimate, Brew: "slack", AppBundle: "Slack.app"},
	{Name: "visual-studio-code", Desc: "VS Code", Cask: true, Tier: Daily, Brew: "visual-studio-code", AppBundle: "Visual Studio Code.app"},
	{Name: "aerospace", Desc: "tiling window manager", Cask: true, Tier: Daily, Brew: "nikitabobko/tap/aerospace", AppBundle: "AeroSpace.app"},
	{Name: "claude", Desc: "Claude desktop", Cask: true, Tier: Daily, Brew: "claude", AppBundle: "Claude.app"},
	{Name: "claude-code", Desc: "Claude Code CLI", Cask: true, Tier: Daily, Brew: "claude-code@latest", Bin: "claude"},
	{Name: "docker", Desc: "Docker Desktop", Cask: true, Tier: Daily, Brew: "docker-desktop", AppBundle: "Docker.app"},

	{Name: "ffmpeg", Desc: "media swiss army knife", Tier: Ultimate, Brew: "ffmpeg", Apt: "ffmpeg", Dnf: "ffmpeg"},
	{Name: "iperf3", Desc: "network throughput testing", Tier: Ultimate, Brew: "iperf3", Apt: "iperf3", Dnf: "iperf3"},
	{Name: "discord", Desc: "Discord", Cask: true, Tier: Ultimate, Brew: "discord", AppBundle: "Discord.app"},
	{Name: "obs", Desc: "screen recording / streaming", Cask: true, Tier: Ultimate, Brew: "obs", AppBundle: "OBS.app"},
	{Name: "betterdisplay", Desc: "display management", Cask: true, Tier: Ultimate, Brew: "betterdisplay", AppBundle: "BetterDisplay.app"},
	{Name: "hidden-bar", Desc: "menu bar tidying", Cask: true, Tier: Ultimate, Brew: "hiddenbar", AppBundle: "Hidden Bar.app"},
	{Name: "firefox", Desc: "Firefox", Cask: true, Tier: Ultimate, Brew: "firefox", AppBundle: "Firefox.app"},
	{Name: "vlc", Desc: "media player", Cask: true, Tier: Ultimate, Brew: "vlc", AppBundle: "VLC.app"},
	{Name: "zoom", Desc: "video calls", Cask: true, Tier: Ultimate, Brew: "zoom", AppBundle: "zoom.us.app"},
	{Name: "anydesk", Desc: "remote desktop", Cask: true, Tier: Ultimate, Brew: "anydesk", AppBundle: "AnyDesk.app"},
	{Name: "tailscale", Desc: "mesh VPN", Cask: true, Tier: Ultimate, Brew: "tailscale-app", AppBundle: "Tailscale.app"},
}

func ByName(name string) (Item, bool) {
	for _, it := range Items {
		if it.Name == name {
			return it, true
		}
	}
	return Item{}, false
}

func ForTier(t Tier) []Item {
	var out []Item
	for _, it := range Items {
		if it.Tier <= t {
			out = append(out, it)
		}
	}
	return out
}

func Names() []string {
	names := make([]string, len(Items))
	for i, it := range Items {
		names[i] = it.Name
	}
	sort.Strings(names)
	return names
}

func Resolve(tier Tier, include, exclude []string) ([]Item, error) {
	selected := map[string]Item{}
	for _, it := range ForTier(tier) {
		selected[it.Name] = it
	}
	for _, name := range include {
		it, ok := ByName(name)
		if !ok {
			return nil, unknownName(name)
		}
		selected[it.Name] = it
	}
	for _, name := range exclude {
		if _, ok := ByName(name); !ok {
			return nil, unknownName(name)
		}
		delete(selected, name)
	}
	var out []Item
	for _, it := range Items {
		if _, ok := selected[it.Name]; ok {
			out = append(out, it)
		}
	}
	return out, nil
}

func unknownName(name string) error {
	return fmt.Errorf("unknown app %q — supported names:\n  %s", name, strings.Join(Names(), ", "))
}
