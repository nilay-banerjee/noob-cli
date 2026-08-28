# App catalog

Names in the first column are what `--include` and `--exclude` accept.
Casks are macOS-only and skipped on Linux. An empty apt/dnf cell means the
installer lists it under "needs manual install" with a hint instead of failing.

## server tier

| Name | What | brew | apt | dnf |
|---|---|---|---|---|
| git | version control | git | git | git |
| zsh | shell | zsh | zsh | zsh |
| zsh-autosuggestions | fish-style suggestions | zsh-autosuggestions | zsh-autosuggestions | zsh-autosuggestions |
| zsh-syntax-highlighting | command highlighting | zsh-syntax-highlighting | zsh-syntax-highlighting | zsh-syntax-highlighting |
| neovim | editor | neovim | neovim | neovim |
| tmux | terminal multiplexer | tmux | tmux | tmux |
| fzf | fuzzy finder | fzf | fzf | fzf |
| fd | fast find (fzf uses it) | fd | fd-find (binary is `fdfind`) | fd-find |
| ripgrep | fast grep | ripgrep | ripgrep | ripgrep |
| bat | cat with highlighting | bat | bat (binary is `batcat`) | bat |
| eza | modern ls | eza | eza (Ubuntu 24.04+) | eza |
| zoxide | smarter cd | zoxide | zoxide | zoxide |
| fastfetch | system info | fastfetch | fastfetch (Ubuntu 24.10+) | fastfetch |
| lazygit | git TUI | lazygit | — (GitHub releases) | lazygit |
| gh | GitHub CLI | gh | gh | gh |
| git-delta | better diffs | git-delta | git-delta | git-delta |
| mise | runtime versions | mise | — (`curl https://mise.run \| sh`) | — (same) |
| tree | directory trees | tree | tree | tree |
| make | build tool | make | make | make |
| mole | Mac cleanup tool (mole.fit) | mole | — (macOS-only) | — (macOS-only) |
| lumen | AI git assistant | jnsahaj/lumen/lumen | — (see repo) | — (see repo) |

## daily tier (adds to server)

| Name | What | brew |
|---|---|---|
| arc | Arc browser | arc (cask) |
| google-chrome | Chrome | google-chrome (cask) |
| ghostty | terminal | ghostty (cask) |
| meslo-nerd-font | MesloLGS Nerd Font (p10k + Ghostty need it) | font-meslo-lg-nerd-font (cask) |
| raycast | launcher | raycast (cask) |
| spotify | music | spotify (cask) |
| whatsapp | WhatsApp | whatsapp (cask) |
| free-download-manager | FDM | free-download-manager (cask) |
| obsidian | notes | obsidian (cask) |
| 1password | password manager | 1password (cask) |
| visual-studio-code | VS Code | visual-studio-code (cask) |
| aerospace | tiling WM | nikitabobko/tap/aerospace (cask) |
| claude | Claude desktop | claude (cask) |
| claude-code | Claude Code CLI | claude-code@latest (cask) |
| colima | container runtime VM | colima (macOS; Linux runs docker natively) |
| docker | Docker CLI | docker (apt: docker.io, dnf: docker) |

## ultimate tier (adds to daily)

| Name | What | brew |
|---|---|---|
| yarn | JS package manager | yarn |
| watchman | file watcher | watchman |
| slack | work chat | slack (cask) |
| ffmpeg | media tooling | ffmpeg (also apt/dnf) |
| iperf3 | network throughput | iperf3 (also apt/dnf) |
| discord | Discord | discord (cask) |
| obs | recording/streaming | obs (cask) |
| betterdisplay | display management | betterdisplay (cask) |
| hidden-bar | menu bar tidying | hiddenbar (cask) |
| firefox | Firefox | firefox (cask) |
| vlc | media player | vlc (cask) |
| zoom | video calls | zoom (cask) |
| anydesk | remote desktop | anydesk (cask) |
| tailscale | mesh VPN | tailscale-app (cask) |

## Extras (not packages)

Git clones that `.zshrc` depends on, done automatically after installs:

- fzf selected → [fzf-git.sh](https://github.com/junegunn/fzf-git.sh) to `~/fzf-git.sh`
- zsh selected → [oh-my-zsh](https://github.com/ohmyzsh/ohmyzsh) to `~/.oh-my-zsh`, plus
  zsh-autosuggestions and zsh-syntax-highlighting into its custom plugins and
  [powerlevel10k](https://github.com/romkatv/powerlevel10k) into its custom themes.
  The omz plugin system loads these clones; the brew formulas of the same plugins
  cover non-omz setups (Linux servers).

## Deliberately not in the catalog

Installed by hand when a machine actually needs them:

- Xcode, Android Studio (App Store / heavyweight installers)
- STM32CubeIDE, Arduino IDE, gcc-arm-embedded, open-ocd (embedded toolchains, per-project)
- Lightshot (Mac App Store only, no cask)
- Work-specific taps: neetodeploy, tunnelto, lumen
- postgresql, redis, mongosh (per-project services, versions matter)
