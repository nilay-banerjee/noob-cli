# macOS settings

Everything noob-cli changes on macOS, the exact command it runs, and how to undo it.
Applied for daily and ultimate tiers unless `--skip-defaults` is passed. In the
interactive flow each one is a checkbox.

## caps-to-ctrl

Remaps Caps Lock to Ctrl at the HID level (works in every app, no System Settings
fiddling per keyboard).

```sh
hidutil property --set '{"UserKeyMapping":[{"HIDKeyboardModifierMappingSrc":0x700000039,"HIDKeyboardModifierMappingDst":0x7000000E0}]}'
```

hidutil mappings reset on reboot, so it also writes a LaunchAgent at
`~/Library/LaunchAgents/com.noob-cli.caps-to-ctrl.plist` that reapplies the mapping
at login.

Revert:

```sh
hidutil property --set '{"UserKeyMapping":[]}'
rm ~/Library/LaunchAgents/com.noob-cli.caps-to-ctrl.plist
```

## raycast-hotkey

Disables Spotlight's Cmd-Space hotkey (symbolic hotkey 64) so Raycast can claim it.
Only applied when Raycast is in the selection.

```sh
defaults write com.apple.symbolichotkeys AppleSymbolicHotKeys -dict-add 64 "<dict><key>enabled</key><false/></dict>"
killall cfprefsd
```

Takes effect after logout/login. You still set Raycast's hotkey to Cmd-Space once,
in Raycast settings.

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
