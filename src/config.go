package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	XMPP     XMPPConfig     `yaml:"xmpp"`
	RSS      RSSConfig      `yaml:"rss"`
	Database DatabaseConfig `yaml:"database"`
}

type XMPPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	JID      string `yaml:"jid"`
	Password string `yaml:"password"`
}

type RSSConfig struct {
	PollInterval int    `yaml:"poll_interval"`
	UserAgent    string `yaml:"user_agent"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.RSS.PollInterval == 0 {
		cfg.RSS.PollInterval = 3600
	}

	if cfg.RSS.UserAgent == "" {
		cfg.RSS.UserAgent = "RSSBot/1.0"
	}

	if cfg.Database.Path == "" {
		cfg.Database.Path = "./rss.db"
	}

	return &cfg, nil
}
