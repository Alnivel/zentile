package config

import (
	"strconv"

	log "github.com/sirupsen/logrus"
)

type workspaceConfigRaw struct {
	StartTiling *bool `toml:"start_tiling"`
	Gap         *int
	Proportion  *float64
	HideDecor   *bool `toml:"remove_decorations"`
	Layouts     []string
}

type configRaw struct {
	WorkspaceConfigs map[string]workspaceConfigRaw `toml:"workspace"`

	ProportionStep  *float64
	Keybindings     map[string]string
	WindowsToIgnore []string `toml:"ignore"`
}

type WorkspaceConfig struct {
	StartTiling bool
	Gap         int
	Proportion  float64
	HideDecor   bool
	Layouts     []string
}

type Config struct {
	globalWorkspaceConfig WorkspaceConfig
	workspaceConfigs      map[uint]WorkspaceConfig

	ProportionStep  float64
	Keybindings     map[string]string
	WindowsToIgnore []string
}

func newWorkspaceConfigFromRaw(raw workspaceConfigRaw, defaults WorkspaceConfig) WorkspaceConfig {
	config := defaults

	if raw.StartTiling != nil {
		config.StartTiling = *raw.StartTiling
	}
	if raw.Gap != nil {
		config.Gap = *raw.Gap
	}
	if raw.Proportion != nil {
		config.Proportion = *raw.Proportion
	}
	if raw.HideDecor != nil {
		config.HideDecor = *raw.HideDecor
	}
	if layouts := validateLayoutsList(raw.Layouts); layouts != nil {
		config.Layouts = layouts
	}

	return config
}

var defaultLayoutOrder = []string{"vertical", "horizontal", "fullscreen"}

func newConfigFromRaw(raw configRaw) (Config, error) {
	handleLegacyKeybindings(&raw)

	wsDefaults := WorkspaceConfig{
		StartTiling: false,
		Gap:         5,
		Proportion:  0.5,
		HideDecor:   false,
		Layouts:     defaultLayoutOrder,
	}

	globalWsConfig := newWorkspaceConfigFromRaw(raw.WorkspaceConfigs["defaults"], wsDefaults)
	delete(raw.WorkspaceConfigs, "defaults")

	workspaceConfigs := make(map[uint]WorkspaceConfig, len(raw.WorkspaceConfigs))
	for keyStr, rawWsConfig := range raw.WorkspaceConfigs {
		workspaceNum, err := strconv.ParseUint(keyStr, 10, 64)
		if err != nil {
			log.Warnf("Error during parsing config: Invalid workspace number %v - %v", keyStr, err)

			continue
		}
		workspaceConfigs[uint(workspaceNum)] = newWorkspaceConfigFromRaw(rawWsConfig, globalWsConfig)
	}

	// Top level defaults handling
	proportionStep := 0.1
	if raw.ProportionStep != nil {
		proportionStep = *raw.ProportionStep
	}

	return Config{
		globalWorkspaceConfig: globalWsConfig,
		workspaceConfigs:      workspaceConfigs,

		ProportionStep:  proportionStep,
		Keybindings:     raw.Keybindings,
		WindowsToIgnore: raw.WindowsToIgnore,
	}, nil
}

func (c Config) WorkspaceConfig(num uint) WorkspaceConfig {
	wsConfig, exists := c.workspaceConfigs[num]
	if exists {
		return wsConfig
	} else {
		return c.globalWorkspaceConfig
	}
}

func validateLayoutsList(list []string) []string {
	var result []string
	for _, layoutName := range list {
		switch layoutName {
		case "vertical":
			fallthrough
		case "horizontal":
			fallthrough
		case "fullscreen":
			result = append(result, layoutName)
		default:
			log.Warnf("Invalid layout name %v", layoutName)
		}
	}

	if result == nil && len(list) > 0 {
		log.Warn("No valid layout names provided, using defaults")
		return nil
	}

	return result
}

var legacyKeybindings = [...]string{
	"tile",
	"untile",
	"make_active_window_master",
	"increase_master",
	"decrease_master",
	"switch_layout",
	"next_window",
	"previous_window",
	"increment_master",
	"decrement_master",
}

func handleLegacyKeybindings(config *configRaw) {
	for _, command := range legacyKeybindings {
		keybind, mappingExists := config.Keybindings[command]
		if mappingExists {
			config.Keybindings[keybind] = command
			delete(config.Keybindings, command)
		}
	}
}
