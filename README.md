# noob-cli

Sets up a fresh machine the way I like it. One binary that installs my apps and CLIs,
links dotfiles from `~/dotfiles`, and applies the macOS settings I always end up
changing by hand.

Targets macOS (full setup) and Linux (CLI tools only, apt or dnf).

## Install

On a fresh machine (no Go, no brew, just curl):

```sh
curl -fsSL https://raw.githubusercontent.com/nilay-banerjee/noob-cli/main/install.sh | sh
```

Downloads the prebuilt binary for your OS/arch from the latest GitHub release into
`~/.local/bin` (override with `BINDIR=/usr/local/bin`).

Or build from source if you have Go:

```sh
git clone https://github.com/nilay-banerjee/noob-cli.git
cd noob-cli
go build -o noob-cli .
```

noob-cli installs Homebrew itself if it's missing. Cutting a new release is
`scripts/release.sh v0.x.y` (cross-compiles, tags, pushes, publishes with gh).

Don't run noob-cli with sudo: Homebrew refuses to run as root, and files written
to `$HOME` would end up root-owned. noob-cli asks for your password once at the
start and keeps the sudo session warm for the rest of the run. GUI permission
dialogs (App Management, Accessibility for AeroSpace/Raycast) come from macOS
privacy protections and can't be pre-approved from a script.

## Usage

```sh
noob-cli                 # interactive: pick a tier, then toggle apps step by step
noob-cli --daily         # the default set, no prompts
noob-cli --server        # lightweight CLI-only set, works on Linux
noob-cli --ultimate      # daily plus the nice-to-haves

noob-cli --daily --exclude spotify,docker
noob-cli --server --include ffmpeg
noob-cli --ultimate --dry-run          # print the plan, change nothing
```

Flags:

| Flag | What it does |
|---|---|
| `--server` / `--daily` / `--ultimate` | Pick a tier. No tier flag on a terminal opens the interactive step-through. |
| `--include a,b` | Add supported apps on top of the tier. |
| `--exclude a,b` | Drop apps from the tier. |
| `--dry-run` | Print every command instead of running it. |
| `--skip-defaults` | Leave macOS settings alone. |
| `--skip-dotfiles` | Don't touch dotfiles. |
| `--dotfiles-repo <url>` | Clone this into `~/dotfiles` if it's missing. |

Unknown names in `--include`/`--exclude` fail fast and print the supported list.

## Tiers

- **server**: git, zsh + plugins, neovim, tmux, fzf, fd, ripgrep, bat, eza, zoxide,
  fastfetch, lazygit, gh, git-delta, mise, tree, make. No GUI apps, no settings
  changes. This is the one to run on a Linux box. When fzf is selected, fzf-git.sh
  is cloned to `~/fzf-git.sh` too (`.zshrc` sources it).
- **daily** (default): server + the mac apps I actually open every day: Arc,
  Chrome, Ghostty, Raycast, Spotify, WhatsApp, FDM, Obsidian, 1Password,
  VS Code, AeroSpace, Claude, Claude Code, Docker Desktop, Meslo Nerd Font.
- **ultimate**: daily + work/dev extras (yarn, watchman, Slack) and Discord,
  OBS, BetterDisplay, Hidden Bar, Firefox, VLC, Zoom, AnyDesk, Tailscale,
  ffmpeg, iperf3.

Full catalog with package names per platform: [docs/apps.md](docs/apps.md).

## Dotfiles

Configs live in a git repo at `~/dotfiles` and get symlinked into place, so editing
`~/.zshrc` edits the repo copy.

```sh
noob-cli dotfiles init   # on this machine: move live configs into ~/dotfiles, symlink back, git init
noob-cli dotfiles link   # on a fresh machine: symlink ~/dotfiles configs into place
```

## Maintenance

```sh
noob-cli doctor    # verify the machine: links, clones, font, settings, zsh/nvim load, dotfiles sync
noob-cli upgrade   # replace this binary with the latest GitHub release
```

`doctor` exits non-zero when something is off and says how to fix each item.

Details and the full symlink map: [docs/dotfiles.md](docs/dotfiles.md).

## macOS settings

Applied for daily and ultimate (never for server, never on Linux):

- Caps Lock remapped to Ctrl, persisted across reboots with a LaunchAgent
- Spotlight's Cmd-Space hotkey disabled so Raycast can take it
- Key repeat faster than System Settings allows
- Finder and Dock sanity (extensions, hidden files, path bar, snappier autohide)

Every default written, and how to revert each one: [docs/defaults.md](docs/defaults.md).

## Layout

```
cmd/                 cobra commands, interactive flow, output formatting
internal/catalog/    the app list: tiers + brew/apt/dnf package names
internal/installer/  brew (macOS) and apt/dnf (Linux) installers
internal/dotfiles/   harvest, clone, and symlink logic
internal/macos/      the settings applied via defaults/hidutil
```

Adding an app is one line in `internal/catalog/catalog.go`.
