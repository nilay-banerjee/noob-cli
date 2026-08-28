# Dotfiles

Configs live in a git repo at `~/dotfiles`. The live paths in `$HOME` are symlinks
into the repo, so editing a config edits the repo copy and `git diff` in `~/dotfiles`
shows what changed.

## Symlink map

| Live path | Repo path |
|---|---|
| `~/.zshrc` | `~/dotfiles/zsh/.zshrc` |
| `~/.p10k.zsh` | `~/dotfiles/zsh/.p10k.zsh` |
| `~/.tmux.conf` | `~/dotfiles/tmux/.tmux.conf` |
| `~/.gitconfig` | `~/dotfiles/git/.gitconfig` |
| `~/.config/nvim` | `~/dotfiles/config/nvim` |
| `~/.config/aerospace` | `~/dotfiles/config/aerospace` |
| `~/.config/ghostty` | `~/dotfiles/config/ghostty` |
| `~/Library/Application Support/com.mitchellh.ghostty/config` | `~/dotfiles/ghostty/config` (macOS) |
| `~/.config/lazygit` | `~/dotfiles/config/lazygit` |

Paths that don't exist on a machine are skipped. Adding a config means adding one
entry to `Links` in `internal/dotfiles/dotfiles.go`.

## First time (the machine that has your configs)

```sh
noob-cli dotfiles init
```

For each live config: moves it into `~/dotfiles`, symlinks it back, then
`git init` + first commit if the repo is new. Already-linked paths are left alone,
and it refuses to overwrite a repo path that already exists.

Then push it somewhere:

```sh
cd ~/dotfiles
git remote add origin git@github.com:nilay-banerjee/dotfiles.git
git push -u origin main
```

## Fresh machine

```sh
noob-cli dotfiles link
```

Clones `https://github.com/nilay-banerjee/dotfiles.git` if `~/dotfiles` is missing
(override with `--dotfiles-repo`), then symlinks every repo config into place.
Existing real files (like the stock `.zshrc` a fresh install ships) are moved to a
timestamped backup dir under `$TMPDIR` first, and the path is printed.

The main `noob-cli` setup run does the same clone-and-link at the end unless
`--skip-dotfiles` is passed.
