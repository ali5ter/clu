// Package config loads ~/.config/clu/config.toml via Viper.
package config

import (
	"github.com/spf13/viper"
)

// Config holds runtime configuration for clu.
type Config struct {
	Theme    string `mapstructure:"theme"`
	CacheTTL int    `mapstructure:"cache_ttl"` // seconds
	Browser  string `mapstructure:"browser"`
}

// Load reads config from ~/.config/clu/config.toml, returning defaults if absent.
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("$HOME/.config/clu")

	viper.SetDefault("theme", "auto")
	viper.SetDefault("cache_ttl", 3600)
	viper.SetDefault("browser", "")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
