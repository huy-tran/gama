package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Github         Github            `mapstructure:"github"`
	Shortcuts      Shortcuts         `mapstructure:"keys"`
	Settings       Settings          `mapstructure:"settings"`
	CustomCommands []CustomCommand   `mapstructure:"custom_commands"`
	// RepoPaths maps "owner/repo" slugs to absolute local filesystem paths.
	// Only repos you actively work on locally need an entry.
	// Supports ~ expansion (e.g. "~/Herd/my-app").
	// Used to resolve {{.RepoPath}} in custom command templates.
	RepoPaths map[string]string `mapstructure:"repo_paths"`
}

// CustomCommand adds an entry to the Repository tab's action bar.
// Command supports Go template variables:
//
//   - {{.RepoName}} — the full "owner/repo" slug of the selected repository
//   - {{.RepoPath}} — local path looked up from repo_paths; empty if not mapped
type CustomCommand struct {
	Description string `mapstructure:"description"`
	Command     string `mapstructure:"command"`
}

type Settings struct {
	LiveMode struct {
		Enabled  bool          `mapstructure:"enabled"`
		Interval time.Duration `mapstructure:"interval"`
	} `mapstructure:"live_mode"`
}

type Github struct {
	Token string `mapstructure:"token"`
}

type Shortcuts struct {
	SwitchTabRight string `mapstructure:"switch_tab_right"`
	SwitchTabLeft  string `mapstructure:"switch_tab_left"`
	Quit           string `mapstructure:"quit"`
	Refresh        string `mapstructure:"refresh"`
	Enter          string `mapstructure:"enter"`
	LiveMode       string `mapstructure:"live_mode"`
	Tab            string `mapstructure:"tab"`
}

func LoadConfig() (*Config, error) {
	var config = new(Config)
	defer func() {
		config = fillDefaultShortcuts(config)
		config = fillDefaultSettings(config)
	}()

	setConfig()

	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	if err := viper.BindEnv("github.token", "GITHUB_TOKEN"); err != nil {
		return nil, fmt.Errorf("failed to bind environment variable: %w", err)
	}
	viper.AutomaticEnv()

	// Read the config file first
	if err := viper.ReadInConfig(); err == nil {
		if err := viper.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
		}
		return config, nil
	}

	// If config file is not found, try to unmarshal from environment variables
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

func setConfig() {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "gama", "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "gama"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		return
	}

	oldConfigPath := filepath.Join(os.Getenv("HOME"), ".gama.yaml")
	if _, err := os.Stat(oldConfigPath); err == nil {
		viper.AddConfigPath(os.Getenv("HOME"))
		viper.SetConfigName(".gama")
		viper.SetConfigType("yaml")
		return
	}
}
