// Package config loads cast settings from TOML, then CAST_* env overrides.
package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds resolved cast settings.
type Config struct {
	Targets  []string `toml:"targets"`
	Outdir   string   `toml:"outdir"`
	Optimize bool     `toml:"optimize"`
	Agent    string   `toml:"agent"`
}

func defaults() *Config {
	return &Config{Targets: []string{"codex", "cursor"}, Outdir: "dist", Optimize: false, Agent: "auto"}
}

// Load reads ~/.config/cast/config.toml (honoring XDG_CONFIG_HOME), then
// applies CAST_* environment overrides. Missing file is not an error.
func Load() (*Config, error) {
	c := defaults()
	if path := configPath(); path != "" {
		if _, err := os.Stat(path); err == nil {
			if _, err := toml.DecodeFile(path, c); err != nil {
				return nil, err
			}
		}
	}
	if v := os.Getenv("CAST_TARGETS"); v != "" {
		c.Targets = strings.Split(v, ",")
	}
	if v := os.Getenv("CAST_OUTDIR"); v != "" {
		c.Outdir = v
	}
	if v := os.Getenv("CAST_OPTIMIZE"); v == "1" || v == "true" {
		c.Optimize = true
	}
	if v := os.Getenv("CAST_AGENT"); v != "" {
		c.Agent = v
	}
	return c, nil
}

func configPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "cast", "config.toml")
}
