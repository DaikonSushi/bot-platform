package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	NapCat        NapCatConfig        `yaml:"napcat"`
	Bot           BotConfig           `yaml:"bot"`
	PluginManager PluginManagerConfig `yaml:"plugin_manager"`
	AdminServer   AdminServerConfig   `yaml:"admin_server"`
	Plugins       PluginsConfig       `yaml:"plugins"`
	Help          HelpConfig          `yaml:"help"`
}

// NapCatConfig holds NapCat connection settings
type NapCatConfig struct {
	HttpURL string `yaml:"http_url"`
	WsURL   string `yaml:"ws_url"`
	Token   string `yaml:"token"`
}

// BotConfig holds bot behavior settings
type BotConfig struct {
	Admins        []int64 `yaml:"admins"`
	CommandPrefix string  `yaml:"command_prefix"`
	Debug         bool    `yaml:"debug"`
}

// PluginManagerConfig holds external plugin manager settings
type PluginManagerConfig struct {
	Enabled   bool     `yaml:"enabled"`
	PluginDir string   `yaml:"plugin_dir"`
	ConfigDir string   `yaml:"config_dir"`
	GRPCPort  int      `yaml:"grpc_port"`
	AutoStart []string `yaml:"auto_start"`
}

// AdminServerConfig holds admin HTTP API settings
type AdminServerConfig struct {
	Enabled bool   `yaml:"enabled"`
	Addr    string `yaml:"addr"`
}

// PluginsConfig holds built-in plugin settings
type PluginsConfig struct {
	Enabled []string `yaml:"enabled"`
}

// HelpConfig holds help plugin customization settings
type HelpConfig struct {
	Title       string `yaml:"title"`        // Custom title for help menu
	Description string `yaml:"description"` // Custom description/header text
	Footer      string `yaml:"footer"`      // Custom footer text
	ShowBuiltin bool   `yaml:"show_builtin"` // Show built-in plugins (default: true)
	ShowExternal bool  `yaml:"show_external"` // Show external plugins (default: true)
}

// Load reads and parses config from yaml file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Bot.CommandPrefix == "" {
		cfg.Bot.CommandPrefix = "/"
	}
	if cfg.PluginManager.PluginDir == "" {
		cfg.PluginManager.PluginDir = "./plugins-bin"
	}
	if cfg.PluginManager.ConfigDir == "" {
		cfg.PluginManager.ConfigDir = "./plugins-config"
	}
	if cfg.PluginManager.GRPCPort == 0 {
		cfg.PluginManager.GRPCPort = 50051
	}
	if cfg.AdminServer.Addr == "" {
		cfg.AdminServer.Addr = ":8080"
	}
	// Help defaults
	if cfg.Help.Title == "" {
		cfg.Help.Title = "📖 Bot Help Menu"
	}
	if !cfg.Help.ShowBuiltin && !cfg.Help.ShowExternal {
		// If both are false, enable both by default
		cfg.Help.ShowBuiltin = true
		cfg.Help.ShowExternal = true
	}

	return &cfg, nil
}

// IsAdmin checks if a user is an admin
func (c *Config) IsAdmin(userID int64) bool {
	for _, admin := range c.Bot.Admins {
		if admin == userID {
			return true
		}
	}
	return false
}
