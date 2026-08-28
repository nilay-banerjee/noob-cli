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

## daily tier (adds to server)

| Name | What | brew |
|---|---|---|
| yarn | JS package manager | yarn |
| watchman | file watcher | watchman |
| arc | Arc browser | arc (cask) |
| google-chrome | Chrome | google-chrome (cask) |
| ghostty | terminal | ghostty (cask) |
| raycast | launcher | raycast (cask) |
| spotify | music | spotify (cask) |
| whatsapp | WhatsApp | whatsapp (cask) |
| free-download-manager | FDM | free-download-manager (cask) |
| obsidian | notes | obsidian (cask) |
| 1password | password manager | 1password (cask) |
| slack | work chat | slack (cask) |
| visual-studio-code | VS Code | visual-studio-code (cask) |
| aerospace | tiling WM | nikitabobko/tap/aerospace (cask) |
| claude | Claude desktop | claude (cask) |
| claude-code | Claude Code CLI | claude-code@latest (cask) |
| docker | Docker Desktop | docker-desktop (cask) |

## ultimate tier (adds to daily)

| Name | What | brew |
|---|---|---|
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

## Deliberately not in the catalog

Installed by hand when a machine actually needs them:

- Xcode, Android Studio (App Store / heavyweight installers)
- STM32CubeIDE, Arduino IDE, gcc-arm-embedded, open-ocd (embedded toolchains, per-project)
- Lightshot (Mac App Store only, no cask)
- Work-specific taps: neetodeploy, tunnelto, lumen
- postgresql, redis, mongosh (per-project services, versions matter)
