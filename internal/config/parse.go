package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/go-homedir"
)

func InitConfig() (Config, error) {
	configRaw := configRaw{}
	writeDefaultConfig()
	_, err := toml.DecodeFile(configFilePath(), &configRaw)
	if err != nil {
		return Config{}, err
	}
	config, err := newConfigFromRaw(configRaw)
	return config, err
}

func writeDefaultConfig() {
	if _, err := os.Stat(configFolderPath()); os.IsNotExist(err) {
		os.MkdirAll(configFolderPath(), 0700)
	}

	if _, err := os.Stat(configFilePath()); os.IsNotExist(err) {
		os.WriteFile(configFilePath(), []byte(defaultConfig), 0644)
	}
}

func configFolderPath() string {
	var configFolder string
	switch runtime.GOOS {
	case "linux":
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome != "" {
			configFolder = filepath.Join(xdgConfigHome, "zentile")
		} else {
			configFolder, _ = homedir.Expand("~/.config/zentile/")
		}
	default:
		configFolder, _ = homedir.Expand("~/.zentile/")
	}

	return configFolder
}

func configFilePath() string {
	return filepath.Join(configFolderPath(), "config.toml")
}

var defaultConfig = `
# ===== Global Settings =====

# How much to increment the master area size.
proportionStep = 0.04

# Zentile will ignore windows added to this list.
# You'll have to add WM_CLASS property of the window you want ignored.
# You can get WM_CLASS property of a window, by running "xprop WM_CLASS" and clicking on the window.
# ignore = ['ulauncher', 'gnome-screenshot']

# Defaults for all workspaces
[workspace.defaults]
# Gap between windows.
gap = 5

# Default propotion between master and stack.
proportion = 0.5

 # Window decorations will be removed when tiling if set to true
remove_decorations = false

# Per workspace overrides. The first workspace is 0
[workspace.1]
gap = 0

[workspace.2]
proportion = 0.6

# ===== Keybindings =====
[keybindings]
# key sequences can have zero or more modifiers and exactly one key.
# example: Control-Shift-t has two modifiers and one key.
# You can view which keys activate which modifier using the 'xmodmap' program.
# Key symbols can be found by pressing keys using the 'xev' program

# Tile the current workspace.
"Control-Shift-t" = "tile"

# Untile the current workspace.
"Control-Shift-u" = "untile"

# Make the active window as master.
"Control-Shift-m" = "make_active_window_master"

# Increase the number of masters.
"Control-Shift-i" = "increase_master"

# Decrease the number of masters.
"Control-Shift-d" = "decrease_master"

# Cycles through the available layouts.
"Control-Shift-s" = "switch_layout"

# Moves focus to the next window.
"Control-Shift-n" = "next_window"

# Moves focus to the previous window.
"Control-Shift-p" = "previous_window"

# Increases the size of the master windows.
"Control-bracketright" = "increment_master"

# Decreases the size of the master windows.
"Control-bracketleft" = "decrement_master"
`
