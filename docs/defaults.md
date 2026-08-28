# macOS settings

Everything noob-cli changes on macOS, the exact command it runs, and how to undo it.
Applied for daily and ultimate tiers unless `--skip-defaults` is passed. In the
interactive flow each one is a checkbox.

## caps-to-ctrl

Writes the same native modifier mapping System Settings > Keyboard > Modifier Keys
does, per keyboard plus a `0-0-0` catch-all, so it survives reboots with no helper
process and shows up in the Settings UI. A one-shot `hidutil` call makes it take
effect immediately in the current session.

```sh
defaults -currentHost write -g com.apple.keyboard.modifiermapping.<vendor>-<product>-0 -array \
  '<dict><key>HIDKeyboardModifierMappingSrc</key><integer>30064771129</integer><key>HIDKeyboardModifierMappingDst</key><integer>30064771300</integer></dict>'
hidutil property --set '{"UserKeyMapping":[{"HIDKeyboardModifierMappingSrc":0x700000039,"HIDKeyboardModifierMappingDst":0x7000000E4}]}'
```

Keyboards are found with `hidutil list`. Older noob-cli versions used a LaunchAgent
instead; the setting removes that agent when it finds one, and `doctor` accepts
either mechanism.

Revert: System Settings > Keyboard > Keyboard Shortcuts > Modifier Keys, set
Caps Lock back to Caps Lock (or `defaults -currentHost delete -g com.apple.keyboard.modifiermapping.<id>`),
then `hidutil property --set '{"UserKeyMapping":[]}'`.

## raycast-hotkey

Disables Spotlight's Cmd-Space hotkey (symbolic hotkey 64), sets Raycast's own
hotkey to Cmd-Space via its preferences, and launches Raycast once so it registers.
Only applied when Raycast is in the selection.

```sh
defaults write com.apple.symbolichotkeys AppleSymbolicHotKeys -dict-add 64 "<dict><key>enabled</key><false/></dict>"
defaults write com.raycast.macos raycastGlobalHotkey -string "Command-49"
killall cfprefsd
open -ga Raycast
```

Spotlight fully releases the key after logout/login.

Revert: System Settings > Keyboard > Keyboard Shortcuts > Spotlight, re-enable
"Show Spotlight search". (Safer than writing the dict back by hand.)

## fast-key-repeat

Faster than the fastest System Settings slider position. Good with vim.

```sh
defaults write -g KeyRepeat -int 2
defaults write -g InitialKeyRepeat -int 15
```

Takes effect after logout/login.

Revert (macOS defaults):

```sh
defaults delete -g KeyRepeat
defaults delete -g InitialKeyRepeat
```

## finder-dock

```sh
defaults write NSGlobalDomain AppleShowAllExtensions -bool true      # show all file extensions
defaults write com.apple.finder AppleShowAllFiles -bool true         # show hidden files
defaults write com.apple.finder ShowPathbar -bool true               # path bar at the bottom
defaults write com.apple.desktopservices DSDontWriteNetworkStores -bool true  # no .DS_Store on network drives
defaults write com.apple.dock autohide-delay -float 0                # Dock appears immediately
defaults write com.apple.dock autohide-time-modifier -float 0.4      # faster Dock animation
killall Finder
killall Dock
```

Revert any of them with `defaults delete <domain> <key>` and the same killall.

## dock-apps

Pins Arc, Spotify, Ghostty, Discord, Obsidian, and Slack to the Dock by appending
to `com.apple.dock persistent-apps`, then restarts the Dock. Append-only: existing
Dock items are never removed or reordered, apps not installed are skipped.

Revert: drag the icons off the Dock.

## default-browser

Sets Arc as the default browser (Chrome when Arc isn't installed) using the
`defaultbrowser` brew formula, which it installs on demand. Arc's LaunchServices
handler name is `browser` (its bundle id is `company.thebrowser.Browser`).

```sh
brew install defaultbrowser
defaultbrowser browser   # or: defaultbrowser chrome
```

macOS always shows one confirmation dialog for default-browser changes; nothing
can suppress it, so click "Use" when it appears.

Revert: System Settings → Desktop & Dock → Default web browser, or
`defaultbrowser safari`.
